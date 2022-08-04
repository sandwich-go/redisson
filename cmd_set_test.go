package redisson

import (
	"context"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func testSAdd(ctx context.Context, c Cmdable) []string {
	var key, value1, value2, value3 = "set", "Hello", "World", "!"
	sAdd := c.SAdd(ctx, key, value1, value2)
	So(sAdd.Err(), ShouldBeNil)
	So(sAdd.Val(), ShouldEqual, 2)

	sAdd = c.SAdd(ctx, key, value3)
	So(sAdd.Err(), ShouldBeNil)
	So(sAdd.Val(), ShouldEqual, 1)

	sAdd = c.SAdd(ctx, key, value2)
	So(sAdd.Err(), ShouldBeNil)
	So(sAdd.Val(), ShouldEqual, 0)

	sMembers := c.SMembers(ctx, key)
	So(sMembers.Err(), ShouldBeNil)
	So(stringSliceEqual(sMembers.Val(), []string{value1, value2, value3}, false), ShouldBeTrue)

	return []string{key}
}

func testSCard(ctx context.Context, c Cmdable) []string {
	var key, value1, value2 = "set", "Hello", "World"

	sAdd := c.SAdd(ctx, key, value1)
	So(sAdd.Err(), ShouldBeNil)
	So(sAdd.Val(), ShouldEqual, 1)

	sAdd = c.SAdd(ctx, key, value2)
	So(sAdd.Err(), ShouldBeNil)
	So(sAdd.Val(), ShouldEqual, 1)

	sCard := c.SCard(ctx, key)
	So(sCard.Err(), ShouldBeNil)
	So(sCard.Val(), ShouldEqual, 2)

	sCard = cacheCmd(c).SCard(ctx, key)
	So(sCard.Err(), ShouldBeNil)
	So(sCard.Val(), ShouldEqual, 2)

	return []string{key}
}

func testSDiff(ctx context.Context, c Cmdable) []string {
	var key, key1 = "set1", "set2"

	sAdd := c.SAdd(ctx, key, "a")
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key, "b")
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key, "c")
	So(sAdd.Err(), ShouldBeNil)

	sAdd = c.SAdd(ctx, key1, "c")
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key1, "d")
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key1, "e")
	So(sAdd.Err(), ShouldBeNil)

	sDiff := c.SDiff(ctx, key, key1)
	So(sDiff.Err(), ShouldBeNil)
	So(stringSliceEqual(sDiff.Val(), []string{"a", "b"}, false), ShouldBeTrue)

	return []string{key, key1}
}

func testSDiffStore(ctx context.Context, c Cmdable) []string {
	var key, key1, key2 = "set", "set1", "set2"

	sAdd := c.SAdd(ctx, key1, "a")
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key1, "b")
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key1, "c")
	So(sAdd.Err(), ShouldBeNil)

	sAdd = c.SAdd(ctx, key2, "c")
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key2, "d")
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key2, "e")
	So(sAdd.Err(), ShouldBeNil)

	sDiffStore := c.SDiffStore(ctx, key, key1, key2)
	So(sDiffStore.Err(), ShouldBeNil)
	So(sDiffStore.Val(), ShouldEqual, 2)

	sMembers := c.SMembers(ctx, "set")
	So(sMembers.Err(), ShouldBeNil)
	So(stringSliceEqual(sMembers.Val(), []string{"a", "b"}, false), ShouldBeTrue)

	return []string{key, key1, key2}
}

func testSInter(ctx context.Context, c Cmdable) []string {
	var key1, key2 = "set1", "set2"

	sAdd := c.SAdd(ctx, key1, "a")
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key1, "b")
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key1, "c")
	So(sAdd.Err(), ShouldBeNil)

	sAdd = c.SAdd(ctx, key2, "c")
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key2, "d")
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key2, "e")
	So(sAdd.Err(), ShouldBeNil)

	sInter := c.SInter(ctx, key1, key2)
	So(sInter.Err(), ShouldBeNil)
	So(stringSliceEqual(sInter.Val(), []string{"c"}, false), ShouldBeTrue)

	return []string{key1, key2}
}

