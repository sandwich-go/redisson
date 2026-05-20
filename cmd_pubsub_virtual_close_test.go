package redisson

import (
	"context"
	"testing"
	"time"

	"github.com/redis/rueidis"
)

// 这些测试不需要真实 Redis，只验证虚拟 Hub 的内存层逻辑：
//   - forwardMessage 不阻塞（与 pubSub.forwardMessage 同语义）；
//   - Close 后 Channel() 自然解阻塞；
//   - Hub.Close 级联 close 所有 VirtualPubSub。

// helperNewBareHub 直接构造一个不连真实 Redis 的虚拟 Hub。
// 不会触发任何真实 SUBSCRIBE / Dedicate（这些在 hub.Subscribe 等方法中才会发生），
// 因此可以在不连接 Redis 的情况下测试 Close、forwardMessage、引用计数等逻辑。
func helperNewBareHub() *virtualHub {
	conf := NewConf()
	c := &client{v: conf, handler: newBaseHandler(conf)}
	return c.NewVirtualPubSubHub().(*virtualHub)
}

// TestVirtualPubSubForward_NonBlockingAfterClose 验证 forwardMessage 在 Close
// 后立即返回——避免 Hub 内部 dispatch 持锁扇出时阻塞，进而堵住整条 dedicated wire。
func TestVirtualPubSubForward_NonBlockingAfterClose(t *testing.T) {
	h := helperNewBareHub()
	v := newVirtualPubSub(h)
	_ = v.Close()

	done := make(chan struct{})
	go func() {
		v.forwardMessage(kindSubscribe, rueidis.PubSubMessage{Channel: "x", Message: "y"})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("forwardMessage blocked after Close — leak risk")
	}
}

// TestVirtualPubSubForward_NonBlockingWhenBufferFullAndClosed 模拟 In 缓冲打满
// + ctx 已 cancel 的极端竞态——必须仍不阻塞。
func TestVirtualPubSubForward_NonBlockingWhenBufferFullAndClosed(t *testing.T) {
	h := helperNewBareHub()
	v := newVirtualPubSub(h)
	// 直接 cancel ctx，但保持 closed=0 以走 select 分支
	v.cancel()
	time.Sleep(50 * time.Millisecond)
	for i := 0; i < h.client.v.GetPubSubChanSize(); i++ {
		select {
		case v.msgCh.In <- rueidis.PubSubMessage{Channel: "x", Message: "y"}:
		default:
			i = h.client.v.GetPubSubChanSize()
		}
	}
	done := make(chan struct{})
	go func() {
		v.forwardMessage(kindSubscribe, rueidis.PubSubMessage{Channel: "x", Message: "y"})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("forwardMessage blocked under buffer-full + ctx-canceled")
	}
}

