package redisson

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/redis/rueidis"
	"github.com/sandwich-go/boost/z"
)

// PubSubMonitorOptions 配置显式启用的 PubSub 观测，不创建额外 Redis Client。
// Now 必须返回单调纳秒偏移；未指定的字段使用默认值。
// 回调必须快速返回，不能执行 Redis 命令或其他阻塞操作。
// Receive 的 context Hook 会与已有业务 Hook 组合，不覆盖原有回调。
type PubSubMonitorOptions struct {
	Now     func() time.Duration
	OnEvent func(PubSubEvent)
	// OnDequeue 统计从本机入队到写入 Out 缓冲的时间，不包含业务从 Out 取走前的等待。
	OnDequeue func(time.Duration)
	// DequeueSampleEvery 仅控制等待时长回调的采样间隔；0 使用默认 64，1 为全量。
	// 队列长度、字节数、最老消息年龄与订阅事件不采样。
	DequeueSampleEvery uint32
}

// DefaultPubSubMonitorOptions 提供单调时钟和空回调；只有 WithPubSubMonitoring 才启用采集。
func DefaultPubSubMonitorOptions() PubSubMonitorOptions {
	return PubSubMonitorOptions{
		Now:                func() time.Duration { return time.Duration(z.MonoOffset()) },
		OnEvent:            func(PubSubEvent) {},
		OnDequeue:          func(time.Duration) {},
		DequeueSampleEvery: 64,
	}
}

func (o PubSubMonitorOptions) withDefaults() PubSubMonitorOptions {
	defaults := DefaultPubSubMonitorOptions()
	if o.Now == nil {
		o.Now = defaults.Now
	}
	if o.OnEvent == nil {
		o.OnEvent = defaults.OnEvent
	}
	if o.OnDequeue == nil {
		o.OnDequeue = defaults.OnDequeue
	}
	if o.DequeueSampleEvery == 0 {
		o.DequeueSampleEvery = defaults.DequeueSampleEvery
	}
	return o
}

// PubSubEvent 是已观测到的订阅确认、接收退出、重试或关闭事件。
// Channels 用于上层定位受影响范围，不适合作为指标标签。
type PubSubEvent struct {
	Kind     string
	Command  string
	Channel  string
	Channels []string
	Err      error
	Dropped  int
}

const (
	// PubSubSubscribeAck 表示 Redis 已确认订阅频道或 pattern，不代表业务缓存已校正。
	PubSubSubscribeAck = "subscribe_ack"
	// PubSubUnsubscribeAck 表示 Redis 已确认取消订阅频道或 pattern。
	PubSubUnsubscribeAck = "unsubscribe_ack"
	// PubSubReceiveReturn 表示底层一次 Receive 返回并清空对应 ACK，之后仍可能内部重试。
	PubSubReceiveReturn = "receive_return"
	// PubSubReceiveError 表示接收任务最终因错误结束，且 context 尚未取消或超时。
	PubSubReceiveError = "receive_error"
	// PubSubRetry 表示分片接收任务准备再次执行 SSUBSCRIBE，不包含 rueidis 内部重试。
	PubSubRetry = "retry"
	// PubSubClosed 表示本地接收队列已停止生产者并完成丢弃清理，Dropped 不含 Out 未读消息。
	PubSubClosed = "closed"
)

// PubSubStats 是本机观测快照；Enabled 为 false 时，其余读数未被采集。
// Acknowledged 是当前接收任务最近观测到的 ACK，不等于业务已经完成缓存校正。
// QueueBytes 仅统计消息、频道和 pattern 字符串的字节数，不包含队列结构开销。
type PubSubStats struct {
	Enabled         bool
	Closed          bool
	QueueDepth      int
	QueueBytes      int64
	QueueOldestAge  time.Duration
	Receivers       int
	Subscriptions   int
	Acknowledged    int
	SubscribeAcks   uint64
	UnsubscribeAcks uint64
	ReceiveReturns  uint64
	ReceiveErrors   uint64
	Retries         uint64
	LastError       error
}

