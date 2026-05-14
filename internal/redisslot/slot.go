// Package redisslot 实现 Redis Cluster 的 slot 计算（XMODEM CRC-16 + hashtag 解析）。
//
// 该包是 redisson 的内部实现细节，外部用户应使用 redisson.Slot。
package redisslot

import "fmt"

// SlotMask Redis Cluster 共 16384 个 slot，对应低 14 位掩码。
const SlotMask = 16383

// Slot 计算 key 在 Redis Cluster 中的槽位。
//
// hashtag 语义（与 Redis 服务端一致）:
//   - 若 key 含完整的 {tag}，且 tag 非空，则只对 tag 计算 CRC16;
//   - 否则对整个 key 计算。
//
// 返回 [0, 16383] 之间的整数。
func Slot(key string) uint16 {
	var s, e int
	for ; s < len(key); s++ {
		if key[s] == '{' {
			break
		}
	}
	if s == len(key) {
		return CRC16(key) & SlotMask
	}
	for e = s + 1; e < len(key); e++ {
		if key[e] == '}' {
			break
		}
	}
	if e == len(key) || e == s+1 {
		return CRC16(key) & SlotMask
	}
	return CRC16(key[s+1:e]) & SlotMask
}

// CommandNamer 提供命令名以便在错误消息中标注。
type CommandNamer interface {
	String() string
}

// CheckSlots 验证多 key 是否落在同一 slot。
// 单 key 总是返回 nil；不同 slot 时返回带命令名的错误。
func CheckSlots(cmd CommandNamer, keys ...string) error {
	if len(keys) <= 1 {
		return nil
	}
	var pre uint16
	for k, v := range keys {
		s := Slot(v)
		if k > 0 && pre != s {
			return fmt.Errorf("[%s]: multiple keys command with different key slots are not allowed", cmd.String())
		}
		pre = s
	}
	return nil
}

// CheckMultipleKeySlots 等价于 CheckSlots，但接收 lazy 的 key getter。
// 当 f 为 nil 时直接返回 nil（用于"未知 key 列表"场景）。
func CheckMultipleKeySlots(cmd CommandNamer, f func() []string) error {
	if f == nil {
		return nil
	}
	return CheckSlots(cmd, f()...)
}
