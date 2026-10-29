package redisson

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/rueidis"
)

func waitPubSubStats(t *testing.T, p PubSub, condition func(PubSubStats) bool) PubSubStats {
	t.Helper()
	observer := p.(PubSubObserver)
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
	tick := time.NewTicker(5 * time.Millisecond)
	defer tick.Stop()
	for {
		snapshot := observer.PubSubStats()
		if condition(snapshot) {
			return snapshot
		}
		select {
		case <-deadline.C:
			t.Fatalf("pubsub observation did not converge: %+v", snapshot)
		case <-tick.C:
		}
	}
}

func TestPubSubMonitorQueueStats(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var now atomic.Int64
	now.Store(10)
	delivered := make(chan time.Duration, 2)
	closed := make(chan int, 1)
	q := newPubSubMonitorQueue(ctx, 2, func() time.Duration { return time.Duration(now.Load()) }, func(wait time.Duration) { delivered <- wait }, func(dropped int) { closed <- dropped }, 1)
	sender := q.newSender(ctx)
	t.Cleanup(sender.release)
	if cap(q.messages.In) != 2 || cap(q.out) != 2 {
		t.Fatal("monitor changed the original input/output buffering")
	}
	first := Message{Channel: "foo", Message: "公会"}
	second := Message{Channel: "foo", Pattern: "f*", Message: "two"}
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	for index, message := range []Message{first, second} {
		now.Store(int64(10 + index*10))
		if !sender.push(message) {
			t.Fatal("enqueue rejected")
		}
		select {
		case wait := <-delivered:
			if wait != 0 {
				t.Fatalf("forwarding to an available Out buffer took %v", wait)
			}
		case <-timer.C:
			t.Fatal("forwarding to buffered Out timed out without a consumer")
		}
	}
	now.Store(30)
	depth, bytes, age, _ := q.snapshot()
	if depth != 2 || bytes != messageBytes(first)+messageBytes(second) || age != 20 {
		t.Fatalf("Out backlog was omitted: %d %d %v", depth, bytes, age)
	}
	for _, want := range []Message{first, second} {
		if got := <-q.out; got != want {
			t.Fatalf("FIFO changed: got=%+v want=%+v", got, want)
		}
	}
	depth, bytes, age, _ = q.snapshot()
	if depth != 0 || bytes != 0 || age != 0 {
		t.Fatalf("drained queue retained observations: %d %d %v", depth, bytes, age)
	}
	// Out 中两条仍可在 Close 后读取，只有 In/环形缓冲中的三条被丢弃。
	for _, message := range []Message{first, second} {
		if !sender.push(message) {
			t.Fatal("enqueue before close rejected")
		}
		select {
		case <-delivered:
		case <-timer.C:
			t.Fatal("output buffer did not fill")
		}
	}
	for range 3 {
		if !sender.push(first) {
			t.Fatal("enqueue beyond output capacity rejected")
		}
	}
	cancel()
	select {
	case dropped := <-closed:
		if dropped != 3 {
			t.Fatalf("unexpected close discard: %d", dropped)
		}
	case <-timer.C:
		t.Fatal("close timed out")
	}
	depth, bytes, age, isClosed := q.snapshot()
	if !isClosed || depth != 0 || bytes != 0 || age != 0 || sender.push(first) {
		t.Fatal("closed queue accepted data or retained state")
	}
	var retained []Message
	for message := range q.out {
		retained = append(retained, message)
	}
	if len(retained) != 2 || retained[0] != first || retained[1] != second {
		t.Fatalf("closing discarded buffered output: %+v", retained)
	}
}

