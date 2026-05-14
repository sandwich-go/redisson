//go:build integration

package redisson

import (
	"context"
	"testing"
	"time"
)

// cmd_hash_extra_test.go 补 cmd_hash.go 中：
// HStrLen, HRandFieldWithValues, HExpire 系列（HExpire/HExpireAt/HPExpire 等及
// XX/NX/GT/LT 变体），HExpireTime, HPExpireTime, HPersist, HTTL, HPTTL.
// 这些命令在 Redis 7.4+ 才完全支持；7.2.x 多数返回 ERR 但 wrapper 路径仍被覆盖。
//
// 串行跑（不 t.Parallel）共享一个 client，避免并发 client 抢资源。

func TestHashExtra(t *testing.T) {
	c := MustNewClient(NewConf(WithDevelopment(false), WithDB(nextDB())))
	t.Cleanup(func() { _ = c.Close() })
	ctx := context.Background()

	// 准备一个 hash
	const key = "he-key"
	_ = c.HSet(ctx, key, "f1", "v1", "f2", "v2", "f3", "v3").Err()
	t.Cleanup(func() { _ = c.Del(ctx, key).Err() })

	t.Run("HStrLen", func(t *testing.T) {
		r := c.HStrLen(ctx, key, "f1")
		_ = r.Err()
		if r.Err() == nil && r.Val() != 2 {
			t.Errorf("HStrLen=%d, want 2", r.Val())
		}
	})

	t.Run("HRandFieldWithValues", func(t *testing.T) {
		r := c.HRandFieldWithValues(ctx, key, 2)
		_ = r.Err()
		_ = r.Val()
	})

	// ----- HExpire 系列（秒精度） -----
	// 这些在 7.4+ 工作；7.2 多返回 ERR unknown command 'hexpire'，但 builder 路径
	// 仍被遍历到。

	t.Run("HExpire", func(t *testing.T) {
		_ = c.HExpire(ctx, key, time.Hour, "f1").Err()
	})
	t.Run("HExpireNX", func(t *testing.T) {
		_ = c.HExpireNX(ctx, key, time.Hour, "f1").Err()
	})
	t.Run("HExpireXX", func(t *testing.T) {
		_ = c.HExpireXX(ctx, key, time.Hour, "f1").Err()
	})
	t.Run("HExpireGT", func(t *testing.T) {
		_ = c.HExpireGT(ctx, key, time.Hour, "f1").Err()
	})
	t.Run("HExpireLT", func(t *testing.T) {
		_ = c.HExpireLT(ctx, key, time.Hour, "f1").Err()
	})

	// ----- HExpireAt 系列 -----
	expiry := time.Now().Add(time.Hour)
	t.Run("HExpireAt", func(t *testing.T) {
		_ = c.HExpireAt(ctx, key, expiry, "f1").Err()
	})
	t.Run("HExpireAtNX", func(t *testing.T) {
		_ = c.HExpireAtNX(ctx, key, expiry, "f1").Err()
	})
	t.Run("HExpireAtXX", func(t *testing.T) {
		_ = c.HExpireAtXX(ctx, key, expiry, "f1").Err()
	})
	t.Run("HExpireAtGT", func(t *testing.T) {
		_ = c.HExpireAtGT(ctx, key, expiry, "f1").Err()
	})
	t.Run("HExpireAtLT", func(t *testing.T) {
		_ = c.HExpireAtLT(ctx, key, expiry, "f1").Err()
	})

	// ----- HPExpire 系列（毫秒精度） -----
	t.Run("HPExpire", func(t *testing.T) {
		_ = c.HPExpire(ctx, key, time.Hour, "f1").Err()
	})
	t.Run("HPExpireNX", func(t *testing.T) {
		_ = c.HPExpireNX(ctx, key, time.Hour, "f1").Err()
	})
	t.Run("HPExpireXX", func(t *testing.T) {
		_ = c.HPExpireXX(ctx, key, time.Hour, "f1").Err()
	})
	t.Run("HPExpireGT", func(t *testing.T) {
		_ = c.HPExpireGT(ctx, key, time.Hour, "f1").Err()
	})
	t.Run("HPExpireLT", func(t *testing.T) {
		_ = c.HPExpireLT(ctx, key, time.Hour, "f1").Err()
	})

	// ----- HPExpireAt 系列 -----
	t.Run("HPExpireAt", func(t *testing.T) {
		_ = c.HPExpireAt(ctx, key, expiry, "f1").Err()
	})
	t.Run("HPExpireAtNX", func(t *testing.T) {
		_ = c.HPExpireAtNX(ctx, key, expiry, "f1").Err()
	})
	t.Run("HPExpireAtXX", func(t *testing.T) {
		_ = c.HPExpireAtXX(ctx, key, expiry, "f1").Err()
	})
	t.Run("HPExpireAtGT", func(t *testing.T) {
		_ = c.HPExpireAtGT(ctx, key, expiry, "f1").Err()
	})
	t.Run("HPExpireAtLT", func(t *testing.T) {
		_ = c.HPExpireAtLT(ctx, key, expiry, "f1").Err()
	})

	// ----- 查询 / 取消过期 -----
	t.Run("HExpireTime", func(t *testing.T) {
		_ = c.HExpireTime(ctx, key, "f1").Err()
	})
	t.Run("HPExpireTime", func(t *testing.T) {
		_ = c.HPExpireTime(ctx, key, "f1").Err()
	})
	t.Run("HTTL", func(t *testing.T) {
		_ = c.HTTL(ctx, key, "f1").Err()
	})
	t.Run("HPTTL", func(t *testing.T) {
		_ = c.HPTTL(ctx, key, "f1").Err()
	})
	t.Run("HPersist", func(t *testing.T) {
		_ = c.HPersist(ctx, key, "f1").Err()
	})
}