func testSInterStore(ctx context.Context, c Cmdable) []string {
	var key, key1, key2 = "set", "set1", "set2"

	sAdd := c.SAdd(ctx, key1, "a")
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key1, "b")
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key1, "c")
	So(sAdd.Err(), ShouldBeNil)

	sAdd = c.SAdd(ctx, key2, "c")
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key2, "d")
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key2, "e")
	So(sAdd.Err(), ShouldBeNil)

	sInterStore := c.SInterStore(ctx, key, key1, key2)
	So(sInterStore.Err(), ShouldBeNil)
	So(sInterStore.Val(), ShouldEqual, 1)

	sMembers := c.SMembers(ctx, key)
	So(sMembers.Err(), ShouldBeNil)
	So(stringSliceEqual(sMembers.Val(), []string{"c"}, false), ShouldBeTrue)

	return []string{key, key1, key2}
}

func testSIsMember(ctx context.Context, c Cmdable) []string {
	var key = "set"

	sAdd := c.SAdd(ctx, key, "one")
	So(sAdd.Err(), ShouldBeNil)

	sIsMember := c.SIsMember(ctx, key, "one")
	So(sIsMember.Err(), ShouldBeNil)
	So(sIsMember.Val(), ShouldBeTrue)

	sIsMember = cacheCmd(c).SIsMember(ctx, key, "one")
	So(sIsMember.Err(), ShouldBeNil)
	So(sIsMember.Val(), ShouldBeTrue)

	sIsMember = c.SIsMember(ctx, key, "two")
	So(sIsMember.Err(), ShouldBeNil)
	So(sIsMember.Val(), ShouldBeFalse)

	sIsMember = cacheCmd(c).SIsMember(ctx, key, "two")
	So(sIsMember.Err(), ShouldBeNil)
	So(sIsMember.Val(), ShouldBeFalse)

	return []string{key}
}

func boolSliceEqual(a, b []bool) bool {
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

func testSMIsMember(ctx context.Context, c Cmdable) []string {
	var key = "set"

	sAdd := c.SAdd(ctx, key, "one")
	So(sAdd.Err(), ShouldBeNil)

	sMIsMember := cacheCmd(c).SMIsMember(ctx, key, "one", "two")
	So(sMIsMember.Err(), ShouldBeNil)
	So(boolSliceEqual(sMIsMember.Val(), []bool{true, false}), ShouldBeTrue)

	sMIsMember = c.SMIsMember(ctx, key, "one", "two")
	So(sMIsMember.Err(), ShouldBeNil)
	So(boolSliceEqual(sMIsMember.Val(), []bool{true, false}), ShouldBeTrue)

	return []string{key}
}

func testSMembers(ctx context.Context, c Cmdable) []string {
	var key, value1, value2 = "set", "Hello", "World"

	sAdd := c.SAdd(ctx, key, value1)
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key, value2)
	So(sAdd.Err(), ShouldBeNil)

	sMembers := c.SMembers(ctx, key)
	So(sMembers.Err(), ShouldBeNil)
	So(stringSliceEqual(sMembers.Val(), []string{value1, value2}, false), ShouldBeTrue)

	sMembers = cacheCmd(c).SMembers(ctx, key)
	So(sMembers.Err(), ShouldBeNil)
	So(stringSliceEqual(sMembers.Val(), []string{value1, value2}, false), ShouldBeTrue)

	return []string{key}
}

func stringStructMapEqual(a, b map[string]struct{}) bool {
	if a == nil && b != nil {
		return false
	}
	if b == nil && a != nil {
		return false
	}
	if len(b) != len(a) {
		return false
	}
	for k := range a {
		if _, ok := b[k]; !ok {
			return false
		}
	}
	return true
}

