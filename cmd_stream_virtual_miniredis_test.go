//go:build miniredis_test

package redisson

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"
)

// 这套测试在 miniredis 中验证虚拟 Stream Hub 端到端工作，与
// cmd_stream_virtual_close_test.go 形成纯单元 + e2e 互补。
//
// 跑法：
//   go test -tags 'miniredis_test redisson_miniredis' -run TestVirtualStreamMiniredis ./...

// streamXAdd 简化的 XAdd 帮手：每次插入一条带 id 自动生成的消息。
func streamXAdd(t *testing.T, c Cmdable, ctx context.Context, stream, value string) {
	t.Helper()
	r := c.XAdd(ctx, XAddArgs{
		Stream: stream,
		Values: map[string]any{"v": value},
	})
	if r.Err() != nil {
		t.Fatalf("xadd %s=%s: %v", stream, value, r.Err())
	}
}

func waitStreamMessages(t *testing.T, v VirtualStream, n int, timeout time.Duration) []StreamMessage {
	t.Helper()
	got := make([]StreamMessage, 0, n)
	deadline := time.After(timeout)
	for len(got) < n {
		select {
		case m := <-v.Channel():
			got = append(got, m)
		case <-deadline:
			t.Fatalf("only got %d/%d messages within %s: %v", len(got), n, timeout, got)
		}
	}
	return got
}

func TestVirtualStreamMiniredis_BasicReadFromStart(t *testing.T) {
	t.Parallel()
	c := MustNewClient(NewConf(WithT(t), WithEnableCache(false)))
	t.Cleanup(func() { _ = c.Close() })
	hub := c.NewVirtualStreamHub()
	t.Cleanup(func() { _ = hub.Close() })

	ctx := context.Background()
	// 先写入 3 条历史消息
	for i := 0; i < 3; i++ {
		streamXAdd(t, c, ctx, "s.basic", "v"+strconv.Itoa(i))
	}

	// 从头订阅（startID="0"）应能读到 3 条历史消息
	v, err := hub.Subscribe(ctx, []string{"s.basic"}, "0")
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	t.Cleanup(func() { _ = v.Close() })

	msgs := waitStreamMessages(t, v, 3, 2*time.Second)
	for i, m := range msgs {
		if m.Stream != "s.basic" {
			t.Fatalf("msg[%d].Stream = %q", i, m.Stream)
		}
		if got := m.Message.Values["v"].(string); got != "v"+strconv.Itoa(i) {
			t.Fatalf("msg[%d] value %q, want v%d", i, got, i)
		}
	}
}

func TestVirtualStreamMiniredis_DefaultDollarReadsOnlyNew(t *testing.T) {
	t.Parallel()
	c := MustNewClient(NewConf(WithT(t), WithEnableCache(false)))
	t.Cleanup(func() { _ = c.Close() })
	hub := c.NewVirtualStreamHub()
	t.Cleanup(func() { _ = hub.Close() })

	ctx := context.Background()
	// 先写入 1 条历史消息
	streamXAdd(t, c, ctx, "s.dollar", "old")

	// 默认 startID="$" 不应读到历史
	v, err := hub.Subscribe(ctx, []string{"s.dollar"})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	t.Cleanup(func() { _ = v.Close() })

	// 给 worker 时间发起 XREAD（订阅时序保证）
	time.Sleep(200 * time.Millisecond)

	// 写新消息
	streamXAdd(t, c, ctx, "s.dollar", "new")

	msgs := waitStreamMessages(t, v, 1, 2*time.Second)
	if got := msgs[0].Message.Values["v"].(string); got != "new" {
		t.Fatalf("got %q, want new (default $ should skip historical)", got)
	}
}

