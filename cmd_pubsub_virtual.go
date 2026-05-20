package redisson

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/redis/rueidis"
)

// 虚拟 PubSub（VirtualPubSub / VirtualPubSubHub）解决的问题：
//
// Redis PubSub（含 sharded pub/sub）每个 SUBSCRIBE/PSUBSCRIBE/SSUBSCRIBE 会
// 占用 rueidis 内部一条 pubsub 专用连接 / dedicated wire。当业务侧在同一个
// 进程内对成百上千个 topic 各自 Subscribe 时，会导致：
//   1. 客户端打开的 TCP 连接数线性增长，触发 ulimit / fd 上限；
//   2. Redis Server 端 maxclients 上限被打爆；
//   3. rueidis 内部 wire 池大量阻塞、影响其他命令。
//
// Hub 提供一个进程级（更准确说是 client 级）共享的 pubsub 复用层：
//   - 在 Hub 内部，对相同的 redis channel/pattern 只发起 ** 一条 ** 真实 SUBSCRIBE。
//   - 所有用户层面的 VirtualPubSub.Subscribe(channel) 在 Hub 内部做引用计数：
//     第一次订阅时下发 SUBSCRIBE，引用归零时下发 UNSUBSCRIBE。
//   - 连接复用通过 rueidis.Dedicate() + SetPubSubHooks 实现：单条 dedicated
//     连接在 SetPubSubHooks 之后可以接收所有 OnMessage 回调，并允许后续通过
//     普通 Do(SUBSCRIBE/UNSUBSCRIBE) 动态增删订阅，不需要为每个新 channel
//     另起新连接。
//
// 适用范围与限制：
//   - 单实例 / 主从：1 个 Hub × 3 种订阅类型 ⇒ 最多 3 条共享连接；
//   - Cluster：本实现的 Subscribe / PSubscribe 与 redisson 原本的 c.cmd.Receive
//     行为一致（rueidis 把 SUBSCRIBE 发到其中一个节点，Redis Cluster 各版本
//     对 publish 的 broadcast 行为不同，跨节点广播的覆盖性由 server 决定）；
//     SSubscribe 由 rueidis 按 channel CRC16 路由到 slot owner 节点，因此每
//     个用到的 owner 节点会持有一条 dedicated 连接。
//
// 不在本实现范畴：
//   - reconnection：底层 rueidis 已自动重连；本实现在 dedicated 连接断开时
//     会让 hooks error chan 出错，Hub 触发重建 dedicated 并把当前路由表里
//     所有 channel 重新 SUBSCRIBE。
//   - 跨进程 / 多 client 共享：Hub 与 *client 绑定；想跨多 client 共享请自行
//     在外层做单例。

// VirtualPubSubHub 是进程内 pubsub 多路复用器。
// 通过 client.NewVirtualPubSubHub() 创建，可生成多个 VirtualPubSub 实例，
// 这些实例共享 Hub 内部的少量真实 redis 连接。
type VirtualPubSubHub interface {
	// Subscribe 创建一个虚拟 PubSub 并订阅 channels。
	// 之后还可在返回的 VirtualPubSub 上继续 Subscribe / Unsubscribe / 切换其它订阅类型。
	Subscribe(ctx context.Context, channels ...string) (VirtualPubSub, error)

	// PSubscribe 创建一个虚拟 PubSub 并按 pattern 订阅。
	PSubscribe(ctx context.Context, patterns ...string) (VirtualPubSub, error)

	// SSubscribe 创建一个虚拟 PubSub 并按 sharded channel 订阅。
	// 仅 Redis 7.0+ 支持。Cluster 下按 channel 名 CRC16 路由到对应 slot 的 owner 节点。
	SSubscribe(ctx context.Context, channels ...string) (VirtualPubSub, error)

	// Close 关闭整个 Hub：停止所有真实订阅连接、关闭所有 VirtualPubSub 的 Channel()。
	// 已发出的 VirtualPubSub.Close() 是幂等的；Hub.Close 后再调 VirtualPubSub 任何方法都返回 ErrVirtualHubClosed。
	Close() error

	// Stats 返回当前 Hub 的统计信息（订阅者数、真实订阅 key 数等），便于监控。
	Stats() VirtualPubSubStats
}

