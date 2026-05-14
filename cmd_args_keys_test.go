package redisson

import (
	"errors"
	"testing"
)

// TestKeyValuesCmd_SetVal 验证 setter/getter 链。
func TestKeyValuesCmd_SetVal(t *testing.T) {
	c := &keyValuesCmd{}
	c.SetVal("mykey", []string{"a", "b", "c"})
	k, vs := c.Val()
	if k != "mykey" {
		t.Errorf("key=%q", k)
	}
	if len(vs) != 3 || vs[0] != "a" {
		t.Errorf("vals=%v", vs)
	}
	k2, vs2, err := c.Result()
	if err != nil {
		t.Fatal(err)
	}
	if k2 != "mykey" || len(vs2) != 3 {
		t.Errorf("Result mismatch")
	}
}

func TestKeyValuesCmd_SetErr(t *testing.T) {
	c := &keyValuesCmd{}
	want := errors.New("e")
	c.SetErr(want)
	if !errors.Is(c.Err(), want) {
		t.Errorf("Err=%v", c.Err())
	}
	_, _, err := c.Result()
	if !errors.Is(err, want) {
		t.Errorf("Result err=%v", err)
	}
}

// TestScanCmd_SetVal 验证 ScanCmd setter/getter。
func TestScanCmd_SetVal(t *testing.T) {
	c := &scanCmd{}
	c.SetVal([]string{"k1", "k2"}, 100)
	keys, cursor := c.Val()
	if len(keys) != 2 || cursor != 100 {
		t.Errorf("Val=%v %d", keys, cursor)
	}
	keys2, cursor2, err := c.Result()
	if err != nil {
		t.Fatal(err)
	}
	if len(keys2) != 2 || cursor2 != 100 {
		t.Errorf("Result=%v %d", keys2, cursor2)
	}
}

// TestScanCmd_ZeroValues 默认零值。
func TestScanCmd_ZeroValues(t *testing.T) {
	var c scanCmd
	keys, cursor := c.Val()
	if keys != nil || cursor != 0 {
		t.Errorf("zero Val=%v %d", keys, cursor)
	}
	if c.Err() != nil {
		t.Errorf("zero Err=%v", c.Err())
	}
}

// TestKeyFlagsCmd_BaseAccess 验证 baseCmd[[]KeyFlags] 路径。
func TestKeyFlagsCmd_BaseAccess(t *testing.T) {
	c := &keyFlagsCmd{}
	c.SetVal([]KeyFlags{{Key: "k", Flags: []string{"a", "b"}}})
	v := c.Val()
	if len(v) != 1 || v[0].Key != "k" {
		t.Errorf("Val=%+v", v)
	}
	c.SetErr(errors.New("e"))
	if c.Err() == nil {
		t.Error("Err should propagate")
	}
}
