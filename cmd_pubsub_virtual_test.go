//go:build integration

package redisson

import (
	"context"
	"sync"
	"testing"
	"time"
)

// 这一套 integration 测试验证虚拟 PubSub 在真实 Redis 上的端到端语义：
//   - 多个虚拟订阅者共享同一条底层连接
//   - 引用计数：相同 channel 第一次 SUBSCRIBE，归零时 UNSUBSCRIBE
//   - 多 channel 在同一个 VirtualPubSub 上扇入
//   - PSubscribe / SSubscribe 一致语义
//   - Hub.Close 优雅级联关闭

func waitForVirtualMessage(t *testing.T, v VirtualPubSub, want string, timeout time.Duration) Message {
	t.Helper()
	select {
	case m := <-v.Channel():
		if m.Message != want {
			t.Fatalf("got message %q, want %q", m.Message, want)
		}
		return m
	case <-time.After(timeout):
		t.Fatalf("timeout waiting for message %q", want)
	}
	return Message{}
}

func TestVirtualPubSubIntegration_BasicFanout(t *testing.T) {
	c := MustNewClient(NewConf(WithEnableCache(false)))
	defer c.Close()
	hub := c.NewVirtualPubSubHub()
	defer hub.Close()

	ctx := context.Background()
	v1, err := hub.Subscribe(ctx, "ch.basic")
	if err != nil {
		t.Fatalf("v1 subscribe: %v", err)
	}
	v2, err := hub.Subscribe(ctx, "ch.basic")
	if err != nil {
		t.Fatalf("v2 subscribe: %v", err)
	}
	defer v1.Close()
	defer v2.Close()

	// 真实连接数：Subscribe 类型仅 1 条（共享）
	st := hub.Stats()
	if st.Channels != 1 || st.Connections < 1 {
		t.Fatalf("stats unexpected: %+v", st)
	}

	// 等待 SUBSCRIBE 真正生效（PUBSUB NUMSUB == 1，因为 Hub 内部只发了一条 SUBSCRIBE）
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if c.PubSubNumSub(ctx, "ch.basic").Val()["ch.basic"] >= 1 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if got := c.PubSubNumSub(ctx, "ch.basic").Val()["ch.basic"]; got != 1 {
		t.Fatalf("PUBSUB NUMSUB ch.basic = %d, want 1 (hub should multiplex)", got)
	}

	// 发布一条，两个虚拟订阅者都应该收到
	if r := c.Publish(ctx, "ch.basic", "hello"); r.Err() != nil {
		t.Fatalf("publish: %v", r.Err())
	}
	waitForVirtualMessage(t, v1, "hello", 2*time.Second)
	waitForVirtualMessage(t, v2, "hello", 2*time.Second)
}

func TestVirtualPubSubIntegration_RefCountUnsubscribe(t *testing.T) {
	c := MustNewClient(NewConf(WithEnableCache(false)))
	defer c.Close()
	hub := c.NewVirtualPubSubHub()
	defer hub.Close()

	ctx := context.Background()
	v1, err := hub.Subscribe(ctx, "ch.ref")
	if err != nil {
		t.Fatalf("v1 subscribe: %v", err)
	}
	v2, err := hub.Subscribe(ctx, "ch.ref")
	if err != nil {
		t.Fatalf("v2 subscribe: %v", err)
	}

	// 等订阅生效
	time.Sleep(100 * time.Millisecond)

	// v1 unsubscribe：底层应仍保持订阅（v2 还在）
	if err := v1.Unsubscribe(ctx, "ch.ref"); err != nil {
		t.Fatalf("v1 unsub: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	if got := c.PubSubNumSub(ctx, "ch.ref").Val()["ch.ref"]; got != 1 {
		t.Fatalf("after v1 unsub: PUBSUB NUMSUB = %d, want 1", got)
	}

	// v2 unsubscribe：refCount 归零，底层 UNSUBSCRIBE
	if err := v2.Unsubscribe(ctx, "ch.ref"); err != nil {
		t.Fatalf("v2 unsub: %v", err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if c.PubSubNumSub(ctx, "ch.ref").Val()["ch.ref"] == 0 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if got := c.PubSubNumSub(ctx, "ch.ref").Val()["ch.ref"]; got != 0 {
		t.Fatalf("after both unsub: PUBSUB NUMSUB = %d, want 0", got)
	}
	_ = v1.Close()
	_ = v2.Close()
}

func TestVirtualPubSubIntegration_PSubscribe(t *testing.T) {
	c := MustNewClient(NewConf(WithEnableCache(false)))
	defer c.Close()
	hub := c.NewVirtualPubSubHub()
	defer hub.Close()

	ctx := context.Background()
	v, err := hub.PSubscribe(ctx, "news.*")
	if err != nil {
		t.Fatalf("psubscribe: %v", err)
	}
	defer v.Close()

	time.Sleep(100 * time.Millisecond)
	if r := c.Publish(ctx, "news.weather", "rain"); r.Err() != nil {
		t.Fatalf("publish: %v", r.Err())
	}
	m := waitForVirtualMessage(t, v, "rain", 2*time.Second)
	if m.Pattern != "news.*" {
		t.Fatalf("Pattern=%q, want news.*", m.Pattern)
	}
}

func TestVirtualPubSubIntegration_HubClose_CascadeShutdown(t *testing.T) {
	c := MustNewClient(NewConf(WithEnableCache(false)))
	defer c.Close()
	hub := c.NewVirtualPubSubHub()

	ctx := context.Background()
	v1, _ := hub.Subscribe(ctx, "ch.close.1")
	v2, _ := hub.PSubscribe(ctx, "ch.close.*")

	// Channel() 必须在 Hub.Close 后被关闭，下游 range 正常退出。
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for range v1.Channel() {
		}
	}()
	go func() {
		defer wg.Done()
		for range v2.Channel() {
		}
	}()

	time.Sleep(100 * time.Millisecond)
	if err := hub.Close(); err != nil {
		t.Fatalf("hub.Close: %v", err)
	}

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatalf("VirtualPubSub.Channel() did not close after Hub.Close")
	}
}

func TestVirtualPubSubIntegration_ManyChannelsOneConnection(t *testing.T) {
	// 验证大量 channel 复用单条 dedicated 连接：1000 个 channel 起 1000 个虚拟订阅者，
	// 全部成功且 Hub Stats.Connections 仍 ≤ 1（非 cluster）或 ≤ 节点数（cluster）。
	c := MustNewClient(NewConf(WithEnableCache(false)))
	defer c.Close()
	hub := c.NewVirtualPubSubHub()
	defer hub.Close()

	ctx := context.Background()
	const N = 1000
	subs := make([]VirtualPubSub, 0, N)
	for i := 0; i < N; i++ {
		ch := "vch." + itoa(i)
		v, err := hub.Subscribe(ctx, ch)
		if err != nil {
			t.Fatalf("subscribe %s: %v", ch, err)
		}
		subs = append(subs, v)
	}
	st := hub.Stats()
	if st.Channels != N {
		t.Fatalf("Stats.Channels = %d, want %d", st.Channels, N)
	}
	// 单实例下一条连接搞定 1000 channel
	if !c.IsCluster() && st.Connections > 1 {
		t.Fatalf("Stats.Connections = %d, want <= 1 on single instance", st.Connections)
	}
	for _, v := range subs {
		_ = v.Close()
	}
}

// itoa 不使用 strconv 以避免改 import（测试文件而已，这里直接复用）。
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b [20]byte
	pos := len(b)
	for i > 0 {
		pos--
		b[pos] = byte('0' + i%10)
		i /= 10
	}
	return string(b[pos:])
}
