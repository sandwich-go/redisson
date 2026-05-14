package redisson

import (
	"errors"
	"testing"
	"time"
)

// TestNewSliceCmdFromSlice_Success 走非 RedisResult 构造路径。
func TestNewSliceCmdFromSlice_Success(t *testing.T) {
	c := newSliceCmdFromSlice([]any{"a", 1, true}, nil, "k1")
	v := c.Val()
	if len(v) != 3 || v[0] != "a" {
		t.Errorf("Val=%v", v)
	}
	if c.Err() != nil {
		t.Errorf("Err=%v", c.Err())
	}
}

// TestNewSliceCmdFromSlice_Error err 直传,Val 不被 set。
func TestNewSliceCmdFromSlice_Error(t *testing.T) {
	want := errors.New("upstream")
	c := newSliceCmdFromSlice(nil, want)
	if !errors.Is(c.Err(), want) {
		t.Errorf("Err=%v", c.Err())
	}
}

// TestSliceCmd_ScanPassesErr 验证 err 路径短路。
func TestSliceCmd_ScanPassesErr(t *testing.T) {
	c := newSliceCmdFromSlice(nil, errors.New("u"))
	if err := c.Scan(new(string)); err == nil {
		t.Error("err should propagate")
	}
}

// TestStringSliceCmd_BaseAndScanSlice baseCmd setter + ScanSlice 委托 scanSlice。
func TestStringSliceCmd_BaseAndScanSlice(t *testing.T) {
	c := &stringSliceCmd{}
	c.SetVal([]string{"1", "2", "3"})
	var dst []int
	if err := c.ScanSlice(&dst); err != nil {
		t.Fatal(err)
	}
	if len(dst) != 3 || dst[0] != 1 || dst[2] != 3 {
		t.Errorf("ScanSlice=%v", dst)
	}
}

func TestStringSliceCmd_ScanSliceBadElem(t *testing.T) {
	c := &stringSliceCmd{}
	c.SetVal([]string{"1", "abc"})
	var dst []int
	if err := c.ScanSlice(&dst); err == nil {
		t.Error("ScanSlice should err on bad element")
	}
}

func TestIntSliceCmd_BaseAccess(t *testing.T) {
	c := &intSliceCmd{}
	c.SetVal([]int64{1, 2, 3})
	v, err := c.Result()
	if err != nil {
		t.Fatal(err)
	}
	if len(v) != 3 {
		t.Errorf("got %v", v)
	}
}

func TestFloatSliceCmd_BaseAccess(t *testing.T) {
	c := &floatSliceCmd{}
	c.SetVal([]float64{1.1, 2.2})
	if v := c.Val(); len(v) != 2 {
		t.Errorf("got %v", v)
	}
}

func TestBoolSliceCmd_BaseAccess(t *testing.T) {
	c := &boolSliceCmd{}
	c.SetVal([]bool{true, false, true})
	if v := c.Val(); len(v) != 3 || !v[0] {
		t.Errorf("got %v", v)
	}
}

func TestDurationSliceCmd_BaseAccess(t *testing.T) {
	c := &durationSliceCmd{precision: time.Second}
	c.SetVal([]time.Duration{time.Second, 2 * time.Second})
	if v := c.Val(); len(v) != 2 {
		t.Errorf("got %v", v)
	}
}
