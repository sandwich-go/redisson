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
// 数据结构（每个 DelayQueue 实例独占；时间单位统一为 毫秒）：
//   - delay:{name}  ZSET    score = 任务到期 unix ms（绝对时间），member = payload
//   - doing:{name}  ZSET    score = visibility deadline unix ms，member = payload
//   - meta:{name}   HASH    field  = payload, value = 已失败次数（int，HINCRBY 维护）
//   - owner:{name}  HASH    field  = payload, value = fence token（拿到任务的 worker 持有）
//   - seq:{name}    STRING  fence token 单调递增源（INCR 维护）
//
// 命名采用 `{name}` hashtag，确保同一队列各键落到同一 cluster slot。
// 所有脚本在 Redis 单线程内原子执行；Go 端只负责调度、重试策略与 fence token 持有。
//
// Fence Token 机制（防止 callback 卡住超过 visibility 时被本实例 reclaim 重复调度）：
//   - poll 时为每个 item 分配一个新 token（INCR 拿到），写入 owner hash 并返回给 worker；
//   - worker 启动后周期 heartbeat：用 token 校验后推后 doing 的 score；
//   - worker ack（成功/重试/死信）：用 token 校验，token 不匹配说明已被 reclaim 移交他人，
//     当前 worker 跳过 ack（避免误删别人的状态），item 由新持有者负责；
//   - reclaim 时清掉 owner，使旧 worker 的后续 heartbeat / ack 全部因 token 不匹配而 noop。
// ============================================================================

// addDelayTaskLua 添加一个延迟任务。
// KEYS: [delaySet, metaSet, ownerSet]
// ARGV: [payload, expireUnixMs]
// 行为：覆盖 delaySet 中同 payload 的旧 score；同时清掉 meta（重置失败计数）和 owner。
var addDelayTaskLua = `
local delay_set, meta_set, owner_set = KEYS[1], KEYS[2], KEYS[3]
local value, score = ARGV[1], tonumber(ARGV[2])
redis.call('ZADD', delay_set, score, value)
redis.call('HDEL', meta_set, value)
redis.call('HDEL', owner_set, value)
return 1
`

// delDelayTaskLua 删除任务（不论它当前在 delay 还是 doing）。
// KEYS: [delaySet, doingSet, metaSet, ownerSet]
// ARGV: [payload]
var delDelayTaskLua = `
local delay_set, doing_set, meta_set, owner_set = KEYS[1], KEYS[2], KEYS[3], KEYS[4]
local value = ARGV[1]
redis.call('ZREM', delay_set, value)
redis.call('ZREM', doing_set, value)
redis.call('HDEL', meta_set, value)
redis.call('HDEL', owner_set, value)
return 1
`

// delayTaskLengthLua 返回 delay+doing 总长度。
// KEYS: [delaySet, doingSet]
var delayTaskLengthLua = `
local delay_set, doing_set = KEYS[1], KEYS[2]
return redis.call('ZCARD', delay_set) + redis.call('ZCARD', doing_set)
`

// pollDelayTaskLua 从 delay 中原子检出 score<=now 的若干 item，移到 doing，
// 为每个 item 分配新的 fence token 写入 owner hash，同时回带 retries 与 token。
// KEYS: [delaySet, doingSet, metaSet, ownerSet, seqKey]
// ARGV: [nowMs, visibilityDeadlineMs, batchLimit]
// 返回：[payload1, retries1, token1, payload2, retries2, token2, ...]
var pollDelayTaskLua = `
local delay_set, doing_set, meta_set, owner_set, seq_key = KEYS[1], KEYS[2], KEYS[3], KEYS[4], KEYS[5]
local now = tonumber(ARGV[1])
local visibility_deadline = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local items = redis.call('ZRANGEBYSCORE', delay_set, '-inf', now, 'LIMIT', 0, limit)
local out = {}
for i, value in ipairs(items) do
    local token = redis.call('INCR', seq_key)
    redis.call('ZADD', doing_set, visibility_deadline, value)
    redis.call('ZREM', delay_set, value)
    redis.call('HSET', owner_set, value, token)
    local retries = redis.call('HGET', meta_set, value)
    if not retries then retries = '0' end
    out[#out+1] = value
    out[#out+1] = retries
    out[#out+1] = tostring(token)
end
return out
`

