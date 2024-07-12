package redisson

import (
	"context"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"strconv"
	"testing"
	"time"
)

func zWithKeyEqual(a, b *ZWithKey) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil && b != nil {
		return false
	}
	if b == nil && a != nil {
		return false
	}
	return a.Key == b.Key && a.Score == b.Score && a.Member == b.Member
}

func zEqual(a, b Z) bool {
	return a.Score == b.Score && a.Member == b.Member
}

func zsEqual(a, b []Z) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil && b != nil {
		return false
	}
	if b == nil && a != nil {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if !zEqual(b[k], v) {
			return false
		}
	}
	return true
}

func testBZPopMax(ctx context.Context, c Cmdable) []string {
	var key1, key2, value1, value2, value3 = "zset1", "zset2", "one", "two", "three"
	err := c.ZAdd(ctx, key1, Z{
		Score:  1,
		Member: value1,
	}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key1, Z{
		Score:  2,
		Member: value2,
	}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key1, Z{
		Score:  3,
		Member: value3,
	}).Err()
	So(err, ShouldBeNil)

	bZPopMax := c.BZPopMax(ctx, 0, key1, key2)
	So(bZPopMax.Err(), ShouldBeNil)
	v := bZPopMax.Val()
	So(zWithKeyEqual(&v, &ZWithKey{
		Z: Z{
			Score:  3,
			Member: value3,
		},
		Key: key1,
	}), ShouldBeTrue)

	d := c.Del(ctx, key1, key2)
	So(d.Err(), ShouldBeNil)

	val := c.BZPopMax(ctx, time.Second, key1)
	So(val.Err(), ShouldNotBeNil)
	So(IsNil(val.Err()), ShouldBeTrue)
	So(val.Val(), ShouldBeNil)

	return []string{key1, key2}
}

func testBZPopMin(ctx context.Context, c Cmdable) []string {
	var key1, key2, value1, value2, value3 = "zset1", "zset2", "one", "two", "three"
	err := c.ZAdd(ctx, key1, Z{
		Score:  1,
		Member: value1,
	}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key1, Z{
		Score:  2,
		Member: value2,
	}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key1, Z{
		Score:  3,
		Member: value3,
	}).Err()
	So(err, ShouldBeNil)

	bZPopMax := c.BZPopMin(ctx, 0, key1, key2)
	So(bZPopMax.Err(), ShouldBeNil)
	v := bZPopMax.Val()
	So(zWithKeyEqual(&v, &ZWithKey{
		Z: Z{
			Score:  1,
			Member: value1,
		},
		Key: key1,
	}), ShouldBeTrue)

	d := c.Del(ctx, key1, key2)
	So(d.Err(), ShouldBeNil)

	val := c.BZPopMin(ctx, time.Second, key1)
	So(val.Err(), ShouldNotBeNil)
	So(IsNil(val.Err()), ShouldBeTrue)
	So(val.Val(), ShouldBeNil)

	return []string{key1, key2}
}

func testZAdd(ctx context.Context, c Cmdable) []string {
	var key = "zset"
	added := c.ZAdd(ctx, key, Z{
		Score:  1,
		Member: "one",
	})
	So(added.Err(), ShouldBeNil)
	So(added.Val(), ShouldEqual, 1)

	added = c.ZAdd(ctx, key, Z{
		Score:  1,
		Member: "uno",
	})
	So(added.Err(), ShouldBeNil)
	So(added.Val(), ShouldEqual, 1)

	added = c.ZAdd(ctx, key, Z{
		Score:  2,
		Member: "two",
	})
	So(added.Err(), ShouldBeNil)
	So(added.Val(), ShouldEqual, 1)

	added = c.ZAdd(ctx, key, Z{
		Score:  3,
		Member: "two",
	})
	So(added.Err(), ShouldBeNil)
	So(added.Val(), ShouldEqual, 0)

	vals := c.ZRangeWithScores(ctx, key, 0, -1)
	So(vals.Err(), ShouldBeNil)
	So(zsEqual(vals.Val(), []Z{{
		Score:  1,
		Member: "one",
	}, {
		Score:  1,
		Member: "uno",
	}, {
		Score:  3,
		Member: "two",
	}}), ShouldBeTrue)

	d := c.Del(ctx, key)
	So(d.Err(), ShouldBeNil)

	added = c.ZAdd(ctx, key, Z{
		Score:  1,
		Member: "one",
	})
	So(added.Err(), ShouldBeNil)
	So(added.Val(), ShouldEqual, 1)

	added = c.ZAdd(ctx, key, Z{
		Score:  1,
		Member: "uno",
	})
	So(added.Err(), ShouldBeNil)
	So(added.Val(), ShouldEqual, 1)

	added = c.ZAdd(ctx, key, Z{
		Score:  2,
		Member: "two",
	})
	So(added.Err(), ShouldBeNil)
	So(added.Val(), ShouldEqual, 1)

	added = c.ZAdd(ctx, key, Z{
		Score:  3,
		Member: "two",
	})
	So(added.Err(), ShouldBeNil)
	So(added.Val(), ShouldEqual, 0)

	vals = c.ZRangeWithScores(ctx, key, 0, -1)
	So(vals.Err(), ShouldBeNil)
	So(zsEqual(vals.Val(), []Z{{
		Score:  1,
		Member: "one",
	}, {
		Score:  1,
		Member: "uno",
	}, {
		Score:  3,
		Member: "two",
	}}), ShouldBeTrue)

	return []string{key}
}

func testZAddNX(ctx context.Context, c Cmdable) []string {
	var key = "zset"

	added := c.ZAddNX(ctx, key, Z{
		Score:  1,
		Member: "one",
	})
	So(added.Err(), ShouldBeNil)
	So(added.Val(), ShouldEqual, 1)

	vals := c.ZRangeWithScores(ctx, key, 0, -1)
	So(vals.Err(), ShouldBeNil)
	So(zsEqual(vals.Val(), []Z{{
		Score:  1,
		Member: "one",
	}}), ShouldBeTrue)

	added = c.ZAddNX(ctx, key, Z{
		Score:  2,
		Member: "one",
	})
	So(added.Err(), ShouldBeNil)
	So(added.Val(), ShouldEqual, 0)

	vals = c.ZRangeWithScores(ctx, key, 0, -1)
	So(vals.Err(), ShouldBeNil)
	So(zsEqual(vals.Val(), []Z{{
		Score:  1,
		Member: "one",
	}}), ShouldBeTrue)

	return []string{key}
}

func testZAddXX(ctx context.Context, c Cmdable) []string {
	var key = "zset"

	added := c.ZAddXX(ctx, key, Z{
		Score:  1,
		Member: "one",
	})
	So(added.Err(), ShouldBeNil)
	So(added.Val(), ShouldEqual, 0)

	vals := c.ZRangeWithScores(ctx, key, 0, -1)
	So(vals.Err(), ShouldBeNil)
	So(vals.Val(), ShouldBeEmpty)

	added = c.ZAdd(ctx, key, Z{
		Score:  1,
		Member: "one",
	})
	So(added.Err(), ShouldBeNil)
	So(added.Val(), ShouldEqual, 1)

	added = c.ZAddXX(ctx, key, Z{
		Score:  2,
		Member: "one",
	})
	So(added.Err(), ShouldBeNil)
	So(added.Val(), ShouldEqual, 0)

	vals = c.ZRangeWithScores(ctx, key, 0, -1)
	So(vals.Err(), ShouldBeNil)
	So(zsEqual(vals.Val(), []Z{{
		Score:  2,
		Member: "one",
	}}), ShouldBeTrue)

	return []string{key}
}

