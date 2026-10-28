package redisson

import (
	"context"
	"errors"
	"sync/atomic"
)

var ErrIsEmpty = errors.New("ring buffer is empty")

// ringBuffer is a ring buffer for common types.
// It never is full and always grows if it will be full.
// It is not thread-safe(goroutine-safe) so you must use the lock-like synchronization primitive to use it in multiple writers and multiple readers.
type ringBuffer[T any] struct {
	buf         []T
	initialSize int
	size        int
	r           int // read pointer
	w           int // write pointer
}

func newRingBuffer[T any](initialSize int) *ringBuffer[T] {
	if initialSize <= 0 {
		panic("initial size must be great than zero")
	}
	// initial size must >= 2
	if initialSize == 1 {
		initialSize = 2
	}

	return &ringBuffer[T]{
		buf:         make([]T, initialSize),
		initialSize: initialSize,
		size:        initialSize,
	}
}

func (r *ringBuffer[T]) Read() (T, error) {
	var t T
	if r.r == r.w {
		return t, ErrIsEmpty
	}

	v := r.buf[r.r]
	r.r++
	if r.r == r.size {
		r.r = 0
	}

	return v, nil
}

func (r *ringBuffer[T]) Pop() T {
	v, err := r.Read()
	if errors.Is(err, ErrIsEmpty) { // Empty
		panic(ErrIsEmpty.Error())
	}

	return v
}

func (r *ringBuffer[T]) Peek() T {
	if r.r == r.w { // Empty
		panic(ErrIsEmpty.Error())
	}

	v := r.buf[r.r]
	return v
}

func (r *ringBuffer[T]) Write(v T) {
	r.buf[r.w] = v
	r.w++

	if r.w == r.size {
		r.w = 0
	}

	if r.w == r.r { // full
		r.grow()
	}
}

func (r *ringBuffer[T]) grow() {
	var size int
	if r.size < 1024 {
		size = r.size * 2
	} else {
		size = r.size + r.size/4
	}

	buf := make([]T, size)

	copy(buf[0:], r.buf[r.r:])
	copy(buf[r.size-r.r:], r.buf[0:r.r])

	r.r = 0
	r.w = r.size
	r.size = size
	r.buf = buf
}

func (r *ringBuffer[T]) IsEmpty() bool {
	return r.r == r.w
}

// Capacity returns the size of the underlying buffer.
func (r *ringBuffer[T]) Capacity() int {
	return r.size
}

func (r *ringBuffer[T]) Len() int {
	if r.r == r.w {
		return 0
	}

	if r.w > r.r {
		return r.w - r.r
	}

	return r.size - r.r + r.w
}

func (r *ringBuffer[T]) Reset() {
	r.r = 0
	r.w = 0
	r.size = r.initialSize
	r.buf = make([]T, r.initialSize)
}

// unboundedChan is an unbounded chan.
// In is used to write without blocking, which supports multiple writers.
// and Out is used to read, which supports multiple readers.
// You can close the in channel if you want.
type unboundedChan[T, Out any] struct {
	bufCount int64
	In       chan<- T       // channel for write
	Out      <-chan Out     // channel for read
	buffer   *ringBuffer[T] // buffer
}

// Len 所有待读取的数据的长度
func (c *unboundedChan[T, Out]) Len() int {
	return len(c.In) + c.BufLen() + len(c.Out)
}

// BufLen 获取缓存中的数据的长度，不包含外发Out channel中数据的长度
func (c *unboundedChan[T, Out]) BufLen() int {
	return int(atomic.LoadInt64(&c.bufCount))
}

// newUnboundedChan creates the unbounded chan.
// in is used to write without blocking, which supports multiple writers.
// and out is used to read, which supports multiple readers.
// You can close the in channel if you want.
func newUnboundedChan[T any](ctx context.Context, initCapacity int) *unboundedChan[T, T] {
	return newUnboundedChanSize[T](ctx, initCapacity, initCapacity, initCapacity)
}

// newUnboundedChanSize is like newUnboundedChan but you can set initial capacity for In, Out, Buffer.
func newUnboundedChanSize[T any](ctx context.Context, initInCapacity, initOutCapacity, initBufCapacity int) *unboundedChan[T, T] {
	return newUnboundedChanMapped(ctx, initInCapacity, initOutCapacity, initBufCapacity, func(value T) T { return value }, nil)
}

// 可选观测在原有转发 goroutine 内执行，不增加队列或转发 goroutine。
type unboundedChanHooks[T any] struct {
	onBuffer func(T)
	onWrite  func(T, *ringBuffer[T])
	onClose  func(int)
}

// 入队元素可携带本机观测信息，输出转换后仍保持原有消息类型和缓冲大小。
func newUnboundedChanMapped[T, Out any](ctx context.Context, inCapacity, outCapacity, bufferCapacity int, project func(T) Out, hooks *unboundedChanHooks[T]) *unboundedChan[T, Out] {
	in := make(chan T, inCapacity)
	out := make(chan Out, outCapacity)
	ch := &unboundedChan[T, Out]{In: in, Out: out, buffer: newRingBuffer[T](bufferCapacity)}
	go process(ctx, in, out, ch, project, hooks)
	return ch
}

func process[T, Out any](ctx context.Context, in chan T, out chan Out, ch *unboundedChan[T, Out], project func(T) Out, hooks *unboundedChanHooks[T]) {
	defer func() {
		close(out)
		if hooks != nil {
			hooks.onClose(len(in) + ch.buffer.Len())
		}
	}()
	drain := func() {
		for !ch.buffer.IsEmpty() {
			value := ch.buffer.Peek()
			select {
			case out <- project(value):
				ch.buffer.Pop()
				atomic.AddInt64(&ch.bufCount, -1)
				if hooks != nil {
					hooks.onWrite(value, ch.buffer)
				}
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
				atomic.AddInt64(&ch.bufCount, 1)
			} else {
				select {
				case out <- project(val):
					if hooks != nil {
						hooks.onWrite(val, ch.buffer)
					}
					continue
				default:
				}
				if hooks != nil {
					hooks.onBuffer(val)
				}
				ch.buffer.Write(val)
				atomic.AddInt64(&ch.bufCount, 1)
			}
			for !ch.buffer.IsEmpty() {
				value := ch.buffer.Peek()
				select {
				case <-ctx.Done():
					return
				case val, ok := <-in:
					if !ok {
						drain()
						return
					}
					ch.buffer.Write(val)
					atomic.AddInt64(&ch.bufCount, 1)
				case out <- project(value):
					ch.buffer.Pop()
					atomic.AddInt64(&ch.bufCount, -1)
					if hooks != nil {
						hooks.onWrite(value, ch.buffer)
					}
					if ch.buffer.IsEmpty() && ch.buffer.size > ch.buffer.initialSize {
						ch.buffer.Reset()
						atomic.StoreInt64(&ch.bufCount, 0)
					}
				}
			}
		}
	}
}
