package redisson

import (
	"context"

	"github.com/sandwich-go/redisson/internal/buffer"
)

// 兼容层：Ring 与 UnboundedChan 已迁至 internal/buffer。
// 主包内部仍以历史名称 ringBuffer / unboundedChan 引用，避免大面积改动。

// ErrIsEmpty 在 ring buffer 为空时由 Read 返回。
var ErrIsEmpty = buffer.ErrIsEmpty

// ringBuffer 仍为内部别名，类型由 internal/buffer 提供。
type ringBuffer[T any] = buffer.Ring[T]

// unboundedChan 仍为内部别名。
type unboundedChan[T any] = buffer.UnboundedChan[T]

// 包内构造函数：保留历史命名以减少调用点改动。
func newRingBuffer[T any](initialSize int) *ringBuffer[T] {
	return buffer.NewRing[T](initialSize)
}

func newUnboundedChan[T any](ctx context.Context, initCapacity int) *unboundedChan[T] {
	return buffer.NewUnbounded[T](ctx, initCapacity)
}

//nolint:unused // 保留导出形式，便于将来拆分非默认尺寸场景
func newUnboundedChanSize[T any](ctx context.Context, in, out, buf int) *unboundedChan[T] {
	return buffer.NewUnboundedSize[T](ctx, in, out, buf)
}
