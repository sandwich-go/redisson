package redisson

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/rueidis"
	"github.com/sandwich-go/boost/z"
)

// 不注册到全局 Registry，回调操作与 Phantom 当前的接收等待直方图一致。
var pubSubBenchmarkWaits = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "benchmark_pubsub_queue_wait_seconds",
	Help:    "Sampled local PubSub forwarding wait.",
	Buckets: []float64{.00001, .0001, .001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
}, []string{"direction"})

var pubSubBenchmarkWait = pubSubBenchmarkWaits.WithLabelValues("receive")
var pubSubBenchmarkEnabled atomic.Bool

func init() {
	pubSubBenchmarkEnabled.Store(true)
}

type pubSubBenchmarkMode struct {
	name        string
	enabled     bool
	sampleEvery uint32
}

var pubSubBenchmarkModes = []pubSubBenchmarkMode{
	{name: "disabled"},
	{name: "full", enabled: true, sampleEvery: 1},
	{name: "sampled", enabled: true, sampleEvery: 64},
}

func pubSubBenchmarkContext(ctx context.Context, mode pubSubBenchmarkMode) context.Context {
	if !mode.enabled {
		return ctx
	}
	return WithPubSubMonitoring(ctx, PubSubMonitorOptions{
		Now:                func() time.Duration { return time.Duration(z.MonoOffset()) },
		DequeueSampleEvery: mode.sampleEvery,
		OnDequeue: func(wait time.Duration) {
			if pubSubBenchmarkEnabled.Load() {
				pubSubBenchmarkWait.Observe(wait.Seconds())
			}
		},
	})
}

func pubSubBenchmarkQueue(b *testing.B, mode pubSubBenchmarkMode) (PubSub, func(Message)) {
	b.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	p := newPubSub(pubSubBenchmarkContext(ctx, mode), &client{v: NewConf()}, nil).(*pubSub)
	b.Cleanup(func() { cancel(); _ = p.Close() })
	receive, release := p.monitor.onReceive(p.ctx)
	b.Cleanup(release)
	return p, receive
}

func BenchmarkPubSubMonitorCompareSerial(b *testing.B) {
	for _, mode := range pubSubBenchmarkModes {
		b.Run(mode.name, func(b *testing.B) {
			p, enqueue := pubSubBenchmarkQueue(b, mode)
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
	for _, mode := range pubSubBenchmarkModes {
		b.Run(mode.name, func(b *testing.B) {
			p, enqueue := pubSubBenchmarkQueue(b, mode)
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
	for _, mode := range pubSubBenchmarkModes {
		b.Run(mode.name, func(b *testing.B) {
			p, _ := pubSubBenchmarkQueue(b, mode)
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
				internal := p.(*pubSub)
				enqueue, release := internal.monitor.onReceive(internal.ctx)
				defer release()
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
	for _, mode := range pubSubBenchmarkModes {
		b.Run(mode.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				ctx, cancel := context.WithCancel(context.Background())
				p := newPubSub(pubSubBenchmarkContext(ctx, mode), &client{v: conf}, nil)
				_ = p.Close()
				<-p.Channel()
				cancel()
			}
		})
	}
}

// 连接断言在建连前执行，基准只访问 TestMain 启动的私有 Redis。
func BenchmarkPubSubMonitorCompareRedis(b *testing.B) {
	for _, mode := range pubSubBenchmarkModes {
		b.Run(mode.name, func(b *testing.B) {
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
			ctx = pubSubBenchmarkContext(ctx, mode)
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
	for _, mode := range pubSubBenchmarkModes {
		b.Run(mode.name, func(b *testing.B) {
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
			p := subscriber.Subscribe(pubSubBenchmarkContext(ctx, mode), "bench.stream")
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

// 分别量化单调时钟、字节计数和输出元数据记录，定位开启监控后的成本。
func BenchmarkPubSubMonitorRecording(b *testing.B) {
	b.Run("clock", func(b *testing.B) {
		var last time.Duration
		for b.Loop() {
			last = time.Duration(z.MonoOffset())
		}
		if last == 0 {
			b.Fatal("clock did not advance")
		}
	})
	for _, callback := range []bool{false, true} {
		b.Run(fmt.Sprintf("output/callback=%v", callback), func(b *testing.B) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var observe func(time.Duration)
			if callback {
				observe = func(time.Duration) {}
			}
			q := newPubSubMonitorQueue(ctx, 1024, func() time.Duration { return time.Duration(z.MonoOffset()) }, observe, nil, 1)
			value := pubSubQueuedMessage{message: Message{Channel: "bench", Message: "payload"}, enqueued: q.now()}
			value.bytes = messageBytes(value.message)
			buffer := newRingBuffer[pubSubQueuedMessage](2)
			b.ReportAllocs()
			for b.Loop() {
				q.onWrite(value, buffer)
			}
		})
	}
}

func BenchmarkPubSubMonitorSnapshot(b *testing.B) {
	for _, backlog := range []int{0, 64, 1024} {
		b.Run(fmt.Sprintf("backlog=%d", backlog), func(b *testing.B) {
			p, enqueue := pubSubBenchmarkQueue(b, pubSubBenchmarkMode{name: "sampled", enabled: true, sampleEvery: 64})
			queue := p.(*pubSub).monitor.queue
			message := Message{Channel: "bench", Message: strings.Repeat("x", 256)}
			for range backlog {
				enqueue(message)
			}
			timer := time.NewTimer(5 * time.Second)
			defer timer.Stop()
			for queue.sent.Load() != uint64(backlog) {
				select {
				case <-timer.C:
					b.Fatal("snapshot benchmark did not fill output")
				default:
					runtime.Gosched()
				}
			}
			wantBytes := int64(backlog) * messageBytes(message)
			b.ReportAllocs()
			for b.Loop() {
				depth, bytes, age, closed := queue.snapshot()
				if depth != backlog || bytes != wantBytes || age < 0 || closed {
					b.Fatalf("snapshot changed pending output: %d %d %v %v", depth, bytes, age, closed)
				}
			}
		})
	}
}
