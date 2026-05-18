//go:build integration

package redisson

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// 本文件补充 delay.go 的 P0/P1 行为测试，覆盖现有 delay_test.go 的 happy path 之外的：
//   - bug 不复发回归（retry-not-via-reclaim、reclaim-not-too-early、Close 幂等等）
//   - API 边界（空 name/nil callback/重复注册/Close 后调用）
//   - 并发/异常（callback panic、并发 Add、callback 内 Close 不死锁）
//
// 全部使用 t.Parallel() + nextDB() 并行运行；每个测试独立 client，避免污染。

// delayExtraDB 本文件所有测试共享的 Redis db。
//
// 不使用 helpers_test.go 中的 nextDB（它给每个测试分配 1..15 之一）：
//   - 本文件 17 个并行 delay 测试 + 仓库内其他 ~45 个 cmd 测试都走 nextDB，
//     必然会撞 db slot；
//   - 而 cmd 测试通过 doTestUnitClean -> FlushDB 频繁清当前 db，会把 delay 测试
//     正在使用的 ZSET/HASH 清空，导致 callback 拉不到任务。
//
// 这里固定使用 db 0（helpers_test.go 注释明确留给"手动调试"，常规测试不使用），
// 配合每个测试独占的 uniquePrefix 实现 key 级别的隔离，互不干扰。
const delayExtraDB = 0

// newTestDelayClient 构造一个 standalone client；t.Cleanup 自动 Close client。
//
// 启动时清掉 uniquePrefix(t):* 下的全部残留 key（避免重跑同名测试时受上轮影响）；
// 不调 FlushDB（避免影响并行运行的其他测试 prefix）。
func newTestDelayClient(t *testing.T) Cmdable {
	t.Helper()
	c := MustNewClient(NewConf(WithDevelopment(false), WithDB(delayExtraDB)))
	t.Cleanup(func() { _ = c.Close() })
	purgePrefix(t, c, uniquePrefix(t))
	return c
}