// VirtualPubSub 暴露给用户的虚拟订阅者。
// 行为与 PubSub 类似，但内部不持有专用 Redis 连接，仅通过 Hub 的路由表参与消息分发。
//
// 与 PubSub 的差异：
//   - 同一个 VirtualPubSub 实例可以混用 Subscribe / PSubscribe / SSubscribe（Hub 会路由）；
//   - Channel() 在 Close 之后会被 unboundedChan process goroutine 关闭，下游 range 自然退出；
//   - 慢消费策略与 pubSub 一致（unboundedChan，慢消费者只影响自己，Hub 内部转发不会被阻塞）。
type VirtualPubSub interface {
	Subscribe(ctx context.Context, channels ...string) error
	Unsubscribe(ctx context.Context, channels ...string) error

	PSubscribe(ctx context.Context, patterns ...string) error
	PUnsubscribe(ctx context.Context, patterns ...string) error

	SSubscribe(ctx context.Context, channels ...string) error
	SUnsubscribe(ctx context.Context, channels ...string) error

	Channel() <-chan Message
	Close() error
}

// VirtualPubSubStats 是 Hub 的运行时统计。
type VirtualPubSubStats struct {
	// Subscribers 当前活跃 VirtualPubSub 数量
	Subscribers int
	// Channels Hub 内真实 SUBSCRIBE 的 channel 数
	Channels int
	// Patterns Hub 内真实 PSUBSCRIBE 的 pattern 数
	Patterns int
	// ShardChannels Hub 内真实 SSUBSCRIBE 的 shard channel 数
	ShardChannels int
	// Connections Hub 持有的真实 dedicated 连接数（每个 pool 内每节点 1 条）
	Connections int
}

// ErrVirtualHubClosed 在 Hub 已 Close 后操作返回。
var ErrVirtualHubClosed = errors.New("virtual pubsub hub closed")

// kind 表示一种订阅类型。
type virtualKind int

const (
	kindSubscribe  virtualKind = iota // SUBSCRIBE / message
	kindPSubscribe                    // PSUBSCRIBE / pmessage
	kindSSubscribe                    // SSUBSCRIBE / smessage
)

func (k virtualKind) String() string {
	switch k {
	case kindSubscribe:
		return "subscribe"
	case kindPSubscribe:
		return "psubscribe"
	case kindSSubscribe:
		return "ssubscribe"
	}
	return "unknown"
}

// virtualHub 是 VirtualPubSubHub 的具体实现。
// 内部对 3 种订阅类型各维护一个 pool，pool 之间彼此独立。
type virtualHub struct {
	client *client

	// pools[kindSubscribe], pools[kindPSubscribe], pools[kindSSubscribe]
	pools [3]*subscriptionPool

	mu     sync.Mutex                  // 保护 subs 表
	subs   map[*virtualPubSub]struct{} // 全部活跃订阅者（用于 Close 全量收尾）
	closed AtomicInt32
}

// NewVirtualPubSubHub 在 client 上创建一个 PubSub Hub。
// 多次调用会返回相互独立的 Hub（每个 Hub 独立持有一组共享连接）。
// 业务侧通常只需 1 个 Hub。
func (c *client) NewVirtualPubSubHub() VirtualPubSubHub {
	h := &virtualHub{client: c, subs: make(map[*virtualPubSub]struct{})}
	h.pools[kindSubscribe] = newSubscriptionPool(c, kindSubscribe)
	h.pools[kindPSubscribe] = newSubscriptionPool(c, kindPSubscribe)
	h.pools[kindSSubscribe] = newSubscriptionPool(c, kindSSubscribe)
	return h
}

func (h *virtualHub) isClosed() bool { return h.closed.Get() == 1 }

func (h *virtualHub) ensureOpen() error {
	if h.isClosed() {
		return ErrVirtualHubClosed
	}
	return nil
}

func (h *virtualHub) newVirtualPubSub() *virtualPubSub {
	v := newVirtualPubSub(h)
	h.mu.Lock()
	if h.isClosed() {
		h.mu.Unlock()
		_ = v.Close()
		return v // 由调用方判定 Close 状态后丢弃
	}
	h.subs[v] = struct{}{}
	h.mu.Unlock()
	return v
}

func (h *virtualHub) removeSubscriber(v *virtualPubSub) {
	h.mu.Lock()
	delete(h.subs, v)
	h.mu.Unlock()
}

func (h *virtualHub) Subscribe(ctx context.Context, channels ...string) (VirtualPubSub, error) {
	if err := h.ensureOpen(); err != nil {
		return nil, err
	}
	v := h.newVirtualPubSub()
	if h.isClosed() {
		return nil, ErrVirtualHubClosed
	}
	if len(channels) > 0 {
		if err := v.Subscribe(ctx, channels...); err != nil {
			_ = v.Close()
			return nil, err
		}
	}
	return v, nil
}

