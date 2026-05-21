package redisson

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// 虚拟 Stream（VirtualStream / VirtualStreamHub）解决的问题：
//
// Redis Stream 的"持续监听"语义通常通过 XREAD BLOCK 0 实现，rueidis 在
// blocking 命令上会从独立的 BlockingPool 取一条连接独占阻塞，直到：
//   1) 服务端有数据返回；
//   2) ctx 取消；
//   3) 服务端断连。
//
// 当业务侧对成百上千个 stream 各自起一个 goroutine 跑 XREAD BLOCK 0 时：
//   - 每个 goroutine 独占一条 blocking 连接；
//   - rueidis BlockingPool 默认 1024 上限会被打爆；
//   - Redis Server 的 maxclients 上限也会被压垮。
//
// 与 PubSub 不同，rueidis 没有给 stream 提供 SetPubSubHooks 那样的"server push
// 异步回调"接口，所以本实现只能采用经典的"客户端轮询 + 多 stream 合并读"思路：
//
//   一个 hub 内部 ** 一组 worker goroutine **（每个负责一组 stream），每个 worker
//   持续做 XREAD BLOCK 0 COUNT N STREAMS k1 k2 ... id1 id2 ...，把多 stream 合并到
//   一次阻塞读里；返回后按 stream 名查路由表扇出到各 virtualStream，并推进每个
//   stream 的 lastID 游标。
//
// 连接数收益（XREAD 是 keyed 命令，按 streams 列表里第一个 key 的 slot 路由）：
//   - 单实例 / 主从：1 个 hub 永远只占用 ** 1 条 ** blocking 连接；
//   - Cluster：worker 数 = 不同 slot 的数量，每个有订阅的 slot owner 节点 1 条
//     blocking 连接。
//
// 控制连接数的方法（cluster 下尤其重要）：
//   * ** XREAD 是 keyed 命令，hashtag 有效 **：给相关 stream 名带相同 {tag}
//     可让它们落到同一 slot，从而共用一条 blocking 连接。例如：
//         hub.Subscribe(ctx, []string{"{events}.a", "{events}.b", "{events}.c"})
//     上面 3 个 stream 名因 hashtag 相同 → CRC16 相同 → slot 相同 → 1 条连接。
//   * 反例：若用 "events.a"、"events.b"、"events.c" 这种 slot 各异的命名，
//     hub 会分别为每个 slot 起 1 个 worker（即 1 条 blocking 连接），同时
//     rueidis 在 cluster builder 模式下还会拒绝跨 slot 多 key 命令，因此本
//     实现强制按 slot 分组（详见 nodeKeyForStream 的注释）。
//   * 即使后端只是单实例，rueidis 默认仍构造 *clusterClient，跨 slot 仍受
//     builder 校验，所以"按 slot 分组"在所有部署形态下都生效。
//
// 使用约束（fan-out 语义）：
//   - 多个 VirtualStream 订阅同一个 stream 时，每条 message 都会被扇出给所有订阅者
//     （与 PubSub 一致）。这是消息流复制语义，不是 XREADGROUP 的消费者组互斥语义。
//     如果需要消费者组语义请直接调用 client.XReadGroup（不要走本 hub）。
//   - 默认起始 ID 为 "$"（仅读订阅之后到达的新消息），与 XREAD ... STREAMS s $ 一致；
//     用户可在 hub.Subscribe(stream, startID) 时显式传入历史 ID（如 "0" 全量重放）。
//
// 关键不变量：
//   - 每个 stream 在 hub 内只有一个游标（lastID）。多订阅者共享读取的物理 cursor，
//     不会让 server 多发同一条消息。fan-out 是"读到一条消息后内存扇出"。
//   - 路由变更（add/remove stream）通过取消当前 worker ctx 让 XREAD 立即返回，
//     再用新的 stream 列表重发；这是唯一可行的"动态增删"机制，因为 XREAD 不像
//     SUBSCRIBE 那样支持在同一连接上追加。