func TestPubSubMonitorOutputBacklogAfterWrap(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var now atomic.Int64
	forwarded := make(chan struct{}, 1)
	q := newPubSubMonitorQueue(ctx, 2, func() time.Duration { return time.Duration(now.Load()) }, func(time.Duration) { forwarded <- struct{}{} }, nil, 1)
	sender := q.newSender(ctx)
	t.Cleanup(sender.release)
	type pending struct {
		message Message
		at      time.Duration
	}
	var expected []pending
	for index := range 64 {
		if len(expected) == 2 {
			if got := <-q.out; got != expected[0].message {
				t.Fatalf("FIFO changed after output metadata wrap: %+v", got)
			}
			expected = expected[1:]
		}
		now.Store(int64(index * 10))
		message := Message{Channel: "channel", Pattern: "ch*", Message: fmt.Sprintf("%0*d", index+1, index)}
		if !sender.push(message) {
			t.Fatal("enqueue rejected")
		}
		select {
		case <-forwarded:
		case <-ctx.Done():
			t.Fatal("forwarding timed out")
		}
		expected = append(expected, pending{message: message, at: time.Duration(index * 10)})
		now.Store(int64(index*10 + 7))
		var wantBytes int64
		for _, value := range expected {
			wantBytes += messageBytes(value.message)
		}
		depth, bytes, age, _ := q.snapshot()
		if depth != len(expected) || bytes != wantBytes || age != time.Duration(now.Load())-expected[0].at {
			t.Fatalf("output metadata retained a consumed message at %d: depth=%d bytes=%d age=%v", index, depth, bytes, age)
		}
	}
}

func TestPubSubMonitorDequeueCanReadStats(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	observed := make(chan int, 1)
	var q *pubSubMonitorQueue
	q = newPubSubMonitorQueue(ctx, 2, func() time.Duration { return 1 }, func(time.Duration) {
		depth, _, _, _ := q.snapshot()
		observed <- depth
	}, nil, 1)
	sender := q.newSender(ctx)
	t.Cleanup(sender.release)
	if !sender.push(Message{Message: "pending"}) {
		t.Fatal("enqueue rejected")
	}
	select {
	case depth := <-observed:
		if depth != 1 {
			t.Fatalf("dequeue callback omitted buffered output: %d", depth)
		}
	case <-ctx.Done():
		t.Fatal("reading stats from the dequeue callback blocked forwarding")
	}
}

func TestPubSubMonitorWaitSamplingKeepsFullBacklog(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var samples atomic.Int64
	q := newPubSubMonitorQueue(ctx, 8, func() time.Duration { return 10 }, func(time.Duration) { samples.Add(1) }, nil, 4)
	sender := q.newSender(ctx)
	t.Cleanup(sender.release)
	message := Message{Channel: "sample", Message: "payload"}
	for range 8 {
		if !sender.push(message) {
			t.Fatal("enqueue rejected")
		}
	}
	waitShardTest(t, func() bool { return q.sent.Load() == 8 && samples.Load() == 2 })
	depth, bytes, age, _ := q.snapshot()
	if depth != 8 || bytes != 8*messageBytes(message) || age != 0 {
		t.Fatalf("wait sampling changed full backlog statistics: depth=%d bytes=%d age=%v", depth, bytes, age)
	}
	defaults := (PubSubMonitorOptions{}).withDefaults()
	if defaults.DequeueSampleEvery != 64 || (PubSubMonitorOptions{DequeueSampleEvery: 1}).withDefaults().DequeueSampleEvery != 1 {
		t.Fatal("default or explicit wait sampling interval was lost")
	}
}