func (h *virtualHub) PSubscribe(ctx context.Context, patterns ...string) (VirtualPubSub, error) {
	if err := h.ensureOpen(); err != nil {
		return nil, err
	}
	v := h.newVirtualPubSub()
	if h.isClosed() {
		return nil, ErrVirtualHubClosed
	}
	if len(patterns) > 0 {
		if err := v.PSubscribe(ctx, patterns...); err != nil {
			_ = v.Close()
			return nil, err
		}
	}
	return v, nil
}

func (h *virtualHub) SSubscribe(ctx context.Context, channels ...string) (VirtualPubSub, error) {
	if err := h.ensureOpen(); err != nil {
		return nil, err
	}
	v := h.newVirtualPubSub()
	if h.isClosed() {
		return nil, ErrVirtualHubClosed
	}
	if len(channels) > 0 {
		if err := v.SSubscribe(ctx, channels...); err != nil {
			_ = v.Close()
			return nil, err
		}
	}
	return v, nil
}

func (h *virtualHub) Close() error {
	if !h.closed.CompareAndSwap(0, 1) {
		return nil
	}
	// 1) 摘除所有 subscriber：close 每个 virtualPubSub 的 ctx，断开它们与 Hub 的关联。
	h.mu.Lock()
	subs := h.subs
	h.subs = nil
	h.mu.Unlock()
	for v := range subs {
		_ = v.closeFromHub()
	}
	// 2) 关闭所有 pool（断开真实 dedicated 连接、停止重连循环）。
	for _, p := range h.pools {
		p.close()
	}
	return nil
}

func (h *virtualHub) Stats() VirtualPubSubStats {
	st := VirtualPubSubStats{}
	h.mu.Lock()
	st.Subscribers = len(h.subs)
	h.mu.Unlock()
	st.Channels = h.pools[kindSubscribe].routeCount()
	st.Patterns = h.pools[kindPSubscribe].routeCount()
	st.ShardChannels = h.pools[kindSSubscribe].routeCount()
	st.Connections = h.pools[kindSubscribe].connCount() +
		h.pools[kindPSubscribe].connCount() +
		h.pools[kindSSubscribe].connCount()
	return st
}

// ============================================================================
// virtualPubSub: VirtualPubSub 的实现
// ============================================================================

type virtualPubSub struct {
	hub   *virtualHub
	msgCh *unboundedChan[Message]

	// ctx 同时驱动两件事：
	//   1. forwardMessage 的"早返回"select 防阻塞；
	//   2. msgCh 的 process goroutine：cancel 后 process 退出 → defer close(Out) → 下游 range 退出。
	// 这与现有 pubSub.Close 完全同形，避免主动 close(In) 引发的"已关闭 channel send" panic。
	ctx    context.Context
	cancel context.CancelFunc
	closed AtomicInt32

	// 跟踪本订阅者通过本 Hub 持有的所有 channel/pattern，用于 Close 时统一释放。
	mu      sync.Mutex
	holding [3]map[string]struct{} // 按 virtualKind 索引
}

func newVirtualPubSub(h *virtualHub) *virtualPubSub {
	ctx, cancel := context.WithCancel(context.Background())
	v := &virtualPubSub{
		hub:    h,
		msgCh:  newUnboundedChan[Message](ctx, h.client.v.GetPubSubChanSize()),
		ctx:    ctx,
		cancel: cancel,
	}
	for i := range v.holding {
		v.holding[i] = make(map[string]struct{})
	}
	return v
}

func (v *virtualPubSub) isClosed() bool { return v.closed.Get() == 1 }

// forwardMessage 由 pool 在收到匹配的 redis 消息后调用，必须永不阻塞。
// 与 pubSub.forwardMessage 同形：双重保护（isClosed 早返回 + select on ctx.Done）。
func (v *virtualPubSub) forwardMessage(kind virtualKind, m rueidis.PubSubMessage) {
	if v.isClosed() {
		return
	}
	select {
	case v.msgCh.In <- m:
	case <-v.ctx.Done():
		warning(fmt.Sprintf("virtualpubsub %s: dropped after close, channel=%s", kind, m.Channel))
	}
}

func (v *virtualPubSub) Channel() <-chan Message { return v.msgCh.Out }

