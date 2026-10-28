package redisson

import (
	"context"
	"sync"
	"time"
)

type pubSubQueuedMessage struct {
	message  Message
	enqueued time.Duration
	next     *pubSubQueuedMessage
}

var pubSubQueuedMessagePool = sync.Pool{New: func() any { return new(pubSubQueuedMessage) }}

// 观测队列替代普通缓冲，不再增加一份消息队列；无界策略保持不变。
// 输出在消费者接收后才移除队首，以统计等待消费的消息，而不是只统计中间缓冲。
type pubSubMonitorQueue struct {
	mu         sync.Mutex
	head, tail *pubSubQueuedMessage
	depth      int
	bytes      int64
	closed     bool
	out        chan Message
	wake       chan struct{}
	now        func() time.Duration
	onDequeue  func(time.Duration)
	onClose    func(int)
}

func newPubSubMonitorQueue(ctx context.Context, now func() time.Duration, onDequeue func(time.Duration), onClose func(int)) *pubSubMonitorQueue {
	q := &pubSubMonitorQueue{out: make(chan Message), wake: make(chan struct{}, 1), now: now, onDequeue: onDequeue, onClose: onClose}
	go q.run(ctx)
	return q
}

func messageBytes(message Message) int64 {
	return int64(len(message.Channel) + len(message.Pattern) + len(message.Message))
}

func (q *pubSubMonitorQueue) push(message Message) bool {
	node := pubSubQueuedMessagePool.Get().(*pubSubQueuedMessage)
	node.message = message
	q.mu.Lock()
	if q.closed {
		q.mu.Unlock()
		releasePubSubQueuedMessage(node)
		return false
	}
	node.enqueued = q.now()
	if q.tail != nil {
		q.tail.next = node
	}
	if q.head == nil {
		q.head = node
	}
	q.tail = node
	q.depth++
	q.bytes += messageBytes(message)
	q.mu.Unlock()
	select {
	case q.wake <- struct{}{}:
	default:
	}
	return true
}

func releasePubSubQueuedMessage(node *pubSubQueuedMessage) {
	*node = pubSubQueuedMessage{}
	pubSubQueuedMessagePool.Put(node)
}

func (q *pubSubMonitorQueue) run(ctx context.Context) {
	defer func() {
		q.mu.Lock()
		q.closed = true
		node, dropped := q.head, q.depth
		q.head, q.tail, q.depth, q.bytes = nil, nil, 0, 0
		q.mu.Unlock()
		for node != nil {
			next := node.next
			releasePubSubQueuedMessage(node)
			node = next
		}
		close(q.out)
		if q.onClose != nil {
			q.onClose(dropped)
		}
	}()
	for {
		q.mu.Lock()
		node := q.head
		q.mu.Unlock()
		if node == nil {
			select {
			case <-ctx.Done():
				return
			case <-q.wake:
				continue
			}
		}
		select {
		case <-ctx.Done():
			return
		case q.out <- node.message:
			q.mu.Lock()
			q.head = node.next
			if q.head == nil {
				q.tail = nil
			}
			q.depth--
			q.bytes -= messageBytes(node.message)
			waited := q.now() - node.enqueued
			q.mu.Unlock()
			releasePubSubQueuedMessage(node)
			if q.onDequeue != nil {
				q.onDequeue(waited)
			}
		}
	}
}

func (q *pubSubMonitorQueue) snapshot() (int, int64, time.Duration, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.head == nil {
		return 0, 0, 0, q.closed
	}
	return q.depth, q.bytes, q.now() - q.head.enqueued, q.closed
}
