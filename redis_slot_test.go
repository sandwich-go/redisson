package redisson

import (
	"testing"
)

// TestCRC16 使用 XMODEM CRC-16 官方测试向量验证 crc16 实现。
// 详见 redis_slot.go 文件头注释（XMODEM/ZMODEM/CRC-16/ACORN）。
func TestCRC16(t *testing.T) {
	cases := []struct {
		in   string
		want uint16
	}{
		// XMODEM 标准测试向量
		{"123456789", 0x31C3},
		// 空字符串
		{"", 0x0000},
		// 单字符
		{"A", 0x58E5},
	}
	for _, c := range cases {
		c := c
		t.Run(c.in, func(t *testing.T) {
			got := crc16(c.in)
			if got != c.want {
				t.Fatalf("crc16(%q) = 0x%04X, want 0x%04X", c.in, got, c.want)
			}
		})
	}
}

// TestSlotRange 验证 slot 落在 0~16383。
func TestSlotRange(t *testing.T) {
	keys := []string{"", "foo", "bar", "{tag}foo", "{tag}bar", "key:1", "key:2", "user:42"}
	for _, k := range keys {
		s := slot(k)
		if s > 16383 {
			t.Fatalf("slot(%q) = %d out of range [0, 16383]", k, s)
		}
	}
}

// TestSlotHashTag 验证大括号 hashtag 语义：
// 含完整 {tag} 时，slot 仅由 tag 决定，不同前后缀但 tag 相同应得到相同 slot。
func TestSlotHashTag(t *testing.T) {
	cases := []struct {
		a, b      string
		shouldEq  bool
		shouldNeq bool
	}{
		// tag 相同 → 同 slot
		{"{user1}.profile", "{user1}.followers", true, false},
		{"foo{1234}bar", "baz{1234}qux", true, false},
		// 不同 tag → 不同 slot（高概率）
		{"{user1}foo", "{user2}foo", false, true},
		// 空 tag {} 退化为对整个 key（含 "{}"）计算 → 与裸 "foo" 不同
		{"{}foo", "foo", false, true},
		// 但两个相同的 "{}foo" 应得到相同 slot（决定性）
		{"{}foo", "{}foo", true, false},
		// 缺少右大括号也退化
		{"{foo", "{foo", true, false},
	}
	for _, c := range cases {
		c := c
		t.Run(c.a+"|"+c.b, func(t *testing.T) {
			sa := slot(c.a)
			sb := slot(c.b)
			if c.shouldEq && sa != sb {
				t.Fatalf("slot(%q)=%d != slot(%q)=%d, want equal", c.a, sa, c.b, sb)
			}
			if c.shouldNeq && sa == sb {
				t.Fatalf("slot(%q)=%d == slot(%q)=%d, want different", c.a, sa, c.b, sb)
			}
		})
	}
}

// TestSlotKnownValues 一组已知 slot 回归保护值。
// 这些值由本实现首次确定后冻结：未来如果 crc16 实现被改坏，这里会立刻失败。
// 其中 "foo"=12182、"bar"=5061 与 redis-cli CLUSTER KEYSLOT 实测一致（XMODEM CRC-16）。
func TestSlotKnownValues(t *testing.T) {
	cases := map[string]uint16{
		"foo":     12182,
		"bar":     5061,
		"hello":   866,
		"world":   9059,
		"key:1":   6657,
		"counter": 6680,
	}
	for k, want := range cases {
		k, want := k, want
		t.Run(k, func(t *testing.T) {
			got := slot(k)
			if got != want {
				t.Fatalf("slot(%q) = %d, want %d", k, got, want)
			}
		})
	}
}

// TestCheckSlots 验证多 key 跨槽检测逻辑。
func TestCheckSlots(t *testing.T) {
	cmd := CommandMGet // 任意一个命令都行
	// 单 key 不报错
	if err := checkSlots(cmd, "foo"); err != nil {
		t.Fatalf("single key should not error: %v", err)
	}
	// 同 slot（hashtag 相同）不报错
	if err := checkSlots(cmd, "{tag}a", "{tag}b"); err != nil {
		t.Fatalf("same slot should not error: %v", err)
	}
	// 不同 slot 应报错
	if err := checkSlots(cmd, "{user1}a", "{user2}b"); err == nil {
		t.Fatal("different slots should error, got nil")
	}
}

// TestCheckMultipleKeySlotsNilFunc 验证 nil getter 函数被安全跳过。
func TestCheckMultipleKeySlotsNilFunc(t *testing.T) {
	if err := checkMultipleKeySlots(CommandMGet, nil); err != nil {
		t.Fatalf("nil getter should not error: %v", err)
	}
}

// TestExportedSlot 验证导出的 Slot 函数与内部 slot 一致（API 稳定性保护）。
func TestExportedSlot(t *testing.T) {
	keys := []string{"foo", "bar", "{tag}xxx", ""}
	for _, k := range keys {
		if Slot(k) != slot(k) {
			t.Fatalf("Slot(%q) != slot(%q)", k, k)
		}
	}
}