// purgePrefix 清掉给定 prefix 开头的所有 key（含 :delay:{name}/:doing:{name}/:meta:{name}）。
// 使用 SCAN（O(N) 但分批拉取，不阻塞 Redis）。
func purgePrefix(t *testing.T, c Cmdable, prefix string) {
	t.Helper()
	ctx := context.Background()
	var cursor uint64
	for {
		// 不直接使用 KEYS：在共享实例上是阻塞调用。改用 SCAN。
		keys, next, err := c.Scan(ctx, cursor, prefix+"*", 200).Result()
		if err != nil {
			t.Fatalf("scan: %v", err)
		}
		for _, k := range keys {
			_ = c.Del(ctx, k).Err()
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
}

// uniquePrefix 用 t.Name() 拼出当次测试独占的 Redis key 前缀。
func uniquePrefix(t *testing.T) string {
	t.Helper()
	return "delay_extra:" + t.Name()
}

// ============================================================================
// P0 回归：retry 路径独立于 reclaim
// ============================================================================

// TestDelay_RetryViaRetryPath_NotReclaim 验证业务失败 → 通过 retry 路径快速重试
// 而不是依赖 reclaim 兜底。
//
// 设计：Timeout 设为 60s，使 reclaim 在测试窗口内（5s）绝对不可能触发；
// 默认 retry backoff = 1s，只有 retry 路径能让 callback 在 5s 内被多次调用。
//
// 旧版 bug：moveDelayTaskLua 把 doing score 覆盖为 now+timeout，
// poll 返回的 score 永远 > 0，retry 计数读不到，导致重试只能靠 reclaim 兜底；
// 此用例下旧版会超时失败。
func TestDelay_RetryViaRetryPath_NotReclaim(t *testing.T) {
	t.Parallel()
	c := newTestDelayClient(t)
	ctx := context.Background()

	var attempts atomic.Int32
	done := make(chan struct{})
	q, err := c.NewDelayQueue("retry-only", func(_ []byte) error {
		n := attempts.Add(1)
		if n >= 3 {
			close(done)
			return nil
		}
		return errors.New("transient")
	},
		WithDelayOptionPrefix(uniquePrefix(t)),
		WithDelayOptionVisibilityTimeout(60*time.Second),
		WithDelayOptionRetryTimes(5),
	)
	if err != nil {
		t.Fatalf("NewDelayQueue: %v", err)
	}
	t.Cleanup(func() { _ = q.Close() })

	if err := q.Add(ctx, []byte("payload"), 500*time.Millisecond); err != nil {
		t.Fatalf("Add: %v", err)
	}

	select {
	case <-done:
		// 至少 3 次：第 1 次 fail、第 2 次 fail、第 3 次 success
		if got := attempts.Load(); got < 3 {
			t.Fatalf("attempts=%d, want >=3", got)
		}
	case <-time.After(8 * time.Second):
		t.Fatalf("retry never converged within 8s; attempts=%d", attempts.Load())
	}
}

// TestDelay_DeadLetterViaRetryPath_NotReclaim 验证累计达到 RetryTimes 后通过 retry 路径触发死信，
// 不依赖 reclaim 兜底。
//
// 设计：Timeout=60s 让 reclaim 在 10s 窗口内不可能触发；RetryTimes=3，
// callback 永远 fail，期望 3 次失败后死信 hook 被调用。
func TestDelay_DeadLetterViaRetryPath_NotReclaim(t *testing.T) {
	t.Parallel()
	c := newTestDelayClient(t)
	ctx := context.Background()

	var attempts atomic.Int32
	dead := make(chan []byte, 1)
	q, err := c.NewDelayQueue("dead-only", func(_ []byte) error {
		attempts.Add(1)
		return errors.New("perm")
	},
		WithDelayOptionPrefix(uniquePrefix(t)),
		WithDelayOptionVisibilityTimeout(60*time.Second),
		WithDelayOptionRetryTimes(3),
		WithDelayOptionHandleDeadLetter(func(bs []byte) {
			select {
			case dead <- bs:
			default:
			}
		}),
	)
	if err != nil {
		t.Fatalf("NewDelayQueue: %v", err)
	}
	t.Cleanup(func() { _ = q.Close() })

	if err := q.Add(ctx, []byte("xx"), 500*time.Millisecond); err != nil {
		t.Fatalf("Add: %v", err)
	}

	select {
	case <-dead:
		if got := attempts.Load(); got != 3 {
			t.Fatalf("attempts=%d, want exactly 3 (RetryTimes=3, retry 路径触发死信)", got)
		}
		// 死信后队列应清空
		eventuallyEqOnce(t, func() int64 {
			l, _ := q.Length(ctx)
			return l
		}, 0, 3*time.Second)
	case <-time.After(10 * time.Second):
		t.Fatalf("dead letter never fired within 10s; attempts=%d", attempts.Load())
	}
}

// ============================================================================
// P0 回归：reclaim 不会过早触发
// ============================================================================

// TestDelay_StuckCallbackDoesNotReschedule 验证 callback 执行时间超过 visibility timeout
// 时，本实例的 reclaim 不会重复调度同一 payload（heartbeat 续期保证）。
//
// 设计:Timeout=2s (visibility=2s),callback 阻塞 5s (远超 visibility)。
// worker 用 heartbeat 每 visibility/3 续 doing score,visibility 永远在未来,
// reclaim 拉不到自己持有的 task;callback 跑完后通过 score fence 安全 ack。
func TestDelay_StuckCallbackDoesNotReschedule(t *testing.T) {
	t.Parallel()
	c := newTestDelayClient(t)
	ctx := context.Background()

	var totalCalls atomic.Int32
	done := make(chan struct{})
	q, err := c.NewDelayQueue("stuck-cb", func(_ []byte) error {
		totalCalls.Add(1)
		// 阻塞远超 visibility，模拟卡住的 callback
		time.Sleep(5 * time.Second)
		select {
		case <-done:
		default:
			close(done)
		}
		return nil
	},
		WithDelayOptionPrefix(uniquePrefix(t)),
		WithDelayOptionVisibilityTimeout(2*time.Second), // visibility=2s
		WithDelayOptionRetryTimes(3),
	)
	if err != nil {
		t.Fatalf("NewDelayQueue: %v", err)
	}
	t.Cleanup(func() { _ = q.Close() })

	if err := q.Add(ctx, []byte("p"), 500*time.Millisecond); err != nil {
		t.Fatalf("Add: %v", err)
	}

	// 等第一次完成
	select {
	case <-done:
	case <-time.After(15 * time.Second):
		t.Fatalf("callback never completed; totalCalls=%d", totalCalls.Load())
	}

	// callback 完成后再多等一段，看 reclaim 是否多触发了几次
	time.Sleep(3 * time.Second)

	if got := totalCalls.Load(); got != 1 {
		t.Fatalf("totalCalls=%d, want 1 (callback 卡住超过 visibility 期间不应被重复调度，"+
			"heartbeat 应保证 doing score 始终在未来，reclaim 拉不到)", got)
	}
}

// TestDelay_ReclaimNotTooEarly 验证一个正在被处理的 item 在 visibility 内
// 不会被 reclaim 拉回造成重复消费。
//
// 设计：callback 阻塞 2s 后成功；Timeout=10s（visibility=10s）。
// 旧版 bug：reclaim 上界用 now+timeout 而非 now，doing 中所有 item（score=now+timeout）
// 都会被立刻拉回 delay，导致 callback 被并发触发多次。
func TestDelay_ReclaimNotTooEarly(t *testing.T) {
	t.Parallel()
	c := newTestDelayClient(t)
	ctx := context.Background()

	var inFlight atomic.Int32
	var maxInFlight atomic.Int32
	var totalCalls atomic.Int32
	done := make(chan struct{})
	q, err := c.NewDelayQueue("reclaim-window", func(_ []byte) error {
		cur := inFlight.Add(1)
		// 记录峰值并发
		for {
			m := maxInFlight.Load()
			if cur <= m || maxInFlight.CompareAndSwap(m, cur) {
				break
			}
		}
		defer inFlight.Add(-1)
		totalCalls.Add(1)
		time.Sleep(2 * time.Second)
		// 第一次成功就关闭 done；后续失败说明被重复触发了
		select {
		case <-done:
		default:
			close(done)
		}
		return nil
	},
		WithDelayOptionPrefix(uniquePrefix(t)),
		WithDelayOptionVisibilityTimeout(10*time.Second),
		WithDelayOptionRetryTimes(3),
	)
	if err != nil {
		t.Fatalf("NewDelayQueue: %v", err)
	}
	t.Cleanup(func() { _ = q.Close() })

	if err := q.Add(ctx, []byte("p"), 500*time.Millisecond); err != nil {
		t.Fatalf("Add: %v", err)
	}

	<-done
	// 给 reclaim ticker 几次机会触发（visibility=10s，下面只等 4s 远小于）
	time.Sleep(4 * time.Second)

	if got := maxInFlight.Load(); got > 1 {
		t.Fatalf("maxInFlight=%d, want=1 (callback 不应被并发触发)", got)
	}
	if got := totalCalls.Load(); got != 1 {
		t.Fatalf("totalCalls=%d, want=1 (callback 不应被重复触发)", got)
	}
}

// ============================================================================
// P0 回归：NewDelayQueue 严格语义
// ============================================================================

// TestDelay_NewDelayQueue_DuplicateRejected 同名重复注册必须返回
// ErrDelayQueueHasRegistered，第一份实例不受影响。
func TestDelay_NewDelayQueue_DuplicateRejected(t *testing.T) {
	t.Parallel()
	c := newTestDelayClient(t)

	q1, err := c.NewDelayQueue("dup-name", func(_ []byte) error { return nil },
		WithDelayOptionPrefix(uniquePrefix(t)))
	if err != nil {
		t.Fatalf("first NewDelayQueue: %v", err)
	}
	t.Cleanup(func() { _ = q1.Close() })

	q2, err := c.NewDelayQueue("dup-name", func(_ []byte) error { return nil },
		WithDelayOptionPrefix(uniquePrefix(t)))
	if !errors.Is(err, ErrDelayQueueHasRegistered) {
		t.Fatalf("second NewDelayQueue err=%v, want ErrDelayQueueHasRegistered", err)
	}
	if q2 != nil {
		t.Fatalf("second NewDelayQueue returned non-nil queue alongside error")
	}
	// 第一份实例仍可用
	if err := q1.Add(context.Background(), []byte("ping"), time.Second); err != nil {
		t.Fatalf("q1 still usable check failed: %v", err)
	}
}

// TestDelay_NewDelayQueue_ReuseAfterClose 同名队列 Close 之后可以重新创建。
func TestDelay_NewDelayQueue_ReuseAfterClose(t *testing.T) {
	t.Parallel()
	c := newTestDelayClient(t)

	q1, err := c.NewDelayQueue("reuse", func(_ []byte) error { return nil },
		WithDelayOptionPrefix(uniquePrefix(t)))
	if err != nil {
		t.Fatalf("first NewDelayQueue: %v", err)
	}
	if err := q1.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	q2, err := c.NewDelayQueue("reuse", func(_ []byte) error { return nil },
		WithDelayOptionPrefix(uniquePrefix(t)))
	if err != nil {
		t.Fatalf("second NewDelayQueue after Close: %v", err)
	}
	t.Cleanup(func() { _ = q2.Close() })
}

// ============================================================================
// P0 回归：Close 幂等 + double close 不 panic
// ============================================================================

// TestDelay_Close_Idempotent 第二次 Close 返回 ErrDelayQueueHasClosed，不 panic。
func TestDelay_Close_Idempotent(t *testing.T) {
	t.Parallel()
	c := newTestDelayClient(t)

	q, err := c.NewDelayQueue("close-idem", func(_ []byte) error { return nil },
		WithDelayOptionPrefix(uniquePrefix(t)))
	if err != nil {
		t.Fatalf("NewDelayQueue: %v", err)
	}

	if err := q.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := q.Close(); !errors.Is(err, ErrDelayQueueHasClosed) {
		t.Fatalf("second Close err=%v, want ErrDelayQueueHasClosed", err)
	}
	if err := q.Close(); !errors.Is(err, ErrDelayQueueHasClosed) {
		t.Fatalf("third Close err=%v, want ErrDelayQueueHasClosed", err)
	}
}

// TestDelay_Close_ConcurrentNoPanic 多 goroutine 同时 Close 不应 double close
// channel panic；只有一个能拿到 nil。
func TestDelay_Close_ConcurrentNoPanic(t *testing.T) {
	t.Parallel()
	c := newTestDelayClient(t)

	q, err := c.NewDelayQueue("close-race", func(_ []byte) error { return nil },
		WithDelayOptionPrefix(uniquePrefix(t)))
	if err != nil {
		t.Fatalf("NewDelayQueue: %v", err)
	}

	const N = 32
	var wg sync.WaitGroup
	var nilCount, closedCount atomic.Int32
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := q.Close()
			switch {
			case err == nil:
				nilCount.Add(1)
			case errors.Is(err, ErrDelayQueueHasClosed):
				closedCount.Add(1)
			default:
				t.Errorf("unexpected Close err: %v", err)
			}
		}()
	}
	wg.Wait()

	if got := nilCount.Load(); got != 1 {
		t.Fatalf("nilCount=%d, want exactly 1", got)
	}
	if got := closedCount.Load(); got != N-1 {
		t.Fatalf("closedCount=%d, want %d", got, N-1)
	}
}

