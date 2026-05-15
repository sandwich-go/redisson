package redisson

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

// TestRunTickerRecoverFromPanic 验证 runTicker 内 fn panic 时 recover 后
// ticker 仍继续:单次 panic 不应让整个 delayQueue 静默失效。
func TestRunTickerRecoverFromPanic(t *testing.T) {
	q := &delayQueue{
		name:  "panic-test",
		exitC: make(chan struct{}),
	}
	q.running.Store(true)

	// 用极短 interval 加快测试
	const interval = 5 * time.Millisecond
	var calls atomic.Int32
	var successCalls atomic.Int32
	fn := func() {
		c := calls.Add(1)
		// 前两次 panic，验证 recover 后 ticker 仍然继续；第三次起正常返回。
		if c <= 2 {
			panic(fmt.Sprintf("synthetic panic #%d", c))
		}
		successCalls.Add(1)
	}

	q.tickerWG.Add(1)
	go q.runTicker(interval, fn)

	// 给足够时间触发 ≥3 次 tick
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if successCalls.Load() >= 1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	close(q.exitC)
	q.tickerWG.Wait()

	total := calls.Load()
	good := successCalls.Load()
	if total < 3 || good < 1 {
		t.Fatalf("expected ticker to recover from panic and continue (total>=3, success>=1), got total=%d success=%d", total, good)
	}
}
