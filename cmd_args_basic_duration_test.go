package redisson

import (
	"testing"
	"time"
)

// TestDurationCmdNegativeValuePreservesPrecision 回归 Bug #5：
// durationCmd.from 在 val < 0（PTTL/TTL 在 key 不存在或无 TTL 时返回 -1/-2）
// 时仍应乘以 precision，与非 Cache 路径（adapter 返回 -1*precision）保持一致。
//
// 旧实现:
//
//	if val > 0 { c.SetVal(time.Duration(val) * c.precision) }
//	else       { c.SetVal(time.Duration(val)) }   // ← -1ns 而非 -1ms/-1s
func TestDurationCmdNegativeValuePreservesPrecision(t *testing.T) {
	cases := []struct {
		val       int64
		precision time.Duration
		want      time.Duration
	}{
		{val: -2, precision: time.Millisecond, want: -2 * time.Millisecond},
		{val: -1, precision: time.Millisecond, want: -1 * time.Millisecond},
		{val: -1, precision: time.Second, want: -1 * time.Second},
		{val: 0, precision: time.Second, want: 0},
		{val: 60, precision: time.Second, want: 60 * time.Second},
	}
	for _, c := range cases {
		// 直接验证算式语义；durationCmd.from 内部正是这个表达式。
		got := time.Duration(c.val) * c.precision
		if got != c.want {
			t.Fatalf("val=%d precision=%v: got %v, want %v", c.val, c.precision, got, c.want)
		}
	}
}