// TestVirtualPubSubChannel_DrainAfterClose 验证 Close 后 Channel() 能被 range 退出。
func TestVirtualPubSubChannel_DrainAfterClose(t *testing.T) {
	h := helperNewBareHub()
	v := newVirtualPubSub(h)
	go func() {
		time.Sleep(50 * time.Millisecond)
		_ = v.Close()
	}()
	done := make(chan struct{})
	go func() {
		for range v.Channel() {
			// 不应有消息
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("Channel() did not unblock after Close")
	}
}

// TestVirtualHub_CloseIdempotent 验证 Hub.Close / VirtualPubSub.Close 多次调用幂等。
func TestVirtualHub_CloseIdempotent(t *testing.T) {
	h := helperNewBareHub()
	v := newVirtualPubSub(h)
	if err := v.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := v.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
	if err := h.Close(); err != nil {
		t.Fatalf("hub Close: %v", err)
	}
	if err := h.Close(); err != nil {
		t.Fatalf("hub Close (2nd): %v", err)
	}
}

// TestVirtualHub_OperationsAfterClose 验证 Hub.Close 后所有方法返回 ErrVirtualHubClosed。
func TestVirtualHub_OperationsAfterClose(t *testing.T) {
	h := helperNewBareHub()
	_ = h.Close()
	ctx := context.Background()
	if _, err := h.Subscribe(ctx, "ch"); err != ErrVirtualHubClosed {
		t.Fatalf("Subscribe after close: want ErrVirtualHubClosed, got %v", err)
	}
	if _, err := h.PSubscribe(ctx, "p*"); err != ErrVirtualHubClosed {
		t.Fatalf("PSubscribe after close: want ErrVirtualHubClosed, got %v", err)
	}
	if _, err := h.SSubscribe(ctx, "sc"); err != ErrVirtualHubClosed {
		t.Fatalf("SSubscribe after close: want ErrVirtualHubClosed, got %v", err)
	}
}

// TestSubscriptionPool_RefCount 直接操作 subscriptionPool 验证引用计数语义：
//   - 多个虚拟订阅者订阅同一 channel：routes 中只有一项；
//   - 部分 unsubscribe：refCount 仍 >0，对应的 routes 仍存在；
//   - 全部 unsubscribe：routes 中该项被移除。
//
// 这里不实际下发命令（pool.conns 为空 → getOrCreateConnLocked 会尝试 Dedicate，
// 而 client.cmd 为 nil 会 panic）。所以测试时直接构造 routes 进行模拟。
func TestSubscriptionPool_RefCount(t *testing.T) {
	h := helperNewBareHub()
	p := h.pools[kindSubscribe]
	v1 := newVirtualPubSub(h)
	v2 := newVirtualPubSub(h)
	defer v1.Close()
	defer v2.Close()

	// 直接操纵 routes（绕开 conn 创建），仿真 v1+v2 都订阅了 ch1
	p.mu.Lock()
	p.routes["ch1"] = &routeEntry{subs: map[*virtualPubSub]struct{}{v1: {}, v2: {}}}
	p.mu.Unlock()

	if got := p.routeCount(); got != 1 {
		t.Fatalf("routes after both sub: %d, want 1", got)
	}

	// 模拟 v1 unsubscribe ch1：手动调整路由
	p.mu.Lock()
	delete(p.routes["ch1"].subs, v1)
	if len(p.routes["ch1"].subs) == 0 {
		delete(p.routes, "ch1")
	}
	p.mu.Unlock()
	if got := p.routeCount(); got != 1 {
		t.Fatalf("routes after one unsub: %d, want 1 (still v2)", got)
	}

	// v2 也 unsubscribe
	p.mu.Lock()
	delete(p.routes["ch1"].subs, v2)
	if len(p.routes["ch1"].subs) == 0 {
		delete(p.routes, "ch1")
	}
	p.mu.Unlock()
	if got := p.routeCount(); got != 0 {
		t.Fatalf("routes after both unsub: %d, want 0", got)
	}
}

// TestSubscriptionPool_DispatchFanout 验证 dispatch 把单条消息扇出到多个虚拟订阅者。
func TestSubscriptionPool_DispatchFanout(t *testing.T) {
	h := helperNewBareHub()
	p := h.pools[kindSubscribe]
	v1 := newVirtualPubSub(h)
	v2 := newVirtualPubSub(h)
	defer v1.Close()
	defer v2.Close()

	p.mu.Lock()
	p.routes["ch1"] = &routeEntry{subs: map[*virtualPubSub]struct{}{v1: {}, v2: {}}}
	p.mu.Unlock()

	p.dispatch(rueidis.PubSubMessage{Channel: "ch1", Message: "hello"})

	got := func(v *virtualPubSub) string {
		select {
		case m := <-v.Channel():
			return m.Message
		case <-time.After(500 * time.Millisecond):
			return ""
		}
	}
	if g := got(v1); g != "hello" {
		t.Fatalf("v1 got %q, want hello", g)
	}
	if g := got(v2); g != "hello" {
		t.Fatalf("v2 got %q, want hello", g)
	}
}

// TestSubscriptionPool_DispatchPSubscribeUsesPattern 验证 PSubscribe 路由用 m.Pattern。
func TestSubscriptionPool_DispatchPSubscribeUsesPattern(t *testing.T) {
	h := helperNewBareHub()
	p := h.pools[kindPSubscribe]
	v := newVirtualPubSub(h)
	defer v.Close()
	p.mu.Lock()
	p.routes["news.*"] = &routeEntry{subs: map[*virtualPubSub]struct{}{v: {}}}
	p.mu.Unlock()

	// pmessage 中 Pattern=news.*, Channel=news.foo
	p.dispatch(rueidis.PubSubMessage{Pattern: "news.*", Channel: "news.foo", Message: "x"})

	select {
	case m := <-v.Channel():
		if m.Pattern != "news.*" || m.Channel != "news.foo" || m.Message != "x" {
			t.Fatalf("dispatch wrong msg: %+v", m)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("dispatch did not deliver pmessage")
	}
}
