package redisson

import (
	"sync"
	"testing"
)

// recordingLogger 抓取 Warnf/Errorf 调用，用于测试 SetLogger 注入路径。
type recordingLogger struct {
	mu       sync.Mutex
	warnings []string
	errors   []string
}

func (r *recordingLogger) Warnf(format string, args ...any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.warnings = append(r.warnings, format)
}

func (r *recordingLogger) Errorf(format string, args ...any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.errors = append(r.errors, format)
}

func (r *recordingLogger) snapshot() (w, e []string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.warnings...), append([]string(nil), r.errors...)
}

func TestLogger_DefaultIsStderr(t *testing.T) {
	// 默认实现应返回非 nil 且实现 Logger 接口
	l := GetLogger()
	if l == nil {
		t.Fatal("GetLogger() default should not be nil")
	}
}

func TestSetLogger_NilFallsBackToNop(t *testing.T) {
	original := GetLogger()
	t.Cleanup(func() { SetLogger(original) })

	SetLogger(nil)
	l := GetLogger()
	if _, ok := l.(NopLogger); !ok {
		t.Fatalf("SetLogger(nil) should fall back to NopLogger, got %T", l)
	}
}

func TestSetLogger_RecordingInjection(t *testing.T) {
	original := GetLogger()
	t.Cleanup(func() { SetLogger(original) })

	rec := &recordingLogger{}
	SetLogger(rec)

	// 通过包内桥接函数调用，验证 warning/e 走到注入的 logger
	warning("hello-warn")
	e("hello-err")

	w, errs := rec.snapshot()
	if len(w) != 1 {
		t.Fatalf("expected 1 warning, got %d (%v)", len(w), w)
	}
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d (%v)", len(errs), errs)
	}
}

func TestNopLogger_DiscardsAll(t *testing.T) {
	var n NopLogger
	// 不 panic 即通过
	n.Warnf("anything %d", 1)
	n.Errorf("anything %d", 2)
}
