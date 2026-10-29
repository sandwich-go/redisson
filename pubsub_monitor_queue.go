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
	sender   *pubSubQueueSender
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

	oldest    atomic.Int64
	ringCount atomic.Int64
	sent      atomic.Uint64
	output    []pubSubOutputRecord

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

// 输入历史只保存时间戳，最多覆盖输入容量及两个交接中的元素。
// produced 由接收任务写，received 由转发任务写；完成的任务保留到输入被接走。
type pubSubQueueSender struct {
	queue         *pubSubMonitorQueue
	writer        *unboundedChanSender[pubSubQueuedMessage, Message]
	now           func() time.Duration
	total         int64
	admitted      atomic.Int64
	input         []atomic.Int64
	totalProduced uint64
	inputNext     int
	produced      atomic.Uint64
	finished      atomic.Bool
	_             [128]byte
	totalReceived uint64
	received      atomic.Uint64
	_             [128]byte
}

func (q *pubSubMonitorQueue) newSender(ctx context.Context) *pubSubQueueSender {
	sender := &pubSubQueueSender{
		queue: q, writer: q.messages.newSender(ctx), now: q.now,
		input: make([]atomic.Int64, cap(q.messages.In)+2),
	}
	q.sendersMu.Lock()
	if !q.closed.Load() {
		q.senders[sender] = struct{}{}
	}
	q.sendersMu.Unlock()
	return sender
}

// Receive 返回后不再有回调；输入尚未接走时，仍需保留其队首观测。
func (s *pubSubQueueSender) release() {
	s.finished.Store(true)
	s.queue.retire(s)
	s.writer.release()
}

func (q *pubSubMonitorQueue) retire(sender *pubSubQueueSender) {
	q.sendersMu.Lock()
	if _, exists := q.senders[sender]; exists && sender.finished.Load() && sender.received.Load() == sender.produced.Load() {
		q.retiredBytes += sender.admitted.Load()
		delete(q.senders, sender)
	}
	q.sendersMu.Unlock()
}

func (s *pubSubQueueSender) push(message Message) bool {
	if !s.writer.begin() {
		s.writer.end()
		return false
	}
	defer s.writer.end()
	value := pubSubQueuedMessage{message: message, enqueued: s.now(), sender: s}
	bytes := messageBytes(message)
	index := s.inputNext
	s.input[index].Store(int64(value.enqueued))
	s.inputNext++
	if s.inputNext == len(s.input) {
		s.inputNext = 0
	}
	s.totalProduced++
	s.produced.Store(s.totalProduced)
	s.total += bytes
	s.admitted.Store(s.total)
	if s.writer.commit(value) {
		return true
	}
	s.totalProduced--
	s.inputNext = index
	s.produced.Store(s.totalProduced)
	s.total -= bytes
	s.admitted.Store(s.total)
	return false
}

func (q *pubSubMonitorQueue) admissions() (bytes int64, pending uint64, oldest time.Duration, hasOldest bool) {
	q.sendersMu.Lock()
	defer q.sendersMu.Unlock()
	bytes = q.retiredBytes
	for sender := range q.senders {
		bytes += sender.admitted.Load()
		for attempt := 0; attempt < 2; attempt++ {
			received := sender.received.Load()
			produced := sender.produced.Load()
			if produced <= received {
				break
			}
			at := time.Duration(sender.input[received%uint64(len(sender.input))].Load())
			if sender.received.Load() != received && attempt == 0 {
				continue
			}
			pending += produced - received
			if !hasOldest || at < oldest {
				oldest, hasOldest = at, true
			}
			break
		}
	}
	return
}

func (q *pubSubMonitorQueue) admittedBytes() int64 {
	bytes, _, _, _ := q.admissions()
	return bytes
}

func (q *pubSubMonitorQueue) acceptInput(sender *pubSubQueueSender) {
	if sender != nil {
		sender.totalReceived++
		sender.received.Store(sender.totalReceived)
		if sender.finished.Load() {
			q.retire(sender)
		}
	}
}

func (q *pubSubMonitorQueue) onBuffer(value pubSubQueuedMessage, buffer *ringBuffer[pubSubQueuedMessage]) {
	// 先发布环形阶段的队首和深度，再核销输入，交接时允许保守重叠但不能漏记。
	if buffer.Len() == 1 {
		q.oldest.Store(int64(value.enqueued))
	}
	q.ringCount.Store(int64(buffer.Len()))
	q.acceptInput(value.sender)
}

func (q *pubSubMonitorQueue) onWrite(value pubSubQueuedMessage, buffer *ringBuffer[pubSubQueuedMessage], fromInput bool) {
	bytes := messageBytes(value.message)
	record := &q.output[q.totalSent%uint64(len(q.output))]
	q.totalBytes += bytes
	record.enqueued.Store(int64(value.enqueued))
	record.totalBytes.Store(q.totalBytes)
	q.totalSent++
	q.sent.Store(q.totalSent)
	if fromInput {
		q.acceptInput(value.sender)
	} else {
		if !buffer.IsEmpty() {
			q.oldest.Store(int64(buffer.Peek().enqueued))
		}
		q.ringCount.Store(int64(buffer.Len()))
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
	q.ringCount.Store(0)
	q.sendersMu.Lock()
	q.senders = nil
	q.retiredBytes = 0
	q.sendersMu.Unlock()
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
		var input uint64
		var oldest time.Duration
		var hasOldest bool
		// 先读输入，再读下游；若输入已被核销，必须看得到此前发布的阶段状态。
		bytes, input, oldest, hasOldest = q.admissions()
		ring := int(q.ringCount.Load())
		sent := q.sent.Load()
		buffered := len(q.out)
		depth = int(input) + ring + buffered
		include := func(at time.Duration) {
			if !hasOldest || at < oldest {
				oldest, hasOldest = at, true
			}
		}
		if ring > 0 {
			include(time.Duration(q.oldest.Load()))
		}
		tracked := min(uint64(buffered), sent)
		consumed := sent - tracked
		if consumed > 0 {
			bytes -= q.output[(consumed-1)%uint64(len(q.output))].totalBytes.Load()
		}
		if tracked > 0 {
			include(time.Duration(q.output[consumed%uint64(len(q.output))].enqueued.Load()))
		}
		age = 0
		if hasOldest {
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