// PubSubObserver 是可选诊断接口，不扩充 PubSub，已有实现与 mock 无需修改。
type PubSubObserver interface{ PubSubStats() PubSubStats }

type pubSubMonitorContextKey struct{}

// WithPubSubMonitoring 为使用该 context 创建的 PubSub 显式启用采集，未指定的字段使用默认值。
// 同一个 context 不重复配置观测，以免静默替换已有时钟或事件回调。
func WithPubSubMonitoring(ctx context.Context, options PubSubMonitorOptions) context.Context {
	if ctx.Value(pubSubMonitorContextKey{}) != nil {
		panic("pubsub monitoring is already configured")
	}
	return context.WithValue(ctx, pubSubMonitorContextKey{}, options.withDefaults())
}

type monitoredSubscription struct {
	command   string
	channels  []string
	confirmed map[string]struct{}
}

type pubSubMonitor struct {
	options       PubSubMonitorOptions
	queue         *pubSubMonitorQueue
	messages      *unboundedChan[Message, Message]
	mu            sync.Mutex
	subscriptions map[*monitoredSubscription]struct{}
	stats         PubSubStats
}

func newPubSubMonitor(ctx context.Context, initialCapacity int) *pubSubMonitor {
	options, enabled := ctx.Value(pubSubMonitorContextKey{}).(PubSubMonitorOptions)
	m := &pubSubMonitor{options: DefaultPubSubMonitorOptions()}
	if !enabled {
		m.messages = newUnboundedChan[Message](ctx, initialCapacity)
		return m
	}
	m.options = options
	m.subscriptions = make(map[*monitoredSubscription]struct{})
	m.queue = newPubSubMonitorQueue(ctx, initialCapacity, options.Now, options.OnDequeue, func(dropped int) {
		m.emit(PubSubEvent{Kind: PubSubClosed, Dropped: dropped})
	}, options.DequeueSampleEvery)
	return m
}

func (m *pubSubMonitor) emit(event PubSubEvent) {
	m.options.OnEvent(event)
}

func (m *pubSubMonitor) observeReceiver(ctx context.Context, command string, channels []string) (context.Context, func(Message), func(error)) {
	receive, release := m.onReceive(ctx)
	if m.queue == nil {
		return ctx, receive, func(error) { release() }
	}
	s := &monitoredSubscription{command: command, channels: append([]string(nil), channels...), confirmed: make(map[string]struct{})}
	m.mu.Lock()
	m.subscriptions[s] = struct{}{}
	m.mu.Unlock()
	hooks := rueidis.WithOnSubscriptionHook(context.Background(), func(ack rueidis.PubSubSubscription) { m.onSubscription(s, ack) })
	hooks = rueidis.WithOnReceiveReturnHook(hooks, func(err error, _ rueidis.CommandClient) error {
		m.onReceiveReturn(ctx, s, err)
		return err
	})
	return &pubSubHooksContext{Context: ctx, hooks: hooks}, receive, func(err error) {
		defer release()
		m.onFinish(ctx, s, err)
	}
}

// 两个 Hook context 仅查询官方工厂创建的 opaque key，不依赖其私有 key 编号。
// 先记录观测，再执行已有业务 Hook；返回 Hook 保留业务对错误返回值的处理。
type pubSubHooksContext struct {
	context.Context
	hooks context.Context
}

func (c *pubSubHooksContext) Value(key any) any {
	own := c.hooks.Value(key)
	switch hook := own.(type) {
	case func(rueidis.PubSubSubscription):
		previous, _ := c.Context.Value(key).(func(rueidis.PubSubSubscription))
		if previous == nil {
			return hook
		}
		return func(event rueidis.PubSubSubscription) { hook(event); previous(event) }
	case func(error, rueidis.CommandClient) error:
		previous, _ := c.Context.Value(key).(func(error, rueidis.CommandClient) error)
		if previous == nil {
			return hook
		}
		return func(err error, client rueidis.CommandClient) error { return previous(hook(err, client), client) }
	default:
		return c.Context.Value(key)
	}
}

