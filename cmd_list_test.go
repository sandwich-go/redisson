package redisson

import (
	"context"
	. "github.com/smartystreets/goconvey/convey"
	"sort"
	"testing"
	"time"
)

func testBLMove(ctx context.Context, c Cmdable) []string {
	var source, destination, value = "lmove1", "lmove2", "ichi"
	bLMove := c.BLMove(ctx, source, destination, "RIGHT", "LEFT", 1*time.Second)
	So(bLMove.Err(), ShouldNotBeNil)
	So(IsNil(bLMove.Err()), ShouldBeTrue)

	rPush := c.RPush(ctx, source, value)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 1)

	bLMove = c.BLMove(ctx, source, destination, "RIGHT", "LEFT", 0)
	So(bLMove.Err(), ShouldBeNil)
	So(bLMove.Val(), ShouldEqual, value)

	return []string{source, destination}
}

func testBLPop(ctx context.Context, c Cmdable) []string {
	var key1, key2, value = "list1", "list2", "ichi"
	bLPop := c.BLPop(ctx, 1*time.Second, key1, key2)
	So(bLPop.Err(), ShouldNotBeNil)
	So(IsNil(bLPop.Err()), ShouldBeTrue)

	rPush := c.RPush(ctx, key1, value, "b", "c")
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 3)

	bLPop = c.BLPop(ctx, 0, key1, key2)
	So(bLPop.Err(), ShouldBeNil)
	So(stringSliceEqual(bLPop.Val(), []string{key1, value}, true), ShouldBeTrue)

	return []string{key1, key2}
}

func testBRPop(ctx context.Context, c Cmdable) []string {
	var key1, key2, value = "list1", "list2", "ichi"
	bRPop := c.BRPop(ctx, 1*time.Second, key1, key2)
	So(bRPop.Err(), ShouldNotBeNil)
	So(IsNil(bRPop.Err()), ShouldBeTrue)

	rPush := c.RPush(ctx, key1, "b", "c", value)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 3)

	bRPop = c.BRPop(ctx, 0, key1, key2)
	So(bRPop.Err(), ShouldBeNil)
	So(stringSliceEqual(bRPop.Val(), []string{key1, value}, true), ShouldBeTrue)

	return []string{key1, key2}
}

func testBRPopLPush(ctx context.Context, c Cmdable) []string {
	var key1, key2, value = "list1", "list2", "ichi"

	bRPopLPush := c.BRPopLPush(ctx, key1, key2, time.Second)
	So(bRPopLPush.Err(), ShouldNotBeNil)
	So(IsNil(bRPopLPush.Err()), ShouldBeTrue)

	rPush := c.RPush(ctx, key1, "a", "b", value)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 3)

	bRPopLPush = c.BRPopLPush(ctx, key1, key2, 0)
	So(bRPopLPush.Err(), ShouldBeNil)
	So(bRPopLPush.Val(), ShouldEqual, value)

	return []string{key1, key2}
}

func testLIndex(ctx context.Context, c Cmdable) []string {
	var key1, value1, value2 = "mylist", "World", "Hello"

	lPush := c.LPush(ctx, key1, value1)
	So(lPush.Err(), ShouldBeNil)
	So(lPush.Val(), ShouldEqual, 1)

	lPush = c.LPush(ctx, key1, value2)
	So(lPush.Err(), ShouldBeNil)
	So(lPush.Val(), ShouldEqual, 2)

	lIndex := c.LIndex(ctx, key1, 0)
	So(lIndex.Err(), ShouldBeNil)
	So(lIndex.Val(), ShouldEqual, value2)

	lIndex = c.LIndex(ctx, key1, -1)
	So(lIndex.Err(), ShouldBeNil)
	So(lIndex.Val(), ShouldEqual, value1)

	lIndex = cacheCmd(c).LIndex(ctx, key1, -1)
	So(lIndex.Err(), ShouldBeNil)
	So(lIndex.Val(), ShouldEqual, value1)

	lIndex = c.LIndex(ctx, key1, 3)
	So(lIndex.Err(), ShouldNotBeNil)
	So(IsNil(lIndex.Err()), ShouldBeTrue)

	return []string{key1}
}