func testZAddCh(ctx context.Context, c Cmdable) []string {
	var key = "zset"

	changed := c.ZAddCh(ctx, key, Z{
		Score:  1,
		Member: "one",
	})
	So(changed.Err(), ShouldBeNil)
	So(changed.Val(), ShouldEqual, 1)

	changed = c.ZAddCh(ctx, key, Z{
		Score:  1,
		Member: "one",
	})
	So(changed.Err(), ShouldBeNil)
	So(changed.Val(), ShouldEqual, 0)

	return []string{key}
}

func testZAddNXCh(ctx context.Context, c Cmdable) []string {
	var key = "zset"

	changed := c.ZAddNXCh(ctx, key, Z{
		Score:  1,
		Member: "one",
	})
	So(changed.Err(), ShouldBeNil)
	So(changed.Val(), ShouldEqual, 1)

	vals := c.ZRangeWithScores(ctx, key, 0, -1)
	So(vals.Err(), ShouldBeNil)
	So(zsEqual(vals.Val(), []Z{{
		Score:  1,
		Member: "one",
	}}), ShouldBeTrue)

	changed = c.ZAddNXCh(ctx, key, Z{
		Score:  2,
		Member: "one",
	})
	So(changed.Err(), ShouldBeNil)
	So(changed.Val(), ShouldEqual, 0)

	vals = c.ZRangeWithScores(ctx, key, 0, -1)
	So(vals.Err(), ShouldBeNil)
	So(zsEqual(vals.Val(), []Z{{
		Score:  1,
		Member: "one",
	}}), ShouldBeTrue)

	return []string{key}
}

func testZAddXXCh(ctx context.Context, c Cmdable) []string {
	var key = "zset"

	changed := c.ZAddXXCh(ctx, key, Z{
		Score:  1,
		Member: "one",
	})
	So(changed.Err(), ShouldBeNil)
	So(changed.Val(), ShouldEqual, 0)

	vals := c.ZRangeWithScores(ctx, key, 0, -1)
	So(vals.Err(), ShouldBeNil)
	So(vals.Val(), ShouldBeEmpty)

	added := c.ZAdd(ctx, key, Z{
		Score:  1,
		Member: "one",
	})
	So(added.Err(), ShouldBeNil)
	So(added.Val(), ShouldEqual, 1)

	changed = c.ZAddXXCh(ctx, key, Z{
		Score:  2,
		Member: "one",
	})
	So(changed.Err(), ShouldBeNil)
	So(changed.Val(), ShouldEqual, 1)

	vals = c.ZRangeWithScores(ctx, key, 0, -1)
	So(vals.Err(), ShouldBeNil)
	So(zsEqual(vals.Val(), []Z{{
		Score:  2,
		Member: "one",
	}}), ShouldBeTrue)

	return []string{key}
}

func testZAddArgs(ctx context.Context, c Cmdable) []string {
	var key = "zset"

	// Test only the GT+LT options.
	added := c.ZAddArgs(ctx, key, ZAddArgs{
		GT:      true,
		Members: []Z{{Score: 1, Member: "one"}},
	})
	So(added.Err(), ShouldBeNil)
	So(added.Val(), ShouldEqual, 1)

	vals := c.ZRangeWithScores(ctx, key, 0, -1)
	So(vals.Err(), ShouldBeNil)
	So(zsEqual(vals.Val(), []Z{{
		Score:  1,
		Member: "one",
	}}), ShouldBeTrue)

	added = c.ZAddArgs(ctx, key, ZAddArgs{
		GT:      true,
		Members: []Z{{Score: 2, Member: "one"}},
	})
	So(added.Err(), ShouldBeNil)
	So(added.Val(), ShouldEqual, 0)

	vals = c.ZRangeWithScores(ctx, key, 0, -1)
	So(vals.Err(), ShouldBeNil)
	So(zsEqual(vals.Val(), []Z{{
		Score:  2,
		Member: "one",
	}}), ShouldBeTrue)

	added = c.ZAddArgs(ctx, key, ZAddArgs{
		LT:      true,
		Members: []Z{{Score: 1, Member: "one"}},
	})
	So(added.Err(), ShouldBeNil)
	So(added.Val(), ShouldEqual, 0)

	vals = c.ZRangeWithScores(ctx, key, 0, -1)
	So(vals.Err(), ShouldBeNil)
	So(zsEqual(vals.Val(), []Z{{
		Score:  1,
		Member: "one",
	}}), ShouldBeTrue)

	return []string{key}
}

func testZAddArgsIncr(ctx context.Context, c Cmdable) []string {
	var key = "zset"

	added := c.ZAddArgs(ctx, key, ZAddArgs{
		Members: []Z{{Score: 1, Member: "one"}},
	})
	So(added.Err(), ShouldBeNil)
	So(added.Val(), ShouldEqual, 1)

	zAddArgsIncr := c.ZAddArgsIncr(ctx, key, ZAddArgs{
		Members: []Z{{Score: 1, Member: "one"}},
	})
	So(zAddArgsIncr.Err(), ShouldBeNil)

	vals := c.ZRangeWithScores(ctx, key, 0, -1)
	So(vals.Err(), ShouldBeNil)
	So(zsEqual(vals.Val(), []Z{{
		Score:  2,
		Member: "one",
	}}), ShouldBeTrue)

	return []string{key}
}

func testZIncr(ctx context.Context, c Cmdable) []string {
	var key = "zset"

	score := c.ZIncr(ctx, key, Z{
		Score:  1,
		Member: "one",
	})
	So(score.Err(), ShouldBeNil)
	So(score.Val(), ShouldEqual, 1)

	vals := c.ZRangeWithScores(ctx, key, 0, -1)
	So(vals.Err(), ShouldBeNil)
	So(zsEqual(vals.Val(), []Z{{
		Score:  1,
		Member: "one",
	}}), ShouldBeTrue)

	score = c.ZIncr(ctx, key, Z{
		Score:  1,
		Member: "one",
	})
	So(score.Err(), ShouldBeNil)
	So(score.Val(), ShouldEqual, 2)

	vals = c.ZRangeWithScores(ctx, key, 0, -1)
	So(vals.Err(), ShouldBeNil)
	So(zsEqual(vals.Val(), []Z{{
		Score:  2,
		Member: "one",
	}}), ShouldBeTrue)

	return []string{key}
}

func testZIncrNX(ctx context.Context, c Cmdable) []string {
	var key = "zset"

	score := c.ZIncrNX(ctx, key, Z{
		Score:  1,
		Member: "one",
	})
	So(score.Err(), ShouldBeNil)
	So(score.Val(), ShouldEqual, 1)

	vals := c.ZRangeWithScores(ctx, key, 0, -1)
	So(vals.Err(), ShouldBeNil)
	So(zsEqual(vals.Val(), []Z{{
		Score:  1,
		Member: "one",
	}}), ShouldBeTrue)

	score = c.ZIncrNX(ctx, key, Z{
		Score:  1,
		Member: "one",
	})
	So(score.Err(), ShouldNotBeNil)
	So(IsNil(score.Err()), ShouldBeTrue)
	So(score.Val(), ShouldEqual, 0)

	vals = c.ZRangeWithScores(ctx, key, 0, -1)
	So(vals.Err(), ShouldBeNil)
	So(zsEqual(vals.Val(), []Z{{
		Score:  1,
		Member: "one",
	}}), ShouldBeTrue)

	return []string{key}
}