// Close 是用户侧关闭：摘除所有路由 → cancel ctx（process goroutine 退出 → close(Out)）。
// 与 pubSub.Close 同形——不主动 close(In)，避免与并发 forwardMessage 形成 send-on-closed panic；
// process goroutine 在 ctx.Done 后 return → defer close(out) → 用户 range 退出。
func (v *virtualPubSub) Close() error {
	if !v.closed.CompareAndSwap(0, 1) {
		return nil
	}
	v.hub.removeSubscriber(v)
	v.unsubscribeAll()
	v.cancel()
	return nil
}

// closeFromHub 由 Hub.Close 触发：与 Close 行为一致但不再回头操作 hub.subs（已被 hub 清空）。
func (v *virtualPubSub) closeFromHub() error {
	if !v.closed.CompareAndSwap(0, 1) {
		return nil
	}
	v.unsubscribeAll()
	v.cancel()
	return nil
}

// unsubscribeAll 释放本订阅者在所有 pool 上的 holding。
func (v *virtualPubSub) unsubscribeAll() {
	v.mu.Lock()
	holdings := v.holding
	v.holding = [3]map[string]struct{}{
		make(map[string]struct{}),
		make(map[string]struct{}),
		make(map[string]struct{}),
	}
	v.mu.Unlock()
	for kind, m := range holdings {
		if len(m) == 0 {
			continue
		}
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		// 即使关闭流程也尽量给一个 ctx；Hub 已退出时 pool.close 会幂等忽略。
		_ = v.hub.pools[kind].unsubscribe(context.Background(), v, keys)
	}
}

func (v *virtualPubSub) trackAdd(kind virtualKind, keys []string) {
	v.mu.Lock()
	for _, k := range keys {
		v.holding[kind][k] = struct{}{}
	}
	v.mu.Unlock()
}

func (v *virtualPubSub) trackRemove(kind virtualKind, keys []string) {
	v.mu.Lock()
	for _, k := range keys {
		delete(v.holding[kind], k)
	}
	v.mu.Unlock()
}

func (v *virtualPubSub) ensureOpen() error {
	if v.isClosed() {
		return ErrVirtualHubClosed
	}
	if v.hub.isClosed() {
		return ErrVirtualHubClosed
	}
	return nil
}

func (v *virtualPubSub) Subscribe(ctx context.Context, channels ...string) error {
	if err := v.ensureOpen(); err != nil {
		return err
	}
	if len(channels) == 0 {
		return nil
	}
	ctx = v.hub.client.handler.before(ctx, CommandSubscribe)
	err := v.hub.pools[kindSubscribe].subscribe(ctx, v, channels)
	if err == nil {
		v.trackAdd(kindSubscribe, channels)
	}
	v.hub.client.handler.after(ctx, err)
	return err
}

func (v *virtualPubSub) Unsubscribe(ctx context.Context, channels ...string) error {
	if v.isClosed() {
		return nil // 关闭语义：再次 Unsubscribe 无副作用
	}
	if len(channels) == 0 {
		return nil
	}
	ctx = v.hub.client.handler.before(ctx, CommandUnsubscribe)
	err := v.hub.pools[kindSubscribe].unsubscribe(ctx, v, channels)
	if err == nil {
		v.trackRemove(kindSubscribe, channels)
	}
	v.hub.client.handler.after(ctx, err)
	return err
}

func (v *virtualPubSub) PSubscribe(ctx context.Context, patterns ...string) error {
	if err := v.ensureOpen(); err != nil {
		return err
	}
	if len(patterns) == 0 {
		return nil
	}
	ctx = v.hub.client.handler.before(ctx, CommandPSubscribe)
	err := v.hub.pools[kindPSubscribe].subscribe(ctx, v, patterns)
	if err == nil {
		v.trackAdd(kindPSubscribe, patterns)
	}
	v.hub.client.handler.after(ctx, err)
	return err
}

func (v *virtualPubSub) PUnsubscribe(ctx context.Context, patterns ...string) error {
	if v.isClosed() {
		return nil
	}
	if len(patterns) == 0 {
		return nil
	}
	ctx = v.hub.client.handler.before(ctx, CommandPUnsubscribe)
	err := v.hub.pools[kindPSubscribe].unsubscribe(ctx, v, patterns)
	if err == nil {
		v.trackRemove(kindPSubscribe, patterns)
	}
	v.hub.client.handler.after(ctx, err)
	return err
}