// ============================================================================
// API 边界
// ============================================================================

// TestDelay_NewDelayQueue_EmptyName 空 name 返回 ErrEmptyDelayQueueName。
func TestDelay_NewDelayQueue_EmptyName(t *testing.T) {
	t.Parallel()
	c := newTestDelayClient(t)

	q, err := c.NewDelayQueue("", func(_ []byte) error { return nil })
	if !errors.Is(err, ErrEmptyDelayQueueName) {
		t.Fatalf("err=%v, want ErrEmptyDelayQueueName", err)
	}
	if q != nil {
		t.Fatalf("queue should be nil on error")
	}
}

// TestDelay_NewDelayQueue_NilCallback nil callback 返回 ErrEmptyDelayQueueCallback。
func TestDelay_NewDelayQueue_NilCallback(t *testing.T) {
	t.Parallel()
	c := newTestDelayClient(t)

	q, err := c.NewDelayQueue("with-nil-cb", nil)
	if !errors.Is(err, ErrEmptyDelayQueueCallback) {
		t.Fatalf("err=%v, want ErrEmptyDelayQueueCallback", err)
	}
	if q != nil {
		t.Fatalf("queue should be nil on error")
	}
}

// TestDelay_AddDelClose_AfterClose Close 后调用 Add/Del 返回 ErrDelayQueueHasClosed。
func TestDelay_AddDelClose_AfterClose(t *testing.T) {
	t.Parallel()
	c := newTestDelayClient(t)

	q, err := c.NewDelayQueue("after-close", func(_ []byte) error { return nil },
		WithDelayOptionPrefix(uniquePrefix(t)))
	if err != nil {
		t.Fatalf("NewDelayQueue: %v", err)
	}
	if err := q.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if err := q.Add(context.Background(), []byte("p"), time.Second); !errors.Is(err, ErrDelayQueueHasClosed) {
		t.Fatalf("Add err=%v, want ErrDelayQueueHasClosed", err)
	}
	if err := q.Del(context.Background(), []byte("p")); !errors.Is(err, ErrDelayQueueHasClosed) {
		t.Fatalf("Del err=%v, want ErrDelayQueueHasClosed", err)
	}
}

