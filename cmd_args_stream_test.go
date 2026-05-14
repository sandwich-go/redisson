package redisson

import (
	"errors"
	"testing"

	"github.com/redis/rueidis"
)

// TestNewXMessage_NilFieldValues nil 字段值返回空 Values。
func TestNewXMessage_NilFieldValues(t *testing.T) {
	m := newXMessage(rueidis.XRangeEntry{ID: "1-0", FieldValues: nil})
	if m.ID != "1-0" {
		t.Errorf("ID=%q", m.ID)
	}
	if m.Values != nil {
		t.Errorf("Values should be nil for nil FieldValues, got %v", m.Values)
	}
}

// TestNewXMessage_WithFieldValues 字段映射正确。
func TestNewXMessage_WithFieldValues(t *testing.T) {
	m := newXMessage(rueidis.XRangeEntry{
		ID:          "2-3",
		FieldValues: map[string]string{"k1": "v1", "k2": "v2"},
	})
	if m.ID != "2-3" {
		t.Errorf("ID=%q", m.ID)
	}
	if len(m.Values) != 2 {
		t.Fatalf("Values len=%d", len(m.Values))
	}
	if m.Values["k1"] != "v1" || m.Values["k2"] != "v2" {
		t.Errorf("Values=%v", m.Values)
	}
}

// TestXMessageSliceCmd_BaseAccess baseCmd[[]XMessage] 路径。
func TestXMessageSliceCmd_BaseAccess(t *testing.T) {
	c := &xMessageSliceCmd{}
	c.SetVal([]XMessage{{ID: "1-0", Values: map[string]any{"k": "v"}}})
	v := c.Val()
	if len(v) != 1 || v[0].ID != "1-0" {
		t.Errorf("got %+v", v)
	}
}

// TestXAutoClaimCmd_SetValAndResult 双值 setter。
func TestXAutoClaimCmd_SetValAndResult(t *testing.T) {
	c := &xAutoClaimCmd{}
	msgs := []XMessage{{ID: "1-0"}, {ID: "1-1"}}
	c.SetVal(msgs, "next-cursor")
	gotMsgs, gotStart := c.Val()
	if gotStart != "next-cursor" {
		t.Errorf("start=%q", gotStart)
	}
	if len(gotMsgs) != 2 {
		t.Errorf("msgs len=%d", len(gotMsgs))
	}
	gotMsgs2, gotStart2, err := c.Result()
	if err != nil {
		t.Fatal(err)
	}
	if gotStart2 != "next-cursor" || len(gotMsgs2) != 2 {
		t.Errorf("Result mismatch")
	}
}

func TestXAutoClaimCmd_SetErr(t *testing.T) {
	c := &xAutoClaimCmd{}
	c.SetErr(errors.New("e"))
	if c.Err() == nil {
		t.Error("Err should propagate")
	}
}

// TestXAutoClaimJustIDCmd_SetVal id-only 变体。
func TestXAutoClaimJustIDCmd_SetVal(t *testing.T) {
	c := &xAutoClaimJustIDCmd{}
	c.SetVal([]string{"1-0", "1-1"}, "cursor")
	ids, start := c.Val()
	if start != "cursor" || len(ids) != 2 {
		t.Errorf("got %v %q", ids, start)
	}
	ids2, start2, err := c.Result()
	if err != nil {
		t.Fatal(err)
	}
	if start2 != "cursor" || len(ids2) != 2 {
		t.Errorf("Result mismatch")
	}
}

func TestXAutoClaimJustIDCmd_SetErr(t *testing.T) {
	c := &xAutoClaimJustIDCmd{}
	c.SetErr(errors.New("e"))
	if c.Err() == nil {
		t.Error("Err should propagate")
	}
}
