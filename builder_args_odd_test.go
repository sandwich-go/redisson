package redisson

import (
	"strings"
	"testing"
)

// mustOddArgsPanic 工具：断言 fn() panic 出 *ParameterError 且消息含 even 关键词。
func mustOddArgsPanic(t *testing.T, name string, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("%s: expected panic on odd args, got none", name)
		}
		if !IsParameterError(r) {
			t.Fatalf("%s: expected *ParameterError, got %T(%v)", name, r, r)
		}
		err := r.(error)
		if !strings.Contains(strings.ToLower(err.Error()), "even") {
			t.Fatalf("%s: error message should mention 'even', got: %s", name, err.Error())
		}
	}()
	fn()
}

// TestMSetCompletedOddArgsPanics 回归 Bug #2：
// MSet/MSetNX 在奇数参数时应当 panic ParameterError，
// 而不是从 partial.KeyValue(args[i+1]) 抛出 runtime 数组越界（信息含糊）。
func TestMSetCompletedOddArgsPanics(t *testing.T) {
	b := builder{}
	mustOddArgsPanic(t, "MSet", func() { b.MSetCompleted("k1", "v1", "k2") })
	mustOddArgsPanic(t, "MSetNX", func() { b.MSetNXCompleted("k1", "v1", "k2") })
}

// TestHMSetCompletedOddArgsPanics 同 #2：HMSet/HMSetX。
func TestHMSetCompletedOddArgsPanics(t *testing.T) {
	b := builder{}
	mustOddArgsPanic(t, "HMSet", func() { b.HMSetCompleted("h", "f1", "v1", "f2") })
	mustOddArgsPanic(t, "HMSetX", func() { b.HMSetXCompleted("h", "f1", "v1", "f2") })
}