// TestDelay_Length_AfterCloseStillWorks Length 在 Close 后仍可读（用于 collector 收尾）。
// 这是和 Add/Del 不同的设计：只读、且 collector 路径会触发。
func TestDelay_Length_AfterCloseStillWorks(t *testing.T) {
	t.Parallel()
	c := newTestDelayClient(t)
	ctx := context.Background()

	q, err := c.NewDelayQueue("len-after-close", func(_ []byte) error { return nil },
		WithDelayOptionPrefix(uniquePrefix(t)))
	if err != nil {
		t.Fatalf("NewDelayQueue: %v", err)
	}
	if err := q.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	// Close 后 Length 不返回 ErrDelayQueueHasClosed；Redis 调用本身返回 0/nil。
	l, err := q.Length(ctx)
	if err != nil {
		t.Fatalf("Length after Close: %v", err)
	}
	if l != 0 {
		t.Fatalf("Length=%d, want 0", l)
	}
}

// TestDelay_Name 返回创建时传入的 name（不含 prefix）。
func TestDelay_Name(t *testing.T) {
	t.Parallel()
	c := newTestDelayClient(t)

	q, err := c.NewDelayQueue("named", func(_ []byte) error { return nil },
		WithDelayOptionPrefix(uniquePrefix(t)))
	if err != nil {
		t.Fatalf("NewDelayQueue: %v", err)
	}
	t.Cleanup(func() { _ = q.Close() })
	if got := q.Name(); got != "named" {
		t.Fatalf("Name=%q, want %q", got, "named")
	}
}

