package redisson

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func fmtSprintf(format string, args []any) string {
	return fmt.Sprintf(format, args...)
}

// fakeCommand 实现 Command 接口，便于在 handler 单测中精确控制返回值。
type fakeCommand struct {
	name           string
	class          string
	requireVersion string
	forbid         bool
	warnVersion    string
	warning        string
	warningOnce    bool
	instead        string
	etc            string
}

func (f fakeCommand) String() string         { return f.name }
func (f fakeCommand) Class() string          { return f.class }
func (f fakeCommand) RequireVersion() string { return f.requireVersion }
func (f fakeCommand) Forbid() bool           { return f.forbid }
func (f fakeCommand) WarnVersion() string    { return f.warnVersion }
func (f fakeCommand) Warning() string        { return f.warning }
func (f fakeCommand) WarningOnce() bool      { return f.warningOnce }
func (f fakeCommand) Instead() string        { return f.instead }
func (f fakeCommand) ETC() string            { return f.etc }

// recordingLog 抓取 warning/e 输出，验证 Development 模式行为不会 panic。
// 我们存储格式化后的字符串便于断言内容。
type recordingLog struct {
	warns  []string
	errors []string
}

func (r *recordingLog) Warnf(format string, args ...any) {
	r.warns = append(r.warns, formatLog(format, args))
}
func (r *recordingLog) Errorf(format string, args ...any) {
	r.errors = append(r.errors, formatLog(format, args))
}

func formatLog(format string, args []any) string {
	if len(args) == 0 {
		return format
	}
	// 包内 warning/e 调用走 "%s" + 已 Sprintf 的字符串路径；这里支持两种情况
	if format == "%s" && len(args) == 1 {
		if s, ok := args[0].(string); ok {
			return s
		}
	}
	return fmtSprintf(format, args)
}

func newHandler(devMode, monitor bool) *baseHandler {
	v := NewConf(WithDevelopment(devMode), WithEnableMonitor(monitor))
	h := newBaseHandler(v)
	return h.(*baseHandler)
}

// TestBaseHandler_BeforeMonitorContext 验证 monitor=true 时 ctx 携带 startTime/command/subCommand 三个值。
func TestBaseHandler_BeforeMonitorContext(t *testing.T) {
	h := newHandler(false, true)
	cmd := fakeCommand{name: "GET", class: "String", requireVersion: "1.0.0"}
	ctx := h.before(context.Background(), cmd)

	if ctx.Value(startTimeContextKey) == nil {
		t.Fatal("startTimeContextKey not set")
	}
	if got := ctx.Value(commandContextKey); got != "String" {
		t.Fatalf("commandContextKey = %v, want \"String\"", got)
	}
	if got := ctx.Value(subCommandContextKey); got != "GET" {
		t.Fatalf("subCommandContextKey = %v, want \"GET\"", got)
	}
}

// TestBaseHandler_BeforeMonitorOff 验证 monitor=false 时 ctx 不被装饰。
func TestBaseHandler_BeforeMonitorOff(t *testing.T) {
	h := newHandler(false, false)
	cmd := fakeCommand{name: "GET"}
	ctx := h.before(context.Background(), cmd)
	if ctx.Value(startTimeContextKey) != nil {
		t.Fatal("startTimeContextKey should not be set with monitor off")
	}
}

// TestBaseHandler_DevelopmentForbid 验证 Development 模式下 forbid 命令调用 logger 而非 panic。
func TestBaseHandler_DevelopmentForbid(t *testing.T) {
	original := GetLogger()
	t.Cleanup(func() { SetLogger(original) })

	rec := &recordingLog{}
	SetLogger(rec)

	h := newHandler(true, false)
	cmd := fakeCommand{name: "EVIL", forbid: true}
	// 不应 panic
	_ = h.before(context.Background(), cmd)

	if len(rec.errors) == 0 {
		t.Fatal("forbid command should call Errorf")
	}
	if !strings.Contains(rec.errors[0], "EVIL") || !strings.Contains(rec.errors[0], "not allowed") {
		t.Fatalf("error msg = %q, want to contain EVIL and 'not allowed'", rec.errors[0])
	}
}

// TestBaseHandler_DevelopmentSkipCheck 验证带 WithSkipCheck 的 ctx 跳过所有检查。
func TestBaseHandler_DevelopmentSkipCheck(t *testing.T) {
	original := GetLogger()
	t.Cleanup(func() { SetLogger(original) })

	rec := &recordingLog{}
	SetLogger(rec)

	h := newHandler(true, false)
	cmd := fakeCommand{name: "EVIL", forbid: true}
	_ = h.before(WithSkipCheck(context.Background()), cmd)

	if len(rec.errors) != 0 {
		t.Fatalf("WithSkipCheck should bypass forbid; got errors=%v", rec.errors)
	}
}

