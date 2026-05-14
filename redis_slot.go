package redisson

import "github.com/sandwich-go/redisson/internal/redisslot"

// checkMultipleKeySlots 检查多 key 命令的所有 key 是否在同一 slot；
// 若不同则返回错误。Development 模式下用于警告（不再 panic）。
func checkMultipleKeySlots(command Command, f func() []string) error {
	return redisslot.CheckMultipleKeySlots(command, f)
}

// checkSlots 内部使用：验证一组 key 的 slot 一致性。
func checkSlots(command Command, keys ...string) error {
	return redisslot.CheckSlots(command, keys...)
}

// slot 内部使用：计算 key 的 slot（保留私有名以兼容现有调用点）。
func slot(key string) uint16 { return redisslot.Slot(key) }

// crc16 内部使用：XMODEM CRC-16 校验。
func crc16(key string) uint16 { return redisslot.CRC16(key) }

// Deprecated: 保留向后兼容；现已不在生产路径中使用，请改用 checkMultipleKeySlots。
//
//nolint:unused // 公开 API 兼容残留，保留至 1.4 主版本前
func panicIfUseMultipleKeySlots(command Command, f func() []string) {
	if err := checkMultipleKeySlots(command, f); err != nil {
		panic(err)
	}
}

// Slot 公开 API：计算 key 在 Redis Cluster 中的槽位。
func Slot(key string) uint16 { return redisslot.Slot(key) }