// VirtualStreamHub 是 stream 多路复用器。
// 通过 client.NewVirtualStreamHub() 创建，可生成多个 VirtualStream 实例，
// 这些实例共享 Hub 内部的少量 blocking 连接。
type VirtualStreamHub interface {
	// Subscribe 创建一个虚拟 Stream 订阅者。
	// 可后续在返回的 VirtualStream 上继续 Subscribe / Unsubscribe 增删 stream。
	// startIDs 与 streams 一一对应，可省略：
	//   - len(startIDs) == 0：所有 stream 起点为 "$"
	//   - len(startIDs) == len(streams)：按位对齐
	Subscribe(ctx context.Context, streams []string, startIDs ...string) (VirtualStream, error)

	// Close 关闭整个 Hub：停止所有 worker、关闭所有 VirtualStream 的 Channel()。
	Close() error

	// Stats 返回当前 Hub 的统计信息。
	Stats() VirtualStreamStats
}

// VirtualStream 暴露给用户的虚拟 stream 订阅者。
// 行为类似订阅器：业务侧 range Channel() 拿 XStreamMessage 即可。
type VirtualStream interface {
	// Subscribe 增加订阅。startIDs 含义与 Hub.Subscribe 一致。
	// 注意：若该 stream 已被 Hub 内其它 VirtualStream 订阅过，本订阅者只会收到
	// **从此刻起**（hub 内部当前游标之后）到达的消息；startID 仅在 stream 为
	// hub 内首次出现时生效。
	Subscribe(ctx context.Context, streams []string, startIDs ...string) error

	// Unsubscribe 移除订阅。被全部订阅者解除后，hub 不再读取该 stream。
	Unsubscribe(ctx context.Context, streams ...string) error

	// Channel 返回扇入的消息 channel；Close 后会被关闭，下游 range 自然退出。
	Channel() <-chan StreamMessage

	// Close 释放所有持有的 stream 订阅。幂等。
	Close() error
}

// StreamMessage 是 Hub 派发给 VirtualStream 的事件单元——单条消息粒度。
// 不直接复用 XStream 是为了让用户在收到一条消息时立即处理（而不是攒一批）。
type StreamMessage struct {
	// Stream 是消息所属 stream 的 key。
	Stream string
	// Message 是消息体（含 ID 与 Values）。
	Message XMessage
}

// VirtualStreamStats 统计信息。
type VirtualStreamStats struct {
	// Subscribers 当前活跃 VirtualStream 数。
	Subscribers int
	// Streams Hub 真实订阅的 stream 数（去重后）。
	Streams int
	// Workers Hub 持有的 worker goroutine / blocking 连接数。
	// 单实例 = 1（仅在有订阅时启动）；Cluster = 有订阅的 slot 节点数。
	Workers int
}

// 默认参数。可通过 ConfVisitor 暴露后再扩展（暂保留为常量，足够 99% 场景）。
const (
	// streamWorkerBlock 是 hub worker 一次 XREAD 的 BLOCK 时长。
	// 用 5s 而不是 0（永久阻塞）的原因：让 worker 周期性回到 select，能处理路由
	// 变更信号；即便 cancel 信号路径正常，5s 也是健壮性 fallback。
	streamWorkerBlock = 5 * time.Second
	// streamWorkerCount 是一次 XREAD 的 COUNT 上限。值偏小让回路更敏捷。
	streamWorkerCount = 100
	// streamReconnectBackoff 是 worker 出错时的重试退避。
	streamReconnectBackoff = 200 * time.Millisecond
)

// ErrVirtualStreamHubClosed 在 Hub 已 Close 后操作返回。
var ErrVirtualStreamHubClosed = errors.New("virtual stream hub closed")

// virtualStreamHub 是 VirtualStreamHub 的实现。
type virtualStreamHub struct {
	client *client

	mu      sync.Mutex
	subs    map[*virtualStream]struct{}
	workers map[string]*streamWorker // key: 节点分组（单实例为 ""，cluster 为 slotKey）
	closed  AtomicInt32
}