// ============================================================================
// 行为细节
// ============================================================================

// TestDelay_AddOverridesExpiry 同 payload 二次 Add 覆盖到期时间。
// 第一次 Add delay=10s，第二次 Add delay=1s 应在 ~1s 后被消费。
func TestDelay_AddOverridesExpiry(t *testing.T) {
	t.Parallel()
	c := newTestDelayClient(t)
	ctx := context.Background()

	got := make(chan time.Time, 1)
	q, err := c.NewDelayQueue("override-exp", func(_ []byte) error {
		select {
		case got <- nowFunc():
		default:
		}
		return nil
	}, WithDelayOptionPrefix(uniquePrefix(t)))
	if err != nil {
		t.Fatalf("NewDelayQueue: %v", err)
	}
	t.Cleanup(func() { _ = q.Close() })

	if err := q.Add(ctx, []byte("p"), 10*time.Second); err != nil {
		t.Fatalf("first Add: %v", err)
	}
	start := nowFunc()
	if err := q.Add(ctx, []byte("p"), 1*time.Second); err != nil {
		t.Fatalf("second Add: %v", err)
	}

	select {
	case at := <-got:
		elapsed := at.Sub(start)
		// 留余量：1s 延迟 + 1s ticker + 1s 抖动
		if elapsed > 4*time.Second {
			t.Fatalf("callback fired after %v, want <=4s (override should take effect)", elapsed)
		}
	case <-time.After(6 * time.Second):
		t.Fatalf("callback never fired; second Add did not override expiry")
	}
}

