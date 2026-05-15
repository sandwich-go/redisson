//go:build integration

package redisson

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// TestZRangeRevByScoreUserCompat 验证 ZRangeStore / ZRangeArgsWithScores 在
// Rev+ByScore 时,redisson 层正确处理 Start/Stop 交换:用户传 Start<Stop+Rev=true
// 仍能拿到正确的反向区间结果 (rueidiscompat v1.0.74+ 不再做这一交换补偿,
// redisson 改走自己的 builder 路径来覆盖)。
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

		// 用户传 Start<Stop+Rev=true,期望按 score 反向取窗口的 [1,2] 段 (Offset=1, Count=2)。
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
