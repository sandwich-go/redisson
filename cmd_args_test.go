package redisson

import (
	"errors"
	"testing"
)

// baseCmd 是泛型基类，所有 *Cmd 实现的公共 setter/getter。
// 这里直接实例化测试。

func TestBaseCmd_SetValAndVal(t *testing.T) {
	var c baseCmd[int]
	c.SetVal(42)
	if c.Val() != 42 {
		t.Errorf("Val=%d, want 42", c.Val())
	}
}

func TestBaseCmd_SetErrAndErr(t *testing.T) {
	var c baseCmd[string]
	want := errors.New("boom")
	c.SetErr(want)
	if !errors.Is(c.Err(), want) {
		t.Errorf("Err=%v, want %v", c.Err(), want)
	}
}

// TestBaseCmd_Result 同时返回 val + err。
func TestBaseCmd_Result(t *testing.T) {
	var c baseCmd[bool]
	c.SetVal(true)
	c.SetErr(nil)
	v, err := c.Result()
	if err != nil {
		t.Fatal(err)
	}
	if !v {
		t.Errorf("Result val=%v, want true", v)
	}
}

// TestBaseCmd_TypedSlice 验证泛型用 slice 类型也能工作。
func TestBaseCmd_TypedSlice(t *testing.T) {
	var c baseCmd[[]string]
	c.SetVal([]string{"a", "b"})
	got := c.Val()
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("Val=%v", got)
	}
}

// TestBaseCmd_ZeroValueErrIsNil 默认零值 Err 应为 nil。
func TestBaseCmd_ZeroValueErrIsNil(t *testing.T) {
	var c baseCmd[int]
	if c.Err() != nil {
		t.Errorf("zero value Err=%v, want nil", c.Err())
	}
}

// TestCmdArgsConstants 公开常量值校验（防止误改）。
func TestCmdArgsConstants(t *testing.T) {
	cases := map[string]string{
		"OK":     OK,
		"BYTE":   BYTE,
		"BIT":    BIT,
		"M":      M,
		"KM":     KM,
		"FT":     FT,
		"MI":     MI,
		"XX":     XX,
		"NX":     NX,
		"BEFORE": BEFORE,
		"AFTER":  AFTER,
		"RIGHT":  RIGHT,
		"LEFT":   LEFT,
		"LADDR":  LADDR,
		"TYPE":   TYPE,
		"ASC":    ASC,
		"DESC":   DESC,
	}
	for name, val := range cases {
		if val != name {
			t.Errorf("constant %s = %q, want %q", name, val, name)
		}
	}
	if EMPTY != "" {
		t.Errorf("EMPTY=%q, want empty string", EMPTY)
	}
}