// NewVirtualStreamHub 在 client 上创建一个 Stream Hub。
// 多次调用返回相互独立的 Hub。
func (c *client) NewVirtualStreamHub() VirtualStreamHub {
	return &virtualStreamHub{
		client:  c,
		subs:    make(map[*virtualStream]struct{}),
		workers: make(map[string]*streamWorker),
	}
}

func (h *virtualStreamHub) isClosed() bool { return h.closed.Get() == 1 }

func (h *virtualStreamHub) Subscribe(ctx context.Context, streams []string, startIDs ...string) (VirtualStream, error) {
	if h.isClosed() {
		return nil, ErrVirtualStreamHubClosed
	}
	v := newVirtualStream(h)
	h.mu.Lock()
	if h.isClosed() {
		h.mu.Unlock()
		_ = v.closeFromHub()
		return nil, ErrVirtualStreamHubClosed
	}
	h.subs[v] = struct{}{}
	h.mu.Unlock()

	if len(streams) > 0 {
		if err := v.Subscribe(ctx, streams, startIDs...); err != nil {
			_ = v.Close()
			return nil, err
		}
	}
	return v, nil
}

func (h *virtualStreamHub) Close() error {
	if !h.closed.CompareAndSwap(0, 1) {
		return nil
	}
	h.mu.Lock()
	subs := h.subs
	workers := h.workers
	h.subs = nil
	h.workers = nil
	h.mu.Unlock()

	// 先停 worker（停止从 redis 拉消息），再关 subscriber（关闭其 Channel）。
	for _, w := range workers {
		w.stop()
	}
	for v := range subs {
		_ = v.closeFromHub()
	}
	return nil
}

func (h *virtualStreamHub) Stats() VirtualStreamStats {
	st := VirtualStreamStats{}
	h.mu.Lock()
	defer h.mu.Unlock()
	st.Subscribers = len(h.subs)
	st.Workers = len(h.workers)
	streams := map[string]struct{}{}
	for _, w := range h.workers {
		w.mu.Lock()
		for k := range w.routes {
			streams[k] = struct{}{}
		}
		w.mu.Unlock()
	}
	st.Streams = len(streams)
	return st
}

// removeSubscriber 用于 v.Close 路径。
func (h *virtualStreamHub) removeSubscriber(v *virtualStream) {
	h.mu.Lock()
	delete(h.subs, v)
	h.mu.Unlock()
}

// nodeKeyForStream 决定 stream 应当落到哪个 worker。
//
// 永远按 slot 分组——即使在单实例下也必须分组：rueidis 默认构造的客户端是
// *rueidis.clusterClient（就算后端只有 1 个 redis 实例），其 Builder 在
// InitSlot 模式下会对多 key 命令校验"同 slot"，跨 slot 直接 panic。所以
// 我们把同一 slot 的 stream 合并到一个 worker，跨 slot 的 stream 各自一个
// worker —— 这样既兼容单实例（实际 rueidis 不路由），也兼容 cluster
// （rueidis 按第一个 key 的 slot 选 owner 节点）。
//
// 注意：跨 slot 的 stream 会增加 worker 数（→ blocking 连接数）；如果业务侧
// 需要把多个相关 stream 合并到一条连接，请用 Redis 的 hashtag（{tag}）让它们
// 落到同一个 slot。
func (h *virtualStreamHub) nodeKeyForStream(stream string) string {
	return clusterSlotKey(stream)
}

// addStream 在指定 worker 分组上注册一个 stream → subscriber 关系，
// 必要时启动 worker。startID 仅在该 stream 是 hub 内首次出现时生效。
func (h *virtualStreamHub) addStream(v *virtualStream, stream, startID string) error {
	nodeKey := h.nodeKeyForStream(stream)

	h.mu.Lock()
	if h.isClosed() {
		h.mu.Unlock()
		return ErrVirtualStreamHubClosed
	}
	w, ok := h.workers[nodeKey]
	if !ok {
		w = newStreamWorker(h, nodeKey)
		h.workers[nodeKey] = w
	}
	h.mu.Unlock()

	w.addStream(v, stream, startID)
	return nil
}

