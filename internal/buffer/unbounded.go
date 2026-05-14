package buffer

import (
	"context"
	"sync/atomic"
)

// UnboundedChan 是无界 channel：
//   - In  支持非阻塞写入（多写者）；
//   - Out 支持读取（多读者）；
//   - 内部用 Ring 缓冲超额数据，元素 FIFO；
//   - 关闭 In 后会 drain 后再关闭 Out。
type UnboundedChan[T any] struct {
	bufCount int64
	In       chan<- T
	Out      <-chan T
	buffer   *Ring[T]
}

// Len 总元素数（in chan + buffer + out chan）。
func (c UnboundedChan[T]) Len() int {
	return len(c.In) + c.BufLen() + len(c.Out)
}

// BufLen 仅计 buffer 中的元素数。
func (c UnboundedChan[T]) BufLen() int {
	return int(atomic.LoadInt64(&c.bufCount))
}

// NewUnbounded 创建一个 In/Out/Buffer 容量都相同的 UnboundedChan。
func NewUnbounded[T any](ctx context.Context, initCapacity int) *UnboundedChan[T] {
	return NewUnboundedSize[T](ctx, initCapacity, initCapacity, initCapacity)
}

// NewUnboundedSize 允许分别指定 In/Out/Buffer 容量。
func NewUnboundedSize[T any](ctx context.Context, initInCapacity, initOutCapacity, initBufCapacity int) *UnboundedChan[T] {
	in := make(chan T, initInCapacity)
	out := make(chan T, initOutCapacity)
	ch := UnboundedChan[T]{In: in, Out: out, buffer: NewRing[T](initBufCapacity)}
	go process(ctx, in, out, &ch)
	return &ch
}

func process[T any](ctx context.Context, in, out chan T, ch *UnboundedChan[T]) {
	defer close(out)
	drain := func() {
		for !ch.buffer.IsEmpty() {
			select {
			case out <- ch.buffer.Pop():
				atomic.AddInt64(&ch.bufCount, -1)
			case <-ctx.Done():
				return
			}
		}
		ch.buffer.Reset()
		atomic.StoreInt64(&ch.bufCount, 0)
	}
	for {
		select {
		case <-ctx.Done():
			return
		case val, ok := <-in:
			if !ok {
				drain()
				return
			}
			if atomic.LoadInt64(&ch.bufCount) > 0 {
				ch.buffer.Write(val)
				_ = atomic.AddInt64(&ch.bufCount, 1)
			} else {
				select {
				case out <- val:
					continue
				default:
				}
				ch.buffer.Write(val)
				_ = atomic.AddInt64(&ch.bufCount, 1)
			}
			for !ch.buffer.IsEmpty() {
				select {
				case <-ctx.Done():
					return
				case val, ok := <-in:
					if !ok {
						drain()
						return
					}
					ch.buffer.Write(val)
					_ = atomic.AddInt64(&ch.bufCount, 1)
				case out <- ch.buffer.Peek():
					ch.buffer.Pop()
					atomic.AddInt64(&ch.bufCount, -1)
					if ch.buffer.IsEmpty() && ch.buffer.Size() > ch.buffer.InitialSize() {
						ch.buffer.Reset()
						atomic.StoreInt64(&ch.bufCount, 0)
					}
				}
			}
		}
	}
}
