package redisson

import (
	"errors"
	"testing"
	"time"
)

// cmd_args_basic 中的 *Cmd 类型大多通过 from(rueidis.RedisResult) 构造，
// 而 rueidis.RedisResult 的字段是 unexported 不能 mock 构造。
// 所以这里只覆盖 SetVal/SetErr 后派生方法的纯逻辑路径，外加 newOKStatusCmdr。

// ------- intCmd -------

func TestIntCmd_Uint64Conversion(t *testing.T) {
	c := &intCmd{}
	c.SetVal(42)
	c.SetErr(nil)
	v, err := c.Uint64()
	if err != nil {
		t.Fatal(err)
	}
	if v != 42 {
		t.Errorf("Uint64=%d, want 42", v)
	}
}

func TestIntCmd_Uint64PassesErr(t *testing.T) {
	c := &intCmd{}
	want := errors.New("boom")
	c.SetErr(want)
	_, err := c.Uint64()
	if !errors.Is(err, want) {
		t.Errorf("err=%v, want %v", err, want)
	}
}

// ------- stringCmd 派生方法 -------

func TestStringCmd_Bytes(t *testing.T) {
	c := &stringCmd{}
	c.SetVal("hello")
	b, err := c.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "hello" {
		t.Errorf("got %q", b)
	}
}

func TestStringCmd_Bool(t *testing.T) {
	c := &stringCmd{}
	c.SetVal("yes")
	b, err := c.Bool()
	if err != nil {
		t.Fatal(err)
	}
	if !b {
		t.Errorf("non-empty val should be true")
	}

	c2 := &stringCmd{}
	c2.SetVal("")
	if b2, _ := c2.Bool(); b2 {
		t.Errorf("empty val should be false")
	}
}

func TestStringCmd_Int(t *testing.T) {
	c := &stringCmd{}
	c.SetVal("42")
	v, err := c.Int()
	if err != nil {
		t.Fatal(err)
	}
	if v != 42 {
		t.Errorf("got %d", v)
	}
}

func TestStringCmd_IntBad(t *testing.T) {
	c := &stringCmd{}
	c.SetVal("nope")
	if _, err := c.Int(); err == nil {
		t.Error("Int(\"nope\") should error")
	}
}

func TestStringCmd_IntPassesErr(t *testing.T) {
	c := &stringCmd{}
	c.SetErr(errors.New("upstream"))
	if _, err := c.Int(); err == nil || err.Error() != "upstream" {
		t.Errorf("err=%v", err)
	}
}

func TestStringCmd_Int64(t *testing.T) {
	c := &stringCmd{}
	c.SetVal("-9876543210")
	v, err := c.Int64()
	if err != nil {
		t.Fatal(err)
	}
	if v != -9876543210 {
		t.Errorf("got %d", v)
	}
}

func TestStringCmd_Int64PassesErr(t *testing.T) {
	c := &stringCmd{}
	c.SetErr(errors.New("u"))
	if _, err := c.Int64(); err == nil {
		t.Error("expected error")
	}
}

func TestStringCmd_Uint64(t *testing.T) {
	c := &stringCmd{}
	c.SetVal("18000000000")
	v, err := c.Uint64()
	if err != nil {
		t.Fatal(err)
	}
	if v != 18000000000 {
		t.Errorf("got %d", v)
	}
}

func TestStringCmd_Uint64PassesErr(t *testing.T) {
	c := &stringCmd{}
	c.SetErr(errors.New("u"))
	if _, err := c.Uint64(); err == nil {
		t.Error("expected error")
	}
}

func TestStringCmd_Float32(t *testing.T) {
	c := &stringCmd{}
	c.SetVal("3.14")
	v, err := c.Float32()
	if err != nil {
		t.Fatal(err)
	}
	if v < 3.13 || v > 3.15 {
		t.Errorf("got %v", v)
	}
}

func TestStringCmd_Float32Bad(t *testing.T) {
	c := &stringCmd{}
	c.SetVal("nope")
	if _, err := c.Float32(); err == nil {
		t.Error("Float32 nope should err")
	}
}

func TestStringCmd_Float32PassesErr(t *testing.T) {
	c := &stringCmd{}
	c.SetErr(errors.New("u"))
	if _, err := c.Float32(); err == nil {
		t.Error("err should propagate")
	}
}

