package redisson

import (
	"context"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func testPFAdd(ctx context.Context, c Cmdable) []string {
	var key = "hll"
	pFAdd := c.PFAdd(ctx, key, "a", "b", "c", "d", "e", "f", "g")
	So(pFAdd.Err(), ShouldBeNil)
	So(pFAdd.Val(), ShouldEqual, 1)

	pFCount := c.PFCount(ctx, key)
	So(pFCount.Err(), ShouldBeNil)
	So(pFCount.Val(), ShouldEqual, 7)

	return []string{key}
}

func testPFCount(ctx context.Context, c Cmdable) []string {
	var key, key1 = "hll", "some-other-hll"

	pFAdd := c.PFAdd(ctx, key, "foo", "bar", "zap")
	So(pFAdd.Err(), ShouldBeNil)
	So(pFAdd.Val(), ShouldEqual, 1)

	pFAdd = c.PFAdd(ctx, key, "zap", "zap", "zap")
	So(pFAdd.Err(), ShouldBeNil)
	So(pFAdd.Val(), ShouldEqual, 0)

	pFAdd = c.PFAdd(ctx, key, "foo", "bar")
	So(pFAdd.Err(), ShouldBeNil)
	So(pFAdd.Val(), ShouldEqual, 0)

	pFCount := c.PFCount(ctx, key)
	So(pFCount.Err(), ShouldBeNil)
	So(pFCount.Val(), ShouldEqual, 3)

	pFAdd = c.PFAdd(ctx, key1, 1, 2, 3)
	So(pFAdd.Err(), ShouldBeNil)
	So(pFAdd.Val(), ShouldEqual, 1)

	pFCount = c.PFCount(ctx, key, key1)
	So(pFCount.Err(), ShouldBeNil)
	So(pFCount.Val(), ShouldEqual, 6)

	return []string{key, key1}
}

func testPFMerge(ctx context.Context, c Cmdable) []string {
	var key, key1, key2 = "hll1", "hll2", "hll3"

	pFAdd := c.PFAdd(ctx, key, "foo", "bar", "zap", "a")
	So(pFAdd.Err(), ShouldBeNil)
	So(pFAdd.Val(), ShouldEqual, 1)

	pFAdd = c.PFAdd(ctx, key1, "a", "b", "c", "foo")
	So(pFAdd.Err(), ShouldBeNil)
	So(pFAdd.Val(), ShouldEqual, 1)

	pFMerge := c.PFMerge(ctx, key2, key, key1)
	So(pFMerge.Err(), ShouldBeNil)
	So(pFMerge.Val(), ShouldEqual, OK)

	pFCount := c.PFCount(ctx, key2)
	So(pFCount.Err(), ShouldBeNil)
	So(pFCount.Val(), ShouldEqual, 6)

	return []string{key, key1, key2}
}

func hyperLogTestUnits() []TestUnit {
	return []TestUnit{
		{CommandPFAdd, testPFAdd},
		{CommandPFCount, testPFCount},
		{CommandPFMerge, testPFMerge},
	}
}

func TestClient_HyperLog(t *testing.T) { doTestUnits(t, hyperLogTestUnits) }
