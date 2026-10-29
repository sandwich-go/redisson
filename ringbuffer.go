package redisson

import (
	"context"
	"errors"
	"sync"
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
	r.buf[r.r] = t
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

// Release 仅由已停止转发的持有方调用，释放丢弃元素及扩容存储。
func (r *ringBuffer[T]) Release() {
	r.buf = nil
	r.size, r.r, r.w = 0, 0, 0
}

func (r *ringBuffer[T]) Reset() {
	r.r = 0
	r.w = 0
	r.size = r.initialSize
	r.buf = make([]T, r.initialSize)
}

// unboundedChan is an unbounded chan.
// Writers bind a sender so shutdown can wait for accepted input operations.
// Out supports multiple readers; CloseInput stops writers and drains accepted data.
type unboundedChan[T, Out any] struct {
	bufCount int64
	In       <-chan T       // occupancy inspection; writes use newSender
	Out      <-chan Out     // channel for read
	buffer   *ringBuffer[T] // buffer
	in       chan T

	_                [128]byte
	stopped          atomic.Bool
	stopping         chan struct{}
	stopOnce         sync.Once
	closeInputOnce   sync.Once
	producersMu      sync.Mutex
	producersChanged *sync.Cond
	producers        map[*unboundedChanSender[T, Out]]struct{}
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
// Each writer binds newSender; Out supports multiple readers.
func newUnboundedChan[T any](ctx context.Context, initCapacity int) *unboundedChan[T, T] {
	return newUnboundedChanSize[T](ctx, initCapacity, initCapacity, initCapacity)
}

// newUnboundedChanSize is like newUnboundedChan but you can set initial capacity for In, Out, Buffer.
func newUnboundedChanSize[T any](ctx context.Context, initInCapacity, initOutCapacity, initBufCapacity int) *unboundedChan[T, T] {
	return newUnboundedChanMapped(ctx, initInCapacity, initOutCapacity, initBufCapacity, func(value T) T { return value }, nil)
}

// 可选观测在原有转发 goroutine 内执行，不增加队列或转发 goroutine。
type unboundedChanHooks[T any] struct {
	onBuffer func(T, *ringBuffer[T])
	onWrite  func(T, *ringBuffer[T], bool)
	onClose  func(int)
}

// 入队元素可携带本机观测信息，输出转换后仍保持原有消息类型和缓冲大小。
func newUnboundedChanMapped[T, Out any](ctx context.Context, inCapacity, outCapacity, bufferCapacity int, project func(T) Out, hooks *unboundedChanHooks[T]) *unboundedChan[T, Out] {
	in := make(chan T, inCapacity)
	out := make(chan Out, outCapacity)
	ch := &unboundedChan[T, Out]{In: in, in: in, Out: out, buffer: newRingBuffer[T](bufferCapacity),
		stopping: make(chan struct{}), producers: make(map[*unboundedChanSender[T, Out]]struct{}),
	}
	ch.producersChanged = sync.NewCond(&ch.producersMu)
	go process(ctx, in, out, ch, project, hooks)
	return ch
}

func process[T, Out any](ctx context.Context, in chan T, out chan Out, ch *unboundedChan[T, Out], project func(T) Out, hooks *unboundedChanHooks[T]) {
	defer func() {
		ch.stopProducers()
		dropped := ch.buffer.Len()
	collect:
		for {
			select {
			case _, ok := <-in:
				if !ok {
					break collect
				}
				dropped++
			default:
				break collect
			}
		}
		ch.buffer.Release()
		atomic.StoreInt64(&ch.bufCount, 0)
		close(out)
		if hooks != nil && hooks.onClose != nil {
			hooks.onClose(dropped)
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
					hooks.onWrite(value, ch.buffer, false)
				}
			case <-ctx.Done():
				return
			}
		}
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
				if hooks != nil {
					hooks.onBuffer(val, ch.buffer)
				}
			} else {
				select {
				case out <- project(val):
					if hooks != nil {
						hooks.onWrite(val, ch.buffer, true)
					}
					continue
				default:
				}
				ch.buffer.Write(val)
				atomic.AddInt64(&ch.bufCount, 1)
				if hooks != nil {
					hooks.onBuffer(val, ch.buffer)
				}
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
					if hooks != nil {
						hooks.onBuffer(val, ch.buffer)
					}
				case out <- project(value):
					ch.buffer.Pop()
					atomic.AddInt64(&ch.bufCount, -1)
					if hooks != nil {
						hooks.onWrite(value, ch.buffer, false)
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