func TestPubSubMonitorQueueCancelsBlockedSubscription(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	release := make(chan struct{})
	t.Cleanup(func() { close(release); cancel() })
	forwarding := make(chan struct{}, 1)
	q := newPubSubMonitorQueue(ctx, 1, func() time.Duration { return 1 }, func(time.Duration) {
		select {
		case forwarding <- struct{}{}:
		default:
		}
		<-release
	}, nil, 1)
	sender := q.newSender(ctx)
	t.Cleanup(sender.release)
	message := Message{Channel: "blocked", Message: "payload"}
	if !sender.push(message) {
		t.Fatal("initial enqueue rejected")
	}
	select {
	case <-forwarding:
	case <-ctx.Done():
		t.Fatal("forwarding did not reach the pause")
	}
	if !sender.push(message) {
		t.Fatal("input buffer did not accept the pending message")
	}
	subscription, unsubscribe := context.WithCancel(ctx)
	defer unsubscribe()
	subscriptionSender := q.newSender(subscription)
	t.Cleanup(subscriptionSender.release)
	result := make(chan bool, 1)
	go func() { result <- subscriptionSender.push(message) }()
	waitShardTest(t, func() bool { return q.admittedBytes() == 3*messageBytes(message) })
	unsubscribe()
	select {
	case accepted := <-result:
		if accepted || ctx.Err() != nil {
			t.Fatal("subscription cancellation must reject the blocked enqueue while the queue stays alive")
		}
	case <-ctx.Done():
		t.Fatal("blocked enqueue waited for the whole PubSub to close")
	}
	depth, bytes, _, closed := q.snapshot()
	if closed || depth != 2 || bytes != 2*messageBytes(message) {
		t.Fatalf("canceled admission retained statistics: depth=%d bytes=%d closed=%v", depth, bytes, closed)
	}
}

func TestPubSubMonitorSenderReuseKeepsFullBytes(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	q := newPubSubMonitorQueue(ctx, 64, func() time.Duration { return 10 }, nil, nil, 1)
	var expected []Message
	var wantBytes int64
	// 接收任务反复创建和取消后，累计入队字节不能丢失或重复计算。
	for index := range 40 {
		subscription, unsubscribe := context.WithCancel(ctx)
		sender := q.newSender(subscription)
		message := Message{Channel: fmt.Sprintf("channel-%d", index), Pattern: "ch*", Message: fmt.Sprintf("%0*d", index+1, index)}
		if !sender.push(message) {
			t.Fatal("new receiver rejected an enqueue")
		}
		unsubscribe()
		sender.release()
		sender.release()
		if sender.push(message) {
			t.Fatal("canceled receiver accepted another message")
		}
		expected = append(expected, message)
		wantBytes += messageBytes(message)
	}
	waitShardTest(t, func() bool { return q.sent.Load() == uint64(len(expected)) })
	q.sendersMu.Lock()
	active := len(q.senders)
	q.sendersMu.Unlock()
	if active != 0 {
		t.Fatalf("finished receivers retained observation objects: %d", active)
	}
	depth, bytes, age, _ := q.snapshot()
	if depth != len(expected) || bytes != wantBytes || age != 0 {
		t.Fatalf("receiver reuse changed full statistics: depth=%d bytes=%d age=%v", depth, bytes, age)
	}
	for _, want := range expected {
		if got := <-q.out; got != want {
			t.Fatalf("receiver reuse changed FIFO: got=%+v want=%+v", got, want)
		}
	}
	depth, bytes, age, _ = q.snapshot()
	if depth != 0 || bytes != 0 || age != 0 {
		t.Fatalf("receiver reuse left pending observations: %d %d %v", depth, bytes, age)
	}
}

type pubSubErrCheckpointContext struct {
	context.Context
	checked chan struct{}
	resume  chan struct{}
}

func (c pubSubErrCheckpointContext) Err() error {
	err := c.Context.Err()
	close(c.checked)
	<-c.resume
	return err
}

