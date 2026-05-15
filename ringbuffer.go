package redisson

import (
	"context"

	"github.com/sandwich-go/redisson/internal/buffer"
)

// 主包内部 ringBuffer / unboundedChan 是 internal/buffer 类型的小写别名,
// 用 lowerCamelCase 命名表达"包内私有"语义,实现仍在 internal/buffer。

// ErrIsEmpty 在 ring buffer 为空时由 Read 返回。
var ErrIsEmpty = buffer.ErrIsEmpty

type ringBuffer[T any] = buffer.Ring[T]

type unboundedChan[T any] = buffer.UnboundedChan[T]

func newRingBuffer[T any](initialSize int) *ringBuffer[T] {
	return buffer.NewRing[T](initialSize)
}

func newUnboundedChan[T any](ctx context.Context, initCapacity int) *unboundedChan[T] {
	return buffer.NewUnbounded[T](ctx, initCapacity)
}

//nolint:unused // 保留对外形式,供将来拆分非默认尺寸场景
func newUnboundedChanSize[T any](ctx context.Context, in, out, buf int) *unboundedChan[T] {
	return buffer.NewUnboundedSize[T](ctx, in, out, buf)
}