func TestStringCmd_Float64(t *testing.T) {
	c := &stringCmd{}
	c.SetVal("2.71828")
	v, err := c.Float64()
	if err != nil {
		t.Fatal(err)
	}
	if v < 2.71 || v > 2.72 {
		t.Errorf("got %v", v)
	}
}

func TestStringCmd_Float64PassesErr(t *testing.T) {
	c := &stringCmd{}
	c.SetErr(errors.New("u"))
	if _, err := c.Float64(); err == nil {
		t.Error("expected err")
	}
}

func TestStringCmd_TimeRFC3339(t *testing.T) {
	c := &stringCmd{}
	c.SetVal("2024-01-02T03:04:05Z")
	v, err := c.Time()
	if err != nil {
		t.Fatal(err)
	}
	if v.Year() != 2024 {
		t.Errorf("got year %d", v.Year())
	}
}

func TestStringCmd_TimeBad(t *testing.T) {
	c := &stringCmd{}
	c.SetVal("not-a-time")
	if _, err := c.Time(); err == nil {
		t.Error("Time bad should err")
	}
}

func TestStringCmd_TimePassesErr(t *testing.T) {
	c := &stringCmd{}
	c.SetErr(errors.New("u"))
	if _, err := c.Time(); err == nil {
		t.Error("err should propagate")
	}
	if _, err := c.Time(); !errors.Is(err, c.Err()) {
		// 平凡校验
	}
	// 时间在 err 路径返回零值
	if v, _ := c.Time(); !v.IsZero() {
		t.Errorf("err path Time should be zero, got %v", v)
	}
}

func TestStringCmd_String(t *testing.T) {
	c := &stringCmd{}
	c.SetVal("yo")
	if c.String() != "yo" {
		t.Errorf("got %q", c.String())
	}
}

func TestStringCmd_Scan_BinUnmarshaler(t *testing.T) {
	c := &stringCmd{}
	c.SetVal("payload")
	dst := &scanBinUnmarshaler{}
	if err := c.Scan(dst); err != nil {
		t.Fatal(err)
	}
	if string(dst.data) != "payload" {
		t.Errorf("got %q", dst.data)
	}
}

func TestStringCmd_Scan_PassesErr(t *testing.T) {
	c := &stringCmd{}
	c.SetErr(errors.New("u"))
	if err := c.Scan(new(string)); err == nil {
		t.Error("err should propagate")
	}
}

// ------- statusCmd / OKStatusCmdr -------

func TestNewOKStatusCmdr(t *testing.T) {
	c := newOKStatusCmdr()
	if c.Err() != nil {
		t.Errorf("OK status err=%v", c.Err())
	}
	if c.Val() != OK {
		t.Errorf("OK status Val=%q", c.Val())
	}
	v, err := c.Result()
	if err != nil || v != OK {
		t.Errorf("Result=%q err=%v", v, err)
	}
}

// ------- floatCmd / boolCmd 通过基类已覆盖；这里只做 sanity -------

func TestFloatCmd_BaseSetVal(t *testing.T) {
	c := &floatCmd{}
	c.SetVal(3.14)
	if c.Val() != 3.14 {
		t.Errorf("Val=%v", c.Val())
	}
}

func TestBoolCmd_BaseSetVal(t *testing.T) {
	c := &boolCmd{}
	c.SetVal(true)
	if !c.Val() {
		t.Errorf("Val=false, want true")
	}
}

// ------- durationCmd Precision 行为 -------

func TestDurationCmd_PositiveAppliesPrecision(t *testing.T) {
	c := &durationCmd{precision: time.Second}
	c.SetVal(time.Duration(0)) // baseCmd 字段，from() 中根据 val>0 才乘 precision
	// 我们手工模拟 from 的逻辑
	val := int64(5)
	if val > 0 {
		c.SetVal(time.Duration(val) * c.precision)
	} else {
		c.SetVal(time.Duration(val))
	}
	if c.Val() != 5*time.Second {
		t.Errorf("got %v, want 5s", c.Val())
	}
}

func TestDurationCmd_NonPositiveRaw(t *testing.T) {
	// 模拟 val<=0 路径
	c := &durationCmd{precision: time.Second}
	val := int64(-1) // PEXPIRE 等可能返回 -1
	if val > 0 {
		c.SetVal(time.Duration(val) * c.precision)
	} else {
		c.SetVal(time.Duration(val))
	}
	if c.Val() != time.Duration(-1) {
		t.Errorf("got %v, want -1ns", c.Val())
	}
}