// removeStream 解除 stream → subscriber，归零时通知 worker 停止读该 stream，
// worker 路由表为空时停止 worker 并从 hub 摘除。
func (h *virtualStreamHub) removeStream(v *virtualStream, stream string) {
	nodeKey := h.nodeKeyForStream(stream)
	h.mu.Lock()
	w := h.workers[nodeKey]
	h.mu.Unlock()
	if w == nil {
		return
	}
	empty := w.removeStream(v, stream)
	if empty {
		// worker 已无任何 stream，可回收
		h.mu.Lock()
		// 二次确认：可能在锁外又被加入了新 stream
		if cur := h.workers[nodeKey]; cur == w && cur.isEmpty() {
			delete(h.workers, nodeKey)
			h.mu.Unlock()
			w.stop()
			return
		}
		h.mu.Unlock()
	}
}

// ============================================================================
// virtualStream
// ============================================================================

type virtualStream struct {
	hub   *virtualStreamHub
	msgCh *unboundedChan[StreamMessage]

	ctx    context.Context
	cancel context.CancelFunc
	closed AtomicInt32

	mu       sync.Mutex
	holding  map[string]struct{} // 本 sub 持有的 stream 集合
}

func newVirtualStream(h *virtualStreamHub) *virtualStream {
	ctx, cancel := context.WithCancel(context.Background())
	return &virtualStream{
		hub:     h,
		msgCh:   newUnboundedChan[StreamMessage](ctx, h.client.v.GetPubSubChanSize()),
		ctx:     ctx,
		cancel:  cancel,
		holding: make(map[string]struct{}),
	}
}

func (v *virtualStream) isClosed() bool { return v.closed.Get() == 1 }

func (v *virtualStream) Channel() <-chan StreamMessage { return v.msgCh.Out }

// forwardMessage 由 worker 调用扇出消息；与 PubSub 的同名函数同形——必须不阻塞。
func (v *virtualStream) forwardMessage(stream string, m XMessage) {
	if v.isClosed() {
		return
	}
	select {
	case v.msgCh.In <- StreamMessage{Stream: stream, Message: m}:
	case <-v.ctx.Done():
		warning(fmt.Sprintf("virtualstream: dropped after close, stream=%s id=%s", stream, m.ID))
	}
}

func (v *virtualStream) Subscribe(ctx context.Context, streams []string, startIDs ...string) error {
	if v.isClosed() || v.hub.isClosed() {
		return ErrVirtualStreamHubClosed
	}
	if len(streams) == 0 {
		return nil
	}
	if len(startIDs) != 0 && len(startIDs) != len(streams) {
		return fmt.Errorf("virtualstream: startIDs length %d must be 0 or equal to streams length %d",
			len(startIDs), len(streams))
	}
	ctx = v.hub.client.handler.before(ctx, CommandXRead)
	var firstErr error
	for i, s := range streams {
		startID := "$"
		if len(startIDs) > 0 {
			startID = startIDs[i]
		}
		if err := v.hub.addStream(v, s, startID); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		v.mu.Lock()
		v.holding[s] = struct{}{}
		v.mu.Unlock()
	}
	v.hub.client.handler.after(ctx, firstErr)
	return firstErr
}

func (v *virtualStream) Unsubscribe(ctx context.Context, streams ...string) error {
	if v.isClosed() {
		return nil
	}
	if len(streams) == 0 {
		return nil
	}
	ctx = v.hub.client.handler.before(ctx, CommandXRead)
	for _, s := range streams {
		v.hub.removeStream(v, s)
		v.mu.Lock()
		delete(v.holding, s)
		v.mu.Unlock()
	}
	v.hub.client.handler.after(ctx, nil)
	return nil
}

