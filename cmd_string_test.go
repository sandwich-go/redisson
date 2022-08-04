package redisson

import (
	"context"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func testAppend(ctx context.Context, c Cmdable) []string {
	var key = "key"
	n := c.Exists(ctx, key)
	So(n.Err(), ShouldBeNil)
	So(n.Val(), ShouldEqual, 0)

	a := c.Append(ctx, key, "Hello")
	So(a.Err(), ShouldBeNil)
	So(a.Val(), ShouldEqual, 5)

	a = c.Append(ctx, key, " World")
	So(a.Err(), ShouldBeNil)
	So(a.Val(), ShouldEqual, 11)

	get := c.Get(ctx, key)
	So(get.Err(), ShouldBeNil)
	So(get.Val(), ShouldEqual, "Hello World")

	return []string{key}
}

func testDecr(ctx context.Context, c Cmdable) []string {
	var key = "key"

	set := c.Set(ctx, key, "10", 0)
	So(set.Err(), ShouldBeNil)
	So(set.Val(), ShouldEqual, OK)

	decr := c.Decr(ctx, key)
	So(decr.Err(), ShouldBeNil)
	So(decr.Val(), ShouldEqual, 9)

	set = c.Set(ctx, key, "234293482390480948029348230948", 0)
	So(set.Err(), ShouldBeNil)
	So(set.Val(), ShouldEqual, OK)

	decr = c.Decr(ctx, key)
	So(decr.Err(), ShouldNotBeNil)
	So(decr.Err().Error(), ShouldEqual, "ERR value is not an integer or out of range")

	return []string{key}
}

func testDecrBy(ctx context.Context, c Cmdable) []string {
	var key = "key"

	set := c.Set(ctx, key, "10", 0)
	So(set.Err(), ShouldBeNil)
	So(set.Val(), ShouldEqual, OK)

	decrBy := c.DecrBy(ctx, key, 5)
	So(decrBy.Err(), ShouldBeNil)
	So(decrBy.Val(), ShouldEqual, 5)

	return []string{key}
}

func testGetDel(ctx context.Context, c Cmdable) []string {
	var key = "key"

	set := c.Set(ctx, key, "value", 0)
	So(set.Err(), ShouldBeNil)
	So(set.Val(), ShouldEqual, OK)

	getDel := c.GetDel(ctx, key)
	So(getDel.Err(), ShouldBeNil)
	So(getDel.Val(), ShouldEqual, "value")

	get := c.Get(ctx, key)
	So(get.Err(), ShouldNotBeNil)
	So(IsNil(get.Err()), ShouldBeTrue)

	return []string{key}
}

func testGetEX(ctx context.Context, c Cmdable) []string {
	var key = "key"

	set := c.Set(ctx, key, "value", 100*time.Second)
	So(set.Err(), ShouldBeNil)
	So(set.Val(), ShouldEqual, OK)

	ttl := c.TTL(ctx, key)
	So(ttl.Err(), ShouldBeNil)
	So(ttl.Val(), ShouldNotEqual, 110*time.Second)
	So(ttl.Val(), ShouldNotEqual, 13*time.Second)

	getEX := c.GetEx(ctx, key, 200*time.Second)
	So(getEX.Err(), ShouldBeNil)
	So(getEX.Val(), ShouldEqual, "value")

	ttl = c.TTL(ctx, key)
	So(ttl.Err(), ShouldBeNil)
	So(ttl.Val(), ShouldBeGreaterThan, 100*time.Second)
	So(ttl.Val(), ShouldNotEqual, 3*time.Second)

	return []string{key}
}

func testGetSet(ctx context.Context, c Cmdable) []string {
	var key = "key"

	incr := c.Incr(ctx, key)
	So(incr.Err(), ShouldBeNil)
	So(incr.Val(), ShouldEqual, 1)

	getSet := c.GetSet(ctx, key, "0")
	So(getSet.Err(), ShouldBeNil)
	So(getSet.Val(), ShouldEqual, "1")

	get := c.Get(ctx, key)
	So(get.Err(), ShouldBeNil)
	So(get.Val(), ShouldEqual, "0")

	return []string{key}
}

func testIncr(ctx context.Context, c Cmdable) []string {
	var key = "key"

	set := c.Set(ctx, key, "10", 0)
	So(set.Err(), ShouldBeNil)
	So(set.Val(), ShouldEqual, OK)

	incr := c.Incr(ctx, key)
	So(incr.Err(), ShouldBeNil)
	So(incr.Val(), ShouldEqual, 11)

	get := c.Get(ctx, key)
	So(get.Err(), ShouldBeNil)
	So(get.Val(), ShouldEqual, "11")

	return []string{key}
}

func testIncrBy(ctx context.Context, c Cmdable) []string {
	var key = "key"

	set := c.Set(ctx, key, "10", 0)
	So(set.Err(), ShouldBeNil)
	So(set.Val(), ShouldEqual, OK)

	incrBy := c.IncrBy(ctx, key, 5)
	So(incrBy.Err(), ShouldBeNil)
	So(incrBy.Val(), ShouldEqual, 15)

	return []string{key}
}

func testIncrByFloat(ctx context.Context, c Cmdable) []string {
	var key, key1 = "key", "key1"

	set := c.Set(ctx, key, "10.50", 0)
	So(set.Err(), ShouldBeNil)
	So(set.Val(), ShouldEqual, OK)

	incrByFloat := c.IncrByFloat(ctx, key, 0.1)
	So(incrByFloat.Err(), ShouldBeNil)
	So(incrByFloat.Val(), ShouldEqual, 10.6)

	set = c.Set(ctx, key, "5.0e3", 0)
	So(set.Err(), ShouldBeNil)
	So(set.Val(), ShouldEqual, OK)

	incrByFloat = c.IncrByFloat(ctx, key, 2.0e2)
	So(incrByFloat.Err(), ShouldBeNil)
	So(incrByFloat.Val(), ShouldEqual, 5200)

	incrByFloat = c.IncrByFloat(ctx, key1, 996945661)
	So(incrByFloat.Err(), ShouldBeNil)
	So(incrByFloat.Val(), ShouldEqual, float64(996945661))

	return []string{key, key1}
}

func testMSet(ctx context.Context, c Cmdable) []string {
	var key, key1 = "key1", "key2"

	mSet := c.MSet(ctx, key, "hello1", key1, "hello2")
	So(mSet.Err(), ShouldBeNil)
	So(mSet.Val(), ShouldEqual, OK)

	mGet := c.MGet(ctx, key, key1, "_")
	So(mGet.Err(), ShouldBeNil)
	So(interfaceSliceEqual(mGet.Val(), []interface{}{"hello1", "hello2", nil}), ShouldBeTrue)

	return []string{key, key1}
}

func testMSetNX(ctx context.Context, c Cmdable) []string {
	var key, key1, key2 = "key1", "key2", "key3"

	mSetNX := c.MSetNX(ctx, key, "hello1", key1, "hello2")
	So(mSetNX.Err(), ShouldBeNil)
	So(mSetNX.Val(), ShouldBeTrue)

	mSetNX = c.MSetNX(ctx, key1, "hello1", key2, "hello2")
	So(mSetNX.Err(), ShouldBeNil)
	So(mSetNX.Val(), ShouldBeFalse)

	return []string{key, key1}
}

func testSet(ctx context.Context, c Cmdable) []string {
	var key = "key1"

	set := c.Set(ctx, key, "hello", 100*time.Millisecond)
	So(set.Err(), ShouldBeNil)
	So(set.Val(), ShouldEqual, OK)

	get := c.Get(ctx, key)
	So(get.Err(), ShouldBeNil)
	So(get.Val(), ShouldEqual, "hello")

	time.Sleep(110 * time.Millisecond)

	get = c.Get(ctx, key)
	So(get.Err(), ShouldNotBeNil)
	So(IsNil(get.Err()), ShouldBeTrue)

	set = c.Set(ctx, key, "hello", 5*time.Second)
	So(set.Err(), ShouldBeNil)
	So(set.Val(), ShouldEqual, OK)

	// set with keepttl
	set = c.Set(ctx, key, "hello1", KeepTTL)
	So(set.Err(), ShouldBeNil)
	So(set.Val(), ShouldEqual, OK)

	ttl := c.TTL(ctx, key)
	So(ttl.Err(), ShouldBeNil)
	So(ttl.Val().Nanoseconds(), ShouldNotEqual, -1)

	return []string{key}
}

func testSetEX(ctx context.Context, c Cmdable) []string {
	var key, value = "key", "hello"

	setEX := c.SetEX(ctx, key, value, 1*time.Second)
	So(setEX.Err(), ShouldBeNil)
	So(setEX.Val(), ShouldEqual, OK)

	get := c.Get(ctx, key)
	So(get.Err(), ShouldBeNil)
	So(get.Val(), ShouldEqual, value)

	time.Sleep(1500 * time.Millisecond)

	get = c.Get(ctx, key)
	So(get.Err(), ShouldNotBeNil)
	So(IsNil(get.Err()), ShouldBeTrue)

	return []string{key}
}

func testSetNX(ctx context.Context, c Cmdable) []string {
	var key, value, value1 = "key", "hello", "hello2"

	setNX := c.SetNX(ctx, key, value, 0)
	So(setNX.Err(), ShouldBeNil)
	So(setNX.Val(), ShouldBeTrue)

	setNX = c.SetNX(ctx, key, value1, 0)
	So(setNX.Err(), ShouldBeNil)
	So(setNX.Val(), ShouldBeFalse)

	get := c.Get(ctx, key)
	So(get.Err(), ShouldBeNil)
	So(get.Val(), ShouldEqual, value)

	d := c.Del(ctx, key)
	So(d.Err(), ShouldBeNil)
	So(d.Val(), ShouldEqual, 1)

	setNX = c.SetNX(ctx, key, value, time.Second)
	So(setNX.Err(), ShouldBeNil)
	So(setNX.Val(), ShouldBeTrue)

	setNX = c.SetNX(ctx, key, value1, time.Second)
	So(setNX.Err(), ShouldBeNil)
	So(setNX.Val(), ShouldBeFalse)

	get = c.Get(ctx, key)
	So(get.Err(), ShouldBeNil)
	So(get.Val(), ShouldEqual, value)

	d = c.Del(ctx, key)
	So(d.Err(), ShouldBeNil)
	So(d.Val(), ShouldEqual, 1)

	setNX = c.SetNX(ctx, key, value1, KeepTTL)
	So(setNX.Err(), ShouldBeNil)
	So(setNX.Val(), ShouldBeTrue)

	ttl := c.TTL(ctx, key)
	So(ttl.Err(), ShouldBeNil)
	So(ttl.Val().Nanoseconds(), ShouldEqual, -1)

	return []string{key}
}

func testSetXX(ctx context.Context, c Cmdable) []string {
	var key, value, value1, value2 = "key", "hello", "hello2", "hello3"

	setXX := c.SetXX(ctx, key, value1, 0)
	So(setXX.Err(), ShouldBeNil)
	So(setXX.Val(), ShouldBeFalse)

	So(c.Set(ctx, key, value, 0).Err(), ShouldBeNil)

	setXX = c.SetXX(ctx, key, value1, 0)
	So(setXX.Err(), ShouldBeNil)
	So(setXX.Val(), ShouldBeTrue)

	get := c.Get(ctx, key)
	So(get.Err(), ShouldBeNil)
	So(get.Val(), ShouldEqual, value1)

	d := c.Del(ctx, key)
	So(d.Err(), ShouldBeNil)
	So(d.Val(), ShouldEqual, 1)

	setXX = c.SetXX(ctx, key, value1, time.Second)
	So(setXX.Err(), ShouldBeNil)
	So(setXX.Val(), ShouldBeFalse)

	So(c.Set(ctx, key, value, time.Second).Err(), ShouldBeNil)

	setXX = c.SetXX(ctx, key, value1, time.Second)
	So(setXX.Err(), ShouldBeNil)
	So(setXX.Val(), ShouldBeTrue)

	get = c.Get(ctx, key)
	So(get.Err(), ShouldBeNil)
	So(get.Val(), ShouldEqual, value1)

	d = c.Del(ctx, key)
	So(d.Err(), ShouldBeNil)
	So(d.Val(), ShouldEqual, 1)

	setXX = c.SetXX(ctx, key, value1, time.Second)
	So(setXX.Err(), ShouldBeNil)
	So(setXX.Val(), ShouldBeFalse)

	So(c.Set(ctx, key, value, time.Second).Err(), ShouldBeNil)

	setXX = c.SetXX(ctx, key, value1, 5*time.Second)
	So(setXX.Err(), ShouldBeNil)
	So(setXX.Val(), ShouldBeTrue)

	setXX = c.SetXX(ctx, key, value2, KeepTTL)
	So(setXX.Err(), ShouldBeNil)
	So(setXX.Val(), ShouldBeTrue)

	get = c.Get(ctx, key)
	So(get.Err(), ShouldBeNil)
	So(get.Val(), ShouldEqual, value2)

	// set keepttl will Retain the ttl associated with the key
	ttl := c.TTL(ctx, key)
	So(ttl.Err(), ShouldBeNil)
	So(ttl.Val(), ShouldNotEqual, -1)

	return []string{key}
}

func testSetArgs(ctx context.Context, c Cmdable) []string {
	var key, value = "key", "hello"

	args := SetArgs{
		TTL: 100 * time.Millisecond,
	}
	So(c.SetArgs(ctx, key, value, args).Err(), ShouldBeNil)

	get := c.Get(ctx, key)
	So(get.Err(), ShouldBeNil)
	So(get.Val(), ShouldEqual, value)

	time.Sleep(200 * time.Millisecond)

	get = c.Get(ctx, key)
	So(get.Err(), ShouldNotBeNil)
	So(IsNil(get.Err()), ShouldBeTrue)

	expireAt := time.Now().AddDate(1, 1, 1)
	args = SetArgs{
		ExpireAt: expireAt,
	}
	So(c.SetArgs(ctx, key, value, args).Err(), ShouldBeNil)

	get = c.Get(ctx, key)
	So(get.Err(), ShouldBeNil)
	So(get.Val(), ShouldEqual, value)

	// check the key has an expiration date
	// (so a TTL value different of -1)
	ttl := c.TTL(ctx, key)
	So(ttl.Err(), ShouldBeNil)
	So(ttl.Val(), ShouldNotEqual, -1)

	d := c.Del(ctx, key)
	So(d.Err(), ShouldBeNil)
	So(d.Val(), ShouldEqual, 1)

	args = SetArgs{
		ExpireAt: time.Now().AddDate(-3, 1, 1),
	}
	// redis accepts a timestamp less than the current date
	// but returns nil when trying to get the key
	So(c.SetArgs(ctx, key, value, args).Err(), ShouldBeNil)

	get = c.Get(ctx, key)
	So(get.Err(), ShouldNotBeNil)
	So(IsNil(get.Err()), ShouldBeTrue)
	So(get.Val(), ShouldBeEmpty)

	// Set with ttl
	argsWithTTL := SetArgs{
		TTL: 5 * time.Second,
	}
	set := c.SetArgs(ctx, key, value, argsWithTTL)
	So(set.Err(), ShouldBeNil)
	So(set.Val(), ShouldEqual, OK)

	// Set with keepttl
	argsWithKeepTTL := SetArgs{
		KeepTTL: true,
	}
	set = c.SetArgs(ctx, key, value, argsWithKeepTTL)
	So(set.Err(), ShouldBeNil)
	So(set.Val(), ShouldEqual, OK)

	ttl = c.TTL(ctx, key)
	So(ttl.Err(), ShouldBeNil)
	So(ttl.Val().Nanoseconds(), ShouldNotEqual, -1)

	d = c.Del(ctx, key)
	So(d.Err(), ShouldBeNil)
	So(d.Val(), ShouldEqual, 1)

	So(c.Set(ctx, key, value, 0).Err(), ShouldBeNil)

	args = SetArgs{
		Mode: NX,
	}
	val := c.SetArgs(ctx, key, value, args)
	So(val.Err(), ShouldNotBeNil)
	So(IsNil(val.Err()), ShouldBeTrue)
	So(val.Val(), ShouldBeEmpty)

	d = c.Del(ctx, key)
	So(d.Err(), ShouldBeNil)
	So(d.Val(), ShouldEqual, 1)

	args = SetArgs{
		Mode: NX,
	}

	val = c.SetArgs(ctx, key, value, args)
	So(val.Err(), ShouldBeNil)
	So(val.Val(), ShouldEqual, OK)

	d = c.Del(ctx, key)
	So(d.Err(), ShouldBeNil)
	So(d.Val(), ShouldEqual, 1)

	args = SetArgs{
		Mode: NX,
		Get:  true,
	}
	val = c.SetArgs(ctx, key, value, args)
	So(val.Err(), ShouldNotBeNil)

	d = c.Del(ctx, key)
	So(d.Err(), ShouldBeNil)
	//So(d.Val(), ShouldEqual, 0)

	args = SetArgs{
		TTL:  100 * time.Millisecond,
		Mode: NX,
	}
	val = c.SetArgs(ctx, key, value, args)
	So(val.Err(), ShouldBeNil)
	So(val.Val(), ShouldEqual, OK)

	time.Sleep(200 * time.Millisecond)

	get = c.Get(ctx, key)
	So(get.Err(), ShouldNotBeNil)
	So(IsNil(get.Err()), ShouldBeTrue)
	So(get.Val(), ShouldBeEmpty)

	e := c.Set(ctx, key, value, 0)
	So(e.Err(), ShouldBeNil)

	args = SetArgs{
		TTL:  500 * time.Millisecond,
		Mode: NX,
	}
	val = c.SetArgs(ctx, key, value, args)
	So(val.Err(), ShouldNotBeNil)
	So(IsNil(val.Err()), ShouldBeTrue)
	So(val.Val(), ShouldBeEmpty)

	d = c.Del(ctx, key)
	So(d.Err(), ShouldBeNil)
	So(d.Val(), ShouldEqual, 1)

	args = SetArgs{
		TTL:  500 * time.Millisecond,
		Mode: NX,
		Get:  true,
	}
	val = c.SetArgs(ctx, key, value, args)
	So(val.Err(), ShouldNotBeNil)

	d = c.Del(ctx, key)
	So(d.Err(), ShouldBeNil)
	//So(d.Val(), ShouldEqual, 1)

	args = SetArgs{
		Mode: XX,
	}
	val = c.SetArgs(ctx, key, value, args)
	So(val.Err(), ShouldNotBeNil)
	So(IsNil(val.Err()), ShouldBeTrue)
	So(val.Val(), ShouldBeEmpty)

	e = c.Set(ctx, key, value, 0)
	So(e.Err(), ShouldBeNil)

	args = SetArgs{
		Mode: XX,
	}
	val = c.SetArgs(ctx, key, value, args)
	So(val.Err(), ShouldBeNil)
	So(val.Val(), ShouldEqual, OK)

	d = c.Del(ctx, key)
	So(d.Err(), ShouldBeNil)
	So(d.Val(), ShouldEqual, 1)

	e = c.Set(ctx, key, value, 0)
	So(e.Err(), ShouldBeNil)

	args = SetArgs{
		Mode: XX,
		Get:  true,
	}
	val = c.SetArgs(ctx, key, value, args)
	So(val.Err(), ShouldBeNil)
	So(val.Val(), ShouldEqual, value)

	d = c.Del(ctx, key)
	So(d.Err(), ShouldBeNil)
	So(d.Val(), ShouldEqual, 1)

	args = SetArgs{
		Mode: XX,
		Get:  true,
	}
	val = c.SetArgs(ctx, key, value, args)
	So(val.Err(), ShouldNotBeNil)
	So(IsNil(val.Err()), ShouldBeTrue)
	So(val.Val(), ShouldBeEmpty)

	args = SetArgs{
		TTL:  500 * time.Millisecond,
		Mode: XX,
		Get:  true,
	}

	val = c.SetArgs(ctx, key, value, args)
	So(val.Err(), ShouldNotBeNil)
	So(IsNil(val.Err()), ShouldBeTrue)
	So(val.Val(), ShouldBeEmpty)

	e = c.Set(ctx, key, value, 0)
	So(e.Err(), ShouldBeNil)

	args = SetArgs{
		TTL:  100 * time.Millisecond,
		Mode: XX,
		Get:  true,
	}

	val = c.SetArgs(ctx, key, value, args)
	So(val.Err(), ShouldBeNil)
	So(val.Val(), ShouldEqual, value)

	time.Sleep(200 * time.Millisecond)

	get = c.Get(ctx, key)
	So(get.Err(), ShouldNotBeNil)
	So(IsNil(get.Err()), ShouldBeTrue)
	So(get.Val(), ShouldBeEmpty)

	args = SetArgs{
		Get: true,
	}
	val = c.SetArgs(ctx, key, value, args)
	So(val.Err(), ShouldNotBeNil)
	So(IsNil(val.Err()), ShouldBeTrue)
	So(val.Val(), ShouldBeEmpty)

	e = c.Set(ctx, key, value, 0)
	So(e.Err(), ShouldBeNil)

	args = SetArgs{
		Get: true,
	}
	val = c.SetArgs(ctx, key, value, args)
	So(val.Err(), ShouldBeNil)
	So(val.Val(), ShouldEqual, value)

	return []string{key}
}

func testSetRange(ctx context.Context, c Cmdable) []string {
	var key = "key"
	set := c.Set(ctx, key, "Hello World", 0)
	So(set.Err(), ShouldBeNil)
	So(set.Val(), ShouldEqual, OK)

	range_ := c.SetRange(ctx, key, 6, "Redis")
	So(range_.Err(), ShouldBeNil)
	So(range_.Val(), ShouldEqual, 11)

	get := c.Get(ctx, key)
	So(get.Err(), ShouldBeNil)
	So(get.Val(), ShouldEqual, "Hello Redis")

	return []string{key}
}

func testGet(ctx context.Context, c Cmdable) []string {
	var key, value = "key", "hello"

	get := c.Get(ctx, "_")
	So(get.Err(), ShouldNotBeNil)
	So(IsNil(get.Err()), ShouldBeTrue)
	So(get.Val(), ShouldBeEmpty)

	get = cacheCmd(c).Get(ctx, "_")
	So(get.Err(), ShouldNotBeNil)
	So(IsNil(get.Err()), ShouldBeTrue)
	So(get.Val(), ShouldBeEmpty)

	set := c.Set(ctx, key, value, 0)
	So(set.Err(), ShouldBeNil)
	So(set.Val(), ShouldEqual, OK)

	get = cacheCmd(c).Get(ctx, key)
	So(get.Err(), ShouldBeNil)
	So(get.Val(), ShouldEqual, value)

	get = c.Get(ctx, key)
	So(get.Err(), ShouldBeNil)
	So(get.Val(), ShouldEqual, value)

	return []string{key}
}

func testMGet(ctx context.Context, c Cmdable) []string {
	var key1, value1, key2, value2 = "key1", "hello1", "key2", "hello2"
	mSet := c.MSet(ctx, key1, value1, key2, value2)
	So(mSet.Err(), ShouldBeNil)
	So(mSet.Val(), ShouldEqual, OK)

	mGet := c.MGet(ctx, key1, key2, "_")
	So(mGet.Err(), ShouldBeNil)
	So(interfaceSliceEqual(mGet.Val(), []interface{}{value1, value2, nil}), ShouldBeTrue)

	mGet = cacheCmd(c).MGet(ctx, key1, key2, "_")
	So(mGet.Err(), ShouldBeNil)
	So(interfaceSliceEqual(mGet.Val(), []interface{}{value1, value2, nil}), ShouldBeTrue)

	return []string{key1, key2}
}

func testGetRange(ctx context.Context, c Cmdable) []string {
	var key = "key"

	set := c.Set(ctx, key, "This is a string", 0)
	So(set.Err(), ShouldBeNil)
	So(set.Val(), ShouldEqual, OK)

	getRange := c.GetRange(ctx, key, 0, 3)
	So(getRange.Err(), ShouldBeNil)
	So(getRange.Val(), ShouldEqual, "This")

	getRange = cacheCmd(c).GetRange(ctx, key, -3, -1)
	So(getRange.Err(), ShouldBeNil)
	So(getRange.Val(), ShouldEqual, "ing")

	getRange = c.GetRange(ctx, key, 0, -1)
	So(getRange.Err(), ShouldBeNil)
	So(getRange.Val(), ShouldEqual, "This is a string")

	getRange = c.GetRange(ctx, key, 10, 100)
	So(getRange.Err(), ShouldBeNil)
	So(getRange.Val(), ShouldEqual, "string")

	return []string{key}
}

func testStrLen(ctx context.Context, c Cmdable) []string {
	var key = "key"

	set := c.Set(ctx, "key", "hello", 0)
	So(set.Err(), ShouldBeNil)
	So(set.Val(), ShouldEqual, OK)

	strLen := c.StrLen(ctx, key)
	So(strLen.Err(), ShouldBeNil)
	So(strLen.Val(), ShouldEqual, 5)

	strLen = cacheCmd(c).StrLen(ctx, key)
	So(strLen.Err(), ShouldBeNil)
	So(strLen.Val(), ShouldEqual, 5)

	strLen = c.StrLen(ctx, "_")
	So(strLen.Err(), ShouldBeNil)
	So(strLen.Val(), ShouldEqual, 0)

	return []string{key}
}

func stringTestUnits() []TestUnit {
	return []TestUnit{
		{CommandAppend, testAppend},
		{CommandDecr, testDecr},
		{CommandDecrBy, testDecrBy},
		{CommandGetDel, testGetDel},
		{CommandGetEX, testGetEX},
		{CommandGetSet, testGetSet},
		{CommandIncr, testIncr},
		{CommandIncrBy, testIncrBy},
		{CommandIncrByFloat, testIncrByFloat},
		{CommandMSet, testMSet},
		{CommandMSetNX, testMSetNX},
		{CommandSet, testSet},
		{CommandSetEX, testSetEX},
		{CommandSetNX, testSetNX},
		{CommandSetXX, testSetXX},
		{CommandSet, testSetArgs},
		{CommandSetRange, testSetRange},
		{CommandGet, testGet},
		{CommandMGet, testMGet},
		{CommandGetRange, testGetRange},
		{CommandStrLen, testStrLen},
	}
}

func TestResp2Client_String(t *testing.T) { doTestUnits(t, RESP2, stringTestUnits) }
func TestResp3Client_String(t *testing.T) { doTestUnits(t, RESP3, stringTestUnits) }