func testLInsert(ctx context.Context, c Cmdable) []string {
	var key1, value1, value2, value3, value4 = "mylist", "Hello", "World", "There", "Four"

	rPush := c.RPush(ctx, key1, value1)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 1)

	rPush = c.RPush(ctx, key1, value2)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 2)

	lInsert := c.LInsert(ctx, key1, "BEFORE", value2, value3)
	So(lInsert.Err(), ShouldBeNil)
	So(lInsert.Val(), ShouldEqual, 3)

	lInsert = c.LInsert(ctx, key1, "AFTER", value2, value4)
	So(lInsert.Err(), ShouldBeNil)
	So(lInsert.Val(), ShouldEqual, 4)

	lRange := c.LRange(ctx, key1, 0, -1)
	So(lRange.Err(), ShouldBeNil)
	So(stringSliceEqual(lRange.Val(), []string{value1, value3, value2, value4}, true), ShouldBeTrue)

	return []string{key1}
}

func testLInsertBefore(ctx context.Context, c Cmdable) []string {
	var key1, value1, value2, value3 = "mylist", "Hello", "World", "There"

	rPush := c.RPush(ctx, key1, value1)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 1)

	rPush = c.RPush(ctx, key1, value2)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 2)

	lInsert := c.LInsertBefore(ctx, key1, value2, value3)
	So(lInsert.Err(), ShouldBeNil)
	So(lInsert.Val(), ShouldEqual, 3)

	lRange := c.LRange(ctx, key1, 0, -1)
	So(lRange.Err(), ShouldBeNil)
	So(stringSliceEqual(lRange.Val(), []string{value1, value3, value2}, true), ShouldBeTrue)

	return []string{key1}
}

func testLInsertAfter(ctx context.Context, c Cmdable) []string {
	var key1, value1, value2, value3 = "mylist", "Hello", "World", "There"

	rPush := c.RPush(ctx, key1, value1)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 1)

	rPush = c.RPush(ctx, key1, value2)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 2)

	lInsert := c.LInsertAfter(ctx, key1, value2, value3)
	So(lInsert.Err(), ShouldBeNil)
	So(lInsert.Val(), ShouldEqual, 3)

	lRange := c.LRange(ctx, key1, 0, -1)
	So(lRange.Err(), ShouldBeNil)
	So(stringSliceEqual(lRange.Val(), []string{value1, value2, value3}, true), ShouldBeTrue)

	return []string{key1}
}

func testLLen(ctx context.Context, c Cmdable) []string {
	var key1, value1, value2, value3 = "mylist", "Hello", "World", "There"

	lPush := c.LPush(ctx, key1, value1)
	So(lPush.Err(), ShouldBeNil)
	So(lPush.Val(), ShouldEqual, 1)

	lPush = c.LPush(ctx, key1, value2)
	So(lPush.Err(), ShouldBeNil)
	So(lPush.Val(), ShouldEqual, 2)

	lPush = c.LPush(ctx, key1, value3)
	So(lPush.Err(), ShouldBeNil)
	So(lPush.Val(), ShouldEqual, 3)

	lLen := c.LLen(ctx, key1)
	So(lLen.Err(), ShouldBeNil)
	So(lLen.Val(), ShouldEqual, 3)

	lLen = cacheCmd(c).LLen(ctx, key1)
	So(lLen.Err(), ShouldBeNil)
	So(lLen.Val(), ShouldEqual, 3)

	return []string{key1}
}

