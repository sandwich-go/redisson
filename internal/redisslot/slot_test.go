package redisslot

import "testing"

type cmdName string

func (c cmdName) String() string { return string(c) }

func TestCRC16Standard(t *testing.T) {
	if got := CRC16("123456789"); got != 0x31C3 {
		t.Fatalf("CRC16(123456789) = 0x%04X, want 0x31C3", got)
	}
}

func TestSlotInRange(t *testing.T) {
	for _, k := range []string{"", "foo", "bar", "{tag}foo", "user:42"} {
		if Slot(k) > SlotMask {
			t.Fatalf("Slot(%q) out of range", k)
		}
	}
}

func TestSlotHashTag(t *testing.T) {
	if Slot("{user1}.profile") != Slot("{user1}.followers") {
		t.Fatal("same hashtag should yield same slot")
	}
	if Slot("{user1}foo") == Slot("{user2}foo") {
		t.Fatal("different hashtags should differ (high prob)")
	}
}

func TestCheckSlots(t *testing.T) {
	cmd := cmdName("MGET")
	if err := CheckSlots(cmd, "foo"); err != nil {
		t.Fatalf("single key shouldn't error: %v", err)
	}
	if err := CheckSlots(cmd, "{tag}a", "{tag}b"); err != nil {
		t.Fatalf("same slot shouldn't error: %v", err)
	}
	if err := CheckSlots(cmd, "{a}x", "{b}y"); err == nil {
		t.Fatal("different slots should error")
	}
}

func TestCheckMultipleKeySlotsNil(t *testing.T) {
	if err := CheckMultipleKeySlots(cmdName("X"), nil); err != nil {
		t.Fatalf("nil getter should be ok: %v", err)
	}
}
