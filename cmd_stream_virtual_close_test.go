package redisson

import (
	"context"
	"testing"
	"time"
)

// 这些测试不需要真实 Redis，验证虚拟 Stream Hub 的内存层逻辑：
//   - forwardMessage 不阻塞（与 PubSub 同语义）；
//   - Close 后 Channel() 自然解阻塞；
//   - Hub.Close 级联 close 所有 VirtualStream；
//   - worker 的 stream 引用计数 / 启停语义。

func helperNewBareStreamHub() *virtualStreamHub {
	conf := NewConf()
	c := &client{v: conf, handler: newBaseHandler(conf)}
	return c.NewVirtualStreamHub().(*virtualStreamHub)
}

func TestVirtualStreamForward_NonBlockingAfterClose(t *testing.T) {
	h := helperNewBareStreamHub()
	v := newVirtualStream(h)
	_ = v.Close()

	done := make(chan struct{})
	go func() {
		v.forwardMessage("s", XMessage{ID: "1-0", Values: map[string]any{"k": "v"}})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("forwardMessage blocked after Close — leak risk")
	}
}

func TestVirtualStreamForward_NonBlockingWhenBufferFullAndClosed(t *testing.T) {
	h := helperNewBareStreamHub()
	v := newVirtualStream(h)
	v.cancel() // 模拟 ctx canceled, closed=0
	time.Sleep(50 * time.Millisecond)
	for i := 0; i < h.client.v.GetPubSubChanSize(); i++ {
		select {
		case v.msgCh.In <- StreamMessage{Stream: "x"}:
		default:
			i = h.client.v.GetPubSubChanSize()
		}
	}
	done := make(chan struct{})
	go func() {
		v.forwardMessage("s", XMessage{ID: "1-0"})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("forwardMessage blocked under buffer-full + ctx-canceled")
	}
}