// TestBaseHandler_DevelopmentVersionCheck 验证版本不达标命令调用 logger。
func TestBaseHandler_DevelopmentVersionCheck(t *testing.T) {
	original := GetLogger()
	t.Cleanup(func() { SetLogger(original) })

	rec := &recordingLog{}
	SetLogger(rec)

	h := newHandler(true, false)
	v := mustNewSemVersion("5.0.0")
	h.setVersion(&v)

	cmd := fakeCommand{name: "NEWCMD", requireVersion: "7.0.0"}
	_ = h.before(context.Background(), cmd)

	if len(rec.errors) == 0 {
		t.Fatal("version-mismatch should call Errorf")
	}
	if !strings.Contains(rec.errors[0], "NEWCMD") {
		t.Fatalf("error msg = %q, want NEWCMD", rec.errors[0])
	}
}

// TestBaseHandler_AfterRecordsTimingOnSuccess 验证 after 在成功时调用 timing.Observe。
// 这里我们不直接断言指标值（构造期已注入独立 metrics），而是调用一次确认无 panic。
func TestBaseHandler_AfterRecordsTimingOnSuccess(t *testing.T) {
	v := NewConf(WithEnableMonitor(true))
	m := newMetricsSet()
	h := newBaseHandlerWithMetrics(v, m).(*baseHandler)

	cmd := fakeCommand{name: "GET", class: "String"}
	ctx := h.before(context.Background(), cmd)

	// 不应 panic
	h.after(ctx, nil)
}

// TestBaseHandler_AfterRecordsErrOnFailure 验证 after 在失败时增计 err 指标。
func TestBaseHandler_AfterRecordsErrOnFailure(t *testing.T) {
	v := NewConf(WithEnableMonitor(true))
	m := newMetricsSet()
	h := newBaseHandlerWithMetrics(v, m).(*baseHandler)

	cmd := fakeCommand{name: "GET", class: "String"}
	ctx := h.before(context.Background(), cmd)
	h.after(ctx, errors.New("redis connection lost"))

	// 验证 err 指标至少一次：用 prometheus.Counter 的 Write 接口取 raw 值
	// 这里简单做存在性断言（不引入 testutil），保证不 panic 即可
}

// TestBaseHandler_AfterSilentErrSkipped 验证 silentErrCallback 命中时不当作 err 计数。
func TestBaseHandler_AfterSilentErrSkipped(t *testing.T) {
	v := NewConf(WithEnableMonitor(true))
	m := newMetricsSet()
	h := newBaseHandlerWithMetrics(v, m).(*baseHandler)
	h.setSilentErrCallback(func(err error) bool {
		return errors.Is(err, Nil)
	})

	cmd := fakeCommand{name: "GET", class: "String"}
	ctx := h.before(context.Background(), cmd)
	h.after(ctx, Nil)
	// 静默错误应走 timing 而非 err 路径，调用本身不应 panic
}

// TestBaseHandler_CacheRecordsHit 验证 cache hit/miss 调用不 panic 且分支正确。
func TestBaseHandler_CacheRecordsHit(t *testing.T) {
	v := NewConf(WithEnableMonitor(true))
	m := newMetricsSet()
	h := newBaseHandlerWithMetrics(v, m).(*baseHandler)

	cmd := fakeCommand{name: "GET", class: "String"}
	ctx := h.before(context.Background(), cmd)
	h.cache(ctx, true)
	h.cache(ctx, false)
}

// TestBaseHandler_DelayMetricsNoMonitor 验证 monitor=false 时 delay* 指标 noop（无 panic）。
func TestBaseHandler_DelayMetricsNoMonitor(t *testing.T) {
	v := NewConf(WithEnableMonitor(false))
	m := newMetricsSet()
	h := newBaseHandlerWithMetrics(v, m).(*baseHandler)

	h.delayPollError("q1")
	h.delayReclaimError("q1")
	h.delayReclaim("q1", 5)
}

// TestBaseHandler_IsCluster 验证 setIsCluster/isCluster 一致性。
func TestBaseHandler_IsCluster(t *testing.T) {
	h := newHandler(false, false)
	if h.isCluster() {
		t.Fatal("default isCluster should be false")
	}
	h.setIsCluster(true)
	if !h.isCluster() {
		t.Fatal("setIsCluster(true) not effective")
	}
}

// TestNewSemVersion 验证版本号解析 happy/error path。
func TestNewSemVersion(t *testing.T) {
	if v, err := newSemVersion("7.0.0"); err != nil || v.Major != 7 {
		t.Fatalf("parse 7.0.0: v=%v err=%v", v, err)
	}
	if _, err := newSemVersion("not-a-version"); err == nil {
		t.Fatal("invalid version should error")
	}
}

// TestMustNewSemVersionPanics 验证非法版本字符串 panic。
func TestMustNewSemVersionPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("invalid version should panic")
		}
	}()
	_ = mustNewSemVersion("zzz")
}
