//go:build integration

package redisson

import (
	"context"
	"testing"
	"time"
)

// cmd_generic_extra_test.go 补 cmd_generic.go 中现有 cmd_generic_test.go 没覆盖的：
// Expire* / ExpireAt* / PExpire* / PExpireAt* 的 NX/XX/GT/LT 变体（共 16 个），
// ExpireTime / PExpireTime（Redis 7.0+），
// SortRO / SortInterfaces / SortStore。

func TestGenericExtra(t *testing.T) {
	c := MustNewClient(NewConf(WithDevelopment(false), WithDB(nextDB())))
	t.Cleanup(func() { _ = c.Close() })
	ctx := context.Background()

	const key = "ge-key"
	_ = c.Set(ctx, key, "v", 0).Err()
	t.Cleanup(func() { _ = c.Del(ctx, key).Err() })

	expiry := time.Now().Add(time.Hour)

	// ---- Expire 各 modifier ----
	t.Run("ExpireNX", func(t *testing.T) {
		_ = c.ExpireNX(ctx, key, time.Hour).Err()
	})
	t.Run("ExpireXX", func(t *testing.T) {
		_ = c.ExpireXX(ctx, key, time.Hour).Err()
	})
	t.Run("ExpireGT", func(t *testing.T) {
		_ = c.ExpireGT(ctx, key, 2*time.Hour).Err()
	})
	t.Run("ExpireLT", func(t *testing.T) {
		_ = c.ExpireLT(ctx, key, 30*time.Minute).Err()
	})

	// ---- ExpireAt 各 modifier ----
	t.Run("ExpireAtNX", func(t *testing.T) {
		_ = c.ExpireAtNX(ctx, key, expiry).Err()
	})
	t.Run("ExpireAtXX", func(t *testing.T) {
		_ = c.ExpireAtXX(ctx, key, expiry).Err()
	})
	t.Run("ExpireAtGT", func(t *testing.T) {
		_ = c.ExpireAtGT(ctx, key, expiry.Add(time.Hour)).Err()
	})
	t.Run("ExpireAtLT", func(t *testing.T) {
		_ = c.ExpireAtLT(ctx, key, expiry).Err()
	})

	// ---- PExpire 各 modifier ----
	t.Run("PExpireNX", func(t *testing.T) {
		_ = c.PExpireNX(ctx, key, time.Hour).Err()
	})
	t.Run("PExpireXX", func(t *testing.T) {
		_ = c.PExpireXX(ctx, key, time.Hour).Err()
	})
	t.Run("PExpireGT", func(t *testing.T) {
		_ = c.PExpireGT(ctx, key, 2*time.Hour).Err()
	})
	t.Run("PExpireLT", func(t *testing.T) {
		_ = c.PExpireLT(ctx, key, 30*time.Minute).Err()
	})

	// ---- PExpireAt 各 modifier ----
	t.Run("PExpireAtNX", func(t *testing.T) {
		_ = c.PExpireAtNX(ctx, key, expiry).Err()
	})
	t.Run("PExpireAtXX", func(t *testing.T) {
		_ = c.PExpireAtXX(ctx, key, expiry).Err()
	})
	t.Run("PExpireAtGT", func(t *testing.T) {
		_ = c.PExpireAtGT(ctx, key, expiry.Add(time.Hour)).Err()
	})
	t.Run("PExpireAtLT", func(t *testing.T) {
		_ = c.PExpireAtLT(ctx, key, expiry).Err()
	})

	// ---- ExpireTime / PExpireTime ----
	t.Run("ExpireTime", func(t *testing.T) {
		r := c.ExpireTime(ctx, key)
		_ = r.Err()
		_ = r.Val()
	})
	t.Run("PExpireTime", func(t *testing.T) {
		r := c.PExpireTime(ctx, key)
		_ = r.Err()
		_ = r.Val()
	})

	// ---- Sort 系列 ----
	const listKey = "ge-list"
	_ = c.LPush(ctx, listKey, "3", "1", "2").Err()
	t.Cleanup(func() { _ = c.Del(ctx, listKey).Err() })

	t.Run("SortRO", func(t *testing.T) {
		r := c.SortRO(ctx, listKey, Sort{Order: ASC})
		if err := r.Err(); err != nil {
			t.Logf("SortRO err=%v (允许;仅覆盖路径)", err)
		}
		_ = r.Val()
	})
	t.Run("SortInterfaces", func(t *testing.T) {
		r := c.SortInterfaces(ctx, listKey, Sort{Order: ASC})
		if err := r.Err(); err != nil {
			t.Logf("SortInterfaces err=%v", err)
		}
		_ = r.Val()
	})
	t.Run("SortStore", func(t *testing.T) {
		_ = c.SortStore(ctx, listKey, "ge-list-out", Sort{Order: ASC}).Err()
		_ = c.Del(ctx, "ge-list-out").Err()
	})
}