// reclaimDelayTaskLua 把 doing 中 score<=now 的 item（已超过 visibility）移回 delay
// 并清掉 owner（让旧持有者的 heartbeat/ack 因 token 不匹配而失效）。
// 不增加重试计数：reclaim 表示上一持有者疑似崩溃，不是业务失败。
// KEYS: [delaySet, doingSet, ownerSet]
// ARGV: [nowMs, batchLimit]
// 返回：被回收的 item 数量
var reclaimDelayTaskLua = `
local delay_set, doing_set, owner_set = KEYS[1], KEYS[2], KEYS[3]
local now = tonumber(ARGV[1])
local limit = tonumber(ARGV[2])
local items = redis.call('ZRANGEBYSCORE', doing_set, '-inf', now, 'LIMIT', 0, limit)
for i, value in ipairs(items) do
    redis.call('ZADD', delay_set, now, value)
    redis.call('ZREM', doing_set, value)
    redis.call('HDEL', owner_set, value)
end
return #items
`

// heartbeatDelayTaskLua worker 续约：校验 token 后把 doing 的 score 推到新 deadline。
// KEYS: [doingSet, ownerSet]
// ARGV: [payload, expectedToken, newDeadlineMs]
// 返回：1=续约成功；0=token 不匹配（已被 reclaim 或已 ack），调用方应停止 heartbeat。
var heartbeatDelayTaskLua = `
local doing_set, owner_set = KEYS[1], KEYS[2]
local value, expected, new_deadline = ARGV[1], ARGV[2], tonumber(ARGV[3])
local cur = redis.call('HGET', owner_set, value)
if cur == false or cur ~= expected then return 0 end
redis.call('ZADD', doing_set, new_deadline, value)
return 1
`

// ackOKLua 业务成功：校验 token，匹配则从 doing 清掉 + 清 meta + 清 owner。
// KEYS: [doingSet, metaSet, ownerSet]
// ARGV: [payload, expectedToken]
// 返回：1=成功；0=token 不匹配（lease 已转移）。
var ackOKLua = `
local doing_set, meta_set, owner_set = KEYS[1], KEYS[2], KEYS[3]
local value, expected = ARGV[1], ARGV[2]
local cur = redis.call('HGET', owner_set, value)
if cur == false or cur ~= expected then return 0 end
redis.call('ZREM', doing_set, value)
redis.call('HDEL', meta_set, value)
redis.call('HDEL', owner_set, value)
return 1
`

// ackRetryLua 业务失败但未达死信阈值：校验 token，匹配则累加失败计数，把 item 移回 delay。
// KEYS: [delaySet, doingSet, metaSet, ownerSet]
// ARGV: [payload, expectedToken, retryAtMs]
// 返回：>0=新失败次数；0=token 不匹配。
var ackRetryLua = `
local delay_set, doing_set, meta_set, owner_set = KEYS[1], KEYS[2], KEYS[3], KEYS[4]
local value, expected, retry_at = ARGV[1], ARGV[2], tonumber(ARGV[3])
local cur = redis.call('HGET', owner_set, value)
if cur == false or cur ~= expected then return 0 end
local retries = redis.call('HINCRBY', meta_set, value, 1)
redis.call('ZREM', doing_set, value)
redis.call('ZADD', delay_set, retry_at, value)
redis.call('HDEL', owner_set, value)
return retries
`