func TestPubSubMonitorDefaultFastPathCancellation(t *testing.T) {
	input := make(chan Message, 1)
	first := Message{Message: "first"}
	input <- first
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()
	checked, resume := make(chan struct{}), make(chan struct{})
	fixture := &unboundedChan[Message, Message]{In: input, in: input, stopping: make(chan struct{}), producers: make(map[*unboundedChanSender[Message, Message]]struct{})}
	fixture.producersChanged = sync.NewCond(&fixture.producersMu)
	sender := fixture.newSender(pubSubErrCheckpointContext{Context: parent, checked: checked, resume: resume})
	defer sender.release()
	result := make(chan bool, 1)
	go func() { result <- sender.send(Message{Message: "canceled"}) }()
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	select {
	case <-checked:
	case <-timer.C:
		t.Fatal("sender did not check cancellation")
	}
	cancel()
	close(resume)
	select {
	case accepted := <-result:
		if accepted {
			t.Fatal("full input accepted a canceled send")
		}
	case <-timer.C:
		t.Fatal("default receiver did not stop waiting on cancellation")
	}
	if len(input) != 1 || <-input != first {
		t.Fatal("canceled wait changed buffered input")
	}
}

func TestPubSubMonitorPrefixBeforePublication(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	q := newPubSubMonitorQueue(ctx, 2, func() time.Duration { return 200 }, nil, nil, 1)
	sender := q.newSender(ctx)
	defer sender.release()
	// 只填观测记录，暂停在“消息已写入 Out，sent 尚未发布”的交接点。
	out := make(chan Message, 2)
	q.out = out
	var messages []Message
	var cumulative []int64
	var total int64
	for index := 1; index <= 11; index++ {
		message := Message{Channel: "c", Pattern: "p", Message: fmt.Sprintf("%0*d", index, index)}
		messages = append(messages, message)
		total += messageBytes(message)
		cumulative = append(cumulative, total)
		if index <= 10 {
			record := &q.output[uint64(index-1)%uint64(len(q.output))]
			record.enqueued.Store(int64(index * 10))
			record.totalBytes.Store(total)
		}
	}
	sender.total = cumulative[9]
	sender.admitted.Store(sender.total)
	q.sent.Store(10)
	out <- messages[8]
	out <- messages[9]
	depth, bytes, age, _ := q.snapshot()
	if depth != 2 || bytes != messageBytes(messages[8])+messageBytes(messages[9]) || age != 110 {
		t.Fatalf("wrapped prefix changed stable output: depth=%d bytes=%d age=%v", depth, bytes, age)
	}
	<-out
	sender.total = cumulative[10]
	sender.admitted.Store(sender.total)
	out <- messages[10]
	record := &q.output[10%uint64(len(q.output))]
	record.enqueued.Store(110)
	record.totalBytes.Store(cumulative[10])
	// 新写入的记录不能覆盖旧 sent 对应的已消费前缀，避免误报零积压。
	depth, bytes, age, _ = q.snapshot()
	wantBytes := messageBytes(messages[9]) + messageBytes(messages[10])
	if depth != 2 || bytes < wantBytes || age < 100 {
		t.Fatalf("unpublished write hid pending output: depth=%d bytes=%d age=%v", depth, bytes, age)
	}
	q.sent.Store(11)
	depth, bytes, age, _ = q.snapshot()
	if depth != 2 || bytes != wantBytes || age != 100 {
		t.Fatalf("published prefix did not converge: depth=%d bytes=%d age=%v", depth, bytes, age)
	}
}

func TestPubSubMonitorQueueConcurrent(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var now atomic.Int64
	const writers, each = 8, 100
	delivered := make(chan struct{}, writers*each)
	q := newPubSubMonitorQueue(ctx, 32, func() time.Duration { return time.Duration(now.Add(1)) }, func(time.Duration) { delivered <- struct{}{} }, nil, 1)
	var wg sync.WaitGroup
	for i := range writers {
		wg.Add(1)
		go func(writer int) {
			defer wg.Done()
			sender := q.newSender(ctx)
			defer sender.release()
			for j := range each {
				if !sender.push(Message{Channel: strconv.Itoa(writer), Message: strconv.Itoa(j)}) {
					return
				}
				q.snapshot()
			}
		}(i)
	}
	seen := make(map[string]int)
	for range writers * each {
		select {
		case msg := <-q.out:
			sequence, err := strconv.Atoi(msg.Message)
			if err != nil || sequence != seen[msg.Channel] {
				t.Fatalf("producer FIFO changed: %+v", msg)
			}
			seen[msg.Channel]++
		case <-ctx.Done():
			t.Fatal("concurrent delivery timed out")
		}
	}
	wg.Wait()
	for range writers * each {
		select {
		case <-delivered:
		case <-ctx.Done():
			t.Fatal("dequeue completion timed out")
		}
	}
	depth, bytes, _, _ := q.snapshot()
	if depth != 0 || bytes != 0 {
		t.Fatalf("concurrent queue did not drain: %d %d", depth, bytes)
	}
}

