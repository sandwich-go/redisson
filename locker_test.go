//go:build integration

package redisson

import (
	"context"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestLocker_TryAndForce(t *testing.T) {
	c := MustNewClient(NewConf(WithDevelopment(false)))
	t.Cleanup(func() { _ = c.Close() })

	l, err := c.NewLocker()
	if err != nil {
		t.Fatalf("NewLocker: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	const name = "rss-locker-test"

	Convey("Try acquires when free", t, func() {
		c1, cancel1, err := l.TryWithContext(ctx, name)
		So(err, ShouldBeNil)
		So(c1, ShouldNotBeNil)
		defer cancel1()

		// 第二次 Try 应该失败（已被持有）
		_, _, err2 := l.TryWithContext(ctx, name)
		So(err2, ShouldNotBeNil)
	})

	Convey("Force takes over", t, func() {
		c1, cancel1, err := l.WithContext(ctx, name)
		So(err, ShouldBeNil)
		defer cancel1()
		So(c1, ShouldNotBeNil)

		// Force 把锁夺过来
		c2, cancel2, err := l.ForceWithContext(ctx, name)
		So(err, ShouldBeNil)
		defer cancel2()
		So(c2, ShouldNotBeNil)
	})
}