func testLMove(ctx context.Context, c Cmdable) []string {
	var key1, key2, value1, value2, value3 = "mylist", "myotherlist", "one", "two", "three"

	rPush := c.RPush(ctx, key1, value1)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 1)

	rPush = c.RPush(ctx, key1, value2)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 2)

	rPush = c.RPush(ctx, key1, value3)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 3)

	lMove := c.LMove(ctx, key1, key2, "RIGHT", "LEFT")
	So(lMove.Err(), ShouldBeNil)
	So(lMove.Val(), ShouldEqual, value3)

	lMove = c.LMove(ctx, key1, key2, "LEFT", "RIGHT")
	So(lMove.Err(), ShouldBeNil)
	So(lMove.Val(), ShouldEqual, value1)

	lRange := c.LRange(ctx, key1, 0, -1)
	So(lRange.Err(), ShouldBeNil)
	So(stringSliceEqual(lRange.Val(), []string{value2}, true), ShouldBeTrue)

	lRange = c.LRange(ctx, key2, 0, -1)
	So(lRange.Err(), ShouldBeNil)
	So(stringSliceEqual(lRange.Val(), []string{value3, value1}, true), ShouldBeTrue)

	return []string{key1, key2}
}

func testLPop(ctx context.Context, c Cmdable) []string {
	var key1, value1, value2, value3, value4, value5 = "mylist", "one", "two", "three", "four", "five"

	rPush := c.RPush(ctx, key1, value1, value2, value3, value4, value5)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 5)

	lPop := c.LPop(ctx, key1)
	So(lPop.Err(), ShouldBeNil)
	So(lPop.Val(), ShouldEqual, value1)

	lPop = c.LPop(ctx, key1)
	So(lPop.Err(), ShouldBeNil)
	So(lPop.Val(), ShouldEqual, value2)

	lRange := c.LRange(ctx, key1, 0, -1)
	So(lRange.Err(), ShouldBeNil)
	So(stringSliceEqual(lRange.Val(), []string{value3, value4, value5}, true), ShouldBeTrue)

	return []string{key1}
}

func testLPopCount(ctx context.Context, c Cmdable) []string {
	var key1, value1, value2, value3, value4, value5 = "mylist", "one", "two", "three", "four", "five"

	rPush := c.RPush(ctx, key1, value1, value2, value3, value4, value5)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 5)

	lPop := c.LPop(ctx, key1)
	So(lPop.Err(), ShouldBeNil)
	So(lPop.Val(), ShouldEqual, value1)

	lPopCount := c.LPopCount(ctx, key1, 2)
	So(lPopCount.Err(), ShouldBeNil)
	So(stringSliceEqual(lPopCount.Val(), []string{value2, value3}, true), ShouldBeTrue)

	lRange := c.LRange(ctx, key1, 0, -1)
	So(lRange.Err(), ShouldBeNil)
	So(stringSliceEqual(lRange.Val(), []string{value4, value5}, true), ShouldBeTrue)

	return []string{key1}
}

func int64SliceEqual(a, b []int64, absolute bool) bool {
	if a == nil && b != nil {
		return false
	}
	if b == nil && a != nil {
		return false
	}
	if len(b) != len(a) {
		return false
	}
	if !absolute {
		sort.Slice(a, func(i, j int) bool { return a[i] < a[j] })
		sort.Slice(b, func(i, j int) bool { return b[i] < b[j] })
	}
	for k, v := range a {
		if v != b[k] {
			return false
		}
	}
	return true
}

func testLPos(ctx context.Context, c Cmdable) []string {
	var key1 = "mylist"

	rPush := c.RPush(ctx, key1, "a", "b", "c", "d", 1, 2, 3, 4, 3, 3, 3)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 11)

	lPos := c.LPos(ctx, key1, "3", LPosArgs{})
	So(lPos.Err(), ShouldBeNil)
	So(lPos.Val(), ShouldEqual, 6)

	lPos = cacheCmd(c).LPos(ctx, key1, "c", LPosArgs{})
	So(lPos.Err(), ShouldBeNil)
	So(lPos.Val(), ShouldEqual, 2)

	lPos = cacheCmd(c).LPos(ctx, key1, "3", LPosArgs{Rank: 2})
	So(lPos.Err(), ShouldBeNil)
	So(lPos.Val(), ShouldEqual, 8)

	lPos = cacheCmd(c).LPos(ctx, key1, "3", LPosArgs{MaxLen: 1})
	So(lPos.Err(), ShouldNotBeNil)
	So(IsNil(lPos.Err()), ShouldBeTrue)

	return []string{key1}
}

