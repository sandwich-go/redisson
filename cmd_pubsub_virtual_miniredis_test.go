//go:build miniredis_test

package redisson

import (
	"context"
	"testing"
	"time"
)

// 这套测试在 miniredis 中验证虚拟 PubSub 端到端可用，
// 与 cmd_pubsub_virtual_test.go (integration) 形成互补：本套不需要真实 Redis。
//
// 仅需要 build tags: miniredis_test redisson_miniredis。

func TestVirtualPubSubMiniredis_BasicFanout(t *testing.T) {
	t.Parallel()
	c := MustNewClient(NewConf(WithT(t), WithEnableCache(false)))
	t.Cleanup(func() { _ = c.Close() })
	hub := c.NewVirtualPubSubHub()
	t.Cleanup(func() { _ = hub.Close() })

	ctx := context.Background()
	v1, err := hub.Subscribe(ctx, "ch.basic")
	if err != nil {
		t.Fatalf("v1 subscribe: %v", err)
	}
	v2, err := hub.Subscribe(ctx, "ch.basic")
	if err != nil {
		t.Fatalf("v2 subscribe: %v", err)
	}
	t.Cleanup(func() { _ = v1.Close() })
	t.Cleanup(func() { _ = v2.Close() })

	// 等订阅生效（miniredis 几乎是同步的，但保险起见 sleep 一下）
	time.Sleep(100 * time.Millisecond)

	st := hub.Stats()
	if st.Channels != 1 {
		t.Fatalf("Stats.Channels = %d, want 1 (multiplexed)", st.Channels)
	}

	if r := c.Publish(ctx, "ch.basic", "hello"); r.Err() != nil {
		t.Fatalf("publish: %v", r.Err())
	}

	recv := func(v VirtualPubSub) string {
		select {
		case m := <-v.Channel():
			return m.Message
		case <-time.After(2 * time.Second):
			return ""
		}
	}
	if g := recv(v1); g != "hello" {
		t.Fatalf("v1 recv %q, want hello", g)
	}
	if g := recv(v2); g != "hello" {
		t.Fatalf("v2 recv %q, want hello", g)
	}
}

func TestVirtualPubSubMiniredis_RefCount(t *testing.T) {
	t.Parallel()
	c := MustNewClient(NewConf(WithT(t), WithEnableCache(false)))
	t.Cleanup(func() { _ = c.Close() })
	hub := c.NewVirtualPubSubHub()
	t.Cleanup(func() { _ = hub.Close() })

	ctx := context.Background()
	v1, _ := hub.Subscribe(ctx, "ch.rc")
	v2, _ := hub.Subscribe(ctx, "ch.rc")

	if got := hub.Stats().Channels; got != 1 {
		t.Fatalf("two subs to same channel → Stats.Channels=%d, want 1", got)
	}
	_ = v1.Unsubscribe(ctx, "ch.rc")
	if got := hub.Stats().Channels; got != 1 {
		t.Fatalf("after v1 unsub, v2 still holds → Stats.Channels=%d, want 1", got)
	}
	_ = v2.Unsubscribe(ctx, "ch.rc")
	// unsubscribe 是异步通过命令下发的，给 miniredis 一点时间消化
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if hub.Stats().Channels == 0 {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if got := hub.Stats().Channels; got != 0 {
		t.Fatalf("after both unsub, Stats.Channels=%d, want 0", got)
	}
	_ = v1.Close()
	_ = v2.Close()
}

func TestVirtualPubSubMiniredis_PSubscribe(t *testing.T) {
	t.Parallel()
	c := MustNewClient(NewConf(WithT(t), WithEnableCache(false)))
	t.Cleanup(func() { _ = c.Close() })
	hub := c.NewVirtualPubSubHub()
	t.Cleanup(func() { _ = hub.Close() })

	ctx := context.Background()
	v, err := hub.PSubscribe(ctx, "news.*")
	if err != nil {
		t.Fatalf("psubscribe: %v", err)
	}
	t.Cleanup(func() { _ = v.Close() })

	time.Sleep(100 * time.Millisecond)
	if r := c.Publish(ctx, "news.weather", "rain"); r.Err() != nil {
		t.Fatalf("publish: %v", r.Err())
	}
	select {
	case m := <-v.Channel():
		if m.Pattern != "news.*" || m.Channel != "news.weather" || m.Message != "rain" {
			t.Fatalf("psubscribe wrong msg: %+v", m)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("psubscribe no msg")
	}
}

func TestVirtualPubSubMiniredis_HubCloseCascades(t *testing.T) {
	t.Parallel()
	c := MustNewClient(NewConf(WithT(t), WithEnableCache(false)))
	t.Cleanup(func() { _ = c.Close() })
	hub := c.NewVirtualPubSubHub()

	ctx := context.Background()
	v, err := hub.Subscribe(ctx, "ch.closed")
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	done := make(chan struct{})
	go func() {
		for range v.Channel() {
		}
		close(done)
	}()
	time.Sleep(100 * time.Millisecond)
	_ = hub.Close()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("Channel() not closed after Hub.Close")
	}

	// Hub 已关：再 Subscribe 必须返回 ErrVirtualHubClosed
	if _, err := hub.Subscribe(ctx, "x"); err != ErrVirtualHubClosed {
		t.Fatalf("Subscribe after hub.Close: got %v, want ErrVirtualHubClosed", err)
	}
}

func TestVirtualPubSubMiniredis_MultiChannelOneSubscriber(t *testing.T) {
	// 同一 VirtualPubSub 订阅多个 channel，所有消息都汇聚到同一个 Channel()。
	t.Parallel()
	c := MustNewClient(NewConf(WithT(t), WithEnableCache(false)))
	t.Cleanup(func() { _ = c.Close() })
	hub := c.NewVirtualPubSubHub()
	t.Cleanup(func() { _ = hub.Close() })

	ctx := context.Background()
	v, err := hub.Subscribe(ctx, "ch.a", "ch.b", "ch.c")
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	t.Cleanup(func() { _ = v.Close() })

	time.Sleep(100 * time.Millisecond)
	c.Publish(ctx, "ch.a", "1")
	c.Publish(ctx, "ch.b", "2")
	c.Publish(ctx, "ch.c", "3")

	got := map[string]string{}
	for i := 0; i < 3; i++ {
		select {
		case m := <-v.Channel():
			got[m.Channel] = m.Message
		case <-time.After(2 * time.Second):
			t.Fatalf("only got %d messages: %v", i, got)
		}
	}
	if len(got) != 3 || got["ch.a"] != "1" || got["ch.b"] != "2" || got["ch.c"] != "3" {
		t.Fatalf("multi-channel fan-in mismatch: %v", got)
	}
}
