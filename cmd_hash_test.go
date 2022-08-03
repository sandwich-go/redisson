package sandwich_redis

import (
	"context"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func testHDel(ctx context.Context, c Cmdable) []string {
	var key, field1, field2, value1 = "myhash", "field1", "field2", "foo"

	hset := c.HSet(ctx, key, field1, value1)
	So(hset.Err(), ShouldBeNil)
	So(hset.Val(), ShouldEqual, 1)

	hdel := c.HDel(ctx, key, field1, field2)
	So(hdel.Err(), ShouldBeNil)
	So(hdel.Val(), ShouldEqual, 1)

	return []string{key}
}

func testHExists(ctx context.Context, c Cmdable) []string {
	var key, key1, field1, field2, value1 = "myhash", "myhash1", "field1", "field2", "foo"

	hset := c.HSet(ctx, key, field1, value1)
	So(hset.Err(), ShouldBeNil)
	So(hset.Val(), ShouldEqual, 1)

	hExists := c.HExists(ctx, key, field1)
	So(hExists.Err(), ShouldBeNil)
	So(hExists.Val(), ShouldBeTrue)

	hExists = cacheCmd(c).HExists(ctx, key, field1)
	So(hExists.Err(), ShouldBeNil)
	So(hExists.Val(), ShouldBeTrue)

	hExists = c.HExists(ctx, key, field2)
	So(hExists.Err(), ShouldBeNil)
	So(hExists.Val(), ShouldBeFalse)

	hExists = cacheCmd(c).HExists(ctx, key, field2)
	So(hExists.Err(), ShouldBeNil)
	So(hExists.Val(), ShouldBeFalse)

	hExists = c.HExists(ctx, key1, field1)
	So(hExists.Err(), ShouldBeNil)
	So(hExists.Val(), ShouldBeFalse)

	hExists = cacheCmd(c).HExists(ctx, key1, field1)
	So(hExists.Err(), ShouldBeNil)
	So(hExists.Val(), ShouldBeFalse)

	del := c.Del(ctx, key)
	So(del.Err(), ShouldBeNil)
	So(del.Val(), ShouldEqual, 1)

	hExists = cacheCmd(c).HExists(ctx, key, field1)
	So(hExists.Err(), ShouldBeNil)
	So(hExists.Val(), ShouldBeFalse)

	return nil
}

func testHGet(ctx context.Context, c Cmdable) []string {
	var key, key1, field1, field2, value1 = "myhash", "myhash1", "field1", "field2", "foo"

	hset := c.HSet(ctx, key, field1, value1)
	So(hset.Err(), ShouldBeNil)
	So(hset.Val(), ShouldEqual, 1)

	hget := c.HGet(ctx, key, field1)
	So(hget.Err(), ShouldBeNil)
	So(hget.Val(), ShouldEqual, value1)

	hget = cacheCmd(c).HGet(ctx, key, field1)
	So(hget.Err(), ShouldBeNil)
	So(hget.Val(), ShouldEqual, value1)

	hget = c.HGet(ctx, key, field2)
	So(hget.Err(), ShouldNotBeNil)
	So(IsNil(hget.Err()), ShouldBeTrue)

	hget = c.HGet(ctx, key1, field1)
	So(hget.Err(), ShouldNotBeNil)
	So(IsNil(hget.Err()), ShouldBeTrue)

	del := c.Del(ctx, key)
	So(del.Err(), ShouldBeNil)
	So(del.Val(), ShouldEqual, 1)

	hget = cacheCmd(c).HGet(ctx, key, field1)
	So(hget.Err(), ShouldNotBeNil)
	So(IsNil(hget.Err()), ShouldBeTrue)

	return nil
}

func stringStringMapEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	if a == nil && b != nil {
		return false
	}
	if a != nil && b == nil {
		return false
	}
	for k, v := range a {
		v1, ok := b[k]
		if !ok || v != v1 {
			return false
		}
	}
	return true
}

func testHGetAll(ctx context.Context, c Cmdable) []string {
	var key, key1, field1, field2, value1, value2 = "myhash", "myhash1", "field1", "field2", "foo", "foo1"

	hset := c.HSet(ctx, key, field1, value1, field2, value2)
	So(hset.Err(), ShouldBeNil)
	So(hset.Val(), ShouldEqual, 2)

	hgetAll := cacheCmd(c).HGetAll(ctx, key)
	So(hgetAll.Err(), ShouldBeNil)
	So(stringStringMapEqual(hgetAll.Val(), map[string]string{field1: value1, field2: value2}), ShouldBeTrue)

	hgetAll = c.HGetAll(ctx, key)
	So(hgetAll.Err(), ShouldBeNil)
	So(stringStringMapEqual(hgetAll.Val(), map[string]string{field1: value1, field2: value2}), ShouldBeTrue)

	hgetAll = c.HGetAll(ctx, key1)
	So(hgetAll.Err(), ShouldBeNil)
	So(len(hgetAll.Val()), ShouldEqual, 0)

	return []string{key}
}