// TestDelay_AddOverridesResetsRetryCount 二次 Add 应清重试计数。
//
// 步骤：先让 payload 失败 1 次（attempts=1），然后用新 Add 覆盖；
// 后续业务永远成功，attempts 不应再受历史失败计数影响导致死信。
func TestDelay_AddOverridesResetsRetryCount(t *testing.T) {
	t.Parallel()
	c := newTestDelayClient(t)
	ctx := context.Background()

	var phase atomic.Int32 // 0=fail, 1=succeed
	done := make(chan struct{})
	q, err := c.NewDelayQueue("reset-retry", func(_ []byte) error {
		if phase.Load() == 0 {
			return errors.New("phase0 fail")
		}
		select {
		case <-done:
		default:
			close(done)
		}
		return nil
	},
		WithDelayOptionPrefix(uniquePrefix(t)),
		WithDelayOptionVisibilityTimeout(60*time.Second),
		WithDelayOptionRetryTimes(2), // 失败 2 次就死信
	)
	if err != nil {
		t.Fatalf("NewDelayQueue: %v", err)
	}
	t.Cleanup(func() { _ = q.Close() })

	if err := q.Add(ctx, []byte("p"), 500*time.Millisecond); err != nil {
		t.Fatalf("Add: %v", err)
	}

	// 等失败一次（meta=1）
	time.Sleep(2500 * time.Millisecond)

	// 切换 phase 并覆盖 Add（应清 meta）
	phase.Store(1)
	if err := q.Add(ctx, []byte("p"), 500*time.Millisecond); err != nil {
		t.Fatalf("override Add: %v", err)
	}

	select {
	case <-done:
	case <-time.After(6 * time.Second):
		t.Fatalf("override Add did not reset retry count (item went dead-letter unexpectedly)")
	}
}

// TestDelay_AddNegativeDelay 负 delay 钳到 0，立即可见。
func TestDelay_AddNegativeDelay(t *testing.T) {
	t.Parallel()
	c := newTestDelayClient(t)
	ctx := context.Background()

	got := make(chan struct{})
	q, err := c.NewDelayQueue("neg-delay", func(_ []byte) error {
		select {
		case <-got:
		default:
			close(got)
		}
		return nil
	}, WithDelayOptionPrefix(uniquePrefix(t)))
	if err != nil {
		t.Fatalf("NewDelayQueue: %v", err)
	}
	t.Cleanup(func() { _ = q.Close() })

	if err := q.Add(ctx, []byte("p"), -5*time.Second); err != nil {
		t.Fatalf("Add: %v", err)
	}
	select {
	case <-got:
	case <-time.After(3 * time.Second):
		t.Fatalf("negative delay should fire immediately within next poll cycle")
	}
}

// TestDelay_DelInDelayState Add 后立即 Del，Length=0 且 callback 不会被触发。
func TestDelay_DelInDelayState(t *testing.T) {
	t.Parallel()
	c := newTestDelayClient(t)
	ctx := context.Background()

	var fired atomic.Int32
	q, err := c.NewDelayQueue("del-delay", func(_ []byte) error {
		fired.Add(1)
		return nil
	}, WithDelayOptionPrefix(uniquePrefix(t)))
	if err != nil {
		t.Fatalf("NewDelayQueue: %v", err)
	}
	t.Cleanup(func() { _ = q.Close() })

	if err := q.Add(ctx, []byte("p"), 3*time.Second); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if err := q.Del(ctx, []byte("p")); err != nil {
		t.Fatalf("Del: %v", err)
	}
	l, err := q.Length(ctx)
	if err != nil {
		t.Fatalf("Length: %v", err)
	}
	if l != 0 {
		t.Fatalf("Length=%d, want 0 after Del", l)
	}
	// 等过原本应该触发的时间窗口
	time.Sleep(5 * time.Second)
	if got := fired.Load(); got != 0 {
		t.Fatalf("callback fired %d times after Del; want 0", got)
	}
}

// TestDelay_LengthCounts 验证 Length 同时统计 delay+doing。
//
// 加 3 个 item，等其中 1 个被 poll 拉走（进 doing 但 callback 阻塞），
// Length 应为 3（delay=2 + doing=1）。
func TestDelay_LengthCounts(t *testing.T) {
	t.Parallel()
	c := newTestDelayClient(t)
	ctx := context.Background()

	block := make(chan struct{})
	var picked atomic.Bool
	q, err := c.NewDelayQueue("length-counts", func(_ []byte) error {
		picked.Store(true)
		<-block
		return nil
	},
		WithDelayOptionPrefix(uniquePrefix(t)),
		WithDelayOptionVisibilityTimeout(60*time.Second),
	)
	if err != nil {
		t.Fatalf("NewDelayQueue: %v", err)
	}
	t.Cleanup(func() {
		close(block) // 解除 callback 阻塞，允许 Close 后正在跑的 goroutine 退出
		_ = q.Close()
	})

	// 分别 0.5s / 5s / 5s 延迟
	if err := q.Add(ctx, []byte("p1"), 500*time.Millisecond); err != nil {
		t.Fatalf("Add p1: %v", err)
	}
	if err := q.Add(ctx, []byte("p2"), 5*time.Second); err != nil {
		t.Fatalf("Add p2: %v", err)
	}
	if err := q.Add(ctx, []byte("p3"), 5*time.Second); err != nil {
		t.Fatalf("Add p3: %v", err)
	}

	// 起初 3 个全在 delay
	l, _ := q.Length(ctx)
	if l != 3 {
		t.Fatalf("initial Length=%d, want 3", l)
	}

	// 等 p1 被拉到 doing
	eventuallyEqOnce(t, picked.Load, true, 3*time.Second)

	l, _ = q.Length(ctx)
	if l != 3 {
		t.Fatalf("Length after p1 picked=%d, want 3 (delay=2+doing=1)", l)
	}
}

