package redisson

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRingBuffer_BasicReadWrite(t *testing.T) {
	rb := newRingBuffer[int](4)
	if !rb.IsEmpty() {
		t.Fatal("new buffer should be empty")
	}
	if rb.Len() != 0 {
		t.Fatalf("Len = %d, want 0", rb.Len())
	}
	if rb.Capacity() != 4 {
		t.Fatalf("Capacity = %d, want 4", rb.Capacity())
	}
	rb.Write(10)
	rb.Write(20)
	if rb.Len() != 2 {
		t.Fatalf("Len = %d, want 2", rb.Len())
	}
	v, err := rb.Read()
	if err != nil || v != 10 {
		t.Fatalf("Read = (%d, %v), want (10, nil)", v, err)
	}
	v, err = rb.Read()
	if err != nil || v != 20 {
		t.Fatalf("Read = (%d, %v), want (20, nil)", v, err)
	}
	_, err = rb.Read()
	if !errors.Is(err, ErrIsEmpty) {
		t.Fatalf("Read empty err = %v, want ErrIsEmpty", err)
	}
}

func TestRingBuffer_Grow(t *testing.T) {
	rb := newRingBuffer[int](2)
	// 写入足够多元素强制扩容
	for i := 0; i < 100; i++ {
		rb.Write(i)
	}
	if rb.Len() != 100 {
		t.Fatalf("Len = %d, want 100", rb.Len())
	}
	if rb.Capacity() <= 2 {
		t.Fatalf("Capacity should grow beyond 2, got %d", rb.Capacity())
	}
	// FIFO 顺序保留
	for i := 0; i < 100; i++ {
		v, err := rb.Read()
		if err != nil || v != i {
			t.Fatalf("Read i=%d got (%d, %v)", i, v, err)
		}
	}
}

func TestRingBuffer_PopPeekPanicOnEmpty(t *testing.T) {
	rb := newRingBuffer[int](2)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Pop on empty should panic")
		}
	}()
	rb.Pop()
}

func TestRingBuffer_Peek(t *testing.T) {
	rb := newRingBuffer[int](2)
	rb.Write(7)
	if v := rb.Peek(); v != 7 {
		t.Fatalf("Peek = %d, want 7", v)
	}
	// Peek 不消费
	if rb.Len() != 1 {
		t.Fatalf("Len after Peek = %d, want 1", rb.Len())
	}
}

func TestRingBuffer_NewWithSizeOne(t *testing.T) {
	// initialSize == 1 应被规整为 2（避免立即满）
	rb := newRingBuffer[int](1)
	if rb.Capacity() != 2 {
		t.Fatalf("Capacity = %d, want 2 (size=1 should normalize to 2)", rb.Capacity())
	}
}

func TestRingBuffer_NewWithZeroPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("newRingBuffer(0) should panic")
		}
	}()
	_ = newRingBuffer[int](0)
}

func TestRingBuffer_Reset(t *testing.T) {
	rb := newRingBuffer[int](2)
	for i := 0; i < 50; i++ {
		rb.Write(i)
	}
	rb.Reset()
	if !rb.IsEmpty() {
		t.Fatal("Reset should empty the buffer")
	}
	if rb.Capacity() != 2 {
		t.Fatalf("Reset capacity = %d, want 2", rb.Capacity())
	}
}

func TestUnboundedChan_BasicProduceConsume(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := newUnboundedChan[int](ctx, 4)
	const N = 1000
	go func() {
		for i := 0; i < N; i++ {
			ch.In <- i
		}
		close(ch.In)
	}()
	got := make([]int, 0, N)
	for v := range ch.Out {
		got = append(got, v)
	}
	if len(got) != N {
		t.Fatalf("got %d items, want %d", len(got), N)
	}
	for i, v := range got {
		if v != i {
			t.Fatalf("got[%d] = %d, want %d (FIFO order broken)", i, v, i)
		}
	}
}

func TestUnboundedChan_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ch := newUnboundedChan[int](ctx, 2)
	// 写入一些再 cancel
	for i := 0; i < 5; i++ {
		ch.In <- i
	}
	cancel()
	// 读到 ctx 取消应能在合理时间内退出
	deadline := time.After(2 * time.Second)
	for {
		select {
		case _, ok := <-ch.Out:
			if !ok {
				return // out 关闭，预期路径
			}
		case <-deadline:
			t.Fatal("unboundedChan did not close Out within 2s after ctx cancel")
		}
	}
}
