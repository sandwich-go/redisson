package redisson

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type pubSubQueuedMessage struct {
	message  Message
	enqueued time.Duration
	bytes    int64
}

type pubSubOutputRecord struct {
	enqueued   atomic.Int64
	totalBytes atomic.Int64
}

// 消息转发完全复用 unboundedChan。固定长度元数据仅用于观测未消费的 Out 缓冲，
// 不持有消息、不增加消息队列或转发 goroutine。
type pubSubMonitorQueue struct {
	ctx       context.Context
	messages  *unboundedChan[pubSubQueuedMessage, Message]
	out       <-chan Message
	now       func() time.Duration
	onDequeue func(time.Duration)
	onClose   func(int)

	closed atomic.Bool

	sendersMu    sync.Mutex
	senders      map[*pubSubQueueSender]struct{}
	retiredBytes int64
	_            [128]byte // 将接收侧只读状态与转发侧热写入隔开。

	oldest atomic.Int64
	sent   atomic.Uint64
	output []pubSubOutputRecord

	// 以下累计量只有 unboundedChan 的转发 goroutine 修改。
	totalBytes   int64
	totalSent    uint64
	dequeueEvery uint32
	dequeueNext  uint32
}

func newPubSubMonitorQueue(ctx context.Context, capacity int, now func() time.Duration, onDequeue func(time.Duration), onClose func(int), sampleEvery uint32) *pubSubMonitorQueue {
	q := &pubSubMonitorQueue{
		ctx: ctx, now: now, onDequeue: onDequeue, onClose: onClose,
		output: make([]pubSubOutputRecord, capacity+2), dequeueEvery: max(sampleEvery, 1),
		senders: make(map[*pubSubQueueSender]struct{}),
	}
	q.messages = newUnboundedChanMapped(q.ctx, capacity, capacity, capacity,
		func(value pubSubQueuedMessage) Message { return value.message },
		&unboundedChanHooks[pubSubQueuedMessage]{onBuffer: q.onBuffer, onWrite: q.onWrite, onClose: q.finish})
	q.out = q.messages.Out
	return q
}

func messageBytes(message Message) int64 {
	return int64(len(message.Channel) + len(message.Pattern) + len(message.Message))
}

// rueidis.Receive 在接收循环内串行执行回调。累计值只有本接收任务写入，
// 快照只读 admitted；后置填充隔开不同接收任务的热数据。
type pubSubQueueSender struct {
	queue    *pubSubMonitorQueue
	ctx      context.Context
	done     <-chan struct{}
	in       chan<- pubSubQueuedMessage
	now      func() time.Duration
	total    int64
	admitted atomic.Int64
	_        [128]byte
}

func (q *pubSubMonitorQueue) newSender(ctx context.Context) *pubSubQueueSender {
	sender := &pubSubQueueSender{queue: q, ctx: ctx, done: ctx.Done(), in: q.messages.In, now: q.now}
	q.sendersMu.Lock()
	q.senders[sender] = struct{}{}
	q.sendersMu.Unlock()
	return sender
}

// 与 push 在同一接收任务中串行调用，必须等 Receive 返回后才能释放。
func (s *pubSubQueueSender) release() {
	q := s.queue
	q.sendersMu.Lock()
	if _, exists := q.senders[s]; exists {
		q.retiredBytes += s.admitted.Load()
		delete(q.senders, s)
	}
	q.sendersMu.Unlock()
}

// ctx 是队列的生命周期 context 或子订阅 context，任一层取消都应结束等待入队。
func (s *pubSubQueueSender) push(message Message) bool {
	if s.queue.closed.Load() || s.ctx.Err() != nil {
		return false
	}
	value := pubSubQueuedMessage{message: message, enqueued: s.now(), bytes: messageBytes(message)}
	bytes := value.bytes
	s.total += bytes
	s.admitted.Store(s.total)
	select {
	case s.in <- value:
		return true
	default:
	}
	select {
	case s.in <- value:
		return true
	case <-s.done:
		s.total -= bytes
		s.admitted.Store(s.total)
		return false
	}
}

func (q *pubSubMonitorQueue) admittedBytes() int64 {
	q.sendersMu.Lock()
	defer q.sendersMu.Unlock()
	total := q.retiredBytes
	for sender := range q.senders {
		total += sender.admitted.Load()
	}
	return total
}

func (q *pubSubMonitorQueue) onBuffer(value pubSubQueuedMessage) {
	q.oldest.Store(int64(value.enqueued))
}

func (q *pubSubMonitorQueue) onWrite(value pubSubQueuedMessage, buffer *ringBuffer[pubSubQueuedMessage]) {
	bytes := value.bytes
	record := &q.output[q.totalSent%uint64(len(q.output))]
	q.totalBytes += bytes
	record.enqueued.Store(int64(value.enqueued))
	record.totalBytes.Store(q.totalBytes)
	q.totalSent++
	q.sent.Store(q.totalSent)
	if !buffer.IsEmpty() {
		q.oldest.Store(int64(buffer.Peek().enqueued))
	}
	if q.onDequeue != nil {
		if q.dequeueNext == 0 {
			q.onDequeue(q.now() - value.enqueued)
			q.dequeueNext = q.dequeueEvery
		}
		q.dequeueNext--
	}
}

func (q *pubSubMonitorQueue) finish(dropped int) {
	q.closed.Store(true)
	if q.onClose != nil {
		q.onClose(dropped)
	}
}

func (q *pubSubMonitorQueue) snapshot() (int, int64, time.Duration, bool) {
	if q.closed.Load() {
		return 0, 0, 0, true
	}
	// 累计入队字节减去已消费消息的前缀和，覆盖 In、环形缓冲和 Out。
	// 多保留两个输出记录，保护最老未读消息之前的前缀及尚未发布的写入。
	var depth int
	var bytes int64
	var age time.Duration
	for attempt := 0; attempt < 2; attempt++ {
		sent := q.sent.Load()
		buffered := len(q.out)
		depth = len(q.messages.In) + q.messages.BufLen() + buffered
		bytes = q.admittedBytes()
		oldest := time.Duration(q.oldest.Load())
		tracked := min(uint64(buffered), sent)
		consumed := sent - tracked
		if consumed > 0 {
			bytes -= q.output[(consumed-1)%uint64(len(q.output))].totalBytes.Load()
		}
		if tracked > 0 {
			oldest = time.Duration(q.output[consumed%uint64(len(q.output))].enqueued.Load())
		}
		age = 0
		if tracked > 0 || q.messages.BufLen() > 0 {
			age = max(q.now()-oldest, 0)
		}
		if q.sent.Load() == sent {
			break
		}
	}
	if q.closed.Load() {
		return 0, 0, 0, true
	}
	if depth == 0 {
		return 0, 0, 0, false
	}
	return depth, max(bytes, 0), age, false
}