func testLPosCount(ctx context.Context, c Cmdable) []string {
	var key1 = "mylist"

	rPush := c.RPush(ctx, key1, "a", "b", "c", "d", 1, 2, 3, 4, 3, 3, 3)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 11)

	lPosCount := c.LPosCount(ctx, key1, "3", 0, LPosArgs{Rank: 2})
	So(lPosCount.Err(), ShouldBeNil)
	So(int64SliceEqual(lPosCount.Val(), []int64{8, 9, 10}, true), ShouldBeTrue)

	lPosCount = cacheCmd(c).LPosCount(ctx, key1, "3", 2, LPosArgs{Rank: 2})
	So(lPosCount.Err(), ShouldBeNil)
	So(int64SliceEqual(lPosCount.Val(), []int64{8, 9}, true), ShouldBeTrue)

	return []string{key1}
}

func testLPush(ctx context.Context, c Cmdable) []string {
	var key1, value1, value2, value3 = "mylist", "Hello", "World", "There"

	lPush := c.LPush(ctx, key1, value1, value2)
	So(lPush.Err(), ShouldBeNil)
	So(lPush.Val(), ShouldEqual, 2)

	lPush = c.LPush(ctx, key1, value3)
	So(lPush.Err(), ShouldBeNil)
	So(lPush.Val(), ShouldEqual, 3)

	lRange := c.LRange(ctx, key1, 0, -1)
	So(lRange.Err(), ShouldBeNil)
	So(stringSliceEqual(lRange.Val(), []string{value3, value2, value1}, true), ShouldBeTrue)

	return []string{key1}
}

func testLPushX(ctx context.Context, c Cmdable) []string {
	var key1, key2, value1, value2 = "mylist", "myotherlist", "Hello", "World"

	lPush := c.LPush(ctx, key1, value2)
	So(lPush.Err(), ShouldBeNil)
	So(lPush.Val(), ShouldEqual, 1)

	lPush = c.LPushX(ctx, key1, value1)
	So(lPush.Err(), ShouldBeNil)
	So(lPush.Val(), ShouldEqual, 2)

	lPush = c.LPushX(ctx, key2, value1)
	So(lPush.Err(), ShouldBeNil)
	So(lPush.Val(), ShouldEqual, 0)

	lRange := c.LRange(ctx, key1, 0, -1)
	So(lRange.Err(), ShouldBeNil)
	So(stringSliceEqual(lRange.Val(), []string{value1, value2}, true), ShouldBeTrue)

	lRange = c.LRange(ctx, key2, 0, -1)
	So(lRange.Err(), ShouldBeNil)
	So(len(lRange.Val()), ShouldEqual, 0)

	return []string{key1, key2}
}

func testLRange(ctx context.Context, c Cmdable) []string {
	var key1, value1, value2, value3 = "mylist", "one", "two", "three"

	rPush := c.RPush(ctx, key1, value1)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 1)

	rPush = c.RPush(ctx, key1, value2)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 2)

	rPush = c.RPush(ctx, key1, value3)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 3)

	lRange := c.LRange(ctx, key1, 0, 0)
	So(lRange.Err(), ShouldBeNil)
	So(stringSliceEqual(lRange.Val(), []string{value1}, true), ShouldBeTrue)

	lRange = c.LRange(ctx, key1, -3, 2)
	So(lRange.Err(), ShouldBeNil)
	So(stringSliceEqual(lRange.Val(), []string{value1, value2, value3}, true), ShouldBeTrue)

	lRange = cacheCmd(c).LRange(ctx, key1, -100, 100)
	So(lRange.Err(), ShouldBeNil)
	So(stringSliceEqual(lRange.Val(), []string{value1, value2, value3}, true), ShouldBeTrue)

	lRange = c.LRange(ctx, key1, 5, 10)
	So(lRange.Err(), ShouldBeNil)
	So(len(lRange.Val()), ShouldEqual, 0)

	return []string{key1}
}