func TestVirtualStreamMiniredis_FanoutToMultipleSubscribers(t *testing.T) {
	t.Parallel()
	c := MustNewClient(NewConf(WithT(t), WithEnableCache(false)))
	t.Cleanup(func() { _ = c.Close() })
	hub := c.NewVirtualStreamHub()
	t.Cleanup(func() { _ = hub.Close() })

	ctx := context.Background()
	v1, err := hub.Subscribe(ctx, []string{"s.fan"}, "0")
	if err != nil {
		t.Fatalf("v1: %v", err)
	}
	v2, err := hub.Subscribe(ctx, []string{"s.fan"}, "0")
	if err != nil {
		t.Fatalf("v2: %v", err)
	}
	t.Cleanup(func() { _ = v1.Close() })
	t.Cleanup(func() { _ = v2.Close() })

	// 等待 hub worker 进入 XREAD
	time.Sleep(200 * time.Millisecond)

	streamXAdd(t, c, ctx, "s.fan", "hello")

	get := func(v VirtualStream) string {
		select {
		case m := <-v.Channel():
			return m.Message.Values["v"].(string)
		case <-time.After(2 * time.Second):
			return ""
		}
	}
	if g := get(v1); g != "hello" {
		t.Fatalf("v1 got %q, want hello", g)
	}
	if g := get(v2); g != "hello" {
		t.Fatalf("v2 got %q, want hello", g)
	}

	// 单实例下只用一条 worker（一条 blocking 连接）
	if got := hub.Stats().Workers; got != 1 {
		t.Fatalf("Workers=%d, want 1", got)
	}
	if got := hub.Stats().Streams; got != 1 {
		t.Fatalf("Streams=%d, want 1", got)
	}
}

func TestVirtualStreamMiniredis_MultiStreamSingleSubscriber(t *testing.T) {
	// 用 hashtag 让三个 stream 落在同一 slot，共用一条 worker。
	// 这是 redisson 推荐的"减少连接数"用法（与 cluster + 同节点相同模式）。
	t.Parallel()
	c := MustNewClient(NewConf(WithT(t), WithEnableCache(false)))
	t.Cleanup(func() { _ = c.Close() })
	hub := c.NewVirtualStreamHub()
	t.Cleanup(func() { _ = hub.Close() })

	ctx := context.Background()
	streams := []string{"{grp}.m1", "{grp}.m2", "{grp}.m3"}
	v, err := hub.Subscribe(ctx, streams, "0", "0", "0")
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	t.Cleanup(func() { _ = v.Close() })

	time.Sleep(100 * time.Millisecond)
	streamXAdd(t, c, ctx, streams[0], "1")
	streamXAdd(t, c, ctx, streams[1], "2")
	streamXAdd(t, c, ctx, streams[2], "3")

	got := map[string]string{}
	deadline := time.After(2 * time.Second)
	for len(got) < 3 {
		select {
		case m := <-v.Channel():
			got[m.Stream] = m.Message.Values["v"].(string)
		case <-deadline:
			t.Fatalf("only got %d streams: %v", len(got), got)
		}
	}
	if got[streams[0]] != "1" || got[streams[1]] != "2" || got[streams[2]] != "3" {
		t.Fatalf("stream fan-in mismatch: %v", got)
	}

	// 三个 stream 用 hashtag 共享同一 slot → 同一 worker
	if w := hub.Stats().Workers; w != 1 {
		t.Fatalf("Workers=%d, want 1 (same hashtag → same slot)", w)
	}
}