// ============================================================================
// 异常 / 并发
// ============================================================================

// TestDelay_CallbackPanicTriggersRetry callback panic 被 recover，
// 视为业务失败一次，触发 retry 路径。
func TestDelay_CallbackPanicTriggersRetry(t *testing.T) {
	t.Parallel()
	c := newTestDelayClient(t)
	ctx := context.Background()

	var attempts atomic.Int32
	done := make(chan struct{})
	q, err := c.NewDelayQueue("panic-retry", func(_ []byte) error {
		n := attempts.Add(1)
		if n == 1 {
			panic("boom")
		}
		select {
		case <-done:
		default:
			close(done)
		}
		return nil
	},
		WithDelayOptionPrefix(uniquePrefix(t)),
		WithDelayOptionVisibilityTimeout(60*time.Second),
		WithDelayOptionRetryTimes(3),
	)
	if err != nil {
		t.Fatalf("NewDelayQueue: %v", err)
	}
	t.Cleanup(func() { _ = q.Close() })

	if err := q.Add(ctx, []byte("p"), 500*time.Millisecond); err != nil {
		t.Fatalf("Add: %v", err)
	}

	select {
	case <-done:
		if got := attempts.Load(); got != 2 {
			t.Fatalf("attempts=%d, want 2 (panic 算 1 次失败 + 第 2 次成功)", got)
		}
	case <-time.After(6 * time.Second):
		t.Fatalf("panic recovery did not retry; attempts=%d", attempts.Load())
	}
}

// TestDelay_CallbackCallsCloseNoDeadlock callback 内部 q.Close() 不应死锁。
// q.Close 不等 worker（在途 worker 由 client.Close 兜底等待），
// 因此 callback 中同步调 q.Close 安全。
func TestDelay_CallbackCallsCloseNoDeadlock(t *testing.T) {
	t.Parallel()
	c := newTestDelayClient(t)
	ctx := context.Background()

	closed := make(chan error, 1)
	var q DelayQueue
	var err error
	q, err = c.NewDelayQueue("close-from-cb", func(_ []byte) error {
		closed <- q.Close()
		return nil
	}, WithDelayOptionPrefix(uniquePrefix(t)))
	if err != nil {
		t.Fatalf("NewDelayQueue: %v", err)
	}

	if err := q.Add(ctx, []byte("p"), 500*time.Millisecond); err != nil {
		t.Fatalf("Add: %v", err)
	}
	select {
	case e := <-closed:
		if e != nil {
			t.Fatalf("Close from callback err=%v, want nil", e)
		}
	case <-time.After(8 * time.Second):
		t.Fatalf("Close from callback deadlocked")
	}
}

// TestDelay_ConcurrentAdd 多 goroutine 并发 Add 不同 payload，全部入队成功。
func TestDelay_ConcurrentAdd(t *testing.T) {
	t.Parallel()
	c := newTestDelayClient(t)
	ctx := context.Background()

	q, err := c.NewDelayQueue("concurrent-add", func(_ []byte) error { return nil },
		WithDelayOptionPrefix(uniquePrefix(t)),
		WithDelayOptionVisibilityTimeout(60*time.Second),
	)
	if err != nil {
		t.Fatalf("NewDelayQueue: %v", err)
	}
	t.Cleanup(func() { _ = q.Close() })

	const N = 50
	var wg sync.WaitGroup
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			payload := []byte{byte(idx)}
			// 远期延迟，确保还没被 poll 拉走
			if err := q.Add(ctx, payload, 30*time.Second); err != nil {
				t.Errorf("Add %d: %v", idx, err)
			}
		}(i)
	}
	wg.Wait()

	l, err := q.Length(ctx)
	if err != nil {
		t.Fatalf("Length: %v", err)
	}
	if l != N {
		t.Fatalf("Length=%d, want %d", l, N)
	}
}

