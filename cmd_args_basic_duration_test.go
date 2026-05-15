package redisson

import (
	"testing"
	"time"
)

// TestDurationCmdNegativeValuePreservesPrecision 验证 durationCmd.from 在 val < 0
// (PTTL/TTL 用 -2 表示 key 不存在、-1 表示无 TTL) 时仍乘以 precision,
// 与非 Cache 路径 (adapter 返回 -1*precision) 保持一致。
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
		got := time.Duration(c.val) * c.precision
		if got != c.want {
			t.Fatalf("val=%d precision=%v: got %v, want %v", c.val, c.precision, got, c.want)
		}
	}
}