func testSMembersMap(ctx context.Context, c Cmdable) []string {
	var key, value1, value2 = "set", "Hello", "World"

	sAdd := c.SAdd(ctx, key, value1)
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key, value2)
	So(sAdd.Err(), ShouldBeNil)

	sMembersMap := cacheCmd(c).SMembersMap(ctx, key)
	So(sMembersMap.Err(), ShouldBeNil)
	So(stringStructMapEqual(sMembersMap.Val(), map[string]struct{}{"Hello": {}, "World": {}}), ShouldBeTrue)

	sMembersMap = c.SMembersMap(ctx, key)
	So(sMembersMap.Err(), ShouldBeNil)
	So(stringStructMapEqual(sMembersMap.Val(), map[string]struct{}{"Hello": {}, "World": {}}), ShouldBeTrue)

	return []string{key}
}

func testSMove(ctx context.Context, c Cmdable) []string {
	var key1, key2 = "set1", "set2"

	sAdd := c.SAdd(ctx, key1, "one")
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key1, "two")
	So(sAdd.Err(), ShouldBeNil)

	sAdd = c.SAdd(ctx, key2, "three")
	So(sAdd.Err(), ShouldBeNil)

	sMove := c.SMove(ctx, key1, key2, "two")
	So(sMove.Err(), ShouldBeNil)
	So(sMove.Val(), ShouldBeTrue)

	sMembers := c.SMembers(ctx, key1)
	So(sMembers.Err(), ShouldBeNil)
	So(stringSliceEqual(sMembers.Val(), []string{"one"}, false), ShouldBeTrue)

	sMembers = c.SMembers(ctx, key2)
	So(sMembers.Err(), ShouldBeNil)
	So(stringSliceEqual(sMembers.Val(), []string{"three", "two"}, false), ShouldBeTrue)

	return []string{key1, key2}
}

func testSPop(ctx context.Context, c Cmdable) []string {
	var key = "set"

	sAdd := c.SAdd(ctx, key, "one")
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key, "two")
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key, "three")
	So(sAdd.Err(), ShouldBeNil)

	sPop := c.SPop(ctx, key)
	So(sPop.Err(), ShouldBeNil)
	So(sPop.Val(), ShouldNotBeEmpty)

	sMembers := c.SMembers(ctx, key)
	So(sMembers.Err(), ShouldBeNil)
	So(len(sMembers.Val()), ShouldEqual, 2)

	return []string{key}
}

func testSPopN(ctx context.Context, c Cmdable) []string {
	var key = "set"

	sAdd := c.SAdd(ctx, key, "one")
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key, "two")
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key, "three")
	So(sAdd.Err(), ShouldBeNil)

	sPop := c.SPopN(ctx, key, 2)
	So(sPop.Err(), ShouldBeNil)
	So(sPop.Val(), ShouldNotBeEmpty)

	sMembers := c.SMembers(ctx, key)
	So(sMembers.Err(), ShouldBeNil)
	So(len(sMembers.Val()), ShouldEqual, 1)

	return []string{key}
}

func testSRandMember(ctx context.Context, c Cmdable) []string {
	var key = "set"

	sAdd := c.SAdd(ctx, key, "one")
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key, "two")
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key, "three")
	So(sAdd.Err(), ShouldBeNil)

	sPop := c.SRandMember(ctx, key)
	So(sPop.Err(), ShouldBeNil)
	So(sPop.Val(), ShouldNotBeEmpty)

	return []string{key}
}

func testSRandMemberN(ctx context.Context, c Cmdable) []string {
	var key = "set"

	sAdd := c.SAdd(ctx, key, "one")
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key, "two")
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key, "three")
	So(sAdd.Err(), ShouldBeNil)

	sPop := c.SRandMemberN(ctx, key, 2)
	So(sPop.Err(), ShouldBeNil)
	So(len(sPop.Val()), ShouldEqual, 2)

	sPop = c.SRandMemberN(ctx, key, 5)
	So(sPop.Err(), ShouldBeNil)
	So(len(sPop.Val()), ShouldEqual, 3)

	sPop = c.SRandMemberN(ctx, key, -5)
	So(sPop.Err(), ShouldBeNil)
	So(len(sPop.Val()), ShouldEqual, 5)

	return []string{key}
}