func testHIncrBy(ctx context.Context, c Cmdable) []string {
	var key, field1 = "myhash", "field1"
	hset := c.HSet(ctx, key, field1, 5)
	So(hset.Err(), ShouldBeNil)
	So(hset.Val(), ShouldEqual, 1)

	hIncrBy := c.HIncrBy(ctx, key, field1, 1)
	So(hIncrBy.Err(), ShouldBeNil)
	So(hIncrBy.Val(), ShouldEqual, 6)

	hIncrBy = c.HIncrBy(ctx, key, field1, -1)
	So(hIncrBy.Err(), ShouldBeNil)
	So(hIncrBy.Val(), ShouldEqual, 5)

	hIncrBy = c.HIncrBy(ctx, key, field1, -10)
	So(hIncrBy.Err(), ShouldBeNil)
	So(hIncrBy.Val(), ShouldEqual, -5)

	return []string{key}
}

func testHIncrByFloat(ctx context.Context, c Cmdable) []string {
	var key, field1 = "myhash", "field1"
	hset := c.HSet(ctx, key, field1, 10.50)
	So(hset.Err(), ShouldBeNil)
	So(hset.Val(), ShouldEqual, 1)

	hIncrBy := c.HIncrByFloat(ctx, key, field1, 0.1)
	So(hIncrBy.Err(), ShouldBeNil)
	So(hIncrBy.Val(), ShouldEqual, 10.6)

	hIncrBy = c.HIncrByFloat(ctx, key, field1, -5)
	So(hIncrBy.Err(), ShouldBeNil)
	So(hIncrBy.Val(), ShouldEqual, 5.6)

	hset = c.HSet(ctx, key, field1, 5.0e3)
	So(hset.Err(), ShouldBeNil)
	So(hset.Val(), ShouldEqual, 0)

	hIncrBy = c.HIncrByFloat(ctx, key, field1, 2.0e2)
	So(hIncrBy.Err(), ShouldBeNil)
	So(hIncrBy.Val(), ShouldEqual, 5200)

	return []string{key}
}

func testHKeys(ctx context.Context, c Cmdable) []string {
	var key, field1, field2, value1, value2 = "myhash", "field1", "field2", "foo", "foo1"

	hset := c.HSet(ctx, key, field1, value1, field2, value2)
	So(hset.Err(), ShouldBeNil)
	So(hset.Val(), ShouldEqual, 2)

	hKeys := c.HKeys(ctx, key)
	So(hKeys.Err(), ShouldBeNil)
	So(stringSliceEqual(hKeys.Val(), []string{field1, field2}, true), ShouldBeTrue)

	hKeys = cacheCmd(c).HKeys(ctx, key)
	So(hKeys.Err(), ShouldBeNil)
	So(stringSliceEqual(hKeys.Val(), []string{field1, field2}, true), ShouldBeTrue)

	return []string{key}
}

func testHLen(ctx context.Context, c Cmdable) []string {
	var key, field1, field2, value1, value2 = "myhash", "field1", "field2", "foo", "foo1"

	hset := c.HSet(ctx, key, field1, value1, field2, value2)
	So(hset.Err(), ShouldBeNil)
	So(hset.Val(), ShouldEqual, 2)

	hLen := c.HLen(ctx, key)
	So(hLen.Err(), ShouldBeNil)
	So(hLen.Val(), ShouldEqual, 2)

	hLen = cacheCmd(c).HLen(ctx, key)
	So(hLen.Err(), ShouldBeNil)
	So(hLen.Val(), ShouldEqual, 2)

	return []string{key}
}

func testHMGet(ctx context.Context, c Cmdable) []string {
	var key, field1, field2, field3, value1, value2 = "myhash", "field1", "field2", "nofield", "foo", "foo1"

	hset := c.HSet(ctx, key, field1, value1, field2, value2)
	So(hset.Err(), ShouldBeNil)
	So(hset.Val(), ShouldEqual, 2)

	hMGet := c.HMGet(ctx, key, field1, field2, field3)
	So(hMGet.Err(), ShouldBeNil)
	So(interfaceSliceEqual(hMGet.Val(), []interface{}{value1, value2, nil}), ShouldBeTrue)

	hMGet = cacheCmd(c).HMGet(ctx, key, field1, field2, field3)
	So(hMGet.Err(), ShouldBeNil)
	So(interfaceSliceEqual(hMGet.Val(), []interface{}{value1, value2, nil}), ShouldBeTrue)

	return []string{key}
}

func testHMSet(ctx context.Context, c Cmdable) []string {
	var key, field1, field2, field3, value1, value2 = "myhash", "field1", "field2", "nofield", "foo", "foo1"

	hset := c.HMSet(ctx, key, field1, value1, field2, value2)
	So(hset.Err(), ShouldBeNil)
	So(hset.Val(), ShouldBeTrue)

	hMGet := c.HMGet(ctx, key, field1, field2, field3)
	So(hMGet.Err(), ShouldBeNil)
	So(interfaceSliceEqual(hMGet.Val(), []interface{}{value1, value2, nil}), ShouldBeTrue)

	return []string{key}
}

