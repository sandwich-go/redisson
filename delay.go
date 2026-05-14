package redisson

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================================================
// Lua 脚本
//
// 数据结构（每个 DelayQueue 实例独占）：
//   - delay:{name}  ZSET   score = 任务到期 unix 秒（绝对时间），member = payload
//   - doing:{name}  ZSET   score = visibility deadline unix 秒（处理超时绝对时间），member = payload
//   - meta:{name}   HASH   field = payload, value = 已失败次数（int，HINCRBY 维护）
//
// 命名采用 `{name}` hashtag，确保同一队列三键落到同一 cluster slot。
// 所有脚本在 Redis 单线程内原子执行；Go 端只负责调度和重试策略。
// ============================================================================

// addDelayTaskLua 添加一个延迟任务。
// KEYS: [delaySet, metaSet]
// ARGV: [payload, expireUnixSeconds]
// 行为：覆盖 delaySet 中同 payload 的旧 score；同时清掉 meta（重置失败计数）。
var addDelayTaskLua = `
local delay_set, meta_set = KEYS[1], KEYS[2]
local value, score = ARGV[1], tonumber(ARGV[2])
redis.call('ZADD', delay_set, score, value)
redis.call('HDEL', meta_set, value)
return 1
`

// delDelayTaskLua 删除任务（不论在 delay 还是 doing）。
// KEYS: [delaySet, doingSet, metaSet]
// ARGV: [payload]
var delDelayTaskLua = `
local delay_set, doing_set, meta_set = KEYS[1], KEYS[2], KEYS[3]
local value = ARGV[1]
redis.call('ZREM', delay_set, value)
redis.call('ZREM', doing_set, value)
redis.call('HDEL', meta_set, value)
return 1
`

// delayTaskLengthLua 返回 delay+doing 总长度。
// KEYS: [delaySet, doingSet]
var delayTaskLengthLua = `
local delay_set, doing_set = KEYS[1], KEYS[2]
return redis.call('ZCARD', delay_set) + redis.call('ZCARD', doing_set)
`

// pollDelayTaskLua 从 delay 中原子检出 score<=now 的若干 item，移到 doing
// 并把 doing 中的 score 设为 now+visibilityTimeout（处理超时回收点）；
// 同时回带每个 item 当前的失败次数（来自 meta hash）。
// KEYS: [delaySet, doingSet, metaSet]
// ARGV: [nowSeconds, visibilityDeadlineSeconds, batchLimit]
// 返回：[payload1, retries1, payload2, retries2, ...]
var pollDelayTaskLua = `
local delay_set, doing_set, meta_set = KEYS[1], KEYS[2], KEYS[3]
local now = tonumber(ARGV[1])
local visibility_deadline = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local items = redis.call('ZRANGEBYSCORE', delay_set, '-inf', now, 'LIMIT', 0, limit)
local out = {}
for i, value in ipairs(items) do
    redis.call('ZADD', doing_set, visibility_deadline, value)
    redis.call('ZREM', delay_set, value)
    local retries = redis.call('HGET', meta_set, value)
    if not retries then retries = '0' end
    out[#out+1] = value
    out[#out+1] = retries
end
return out
`

// reclaimDelayTaskLua 把 doing 中 score<=now 的 item（已超过 visibility）移回 delay，
// 立即可重试；不增加重试计数（reclaim 不视为业务失败，仅恢复消费权）。
// KEYS: [delaySet, doingSet]
// ARGV: [nowSeconds, batchLimit]
// 返回：被回收的 item 数量
var reclaimDelayTaskLua = `
local delay_set, doing_set = KEYS[1], KEYS[2]
local now = tonumber(ARGV[1])
local limit = tonumber(ARGV[2])
local items = redis.call('ZRANGEBYSCORE', doing_set, '-inf', now, 'LIMIT', 0, limit)
for i, value in ipairs(items) do
    redis.call('ZADD', delay_set, now, value)
    redis.call('ZREM', doing_set, value)
end
return #items
`

// ackOKLua 业务成功：从 doing 清掉 + 清 meta。
// KEYS: [doingSet, metaSet]
// ARGV: [payload]
var ackOKLua = `
local doing_set, meta_set = KEYS[1], KEYS[2]
local value = ARGV[1]
redis.call('ZREM', doing_set, value)
redis.call('HDEL', meta_set, value)
return 1
`