func testLRem(ctx context.Context, c Cmdable) []string {
	var key1, value1, value2 = "mylist", "hello", "foo"

	rPush := c.RPush(ctx, key1, value1)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 1)

	rPush = c.RPush(ctx, key1, value1)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 2)

	rPush = c.RPush(ctx, key1, value2)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 3)

	rPush = c.RPush(ctx, key1, value1)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 4)

	lRem := c.LRem(ctx, key1, -2, value1)
	So(lRem.Err(), ShouldBeNil)
	So(lRem.Val(), ShouldEqual, 2)

	lRange := cacheCmd(c).LRange(ctx, key1, 0, -1)
	So(lRange.Err(), ShouldBeNil)
	So(stringSliceEqual(lRange.Val(), []string{value1, value2}, true), ShouldBeTrue)

	return []string{key1}
}

func testLSet(ctx context.Context, c Cmdable) []string {
	var key1, value1, value2, value3, value4, value5 = "mylist", "one", "two", "three", "four", "five"

	rPush := c.RPush(ctx, key1, value1)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 1)

	rPush = c.RPush(ctx, key1, value2)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 2)

	rPush = c.RPush(ctx, key1, value3)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 3)

	lSet := c.LSet(ctx, key1, 0, value4)
	So(lSet.Err(), ShouldBeNil)
	So(lSet.Val(), ShouldEqual, OK)

	lSet = c.LSet(ctx, key1, -2, value5)
	So(lSet.Err(), ShouldBeNil)
	So(lSet.Val(), ShouldEqual, OK)

	lRange := cacheCmd(c).LRange(ctx, key1, 0, -1)
	So(lRange.Err(), ShouldBeNil)
	So(stringSliceEqual(lRange.Val(), []string{value4, value5, value3}, true), ShouldBeTrue)

	return []string{key1}
}

func testLTrim(ctx context.Context, c Cmdable) []string {
	var key1, value1, value2, value3 = "mylist", "one", "two", "three"

	rPush := c.RPush(ctx, key1, value1)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 1)

	rPush = c.RPush(ctx, key1, value2)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 2)

	rPush = c.RPush(ctx, key1, value3)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 3)

	lTrim := c.LTrim(ctx, key1, 1, -1)
	So(lTrim.Err(), ShouldBeNil)
	So(lTrim.Val(), ShouldEqual, OK)

	lRange := cacheCmd(c).LRange(ctx, key1, 0, -1)
	So(lRange.Err(), ShouldBeNil)
	So(stringSliceEqual(lRange.Val(), []string{value2, value3}, true), ShouldBeTrue)

	return []string{key1}
}

func testRPop(ctx context.Context, c Cmdable) []string {
	var key1, value1, value2, value3, value4, value5 = "mylist", "one", "two", "three", "four", "five"

	rPush := c.RPush(ctx, key1, value1, value2, value3, value4, value5)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 5)

	rPop := c.RPop(ctx, key1)
	So(rPop.Err(), ShouldBeNil)
	So(rPop.Val(), ShouldEqual, value5)

	rPop = c.RPop(ctx, key1)
	So(rPop.Err(), ShouldBeNil)
	So(rPop.Val(), ShouldEqual, value4)

	lRange := cacheCmd(c).LRange(ctx, key1, 0, -1)
	So(lRange.Err(), ShouldBeNil)
	So(stringSliceEqual(lRange.Val(), []string{value1, value2, value3}, true), ShouldBeTrue)

	return []string{key1}
}

func testRPopCount(ctx context.Context, c Cmdable) []string {
	var key1, value1, value2, value3, value4, value5 = "mylist", "one", "two", "three", "four", "five"

	rPush := c.RPush(ctx, key1, value1, value2, value3, value4, value5)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 5)

	rPop := c.RPop(ctx, key1)
	So(rPop.Err(), ShouldBeNil)
	So(rPop.Val(), ShouldEqual, value5)

	rPopCount := c.RPopCount(ctx, key1, 2)
	So(rPopCount.Err(), ShouldBeNil)
	So(stringSliceEqual(rPopCount.Val(), []string{value4, value3}, true), ShouldBeTrue)

	lRange := cacheCmd(c).LRange(ctx, key1, 0, -1)
	So(lRange.Err(), ShouldBeNil)
	So(stringSliceEqual(lRange.Val(), []string{value1, value2}, true), ShouldBeTrue)

	return []string{key1}
}