func (v *virtualPubSub) SSubscribe(ctx context.Context, channels ...string) error {
	if err := v.ensureOpen(); err != nil {
		return err
	}
	if len(channels) == 0 {
		return nil
	}
	ctx = v.hub.client.handler.before(ctx, CommandSSubscribe)
	err := v.hub.pools[kindSSubscribe].subscribe(ctx, v, channels)
	if err == nil {
		v.trackAdd(kindSSubscribe, channels)
	}
	v.hub.client.handler.after(ctx, err)
	return err
}

func (v *virtualPubSub) SUnsubscribe(ctx context.Context, channels ...string) error {
	if v.isClosed() {
		return nil
	}
	if len(channels) == 0 {
		return nil
	}
	ctx = v.hub.client.handler.before(ctx, CommandSUnsubscribe)
	err := v.hub.pools[kindSSubscribe].unsubscribe(ctx, v, channels)
	if err == nil {
		v.trackRemove(kindSSubscribe, channels)
	}
	v.hub.client.handler.after(ctx, err)
	return err
}

// ============================================================================
// subscriptionPool: 单种订阅类型对应的 routing 表 + 真实连接池
// ============================================================================

// subscriptionPool 维护：
//   1. routes: redis 侧 channel/pattern → (refCount + subscribers)
//   2. conns:  每个 redis 节点一条 dedicated 连接
//
// 每条 conn 注册了 SetPubSubHooks，OnMessage 回调进入 pool.dispatch，
// 后者根据 m.Channel / m.Pattern 查 routes 把消息扇出到所有订阅了该 key 的 virtualPubSub。
//
// 引用计数策略：
//   - subscribe(v, [k1,k2,...]):
//       对每个 ki：route[ki].subs.add(v)；若该 key 是 0->1 增长，下发真正的 SUBSCRIBE ki。
//   - unsubscribe(v, [k1,k2,...]):
//       对每个 ki：route[ki].subs.delete(v)；若该 key 1->0 降为 0，下发 UNSUBSCRIBE ki。
//
// 并发：pool.mu 守护 routes 与 conns；下发 SUBSCRIBE 命令是 rueidis 提供的并发安全
// 操作，但为了让"决定是否下发"与"实际下发"两步原子，下发动作仍持锁完成（命令是
// fire-and-forget 的 Do，时间复杂度低，持锁影响可控）。
type subscriptionPool struct {
	client *client
	kind   virtualKind

	mu     sync.Mutex
	routes map[string]*routeEntry
	conns  map[string]*nodeConn // key: 节点 addr 或 ""（单实例）
	closed bool
}

type routeEntry struct {
	subs map[*virtualPubSub]struct{}
	// node 标识本 key 当前实际订阅在哪条 conn 上；用于 UNSUBSCRIBE 时找回。
	// 单实例下永远是 ""；cluster 下 SSUBSCRIBE 是 slot owner addr，普通 SUBSCRIBE 也固定到 ""。
	node string
}

func newSubscriptionPool(c *client, kind virtualKind) *subscriptionPool {
	return &subscriptionPool{
		client: c,
		kind:   kind,
		routes: make(map[string]*routeEntry),
		conns:  make(map[string]*nodeConn),
	}
}

func (p *subscriptionPool) routeCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.routes)
}

func (p *subscriptionPool) connCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.conns)
}

// nodeKey 决定一个 key 应当订阅到哪条 conn。
//   - 非 cluster：一律 ""
//   - cluster + SSUBSCRIBE：按 channel CRC16 → slot → owner addr（依赖 rueidis 内部路由）
//   - cluster + SUBSCRIBE/PSUBSCRIBE：固定 ""，让 rueidis 自由选节点（与原 redisson 行为一致）
//
// 注：本实现不再针对 cluster 下的 SUBSCRIBE/PSUBSCRIBE 做 fan-out 全节点订阅；
// 用户若关心跨节点 publish 覆盖，请在 publish 端做或参考 README 说明。
func (p *subscriptionPool) nodeKey(key string) string {
	if !p.client.IsCluster() {
		return ""
	}
	if p.kind == kindSSubscribe {
		// 用 channel 名做 slot 路由；rueidis Builder.Ssubscribe().Channel() 内部一致。
		// 但本层并不直接用 slot，而是通过 rueidis Nodes() 找 addr：rueidis 自身已经
		// 把 SSUBSCRIBE 路由实现好了，对我们来说只要保证 dedicated 连接选对节点。
		// rueidis Dedicate() 在第一次发命令时按命令 Slot 自动绑定节点，所以这里
		// 我们以 channel 名作为 conn map 的 key，复用即可（同 slot 的多个 channel 落到同一 addr）。
		return clusterSlotKey(key)
	}
	return ""
}