// ackRetryLua 业务失败但未达死信阈值：累加失败计数，把 item 移回 delay。
// KEYS: [delaySet, doingSet, metaSet]
// ARGV: [payload, retryAtUnixSeconds]
// 返回：累加后的失败次数
var ackRetryLua = `
local delay_set, doing_set, meta_set = KEYS[1], KEYS[2], KEYS[3]
local value, retry_at = ARGV[1], tonumber(ARGV[2])
local retries = redis.call('HINCRBY', meta_set, value, 1)
redis.call('ZREM', doing_set, value)
redis.call('ZADD', delay_set, retry_at, value)
return retries
`

// ackDeadLua 业务失败且达到死信阈值：从 doing 清掉 + 清 meta。
// 和 ackOKLua 行为一致；分名是为了语义清晰、便于将来扩展（如写入 dead-letter list）。
// KEYS: [doingSet, metaSet]
// ARGV: [payload]
var ackDeadLua = ackOKLua

// ============================================================================
// 错误与常量
// ============================================================================

var (
	// ErrEmptyDelayQueueName 创建 DelayQueue 时 name 为空。
	ErrEmptyDelayQueueName = errors.New("delay queue name cannot be empty")
	// ErrEmptyDelayQueueCallback 创建 DelayQueue 时 callback 为 nil。
	ErrEmptyDelayQueueCallback = errors.New("delay queue callback cannot be empty")
	// ErrDelayQueueHasClosed 已关闭的 DelayQueue 上调用 Close 之外的方法。
	ErrDelayQueueHasClosed = errors.New("delay queue has closed")
	// ErrDelayQueueHasStarted ticker 重复启动时返回（内部错误，正常路径不会暴露给用户）。
	ErrDelayQueueHasStarted = errors.New("delay queue has started")
	// ErrDelayQueueHasRegistered 同名 DelayQueue 已经注册到 client。
	// 旧代码可能调用 NewDelayQueue 期望"已存在则返回旧实例"——v2 起改为显式报错；
	// 调用方应在创建前确保旧实例已 Close 或使用不同 name。
	ErrDelayQueueHasRegistered = errors.New("delay queue has registered")
)

const (
	delayLogPrefix      = "[redis-delay]:"
	delayKeyFormat      = "delay:{%s}"
	delayDoingKeyFormat = "doing:{%s}"
	delayMetaKeyFormat  = "meta:{%s}"

	// defaultPollInterval ticker 默认间隔；用户暂不可配。
	defaultPollInterval = time.Second
	// defaultPollBatch 单次 poll/reclaim 处理的最大 item 数；防止单次脚本耗时过长。
	defaultPollBatch = 64
	// defaultRedisOpTimeout 单次 Redis 操作的兜底超时；与业务 callback 超时解耦。
	defaultRedisOpTimeout = 5 * time.Second
	// defaultRetryBackoff 业务失败后再次可见的间隔（与 visibility timeout 解耦）。
	// 旧版本错误地用 Timeout 既当 visibility 又当 retry 间隔，导致默认 Timeout=1min 时
	// 业务失败后要等 1 分钟才重试。新版本拆开：visibility 用 Timeout（reclaim 用），
	// retry backoff 用此常量，保持业务失败 → 短时再投递 的预期。
	defaultRetryBackoff = time.Second
)

// DelayQueue 单 namespace 的 Redis 延迟队列。
//
// 实现细节：
//   - 数据结构见 delay.go 顶部的 Lua 脚本说明（delay/doing 两个 ZSET + meta HASH）。
//   - 所有方法对并发调用安全。
//   - Close 是幂等的：首次 Close 释放资源；后续 Close 直接返回 ErrDelayQueueHasClosed 而不 panic。
//   - 单个 DelayQueue 实例假设由唯一 callback 处理；要换 callback 请先 Close 再 NewDelayQueue。
type DelayQueue interface {
	// Name 返回队列名（创建时传入的 name，不含 prefix）。
	Name() string
	// Add 添加一条延迟任务。
	//
	// payload 在队列内具有唯一性：同 name 队列中同 payload 的二次 Add 会覆盖到期时间，
	// 同时重置失败计数（和"创建新任务"语义等价）。
	// delay 表示从当前时间起的相对延迟，最小精度 1 秒（亚秒会向下取整）。
	Add(ctx context.Context, payload []byte, delay time.Duration) error
	// Del 删除一条任务（不论它当前在 delay 还是 doing 状态）。
	// 已经被 callback 拉走但尚未 ack 的 item 也会被删除，但若 callback 已开始执行则无法终止。
	Del(ctx context.Context, payload []byte) error
	// Length 返回 delay+doing 总长度（包含正在处理中的 item）。
	Length(ctx context.Context) (int64, error)
	// Close 停止 ticker、取消正在进行的 Redis 调用、从 client 摘除该队列。
	// 不会等待用户 callback 返回；callback 可以在另一 goroutine 中安全调用 Close（不会死锁）。
	// 重复 Close 返回 ErrDelayQueueHasClosed。
	Close() error
}