func isolatedMonitorClient(t *testing.T, resp2 bool) Cmdable {
	t.Helper()
	conf := NewConf(WithDevelopment(false), WithEnableCache(false), WithAlwaysRESP2(resp2))
	if conf.Net != "unix" {
		t.Fatal("integration test must use its owned Redis socket")
	}
	c := MustNewClient(conf)
	t.Cleanup(func() { _ = c.Close() })
	return c
}

func TestPubSubMonitoringDelivery(t *testing.T) {
	for _, resp2 := range []bool{false, true} {
		for _, pattern := range []bool{false, true} {
			t.Run(fmt.Sprintf("RESP2=%v/pattern=%v", resp2, pattern), func(t *testing.T) {
				c := isolatedMonitorClient(t, resp2)
				var ticks atomic.Int64
				dequeued := make(chan struct{}, 1)
				var ackEvents, returnEvents atomic.Int64
				closed := make(chan struct{}, 1)
				ctx := WithPubSubMonitoring(context.Background(), PubSubMonitorOptions{Now: func() time.Duration { return time.Duration(ticks.Add(1)) },
					OnEvent: func(event PubSubEvent) {
						switch event.Kind {
						case PubSubSubscribeAck:
							ackEvents.Add(1)
						case PubSubReceiveReturn:
							returnEvents.Add(1)
						case PubSubClosed:
							select {
							case closed <- struct{}{}:
							default:
							}
						}
					}, OnDequeue: func(time.Duration) { dequeued <- struct{}{} }})
				p := c.Subscribe(ctx)
				t.Cleanup(func() { _ = p.Close() })
				var err error
				if pattern {
					err = p.PSubscribe(ctx, "observed.*")
				} else {
					err = p.Subscribe(ctx, "observed.one")
				}
				if err != nil {
					t.Fatal(err)
				}
				snapshot := waitPubSubStats(t, p, func(s PubSubStats) bool { return s.Acknowledged == 1 })
				if !snapshot.Enabled || snapshot.Receivers != 1 || snapshot.Subscriptions != 1 || snapshot.SubscribeAcks < 1 {
					t.Fatalf("invalid subscription observation: %+v", snapshot)
				}
				if err := c.Publish(ctx, "observed.one", "公会").Err(); err != nil {
					t.Fatal(err)
				}
				snapshot = waitPubSubStats(t, p, func(s PubSubStats) bool { return s.QueueDepth == 1 })
				want := int64(len("observed.one") + len("公会"))
				if pattern {
					want += int64(len("observed.*"))
				}
				if snapshot.QueueBytes != want {
					t.Fatalf("unexpected queue bytes: %d want=%d", snapshot.QueueBytes, want)
				}
				timer := time.NewTimer(5 * time.Second)
				defer timer.Stop()
				select {
				case msg := <-p.Channel():
					if msg.Message != "公会" {
						t.Fatalf("message changed: %+v", msg)
					}
				case <-timer.C:
					t.Fatal("message timed out")
				}
				select {
				case <-dequeued:
				case <-timer.C:
					t.Fatal("dequeue callback timed out")
				}
				if err := p.Close(); err != nil {
					t.Fatal(err)
				}
				snapshot = waitPubSubStats(t, p, func(s PubSubStats) bool { return s.Closed && s.Receivers == 0 })
				select {
				case <-closed:
				case <-timer.C:
					t.Fatal("close event timed out")
				}
				if ackEvents.Load() == 0 || returnEvents.Load() == 0 {
					t.Fatal("configured event callback was not invoked")
				}

				if snapshot.Acknowledged != 0 || snapshot.ReceiveErrors != 0 || snapshot.QueueDepth != 0 {
					t.Fatalf("close retained subscription or queue state: %+v", snapshot)
				}
			})
		}
	}
}