// clusterSlotKey 把 channel/pattern 转成"同 slot 同节点"的 key。
// 实际值无所谓，只要同 slot 的 channel 算出相同的字符串即可。
func clusterSlotKey(channel string) string {
	return fmt.Sprintf("slot:%d", slot(channel))
}

func (p *subscriptionPool) close() {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	conns := p.conns
	p.conns = nil
	p.routes = nil
	p.mu.Unlock()
	for _, c := range conns {
		c.close()
	}
}

// subscribe 把一组 keys 加到本订阅者的路由中，必要时下发真实 SUBSCRIBE。
func (p *subscriptionPool) subscribe(ctx context.Context, v *virtualPubSub, keys []string) error {
	// 按节点分组：cluster + SSUBSCRIBE 下不同 key 可能落在不同 owner 节点。
	groups := map[string][]string{}
	for _, k := range keys {
		nk := p.nodeKey(k)
		groups[nk] = append(groups[nk], k)
	}

	for nodeKey, gKeys := range groups {
		// 在锁内确定哪些 key 是首次订阅（0→1）；锁外不持有 conn 操作。
		p.mu.Lock()
		if p.closed {
			p.mu.Unlock()
			return ErrVirtualHubClosed
		}
		var newKeys []string
		for _, k := range gKeys {
			r := p.routes[k]
			if r == nil {
				r = &routeEntry{subs: make(map[*virtualPubSub]struct{}), node: nodeKey}
				p.routes[k] = r
				newKeys = append(newKeys, k)
			}
			r.subs[v] = struct{}{}
		}
		conn, err := p.getOrCreateConnLocked(ctx, nodeKey, gKeys)
		if err != nil {
			// 回滚刚加入的路由（避免泄漏）。
			for _, k := range gKeys {
				if r := p.routes[k]; r != nil {
					delete(r.subs, v)
					if len(r.subs) == 0 {
						delete(p.routes, k)
					}
				}
			}
			p.mu.Unlock()
			return err
		}
		p.mu.Unlock()

		// 真实下发 SUBSCRIBE 在锁外进行（命令本身可能短暂阻塞 dedicated wire）。
		if len(newKeys) > 0 {
			if err := conn.subscribe(ctx, p.kind, newKeys); err != nil {
				// 失败回滚路由；conn 仍保留，下次再用。
				p.mu.Lock()
				for _, k := range newKeys {
					if r := p.routes[k]; r != nil {
						delete(r.subs, v)
						if len(r.subs) == 0 {
							delete(p.routes, k)
						}
					}
				}
				p.mu.Unlock()
				return err
			}
		}
	}
	return nil
}