// ============================================================================
// 实现
// ============================================================================

type delayQueue struct {
	c    *client
	spec DelayOptionsVisitor
	name string

	// running 状态机：true=运行中,false=已关闭。
	// 用 CompareAndSwap 守护 Close 幂等且避免 double close panic。
	running atomic.Bool

	// ctx/cancel：所有内部 Redis 调用都使用 ctx 派生超时；Close 时 cancel 让在途调用立即返回。
	ctx    context.Context
	cancel context.CancelFunc

	// tickerWG 等 ticker goroutine 退出。
	// 注意：worker（执行用户 callback 的 goroutine）刻意不放进 wg，
	// 这样 callback 可以在自身内部安全调用 q.Close()。
	tickerWG sync.WaitGroup
	exitC    chan struct{}

	// keys 顺序按 Lua 脚本声明对齐，避免每次调用临时构造切片
	pollKeys    []string // [delay, doing, meta]
	delKeys     []string // [delay, doing, meta]
	lengthKeys  []string // [delay, doing]
	reclaimKeys []string // [delay, doing]
	addKeys     []string // [delay, meta]
	ackOKKeys   []string // [doing, meta]
	ackDeadKeys []string // [doing, meta]
	// ackRetry 复用 delKeys（[delay, doing, meta]）

	addScript     Scripter
	delScript     Scripter
	lengthScript  Scripter
	pollScript    Scripter
	reclaimScript Scripter
	ackOKScript   Scripter
	ackRetry      Scripter
	ackDeadScript Scripter

	callback func([]byte) error
}

// NewDelayQueue 在 client 上创建一个 DelayQueue。
//
// 同名队列已存在时返回 ErrDelayQueueHasRegistered；这是与 v1 的行为差异：
// v1 会静默返回旧实例（丢弃新 callback），容易让调用方误以为新 callback 生效。
//
// callback 以"成功 ack=nil err / 失败重试=non-nil err"语义运行；超过 RetryTimes 后触发 HandleDeadLetter。
// callback panic 会被 recover 视为失败一次。
func (c *client) NewDelayQueue(name string, f func([]byte) error, opts ...DelayOption) (DelayQueue, error) {
	q, err := newDelayQueue(c, name, f, opts...)
	if err != nil {
		return nil, err
	}
	// LoadOrStore 保证并发安全的"严格唯一注册"
	if _, loaded := c.delayQueues.LoadOrStore(q.name, q); loaded {
		// 撞名：把刚启动的 ticker 收回，返回错误
		q.running.Store(false)
		q.shutdown()
		return nil, ErrDelayQueueHasRegistered
	}
	return q, nil
}

func newDelayQueue(c *client, name string, f func([]byte) error, opts ...DelayOption) (*delayQueue, error) {
	if name == "" {
		return nil, ErrEmptyDelayQueueName
	}
	if f == nil {
		return nil, ErrEmptyDelayQueueCallback
	}
	spec := newDelayOptions(opts...)
	ctx, cancel := context.WithCancel(context.Background())
	q := &delayQueue{
		c:        c,
		spec:     spec,
		name:     name,
		ctx:      ctx,
		cancel:   cancel,
		exitC:    make(chan struct{}),
		callback: f,
	}
	q.running.Store(true)

	// 三个 key 用 hashtag 锁定到同一 slot
	delayKey := fmt.Sprintf(delayKeyFormat, name)
	doingKey := fmt.Sprintf(delayDoingKeyFormat, name)
	metaKey := fmt.Sprintf(delayMetaKeyFormat, name)
	if prefix := spec.GetPrefix(); prefix != "" {
		delayKey = fmt.Sprintf("%s:%s", prefix, delayKey)
		doingKey = fmt.Sprintf("%s:%s", prefix, doingKey)
		metaKey = fmt.Sprintf("%s:%s", prefix, metaKey)
	}
	q.pollKeys = []string{delayKey, doingKey, metaKey}
	q.delKeys = []string{delayKey, doingKey, metaKey}
	q.lengthKeys = []string{delayKey, doingKey}
	q.reclaimKeys = []string{delayKey, doingKey}
	q.addKeys = []string{delayKey, metaKey}
	q.ackOKKeys = []string{doingKey, metaKey}
	q.ackDeadKeys = []string{doingKey, metaKey}

	q.addScript = c.CreateScript(addDelayTaskLua)
	q.delScript = c.CreateScript(delDelayTaskLua)
	q.lengthScript = c.CreateScript(delayTaskLengthLua)
	q.pollScript = c.CreateScript(pollDelayTaskLua)
	q.reclaimScript = c.CreateScript(reclaimDelayTaskLua)
	q.ackOKScript = c.CreateScript(ackOKLua)
	q.ackRetry = c.CreateScript(ackRetryLua)
	q.ackDeadScript = c.CreateScript(ackDeadLua)

	q.startTickers()
	return q, nil
}

