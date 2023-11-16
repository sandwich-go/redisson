package redisson

import (
	"context"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"sort"
	"testing"
	"time"
)

func testCopy(ctx context.Context, c Cmdable) []string {
	var key, dest1, dest2, value1, value2 = "dolly:{1}", "clone1:{1}", "clone2:{1}", "sheep1", "sheep2"
	s := c.Set(ctx, key, value1, 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	s = c.Set(ctx, dest2, value2, 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	i := c.Copy(ctx, key, dest1, 0, false)
	So(i.Err(), ShouldBeNil)
	So(i.Val(), ShouldEqual, 1)

	i = c.Copy(ctx, key, dest2, 0, false)
	So(i.Err(), ShouldBeNil)
	So(i.Val(), ShouldEqual, 0)

	s = c.Get(ctx, dest1)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, value1)

	s = c.Get(ctx, dest2)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, value2)

	i = c.Copy(ctx, key, dest2, 0, true)
	So(i.Err(), ShouldBeNil)
	So(i.Val(), ShouldEqual, 1)

	s = c.Get(ctx, dest2)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, value1)

	return []string{key, dest1, dest2}
}

func testDel(ctx context.Context, c Cmdable) []string {
	var key1, key2, value1, value2 = "key1:{1}", "key2:{1}", "Hello", "World"
	s := c.Set(ctx, key1, value1, 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	s = c.Set(ctx, key2, value2, 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	return []string{key1, key2}
}

func testDump(ctx context.Context, c Cmdable) []string {
	var key1 = "mykey"
	s := c.Set(ctx, key1, 10, 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	s = c.Dump(ctx, key1)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldNotBeEmpty)
	//So(s.Val(), ShouldEqual, "\x00\xc0\n\t\x00\xbem\x06\x89Z(\x00\n")

	return []string{key1}
}

func testExists(ctx context.Context, c Cmdable) []string {
	var key1, key2, nosuchkey, value1, value2 = "key1:{1}", "key2:{1}", "nosuchkey:{1}", "Hello", "World"
	s := c.Set(ctx, key1, value1, 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	b := c.Exists(ctx, key1)
	So(b.Err(), ShouldBeNil)
	So(b.Val(), ShouldEqual, 1)

	b = c.Exists(ctx, nosuchkey)
	So(b.Err(), ShouldBeNil)
	So(b.Val(), ShouldEqual, 0)

	b = cacheCmd(c).Exists(ctx, nosuchkey)
	So(b.Err(), ShouldBeNil)
	So(b.Val(), ShouldEqual, 0)

	s = c.Set(ctx, key2, value2, 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	b = cacheCmd(c).Exists(ctx, key1, nosuchkey, key2)
	So(b.Err(), ShouldBeNil)
	So(b.Val(), ShouldEqual, 2)

	b = c.Exists(ctx, key1, nosuchkey, key2)
	So(b.Err(), ShouldBeNil)
	So(b.Val(), ShouldEqual, 2)

	d := c.Del(ctx, key1, key2)
	So(d.Err(), ShouldBeNil)
	So(d.Val(), ShouldEqual, 2)

	time.Sleep(1 * time.Second)

	b = cacheCmd(c).Exists(ctx, key1, nosuchkey, key2)
	So(b.Err(), ShouldBeNil)
	So(b.Val(), ShouldEqual, 0)

	b = c.Exists(ctx, key1, nosuchkey, key2)
	So(b.Err(), ShouldBeNil)
	So(b.Val(), ShouldEqual, 0)

	return nil
}

func testExpire(ctx context.Context, c Cmdable) []string {
	var key, nonexistent_key = "key", "nonexistent_key"
	s := c.Set(ctx, key, "Hello", 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	e := c.Expire(ctx, key, 10*time.Second)
	So(e.Err(), ShouldBeNil)
	So(e.Val(), ShouldBeTrue)

	ttl := c.TTL(ctx, key)
	So(ttl.Err(), ShouldBeNil)
	So(ttl.Val(), ShouldEqual, 10*time.Second)

	s = c.Set(ctx, key, "Hello World", 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	ttl = c.TTL(ctx, key)
	So(ttl.Err(), ShouldBeNil)
	So(ttl.Val(), ShouldEqual, -1)

	ttl = c.TTL(ctx, nonexistent_key)
	So(ttl.Err(), ShouldBeNil)
	So(ttl.Val(), ShouldEqual, -2)

	return []string{key}
}

func testExpireAt(ctx context.Context, c Cmdable) []string {
	var key = "key"
	s := c.Set(ctx, key, "Hello", 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	n := c.Exists(ctx, key)
	So(n.Err(), ShouldBeNil)
	So(n.Val(), ShouldEqual, 1)

	expireAt := c.ExpireAt(ctx, key, time.Now().Add(-time.Hour))
	So(expireAt.Err(), ShouldBeNil)
	So(expireAt.Val(), ShouldBeTrue)

	n = c.Exists(ctx, key)
	So(n.Err(), ShouldBeNil)
	So(n.Val(), ShouldEqual, 0)

	return []string{key}
}

func interfaceSliceEqual(a, b []interface{}) bool {
	if a == nil && b != nil {
		return false
	}
	if b == nil && a != nil {
		return false
	}
	if len(b) != len(a) {
		return false
	}
	for k, v := range a {
		if v != b[k] {
			return false
		}
	}
	return true
}

func stringSliceEqual(a, b []string, absolute bool) bool {
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
		sort.Strings(a)
		sort.Strings(b)
	}
	for k, v := range a {
		if v != b[k] {
			return false
		}
	}
	return true
}

func testKeys(ctx context.Context, c Cmdable) []string {
	mset := c.MSet(ctx, "one", "1", "two", "2", "three", "3", "four", "4")
	So(mset.Err(), ShouldBeNil)
	So(mset.Val(), ShouldEqual, OK)

	keys := c.Keys(ctx, "*o*")
	So(keys.Err(), ShouldBeNil)
	So(stringSliceEqual(keys.Val(), []string{"four", "one", "two"}, false), ShouldBeTrue)

	keys = c.Keys(ctx, "t??")
	So(keys.Err(), ShouldBeNil)
	So(stringSliceEqual(keys.Val(), []string{"two"}, false), ShouldBeTrue)

	keys = c.Keys(ctx, "*")
	So(keys.Err(), ShouldBeNil)
	So(stringSliceEqual(keys.Val(), []string{"four", "one", "three", "two"}, false), ShouldBeTrue)

	return keys.Val()
}

func testMigrate(ctx context.Context, c Cmdable) []string {
	var key, redisSecondaryPort = "key", "6389"
	migrate := c.Migrate(ctx, "localhost", redisSecondaryPort, key, 0, 0)
	So(migrate.Err(), ShouldBeNil)
	So(migrate.Val(), ShouldEqual, "NOKEY")

	s := c.Set(ctx, key, "hello", 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	migrate = c.Migrate(ctx, "localhost", redisSecondaryPort, key, 0, 0)
	So(migrate.Err(), ShouldNotBeNil)
	So(migrate.Err().Error(), ShouldContainSubstring, "IOERR error or timeout writing to target instance")

	return []string{key}
}

func testMove(ctx context.Context, c Cmdable) []string {
	_ = c.FlushAll(ctx)

	var key = "key"
	move := c.Move(ctx, key, 2)
	So(move.Err(), ShouldBeNil)
	So(move.Val(), ShouldBeFalse)

	s := c.Set(ctx, key, "hello", 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	move = c.Move(ctx, key, 2)
	So(move.Err(), ShouldBeNil)
	So(move.Val(), ShouldBeTrue)

	g := c.Get(ctx, key)
	So(g.Err(), ShouldNotBeNil)
	So(IsNil(g.Err()), ShouldBeTrue)
	So(g.Val(), ShouldBeEmpty)

	del := c.FlushAll(ctx)
	So(del.Err(), ShouldBeNil)

	return []string{key}
}

func testObjectEncoding(ctx context.Context, c Cmdable) []string {
	var key = "key"
	s := c.Set(ctx, key, "hello", 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	err := c.ObjectEncoding(ctx, key).Err()
	So(err, ShouldBeNil)

	return []string{key}
}

func testObjectRefCount(ctx context.Context, c Cmdable) []string {
	var key = "key"
	s := c.Set(ctx, key, "hello", 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	refCount := c.ObjectRefCount(ctx, key)
	So(refCount.Err(), ShouldBeNil)
	So(refCount.Val(), ShouldEqual, 1)

	return []string{key}
}

func testObjectIdleTime(ctx context.Context, c Cmdable) []string {
	start := nowFunc()
	var key = "key"
	s := c.Set(ctx, key, "hello", 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	idleTime := c.ObjectIdleTime(ctx, key)
	So(idleTime.Err(), ShouldBeNil)

	//Redis returned milliseconds/1000, which may cause ObjectIdleTime to be at a critical value,
	//should be +1s to deal with the critical value problem.
	//if too much time (>1s) is used during command execution, it may also cause the test to fail.
	//so the ObjectIdleTime result should be <=now-start+1s
	//link: https://github.com/redis/redis/blob/5b48d900498c85bbf4772c1d466c214439888115/src/object.c#L1265-L1272
	So(idleTime.Val(), ShouldBeLessThanOrEqualTo, time.Now().Sub(start)+time.Second)

	return []string{key}
}

func testPersist(ctx context.Context, c Cmdable) []string {
	var key = "key"
	s := c.Set(ctx, key, "Hello", 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	expire := c.Expire(ctx, key, 10*time.Second)
	So(expire.Err(), ShouldBeNil)
	So(expire.Val(), ShouldBeTrue)

	ttl := c.TTL(ctx, key)
	So(ttl.Err(), ShouldBeNil)
	So(ttl.Val(), ShouldEqual, 10*time.Second)

	persist := c.Persist(ctx, key)
	So(persist.Err(), ShouldBeNil)
	So(persist.Val(), ShouldBeTrue)

	ttl = c.TTL(ctx, key)
	So(ttl.Err(), ShouldBeNil)
	So(ttl.Val() < 0, ShouldBeTrue)

	return []string{key}
}

func testPExpire(ctx context.Context, c Cmdable) []string {
	var key = "key"
	s := c.Set(ctx, key, "Hello", 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	expiration := 900 * time.Millisecond
	pexpire := c.PExpire(ctx, key, expiration)
	So(pexpire.Err(), ShouldBeNil)
	So(pexpire.Val(), ShouldBeTrue)

	ttl := c.TTL(ctx, key)
	So(ttl.Err(), ShouldBeNil)
	So(ttl.Val(), ShouldEqual, time.Second)

	pttl := c.PTTL(ctx, key)
	So(pttl.Err(), ShouldBeNil)
	So(pttl.Val(), ShouldNotEqual, 100*time.Millisecond)

	return []string{key}
}

func testPExpireAt(ctx context.Context, c Cmdable) []string {
	var key = "key"
	s := c.Set(ctx, key, "Hello", 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	expiration := 900 * time.Millisecond
	pexpireat := c.PExpireAt(ctx, key, time.Now().Add(expiration))
	So(pexpireat.Err(), ShouldBeNil)
	So(pexpireat.Val(), ShouldBeTrue)

	ttl := c.TTL(ctx, key)
	So(ttl.Err(), ShouldBeNil)
	So(ttl.Val(), ShouldEqual, time.Second)

	pttl := c.PTTL(ctx, key)
	So(pttl.Err(), ShouldBeNil)
	So(pttl.Val(), ShouldNotEqual, 100*time.Millisecond)

	return []string{key}
}

func testPTTL(ctx context.Context, c Cmdable) []string {
	var key = "key"
	s := c.Set(ctx, key, "Hello", 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	expiration := time.Second
	expire := c.Expire(ctx, key, expiration)
	So(expire.Err(), ShouldBeNil)
	So(expire.Val(), ShouldBeTrue)

	pttl := c.PTTL(ctx, key)
	So(pttl.Err(), ShouldBeNil)
	So(pttl.Val(), ShouldNotEqual, 100*time.Millisecond)

	return []string{key}
}

func testRandomKey(ctx context.Context, c Cmdable) []string {
	randomKey := c.RandomKey(ctx)
	So(randomKey.Err(), ShouldNotBeNil)
	So(IsNil(randomKey.Err()), ShouldBeTrue)
	So(randomKey.Val(), ShouldBeEmpty)

	var key = "key"
	s := c.Set(ctx, key, "Hello", 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	randomKey = c.RandomKey(ctx)
	So(randomKey.Err(), ShouldBeNil)
	So(randomKey.Val(), ShouldEqual, key)

	return []string{key}
}

func testRename(ctx context.Context, c Cmdable) []string {
	var key, key1, value = "key:{1}", "key1:{1}", "hello"
	s := c.Set(ctx, key, value, 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	status := c.Rename(ctx, key, key1)
	So(status.Err(), ShouldBeNil)
	So(status.Val(), ShouldEqual, OK)

	get := c.Get(ctx, key1)
	So(get.Err(), ShouldBeNil)
	So(get.Val(), ShouldEqual, value)

	return []string{key, key1}
}

func testRenameNX(ctx context.Context, c Cmdable) []string {
	var key, key1, key2, value = "key:{1}", "key1:{1}", "key2:{1}", "hello"
	s := c.Set(ctx, key, value, 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	s = c.Set(ctx, key2, value, 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	renameNX := c.RenameNX(ctx, key, key1)
	So(renameNX.Err(), ShouldBeNil)
	So(renameNX.Val(), ShouldBeTrue)

	get := c.Get(ctx, key1)
	So(get.Err(), ShouldBeNil)
	So(get.Val(), ShouldEqual, value)

	renameNX = c.RenameNX(ctx, key1, key2)
	So(renameNX.Err(), ShouldBeNil)
	So(renameNX.Val(), ShouldBeFalse)

	return []string{key, key1, key2}
}

func testRestore(ctx context.Context, c Cmdable) []string {
	var key, key1, value = "key:{1}", "key1:{1}", "hello"
	s := c.Set(ctx, key, value, 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	s = c.Set(ctx, key1, value, 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	dump := c.Dump(ctx, key)
	So(dump.Err(), ShouldBeNil)

	d := c.Del(ctx, key)
	So(d.Err(), ShouldBeNil)
	So(d.Val(), ShouldEqual, 1)

	expiration := time.Second
	restore := c.Restore(ctx, key, expiration, dump.Val())
	So(restore.Err(), ShouldBeNil)
	So(restore.Val(), ShouldEqual, OK)

	restore = c.Restore(ctx, key1, 0, dump.Val())
	So(restore.Err(), ShouldNotBeNil)
	//So(restore.Val(), ShouldBeEmpty)

	ty := c.Type(ctx, key)
	So(ty.Err(), ShouldBeNil)
	So(ty.Val(), ShouldEqual, "string")

	val := c.Get(ctx, key)
	So(val.Err(), ShouldBeNil)
	So(val.Val(), ShouldEqual, value)

	pttl := c.PTTL(ctx, key)
	So(pttl.Err(), ShouldBeNil)
	So(pttl.Val(), ShouldBeGreaterThan, 0)
	So(pttl.Val(), ShouldNotEqual, 100*time.Millisecond)

	return []string{key, key1}
}

func testRestoreReplace(ctx context.Context, c Cmdable) []string {
	var key, key1, value, value1 = "key:{1}", "key1:{1}", "hello", "world"
	s := c.Set(ctx, key, value, 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	s = c.Set(ctx, key1, value1, 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	dump := c.Dump(ctx, key)
	So(dump.Err(), ShouldBeNil)

	expiration := time.Second
	restore := c.RestoreReplace(ctx, key, expiration, dump.Val())
	So(restore.Err(), ShouldBeNil)
	So(restore.Val(), ShouldEqual, OK)

	restore = c.RestoreReplace(ctx, key1, expiration, dump.Val())
	So(restore.Err(), ShouldBeNil)
	So(restore.Val(), ShouldEqual, OK)

	ty := c.Type(ctx, key)
	So(ty.Err(), ShouldBeNil)
	So(ty.Val(), ShouldEqual, "string")

	pttl := c.PTTL(ctx, key)
	So(pttl.Err(), ShouldBeNil)
	So(pttl.Val(), ShouldBeGreaterThan, 0)
	So(pttl.Val(), ShouldNotEqual, 100*time.Millisecond)

	val := c.Get(ctx, key)
	So(val.Err(), ShouldBeNil)
	So(val.Val(), ShouldEqual, value)

	val = c.Get(ctx, key)
	So(val.Err(), ShouldBeNil)
	So(val.Val(), ShouldEqual, value)

	return []string{key, key1}
}

func testSort(ctx context.Context, c Cmdable) []string {
	var key = "list"
	size := c.LPush(ctx, key, "1")
	So(size.Err(), ShouldBeNil)
	So(size.Val(), ShouldEqual, 1)

	size = c.LPush(ctx, key, "3")
	So(size.Err(), ShouldBeNil)
	So(size.Val(), ShouldEqual, 2)

	size = c.LPush(ctx, key, "2")
	So(size.Err(), ShouldBeNil)
	So(size.Val(), ShouldEqual, 3)

	els := c.Sort(ctx, key, Sort{
		Offset: 0,
		Count:  2,
		Order:  "ASC",
	})
	So(els.Err(), ShouldBeNil)
	So(stringSliceEqual(els.Val(), []string{"1", "2"}, true), ShouldBeTrue)

	els = c.Sort(ctx, key, Sort{
		Offset: 0,
		Count:  2,
		Order:  "ASC",
	})
	So(els.Err(), ShouldBeNil)
	So(stringSliceEqual(els.Val(), []string{"1", "2"}, true), ShouldBeTrue)

	return []string{key}
}

func testSortAndGet(ctx context.Context, c Cmdable) []string {
	var key, key1, key2 = "list", "object_2", "hello_3"
	size := c.LPush(ctx, key, "1")
	So(size.Err(), ShouldBeNil)
	So(size.Val(), ShouldEqual, 1)

	size = c.LPush(ctx, key, "3")
	So(size.Err(), ShouldBeNil)
	So(size.Val(), ShouldEqual, 2)

	size = c.LPush(ctx, key, "2")
	So(size.Err(), ShouldBeNil)
	So(size.Val(), ShouldEqual, 3)

	s := c.Set(ctx, key1, "value2", 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	s = c.Set(ctx, key2, "value3", 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	{
		els := c.Sort(ctx, key, Sort{
			Get: []string{"object_*", "hello_*"},
		})
		So(els.Err(), ShouldBeNil)
		So(stringSliceEqual(els.Val(), []string{"", "", "value2", "", "", "value3"}, true), ShouldBeTrue)

		els = c.Sort(ctx, key, Sort{
			Get: []string{"object_*", "hello_*"},
		})
		So(els.Err(), ShouldBeNil)
		So(stringSliceEqual(els.Val(), []string{"", "", "value2", "", "", "value3"}, true), ShouldBeTrue)
	}

	{
		els := c.SortInterfaces(ctx, key, Sort{
			Get: []string{"object_*", "hello_*"},
		})
		So(els.Err(), ShouldBeNil)
		So(interfaceSliceEqual(els.Val(), []interface{}{nil, nil, "value2", nil, nil, "value3"}), ShouldBeTrue)

		els = c.SortInterfaces(ctx, key, Sort{
			Get: []string{"object_*", "hello_*"},
		})
		So(els.Err(), ShouldBeNil)
		So(interfaceSliceEqual(els.Val(), []interface{}{nil, nil, "value2", nil, nil, "value3"}), ShouldBeTrue)
	}

	return []string{key, key1, key2}
}

func testSortAndStore(ctx context.Context, c Cmdable) []string {
	var key, key1 = "list:{1}", "list2:{1}"
	size := c.LPush(ctx, key, "1")
	So(size.Err(), ShouldBeNil)
	So(size.Val(), ShouldEqual, 1)

	size = c.LPush(ctx, key, "3")
	So(size.Err(), ShouldBeNil)
	So(size.Val(), ShouldEqual, 2)

	size = c.LPush(ctx, key, "2")
	So(size.Err(), ShouldBeNil)
	So(size.Val(), ShouldEqual, 3)

	n := c.SortStore(ctx, key, key1, Sort{
		Offset: 0,
		Count:  2,
		Order:  "ASC",
	})
	So(n.Err(), ShouldBeNil)
	So(n.Val(), ShouldEqual, 2)

	els := c.LRange(ctx, key1, 0, -1)
	So(els.Err(), ShouldBeNil)
	So(stringSliceEqual(els.Val(), []string{"1", "2"}, true), ShouldBeTrue)

	return []string{key, key1}
}

func testTouch(ctx context.Context, c Cmdable) []string {
	var key, key1, key2 = "touch1", "touch2", "touch3"
	set1 := c.Set(ctx, key, "hello", 0)
	So(set1.Err(), ShouldBeNil)
	So(set1.Val(), ShouldEqual, OK)

	set2 := c.Set(ctx, key1, "hello", 0)
	So(set2.Err(), ShouldBeNil)
	So(set2.Val(), ShouldEqual, OK)

	touch := c.Touch(ctx, key, key1, key2)
	So(touch.Err(), ShouldBeNil)
	So(touch.Val(), ShouldEqual, 2)

	return []string{key, key1, key2}
}

func testTTL(ctx context.Context, c Cmdable) []string {
	var key = "key"
	ttl := c.TTL(ctx, key)
	So(ttl.Err(), ShouldBeNil)
	So(ttl.Val() < 0, ShouldBeTrue)

	set := c.Set(ctx, key, "hello", 0)
	So(set.Err(), ShouldBeNil)
	So(set.Val(), ShouldEqual, OK)

	expire := c.Expire(ctx, key, 60*time.Second)
	So(expire.Err(), ShouldBeNil)
	So(expire.Val(), ShouldBeTrue)

	ttl = c.TTL(ctx, key)
	So(ttl.Err(), ShouldBeNil)
	So(ttl.Val(), ShouldEqual, 60*time.Second)

	return []string{key}
}

func testType(ctx context.Context, c Cmdable) []string {
	var key = "key"
	set := c.Set(ctx, key, "hello", 0)
	So(set.Err(), ShouldBeNil)
	So(set.Val(), ShouldEqual, OK)

	ty := c.Type(ctx, key)
	So(ty.Err(), ShouldBeNil)
	So(ty.Val(), ShouldEqual, "string")

	ty = cacheCmd(c).Type(ctx, key)
	So(ty.Err(), ShouldBeNil)
	So(ty.Val(), ShouldEqual, "string")

	return []string{key}
}

func testUnlink(ctx context.Context, c Cmdable) []string {
	var key1, key2, key3 = "key1", "key2", "key3"
	set := c.Set(ctx, key1, "Hello", 0)
	So(set.Err(), ShouldBeNil)
	So(set.Val(), ShouldEqual, OK)

	set = c.Set(ctx, key2, "World", 0)
	So(set.Err(), ShouldBeNil)
	So(set.Val(), ShouldEqual, OK)

	n := c.Unlink(ctx, key1, key2, key3)
	So(n.Err(), ShouldBeNil)
	So(n.Val(), ShouldEqual, 2)

	get := c.Get(ctx, key1)
	So(get.Err(), ShouldNotBeNil)

	e := c.Exists(ctx, key2)
	So(n.Err(), ShouldBeNil)
	So(e.Val(), ShouldEqual, 0)

	return nil
}

func testScan(ctx context.Context, c Cmdable) []string {
	var allKeys []string
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		allKeys = append(allKeys, key)
		set := c.Set(ctx, key, "hello", 0)
		So(set.Err(), ShouldBeNil)
	}

	keys, cursor, err := c.Scan(ctx, 0, "", 0).Result()
	So(err, ShouldBeNil)
	So(keys, ShouldNotBeEmpty)
	So(cursor, ShouldNotBeZeroValue)

	return allKeys
}

func testScanType(ctx context.Context, c Cmdable) []string {
	var allKeys []string
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		allKeys = append(allKeys, key)
		set := c.Set(ctx, key, "hello", 0)
		So(set.Err(), ShouldBeNil)
	}

	keys, cursor, err := c.ScanType(ctx, 0, "", 0, "string").Result()
	So(err, ShouldBeNil)
	So(keys, ShouldNotBeEmpty)
	So(cursor, ShouldNotBeZeroValue)

	return allKeys
}

func genericTestUnits() []TestUnit {
	return []TestUnit{
		{CommandCopy, testCopy},
		{CommandDel, testDel},
		{CommandDump, testDump},
		{CommandExists, testExists},
		{CommandExpire, testExpire},
		{CommandExpireAt, testExpireAt},
		{CommandKeys, testKeys},
		{CommandMigrate, testMigrate},
		{CommandMove, testMove},
		{CommandObjectEncoding, testObjectEncoding},
		{CommandObjectRefCount, testObjectRefCount},
		{CommandObjectIdleTime, testObjectIdleTime},
		{CommandPersist, testPersist},
		{CommandPExpire, testPExpire},
		{CommandPExpireAt, testPExpireAt},
		{CommandPTTL, testPTTL},
		{CommandRandomKey, testRandomKey},
		{CommandRename, testRename},
		{CommandRenameNX, testRenameNX},
		{CommandRestore, testRestore},
		{CommandRestoreReplace, testRestoreReplace},
		{CommandSort, testSort},
		{CommandSort, testSortAndGet},
		{CommandSort, testSortAndStore},
		{CommandTouch, testTouch},
		{CommandTTL, testTTL},
		{CommandType, testType},
		{CommandUnlink, testUnlink},
		{CommandScan, testScan},
		{CommandScanType, testScanType},
	}
}

func TestResp2Client_Generic(t *testing.T) { doTestUnits(t, RESP2, genericTestUnits) }
func TestResp3Client_Generic(t *testing.T) { doTestUnits(t, RESP3, genericTestUnits) }