// unsubscribe 删除一组 keys 的引用，必要时下发真实 UNSUBSCRIBE。
func (p *subscriptionPool) unsubscribe(ctx context.Context, v *virtualPubSub, keys []string) error {
	// 按 nodeKey 分组（要 UNSUBSCRIBE 到对应 conn）
	type nodeWork struct {
		conn       *nodeConn
		toUnsubKey []string
	}
	work := map[string]*nodeWork{}

	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	for _, k := range keys {
		r := p.routes[k]
		if r == nil {
			continue
		}
		delete(r.subs, v)
		if len(r.subs) == 0 {
			// 引用归零：从 routes 摘除，记账到对应 conn 进行真正 UNSUBSCRIBE
			delete(p.routes, k)
			conn := p.conns[r.node]
			if conn != nil {
				w, ok := work[r.node]
				if !ok {
					w = &nodeWork{conn: conn}
					work[r.node] = w
				}
				w.toUnsubKey = append(w.toUnsubKey, k)
			}
		}
	}
	p.mu.Unlock()

	// 锁外下发 UNSUBSCRIBE
	var firstErr error
	for _, w := range work {
		if err := w.conn.unsubscribe(ctx, p.kind, w.toUnsubKey); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// dispatch 由 conn.OnMessage 调用：把 msg 扇出到所有匹配 route 的订阅者。
func (p *subscriptionPool) dispatch(m rueidis.PubSubMessage) {
	// pmessage 时 m.Pattern 是 hub 注册的 pattern，m.Channel 是真实命中 channel；
	// route key 取决于 kind:
	//   subscribe / ssubscribe → m.Channel
	//   psubscribe → m.Pattern
	var key string
	if p.kind == kindPSubscribe {
		key = m.Pattern
	} else {
		key = m.Channel
	}

	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	r := p.routes[key]
	if r == nil {
		p.mu.Unlock()
		return
	}
	// 复制 subs 到栈上的 slice，锁外做扇出，避免 forwardMessage 内部任何阻塞放大持锁时长。
	subs := make([]*virtualPubSub, 0, len(r.subs))
	for s := range r.subs {
		subs = append(subs, s)
	}
	p.mu.Unlock()

	for _, s := range subs {
		s.forwardMessage(p.kind, m)
	}
}

// getOrCreateConnLocked 必须在 p.mu 持锁下调用。
// pickKey 用于 cluster 下首次绑定 dedicated 连接到正确节点（rueidis 在 Dedicate() 后
// 第一个命令的 Slot 决定绑定的 node）。
func (p *subscriptionPool) getOrCreateConnLocked(ctx context.Context, nodeKey string, pickKey []string) (*nodeConn, error) {
	if c, ok := p.conns[nodeKey]; ok {
		return c, nil
	}
	pinKey := ""
	if len(pickKey) > 0 {
		pinKey = pickKey[0]
	}
	c, err := newNodeConn(p, nodeKey, pinKey)
	if err != nil {
		return nil, err
	}
	p.conns[nodeKey] = c
	return c, nil
}

// onConnLost 由 nodeConn 回调：底层 dedicated 连接断开，清理并尝试重建。
// 重建后会把当前 routes 中归属该节点的所有 key 重新订阅一次。
func (p *subscriptionPool) onConnLost(nodeKey string) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	old := p.conns[nodeKey]
	delete(p.conns, nodeKey)
	// 收集需要重订的 keys
	var resubKeys []string
	for k, r := range p.routes {
		if r.node == nodeKey {
			resubKeys = append(resubKeys, k)
		}
	}
	p.mu.Unlock()

	if old != nil {
		old.close()
	}
	if len(resubKeys) == 0 {
		return
	}

	// 异步触发一次 lazy 重订：取一个 key 引导 dedicated 节点绑定，再 subscribe 全部。
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), p.client.v.GetWriteTimeout())
		defer cancel()
		p.mu.Lock()
		if p.closed {
			p.mu.Unlock()
			return
		}
		conn, err := p.getOrCreateConnLocked(ctx, nodeKey, resubKeys)
		p.mu.Unlock()
		if err != nil {
			e(fmt.Sprintf("virtualpubsub %s: reconnect node=%q failed: %s", p.kind, nodeKey, err.Error()))
			return
		}
		if err := conn.subscribe(ctx, p.kind, resubKeys); err != nil {
			e(fmt.Sprintf("virtualpubsub %s: resubscribe %d keys on node=%q failed: %s",
				p.kind, len(resubKeys), nodeKey, err.Error()))
		}
	}()
}

// ============================================================================
// nodeConn: 单条 dedicated 真实连接
// ============================================================================

// nodeConn 包装一条 rueidis dedicated 连接 + SetPubSubHooks 注册。
// 所有订阅消息通过 hooks.OnMessage 进入 pool.dispatch。
type nodeConn struct {
	pool       *subscriptionPool
	nodeKey    string
	dedicated  rueidis.DedicatedClient
	cancelFn   func()
	hookErrCh  <-chan error
	closeOnce  sync.Once
	closedFlag AtomicInt32
}

// newNodeConn 在指定 nodeKey 上建立 dedicated 连接、注册 hooks。
//   - pinKey 用于 cluster 下首次发命令绑定节点：rueidis 文档要求 cluster client 的
//     dedicated 第一条命令必须有 Key()。我们通过给 SUBSCRIBE/SSUBSCRIBE 提供一个
//     channel 来满足，但 SetPubSubHooks 本身不算命令；所以 pinKey 在外层 subscribe()
//     调用时已经会带着 newKeys 一起下发，这里不需要主动 ping。
func newNodeConn(p *subscriptionPool, nodeKey, pinKey string) (*nodeConn, error) {
	dc, cancel := p.client.cmd.Dedicate()
	nc := &nodeConn{pool: p, nodeKey: nodeKey, dedicated: dc, cancelFn: cancel}
	hooks := rueidis.PubSubHooks{
		OnMessage: func(m rueidis.PubSubMessage) {
			// 关闭后到达的回调直接丢弃（pool.dispatch 内会再做一次 closed 检查）。
			if nc.closedFlag.Get() == 1 {
				return
			}
			p.dispatch(m)
		},
		// OnSubscription 用于确认订阅状态变化，目前不需要使用，留空保持轻量。
	}
	errCh := dc.SetPubSubHooks(hooks)
	if errCh == nil {
		// hooks 必为非零；理论上不会进入这里，作为防御。
		cancel()
		return nil, errors.New("virtualpubsub: SetPubSubHooks returned nil error chan")
	}
	nc.hookErrCh = errCh
	// 监听 hookErrCh：dedicated 断开/recycle 时触发 pool 重建。
	go nc.watch()
	return nc, nil
}