func testZIncrXX(ctx context.Context, c Cmdable) []string {
	var key = "zset"

	score := c.ZIncrXX(ctx, key, Z{
		Score:  1,
		Member: "one",
	})
	So(score.Err(), ShouldNotBeNil)
	So(IsNil(score.Err()), ShouldBeTrue)
	So(score.Val(), ShouldEqual, 0)

	vals := c.ZRangeWithScores(ctx, key, 0, -1)
	So(vals.Err(), ShouldBeNil)
	So(vals.Val(), ShouldBeEmpty)

	added := c.ZAdd(ctx, key, Z{
		Score:  1,
		Member: "one",
	})
	So(added.Err(), ShouldBeNil)
	So(added.Val(), ShouldEqual, 1)

	score = c.ZIncrXX(ctx, key, Z{
		Score:  1,
		Member: "one",
	})
	So(score.Err(), ShouldBeNil)
	So(score.Val(), ShouldEqual, 2)

	vals = c.ZRangeWithScores(ctx, key, 0, -1)
	So(vals.Err(), ShouldBeNil)
	So(zsEqual(vals.Val(), []Z{{
		Score:  2,
		Member: "one",
	}}), ShouldBeTrue)

	return []string{key}
}

func testZDiffStore(ctx context.Context, c Cmdable) []string {
	var key1, key2, dest = "zset1", "zset2", "out1"

	err := c.ZAdd(ctx, key1, Z{Score: 1, Member: "one"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key1, Z{Score: 2, Member: "two"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key2, Z{Score: 1, Member: "one"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key2, Z{Score: 2, Member: "two"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key2, Z{Score: 3, Member: "three"}).Err()
	So(err, ShouldBeNil)
	v := c.ZDiffStore(ctx, dest, key1, key2)
	So(v.Err(), ShouldBeNil)
	So(v.Val(), ShouldEqual, 0)
	v = c.ZDiffStore(ctx, dest, key2, key1)
	So(v.Err(), ShouldBeNil)
	So(v.Val(), ShouldEqual, 1)
	vals := c.ZRangeWithScores(ctx, dest, 0, -1)
	So(vals.Err(), ShouldBeNil)
	So(zsEqual(vals.Val(), []Z{{
		Score:  3,
		Member: "three",
	}}), ShouldBeTrue)

	return []string{key1, key2, dest}
}

func testZIncrBy(ctx context.Context, c Cmdable) []string {
	var key = "zset"

	err := c.ZAdd(ctx, key, Z{
		Score:  1,
		Member: "one",
	}).Err()
	So(err, ShouldBeNil)

	err = c.ZAdd(ctx, key, Z{
		Score:  2,
		Member: "two",
	}).Err()
	So(err, ShouldBeNil)

	n := c.ZIncrBy(ctx, key, 2, "one")
	So(n.Err(), ShouldBeNil)
	So(n.Val(), ShouldEqual, 3)

	val := c.ZRangeWithScores(ctx, key, 0, -1)
	So(val.Err(), ShouldBeNil)
	So(zsEqual(val.Val(), []Z{{
		Score:  2,
		Member: "two",
	}, {
		Score:  3,
		Member: "one",
	}}), ShouldBeTrue)

	return []string{key}
}

func testZInterStore(ctx context.Context, c Cmdable) []string {
	var key1, key2, key3, dest = "zset1", "zset2", "zset3", "out1"

	err := c.ZAdd(ctx, key1, Z{
		Score:  1,
		Member: "one",
	}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key1, Z{
		Score:  2,
		Member: "two",
	}).Err()
	So(err, ShouldBeNil)

	err = c.ZAdd(ctx, key2, Z{Score: 1, Member: "one"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key2, Z{Score: 2, Member: "two"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key3, Z{Score: 3, Member: "two"}).Err()
	So(err, ShouldBeNil)

	n := c.ZInterStore(ctx, dest, ZStore{
		Keys:    []string{key1, key2},
		Weights: []int64{2, 3},
	})
	So(n.Err(), ShouldBeNil)
	So(n.Val(), ShouldEqual, 2)

	vals := c.ZRangeWithScores(ctx, dest, 0, -1)
	So(vals.Err(), ShouldBeNil)
	So(zsEqual(vals.Val(), []Z{{
		Score:  5,
		Member: "one",
	}, {
		Score:  10,
		Member: "two",
	}}), ShouldBeTrue)

	return []string{key1, key2, key3, dest}
}

func testZPopMax(ctx context.Context, c Cmdable) []string {
	var key = "zset"

	err := c.ZAdd(ctx, key, Z{
		Score:  1,
		Member: "one",
	}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{
		Score:  2,
		Member: "two",
	}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{
		Score:  3,
		Member: "three",
	}).Err()
	So(err, ShouldBeNil)

	members := c.ZPopMax(ctx, key)
	So(members.Err(), ShouldBeNil)
	So(zsEqual(members.Val(), []Z{{
		Score:  3,
		Member: "three",
	}}), ShouldBeTrue)

	// adding back 3
	err = c.ZAdd(ctx, key, Z{
		Score:  3,
		Member: "three",
	}).Err()
	So(err, ShouldBeNil)
	members = c.ZPopMax(ctx, key, 2)
	So(members.Err(), ShouldBeNil)
	So(zsEqual(members.Val(), []Z{{
		Score:  3,
		Member: "three",
	}, {
		Score:  2,
		Member: "two",
	}}), ShouldBeTrue)

	// adding back 2 & 3
	err = c.ZAdd(ctx, key, Z{
		Score:  3,
		Member: "three",
	}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{
		Score:  2,
		Member: "two",
	}).Err()
	So(err, ShouldBeNil)
	members = c.ZPopMax(ctx, key, 10)
	So(members.Err(), ShouldBeNil)
	So(zsEqual(members.Val(), []Z{{
		Score:  3,
		Member: "three",
	}, {
		Score:  2,
		Member: "two",
	}, {
		Score:  1,
		Member: "one",
	}}), ShouldBeTrue)

	return []string{key}
}

func testZPopMin(ctx context.Context, c Cmdable) []string {
	var key = "zset"

	err := c.ZAdd(ctx, key, Z{
		Score:  1,
		Member: "one",
	}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{
		Score:  2,
		Member: "two",
	}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{
		Score:  3,
		Member: "three",
	}).Err()
	So(err, ShouldBeNil)

	members := c.ZPopMin(ctx, key)
	So(members.Err(), ShouldBeNil)
	So(zsEqual(members.Val(), []Z{{
		Score:  1,
		Member: "one",
	}}), ShouldBeTrue)

	// adding back 1
	err = c.ZAdd(ctx, key, Z{
		Score:  1,
		Member: "one",
	}).Err()
	So(err, ShouldBeNil)
	members = c.ZPopMin(ctx, key, 2)
	So(members.Err(), ShouldBeNil)
	So(zsEqual(members.Val(), []Z{{
		Score:  1,
		Member: "one",
	}, {
		Score:  2,
		Member: "two",
	}}), ShouldBeTrue)

	// adding back 1 & 2
	err = c.ZAdd(ctx, key, Z{
		Score:  1,
		Member: "one",
	}).Err()
	So(err, ShouldBeNil)

	err = c.ZAdd(ctx, key, Z{
		Score:  2,
		Member: "two",
	}).Err()
	So(err, ShouldBeNil)

	members = c.ZPopMin(ctx, key, 10)
	So(members.Err(), ShouldBeNil)
	So(zsEqual(members.Val(), []Z{{
		Score:  1,
		Member: "one",
	}, {
		Score:  2,
		Member: "two",
	}, {
		Score:  3,
		Member: "three",
	}}), ShouldBeTrue)

	return []string{key}
}

func testZRangeStore(ctx context.Context, c Cmdable) []string {
	var key1, key2 = "zset", "new-zset"

	added := c.ZAddArgs(ctx, key1, ZAddArgs{
		Members: []Z{
			{Score: 1, Member: "one"},
			{Score: 2, Member: "two"},
			{Score: 3, Member: "three"},
			{Score: 4, Member: "four"},
		},
	})
	So(added.Err(), ShouldBeNil)
	So(added.Val(), ShouldEqual, 4)

	rangeStore := c.ZRangeStore(ctx, key2, ZRangeArgs{
		Key:     key1,
		Start:   1,
		Stop:    4,
		ByScore: true,
		Rev:     true,
		Offset:  1,
		Count:   2,
	})
	So(rangeStore.Err(), ShouldBeNil)
	So(rangeStore.Val(), ShouldEqual, 2)

	zRange := c.ZRange(ctx, key2, 0, -1)
	So(zRange.Err(), ShouldBeNil)
	So(stringSliceEqual(zRange.Val(), []string{"two", "three"}, true), ShouldBeTrue)

	return []string{key1, key2}
}

func testZRem(ctx context.Context, c Cmdable) []string {
	var key = "zset"

	err := c.ZAdd(ctx, key, Z{Score: 1, Member: "one"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{Score: 2, Member: "two"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{Score: 3, Member: "three"}).Err()
	So(err, ShouldBeNil)

	zRem := c.ZRem(ctx, key, "two")
	So(zRem.Err(), ShouldBeNil)
	So(zRem.Val(), ShouldEqual, 1)

	vals := c.ZRangeWithScores(ctx, key, 0, -1)
	So(vals.Err(), ShouldBeNil)
	So(zsEqual(vals.Val(), []Z{{
		Score:  1,
		Member: "one",
	}, {
		Score:  3,
		Member: "three",
	}}), ShouldBeTrue)

	return []string{key}
}

func testZRemRangeByLex(ctx context.Context, c Cmdable) []string {
	var key = "zset"

	zz := []Z{
		{Score: 0, Member: "aaaa"},
		{Score: 0, Member: "b"},
		{Score: 0, Member: "c"},
		{Score: 0, Member: "d"},
		{Score: 0, Member: "e"},
		{Score: 0, Member: "foo"},
		{Score: 0, Member: "zap"},
		{Score: 0, Member: "zip"},
		{Score: 0, Member: "ALPHA"},
		{Score: 0, Member: "alpha"},
	}
	for _, z := range zz {
		err := c.ZAdd(ctx, key, z).Err()
		So(err, ShouldBeNil)
	}

	n := c.ZRemRangeByLex(ctx, key, "[alpha", "[omega")
	So(n.Err(), ShouldBeNil)
	So(n.Val(), ShouldEqual, 6)

	vals := c.ZRange(ctx, key, 0, -1)
	So(vals.Err(), ShouldBeNil)
	So(stringSliceEqual(vals.Val(), []string{"ALPHA", "aaaa", "zap", "zip"}, true), ShouldBeTrue)

	return []string{key}
}

func testZRemRangeByRank(ctx context.Context, c Cmdable) []string {
	var key = "zset"

	err := c.ZAdd(ctx, key, Z{Score: 1, Member: "one"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{Score: 2, Member: "two"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{Score: 3, Member: "three"}).Err()
	So(err, ShouldBeNil)

	zRemRangeByRank := c.ZRemRangeByRank(ctx, key, 0, 1)
	So(zRemRangeByRank.Err(), ShouldBeNil)
	So(zRemRangeByRank.Val(), ShouldEqual, 2)

	vals := c.ZRangeWithScores(ctx, key, 0, -1)
	So(vals.Err(), ShouldBeNil)
	So(zsEqual(vals.Val(), []Z{{
		Score:  3,
		Member: "three",
	}}), ShouldBeTrue)

	return []string{key}
}

func testZRemRangeByScore(ctx context.Context, c Cmdable) []string {
	var key = "zset"

	err := c.ZAdd(ctx, key, Z{Score: 1, Member: "one"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{Score: 2, Member: "two"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{Score: 3, Member: "three"}).Err()
	So(err, ShouldBeNil)

	zRemRangeByScore := c.ZRemRangeByScore(ctx, key, "-inf", "(2")
	So(zRemRangeByScore.Err(), ShouldBeNil)
	So(zRemRangeByScore.Val(), ShouldEqual, 1)

	vals := c.ZRangeWithScores(ctx, key, 0, -1)
	So(vals.Err(), ShouldBeNil)
	So(zsEqual(vals.Val(), []Z{{
		Score:  2,
		Member: "two",
	}, {
		Score:  3,
		Member: "three",
	}}), ShouldBeTrue)

	return []string{key}
}

func testZUnionStore(ctx context.Context, c Cmdable) []string {
	var key1, key2, dest = "zset1", "zset2", "out"

	err := c.ZAdd(ctx, key1, Z{Score: 1, Member: "one"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key1, Z{Score: 2, Member: "two"}).Err()
	So(err, ShouldBeNil)

	err = c.ZAdd(ctx, key2, Z{Score: 1, Member: "one"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key2, Z{Score: 2, Member: "two"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key2, Z{Score: 3, Member: "three"}).Err()
	So(err, ShouldBeNil)

	n := c.ZUnionStore(ctx, dest, ZStore{
		Keys:    []string{key1, key2},
		Weights: []int64{2, 3},
	})
	So(n.Err(), ShouldBeNil)
	So(n.Val(), ShouldEqual, 3)

	val := c.ZRangeWithScores(ctx, dest, 0, -1)
	So(val.Err(), ShouldBeNil)
	So(zsEqual(val.Val(), []Z{{
		Score:  5,
		Member: "one",
	}, {
		Score:  9,
		Member: "three",
	}, {
		Score:  10,
		Member: "two",
	}}), ShouldBeTrue)

	return []string{key1, key2, dest}
}

func testZInter(ctx context.Context, c Cmdable) []string {
	var key1, key2 = "zset1", "zset2"

	err := c.ZAdd(ctx, key1, Z{Score: 1, Member: "one"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key1, Z{Score: 2, Member: "two"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key2, Z{Score: 1, Member: "one"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key2, Z{Score: 2, Member: "two"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key2, Z{Score: 3, Member: "three"}).Err()
	So(err, ShouldBeNil)

	v := c.ZInter(ctx, ZStore{
		Keys: []string{key1, key2},
	})
	So(v.Err(), ShouldBeNil)
	So(stringSliceEqual(v.Val(), []string{"one", "two"}, true), ShouldBeTrue)

	return []string{key1, key2}
}

func testZInterWithScores(ctx context.Context, c Cmdable) []string {
	var key1, key2 = "zset1", "zset2"

	err := c.ZAdd(ctx, key1, Z{Score: 1, Member: "one"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key1, Z{Score: 2, Member: "two"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key2, Z{Score: 1, Member: "one"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key2, Z{Score: 2, Member: "two"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key2, Z{Score: 3, Member: "three"}).Err()
	So(err, ShouldBeNil)

	v := c.ZInterWithScores(ctx, ZStore{
		Keys:      []string{key1, key2},
		Weights:   []int64{2, 3},
		Aggregate: "Max",
	})
	So(v.Err(), ShouldBeNil)
	So(zsEqual(v.Val(), []Z{{
		Member: "one",
		Score:  3,
	}, {
		Member: "two",
		Score:  6,
	}}), ShouldBeTrue)

	return []string{key1, key2}
}

func testZRandMember(ctx context.Context, c Cmdable) []string {
	var key = "key"

	err := c.ZAdd(ctx, key, Z{Score: 1, Member: "one"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{Score: 2, Member: "two"}).Err()
	So(err, ShouldBeNil)

	v := c.ZRandMember(ctx, key, 2)
	So(v.Err(), ShouldBeNil)
	So(stringSliceEqual(v.Val(), []string{"one", "two"}, false), ShouldBeTrue)

	v = c.ZRandMember(ctx, key, 0)
	So(v.Err(), ShouldBeNil)
	So(v.Val(), ShouldBeEmpty)

	var slice []string
	var zs []Z
	zs, err = c.ZRandMemberWithScores(ctx, key, 2).Result()
	So(err, ShouldBeNil)
	for _, _zs := range zs {
		slice = append(slice, _zs.Member, strconv.FormatInt(int64(_zs.Score), 10))
	}
	So(stringSliceEqual(slice, []string{"one", "1", "two", "2"}, false), ShouldBeTrue)

	return []string{key}
}

func testZScan(ctx context.Context, c Cmdable) []string {
	var key = "key"

	for i := 0; i < 1000; i++ {
		err := c.ZAdd(ctx, key, Z{
			Score:  float64(i),
			Member: fmt.Sprintf("member%d", i),
		}).Err()
		So(err, ShouldBeNil)
	}

	keys, cursor, err := c.ZScan(ctx, key, 0, "", 0).Result()
	So(err, ShouldBeNil)
	So(len(keys), ShouldBeGreaterThan, 0)
	So(cursor, ShouldBeGreaterThan, 0)

	return []string{key}
}

func testZDiff(ctx context.Context, c Cmdable) []string {
	var key1, key2 = "zset1", "zset2"

	err := c.ZAdd(ctx, key1, Z{Score: 1, Member: "one"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key1, Z{Score: 2, Member: "two"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key1, Z{Score: 3, Member: "three"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key2, Z{Score: 1, Member: "one"}).Err()
	So(err, ShouldBeNil)

	v := c.ZDiff(ctx, key1, key2)
	So(v.Err(), ShouldBeNil)
	So(stringSliceEqual(v.Val(), []string{"two", "three"}, false), ShouldBeTrue)

	return []string{key1, key2}
}

func testZDiffWithScores(ctx context.Context, c Cmdable) []string {
	var key1, key2 = "zset1", "zset2"

	err := c.ZAdd(ctx, key1, Z{Score: 1, Member: "one"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key1, Z{Score: 2, Member: "two"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key1, Z{Score: 3, Member: "three"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key2, Z{Score: 1, Member: "one"}).Err()
	So(err, ShouldBeNil)

	v := c.ZDiffWithScores(ctx, key1, key2)
	So(v.Err(), ShouldBeNil)
	So(zsEqual(v.Val(), []Z{
		{
			Member: "two",
			Score:  2,
		},
		{
			Member: "three",
			Score:  3,
		},
	}), ShouldBeTrue)

	return []string{key1, key2}
}

func testZUnion(ctx context.Context, c Cmdable) []string {
	var key1, key2 = "zset1", "zset2"
	err := c.ZAddArgs(ctx, key1, ZAddArgs{
		Members: []Z{
			{Score: 1, Member: "one"},
			{Score: 2, Member: "two"},
		},
	}).Err()
	So(err, ShouldBeNil)

	err = c.ZAddArgs(ctx, key2, ZAddArgs{
		Members: []Z{
			{Score: 1, Member: "one"},
			{Score: 2, Member: "two"},
			{Score: 3, Member: "three"},
		},
	}).Err()
	So(err, ShouldBeNil)

	union := c.ZUnion(ctx, ZStore{
		Keys:      []string{key1, key2},
		Weights:   []int64{2, 3},
		Aggregate: "sum",
	})
	So(union.Err(), ShouldBeNil)
	So(stringSliceEqual(union.Val(), []string{"one", "three", "two"}, false), ShouldBeTrue)

	unionScores := c.ZUnionWithScores(ctx, ZStore{
		Keys:      []string{key1, key2},
		Weights:   []int64{2, 3},
		Aggregate: "sum",
	})
	So(unionScores.Err(), ShouldBeNil)
	So(zsEqual(unionScores.Val(), []Z{
		{Score: 5, Member: "one"},
		{Score: 9, Member: "three"},
		{Score: 10, Member: "two"},
	}), ShouldBeTrue)

	return []string{key1, key2}
}

func testZUnionWithScores(ctx context.Context, c Cmdable) []string {
	var key1, key2, dest = "zset1", "zset2", "out"

	err := c.ZAdd(ctx, key1, Z{Score: 1, Member: "one"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key1, Z{Score: 2, Member: "two"}).Err()
	So(err, ShouldBeNil)

	err = c.ZAdd(ctx, key2, Z{Score: 1, Member: "one"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key2, Z{Score: 2, Member: "two"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key2, Z{Score: 3, Member: "three"}).Err()
	So(err, ShouldBeNil)

	n := c.ZUnionStore(ctx, dest, ZStore{
		Keys:    []string{key1, key2},
		Weights: []int64{2, 3},
	})
	So(n.Err(), ShouldBeNil)
	So(n.Val(), ShouldEqual, 3)

	val := c.ZRangeWithScores(ctx, dest, 0, -1)
	So(val.Err(), ShouldBeNil)
	So(zsEqual(val.Val(), []Z{
		{
			Score:  5,
			Member: "one",
		}, {
			Score:  9,
			Member: "three",
		}, {
			Score:  10,
			Member: "two",
		},
	}), ShouldBeTrue)

	return []string{key1, key2, dest}
}

func testZCard(ctx context.Context, c Cmdable) []string {
	var key = "zset"

	err := c.ZAdd(ctx, key, Z{
		Score:  1,
		Member: "one",
	}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{
		Score:  2,
		Member: "two",
	}).Err()
	So(err, ShouldBeNil)

	card := c.ZCard(ctx, key)
	So(card.Err(), ShouldBeNil)
	So(card.Val(), ShouldEqual, 2)

	card = cacheCmd(c).ZCard(ctx, key)
	So(card.Err(), ShouldBeNil)
	So(card.Val(), ShouldEqual, 2)

	return []string{key}
}

func testZCount(ctx context.Context, c Cmdable) []string {
	var key = "zset"

	err := c.ZAdd(ctx, key, Z{
		Score:  1,
		Member: "one",
	}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{
		Score:  2,
		Member: "two",
	}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{
		Score:  3,
		Member: "three",
	}).Err()
	So(err, ShouldBeNil)

	count := c.ZCount(ctx, key, "-inf", "+inf")
	So(count.Err(), ShouldBeNil)
	So(count.Val(), ShouldEqual, 3)

	count = cacheCmd(c).ZCount(ctx, key, "(1", "3")
	So(count.Err(), ShouldBeNil)
	So(count.Val(), ShouldEqual, 2)

	count = c.ZLexCount(ctx, key, "-", "+")
	So(count.Err(), ShouldBeNil)
	So(count.Val(), ShouldEqual, 3)

	return []string{key}
}

func testZLexCount(ctx context.Context, c Cmdable) []string {
	var key = "myzset"

	err := c.ZAdd(ctx, key, Z{
		Score:  0,
		Member: "a",
	}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{
		Score:  0,
		Member: "b",
	}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{
		Score:  0,
		Member: "c",
	}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{
		Score:  0,
		Member: "d",
	}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{
		Score:  0,
		Member: "e",
	}).Err()
	So(err, ShouldBeNil)

	err = c.ZAdd(ctx, key, Z{
		Score:  0,
		Member: "f",
	}, Z{
		Score:  0,
		Member: "g",
	}).Err()
	So(err, ShouldBeNil)

	count := c.ZLexCount(ctx, key, "-", "+")
	So(count.Err(), ShouldBeNil)
	So(count.Val(), ShouldEqual, 7)

	count = cacheCmd(c).ZLexCount(ctx, key, "[b", "[f")
	So(count.Err(), ShouldBeNil)
	So(count.Val(), ShouldEqual, 5)

	return []string{key}
}

func testZMScore(ctx context.Context, c Cmdable) []string {
	var key = "myzset"

	zmScore := c.ZMScore(ctx, key, "one", "three")
	So(zmScore.Err(), ShouldBeNil)
	So(len(zmScore.Val()), ShouldEqual, 2)
	So(zmScore.Val()[0], ShouldEqual, 0)

	err := c.ZAdd(ctx, key, Z{Score: 1, Member: "one"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{Score: 2, Member: "two"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{Score: 3, Member: "three"}).Err()
	So(err, ShouldBeNil)

	zmScore = c.ZMScore(ctx, key, "one", "three")
	So(zmScore.Err(), ShouldBeNil)
	So(len(zmScore.Val()), ShouldEqual, 2)
	So(zmScore.Val()[0], ShouldEqual, 1)

	zmScore = cacheCmd(c).ZMScore(ctx, key, "four")
	So(zmScore.Err(), ShouldBeNil)
	So(len(zmScore.Val()), ShouldEqual, 1)

	zmScore = c.ZMScore(ctx, key, "four", "one")
	So(zmScore.Err(), ShouldBeNil)
	So(len(zmScore.Val()), ShouldEqual, 2)

	return []string{key}
}

func testZRange(ctx context.Context, c Cmdable) []string {
	var key = "zset"

	err := c.ZAdd(ctx, key, Z{Score: 1, Member: "one"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{Score: 2, Member: "two"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{Score: 3, Member: "three"}).Err()
	So(err, ShouldBeNil)

	zRange := c.ZRange(ctx, key, 0, -1)
	So(zRange.Err(), ShouldBeNil)
	So(stringSliceEqual(zRange.Val(), []string{"one", "two", "three"}, true), ShouldBeTrue)

	zRange = cacheCmd(c).ZRange(ctx, key, 2, 3)
	So(zRange.Err(), ShouldBeNil)
	So(stringSliceEqual(zRange.Val(), []string{"three"}, true), ShouldBeTrue)

	zRange = c.ZRange(ctx, key, -2, -1)
	So(zRange.Err(), ShouldBeNil)
	So(stringSliceEqual(zRange.Val(), []string{"two", "three"}, true), ShouldBeTrue)

	return []string{key}
}

func testZRangeWithScores(ctx context.Context, c Cmdable) []string {
	var key = "zset"

	err := c.ZAdd(ctx, key, Z{Score: 1, Member: "one"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{Score: 2, Member: "two"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{Score: 3, Member: "three"}).Err()
	So(err, ShouldBeNil)

	vals := c.ZRangeWithScores(ctx, key, 0, -1)
	So(vals.Err(), ShouldBeNil)
	So(zsEqual(vals.Val(), []Z{{
		Score:  1,
		Member: "one",
	}, {
		Score:  2,
		Member: "two",
	}, {
		Score:  3,
		Member: "three",
	}}), ShouldBeTrue)

	vals = cacheCmd(c).ZRangeWithScores(ctx, key, 2, 3)
	So(vals.Err(), ShouldBeNil)
	So(zsEqual(vals.Val(), []Z{{Score: 3, Member: "three"}}), ShouldBeTrue)

	vals = c.ZRangeWithScores(ctx, key, -2, -1)
	So(vals.Err(), ShouldBeNil)
	So(zsEqual(vals.Val(), []Z{{
		Score:  2,
		Member: "two",
	}, {
		Score:  3,
		Member: "three",
	}}), ShouldBeTrue)

	return []string{key}
}

func testZRangeArgs(ctx context.Context, c Cmdable) []string {
	var key = "zset"

	added := c.ZAddArgs(ctx, key, ZAddArgs{
		Members: []Z{
			{Score: 1, Member: "one"},
			{Score: 2, Member: "two"},
			{Score: 3, Member: "three"},
			{Score: 4, Member: "four"},
		},
	})
	So(added.Err(), ShouldBeNil)
	So(added.Val(), ShouldEqual, 4)

	zRange := c.ZRangeArgs(ctx, ZRangeArgs{
		Key:     key,
		Start:   1,
		Stop:    4,
		ByScore: true,
		Rev:     true,
		Offset:  1,
		Count:   2,
	})
	So(zRange.Err(), ShouldBeNil)
	So(stringSliceEqual(zRange.Val(), []string{"three", "two"}, true), ShouldBeTrue)

	zRange = cacheCmd(c).ZRangeArgs(ctx, ZRangeArgs{
		Key:    key,
		Start:  "-",
		Stop:   "+",
		ByLex:  true,
		Rev:    true,
		Offset: 2,
		Count:  2,
	})
	So(zRange.Err(), ShouldBeNil)
	So(stringSliceEqual(zRange.Val(), []string{"two", "one"}, true), ShouldBeTrue)

	zRange = c.ZRangeArgs(ctx, ZRangeArgs{
		Key:     key,
		Start:   "(1",
		Stop:    "(4",
		ByScore: true,
	})
	So(zRange.Err(), ShouldBeNil)
	So(stringSliceEqual(zRange.Val(), []string{"two", "three"}, true), ShouldBeTrue)

	// withScores.
	zSlice := c.ZRangeArgsWithScores(ctx, ZRangeArgs{
		Key:     key,
		Start:   1,
		Stop:    4,
		ByScore: true,
		Rev:     true,
		Offset:  1,
		Count:   2,
	})
	So(zSlice.Err(), ShouldBeNil)
	So(zsEqual(zSlice.Val(), []Z{
		{Score: 3, Member: "three"},
		{Score: 2, Member: "two"},
	}), ShouldBeTrue)

	return []string{key}
}

func testZRangeArgsWithScores(ctx context.Context, c Cmdable) []string {
	var key = "zset"

	added := c.ZAddArgs(ctx, key, ZAddArgs{
		Members: []Z{
			{Score: 1, Member: "one"},
			{Score: 2, Member: "two"},
			{Score: 3, Member: "three"},
			{Score: 4, Member: "four"},
		},
	})
	So(added.Err(), ShouldBeNil)
	So(added.Val(), ShouldEqual, 4)

	zSlice := c.ZRangeArgsWithScores(ctx, ZRangeArgs{
		Key:     key,
		Start:   1,
		Stop:    4,
		ByScore: true,
		Rev:     true,
		Offset:  1,
		Count:   2,
	})
	So(zSlice.Err(), ShouldBeNil)
	So(zsEqual(zSlice.Val(), []Z{
		{Score: 3, Member: "three"},
		{Score: 2, Member: "two"},
	}), ShouldBeTrue)

	zSlice = cacheCmd(c).ZRangeArgsWithScores(ctx, ZRangeArgs{
		Key:     key,
		Start:   1,
		Stop:    4,
		ByScore: true,
		Rev:     true,
		Offset:  1,
		Count:   2,
	})
	So(zSlice.Err(), ShouldBeNil)
	So(zsEqual(zSlice.Val(), []Z{
		{Score: 3, Member: "three"},
		{Score: 2, Member: "two"},
	}), ShouldBeTrue)

	return []string{key}
}

func testZRangeByLex(ctx context.Context, c Cmdable) []string {
	var key = "zset"

	err := c.ZAdd(ctx, key, Z{
		Score:  0,
		Member: "a",
	}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{
		Score:  0,
		Member: "b",
	}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{
		Score:  0,
		Member: "c",
	}).Err()
	So(err, ShouldBeNil)

	zRangeByLex := c.ZRangeByLex(ctx, key, ZRangeBy{
		Min: "-",
		Max: "+",
	})
	So(zRangeByLex.Err(), ShouldBeNil)
	So(stringSliceEqual(zRangeByLex.Val(), []string{"a", "b", "c"}, true), ShouldBeTrue)

	zRangeByLex = c.ZRangeByLex(ctx, key, ZRangeBy{
		Min: "[a",
		Max: "[b",
	})
	So(zRangeByLex.Err(), ShouldBeNil)
	So(stringSliceEqual(zRangeByLex.Val(), []string{"a", "b"}, true), ShouldBeTrue)

	zRangeByLex = cacheCmd(c).ZRangeByLex(ctx, key, ZRangeBy{
		Min: "(a",
		Max: "[b",
	})
	So(zRangeByLex.Err(), ShouldBeNil)
	So(stringSliceEqual(zRangeByLex.Val(), []string{"b"}, true), ShouldBeTrue)

	zRangeByLex = c.ZRangeByLex(ctx, key, ZRangeBy{
		Min: "(a",
		Max: "(b",
	})
	So(zRangeByLex.Err(), ShouldBeNil)
	So(stringSliceEqual(zRangeByLex.Val(), []string{}, true), ShouldBeTrue)

	return []string{key}
}

func testZRangeByScore(ctx context.Context, c Cmdable) []string {
	var key = "zset"

	err := c.ZAdd(ctx, key, Z{Score: 1, Member: "one"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{Score: 2, Member: "two"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{Score: 3, Member: "three"}).Err()
	So(err, ShouldBeNil)

	zRangeByScore := c.ZRangeByScore(ctx, key, ZRangeBy{
		Min: "-inf",
		Max: "+inf",
	})
	So(zRangeByScore.Err(), ShouldBeNil)
	So(stringSliceEqual(zRangeByScore.Val(), []string{"one", "two", "three"}, true), ShouldBeTrue)

	zRangeByScore = cacheCmd(c).ZRangeByScore(ctx, key, ZRangeBy{
		Min: "1",
		Max: "2",
	})
	So(zRangeByScore.Err(), ShouldBeNil)
	So(stringSliceEqual(zRangeByScore.Val(), []string{"one", "two"}, true), ShouldBeTrue)

	zRangeByScore = c.ZRangeByScore(ctx, key, ZRangeBy{
		Min: "(1",
		Max: "2",
	})
	So(zRangeByScore.Err(), ShouldBeNil)
	So(stringSliceEqual(zRangeByScore.Val(), []string{"two"}, true), ShouldBeTrue)

	zRangeByScore = c.ZRangeByScore(ctx, key, ZRangeBy{
		Min: "(1",
		Max: "(2",
	})
	So(zRangeByScore.Err(), ShouldBeNil)
	So(stringSliceEqual(zRangeByScore.Val(), []string{}, true), ShouldBeTrue)

	return []string{key}
}

func testZRangeByScoreWithScores(ctx context.Context, c Cmdable) []string {
	var key = "zset"

	err := c.ZAdd(ctx, key, Z{Score: 1, Member: "one"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{Score: 2, Member: "two"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{Score: 3, Member: "three"}).Err()
	So(err, ShouldBeNil)

	vals := c.ZRangeByScoreWithScores(ctx, key, ZRangeBy{
		Min: "-inf",
		Max: "+inf",
	})
	So(vals.Err(), ShouldBeNil)
	So(zsEqual(vals.Val(), []Z{{
		Score:  1,
		Member: "one",
	}, {
		Score:  2,
		Member: "two",
	}, {
		Score:  3,
		Member: "three",
	}}), ShouldBeTrue)

	vals = cacheCmd(c).ZRangeByScoreWithScores(ctx, key, ZRangeBy{
		Min: "1",
		Max: "2",
	})
	So(vals.Err(), ShouldBeNil)
	So(zsEqual(vals.Val(), []Z{{
		Score:  1,
		Member: "one",
	}, {
		Score:  2,
		Member: "two",
	}}), ShouldBeTrue)

	vals = c.ZRangeByScoreWithScores(ctx, key, ZRangeBy{
		Min: "(1",
		Max: "2",
	})
	So(vals.Err(), ShouldBeNil)
	So(zsEqual(vals.Val(), []Z{{Score: 2, Member: "two"}}), ShouldBeTrue)

	vals = c.ZRangeByScoreWithScores(ctx, key, ZRangeBy{
		Min: "(1",
		Max: "(2",
	})
	So(vals.Err(), ShouldBeNil)
	So(zsEqual(vals.Val(), []Z{}), ShouldBeTrue)

	return []string{key}
}

func testZRank(ctx context.Context, c Cmdable) []string {
	var key = "zset"

	err := c.ZAdd(ctx, key, Z{Score: 1, Member: "one"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{Score: 2, Member: "two"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{Score: 3, Member: "three"}).Err()
	So(err, ShouldBeNil)

	zRank := c.ZRank(ctx, key, "three")
	So(zRank.Err(), ShouldBeNil)
	So(zRank.Val(), ShouldEqual, 2)

	zRank = cacheCmd(c).ZRank(ctx, key, "four")
	So(zRank.Err(), ShouldNotBeNil)
	So(IsNil(zRank.Err()), ShouldBeTrue)
	So(zRank.Val(), ShouldEqual, 0)

	return []string{key}
}

func testZRevRange(ctx context.Context, c Cmdable) []string {
	var key = "zset"

	err := c.ZAdd(ctx, key, Z{Score: 1, Member: "one"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{Score: 2, Member: "two"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{Score: 3, Member: "three"}).Err()
	So(err, ShouldBeNil)

	zRevRange := c.ZRevRange(ctx, key, 0, -1)
	So(zRevRange.Err(), ShouldBeNil)
	So(stringSliceEqual(zRevRange.Val(), []string{"three", "two", "one"}, true), ShouldBeTrue)

	zRevRange = cacheCmd(c).ZRevRange(ctx, key, 2, 3)
	So(zRevRange.Err(), ShouldBeNil)
	So(stringSliceEqual(zRevRange.Val(), []string{"one"}, true), ShouldBeTrue)

	zRevRange = c.ZRevRange(ctx, key, -2, -1)
	So(zRevRange.Err(), ShouldBeNil)
	So(stringSliceEqual(zRevRange.Val(), []string{"two", "one"}, true), ShouldBeTrue)

	return []string{key}
}

func testZRevRangeWithScores(ctx context.Context, c Cmdable) []string {
	var key = "zset"

	err := c.ZAdd(ctx, key, Z{Score: 1, Member: "one"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{Score: 2, Member: "two"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{Score: 3, Member: "three"}).Err()
	So(err, ShouldBeNil)

	val := c.ZRevRangeWithScores(ctx, key, 0, -1)
	So(val.Err(), ShouldBeNil)
	So(zsEqual(val.Val(), []Z{{
		Score:  3,
		Member: "three",
	}, {
		Score:  2,
		Member: "two",
	}, {
		Score:  1,
		Member: "one",
	}}), ShouldBeTrue)

	val = cacheCmd(c).ZRevRangeWithScores(ctx, key, 2, 3)
	So(val.Err(), ShouldBeNil)
	So(zsEqual(val.Val(), []Z{{Score: 1, Member: "one"}}), ShouldBeTrue)

	val = c.ZRevRangeWithScores(ctx, key, -2, -1)
	So(val.Err(), ShouldBeNil)
	So(zsEqual(val.Val(), []Z{{
		Score:  2,
		Member: "two",
	}, {
		Score:  1,
		Member: "one",
	}}), ShouldBeTrue)

	return []string{key}
}

func testZRevRangeByLex(ctx context.Context, c Cmdable) []string {
	var key = "zset"

	err := c.ZAdd(ctx, key, Z{Score: 0, Member: "a"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{Score: 0, Member: "b"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{Score: 0, Member: "c"}).Err()
	So(err, ShouldBeNil)

	vals := c.ZRevRangeByLex(ctx, key, ZRangeBy{Max: "+", Min: "-"})
	So(vals.Err(), ShouldBeNil)
	So(stringSliceEqual(vals.Val(), []string{"c", "b", "a"}, true), ShouldBeTrue)

	vals = cacheCmd(c).ZRevRangeByLex(ctx, key, ZRangeBy{Max: "[b", Min: "(a"})
	So(vals.Err(), ShouldBeNil)
	So(stringSliceEqual(vals.Val(), []string{"b"}, true), ShouldBeTrue)

	vals = c.ZRevRangeByLex(ctx, key, ZRangeBy{Max: "(b", Min: "(a"})
	So(vals.Err(), ShouldBeNil)
	So(stringSliceEqual(vals.Val(), []string{}, true), ShouldBeTrue)

	return []string{key}
}

func testZRevRangeByScore(ctx context.Context, c Cmdable) []string {
	var key = "zset"

	err := c.ZAdd(ctx, key, Z{Score: 1, Member: "one"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{Score: 2, Member: "two"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{Score: 3, Member: "three"}).Err()
	So(err, ShouldBeNil)

	vals := cacheCmd(c).ZRevRangeByScore(ctx, key, ZRangeBy{Max: "+inf", Min: "-inf"})
	So(vals.Err(), ShouldBeNil)
	So(stringSliceEqual(vals.Val(), []string{"three", "two", "one"}, true), ShouldBeTrue)

	vals = c.ZRevRangeByScore(ctx, key, ZRangeBy{Max: "2", Min: "(1"})
	So(vals.Err(), ShouldBeNil)
	So(stringSliceEqual(vals.Val(), []string{"two"}, true), ShouldBeTrue)

	vals = c.ZRevRangeByScore(ctx, key, ZRangeBy{Max: "(2", Min: "(1"})
	So(vals.Err(), ShouldBeNil)
	So(stringSliceEqual(vals.Val(), []string{}, true), ShouldBeTrue)

	return []string{key}
}

func testZRevRangeByScoreWithScores(ctx context.Context, c Cmdable) []string {
	var key = "zset"

	err := c.ZAdd(ctx, key, Z{Score: 1, Member: "one"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{Score: 2, Member: "two"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{Score: 3, Member: "three"}).Err()
	So(err, ShouldBeNil)

	vals := c.ZRevRangeByScoreWithScores(ctx, key, ZRangeBy{Max: "+inf", Min: "-inf"})
	So(vals.Err(), ShouldBeNil)
	So(zsEqual(vals.Val(), []Z{{
		Score:  3,
		Member: "three",
	}, {
		Score:  2,
		Member: "two",
	}, {
		Score:  1,
		Member: "one",
	}}), ShouldBeTrue)

	err = c.Del(ctx, key).Err()
	So(err, ShouldBeNil)

	err = c.ZAdd(ctx, key, Z{Score: 1, Member: "one"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{Score: 2, Member: "two"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{Score: 3, Member: "three"}).Err()
	So(err, ShouldBeNil)

	vals = c.ZRevRangeByScoreWithScores(ctx, key, ZRangeBy{Max: "+inf", Min: "-inf"})
	So(vals.Err(), ShouldBeNil)
	So(zsEqual(vals.Val(), []Z{{
		Score:  3,
		Member: "three",
	}, {
		Score:  2,
		Member: "two",
	}, {
		Score:  1,
		Member: "one",
	}}), ShouldBeTrue)

	vals = cacheCmd(c).ZRevRangeByScoreWithScores(ctx, key, ZRangeBy{Max: "2", Min: "(1"})
	So(vals.Err(), ShouldBeNil)
	So(zsEqual(vals.Val(), []Z{{Score: 2, Member: "two"}}), ShouldBeTrue)

	vals = c.ZRevRangeByScoreWithScores(ctx, key, ZRangeBy{Max: "(2", Min: "(1"})
	So(vals.Err(), ShouldBeNil)
	So(zsEqual(vals.Val(), []Z{}), ShouldBeTrue)

	return []string{key}
}

func testZRevRank(ctx context.Context, c Cmdable) []string {
	var key = "zset"

	err := c.ZAdd(ctx, key, Z{Score: 1, Member: "one"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{Score: 2, Member: "two"}).Err()
	So(err, ShouldBeNil)
	err = c.ZAdd(ctx, key, Z{Score: 3, Member: "three"}).Err()
	So(err, ShouldBeNil)

	zRevRank := cacheCmd(c).ZRevRank(ctx, key, "one")
	So(zRevRank.Err(), ShouldBeNil)
	So(zRevRank.Val(), ShouldEqual, 2)

	zRevRank = c.ZRevRank(ctx, key, "four")
	So(zRevRank.Err(), ShouldNotBeNil)
	So(IsNil(zRevRank.Err()), ShouldBeTrue)
	So(zRevRank.Val(), ShouldEqual, 0)

	return []string{key}
}

func testZScore(ctx context.Context, c Cmdable) []string {
	var key = "zset"

	zAdd := c.ZAdd(ctx, key, Z{Score: 1.001, Member: "one"})
	So(zAdd.Err(), ShouldBeNil)

	zScore := c.ZScore(ctx, key, "one")
	So(zScore.Err(), ShouldBeNil)
	So(zScore.Val(), ShouldEqual, 1.001)

	zScore = cacheCmd(c).ZScore(ctx, key, "one")
	So(zScore.Err(), ShouldBeNil)
	So(zScore.Val(), ShouldEqual, 1.001)

	return []string{key}
}

func sortedSetTestUnits() []TestUnit {
	return []TestUnit{
		{CommandBZPopMax, testBZPopMax},
		{CommandBZPopMin, testBZPopMin},
		{CommandZAdd, testZAdd},
		{CommandZAddNX, testZAddNX},
		{CommandZAddXX, testZAddXX},
		{CommandZAddCh, testZAddCh},
		{CommandZAddNX, testZAddNXCh},
		{CommandZAddXX, testZAddXXCh},
		{CommandZAdd, testZAddArgs},
		{CommandZAddIncr, testZAddArgsIncr},
		{CommandZAddIncr, testZIncr},
		{CommandZAddIncr, testZIncrNX},
		{CommandZAddIncr, testZIncrXX},
		{CommandZDiffStore, testZDiffStore},
		{CommandZIncrBy, testZIncrBy},
		{CommandZInterStore, testZInterStore},
		{CommandZPopMax, testZPopMax},
		{CommandZPopMin, testZPopMin},
		{CommandZRangeStore, testZRangeStore},
		{CommandZRem, testZRem},
		{CommandZRemRangeByLex, testZRemRangeByLex},
		{CommandZRemRangeByRank, testZRemRangeByRank},
		{CommandZRemRangeByScore, testZRemRangeByScore},
		{CommandZUnionStore, testZUnionStore},
		{CommandZInter, testZInter},
		{CommandZInter, testZInterWithScores},
		{CommandZRandMember, testZRandMember},
		{CommandZScan, testZScan},
		{CommandZDiff, testZDiff},
		{CommandZDiff, testZDiffWithScores},
		{CommandZUnion, testZUnion},
		{CommandZUnion, testZUnionWithScores},
		{CommandZCard, testZCard},
		{CommandZCount, testZCount},
		{CommandZLexCount, testZLexCount},
		{CommandZMScore, testZMScore},
		{CommandZRange, testZRange},
		{CommandZRange, testZRangeWithScores},
		{CommandZRange, testZRangeArgs},
		{CommandZRange, testZRangeArgsWithScores},
		{CommandZRangeByLex, testZRangeByLex},
		{CommandZRangeByScore, testZRangeByScore},
		{CommandZRangeByScore, testZRangeByScoreWithScores},
		{CommandZRank, testZRank},
		{CommandZRevRange, testZRevRange},
		{CommandZRevRange, testZRevRangeWithScores},
		{CommandZRevRangeByLex, testZRevRangeByLex},
		{CommandZRevRangeByScore, testZRevRangeByScore},
		{CommandZRevRangeByScore, testZRevRangeByScoreWithScores},
		{CommandZRevRank, testZRevRank},
		{CommandZScore, testZScore},
	}
}

func TestResp2Client_SortedSet(t *testing.T) { doTestUnits(t, RESP2, sortedSetTestUnits) }
func TestResp3Client_SortedSet(t *testing.T) { doTestUnits(t, RESP3, sortedSetTestUnits) }
