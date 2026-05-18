package redisson

import (
	"errors"
	"testing"
	"time"
)

// options_gen_test.go 集中覆盖 gen_*_optiongen.go 中的 ApplyOption / Visitor / WatchDog 路径。
//
// 每个 *Options 类型（Bloom/Delay/Locker/RateLimiter）都是 optiongen 生成，
// 模式一致：newXxxOptions(opts) + Visitor + ApplyOption + InstallWatchDog。

// ----- BloomOptions -----

func TestBloomOptions_Defaults(t *testing.T) {
	opts := newBloomOptions()
	if opts.GetEnableReadOperation() {
		t.Errorf("EnableReadOperation default should be false")
	}
}

func TestBloomOptions_ApplyAndRevert(t *testing.T) {
	opts := newBloomOptions(WithBloomOptionEnableReadOperation(true))
	if !opts.GetEnableReadOperation() {
		t.Errorf("EnableReadOperation=false, want true")
	}
	prev := opts.ApplyOption(WithBloomOptionEnableReadOperation(false))
	if opts.GetEnableReadOperation() {
		t.Errorf("after Apply EnableReadOperation=true")
	}
	opts.ApplyOption(prev...)
	if !opts.GetEnableReadOperation() {
		t.Errorf("after revert EnableReadOperation=false")
	}
}

func TestBloomOptions_WatchDog(t *testing.T) {
	old := watchDogBloomOptions
	t.Cleanup(func() { watchDogBloomOptions = old })

	called := false
	InstallBloomOptionsWatchDog(func(cc *BloomOptions) { called = true })
	_ = newBloomOptions()
	if !called {
		t.Errorf("watchdog should be called")
	}
}

// ----- DelayOptions -----

func TestDelayOptions_Defaults(t *testing.T) {
	opts := newDelayOptions()
	if opts.GetVisibilityTimeout() <= 0 {
		t.Errorf("VisibilityTimeout default should be >0, got %v", opts.GetVisibilityTimeout())
	}
	if opts.GetRetryTimes() <= 0 {
		t.Errorf("RetryTimes default should be >0, got %d", opts.GetRetryTimes())
	}
	if opts.GetHandleDeadLetter() == nil {
		t.Errorf("HandleDeadLetter default should be non-nil func")
	}
	// 新增字段默认 0（落到 default* 常量兜底，由 delay.go 内部转换）。
	if opts.GetPollInterval() != 0 {
		t.Errorf("PollInterval default should be 0 (use defaultDelayPollInterval), got %v", opts.GetPollInterval())
	}
	if opts.GetPollBatch() != 0 {
		t.Errorf("PollBatch default should be 0 (use defaultDelayPollBatch), got %d", opts.GetPollBatch())
	}
	if opts.GetRedisOpTimeout() != 0 {
		t.Errorf("RedisOpTimeout default should be 0 (use defaultDelayRedisOpTimeout), got %v", opts.GetRedisOpTimeout())
	}
	if opts.GetRetryBackoff() != 0 {
		t.Errorf("RetryBackoff default should be 0 (use defaultDelayRetryBackoff), got %v", opts.GetRetryBackoff())
	}
}

func TestDelayOptions_ApplyAll(t *testing.T) {
	opts := newDelayOptions(
		WithDelayOptionPrefix("p"),
		WithDelayOptionVisibilityTimeout(2*time.Second),
		WithDelayOptionRetryTimes(7),
		WithDelayOptionHandleDeadLetter(nil),
		WithDelayOptionPollInterval(500*time.Millisecond),
		WithDelayOptionPollBatch(32),
		WithDelayOptionRedisOpTimeout(3*time.Second),
		WithDelayOptionRetryBackoff(2*time.Second),
	)
	if opts.GetPrefix() != "p" {
		t.Errorf("Prefix=%q", opts.GetPrefix())
	}
	if opts.GetVisibilityTimeout() != 2*time.Second {
		t.Errorf("VisibilityTimeout=%v", opts.GetVisibilityTimeout())
	}
	if opts.GetRetryTimes() != 7 {
		t.Errorf("RetryTimes=%d", opts.GetRetryTimes())
	}
	if opts.GetHandleDeadLetter() != nil {
		t.Errorf("HandleDeadLetter should be nil")
	}
	if opts.GetPollInterval() != 500*time.Millisecond {
		t.Errorf("PollInterval=%v", opts.GetPollInterval())
	}
	if opts.GetPollBatch() != 32 {
		t.Errorf("PollBatch=%d", opts.GetPollBatch())
	}
	if opts.GetRedisOpTimeout() != 3*time.Second {
		t.Errorf("RedisOpTimeout=%v", opts.GetRedisOpTimeout())
	}
	if opts.GetRetryBackoff() != 2*time.Second {
		t.Errorf("RetryBackoff=%v", opts.GetRetryBackoff())
	}
}

func TestDelayOptions_WatchDog(t *testing.T) {
	old := watchDogDelayOptions
	t.Cleanup(func() { watchDogDelayOptions = old })
	called := false
	InstallDelayOptionsWatchDog(func(cc *DelayOptions) { called = true })
	_ = newDelayOptions()
	if !called {
		t.Errorf("watchdog should be called")
	}
}

