package buffer

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRingFIFO(t *testing.T) {
	r := NewRing[int](2)
	for i := 0; i < 50; i++ {
		r.Write(i)
	}
	for i := 0; i < 50; i++ {
		v, err := r.Read()
		if err != nil || v != i {
			t.Fatalf("Read i=%d got (%d, %v)", i, v, err)
		}
	}
}

func TestRingEmpty(t *testing.T) {
	r := NewRing[int](2)
	if _, err := r.Read(); !errors.Is(err, ErrIsEmpty) {
		t.Fatalf("Read empty err = %v, want ErrIsEmpty", err)
	}
}

func TestRingPanicOnZero(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("NewRing(0) should panic")
		}
	}()
	_ = NewRing[int](0)
}

func TestRingNormalizeOne(t *testing.T) {
	r := NewRing[int](1)
	if r.Capacity() != 2 {
		t.Fatalf("Capacity = %d, want 2", r.Capacity())
	}
}

func TestUnboundedChanFIFO(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := NewUnbounded[int](ctx, 4)
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
			t.Fatalf("got[%d] = %d, want %d", i, v, i)
		}
	}
}

func TestUnboundedChanCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ch := NewUnbounded[int](ctx, 2)
	for i := 0; i < 3; i++ {
		ch.In <- i
	}
	cancel()
	deadline := time.After(2 * time.Second)
	for {
		select {
		case _, ok := <-ch.Out:
			if !ok {
				return
			}
		case <-deadline:
			t.Fatal("Out did not close within 2s")
		}
	}
}