func TestPubSubMonitoringReceiveError(t *testing.T) {
	c := isolatedMonitorClient(t, false)
	ctx := WithPubSubMonitoring(context.Background(), PubSubMonitorOptions{Now: func() time.Duration { return 1 }})
	p := c.Subscribe(ctx)
	t.Cleanup(func() { _ = p.Close() })
	// 无频道的 SUBSCRIBE 会被 Redis 拒绝，异步错误必须出现在观测中。
	if err := p.Subscribe(ctx); err != nil {
		t.Fatal(err)
	}
	snapshot := waitPubSubStats(t, p, func(s PubSubStats) bool { return s.ReceiveErrors == 1 && s.Receivers == 0 })
	if snapshot.LastError == nil || snapshot.Acknowledged != 0 || snapshot.ReceiveReturns == 0 {
		t.Fatalf("receive failure was hidden: %+v", snapshot)
	}
}

// 仅断开 TestMain 创建的私有 Redis 订阅连接，验证 rueidis 内部重连也会发出事件。
func TestPubSubMonitoringDisconnect(t *testing.T) {
	for _, resp2 := range []bool{false, true} {
		t.Run(fmt.Sprintf("RESP2=%v", resp2), func(t *testing.T) {
			c := isolatedMonitorClient(t, resp2)
			controller := isolatedMonitorClient(t, false)
			var interrupted atomic.Int64
			gap := make(chan struct{}, 1)
			var observer PubSubObserver
			ctx := WithPubSubMonitoring(context.Background(), PubSubMonitorOptions{
				Now: func() time.Duration { return 1 },
				OnEvent: func(event PubSubEvent) {
					if event.Kind == PubSubReceiveReturn && event.Err != nil {
						interrupted.Add(1)
						if observer.PubSubStats().Acknowledged == 0 {
							select {
							case gap <- struct{}{}:
							default:
							}
						}
					}
				},
			})
			p := c.Subscribe(ctx)
			observer = p.(PubSubObserver)
			if err := p.Subscribe(ctx, "reconnect.observed"); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = p.Close() })
			waitPubSubStats(t, p, func(s PubSubStats) bool { return s.Acknowledged == 1 })
			killed := controller.ClientKillByFilter(ctx, "TYPE", "pubsub", "SKIPME", "yes")
			if killed.Err() != nil || killed.Val() < 1 {
				t.Fatalf("failed to disconnect isolated subscription: %v %d", killed.Err(), killed.Val())
			}
			timer := time.NewTimer(5 * time.Second)
			defer timer.Stop()
			select {
			case <-gap:
			case <-timer.C:
				t.Fatal("receive gap did not clear subscription acknowledgement")
			}
			if resp2 {
				// RESP2 的订阅管道断线恢复另行验收；此处只验证故障可见且 ACK 已失效。
				return
			}
			snapshot := waitPubSubStats(t, p, func(s PubSubStats) bool {
				return s.SubscribeAcks >= 2 && s.Acknowledged == 1 && interrupted.Load() > 0
			})
			if snapshot.ReceiveErrors != 0 || snapshot.LastError == nil || snapshot.Receivers != 1 {
				t.Fatalf("transparent retry observation was lost: %+v", snapshot)
			}
			if err := c.Publish(ctx, "reconnect.observed", "after reconnect").Err(); err != nil {
				t.Fatal(err)
			}
			select {
			case message := <-p.Channel():
				if message.Message != "after reconnect" {
					t.Fatalf("message changed: %+v", message)
				}
			case <-timer.C:
				t.Fatal("delivery after reconnect timed out")
			}
		})
	}
}

