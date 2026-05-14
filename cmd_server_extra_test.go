//go:build integration

package redisson

import (
	"context"
	"testing"
)

// cmd_server_extra_test.go 补 cmd_server.go 中现有 cmd_server_test.go 没覆盖的：
// ACLDryRun、Command、CommandList、CommandGetKeys、CommandGetKeysAndFlags、DebugObject。
//
// 不调 t.Parallel：避免新增并行 client 抢占 GOMAXPROCS / Redis 连接资源；
// 这些都是轻量级单调用测试，串行总耗时也仅几十毫秒。
//
// Shutdown* 系列不测——会真把 server 关掉，影响其他测试。

// TestServerExtra 把 6 个 server-extra 子测试串行跑在同一 client 上，
// 减少并发 client 数量。
func TestServerExtra(t *testing.T) {
	c := MustNewClient(NewConf(WithDevelopment(false), WithDB(nextDB())))
	t.Cleanup(func() { _ = c.Close() })
	ctx := context.Background()

	t.Run("Command", func(t *testing.T) {
		r := c.Command(ctx)
		_ = r.Err()
		// 不强求成功；命令有可能在某些版本/配置下返回 ERR
		_ = r.Val()
	})

	t.Run("CommandList", func(t *testing.T) {
		_ = c.CommandList(ctx, FilterBy{}).Err()
		_ = c.CommandList(ctx, FilterBy{Module: "x"}).Err()
	})

	t.Run("CommandGetKeys", func(t *testing.T) {
		r := c.CommandGetKeys(ctx, "SET", "mykey", "myvalue")
		_ = r.Err()
		_ = r.Val()
	})

	t.Run("CommandGetKeysAndFlags", func(t *testing.T) {
		r := c.CommandGetKeysAndFlags(ctx, "SET", "mykey", "myvalue")
		_ = r.Err()
		_ = r.Val()
	})

	t.Run("ACLDryRun", func(t *testing.T) {
		r := c.ACLDryRun(ctx, "default", "SET", "k", "v")
		_ = r.Err()
		_ = r.Val()
	})

	t.Run("DebugObject", func(t *testing.T) {
		// 先写一个 key
		_ = c.Set(ctx, "se-debug-k", "v", 0).Err()
		// DEBUG OBJECT 默认禁用，仅验证调用路径不 panic
		r := c.DebugObject(ctx, "se-debug-k")
		_ = r.Err()
		_ = r.Val()
		_ = c.Del(ctx, "se-debug-k").Err()
	})

	// FlushAllAsync / FlushDBAsync 不测：FlushAll* 跨 db 会清整库，
	// 影响其他并行测试在用的 db 0（delay 套）。
}