func testSRem(ctx context.Context, c Cmdable) []string {
	var key = "set"

	sAdd := c.SAdd(ctx, key, "one")
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key, "two")
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key, "three")
	So(sAdd.Err(), ShouldBeNil)

	sRem := c.SRem(ctx, key, "one")
	So(sRem.Err(), ShouldBeNil)
	So(sRem.Val(), ShouldEqual, 1)

	sRem = c.SRem(ctx, key, "four")
	So(sRem.Err(), ShouldBeNil)
	So(sRem.Val(), ShouldEqual, 0)

	sMembers := c.SMembers(ctx, key)
	So(sMembers.Err(), ShouldBeNil)
	So(stringSliceEqual(sMembers.Val(), []string{"three", "two"}, false), ShouldBeTrue)

	return []string{key}
}

func testSScan(ctx context.Context, c Cmdable) []string {
	var key = "set"

	for i := 0; i < 1000; i++ {
		sAdd := c.SAdd(ctx, key, fmt.Sprintf("member%d", i))
		So(sAdd.Err(), ShouldBeNil)
	}

	keys, cursor, err := c.SScan(ctx, key, 0, "", 0).Result()
	So(err, ShouldBeNil)
	So(keys, ShouldNotBeEmpty)
	So(len(keys), ShouldBeGreaterThan, 0)
	So(cursor, ShouldBeGreaterThan, 0)

	return []string{key}
}

func testSUnion(ctx context.Context, c Cmdable) []string {
	var key1, key2 = "set1", "set2"

	sAdd := c.SAdd(ctx, key1, "a")
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key1, "b")
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key1, "c")
	So(sAdd.Err(), ShouldBeNil)

	sAdd = c.SAdd(ctx, key2, "c")
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key2, "d")
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key2, "e")
	So(sAdd.Err(), ShouldBeNil)

	sUnion := c.SUnion(ctx, key1, key2)
	So(sUnion.Err(), ShouldBeNil)
	So(len(sUnion.Val()), ShouldEqual, 5)

	return []string{key1, key2}
}

func testSUnionStore(ctx context.Context, c Cmdable) []string {
	var key, key1, key2 = "set", "set1", "set2"

	sAdd := c.SAdd(ctx, key1, "a")
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key1, "b")
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key1, "c")
	So(sAdd.Err(), ShouldBeNil)

	sAdd = c.SAdd(ctx, key2, "c")
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key2, "d")
	So(sAdd.Err(), ShouldBeNil)
	sAdd = c.SAdd(ctx, key2, "e")
	So(sAdd.Err(), ShouldBeNil)

	sUnionStore := c.SUnionStore(ctx, key, key1, key2)
	So(sUnionStore.Err(), ShouldBeNil)
	So(sUnionStore.Val(), ShouldEqual, 5)

	sMembers := c.SMembers(ctx, key)
	So(sMembers.Err(), ShouldBeNil)
	So(len(sMembers.Val()), ShouldEqual, 5)

	return []string{key, key1, key2}
}

func setTestUnits() []TestUnit {
	return []TestUnit{
		{CommandSAdd, testSAdd},
		{CommandSCard, testSCard},
		{CommandSDiff, testSDiff},
		{CommandSDiffStore, testSDiffStore},
		{CommandSInter, testSInter},
		{CommandSInterStore, testSInterStore},
		{CommandSIsMember, testSIsMember},
		{CommandSMIsMember, testSMIsMember},
		{CommandSMembers, testSMembers},
		{CommandSMembers, testSMembersMap},
		{CommandSMove, testSMove},
		{CommandSPop, testSPop},
		{CommandSPopN, testSPopN},
		{CommandSRandMember, testSRandMember},
		{CommandSRandMemberN, testSRandMemberN},
		{CommandSRem, testSRem},
		{CommandSScan, testSScan},
		{CommandSUnion, testSUnion},
		{CommandSUnionStore, testSUnionStore},
	}
}

func TestResp2Client_Set(t *testing.T) { doTestUnits(t, RESP2, scriptTestUnits) }
func TestResp3Client_Set(t *testing.T) { doTestUnits(t, RESP3, setTestUnits) }