func TestVirtualStreamMiniredis_DynamicAddStream(t *testing.T) {
	// 验证：订阅期间动态加入新 stream，新消息也能被正确 fan-out。
	t.Parallel()
	c := MustNewClient(NewConf(WithT(t), WithEnableCache(false)))
	t.Cleanup(func() { _ = c.Close() })
	hub := c.NewVirtualStreamHub()
	t.Cleanup(func() { _ = hub.Close() })

	ctx := context.Background()
	v, err := hub.Subscribe(ctx, []string{"s.dyn1"}, "0")
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	t.Cleanup(func() { _ = v.Close() })

	time.Sleep(100 * time.Millisecond)
	streamXAdd(t, c, ctx, "s.dyn1", "a")
	waitStreamMessages(t, v, 1, 2*time.Second)

	// 动态加 s.dyn2
	if err := v.Subscribe(ctx, []string{"s.dyn2"}, "0"); err != nil {
		t.Fatalf("dyn subscribe: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	streamXAdd(t, c, ctx, "s.dyn2", "b")

	msgs := waitStreamMessages(t, v, 1, 2*time.Second)
	if msgs[0].Stream != "s.dyn2" || msgs[0].Message.Values["v"].(string) != "b" {
		t.Fatalf("dyn add stream: got %v", msgs[0])
	}
}

func TestVirtualStreamMiniredis_UnsubscribeStopsDelivery(t *testing.T) {
	t.Parallel()
	c := MustNewClient(NewConf(WithT(t), WithEnableCache(false)))
	t.Cleanup(func() { _ = c.Close() })
	hub := c.NewVirtualStreamHub()
	t.Cleanup(func() { _ = hub.Close() })

	ctx := context.Background()
	v, err := hub.Subscribe(ctx, []string{"s.unsub"}, "0")
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	t.Cleanup(func() { _ = v.Close() })

	time.Sleep(100 * time.Millisecond)
	streamXAdd(t, c, ctx, "s.unsub", "first")
	waitStreamMessages(t, v, 1, 2*time.Second)

	// 取消订阅
	if err := v.Unsubscribe(ctx, "s.unsub"); err != nil {
		t.Fatalf("unsub: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	// 取消后写入新消息不应被收到
	streamXAdd(t, c, ctx, "s.unsub", "after")
	select {
	case m := <-v.Channel():
		t.Fatalf("got message after Unsubscribe: %+v", m)
	case <-time.After(500 * time.Millisecond):
		// OK
	}
	// 路由表应空
	if got := hub.Stats().Streams; got != 0 {
		t.Fatalf("Streams=%d after unsub, want 0", got)
	}
	if got := hub.Stats().Workers; got != 0 {
		t.Fatalf("Workers=%d after unsub, want 0 (worker should stop when empty)", got)
	}
}

func TestVirtualStreamMiniredis_HubCloseCascades(t *testing.T) {
	t.Parallel()
	c := MustNewClient(NewConf(WithT(t), WithEnableCache(false)))
	t.Cleanup(func() { _ = c.Close() })
	hub := c.NewVirtualStreamHub()

	ctx := context.Background()
	v, err := hub.Subscribe(ctx, []string{"s.cascade"}, "0")
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range v.Channel() {
		}
	}()

	time.Sleep(100 * time.Millisecond)
	if err := hub.Close(); err != nil {
		t.Fatalf("hub close: %v", err)
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatalf("Channel() did not close after Hub.Close")
	}

	// Hub 已关：再 Subscribe 必须返回 ErrVirtualStreamHubClosed
	if _, err := hub.Subscribe(ctx, []string{"x"}); err != ErrVirtualStreamHubClosed {
		t.Fatalf("Subscribe after hub.Close: got %v, want ErrVirtualStreamHubClosed", err)
	}
}

func TestVirtualStreamMiniredis_ManyStreamsOneConnection(t *testing.T) {
	// 验证大量 stream（同 slot）复用一条 blocking 连接：用 hashtag {grp} 让 N 个
	// stream 落到同一个 slot，hub Workers 应只有 1 条。
	t.Parallel()
	c := MustNewClient(NewConf(WithT(t), WithEnableCache(false)))
	t.Cleanup(func() { _ = c.Close() })
	hub := c.NewVirtualStreamHub()
	t.Cleanup(func() { _ = hub.Close() })

	ctx := context.Background()
	const N = 100
	subs := make([]VirtualStream, 0, N)
	for i := 0; i < N; i++ {
		stream := "{grp}.vs." + strconv.Itoa(i)
		v, err := hub.Subscribe(ctx, []string{stream}, "0")
		if err != nil {
			t.Fatalf("subscribe %s: %v", stream, err)
		}
		subs = append(subs, v)
	}
	st := hub.Stats()
	if st.Streams != N {
		t.Fatalf("Streams=%d, want %d", st.Streams, N)
	}
	if st.Workers != 1 {
		t.Fatalf("Workers=%d, want 1 (same hashtag should multiplex)", st.Workers)
	}
	for _, v := range subs {
		_ = v.Close()
	}
}