func TestVirtualStreamChannel_DrainAfterClose(t *testing.T) {
	h := helperNewBareStreamHub()
	v := newVirtualStream(h)
	go func() {
		time.Sleep(50 * time.Millisecond)
		_ = v.Close()
	}()
	done := make(chan struct{})
	go func() {
		for range v.Channel() {
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatalf("Channel() did not unblock after Close")
	}
}

func TestVirtualStreamHub_CloseIdempotent(t *testing.T) {
	h := helperNewBareStreamHub()
	v := newVirtualStream(h)
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

func TestVirtualStreamHub_OperationsAfterClose(t *testing.T) {
	h := helperNewBareStreamHub()
	_ = h.Close()
	ctx := context.Background()
	if _, err := h.Subscribe(ctx, []string{"s"}); err != ErrVirtualStreamHubClosed {
		t.Fatalf("Subscribe after close: want ErrVirtualStreamHubClosed, got %v", err)
	}
}

// helperNewWorkerStub 直接构造一个不启动后台 goroutine 的 streamWorker，
// 用来在没有真实 client.adapter 的纯单元测试下验证内部数据结构（routes、
// signalUpdate、dispatch 推进 lastID 等）。
//
// 与 newStreamWorker 的区别：不调用 go w.run()，因此 addStream/snapshot
// 等不会触发任何 XREAD；test 调用方只需要直接调用 dispatch / addStream /
// removeStream 等观察行为。loopDone 显式 close 让 stop() 不会卡住。
func helperNewWorkerStub(h *virtualStreamHub) *streamWorker {
	ctx, cancel := context.WithCancel(context.Background())
	loopDone := make(chan struct{})
	close(loopDone) // pretend loop already exited
	return &streamWorker{
		hub:        h,
		nodeKey:    "",
		routes:     make(map[string]*streamRoute),
		updateSig:  make(chan struct{}),
		loopCtx:    ctx,
		loopCancel: cancel,
		loopDone:   loopDone,
	}
}

// TestStreamWorker_DispatchFanout 验证 worker.dispatch 把消息扇出到多个订阅者，
// 同时推进 lastID 游标。
func TestStreamWorker_DispatchFanout(t *testing.T) {
	h := helperNewBareStreamHub()
	w := helperNewWorkerStub(h)

	v1 := newVirtualStream(h)
	v2 := newVirtualStream(h)
	defer v1.Close()
	defer v2.Close()

	// 直接构造 routes 进行模拟（绕开真实 XREAD）
	w.mu.Lock()
	w.routes["s1"] = &streamRoute{
		subs:   map[*virtualStream]struct{}{v1: {}, v2: {}},
		lastID: "$",
	}
	w.mu.Unlock()

	w.dispatch(XStream{
		Stream: "s1",
		Messages: []XMessage{
			{ID: "1-0", Values: map[string]any{"a": "1"}},
			{ID: "2-0", Values: map[string]any{"a": "2"}},
		},
	})

	// 两位订阅者各自收到 2 条消息
	read := func(v *virtualStream) []string {
		var ids []string
		deadline := time.After(500 * time.Millisecond)
		for len(ids) < 2 {
			select {
			case m := <-v.Channel():
				ids = append(ids, m.Message.ID)
			case <-deadline:
				return ids
			}
		}
		return ids
	}
	if got := read(v1); len(got) != 2 || got[0] != "1-0" || got[1] != "2-0" {
		t.Fatalf("v1 got %v, want [1-0 2-0]", got)
	}
	if got := read(v2); len(got) != 2 || got[0] != "1-0" || got[1] != "2-0" {
		t.Fatalf("v2 got %v, want [1-0 2-0]", got)
	}

	// 游标推进
	w.mu.Lock()
	if w.routes["s1"].lastID != "2-0" {
		t.Fatalf("lastID = %q, want 2-0", w.routes["s1"].lastID)
	}
	w.mu.Unlock()
}

// TestStreamWorker_RefCount 直接操作 worker 验证引用计数：
// 多 sub 订阅同一 stream → 路由唯一；部分 unsubscribe → 不删除；全部 unsubscribe → 删除并 isEmpty。
func TestStreamWorker_RefCount(t *testing.T) {
	h := helperNewBareStreamHub()
	w := helperNewWorkerStub(h)

	v1 := newVirtualStream(h)
	v2 := newVirtualStream(h)
	defer v1.Close()
	defer v2.Close()

	w.addStream(v1, "s", "$")
	w.addStream(v2, "s", "$")

	w.mu.Lock()
	if got := len(w.routes); got != 1 {
		w.mu.Unlock()
		t.Fatalf("after 2 subs same stream: routes=%d, want 1", got)
	}
	if got := len(w.routes["s"].subs); got != 2 {
		w.mu.Unlock()
		t.Fatalf("after 2 subs same stream: subs=%d, want 2", got)
	}
	w.mu.Unlock()

	if empty := w.removeStream(v1, "s"); empty {
		t.Fatalf("after one removeStream: should not be empty")
	}
	if empty := w.removeStream(v2, "s"); !empty {
		t.Fatalf("after both removeStream: should be empty")
	}
	if !w.isEmpty() {
		t.Fatalf("worker not empty after both unsubscribe")
	}
}

// TestStreamWorker_Snapshot 在路由变更时 signal 触发，并验证 snapshot 拿到当前的
// streams + ids，且 signalUpdate 会让旧 sig 立即可读。
func TestStreamWorker_SnapshotAndUpdateSignal(t *testing.T) {
	h := helperNewBareStreamHub()
	w := helperNewWorkerStub(h)

	v := newVirtualStream(h)
	defer v.Close()

	// 初始空：snapshot 应不 ok 但返回 sig。
	_, _, sig, ok := w.snapshot()
	if ok {
		t.Fatalf("empty worker snapshot should be !ok")
	}
	// 加一个 stream → sig 应被关闭（旧 sig）。
	w.addStream(v, "s1", "0")
	select {
	case <-sig:
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("addStream did not signal old updateSig")
	}

	streams, ids, _, ok := w.snapshot()
	if !ok {
		t.Fatalf("snapshot after add: !ok")
	}
	if len(streams) != 1 || streams[0] != "s1" || ids[0] != "0" {
		t.Fatalf("snapshot mismatch: streams=%v ids=%v", streams, ids)
	}
}