func TestPubSubMonitoringSlotMigration(t *testing.T) {
	for _, resp2 := range []bool{false, true} {
		t.Run(fmt.Sprintf("RESP2=%v", resp2), func(t *testing.T) {
			cluster := newShardTestCluster(t)
			c := shardTestClient(t, cluster, resp2)
			var ticks atomic.Int64
			ctx := WithPubSubMonitoring(context.Background(), PubSubMonitorOptions{Now: func() time.Duration { return time.Duration(ticks.Add(1)) }})
			p := c.SSubscribe(ctx, "foo")
			t.Cleanup(func() { _ = p.Close() })
			waitPubSubStats(t, p, func(s PubSubStats) bool { return s.Acknowledged == 1 })
			assertShardMessage(t, c, p, "before migration")
			cluster.move()
			snapshot := waitPubSubStats(t, p, func(s PubSubStats) bool { return s.Retries >= 1 && s.Acknowledged == 1 && s.SubscribeAcks >= 2 })
			if snapshot.UnsubscribeAcks == 0 || snapshot.ReceiveReturns == 0 || snapshot.Receivers != 1 {
				t.Fatalf("migration events were hidden: %+v", snapshot)
			}
			assertShardMessage(t, c, p, "after migration")
			if err := p.SUnsubscribe(ctx, "foo"); err != nil {
				t.Fatal(err)
			}
			snapshot = waitPubSubStats(t, p, func(s PubSubStats) bool { return s.Receivers == 0 })
			if snapshot.Acknowledged != 0 || snapshot.ReceiveErrors != 0 {
				t.Fatalf("unsubscribe retained acknowledgement or reported false error: %+v", snapshot)
			}
		})
	}
}

func TestPubSubMonitoringDefaultPreservesHook(t *testing.T) {
	c := isolatedMonitorClient(t, false)
	var confirmations atomic.Int64
	ctx := rueidis.WithOnSubscriptionHook(context.Background(), func(rueidis.PubSubSubscription) { confirmations.Add(1) })
	p := c.Subscribe(ctx, "legacy-hook")
	t.Cleanup(func() { _ = p.Close() })
	waitShardTest(t, func() bool { return confirmations.Load() > 0 })
	internal := p.(*pubSub)
	if internal.monitor == nil || internal.monitor.messages == nil || internal.monitor.queue != nil {
		t.Fatal("default subscription must use its monitor facade and original buffered queue")
	}
	if cap(p.Channel()) != c.Options().GetPubSubChanSize() || p.(PubSubObserver).PubSubStats().Enabled {
		t.Fatal("default subscription changed buffering or enabled collection")
	}
	if err := c.Publish(ctx, "legacy-hook", "default monitor").Err(); err != nil {
		t.Fatal(err)
	}
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	select {
	case message := <-p.Channel():
		if message.Message != "default monitor" {
			t.Fatalf("message changed: %+v", message)
		}
	case <-timer.C:
		t.Fatal("default monitor delivery timed out")
	}
}

