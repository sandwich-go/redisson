package redisson

// 单元层 benchmark：覆盖纯逻辑热路径，无需 Redis。
// 集成层 benchmark 见 bench_integration_test.go（带 integration tag）。

import (
	"testing"
)

// BenchmarkSlot 测量 slot 计算（CRC16 + hashtag 解析）的吞吐。
// 参考基线（M1 Pro 本地，仅供观察）：~30ns/op。
func BenchmarkSlot(b *testing.B) {
	keys := []string{"foo", "bar", "{user1}.profile", "abcdefghij" + "kkkkkkkkkk"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = slot(keys[i%len(keys)])
	}
}

// BenchmarkCRC16_Short 短字符串（几字节 key）路径。
func BenchmarkCRC16_Short(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = crc16("foo")
	}
}

// BenchmarkCRC16_Long 较长 key 的吞吐表现。
func BenchmarkCRC16_Long(b *testing.B) {
	key := "a-very-long-redis-key-name-with-some-payload:" +
		"12345678901234567890123456789012345678901234567890"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = crc16(key)
	}
}

// BenchmarkRingBuffer_WriteRead 一写一读交替的稳态吞吐。
func BenchmarkRingBuffer_WriteRead(b *testing.B) {
	rb := newRingBuffer[int](16)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.Write(i)
		_, _ = rb.Read()
	}
}

// BenchmarkRingBuffer_BurstWrite 持续写入触发扩容的最差情况。
func BenchmarkRingBuffer_BurstWrite(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rb := newRingBuffer[int](2)
		for j := 0; j < 1024; j++ {
			rb.Write(j)
		}
	}
}

// BenchmarkCheckSlots_AllSame 同槽多 key 检测。
func BenchmarkCheckSlots_AllSame(b *testing.B) {
	keys := []string{"{tag}a", "{tag}b", "{tag}c", "{tag}d"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = checkSlots(CommandMGet, keys...)
	}
}
