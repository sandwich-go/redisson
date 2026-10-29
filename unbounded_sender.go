package redisson

import (
	"context"
	"sync/atomic"
)

// 每个接收任务串行使用自己的发送者；关闭队列只等待正在入队的调用。
type unboundedChanSender[T, Out any] struct {
	queue *unboundedChan[T, Out]
	ctx   context.Context
	done  <-chan struct{}
	in    chan<- T
	busy  atomic.Bool
}

func (q *unboundedChan[T, Out]) newSender(ctx context.Context) *unboundedChanSender[T, Out] {
	sender := &unboundedChanSender[T, Out]{queue: q, ctx: ctx, done: ctx.Done(), in: q.in}
	q.producersMu.Lock()
	if !q.stopped.Load() {
		q.producers[sender] = struct{}{}
	}
	q.producersMu.Unlock()
	return sender
}

func (s *unboundedChanSender[T, Out]) begin() bool {
	// 必须先登记 busy 再检查停止标记，关闭方才不会漏过已进入的调用。
	s.busy.Store(true)
	return !s.queue.stopped.Load() && s.ctx.Err() == nil
}

func (s *unboundedChanSender[T, Out]) end() {
	s.busy.Store(false)
	if s.queue.stopped.Load() {
		s.queue.signalProducerExit()
	}
}

func (q *unboundedChan[T, Out]) signalProducerExit() {
	q.producersMu.Lock()
	q.producersChanged.Broadcast()
	q.producersMu.Unlock()
}

func (s *unboundedChanSender[T, Out]) commit(value T) bool {
	select {
	case s.in <- value:
		return true
	default:
	}
	select {
	case s.in <- value:
		return true
	case <-s.done:
		return false
	case <-s.queue.stopping:
		return false
	}
}

func (s *unboundedChanSender[T, Out]) send(value T) bool {
	if !s.begin() {
		s.end()
		return false
	}
	defer s.end()
	return s.commit(value)
}

// 调用方须等本发送任务结束后再释放。
func (s *unboundedChanSender[T, Out]) release() {
	q := s.queue
	q.producersMu.Lock()
	delete(q.producers, s)
	q.producersChanged.Broadcast()
	q.producersMu.Unlock()
}

func (q *unboundedChan[T, Out]) stopProducers() {
	q.stopOnce.Do(func() {
		q.stopped.Store(true)
		close(q.stopping)
		q.producersMu.Lock()
		defer q.producersMu.Unlock()
		for {
			busy := false
			for sender := range q.producers {
				if sender.busy.Load() {
					busy = true
					break
				}
			}
			if !busy {
				q.producers = nil
				return
			}
			q.producersChanged.Wait()
		}
	})
}

// CloseInput 停止生产者，再关闭输入，让转发任务排空已接受的消息。
func (q *unboundedChan[T, Out]) CloseInput() {
	q.closeInputOnce.Do(func() {
		q.stopProducers()
		close(q.in)
	})
}