// TestDelay_PrefixIsolatesQueues 同 client 上不同 prefix 隔离同名队列的 key。
//
// 前一份失败,会发现两份共用 key:这里通过创建 q1 -> Add -> Length=1 -> Close,
// 然后创建同 name 不同 prefix 的 q2,Length 应=0。
func TestDelay_PrefixIsolatesQueues(t *testing.T) {
	t.Parallel()
	c := newTestDelayClient(t)
	ctx := context.Background()

	q1, err := c.NewDelayQueue("isolated", func(_ []byte) error { return nil },
		WithDelayOptionPrefix(uniquePrefix(t)+":p1"),
		WithDelayOptionVisibilityTimeout(60*time.Second),
	)
	if err != nil {
		t.Fatalf("q1 NewDelayQueue: %v", err)
	}
	if err := q1.Add(ctx, []byte("payload"), 30*time.Second); err != nil {
		t.Fatalf("q1 Add: %v", err)
	}
	l1, _ := q1.Length(ctx)
	if l1 != 1 {
		t.Fatalf("q1 Length=%d, want 1", l1)
	}
	if err := q1.Close(); err != nil {
		t.Fatalf("q1 Close: %v", err)
	}

	q2, err := c.NewDelayQueue("isolated", func(_ []byte) error { return nil },
		WithDelayOptionPrefix(uniquePrefix(t)+":p2"),
		WithDelayOptionVisibilityTimeout(60*time.Second),
	)
	if err != nil {
		t.Fatalf("q2 NewDelayQueue: %v", err)
	}
	t.Cleanup(func() { _ = q2.Close() })
	l2, _ := q2.Length(ctx)
	if l2 != 0 {
		t.Fatalf("q2 Length=%d, want 0 (prefix should isolate)", l2)
	}
}

// TestDelay_DeadLetterHookNil HandleDeadLetter=nil 时死信不 panic。
func TestDelay_DeadLetterHookNil(t *testing.T) {
	t.Parallel()
	c := newTestDelayClient(t)
	ctx := context.Background()

	var attempts atomic.Int32
	q, err := c.NewDelayQueue("dl-nil-hook", func(_ []byte) error {
		attempts.Add(1)
		return errors.New("perm")
	},
		WithDelayOptionPrefix(uniquePrefix(t)),
		WithDelayOptionVisibilityTimeout(60*time.Second),
		WithDelayOptionRetryTimes(2),
		WithDelayOptionHandleDeadLetter(nil),
	)
	if err != nil {
		t.Fatalf("NewDelayQueue: %v", err)
	}
	t.Cleanup(func() { _ = q.Close() })

	if err := q.Add(ctx, []byte("p"), 500*time.Millisecond); err != nil {
		t.Fatalf("Add: %v", err)
	}
	// 等到失败 2 次后死信清理；queue length=0
	eventuallyEqOnce(t, func() int64 {
		l, _ := q.Length(ctx)
		return l
	}, 0, 8*time.Second)
	if got := attempts.Load(); got != 2 {
		t.Fatalf("attempts=%d, want 2", got)
	}
}

// ============================================================================
// 测试辅助
// ============================================================================

// eventuallyEqOnce 是不依赖 Convey 上下文的 polling 等待（avoid global 'So' panic when called outside Convey）。
//
// 与 helpers_test.go 中的 eventuallyEq 不同：这里失败时直接 t.Fatalf，无需 Convey 注册。
func eventuallyEqOnce[T comparable](t *testing.T, actual func() T, expected T, timeout time.Duration) {
	t.Helper()
	if actual() == expected {
		return
	}
	deadline := time.Now().Add(timeout)
	tick := 10 * time.Millisecond
	for time.Now().Before(deadline) {
		time.Sleep(tick)
		if actual() == expected {
			return
		}
		if tick < 100*time.Millisecond {
			tick *= 2
		}
	}
	t.Fatalf("eventuallyEq timed out after %v: got %v want %v", timeout, actual(), expected)
}