func testRPopLPush(ctx context.Context, c Cmdable) []string {
	var key1, key2, value1, value2, value3 = "mylist", "myotherlist", "one", "two", "three"

	rPush := c.RPush(ctx, key1, value1)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 1)

	rPush = c.RPush(ctx, key1, value2)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 2)

	rPush = c.RPush(ctx, key1, value3)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 3)

	rPopLPush := c.RPopLPush(ctx, key1, key2)
	So(rPopLPush.Err(), ShouldBeNil)
	So(rPopLPush.Val(), ShouldEqual, value3)

	lRange := cacheCmd(c).LRange(ctx, key1, 0, -1)
	So(lRange.Err(), ShouldBeNil)
	So(stringSliceEqual(lRange.Val(), []string{value1, value2}, true), ShouldBeTrue)

	lRange = cacheCmd(c).LRange(ctx, key2, 0, -1)
	So(lRange.Err(), ShouldBeNil)
	So(stringSliceEqual(lRange.Val(), []string{value3}, true), ShouldBeTrue)

	return []string{key1, key2}
}

func testRPush(ctx context.Context, c Cmdable) []string {
	var key1, value1, value2 = "mylist", "one", "two"

	rPush := c.RPush(ctx, key1, value1)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 1)

	rPush = c.RPush(ctx, key1, value2)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 2)

	lRange := c.LRange(ctx, key1, 0, -1)
	So(lRange.Err(), ShouldBeNil)
	So(stringSliceEqual(lRange.Val(), []string{value1, value2}, true), ShouldBeTrue)

	return []string{key1}
}

func testRPushX(ctx context.Context, c Cmdable) []string {
	var key1, key2, value1, value2 = "mylist", "myotherlist", "one", "two"

	rPush := c.RPush(ctx, key1, value1)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 1)

	rPush = c.RPushX(ctx, key1, value2)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 2)

	rPush = c.RPushX(ctx, key2, value2)
	So(rPush.Err(), ShouldBeNil)
	So(rPush.Val(), ShouldEqual, 0)

	lRange := c.LRange(ctx, key1, 0, -1)
	So(lRange.Err(), ShouldBeNil)
	So(stringSliceEqual(lRange.Val(), []string{value1, value2}, true), ShouldBeTrue)

	lRange = c.LRange(ctx, key2, 0, -1)
	So(lRange.Err(), ShouldBeNil)
	So(len(lRange.Val()), ShouldEqual, 0)

	return []string{key1}
}

func listTestUnits() []TestUnit {
	return []TestUnit{
		{CommandBLMove, testBLMove},
		{CommandBLPop, testBLPop},
		{CommandBRPop, testBRPop},
		{CommandBRPopLPush, testBRPopLPush},
		{CommandLIndex, testLIndex},
		{CommandLInsert, testLInsert},
		{CommandLInsert, testLInsertBefore},
		{CommandLInsert, testLInsertAfter},
		{CommandLLen, testLLen},
		{CommandLMove, testLMove},
		{CommandLPop, testLPop},
		{CommandLPopCount, testLPopCount},
		{CommandLPos, testLPos},
		{CommandLPos, testLPosCount},
		{CommandLPush, testLPush},
		{CommandLPushX, testLPushX},
		{CommandLRange, testLRange},
		{CommandLRem, testLRem},
		{CommandLSet, testLSet},
		{CommandLTrim, testLTrim},
		{CommandRPop, testRPop},
		{CommandRPopCount, testRPopCount},
		{CommandRPopLPush, testRPopLPush},
		{CommandRPush, testRPush},
		{CommandRPushX, testRPushX},
	}
}

func TestResp2Client_List(t *testing.T) { doTestUnits(t, RESP2, listTestUnits) }
func TestResp3Client_List(t *testing.T) { doTestUnits(t, RESP3, listTestUnits) }
