// Package buffer 提供 redisson 内部使用的环形缓冲与无界 channel 实现。
//
// 该包是 redisson 的内部实现细节，外部用户不应直接依赖。
package buffer

import (
	"errors"
)

// ErrIsEmpty 是 Ring.Read 在缓冲为空时返回的错误。
var ErrIsEmpty = errors.New("ring buffer is empty")

// Ring 是一个泛型环形缓冲：
//   - 写满时自动扩容（不会阻塞）；
//   - 非线程安全，需要外部同步原语保护。
type Ring[T any] struct {
	buf         []T
	initialSize int
	size        int
	r           int // read pointer
	w           int // write pointer
}

// NewRing 创建一个初始容量为 initialSize 的 Ring。
// initialSize 必须 > 0；为 1 时会被规整为 2，避免初始即满。
func NewRing[T any](initialSize int) *Ring[T] {
	if initialSize <= 0 {
		panic("buffer: initial size must be greater than zero")
	}
	if initialSize == 1 {
		initialSize = 2
	}
	return &Ring[T]{
		buf:         make([]T, initialSize),
		initialSize: initialSize,
		size:        initialSize,
	}
}

// Read 读取一个元素；若空则返回 ErrIsEmpty。
func (r *Ring[T]) Read() (T, error) {
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

// Pop 读取并消费一个元素；空时 panic（ErrIsEmpty）。
func (r *Ring[T]) Pop() T {
	v, err := r.Read()
	if errors.Is(err, ErrIsEmpty) {
		panic(ErrIsEmpty.Error())
	}
	return v
}

// Peek 查看队首元素但不消费；空时 panic。
func (r *Ring[T]) Peek() T {
	if r.r == r.w {
		panic(ErrIsEmpty.Error())
	}
	return r.buf[r.r]
}

// Write 写入一个元素；写满时自动扩容。
func (r *Ring[T]) Write(v T) {
	r.buf[r.w] = v
	r.w++
	if r.w == r.size {
		r.w = 0
	}
	if r.w == r.r { // full
		r.grow()
	}
}

func (r *Ring[T]) grow() {
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

// IsEmpty 是否空。
func (r *Ring[T]) IsEmpty() bool { return r.r == r.w }

// Capacity 当前底层缓冲容量。
func (r *Ring[T]) Capacity() int { return r.size }

// Size 与 Capacity 同义；为 unboundedChan 内部缩容判断使用。
func (r *Ring[T]) Size() int { return r.size }

// InitialSize 创建时的初始容量；扩容后的 Reset 会缩回到这个值。
func (r *Ring[T]) InitialSize() int { return r.initialSize }

// Len 当前已写入未读取的元素数。
func (r *Ring[T]) Len() int {
	if r.r == r.w {
		return 0
	}
	if r.w > r.r {
		return r.w - r.r
	}
	return r.size - r.r + r.w
}

// Reset 清空并缩回初始容量。
func (r *Ring[T]) Reset() {
	r.r = 0
	r.w = 0
	r.size = r.initialSize
	r.buf = make([]T, r.initialSize)
}
