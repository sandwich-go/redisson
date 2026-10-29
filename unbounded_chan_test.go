package redisson

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestUnboundedChanMappedFIFOAndDrain(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	const count = 4096
	closed := make(chan int, 1)
	var buffered, writes int
	queue := newUnboundedChanMapped(ctx, 2, 3, 2, strconv.Itoa, &unboundedChanHooks[int]{
		onBuffer: func(int, *ringBuffer[int]) { buffered++ },
		onWrite:  func(int, *ringBuffer[int], bool) { writes++ },
		onClose:  func(dropped int) { closed <- dropped },
	})
	if cap(queue.In) != 2 || cap(queue.Out) != 3 {
		t.Fatal("mapped queue changed configured capacities")
	}
	// 无消费者时也可写入远超过频道容量的数据，覆盖环形缓冲扩容路径。
	sender := queue.newSender(ctx)
	defer sender.release()
	for value := range count {
		if !sender.send(value) {
			t.Fatal("queue stopped accepting data without a consumer")
		}
	}
	queue.CloseInput()
	for value := range count {
		select {
		case got, ok := <-queue.Out:
			if !ok || got != strconv.Itoa(value) {
				t.Fatalf("FIFO changed at %d: %q, open=%v", value, got, ok)
			}
		case <-ctx.Done():
			t.Fatal("input close did not drain pending data")
		}
	}
	select {
	case dropped := <-closed:
		if dropped != 0 || buffered == 0 || writes != count {
			t.Fatalf("incomplete forwarding: dropped=%d buffered=%d writes=%d", dropped, buffered, writes)
		}
	case <-ctx.Done():
		t.Fatal("queue did not close after draining")
	}
}

func TestUnboundedCloseRacesProducers(t *testing.T) {
	for range 20 {
		ctx, cancel := context.WithCancel(context.Background())
		queue := newUnboundedChan[[]byte](ctx, 1)
		var writers sync.WaitGroup
		for range 8 {
			writers.Add(1)
			go func() {
				defer writers.Done()
				sender := queue.newSender(ctx)
				defer sender.release()
				for range 100 {
					if !sender.send(make([]byte, 128)) {
						return
					}
				}
			}()
		}
		waitShardTest(t, func() bool { return queue.BufLen() > 0 })
		cancel()
		done := make(chan struct{})
		go func() { writers.Wait(); close(done) }()
		timer := time.NewTimer(5 * time.Second)
		select {
		case <-done:
		case <-timer.C:
			t.Fatal("queue close left producers blocked")
		}
		for range queue.Out {
		}
		timer.Stop()
		if len(queue.In) != 0 || queue.BufLen() != 0 || len(queue.buffer.buf) != 0 || queue.buffer.Capacity() != 0 {
			t.Fatal("queue close retained discarded values or expanded storage")
		}
	}
}

func TestRingBufferReadReleasesConsumedReference(t *testing.T) {
	buffer := newRingBuffer[[]byte](2)
	buffer.Write(make([]byte, 4096))
	value := buffer.Pop()
	if len(value) != 4096 || !buffer.IsEmpty() {
		t.Fatal("read changed the returned value")
	}
	for _, retained := range buffer.buf {
		if retained != nil {
			t.Fatal("consumed slot retained its payload")
		}
	}
}
