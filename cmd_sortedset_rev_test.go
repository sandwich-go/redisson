//go:build integration

package redisson

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// TestZRangeRevByScoreUserCompat 回归 rueidiscompat v1.0.74+ regression：
// PR #979 删除了 ZRange / ZRangeStore / ZRangeArgsWithScores 的 Rev+ByScore 时
// Start/Stop 自动交换补偿。redisson 旧 API 约定下用户传 Min<Max 顺序即可，
// 升级 rueidis 后会静默返回空集；本测试验证 redisson 层补回该补偿。
func TestZRangeRevByScoreUserCompat(t *testing.T) {
	if !realRedisAvailable {
		t.Skip("real Redis not available; this test runs under -tags integration")
	}
	c := MustNewClient(NewConf(WithDevelopment(false), WithDB(nextDB())))
	t.Cleanup(func() { _ = c.Close() })
	ctx := context.Background()
	c.FlushDB(ctx)

	Convey("ZRangeStore with Rev+ByScore", t, func() {
		c.ZAddArgs(ctx, "src", ZAddArgs{
			Members: []Z{
				{Score: 1, Member: "one"},
				{Score: 2, Member: "two"},
				{Score: 3, Member: "three"},
				{Score: 4, Member: "four"},
			},
		})

		// 用户传 Start < Stop 顺序 + Rev=true。
		// rueidis v1.0.74+ regression：如果不补偿会返回 0，目标集合为空。
		got := c.ZRangeStore(ctx, "dst", ZRangeArgs{
			Key:     "src",
			Start:   1,
			Stop:    4,
			ByScore: true,
			Rev:     true,
			Offset:  1,
			Count:   2,
		})
		So(got.Err(), ShouldBeNil)
		So(got.Val(), ShouldEqual, 2)

		// dst 应该按 score 升序保存（two=2, three=3）
		members := c.ZRange(ctx, "dst", 0, -1)
		So(members.Err(), ShouldBeNil)
		So(stringSliceEqual(members.Val(), []string{"two", "three"}, true), ShouldBeTrue)

		c.Del(ctx, "src", "dst")
	})

	Convey("ZRangeArgsWithScores with Rev+ByScore", t, func() {
		c.ZAddArgs(ctx, "rs", ZAddArgs{
			Members: []Z{
				{Score: 1, Member: "one"},
				{Score: 2, Member: "two"},
				{Score: 3, Member: "three"},
			},
		})

		got := c.ZRangeArgsWithScores(ctx, ZRangeArgs{
			Key:     "rs",
			Start:   1,
			Stop:    3,
			ByScore: true,
			Rev:     true,
		})
		So(got.Err(), ShouldBeNil)
		So(zsEqual(got.Val(), []Z{
			{Score: 3, Member: "three"},
			{Score: 2, Member: "two"},
			{Score: 1, Member: "one"},
		}), ShouldBeTrue)

		c.Del(ctx, "rs")
	})
}