// ackDeadLua 业务达到死信阈值：校验 token 后清干净。
// 行为等同 ackOKLua（都从 doing/meta/owner 清掉）；分名为了语义清晰，
// 便于将来扩展（如写入 dead-letter list）。
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
	delayOwnerKeyFormat = "owner:{%s}"
	delaySeqKeyFormat   = "seq:{%s}"

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
	// delayCloseWorkerWaitTimeout q.Close 等本 queue 在途 worker 完成 ack 的最长时间。
	// 主路径毫秒级返回；callback 内调 q.Close 的反模式下用此 timeout 打破死锁。
	// client.Close 仍会在更高层等所有 worker 收尾，保证无 c.cmd race。
	delayCloseWorkerWaitTimeout = 5 * time.Second
	// heartbeatRatio worker heartbeat 间隔 = visibilityTimeout / heartbeatRatio。
	// 默认 3：每过 1/3 visibility 续一次，给网络抖动留 2 个 RTT 的容错窗口。
	heartbeatRatio = 3
	// minHeartbeatInterval heartbeat 间隔下限，避免 visibility 极短时打爆 Redis。
	minHeartbeatInterval = 200 * time.Millisecond
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
	tickerWG sync.WaitGroup
	// workerWG 跟踪本 queue 的在途 worker；q.Close 等之，但带超时
	// （避免 callback 内调 q.Close 时形成"worker 等 callback ↔ callback 等 Close"死锁）。
	// 即使超时，client.delayWorkerWG 仍会兜底等到 worker 完全收尾。
	workerWG sync.WaitGroup
	exitC    chan struct{}

	// keys 顺序按 Lua 脚本声明对齐，避免每次调用临时构造切片
	pollKeys      []string // [delay, doing, meta, owner, seq]
	delKeys       []string // [delay, doing, meta, owner]
	lengthKeys    []string // [delay, doing]
	reclaimKeys   []string // [delay, doing, owner]
	addKeys       []string // [delay, meta, owner]
	ackOKKeys     []string // [doing, meta, owner]
	ackRetryKeys  []string // [delay, doing, meta, owner]
	ackDeadKeys   []string // [doing, meta, owner]
	heartbeatKeys []string // [doing, owner]

	addScript       Scripter
	delScript       Scripter
	lengthScript    Scripter
	pollScript      Scripter
	reclaimScript   Scripter
	ackOKScript     Scripter
	ackRetryScript  Scripter
	ackDeadScript   Scripter
	heartbeatScript Scripter

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
		// 这里 q 刚创建，不可能有 worker；只需关 ticker。
		if q.running.CompareAndSwap(true, false) {
			close(q.exitC)
			q.tickerWG.Wait()
			q.cancel()
		}
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

	// 五个 key 用同一 {name} hashtag 锁定到同一 cluster slot
	delayKey := fmt.Sprintf(delayKeyFormat, name)
	doingKey := fmt.Sprintf(delayDoingKeyFormat, name)
	metaKey := fmt.Sprintf(delayMetaKeyFormat, name)
	ownerKey := fmt.Sprintf(delayOwnerKeyFormat, name)
	seqKey := fmt.Sprintf(delaySeqKeyFormat, name)
	if prefix := spec.GetPrefix(); prefix != "" {
		delayKey = fmt.Sprintf("%s:%s", prefix, delayKey)
		doingKey = fmt.Sprintf("%s:%s", prefix, doingKey)
		metaKey = fmt.Sprintf("%s:%s", prefix, metaKey)
		ownerKey = fmt.Sprintf("%s:%s", prefix, ownerKey)
		seqKey = fmt.Sprintf("%s:%s", prefix, seqKey)
	}
	q.pollKeys = []string{delayKey, doingKey, metaKey, ownerKey, seqKey}
	q.delKeys = []string{delayKey, doingKey, metaKey, ownerKey}
	q.lengthKeys = []string{delayKey, doingKey}
	q.reclaimKeys = []string{delayKey, doingKey, ownerKey}
	q.addKeys = []string{delayKey, metaKey, ownerKey}
	q.ackOKKeys = []string{doingKey, metaKey, ownerKey}
	q.ackRetryKeys = []string{delayKey, doingKey, metaKey, ownerKey}
	q.ackDeadKeys = []string{doingKey, metaKey, ownerKey}
	q.heartbeatKeys = []string{doingKey, ownerKey}

	q.addScript = c.CreateScript(addDelayTaskLua)
	q.delScript = c.CreateScript(delDelayTaskLua)
	q.lengthScript = c.CreateScript(delayTaskLengthLua)
	q.pollScript = c.CreateScript(pollDelayTaskLua)
	q.reclaimScript = c.CreateScript(reclaimDelayTaskLua)
	q.ackOKScript = c.CreateScript(ackOKLua)
	q.ackRetryScript = c.CreateScript(ackRetryLua)
	q.ackDeadScript = c.CreateScript(ackDeadLua)
	q.heartbeatScript = c.CreateScript(heartbeatDelayTaskLua)

	q.startTickers()
	return q, nil
}

