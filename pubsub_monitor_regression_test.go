package redisson

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestQAMonitorInputAge(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var now atomic.Int64
	now.Store(10)
	paused, resume := make(chan struct{}), make(chan struct{})
	var once sync.Once
	q := newPubSubMonitorQueue(ctx, 1, func() time.Duration { return time.Duration(now.Load()) }, func(time.Duration) {
		once.Do(func() { close(paused); <-resume })
	}, nil, 1)
	sender := q.newSender(ctx)
	t.Cleanup(func() { close(resume); cancel(); sender.release() })
	if !sender.push(Message{Message: "old"}) {
		t.Fatal("initial enqueue rejected")
	}
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	select {
	case <-paused:
	case <-timer.C:
		t.Fatal("forwarder did not pause")
	}
	<-q.out
	now.Store(20)
	if !sender.push(Message{Channel: "c", Message: "123456"}) {
		t.Fatal("input enqueue rejected")
	}
	sender.release()
	now.Store(100)
	depth, bytes, age, closed := q.snapshot()
	if closed || depth != 1 || bytes != 7 || age != 80 {
		t.Fatalf("input stage lost queue age: depth=%d bytes=%d age=%v closed=%v", depth, bytes, age, closed)
	}
}

func TestQAMonitorCloseReleasesDiscarded(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	closed := make(chan int, 1)
	q := newPubSubMonitorQueue(ctx, 1, func() time.Duration { return 1 }, nil, func(dropped int) { closed <- dropped }, 1)
	sender := q.newSender(ctx)
	defer sender.release()
	for index := range 4 {
		if !sender.push(Message{Channel: "c", Message: strings.Repeat(string(rune('a'+index)), 4096)}) {
			t.Fatal("enqueue rejected")
		}
	}
	waitShardTest(t, func() bool { return q.messages.BufLen() == 3 })
	cancel()
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	select {
	case dropped := <-closed:
		if dropped != 3 {
			t.Fatalf("discard count changed: %d", dropped)
		}
	case <-timer.C:
		t.Fatal("queue did not close")
	}
	if len(q.messages.In) != 0 || q.messages.BufLen() != 0 || len(q.messages.buffer.buf) != 0 || q.messages.buffer.Capacity() != 0 {
		t.Fatalf("discarded storage retained: in=%d pending=%d slots=%d capacity=%d", len(q.messages.In), q.messages.BufLen(), len(q.messages.buffer.buf), q.messages.buffer.Capacity())
	}
	message, ok := <-q.out
	if !ok || message.Channel != "c" || len(message.Message) != 4096 {
		t.Fatal("closing discarded unread Out")
	}
	if _, ok := <-q.out; ok {
		t.Fatal("closed output retained unexpected messages")
	}
}

func TestQAMonitorTransferAge(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	q := newPubSubMonitorQueue(ctx, 1, func() time.Duration { return 100 }, nil, nil, 1)
	sender := q.newSender(ctx)
	defer sender.release()
	message := Message{Channel: "c", Message: "123456"}
	sender.input[0].Store(20)
	sender.totalProduced = 1
	sender.produced.Store(1)
	sender.total = messageBytes(message)
	sender.admitted.Store(sender.total)
	value := pubSubQueuedMessage{message: message, enqueued: 20, sender: sender}
	q.messages.buffer.Write(value)
	// 内部缓冲已收到数据但尚未发布，输入观测仍必须覆盖交接中的消息。
	depth, bytes, age, closed := q.snapshot()
	if closed || depth != 1 || bytes != 7 || age != 80 {
		t.Fatalf("unpublished handoff lost age: depth=%d bytes=%d age=%v closed=%v", depth, bytes, age, closed)
	}
	q.onBuffer(value, q.messages.buffer)
	depth, bytes, age, closed = q.snapshot()
	if closed || depth != 1 || bytes != 7 || age != 80 {
		t.Fatalf("transfer stage lost age: depth=%d bytes=%d age=%v closed=%v", depth, bytes, age, closed)
	}
}

func TestQAMonitorCloseClearsInput(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	paused, resume := make(chan struct{}), make(chan struct{})
	closed := make(chan int, 1)
	var once sync.Once
	q := newPubSubMonitorQueue(ctx, 1, func() time.Duration { return 1 }, func(time.Duration) {
		once.Do(func() { close(paused); <-resume })
	}, func(n int) { closed <- n }, 1)
	sender := q.newSender(ctx)
	var resumeOnce sync.Once
	unblock := func() { resumeOnce.Do(func() { close(resume) }) }
	t.Cleanup(func() { unblock(); cancel(); sender.release() })
	if !sender.push(Message{Message: "keep"}) {
		t.Fatal("initial enqueue rejected")
	}
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	select {
	case <-paused:
	case <-timer.C:
		t.Fatal("forwarder did not pause")
	}
	if !sender.push(Message{Message: strings.Repeat("discard", 1024)}) {
		t.Fatal("input enqueue rejected")
	}
	cancel()
	unblock()
	select {
	case n := <-closed:
		if n != 1 {
			t.Fatalf("input discard count: %d", n)
		}
	case <-timer.C:
		t.Fatal("close did not complete")
	}
	if len(q.messages.In) != 0 || len(q.messages.buffer.buf) != 0 || q.messages.BufLen() != 0 {
		t.Fatal("closed queue retained discarded input")
	}
	if message, ok := <-q.out; !ok || message.Message != "keep" {
		t.Fatal("unread output was discarded")
	}
}

func TestQAMonitorCloseWaitsPreparingProducer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	entered, resume := make(chan struct{}), make(chan struct{})
	closed := make(chan int, 1)
	q := newPubSubMonitorQueue(ctx, 1, func() time.Duration { close(entered); <-resume; return 20 }, nil, func(n int) { closed <- n }, 1)
	sender := q.newSender(ctx)
	var once sync.Once
	unblock := func() { once.Do(func() { close(resume) }) }
	result := make(chan bool, 1)
	t.Cleanup(func() { unblock(); cancel() })
	go func() { result <- sender.push(Message{Message: strings.Repeat("payload", 1024)}); sender.release() }()
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	select {
	case <-entered:
	case <-timer.C:
		t.Fatal("producer did not reach preparation")
	}
	cancel()
	waitShardTest(t, func() bool { return q.messages.stopped.Load() })
	select {
	case <-closed:
		t.Fatal("close completed before an entered producer exited")
	default:
	}
	unblock()
	var accepted bool
	select {
	case accepted = <-result:
	case <-timer.C:
		t.Fatal("producer did not stop")
	}
	select {
	case n := <-closed:
		want := 0
		if accepted {
			want = 1
		}
		if n != want {
			t.Fatalf("late enqueue discard: got=%d want=%d", n, want)
		}
	case <-timer.C:
		t.Fatal("close did not wait for producer quiescence")
	}
	if len(q.messages.In) != 0 || q.messages.BufLen() != 0 || len(q.messages.buffer.buf) != 0 {
		t.Fatal("late producer left discarded storage")
	}
}
