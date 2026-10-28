package redisson

import (
	"context"
	"strconv"
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
		onBuffer: func(int) { buffered++ },
		onWrite:  func(int, *ringBuffer[int]) { writes++ },
		onClose:  func(dropped int) { closed <- dropped },
	})
	if cap(queue.In) != 2 || cap(queue.Out) != 3 {
		t.Fatal("mapped queue changed configured capacities")
	}
	// 无消费者时也可写入远超过频道容量的数据，覆盖环形缓冲扩容路径。
	for value := range count {
		select {
		case queue.In <- value:
		case <-ctx.Done():
			t.Fatal("queue stopped accepting data without a consumer")
		}
	}
	close(queue.In)
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