func testHRandField(ctx context.Context, c Cmdable) []string {
	var key = "coin"
	hset := c.HMSet(ctx, key, "heads", "obverse", "tails", "reverse", "edge", "null")
	So(hset.Err(), ShouldBeNil)
	So(hset.Val(), ShouldBeTrue)

	h := c.HRandField(ctx, key, 0, false)
	So(h.Err(), ShouldBeNil)
	So(len(h.Val()), ShouldEqual, 0)

	h = c.HRandField(ctx, key, 1, false)
	So(h.Err(), ShouldBeNil)
	So(len(h.Val()), ShouldEqual, 1)

	h = c.HRandField(ctx, key, 1, true)
	So(h.Err(), ShouldBeNil)
	So(len(h.Val()), ShouldEqual, 2)

	h = c.HRandField(ctx, key, -5, true)
	So(h.Err(), ShouldBeNil)
	So(len(h.Val()), ShouldEqual, 10)

	return []string{key}
}

func testHScan(ctx context.Context, c Cmdable) []string {
	var key = "myhash"
	for i := 0; i < 1000; i++ {
		sadd := c.HSet(ctx, key, fmt.Sprintf("key%d", i), "hello")
		So(sadd.Err(), ShouldBeNil)
	}
	keys, cursor, err := c.HScan(ctx, key, 0, "", 0).Result()
	So(err, ShouldBeNil)
	So(len(keys), ShouldNotBeZeroValue)
	So(cursor, ShouldNotBeZeroValue)

	return []string{key}
}

func testHSet(ctx context.Context, c Cmdable) []string {
	var key, field1, field2, value1, value2 = "myhash", "field1", "field2", "foo", "foo1"
	hset := c.HSet(ctx, key, field1, value1, field2, value2)
	So(hset.Err(), ShouldBeNil)
	So(hset.Val(), ShouldEqual, 2)

	hGet := c.HGet(ctx, key, field1)
	So(hGet.Err(), ShouldBeNil)
	So(hGet.Val(), ShouldEqual, value1)

	return []string{key}
}

func testHSetNX(ctx context.Context, c Cmdable) []string {
	var key, field1, value1, value2 = "myhash", "field1", "Hello", "World"
	hset := c.HSetNX(ctx, key, field1, value1)
	So(hset.Err(), ShouldBeNil)
	So(hset.Val(), ShouldBeTrue)

	hset = c.HSetNX(ctx, key, field1, value2)
	So(hset.Err(), ShouldBeNil)
	So(hset.Val(), ShouldBeFalse)

	hGet := c.HGet(ctx, key, field1)
	So(hGet.Err(), ShouldBeNil)
	So(hGet.Val(), ShouldEqual, value1)

	return []string{key}
}

func testHVals(ctx context.Context, c Cmdable) []string {
	var key, field1, field2, value1, value2 = "myhash", "field1", "field2", "foo", "foo1"
	hset := c.HSet(ctx, key, field1, value1, field2, value2)
	So(hset.Err(), ShouldBeNil)
	So(hset.Val(), ShouldEqual, 2)

	hVals := c.HVals(ctx, key)
	So(hVals.Err(), ShouldBeNil)
	So(len(hVals.Val()), ShouldEqual, 2)
	So(stringSliceEqual(hVals.Val(), []string{value1, value2}, true), ShouldBeTrue)

	hVals = cacheCmd(c).HVals(ctx, key)
	So(hVals.Err(), ShouldBeNil)
	So(len(hVals.Val()), ShouldEqual, 2)
	So(stringSliceEqual(hVals.Val(), []string{value1, value2}, true), ShouldBeTrue)

	return []string{key}
}

func hashTestUnits() []TestUnit {
	return []TestUnit{
		{CommandHDel, testHDel},
		{CommandHExists, testHExists},
		{CommandHGet, testHGet},
		{CommandHGetAll, testHGetAll},
		{CommandHIncrBy, testHIncrBy},
		{CommandHIncrByFloat, testHIncrByFloat},
		{CommandHKeys, testHKeys},
		{CommandHLen, testHLen},
		{CommandHMGet, testHMGet},
		{CommandHMSet, testHMSet},
		{CommandHRandField, testHRandField},
		{CommandHScan, testHScan},
		{CommandHSet, testHSet},
		{CommandHSetNX, testHSetNX},
		{CommandHVals, testHVals},
	}
}

func TestResp2Client_Hash(t *testing.T) { doTestUnits(t, RESP2, hashTestUnits) }
func TestResp3Client_Hash(t *testing.T) { doTestUnits(t, RESP3, hashTestUnits) }