func (q *delayQueue) Name() string { return q.name }

func (q *delayQueue) Add(ctx context.Context, payload []byte, delay time.Duration) error {
	if !q.running.Load() {
		return ErrDelayQueueHasClosed
	}
	sec := formatSec(delay)
	if sec < 0 {
		sec = 0
	}
	expireAt := nowFunc().Unix() + sec
	return q.addScript.Run(ctx, q.addKeys, payload, expireAt).Err()
}

func (q *delayQueue) Del(ctx context.Context, payload []byte) error {
	if !q.running.Load() {
		return ErrDelayQueueHasClosed
	}
	return q.delScript.Run(ctx, q.delKeys, payload).Err()
}

func (q *delayQueue) Length(ctx context.Context) (int64, error) {
	// Length 在 Close 后仍允许查询（只读、用于 collector），不强制状态检查。
	return q.lengthScript.Run(ctx, q.lengthKeys).Int64()
}

// Close 实现 DelayQueue.Close。CAS 守护幂等。
func (q *delayQueue) Close() error {
	if !q.running.CompareAndSwap(true, false) {
		return ErrDelayQueueHasClosed
	}
	q.shutdown()
	q.c.delayQueues.Delete(q.name)
	return nil
}

// shutdown 是 Close 的幂等内部实现：取消 ctx、关闭 exitC、等 ticker 退出。
// 不操作 c.delayQueues，由 Close 或 NewDelayQueue 撞名兜底自行处理。
//
// shutdown 不等待 worker（callback）完成，这样 callback 内部可以安全调用 q.Close。
// 正在进行的 Redis 调用会被 ctx cancel 立即终止；未完成 ack 的 item 会被下一进程的
// reclaim ticker 在 visibility 超时后回收。
func (q *delayQueue) shutdown() {
	q.cancel()
	close(q.exitC)
	q.tickerWG.Wait()
}

// startTickers 启动 poll/reclaim 两个 ticker goroutine。
// ticker 内同步调用任务函数，避免旧版本"go ti.f()"导致的并发触发与 wg 不等待 bug。
func (q *delayQueue) startTickers() {
	q.tickerWG.Add(2)
	go q.runTicker(defaultPollInterval, q.pollOnce)
	go q.runTicker(defaultPollInterval, q.reclaimOnce)
}

// runTicker 周期触发 fn；同步执行确保前一轮完成后才进入下一轮。
// fn 内可能 spawn worker goroutine 跑用户 callback；那部分独立于 ticker wg。
func (q *delayQueue) runTicker(interval time.Duration, fn func()) {
	defer q.tickerWG.Done()

	t := time.NewTimer(interval)
	defer t.Stop()
	for {
		select {
		case <-q.exitC:
			return
		case <-t.C:
			fn()
			// Reset 在 fn 返回后才进入下一轮；保证不并发触发。
			t.Reset(interval)
		}
	}
}

// pollOnce 拉取一批到期 item 并逐个 spawn worker 执行 callback。
// worker 不被 tickerWG 等待；ctx 取消时 worker 内的 ack/retry 会快速失败。
func (q *delayQueue) pollOnce() {
	if !q.running.Load() {
		return
	}
	now := nowFunc()
	visibilityDeadline := now.Add(q.spec.GetTimeout()).Unix()

	ctx, cancel := q.opCtx()
	defer cancel()
	res, err := q.pollScript.Run(ctx, q.pollKeys, now.Unix(), visibilityDeadline, defaultPollBatch).Slice()
	if err != nil {
		q.c.handler.delayPollError(q.name)
		if !errors.Is(err, context.Canceled) {
			e(fmt.Sprintf("%s poll error, %v", delayLogPrefix, err))
		}
		return
	}
	if len(res) == 0 {
		return
	}
	for i := 0; i+1 < len(res); i += 2 {
		payload := []byte(toString(res[i]))
		retries, _ := strconv.Atoi(toString(res[i+1]))
		go q.executeOne(payload, retries)
	}
}

