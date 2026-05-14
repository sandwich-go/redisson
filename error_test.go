package redisson

import (
	"errors"
	"fmt"
	"testing"
)

func TestIsNoScriptError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"plain string match", errors.New("NOSCRIPT No matching script. Please use EVAL"), true},
		{"sentinel via errors.Is", fmt.Errorf("wrap: %w", ErrNoScript), true},
		{"unrelated error", errors.New("some other error"), false},
		{"NOSCRIPT in middle - 不应匹配（必须前缀）", errors.New("server says NOSCRIPT here"), false},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			if got := IsNoScriptError(c.err); got != c.want {
				t.Fatalf("IsNoScriptError(%v) = %v, want %v", c.err, got, c.want)
			}
		})
	}
}

// TestErrors_PushAndErr 验证 Errors 收集器的基本行为。
func TestErrors_PushAndErr(t *testing.T) {
	var es Errors
	if es.Err() != nil {
		t.Fatal("empty Errors should yield nil Err")
	}
	es.Push(nil) // nil 应被忽略
	if es.Err() != nil {
		t.Fatal("Push(nil) should be no-op")
	}
	es.Push(errors.New("a"))
	es.Push(errors.New("b"))
	if es.Err() == nil {
		t.Fatal("non-empty Errors should yield non-nil Err")
	}
	if es.LastErr() == nil || es.LastErr().Error() != "b" {
		t.Fatalf("LastErr() = %v, want b", es.LastErr())
	}
	if len(es.WrappedErrors()) != 2 {
		t.Fatalf("WrappedErrors len = %d, want 2", len(es.WrappedErrors()))
	}
}

// TestParameterError 验证参数错误类型与 IsParameterError 识别。
func TestParameterError(t *testing.T) {
	pe := NewParameterError("invalid unit %s", "FOO")
	if pe == nil {
		t.Fatal("NewParameterError returned nil")
	}
	if pe.Error() != "redisson: invalid unit FOO" {
		t.Fatalf("Error() = %q", pe.Error())
	}
	if !IsParameterError(pe) {
		t.Fatal("IsParameterError(pe) should be true")
	}
	if IsParameterError(errors.New("plain")) {
		t.Fatal("plain error should not be ParameterError")
	}
}

// TestParameterError_Recover 验证 panic(NewParameterError(...)) 可被 recover 识别。
func TestParameterError_Recover(t *testing.T) {
	defer func() {
		r := recover()
		if !IsParameterError(r) {
			t.Fatalf("recover got %T (%v), want *ParameterError", r, r)
		}
	}()
	panic(NewParameterError("test"))
}

// TestErrors_FormatFunc 验证可替换的 format 函数。
func TestErrors_FormatFunc(t *testing.T) {
	var es Errors
	es.Push(errors.New("e1"))
	es.Push(errors.New("e2"))

	// 默认 ListFormatFunc
	if got := es.Error(); got == "" {
		t.Fatal("default Error() should not be empty")
	}

	// 切换到 DotFormatFunc
	es.SetFormatFunc(DotFormatFunc)
	if got := es.Error(); got != "e1,e2" {
		t.Fatalf("DotFormatFunc Error() = %q, want %q", got, "e1,e2")
	}
}

// TestErrors_NilSafe 在 nil 接收者上调用 Err / LastErr 不应 panic。
func TestErrors_NilSafe(t *testing.T) {
	var es *Errors
	if es.LastErr() != nil {
		t.Errorf("nil Errors.LastErr() should be nil")
	}
	if es.Err() != nil {
		t.Errorf("nil Errors.Err() should be nil")
	}
}

// TestErrors_String 验证 String 输出非空。
func TestErrors_String(t *testing.T) {
	var es Errors
	es.Push(errors.New("e"))
	s := es.String()
	if s == "" {
		t.Errorf("String() should be non-empty")
	}
}

// TestErrors_ListFormatFunc 显式调用 ListFormatFunc 的输出格式。
func TestErrors_ListFormatFunc(t *testing.T) {
	got := ListFormatFunc([]error{errors.New("a"), errors.New("b")})
	want := "2 errors occurred:\n#1: a\n#2: b"
	if got != want {
		t.Errorf("ListFormatFunc=%q, want %q", got, want)
	}
}

// TestErrors_DotFormatFunc_Empty 空集 DotFormat 应返回空字符串。
func TestErrors_DotFormatFunc_Empty(t *testing.T) {
	if got := DotFormatFunc(nil); got != "" {
		t.Errorf("DotFormatFunc(nil)=%q, want empty", got)
	}
}

// TestParameterError_FormatArgs 验证格式化参数路径。
func TestParameterError_FormatArgs(t *testing.T) {
	pe := NewParameterError("got %d items, want %d", 3, 5)
	if pe.Error() != "redisson: got 3 items, want 5" {
		t.Errorf("formatted Error()=%q", pe.Error())
	}
}

// TestIsParameterError_Nil nil recover 值不是 ParameterError。
func TestIsParameterError_Nil(t *testing.T) {
	if IsParameterError(nil) {
		t.Errorf("IsParameterError(nil) should be false")
	}
}

// TestIsNoScriptError_LegacyAlias 包内别名 isNoScriptError 与导出版本等价。
func TestIsNoScriptError_LegacyAlias(t *testing.T) {
	if !isNoScriptError(ErrNoScript) {
		t.Errorf("isNoScriptError sentinel mismatch")
	}
	if isNoScriptError(nil) {
		t.Errorf("isNoScriptError(nil) should be false")
	}
}
