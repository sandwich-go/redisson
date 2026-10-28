package redisson

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/redis/rueidis"
	"github.com/sandwich-go/boost/z"
)

// 同一份基准可在 781cd47 与当前实现上运行；两者注入相同的真实单调时钟。
func pubSubBenchmarkContext(ctx context.Context, custom bool) context.Context {
	if !custom {
		return ctx
	}
	return WithPubSubMonitoring(ctx, PubSubMonitorOptions{
		Now: func() time.Duration { return time.Duration(z.MonoOffset()) },
	})
}

func pubSubBenchmarkQueue(b *testing.B, custom bool) (PubSub, func(Message)) {
	b.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	p := newPubSub(pubSubBenchmarkContext(ctx, custom), &client{v: NewConf()}, nil).(*pubSub)
	b.Cleanup(func() { cancel(); _ = p.Close() })
	// 仅在基准初始化时选择入口，计时循环中两者都经同一种函数调用。
	if receiver, ok := any(p).(interface {
		enqueueMessage(context.Context, Message)
	}); ok {
		return p, func(message Message) { receiver.enqueueMessage(ctx, message) }
	}
	receiver, ok := any(p.monitor).(interface {
		onMessage(context.Context, Message)
	})
	if !ok {
		b.Fatal("unsupported PubSub message entry")
	}
	return p, func(message Message) { receiver.onMessage(ctx, message) }
}

func BenchmarkPubSubMonitorCompareSerial(b *testing.B) {
	for _, custom := range []bool{false, true} {
		b.Run(fmt.Sprintf("custom=%v", custom), func(b *testing.B) {
			p, enqueue := pubSubBenchmarkQueue(b, custom)
			out := p.Channel()
			message := Message{Channel: "bench", Message: strings.Repeat("x", 256)}
			b.ReportAllocs()
			b.SetBytes(256)
			for b.Loop() {
				enqueue(message)
				<-out
			}
		})
	}
}

func BenchmarkPubSubMonitorCompareBurst(b *testing.B) {
	const batch = 256
	for _, custom := range []bool{false, true} {
		b.Run(fmt.Sprintf("custom=%v", custom), func(b *testing.B) {
			p, enqueue := pubSubBenchmarkQueue(b, custom)
			out := p.Channel()
			message := Message{Channel: "bench", Message: strings.Repeat("x", 256)}
			b.ReportAllocs()
			b.SetBytes(256 * batch)
			for b.Loop() {
				for range batch {
					enqueue(message)
				}
				for range batch {
					<-out
				}
			}
			b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*batch), "ns/message")
		})
	}
}

func BenchmarkPubSubMonitorCompareParallel(b *testing.B) {
	const window = 1024
	for _, custom := range []bool{false, true} {
		b.Run(fmt.Sprintf("custom=%v", custom), func(b *testing.B) {
			p, enqueue := pubSubBenchmarkQueue(b, custom)
			out := p.Channel()
			message := Message{Channel: "bench", Message: strings.Repeat("x", 256)}
			credits := make(chan struct{}, window)
			done := make(chan struct{})
			go func() {
				for range b.N {
					<-out
					<-credits
				}
				close(done)
			}()
			b.ReportAllocs()
			b.SetBytes(256)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					credits <- struct{}{}
					enqueue(message)
				}
			})
			<-done
			b.StopTimer()
		})
	}
}

func BenchmarkPubSubMonitorCompareInit(b *testing.B) {
	conf := NewConf()
	for _, custom := range []bool{false, true} {
		b.Run(fmt.Sprintf("custom=%v", custom), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				ctx, cancel := context.WithCancel(context.Background())
				p := newPubSub(pubSubBenchmarkContext(ctx, custom), &client{v: conf}, nil)
				_ = p.Close()
				<-p.Channel()
				cancel()
			}
		})
	}
}

// 连接断言在建连前执行，基准只访问 TestMain 启动的私有 Redis。
func BenchmarkPubSubMonitorCompareRedis(b *testing.B) {
	for _, custom := range []bool{false, true} {
		b.Run(fmt.Sprintf("custom=%v", custom), func(b *testing.B) {
			conf := NewConf(WithDevelopment(false), WithEnableMonitor(false), WithEnableCache(false))
			if conf.Net != "unix" {
				b.Fatal("benchmark requires its owned Redis socket")
			}
			client := MustNewClient(conf)
			defer func() { _ = client.Close() }()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			ack := make(chan struct{}, 1)
			ctx = rueidis.WithOnSubscriptionHook(ctx, func(rueidis.PubSubSubscription) {
				select {
				case ack <- struct{}{}:
				default:
				}
			})
			ctx = pubSubBenchmarkContext(ctx, custom)
			p := client.Subscribe(ctx, "bench.delivery")
			defer func() { _ = p.Close() }()
			timer := time.NewTimer(5 * time.Second)
			defer timer.Stop()
			select {
			case <-ack:
			case <-timer.C:
				b.Fatal("benchmark subscription timed out")
			}
			out := p.Channel()
			message := strings.Repeat("x", 256)
			b.ReportAllocs()
			b.SetBytes(256)
			for b.Loop() {
				if err := client.Publish(ctx, "bench.delivery", message).Err(); err != nil {
					b.Fatal(err)
				}
				if received := <-out; received.Message != message {
					b.Fatal("benchmark payload changed")
				}
			}
		})
	}
}

// 批量 PUBLISH 模拟连续通知流，消费与发布并行，避免单条往返掩盖接收开销。
func BenchmarkPubSubMonitorCompareRedisStream(b *testing.B) {
	const batch = 256
	for _, custom := range []bool{false, true} {
		b.Run(fmt.Sprintf("custom=%v", custom), func(b *testing.B) {
			conf := NewConf(WithDevelopment(false), WithEnableMonitor(false), WithEnableCache(false))
			if conf.Net != "unix" {
				b.Fatal("benchmark requires its owned Redis socket")
			}
			publisher := MustNewClient(conf)
			subscriber := MustNewClient(conf)
			defer func() { _ = publisher.Close(); _ = subscriber.Close() }()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			ack := make(chan struct{}, 1)
			ctx = rueidis.WithOnSubscriptionHook(ctx, func(rueidis.PubSubSubscription) {
				select {
				case ack <- struct{}{}:
				default:
				}
			})
			p := subscriber.Subscribe(pubSubBenchmarkContext(ctx, custom), "bench.stream")
			defer func() { _ = p.Close() }()
			timer := time.NewTimer(5 * time.Second)
			defer timer.Stop()
			select {
			case <-ack:
			case <-timer.C:
				b.Fatal("benchmark subscription timed out")
			}
			commandClient := publisher.(*client).cmd
			message := strings.Repeat("x", 256)
			commands := make([]rueidis.Completed, batch)
			for i := range commands {
				commands[i] = commandClient.B().Publish().Channel("bench.stream").Message(message).Build().Pin()
			}
			out := p.Channel()
			completed := make(chan struct{})
			go func() {
				count := 0
				for range out {
					count++
					if count%batch == 0 {
						select {
						case completed <- struct{}{}:
						case <-ctx.Done():
							return
						}
					}
				}
			}()
			b.ReportAllocs()
			b.SetBytes(256 * batch)
			for b.Loop() {
				for _, result := range commandClient.DoMulti(ctx, commands...) {
					if err := result.Error(); err != nil {
						b.Fatal(err)
					}
				}
				<-completed
			}
			b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*batch), "ns/message")
		})
	}
}