func (q *delayQueue) Name() string { return q.name }

func (q *delayQueue) Add(ctx context.Context, payload []byte, delay time.Duration) error {
	if !q.running.Load() {
		return ErrDelayQueueHasClosed
	}
	if delay < 0 {
		delay = 0
	}
	expireAt := nowFunc().Add(delay).UnixMilli()
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
//
// 关闭顺序：
//  1. running CAS true→false，阻止新 Add/Del 与 ticker fire；
//  2. close(exitC) 让 ticker goroutine 退出；
//  3. tickerWG.Wait——确保不再 spawn 新 worker；
//  4. 从 client.delayQueues 摘除（NewDelayQueue 同名可立即重建）；
//  5. 带超时等 worker 完成 ack（workerWG.Wait with timeout）；
//  6. cancel ctx，让仍未完成的 worker ack 快速失败短路。
//
// 等 worker 时为什么带超时？
// 在用户的 callback 内同步调 q.Close 是支持的反模式（业务侧"自我崩溃通知"）。
// 此时形成"worker 等 callback 返回 ↔ callback 等 Close 返回"环；带超时让
// Close 强制返回打破环，callback 才能继续走完。
// 即使超时漏掉的 worker，也由 client.delayWorkerWG 在 client.Close 时兜底等待，
// 仍能保证 c.cmd 释放前所有脚本调用收尾，无 race。
func (q *delayQueue) Close() error {
	if !q.running.CompareAndSwap(true, false) {
		return ErrDelayQueueHasClosed
	}
	close(q.exitC)
	q.tickerWG.Wait()
	q.c.delayQueues.Delete(q.name)
	q.waitWorkersWithTimeout(delayCloseWorkerWaitTimeout)
	q.cancel()
	return nil
}

// waitWorkersWithTimeout 等 workerWG 至多 timeout 时长。
// 主路径：所有 worker 已完成 → wg.Wait 立即返回。
// 反模式路径（callback 内调 q.Close）：wg.Wait 永等 → timeout 后强制返回。
func (q *delayQueue) waitWorkersWithTimeout(timeout time.Duration) {
	done := make(chan struct{})
	go func() {
		q.workerWG.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(timeout):
	}
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
//
// fn 出现 panic 时本 ticker 必须能继续：旧实现没有 recover，一次 panic
// 直接让 ticker goroutine 退出，从此 pollOnce/reclaimOnce 永不再触发，
// 整个 delayQueue 静默失效。这里 recover 后记录日志、保留 ticker 节奏，
// 避免被异常的 metric/handler 回调拖垮整个延迟队列。
func (q *delayQueue) runTicker(interval time.Duration, fn func()) {
	defer q.tickerWG.Done()

	safeFn := func() {
		defer func() {
			if r := recover(); r != nil {
				e(fmt.Sprintf("%s ticker panic recovered, queue=%s, %v", delayLogPrefix, q.name, r))
			}
		}()
		fn()
	}

	t := time.NewTimer(interval)
	defer t.Stop()
	for {
		select {
		case <-q.exitC:
			return
		case <-t.C:
			safeFn()
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
	visibilityDeadline := now.Add(q.spec.GetTimeout()).UnixMilli()

	ctx, cancel := q.opCtx()
	defer cancel()
	res, err := q.pollScript.Run(ctx, q.pollKeys, now.UnixMilli(), visibilityDeadline, defaultPollBatch).Slice()
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
	// 返回格式：[payload, retries, token, payload, retries, token, ...]
	for i := 0; i+2 < len(res); i += 3 {
		payload := []byte(toString(res[i]))
		retries, _ := strconv.Atoi(toString(res[i+1]))
		token := toString(res[i+2])
		// 双重跟踪：
		//   - q.workerWG  ：q.Close 带超时等之，主路径用于让 ack 在 ctx cancel 前完成；
		//   - client.delayWorkerWG：client.Close 兜底等之，避免 c.cmd race。
		// Add 在 spawn 之前完成，安全：q.Close 阻断后续 ticker fire，不与 wg.Wait race。
		q.workerWG.Add(1)
		q.c.delayWorkerWG.Add(1)
		go q.executeOne(payload, retries, token)
	}
}

// executeOne 在独立 goroutine 中执行一个 item 的 callback 与 ack。
//
// 并发模型：
//   - worker 计入 client.delayWorkerWG（client.Close 兜底等之）和 q.workerWG（q.Close 等之，带超时）；
//   - 用户 callback 跑在嵌套 goroutine 中独立调度；
//   - worker 启动 heartbeat ticker 周期续 doing 的 visibility；
//     续期失败（token 不匹配，意味着已被 reclaim 转移给他人）时停止续期，
//     callback 跑完后的 ack 同样会因 token 不匹配而 noop（fence 语义）；
//   - worker 同步等 cbDone 后尝试 ack；ack 用 token 校验防误删。
//
// recover panic 在嵌套 goroutine 中执行，视为业务失败一次。
func (q *delayQueue) executeOne(payload []byte, retries int, token string) {
	defer q.c.delayWorkerWG.Done()
	defer q.workerWG.Done()

	// heartbeat：每 visibility/3 续期，停止信号通过 close(stopHB) 发出。
	stopHB := make(chan struct{})
	hbDone := make(chan struct{})
	go q.heartbeatLoop(payload, token, stopHB, hbDone)

	cbDone := make(chan error, 1)
	go func() {
		cbDone <- q.runCallback(payload)
	}()

	err := <-cbDone

	// 停 heartbeat（callback 已完成，接下来要 ack；不需要再续期）。
	close(stopHB)
	<-hbDone

	if err == nil {
		q.ackSuccess(payload, token)
		return
	}
	// 业务失败：判断是否达到死信阈值
	// 注意：retries 是本次执行前的失败次数；本次失败后实际计数为 retries+1。
	if retries+1 >= q.spec.GetRetryTimes() {
		q.ackDead(payload, token)
		return
	}
	q.ackRetryWithBackoff(payload, token)
}

// heartbeatLoop 周期续 doing 的 visibility，直到 stopHB 关闭。
// 单次续期失败（token 不匹配）即停止——表示 lease 已被 reclaim 转移给其他 worker。
func (q *delayQueue) heartbeatLoop(payload []byte, token string, stopHB <-chan struct{}, done chan<- struct{}) {
	defer close(done)
	interval := q.heartbeatInterval()
	if interval <= 0 {
		return
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-stopHB:
			return
		case <-q.exitC:
			return
		case <-t.C:
			if !q.heartbeatOnce(payload, token) {
				return
			}
		}
	}
}

// heartbeatInterval 计算 heartbeat 间隔：visibility / heartbeatRatio，下限 minHeartbeatInterval。
func (q *delayQueue) heartbeatInterval() time.Duration {
	visibility := q.spec.GetTimeout()
	if visibility <= 0 {
		return 0
	}
	d := visibility / heartbeatRatio
	if d < minHeartbeatInterval {
		d = minHeartbeatInterval
	}
	return d
}

// heartbeatOnce 续期一次。返回 false 表示 token 已失效（不应再续期）。
func (q *delayQueue) heartbeatOnce(payload []byte, token string) bool {
	now := nowFunc()
	newDeadline := now.Add(q.spec.GetTimeout()).UnixMilli()
	ctx, cancel := q.opCtx()
	defer cancel()
	ok, err := q.heartbeatScript.Run(ctx, q.heartbeatKeys, payload, token, newDeadline).Int64()
	if err != nil {
		// 网络错误不立即放弃续期：单次失败不能可靠判断 lease 状态，下个周期重试。
		// ctx 取消（Close 路径）会让 select 通过 q.exitC 或 stopHB 自然退出。
		if !errors.Is(err, context.Canceled) {
			e(fmt.Sprintf("%s heartbeat error, payload=%q err=%v", delayLogPrefix, payload, err))
		}
		return true
	}
	return ok == 1
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

// ackSuccess 业务成功：校验 token，匹配则移出 doing + 清 meta + 清 owner。
// token 不匹配（lease 已被 reclaim 转移）时 noop，由新持有者负责。
func (q *delayQueue) ackSuccess(payload []byte, token string) {
	ctx, cancel := q.opCtx()
	defer cancel()
	if err := q.ackOKScript.Run(ctx, q.ackOKKeys, payload, token).Err(); err != nil {
		if !errors.Is(err, context.Canceled) {
			e(fmt.Sprintf("%s ack ok failed, payload=%q err=%v", delayLogPrefix, payload, err))
		}
	}
}

// ackRetryWithBackoff 业务失败但未达死信阈值：校验 token 后累加 meta 计数 + 把 item 放回 delay。
// 重新可见时间 = now + defaultRetryBackoff，与 visibility timeout 解耦。
// token 不匹配时 noop。
func (q *delayQueue) ackRetryWithBackoff(payload []byte, token string) {
	retryAt := nowFunc().Add(defaultRetryBackoff).UnixMilli()
	ctx, cancel := q.opCtx()
	defer cancel()
	if err := q.ackRetryScript.Run(ctx, q.ackRetryKeys, payload, token, retryAt).Err(); err != nil {
		if !errors.Is(err, context.Canceled) {
			e(fmt.Sprintf("%s ack retry failed, payload=%q err=%v", delayLogPrefix, payload, err))
		}
	}
}

// ackDead 业务达到死信阈值：校验 token 后清理 + 触发用户 dead-letter hook。
// 先清 Redis 状态再调 hook；hook panic 不影响清理结果。
// token 不匹配时跳过清理，但仍**不**触发死信 hook（lease 已转移，新持有者负责）。
func (q *delayQueue) ackDead(payload []byte, token string) {
	ctx, cancel := q.opCtx()
	defer cancel()
	cleared, err := q.ackDeadScript.Run(ctx, q.ackDeadKeys, payload, token).Int64()
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			e(fmt.Sprintf("%s ack dead failed, payload=%q err=%v", delayLogPrefix, payload, err))
		}
		return
	}
	if cleared != 1 {
		// token 不匹配：lease 已转移，避免与新持有者重复触发 hook。
		return
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
// reclaim 同时会清掉 owner，使旧持有者的 heartbeat / ack 因 token 不匹配而 noop。
func (q *delayQueue) reclaimOnce() {
	if !q.running.Load() {
		return
	}
	now := nowFunc()
	ctx, cancel := q.opCtx()
	defer cancel()
	count, err := q.reclaimScript.Run(ctx, q.reclaimKeys, now.UnixMilli(), defaultPollBatch).Int64()
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
