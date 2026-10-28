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
	q := newPubSubMonitorQueue(ctx, func() time.Duration { return time.Duration(now.Load()) }, func(wait time.Duration) { delivered <- wait }, func(dropped int) { closed <- dropped })
	first := Message{Channel: "foo", Message: "公会"}
	second := Message{Channel: "foo", Pattern: "f*", Message: "two"}
	if !q.push(first) {
		t.Fatal("first enqueue rejected")
	}
	now.Store(20)
	if !q.push(second) {
		t.Fatal("second enqueue rejected")
	}
	now.Store(30)
	depth, bytes, age, _ := q.snapshot()
	if depth != 2 || bytes != messageBytes(first)+messageBytes(second) || age != 20 {
		t.Fatalf("unexpected queue stats: %d %d %v", depth, bytes, age)
	}
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	for index, want := range []Message{first, second} {
		select {
		case got := <-q.out:
			if got != want {
				t.Fatalf("FIFO changed: got=%+v want=%+v", got, want)
			}
		case <-timer.C:
			t.Fatal("delivery timed out")
		}
		select {
		case wait := <-delivered:
			if wait != time.Duration(20-index*10) {
				t.Fatalf("unexpected wait: %v", wait)
			}
		case <-timer.C:
			t.Fatal("dequeue observation timed out")
		}
	}
	depth, bytes, age, _ = q.snapshot()
	if depth != 0 || bytes != 0 || age != 0 {
		t.Fatalf("drained queue retained observations: %d %d %v", depth, bytes, age)
	}
	if !q.push(first) || !q.push(second) {
		t.Fatal("enqueue before close rejected")
	}
	cancel()
	select {
	case dropped := <-closed:
		if dropped != 2 {
			t.Fatalf("unexpected close discard: %d", dropped)
		}
	case <-timer.C:
		t.Fatal("close timed out")
	}
	depth, bytes, age, isClosed := q.snapshot()
	if !isClosed || depth != 0 || bytes != 0 || age != 0 || q.push(first) {
		t.Fatal("closed queue accepted data or retained state")
	}
}

func TestPubSubMonitorQueueConcurrent(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var now atomic.Int64
	const writers, each = 8, 100
	delivered := make(chan struct{}, writers*each)
	q := newPubSubMonitorQueue(ctx, func() time.Duration { return time.Duration(now.Add(1)) }, func(time.Duration) { delivered <- struct{}{} }, nil)
	var wg sync.WaitGroup
	for i := range writers {
		wg.Add(1)
		go func(writer int) {
			defer wg.Done()
			for j := range each {
				if !q.push(Message{Channel: strconv.Itoa(writer), Message: strconv.Itoa(j)}) {
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
				for b.Loop() {
					queue.In <- message
					<-queue.Out
				}
				return
			}
			queue := newPubSubMonitorQueue(ctx, func() time.Duration { return 1 }, nil, nil)
			for b.Loop() {
				queue.push(message)
				<-queue.out
			}
		})
	}
}
