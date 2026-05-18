//go:build integration

package redisson

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestClient_Close_TimeoutBreaksDeadlock 验证 client.Close 在 callback 永久阻塞
// 的反模式场景下能依靠 clientCloseWaitTimeout 超时打破死锁,不会让 t.Cleanup 卡死
// 整个测试进程 (CI 上曾出现 5 分钟 panic 现象,见 GitHub Actions Redis 6 矩阵)。
//
// 反模式:用户 callback 内通过自定义 chan 等待外部信号,但外部信号永远不来。
// 此时 worker 等 callback,q.Close 等 worker (5s 后超时), client.Close 等
// delayWorkerWG (本测试验证此处 10s 超时也能强制返回)。
func TestClient_Close_TimeoutBreaksDeadlock(t *testing.T) {
	if !realRedisAvailable {
		t.Skip("real Redis not available; this test runs under -tags integration")
	}
	c := MustNewClient(NewConf(WithDevelopment(false), WithDB(nextDB())))
	ctx := context.Background()
	c.FlushDB(ctx)

	// callback 永久阻塞 (用户反模式: 等永远不会来的信号)。
	blockForever := make(chan struct{})
	var cbStarted sync.WaitGroup
	cbStarted.Add(1)
	q, err := c.NewDelayQueue("close-timeout-test", func(_ []byte) error {
		cbStarted.Done()
		<-blockForever
		return nil
	}, WithDelayOptionPrefix("close_timeout_test"), WithDelayOptionVisibilityTimeout(10*time.Second))
	if err != nil {
		t.Fatalf("NewDelayQueue: %v", err)
	}
	defer close(blockForever) // 测试结束后让 callback 退出避免 goroutine 泄漏

	// 投递 task,确保 callback 被触发
	if err := q.Add(ctx, []byte("p"), 100*time.Millisecond); err != nil {
		t.Fatalf("Add: %v", err)
	}
	cbStarted.Wait() // callback 进入 blockForever 阻塞

	// client.Close 应在 clientCloseWaitTimeout (10s) 内返回,即使 callback 永不返回。
	closeStart := time.Now()
	closeDone := make(chan error, 1)
	go func() { closeDone <- c.Close() }()
	select {
	case <-closeDone:
		elapsed := time.Since(closeStart)
		// q.Close 5s + client.Close 10s = 15s 上限; 加余量
		if elapsed > 20*time.Second {
			t.Fatalf("client.Close took %v, want < 20s", elapsed)
		}
	case <-time.After(20 * time.Second):
		t.Fatalf("client.Close blocked > 20s; timeout did not break deadlock")
	}

	// q.Close 内部已经 5s 超时返回,这里再调一次应直接 ErrDelayQueueHasClosed
	if err := q.Close(); err != ErrDelayQueueHasClosed {
		t.Fatalf("q.Close after client.Close: got %v, want ErrDelayQueueHasClosed", err)
	}
}