func TestPubSubMonitoringPartialOptions(t *testing.T) {
	for _, callbacks := range []bool{false, true} {
		t.Run(fmt.Sprintf("callbacks=%v", callbacks), func(t *testing.T) {
			client := isolatedMonitorClient(t, false)
			var events atomic.Int64
			options := PubSubMonitorOptions{}
			if callbacks {
				options.OnEvent = func(PubSubEvent) { events.Add(1) }
			}
			ctx := WithPubSubMonitoring(context.Background(), options)
			p := client.Subscribe(ctx, "partial.options")
			t.Cleanup(func() { _ = p.Close() })
			waitPubSubStats(t, p, func(s PubSubStats) bool {
				return s.Acknowledged == 1 && (!callbacks || events.Load() > 0)
			})
			if err := client.Publish(ctx, "partial.options", "message").Err(); err != nil {
				t.Fatal(err)
			}
			snapshot := waitPubSubStats(t, p, func(s PubSubStats) bool { return s.QueueDepth == 1 })
			if snapshot.QueueBytes != int64(len("partial.options")+len("message")) || snapshot.QueueOldestAge < 0 {
				t.Fatalf("partial options lost queue observation: %+v", snapshot)
			}
			// 未设置 OnDequeue 时也能完成交付，空回调不影响消息。
			timer := time.NewTimer(5 * time.Second)
			defer timer.Stop()
			select {
			case message := <-p.Channel():
				if message.Message != "message" {
					t.Fatalf("message changed: %+v", message)
				}
			case <-timer.C:
				t.Fatal("partial options delivery timed out")
			}
		})
	}
}

func TestPubSubMonitoringComposesExistingHooks(t *testing.T) {
	c := isolatedMonitorClient(t, false)
	var ticks, acks, returns atomic.Int64
	parent := rueidis.WithOnSubscriptionHook(context.Background(), func(rueidis.PubSubSubscription) { acks.Add(1) })
	parent = rueidis.WithOnReceiveReturnHook(parent, func(err error, _ rueidis.CommandClient) error { returns.Add(1); return err })
	ctx := WithPubSubMonitoring(parent, PubSubMonitorOptions{Now: func() time.Duration { return time.Duration(ticks.Add(1)) }})
	p := c.Subscribe(ctx, "composed-hook")
	t.Cleanup(func() { _ = p.Close() })
	waitPubSubStats(t, p, func(s PubSubStats) bool { return s.Acknowledged == 1 && acks.Load() > 0 })
	if err := p.Close(); err != nil {
		t.Fatal(err)
	}
	waitPubSubStats(t, p, func(s PubSubStats) bool { return s.Receivers == 0 && returns.Load() > 0 })
}

func TestPubSubMonitoringPreservesReturnHookResult(t *testing.T) {
	c := isolatedMonitorClient(t, false)
	var calls atomic.Int64
	parent := rueidis.WithOnReceiveReturnHook(context.Background(), func(err error, _ rueidis.CommandClient) error {
		if err != nil {
			calls.Add(1)
		}
		return nil
	})
	ctx := WithPubSubMonitoring(parent, PubSubMonitorOptions{Now: func() time.Duration { return 1 }})
	p := c.Subscribe(ctx)
	t.Cleanup(func() { _ = p.Close() })
	if err := p.Subscribe(ctx); err != nil {
		t.Fatal(err)
	}
	snapshot := waitPubSubStats(t, p, func(s PubSubStats) bool { return s.Receivers == 0 && s.ReceiveReturns == 1 && calls.Load() == 1 })
	if snapshot.ReceiveErrors != 0 || snapshot.LastError == nil {
		t.Fatalf("business return hook result or raw observation changed: %+v", snapshot)
	}
}

// 只比较队列传递开销，注入固定时钟，不包含 Redis、序列化与真实时钟读取。
func BenchmarkPubSubQueueMonitoring(b *testing.B) {
	for _, observed := range []bool{false, true} {
		b.Run(fmt.Sprintf("observed=%v", observed), func(b *testing.B) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			message := Message{Channel: "bench", Message: "payload"}
			b.ReportAllocs()
			if !observed {
				queue := newUnboundedChan[Message](ctx, 1024)
				sender := queue.newSender(ctx)
				defer sender.release()
				for b.Loop() {
					sender.send(message)
					<-queue.Out
				}
				return
			}
			queue := newPubSubMonitorQueue(ctx, 1024, func() time.Duration { return 1 }, nil, nil, 1)
			sender := queue.newSender(ctx)
			defer sender.release()
			for b.Loop() {
				sender.push(message)
				<-queue.out
			}
		})
	}
}