func (m *pubSubMonitor) onSubscription(s *monitoredSubscription, ack rueidis.PubSubSubscription) {
	m.mu.Lock()
	if _, exists := m.subscriptions[s]; !exists {
		m.mu.Unlock()
		return
	}
	if !slices.Contains(s.channels, ack.Channel) {
		m.mu.Unlock()
		return
	}
	event := PubSubEvent{Command: s.command, Channel: ack.Channel}
	switch ack.Kind {
	case "subscribe", "ssubscribe", "psubscribe":
		s.confirmed[ack.Channel] = struct{}{}
		m.stats.SubscribeAcks++
		event.Kind = PubSubSubscribeAck
	case "unsubscribe", "sunsubscribe", "punsubscribe":
		delete(s.confirmed, ack.Channel)
		m.stats.UnsubscribeAcks++
		event.Kind = PubSubUnsubscribeAck
	default:
		m.mu.Unlock()
		return
	}
	m.mu.Unlock()
	m.emit(event)
}

// 底层可能在一个接收任务内重试，每次返回都先让对应 ACK 失效。
func (m *pubSubMonitor) onReceiveReturn(ctx context.Context, s *monitoredSubscription, err error) {
	m.mu.Lock()
	if _, exists := m.subscriptions[s]; !exists {
		m.mu.Unlock()
		return
	}
	clear(s.confirmed)
	m.stats.ReceiveReturns++
	if err != nil && ctx.Err() == nil {
		m.stats.LastError = err
	}
	m.mu.Unlock()
	m.emit(PubSubEvent{Kind: PubSubReceiveReturn, Command: s.command, Channels: append([]string(nil), s.channels...), Err: err})
}

// 整个接收任务结束才移除登记，主动取消不计入终止错误。
func (m *pubSubMonitor) onFinish(ctx context.Context, s *monitoredSubscription, err error) {
	m.mu.Lock()
	delete(m.subscriptions, s)
	unexpected := err != nil && ctx.Err() == nil
	if unexpected {
		m.stats.ReceiveErrors++
		m.stats.LastError = err
	}
	m.mu.Unlock()
	if unexpected {
		m.emit(PubSubEvent{Kind: PubSubReceiveError, Command: s.command, Channels: append([]string(nil), s.channels...), Err: err})
	}
}

func (m *pubSubMonitor) observeRetry(command, channel string) {
	if m.queue == nil {
		return
	}
	m.mu.Lock()
	m.stats.Retries++
	m.mu.Unlock()
	m.emit(PubSubEvent{Kind: PubSubRetry, Command: command, Channel: channel})
}

// 在接收任务初始化时选择入口，消息路径不重复判断是否启用观测。
func (m *pubSubMonitor) onReceive(ctx context.Context) (func(Message), func()) {
	if m.queue != nil {
		sender := m.queue.newSender(ctx)
		return func(message Message) { sender.push(message) }, sender.release
	}
	sender := m.messages.newSender(ctx)
	return func(message Message) { sender.send(message) }, sender.release
}

func (m *pubSubMonitor) channel() <-chan Message {
	if m.queue != nil {
		return m.queue.out
	}
	return m.messages.Out
}

func (p *pubSub) PubSubStats() PubSubStats {
	return p.monitor.snapshot()
}

func (m *pubSubMonitor) snapshot() PubSubStats {
	if m.queue == nil {
		return PubSubStats{}
	}
	m.mu.Lock()
	snapshot := m.stats
	snapshot.Enabled = true
	snapshot.Receivers = len(m.subscriptions)
	type identity struct{ command, channel string }
	requested := make(map[identity]struct{})
	confirmed := make(map[identity]struct{})
	for s := range m.subscriptions {
		for _, channel := range s.channels {
			requested[identity{s.command, channel}] = struct{}{}
		}
		for channel := range s.confirmed {
			confirmed[identity{s.command, channel}] = struct{}{}
		}
	}
	snapshot.Subscriptions = len(requested)
	snapshot.Acknowledged = len(confirmed)
	m.mu.Unlock()
	snapshot.QueueDepth, snapshot.QueueBytes, snapshot.QueueOldestAge, snapshot.Closed = m.queue.snapshot()
	return snapshot
}