func (nc *nodeConn) watch() {
	err, ok := <-nc.hookErrCh
	if nc.closedFlag.Get() == 1 {
		return
	}
	if !ok {
		// 正常关闭路径
		return
	}
	if err != nil {
		warning(fmt.Sprintf("virtualpubsub %s: dedicated lost on node=%q: %s",
			nc.pool.kind, nc.nodeKey, err.Error()))
	}
	nc.pool.onConnLost(nc.nodeKey)
}

// subscribe 在该连接上下发 SUBSCRIBE / PSUBSCRIBE / SSUBSCRIBE 命令。
func (nc *nodeConn) subscribe(ctx context.Context, kind virtualKind, keys []string) error {
	if nc.closedFlag.Get() == 1 {
		return ErrVirtualHubClosed
	}
	cmd := buildSubscribe(nc.dedicated.B(), kind, keys)
	res := nc.dedicated.Do(ctx, cmd)
	if err := res.Error(); err != nil && !isSubscribeOK(err) {
		return fmt.Errorf("virtualpubsub %s subscribe %v: %w", kind, keys, err)
	}
	return nil
}

func (nc *nodeConn) unsubscribe(ctx context.Context, kind virtualKind, keys []string) error {
	if nc.closedFlag.Get() == 1 {
		return nil // 连接已关，路由已被 pool 清空，等价 noop
	}
	cmd := buildUnsubscribe(nc.dedicated.B(), kind, keys)
	res := nc.dedicated.Do(ctx, cmd)
	if err := res.Error(); err != nil && !isSubscribeOK(err) {
		return fmt.Errorf("virtualpubsub %s unsubscribe %v: %w", kind, keys, err)
	}
	return nil
}

func (nc *nodeConn) close() {
	nc.closeOnce.Do(func() {
		nc.closedFlag.Set(1)
		// 不主动发 UNSUBSCRIBE：rueidis 在 cancelFn 里会回收 dedicated 连接，
		// 该连接关闭即可，路由表由 pool 自己维护。
		if nc.cancelFn != nil {
			nc.cancelFn()
		}
	})
}

// buildSubscribe 根据订阅类型构造对应的 rueidis Completed 命令。
func buildSubscribe(b Builder, kind virtualKind, keys []string) Completed {
	switch kind {
	case kindSubscribe:
		return b.Subscribe().Channel(keys...).Build()
	case kindPSubscribe:
		return b.Psubscribe().Pattern(keys...).Build()
	case kindSSubscribe:
		return b.Ssubscribe().Channel(keys...).Build()
	}
	panic(fmt.Sprintf("virtualpubsub: unknown kind %d", kind))
}

func buildUnsubscribe(b Builder, kind virtualKind, keys []string) Completed {
	switch kind {
	case kindSubscribe:
		return b.Unsubscribe().Channel(keys...).Build()
	case kindPSubscribe:
		return b.Punsubscribe().Pattern(keys...).Build()
	case kindSSubscribe:
		return b.Sunsubscribe().Channel(keys...).Build()
	}
	panic(fmt.Sprintf("virtualpubsub: unknown kind %d", kind))
}

// isSubscribeOK 判断 rueidis 对 SUBSCRIBE 类命令返回的错误是否可视为成功。
// rueidis 的 SUBSCRIBE/UNSUBSCRIBE 在 dedicated 连接上 Do 完成后通常返回 nil，
// 但某些版本会在订阅确认时返回 "subscribe" / "unsubscribe" 类型的非 error 反馈，
// 仅在 Error() 上做过滤即可。这里把"包含 unsubscribe/subscribe 关键字"的错误视为 OK，
// 主要为兼容上游可能的实现变化。
func isSubscribeOK(err error) bool {
	if err == nil {
		return true
	}
	s := err.Error()
	if strings.Contains(s, "subscribe") || strings.Contains(s, "unsubscribe") {
		return true
	}
	return false
}