func (v *virtualStream) Close() error {
	if !v.closed.CompareAndSwap(0, 1) {
		return nil
	}
	v.hub.removeSubscriber(v)
	v.unsubscribeAll()
	v.cancel()
	return nil
}

func (v *virtualStream) closeFromHub() error {
	if !v.closed.CompareAndSwap(0, 1) {
		return nil
	}
	v.unsubscribeAll()
	v.cancel()
	return nil
}

func (v *virtualStream) unsubscribeAll() {
	v.mu.Lock()
	streams := make([]string, 0, len(v.holding))
	for s := range v.holding {
		streams = append(streams, s)
	}
	v.holding = make(map[string]struct{})
	v.mu.Unlock()
	for _, s := range streams {
		v.hub.removeStream(v, s)
	}
}

// ============================================================================
// streamWorker: 一组 stream 的后台 XREAD 轮询者
// ============================================================================

// streamWorker 维护一组 stream（同一 cluster slot/节点）的路由表，
// 并跑一个后台 goroutine 持续 XREAD BLOCK 这一组。
type streamWorker struct {
	hub     *virtualStreamHub
	nodeKey string

	mu     sync.Mutex
	routes map[string]*streamRoute // key: stream name
	closed bool

	// updateSig 在 routes 增删时被关闭→重建，让阻塞中的 XREAD ctx 取消、worker
	// 立即用新 routes 重发。用 chan struct{} + close 而非 sync.Cond 是为了能与
	// ctx.Done() 在 select 中并存。
	updateSig chan struct{}

	// workerCancel 触发当前阻塞 XREAD 立即返回（用于路由变更或 stop）。
	workerCancel context.CancelFunc

	// loopCtx 控制 worker goroutine 主循环；stop 时 cancel。
	loopCtx    context.Context
	loopCancel context.CancelFunc
	loopDone   chan struct{}
}

type streamRoute struct {
	subs   map[*virtualStream]struct{}
	lastID string
}

func newStreamWorker(h *virtualStreamHub, nodeKey string) *streamWorker {
	ctx, cancel := context.WithCancel(context.Background())
	w := &streamWorker{
		hub:        h,
		nodeKey:    nodeKey,
		routes:     make(map[string]*streamRoute),
		updateSig:  make(chan struct{}),
		loopCtx:    ctx,
		loopCancel: cancel,
		loopDone:   make(chan struct{}),
	}
	go w.run()
	return w
}

func (w *streamWorker) isEmpty() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.routes) == 0
}

// signalUpdate 让 worker 主循环退出当前 XREAD 等待并使用新路由重发。
// 必须在持锁状态下调用，避免与 run 循环新建 updateSig 的 race。
func (w *streamWorker) signalUpdateLocked() {
	close(w.updateSig)
	w.updateSig = make(chan struct{})
	if w.workerCancel != nil {
		w.workerCancel()
		w.workerCancel = nil
	}
}

func (w *streamWorker) addStream(v *virtualStream, stream, startID string) {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return
	}
	r, ok := w.routes[stream]
	changed := false
	if !ok {
		r = &streamRoute{subs: map[*virtualStream]struct{}{}, lastID: startID}
		w.routes[stream] = r
		changed = true
	}
	r.subs[v] = struct{}{}
	if changed {
		w.signalUpdateLocked()
	}
	w.mu.Unlock()
}

// removeStream 从 worker 路由表移除一个 (stream, subscriber) 关系。
// 返回 true 表示 worker 已无任何路由（调用方可决定是否回收 worker）。
func (w *streamWorker) removeStream(v *virtualStream, stream string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return true
	}
	r, ok := w.routes[stream]
	if !ok {
		return len(w.routes) == 0
	}
	delete(r.subs, v)
	if len(r.subs) == 0 {
		delete(w.routes, stream)
		w.signalUpdateLocked()
	}
	return len(w.routes) == 0
}