// ----- LockerOptions -----

func TestLockerOptions_Defaults(t *testing.T) {
	opts := newLockerOptions()
	if opts.GetKeyValidity() <= 0 {
		t.Errorf("KeyValidity default should be >0, got %v", opts.GetKeyValidity())
	}
	if opts.GetKeyPrefix() == "" {
		t.Errorf("KeyPrefix default should be non-empty")
	}
}

func TestLockerOptions_Apply(t *testing.T) {
	opts := newLockerOptions(
		WithLockerOptionKeyValidity(10*time.Second),
		WithLockerOptionKeyPrefix("custom"),
		WithLockerOptionKeyMajority(3),
		WithLockerOptionNoLoopTracking(true),
		WithLockerOptionFallbackSETPX(true),
	)
	if opts.GetKeyValidity() != 10*time.Second {
		t.Errorf("KeyValidity=%v", opts.GetKeyValidity())
	}
	if opts.GetKeyPrefix() != "custom" {
		t.Errorf("KeyPrefix=%q", opts.GetKeyPrefix())
	}
	if opts.GetKeyMajority() != 3 {
		t.Errorf("KeyMajority=%d", opts.GetKeyMajority())
	}
	if !opts.GetNoLoopTracking() {
		t.Errorf("NoLoopTracking=false")
	}
	if !opts.GetFallbackSETPX() {
		t.Errorf("FallbackSETPX=false")
	}
}

func TestLockerOptions_WatchDog(t *testing.T) {
	old := watchDogLockerOptions
	t.Cleanup(func() { watchDogLockerOptions = old })
	called := false
	InstallLockerOptionsWatchDog(func(cc *LockerOptions) { called = true })
	_ = newLockerOptions()
	if !called {
		t.Errorf("watchdog should be called")
	}
}

// ----- RateLimiterOptions -----

func TestRateLimiterOptions_Defaults(t *testing.T) {
	opts := newRateLimiterOptions()
	_ = opts.GetWindow() // 至少不 panic
}

func TestRateLimiterOptions_Apply(t *testing.T) {
	opts := newRateLimiterOptions(
		WithRateLimiterOptionWindow(time.Minute),
	)
	if opts.GetWindow() != time.Minute {
		t.Errorf("Window=%v", opts.GetWindow())
	}
}

// TestNewRateLimiter_RejectsZeroValues 在 newRateLimiter 入口拒绝零值 Limit/Window，
// 让用户拿到一个清晰的、调用方可识别的 sentinel error；不必跑到 rueidislimiter 内部才报错。
//
// 校验在访问 c.v 之前发生，这里直接传 nil client 来覆盖校验路径而无需起 Redis。
func TestNewRateLimiter_RejectsZeroValues(t *testing.T) {
	// 默认 0/0 全拒
	if _, err := newRateLimiter(nil); !errors.Is(err, ErrRateLimiterInvalidLimit) {
		t.Errorf("expect ErrRateLimiterInvalidLimit on default options, got %v", err)
	}
	// 仅给 Window：Limit 仍为 0
	if _, err := newRateLimiter(nil, WithRateLimiterOptionWindow(time.Second)); !errors.Is(err, ErrRateLimiterInvalidLimit) {
		t.Errorf("expect ErrRateLimiterInvalidLimit when Limit=0, got %v", err)
	}
	// 仅给 Limit：Window 仍为 0
	if _, err := newRateLimiter(nil, WithRateLimiterOptionLimit(10)); !errors.Is(err, ErrRateLimiterInvalidWindow) {
		t.Errorf("expect ErrRateLimiterInvalidWindow when Window=0, got %v", err)
	}
	// Window=1ms 边界值（必须 > 1ms）
	if _, err := newRateLimiter(nil, WithRateLimiterOptionLimit(10), WithRateLimiterOptionWindow(time.Millisecond)); !errors.Is(err, ErrRateLimiterInvalidWindow) {
		t.Errorf("expect ErrRateLimiterInvalidWindow when Window==1ms, got %v", err)
	}
}

func TestRateLimiterOptions_WatchDog(t *testing.T) {
	old := watchDogRateLimiterOptions
	t.Cleanup(func() { watchDogRateLimiterOptions = old })
	called := false
	InstallRateLimiterOptionsWatchDog(func(cc *RateLimiterOptions) { called = true })
	_ = newRateLimiterOptions()
	if !called {
		t.Errorf("watchdog should be called")
	}
}

// ----- ConfOptions（已有 TestNewConf_Defaults，这里补 watchdog/atomic 路径） -----

func TestConfWatchDog(t *testing.T) {
	old := watchDogConf
	t.Cleanup(func() { watchDogConf = old })
	called := false
	InstallConfWatchDog(func(cc *Conf) { called = true })
	_ = NewConf()
	if !called {
		t.Errorf("conf watchdog should be called")
	}
}

func TestConfAtomicSetCallback(t *testing.T) {
	old := onAtomicConfSet
	t.Cleanup(func() { onAtomicConfSet = old })
	InstallCallbackOnAtomicConfSet(func(cc ConfInterface) bool { return true })
	// 此处只验证安装路径不 panic；实际 atomic set 路径需要内部触发，覆盖率上算 line。
}
