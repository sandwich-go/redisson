//go:build integration

package redisson

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRateLimiter_AllowAndCheck(t *testing.T) {
	c := MustNewClient(NewConf(WithDevelopment(false)))
	t.Cleanup(func() { _ = c.Close() })

	rl, err := c.NewRateLimiter(
		WithRateLimiterOptionLimit(3),
		WithRateLimiterOptionWindow(time.Second),
		WithRateLimiterOptionKeyPrefix("rss-rl-test"),
	)
	if err != nil {
		t.Fatalf("NewRateLimiter: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const id = "user-1"

	Convey("Allow consumes one slot", t, func() {
		// 用 unix nano 拼出独立 identifier，避免不同 sub-test 串扰
		uid := id + "-" + time.Now().Format("150405.000000")
		r1, err := rl.Allow(ctx, uid)
		So(err, ShouldBeNil)
		So(r1.Allowed, ShouldBeTrue)
	})

	Convey("Check does not consume", t, func() {
		uid := id + "-check-" + time.Now().Format("150405.000000")
		// Check N 次都不应消耗配额
		for i := 0; i < 5; i++ {
			r, err := rl.Check(ctx, uid)
			So(err, ShouldBeNil)
			So(r.Allowed, ShouldBeTrue)
		}
	})

	Convey("Limit returns configured value", t, func() {
		So(rl.Limit(), ShouldEqual, 3)
	})

	Convey("AllowN consumes N slots", t, func() {
		uid := id + "-alln-" + time.Now().Format("150405.000000")
		r, err := rl.AllowN(ctx, uid, 3)
		So(err, ShouldBeNil)
		So(r.Allowed, ShouldBeTrue)

		// 第 4 次再来 1 个应被拒
		r2, err := rl.AllowN(ctx, uid, 1)
		So(err, ShouldBeNil)
		So(r2.Allowed, ShouldBeFalse)
	})
}