func (w *streamWorker) stop() {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return
	}
	w.closed = true
	if w.workerCancel != nil {
		w.workerCancel()
	}
	w.loopCancel()
	w.mu.Unlock()
	<-w.loopDone
}

// snapshot 在锁内拷贝一份当前 routes 给 worker 用作下一轮 XREAD 输入。
func (w *streamWorker) snapshot() (streams []string, ids []string, sig <-chan struct{}, ok bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed || len(w.routes) == 0 {
		return nil, nil, w.updateSig, false
	}
	streams = make([]string, 0, len(w.routes))
	ids = make([]string, 0, len(w.routes))
	for k, r := range w.routes {
		streams = append(streams, k)
		ids = append(ids, r.lastID)
	}
	return streams, ids, w.updateSig, true
}

// updateLastID 在收到一批消息后推进 hub 内的全局游标。
// dispatch 阶段也在锁内进行，保证 routes 删除与 forwardMessage 之间不竞争——
// 实际把 forwardMessage 移到锁外（避免被慢消费者拖慢路由 mutate）。
func (w *streamWorker) dispatch(xs XStream) {
	if len(xs.Messages) == 0 {
		return
	}
	w.mu.Lock()
	r, ok := w.routes[xs.Stream]
	if !ok {
		w.mu.Unlock()
		return
	}
	subs := make([]*virtualStream, 0, len(r.subs))
	for s := range r.subs {
		subs = append(subs, s)
	}
	// 推进游标到最后一条消息的 ID
	r.lastID = xs.Messages[len(xs.Messages)-1].ID
	w.mu.Unlock()
	for _, m := range xs.Messages {
		for _, s := range subs {
			s.forwardMessage(xs.Stream, m)
		}
	}
}

// run 是 worker 主循环：拿 routes 快照 → 发起 XREAD BLOCK → 推进 lastID → 扇出 → 重来。
// 退出条件：loopCtx done（hub 关闭或 worker stop）。
func (w *streamWorker) run() {
	defer close(w.loopDone)
	for {
		// 优先检查 loop ctx
		select {
		case <-w.loopCtx.Done():
			return
		default:
		}

		streams, ids, sig, ok := w.snapshot()
		if !ok {
			// 路由表为空：等待更新或退出
			select {
			case <-w.loopCtx.Done():
				return
			case <-sig:
				continue
			}
		}

		// 给本次 XREAD 准备一个独立 ctx，可被 signalUpdate / stop 取消。
		readCtx, readCancel := context.WithCancel(w.loopCtx)
		w.mu.Lock()
		if w.closed {
			w.mu.Unlock()
			readCancel()
			return
		}
		w.workerCancel = readCancel
		w.mu.Unlock()

		args := XReadArgs{
			Streams: append(append([]string{}, streams...), ids...),
			Count:   streamWorkerCount,
			Block:   streamWorkerBlock,
		}
		result := w.hub.client.adapter.XRead(readCtx, args)
		readCancel()
		w.mu.Lock()
		w.workerCancel = nil
		closed := w.closed
		w.mu.Unlock()
		if closed {
			return
		}

		if err := result.Err(); err != nil {
			// 三类预期错误：
			//   - context canceled（路由变更触发的主动取消）：不报错，直接重发；
			//   - rueidis Nil（XREAD BLOCK 超时无数据）：直接进入下一轮；
			//   - 其他错误：日志 + 退避后重试，避免错误循环把 CPU 打满。
			if errors.Is(err, context.Canceled) {
				continue
			}
			if IsNil(err) {
				continue
			}
			warning(fmt.Sprintf("virtualstream worker (node=%q): XREAD failed: %s",
				w.nodeKey, err.Error()))
			select {
			case <-w.loopCtx.Done():
				return
			case <-time.After(streamReconnectBackoff):
			}
			continue
		}

		for _, xs := range result.Val() {
			w.dispatch(xs)
		}
	}
}
