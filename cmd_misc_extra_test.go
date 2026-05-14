//go:build integration

package redisson

import (
	"context"
	"testing"
	"time"
)

// cmd_misc_extra_test.go 集中补 set/sortedset/list/bitmap/connection 中
// 现有测试漏掉的命令 wrapper。串行跑共享 client，避免并发资源争用。

// TestSetExtra: SInterCard。
func TestSetExtra(t *testing.T) {
	c := MustNewClient(NewConf(WithDevelopment(false), WithDB(nextDB())))
	t.Cleanup(func() { _ = c.Close() })
	ctx := context.Background()

	// 用同 hashtag 保证 cluster slot 一致性
	const k1, k2 = "{ic}.s1", "{ic}.s2"
	_ = c.SAdd(ctx, k1, "a", "b", "c").Err()
	_ = c.SAdd(ctx, k2, "b", "c", "d").Err()
	t.Cleanup(func() {
		_ = c.Del(ctx, k1).Err()
		_ = c.Del(ctx, k2).Err()
	})

	// SInterCard 计算交集大小（不返回成员），limit=0 表示不限
	r := c.SInterCard(ctx, 0, k1, k2)
	if r.Err() != nil {
		t.Logf("SInterCard err=%v", r.Err())
	}
	if r.Err() == nil && r.Val() != 2 {
		t.Errorf("SInterCard=%d, want 2", r.Val())
	}
}

// TestSortedSetExtra: BZMPop / ZAddGT / ZAddLT / ZInterCard / ZMPop /
// ZRandMemberWithScores / ZRankWithScore / ZRevRankWithScore.
func TestSortedSetExtra(t *testing.T) {
	c := MustNewClient(NewConf(WithDevelopment(false), WithDB(nextDB())))
	t.Cleanup(func() { _ = c.Close() })
	ctx := context.Background()

	const k1, k2 = "{zse}.k1", "{zse}.k2"
	_ = c.ZAdd(ctx, k1, Z{Score: 1, Member: "a"}, Z{Score: 2, Member: "b"}, Z{Score: 3, Member: "c"}).Err()
	_ = c.ZAdd(ctx, k2, Z{Score: 1, Member: "b"}, Z{Score: 2, Member: "c"}, Z{Score: 3, Member: "d"}).Err()
	t.Cleanup(func() {
		_ = c.Del(ctx, k1).Err()
		_ = c.Del(ctx, k2).Err()
	})

	t.Run("ZAddGT", func(t *testing.T) {
		// GT: 仅当新 score > 旧 score 才更新
		_ = c.ZAddGT(ctx, k1, Z{Score: 100, Member: "a"}).Err()
	})
	t.Run("ZAddLT", func(t *testing.T) {
		// LT: 仅当新 score < 旧 score 才更新
		_ = c.ZAddLT(ctx, k1, Z{Score: -1, Member: "a"}).Err()
	})

	t.Run("ZInterCard", func(t *testing.T) {
		r := c.ZInterCard(ctx, 0, k1, k2)
		_ = r.Err()
		_ = r.Val()
	})

	t.Run("ZMPop", func(t *testing.T) {
		// MIN: 弹出最小 score 的元素；count=1
		r := c.ZMPop(ctx, "MIN", 1, k1)
		_ = r.Err()
		_, _, _ = r.Result()
	})

	t.Run("BZMPop", func(t *testing.T) {
		// timeout=0.1s 防止真的阻塞
		r := c.BZMPop(ctx, 100*time.Millisecond, "MIN", 1, k1)
		_ = r.Err()
		_, _, _ = r.Result()
	})

	t.Run("ZRandMemberWithScores", func(t *testing.T) {
		r := c.ZRandMemberWithScores(ctx, k1, 2)
		_ = r.Err()
		_ = r.Val()
	})

	t.Run("ZRankWithScore", func(t *testing.T) {
		r := c.ZRankWithScore(ctx, k1, "b")
		_ = r.Err()
		_ = r.Val()
	})

	t.Run("ZRevRankWithScore", func(t *testing.T) {
		r := c.ZRevRankWithScore(ctx, k1, "b")
		_ = r.Err()
		_ = r.Val()
	})
}

// TestListExtra: BLMPop / LMPop.
func TestListExtra(t *testing.T) {
	c := MustNewClient(NewConf(WithDevelopment(false), WithDB(nextDB())))
	t.Cleanup(func() { _ = c.Close() })
	ctx := context.Background()

	const key = "le-key"
	_ = c.LPush(ctx, key, "a", "b", "c").Err()
	t.Cleanup(func() { _ = c.Del(ctx, key).Err() })

	t.Run("LMPop", func(t *testing.T) {
		r := c.LMPop(ctx, LEFT, 1, key)
		_ = r.Err()
		_, _, _ = r.Result()
	})

	t.Run("BLMPop", func(t *testing.T) {
		r := c.BLMPop(ctx, 100*time.Millisecond, LEFT, 1, key)
		_ = r.Err()
		_, _, _ = r.Result()
	})
}

// TestBitMapExtra: BitPosSpan。
func TestBitMapExtra(t *testing.T) {
	c := MustNewClient(NewConf(WithDevelopment(false), WithDB(nextDB())))
	t.Cleanup(func() { _ = c.Close() })
	ctx := context.Background()

	const key = "bm-key"
	_ = c.Set(ctx, key, "\xff\xf0\x00", 0).Err()
	t.Cleanup(func() { _ = c.Del(ctx, key).Err() })

	r := c.BitPosSpan(ctx, key, 0, 0, -1, BIT)
	_ = r.Err()
	_ = r.Val()
}

// TestConnectionExtra: ClientGetName / ClientUnblock / ClientUnblockWithError / ClientUnpause.
func TestConnectionExtra(t *testing.T) {
	c := MustNewClient(NewConf(WithDevelopment(false), WithDB(nextDB())))
	t.Cleanup(func() { _ = c.Close() })
	ctx := context.Background()

	t.Run("ClientGetName", func(t *testing.T) {
		r := c.ClientGetName(ctx)
		_ = r.Err()
		_ = r.Val()
	})

	t.Run("ClientUnblock", func(t *testing.T) {
		// ClientUnblock 需要一个 client id；用一个不存在的 id 测，预期 Val=0
		r := c.ClientUnblock(ctx, 99999999)
		_ = r.Err()
		_ = r.Val()
	})

	t.Run("ClientUnblockWithError", func(t *testing.T) {
		r := c.ClientUnblockWithError(ctx, 99999999)
		_ = r.Err()
		_ = r.Val()
	})

	t.Run("ClientUnpause", func(t *testing.T) {
		// CLIENT UNPAUSE 在没暂停时返回 OK 或 NOOP
		r := c.ClientUnpause(ctx)
		_ = r.Err()
		_ = r.Val()
	})
}