// executeOne 在独立 goroutine 中执行一个 item 的 callback 与 ack。
// recover panic 视为失败一次。
func (q *delayQueue) executeOne(payload []byte, retries int) {
	err := q.runCallback(payload)
	if err == nil {
		q.ackSuccess(payload)
		return
	}
	// 业务失败：判断是否达到死信阈值
	// 注意：retries 是本次执行前的失败次数；本次失败后实际计数为 retries+1。
	if retries+1 >= q.spec.GetRetryTimes() {
		q.ackDead(payload)
		return
	}
	q.ackRetryWithBackoff(payload)
}

// runCallback 执行用户 callback；recover panic。
func (q *delayQueue) runCallback(payload []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("handle task panic, %v", r)
			e(fmt.Sprintf("%s %v", delayLogPrefix, err))
		}
	}()
	return q.callback(payload)
}

// ackSuccess 业务成功：移出 doing + 清 meta。失败仅日志（reclaim 兜底）。
func (q *delayQueue) ackSuccess(payload []byte) {
	ctx, cancel := q.opCtx()
	defer cancel()
	if err := q.ackOKScript.Run(ctx, q.ackOKKeys, payload).Err(); err != nil {
		if !errors.Is(err, context.Canceled) {
			e(fmt.Sprintf("%s ack ok failed, payload=%q err=%v", delayLogPrefix, payload, err))
		}
	}
}

// ackRetryWithBackoff 业务失败但未达死信阈值：累加 meta 计数 + 把 item 放回 delay。
// 重新可见时间 = now + defaultRetryBackoff，与 visibility timeout 解耦。
func (q *delayQueue) ackRetryWithBackoff(payload []byte) {
	retryAt := nowFunc().Add(defaultRetryBackoff).Unix()
	ctx, cancel := q.opCtx()
	defer cancel()
	if err := q.ackRetry.Run(ctx, q.delKeys, payload, retryAt).Err(); err != nil {
		if !errors.Is(err, context.Canceled) {
			e(fmt.Sprintf("%s ack retry failed, payload=%q err=%v", delayLogPrefix, payload, err))
		}
	}
}

// ackDead 业务达到死信阈值：清理 + 触发用户 dead-letter hook。
// 先清 Redis 状态再调 hook；hook panic 不影响清理结果。
func (q *delayQueue) ackDead(payload []byte) {
	ctx, cancel := q.opCtx()
	defer cancel()
	if err := q.ackDeadScript.Run(ctx, q.ackDeadKeys, payload).Err(); err != nil {
		if !errors.Is(err, context.Canceled) {
			e(fmt.Sprintf("%s ack dead failed, payload=%q err=%v", delayLogPrefix, payload, err))
		}
	}
	if hook := q.spec.GetHandleDeadLetter(); hook != nil {
		defer func() {
			if r := recover(); r != nil {
				e(fmt.Sprintf("%s handle dead letter panic, %v", delayLogPrefix, r))
			}
		}()
		hook(payload)
	}
}

// reclaimOnce 把 doing 中超过 visibility 的 item 拉回 delay 重新可见。
// 不增加重试计数：reclaim 表示上一次执行进程崩溃/卡死，并非业务失败。
func (q *delayQueue) reclaimOnce() {
	if !q.running.Load() {
		return
	}
	now := nowFunc()
	ctx, cancel := q.opCtx()
	defer cancel()
	count, err := q.reclaimScript.Run(ctx, q.reclaimKeys, now.Unix(), defaultPollBatch).Int64()
	if err != nil {
		q.c.handler.delayReclaimError(q.name)
		if !errors.Is(err, context.Canceled) {
			e(fmt.Sprintf("%s reclaim error, %v", delayLogPrefix, err))
		}
		return
	}
	if count > 0 {
		q.c.handler.delayReclaim(q.name, int(count))
	}
}

// opCtx 构造一个继承自 q.ctx 的子 ctx，带 Redis 操作超时；Close 时一并取消。
func (q *delayQueue) opCtx() (context.Context, context.CancelFunc) {
	timeout := q.spec.GetTimeout()
	if timeout <= 0 || timeout > defaultRedisOpTimeout {
		timeout = defaultRedisOpTimeout
	}
	return context.WithTimeout(q.ctx, timeout)
}

// toString 把 Lua 返回值兜底转 string。Lua 通常返回 string；插入数字时返回 int64。
func toString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case []byte:
		return string(x)
	case int64:
		return strconv.FormatInt(x, 10)
	case int:
		return strconv.Itoa(x)
	default:
		return fmt.Sprint(v)
	}
}
