package redisson

import (
	"context"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

const testCacheTTL = 1 * time.Minute

func cacheCmd(c Cmdable) CacheCmdable {
	return c.Cache(testCacheTTL)
}

func testBitCount(ctx context.Context, c Cmdable) []string {
	var key, value = "mykey", "foobar"
	s := c.Set(ctx, key, value, 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	for _, v := range []struct {
		key      string
		bc       *BitCount
		expected int64
	}{
		{key, nil, 26},
		{key, &BitCount{Start: 0, End: 0}, 4},
		{key, &BitCount{Start: 1, End: 1}, 6},
		{value, nil, 0},
	} {
		bc := c.BitCount(ctx, v.key, v.bc)
		So(bc.Err(), ShouldBeNil)
		So(bc.Val(), ShouldEqual, v.expected)

		bc = cacheCmd(c).BitCount(ctx, v.key, v.bc)
		So(bc.Err(), ShouldBeNil)
		So(bc.Val(), ShouldEqual, v.expected)

		bc = cacheCmd(c).BitCount(ctx, v.key, v.bc)
		So(bc.Err(), ShouldBeNil)
		So(bc.Val(), ShouldEqual, v.expected)
	}
	return []string{key}
}

func testBitField(ctx context.Context, c Cmdable) []string {
	var keys = []string{"mykey:{1}", "mystring:{1}"}
	for _, v := range []struct {
		key      string
		args     []interface{}
		expected []int64
	}{
		{keys[0], []interface{}{"INCRBY", "i5", 100, 1, "GET", "u4", 0}, []int64{1, 0}},
		{keys[1], []interface{}{"SET", "i8", "#0", 100, "SET", "i8", "#1", 200}, []int64{0, 0}},
		{keys[0], []interface{}{"incrby", "u2", 100, 1, "OVERFLOW", "SAT", "incrby", "u2", 102, 1}, []int64{1, 1}},
		{keys[0], []interface{}{"incrby", "u2", 100, 1, "OVERFLOW", "SAT", "incrby", "u2", 102, 1}, []int64{2, 2}},
		{keys[0], []interface{}{"incrby", "u2", 100, 1, "OVERFLOW", "SAT", "incrby", "u2", 102, 1}, []int64{3, 3}},
		{keys[0], []interface{}{"incrby", "u2", 100, 1, "OVERFLOW", "SAT", "incrby", "u2", 102, 1}, []int64{0, 3}},
	} {
		bc := c.BitField(ctx, v.key, v.args...)
		So(bc.Err(), ShouldBeNil)
		So(len(bc.Val()), ShouldEqual, len(v.expected))
		So(bc.Val(), ShouldResemble, v.expected)
	}
	return keys
}

func testBitOpAnd(ctx context.Context, c Cmdable) []string {
	var key1, key2, dest = "key1:{1}", "key2:{1}", "dest:{1}"
	s := c.Set(ctx, key1, "1", 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	s = c.Set(ctx, key2, "0", 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	bo := c.BitOpAnd(ctx, dest, key1, key2)
	So(bo.Err(), ShouldBeNil)
	So(bo.Val(), ShouldEqual, 1)

	g := c.Get(ctx, dest)
	So(g.Err(), ShouldBeNil)
	So(g.Val(), ShouldEqual, "0")
	return nil
}

func testBitOpOr(ctx context.Context, c Cmdable) []string {
	var key1, key2, dest = "key1:{1}", "key2:{1}", "dest:{1}"
	s := c.Set(ctx, key1, "1", 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	s = c.Set(ctx, key2, "0", 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	bo := c.BitOpOr(ctx, dest, key1, key2)
	So(bo.Err(), ShouldBeNil)
	So(bo.Val(), ShouldEqual, 1)

	g := c.Get(ctx, dest)
	So(g.Err(), ShouldBeNil)
	So(g.Val(), ShouldEqual, "1")
	return nil
}

func testBitOpXor(ctx context.Context, c Cmdable) []string {
	var key1, key2, dest = "key1:{1}", "key2:{1}", "dest:{1}"
	s := c.Set(ctx, key1, "\xff", 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	s = c.Set(ctx, key2, "\x0f", 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	bo := c.BitOpXor(ctx, dest, key1, key2)
	So(bo.Err(), ShouldBeNil)
	So(bo.Val(), ShouldEqual, 1)

	g := c.Get(ctx, dest)
	So(g.Err(), ShouldBeNil)
	So(g.Val(), ShouldEqual, "\xf0")

	return nil
}

func testBitOpNot(ctx context.Context, c Cmdable) []string {
	var key1, key2, dest = "key1:{1}", "key2:{1}", "dest:{1}"
	s := c.Set(ctx, key1, "\x00", 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	bo := c.BitOpNot(ctx, dest, key1)
	So(bo.Err(), ShouldBeNil)
	So(bo.Val(), ShouldEqual, 1)

	g := c.Get(ctx, dest)
	So(g.Err(), ShouldBeNil)
	So(g.Val(), ShouldEqual, "\xff")

	return []string{key1, key2, dest}
}

func testBitPos(ctx context.Context, c Cmdable) []string {
	var key1 = "key1"
	s := c.Set(ctx, key1, "\xff\xf0\x00", 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, OK)

	pos := c.BitPos(ctx, key1, 0)
	So(pos.Err(), ShouldBeNil)
	So(pos.Val(), ShouldEqual, 12)

	pos = c.BitPos(ctx, key1, 1)
	So(pos.Err(), ShouldBeNil)
	So(pos.Val(), ShouldEqual, 0)

	pos = c.BitPos(ctx, key1, 0, 2)
	So(pos.Err(), ShouldBeNil)
	So(pos.Val(), ShouldEqual, 16)

	pos = c.BitPos(ctx, key1, 1, 2)
	So(pos.Err(), ShouldBeNil)
	So(pos.Val(), ShouldEqual, -1)

	pos = c.BitPos(ctx, key1, 0, -1)
	So(pos.Err(), ShouldBeNil)
	So(pos.Val(), ShouldEqual, 16)

	pos = c.BitPos(ctx, key1, 1, -1)
	So(s.Err(), ShouldBeNil)
	So(pos.Val(), ShouldEqual, -1)

	pos = c.BitPos(ctx, key1, 0, 2, 1)
	So(pos.Err(), ShouldBeNil)
	So(pos.Val(), ShouldEqual, -1)

	pos = c.BitPos(ctx, key1, 0, 0, -3)
	So(pos.Err(), ShouldBeNil)
	So(pos.Val(), ShouldEqual, -1)

	pos = c.BitPos(ctx, key1, 0, 0, 0)
	So(pos.Err(), ShouldBeNil)
	So(pos.Val(), ShouldEqual, -1)

	return []string{key1}
}

func testGetBit(ctx context.Context, c Cmdable) []string {
	var key1 = "key1"
	s := c.SetBit(ctx, key1, 7, 1)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, 0)

	s = c.SetBit(ctx, key1, 7, 1)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, 1)

	g := c.GetBit(ctx, key1, 0)
	So(g.Err(), ShouldBeNil)
	So(g.Val(), ShouldEqual, 0)

	g = c.GetBit(ctx, key1, 7)
	So(g.Err(), ShouldBeNil)
	So(g.Val(), ShouldEqual, 1)

	g = cacheCmd(c).GetBit(ctx, key1, 100)
	So(g.Err(), ShouldBeNil)
	So(g.Val(), ShouldEqual, 0)

	d := c.Del(ctx, key1)
	So(d.Err(), ShouldBeNil)
	So(d.Val(), ShouldEqual, 1)

	return []string{key1}
}

func testSetBit(ctx context.Context, c Cmdable) []string {
	var key1 = "key1"
	s := c.SetBit(ctx, key1, 7, 1)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, 0)

	s = c.SetBit(ctx, key1, 7, 0)
	So(s.Err(), ShouldBeNil)
	So(s.Val(), ShouldEqual, 1)

	g := c.Get(ctx, key1)
	So(g.Err(), ShouldBeNil)
	So(g.Val(), ShouldEqual, "\x00")

	d := c.Del(ctx, key1)
	So(d.Err(), ShouldBeNil)
	So(d.Val(), ShouldEqual, 1)

	return []string{key1}
}

func doTestUnitClean(ctx context.Context, c Cmdable, keys []string) {
	if len(keys) > 0 {
		So(c.Del(ctx, keys...).Err(), ShouldBeNil)
	}
}

type TestUnitName interface {
	String() string
}

type TestUnit struct {
	Name TestUnitName
	Func func(ctx context.Context, c Cmdable) []string
}

func bitMapTestUnits() []TestUnit {
	return []TestUnit{
		{CommandBitCount, testBitCount},
		{CommandBitField, testBitField},
		{CommandBitOpAnd, testBitOpAnd},
		{CommandBitOpOr, testBitOpOr},
		{CommandBitOpXor, testBitOpXor},
		{CommandBitOpNot, testBitOpNot},
		{CommandBitPos, testBitPos},
		{CommandGetBit, testGetBit},
		{CommandSetBit, testSetBit},
	}
}

func doTestUnits(t *testing.T, r RESP, unitsFunc func() []TestUnit) {
	c := MustNewClient(NewConf(WithResp(r), WithDevelopment(false)))
	t.Cleanup(func() {
		_ = c.Close()
	})
	var ctx = context.Background()
	for _, v := range unitsFunc() {
		Convey(v.Name.String(), t, func() { doTestUnitClean(ctx, c, v.Func(ctx, c)) })
	}
}

func TestResp2Client_BitMap(t *testing.T) { doTestUnits(t, RESP2, bitMapTestUnits) }
func TestResp3Client_BitMap(t *testing.T) { doTestUnits(t, RESP3, bitMapTestUnits) }
