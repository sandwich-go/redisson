package redisson

import (
	"context"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func testSafeMGet(ctx context.Context, c Cmdable) []string {
	var key, key1, key2, key3 = "key1:{1}", "key2", "key3", "key4:{1}"
	So(slot(key), ShouldNotEqual, slot(key1))
	So(slot(key1), ShouldNotEqual, slot(key2))
	So(slot(key), ShouldEqual, slot(key3))

	mSet := c.MSet(ctx, key, "hello1")
	So(mSet.Err(), ShouldBeNil)
	So(mSet.Val(), ShouldEqual, OK)

	So(c.Get(ctx, key).Val(), ShouldEqual, "hello1")

	mSet = c.MSet(ctx, key3, "hello4")
	So(mSet.Err(), ShouldBeNil)
	So(mSet.Val(), ShouldEqual, OK)

	mSet = c.MSet(ctx, key2, "hello3")
	So(mSet.Err(), ShouldBeNil)
	So(mSet.Val(), ShouldEqual, OK)

	mSet = c.MSet(ctx, key1, "hello2")
	So(mSet.Err(), ShouldBeNil)
	So(mSet.Val(), ShouldEqual, OK)

	mGet := c.SafeMGet(ctx, key, key1, key2, key3, "_")
	So(mGet.Err(), ShouldBeNil)
	So(interfaceSliceEqual(mGet.Val(), []any{"hello1", "hello2", "hello3", "hello4", nil}), ShouldBeTrue)

	return []string{key, key1}
}

func safeTestUnits() []TestUnit {
	return []TestUnit{
		{CommandMGet, testSafeMGet},
	}
}

func TestClient_Safe(t *testing.T) { doClusterTestUnits(t, safeTestUnits) }
