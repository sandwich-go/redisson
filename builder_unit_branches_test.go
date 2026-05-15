//go:build miniredis_test

package redisson

import (
	"strings"
	"testing"
	"time"
)

// builder_unit_branches_test.go 补齐 builder_*.go 中分支条件下的覆盖（80% → 90%+）。
// 与 builder_unit_test.go 互补：本文件聚焦各 if/switch 的非默认分支、panic 路径，
// 以及单参 vs 变参的多种取值组合。

// ---------------------------------------------------------------------------
// helper
// ---------------------------------------------------------------------------

// expectPanicContains 断言 fn() panic 出 *ParameterError 且 message 含 want 子串。
func expectPanicContains(t *testing.T, name, want string, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("%s: expected panic, got none", name)
		}
		if !IsParameterError(r) {
			t.Fatalf("%s: expected *ParameterError, got %T(%v)", name, r, r)
		}
		err := r.(error)
		if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(want)) {
			t.Fatalf("%s: error should contain %q, got: %s", name, want, err.Error())
		}
	}()
	fn()
}

// ---------------------------------------------------------------------------
// odd args panic 路径（与 builder_args_odd_test.go 互补，让 miniredis_test 套
// 也覆盖到 builder_string.go / builder_hash.go 的 odd panic 行）
// ---------------------------------------------------------------------------

func TestBuilder_OddArgsPanics_MR(t *testing.T) {
	c := freshBuilderClient(t)
	b := c.builder

	expectPanicContains(t, "MSet-Odd", "even", func() {
		_ = b.MSetCompleted("k1", "v1", "k2")
	})
	expectPanicContains(t, "MSetNX-Odd", "even", func() {
		_ = b.MSetNXCompleted("k1", "v1", "k2")
	})
	expectPanicContains(t, "HMSet-Odd", "even", func() {
		_ = b.HMSetCompleted("h", "f1", "v1", "f2")
	})
	expectPanicContains(t, "HMSetX-Odd", "even", func() {
		_ = b.HMSetXCompleted("h", "f1", "v1", "f2")
	})
}

// ---------------------------------------------------------------------------
// builder_string.go SetXXCompleted + SetArgs default panic
// ---------------------------------------------------------------------------

func TestBuilder_String_Branches(t *testing.T) {
	c := freshBuilderClient(t)
	b := c.builder

	cases := []struct {
		name string
		make func() Completed
		want string
	}{
		{"SetXX-Px", func() Completed { return b.SetXXCompleted("k", "v", 500*time.Millisecond) }, "PX"},
		{"SetXX-Ex", func() Completed { return b.SetXXCompleted("k", "v", 5*time.Second) }, "EX"},
		{"SetXX-KeepTTL", func() Completed { return b.SetXXCompleted("k", "v", KeepTTL) }, "KEEPTTL"},
		{"SetXX-NoExpire", func() Completed { return b.SetXXCompleted("k", "v", 0) }, "XX"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			cmd := c.make()
			assertCmdsContains(t, c.name, cmd.Commands(), c.want)
		})
	}

	// SetArgs default 分支:Mode 是非法值 → panic ParameterError
	expectPanicContains(t, "SetArgs-InvalidMode", "invalid mode", func() {
		_ = b.SetArgsCompleted("k", "v", SetArgs{Mode: "INVALID"})
	})
}

// ---------------------------------------------------------------------------
// builder_bitmap.go BitCount default panic + BitPos panic + BitField
// ---------------------------------------------------------------------------

func TestBuilder_Bitmap_Branches(t *testing.T) {
	c := freshBuilderClient(t)
	b := c.builder

	// BitField 多个 args 触发 cmd = cmd.Args(...) 循环
	cmd := b.BitFieldCompleted("k", "GET", "u8", "100", "SET", "u8", "100", "200")
	assertCmdsContains(t, "BitField", cmd.Commands(), "BITFIELD")

	// BitCount 非法 unit → panic
	expectPanicContains(t, "BitCount-InvalidUnit", "invalid unit", func() {
		_ = b.BitCountCompleted("k", &BitCount{Start: 0, End: 10, Unit: "INVALID"})
	})

	// BitPos 太多参数 → panic
	expectPanicContains(t, "BitPos-TooMany", "too many", func() {
		_ = b.BitPosCompleted("k", 1, 0, 10, 20)
	})
}

// ---------------------------------------------------------------------------
// builder_geospatial.go GeoDist 各 unit + GeoRadius* panic 分支
// ---------------------------------------------------------------------------

func TestBuilder_Geospatial_Branches(t *testing.T) {
	c := freshBuilderClient(t)
	b := c.builder

	cases := []struct {
		name string
		make func() Completed
		want string
	}{
		{"GeoDist-M", func() Completed { return b.GeoDistCompleted("k", "m1", "m2", "M") }, "GEODIST"},
		{"GeoDist-Mi", func() Completed { return b.GeoDistCompleted("k", "m1", "m2", "MI") }, "GEODIST"},
		{"GeoDist-Ft", func() Completed { return b.GeoDistCompleted("k", "m1", "m2", "FT") }, "GEODIST"},
		{"GeoDist-Km", func() Completed { return b.GeoDistCompleted("k", "m1", "m2", "KM") }, "GEODIST"},
		{"GeoDist-Empty", func() Completed { return b.GeoDistCompleted("k", "m1", "m2", "") }, "GEODIST"},
		{"GeoSearchStore-StoreDist", func() Completed {
			return b.GeoSearchStoreCompleted("{t}.dst", "{t}.k", GeoSearchStoreQuery{
				GeoSearchQuery: GeoSearchQuery{Member: "m", Radius: 100, RadiusUnit: KM},
				StoreDist:      true,
			})
		}, "STOREDIST"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			cmd := c.make()
			assertCmdsContains(t, c.name, cmd.Commands(), c.want)
		})
	}

	// 非法 unit → panic
	expectPanicContains(t, "GeoDist-InvalidUnit", "invalid unit", func() {
		_ = b.GeoDistCompleted("k", "m1", "m2", "BAD")
	})

	// GeoRadiusByMember 不允许 Store/StoreDist
	expectPanicContains(t, "GeoRadiusByMember-Forbidden", "does not support", func() {
		_ = b.GeoRadiusByMemberCompleted("k", "m", GeoRadiusQuery{Radius: 1, Unit: KM, Store: "{t}.dst"})
	})
	// GeoRadiusByMemberStore 必须有 Store/StoreDist
	expectPanicContains(t, "GeoRadiusByMemberStore-Required", "requires", func() {
		_ = b.GeoRadiusByMemberStoreCompleted("k", "m", GeoRadiusQuery{Radius: 1, Unit: KM})
	})
	// GeoRadius 不允许 Store/StoreDist
	expectPanicContains(t, "GeoRadius-Forbidden", "does not support", func() {
		_ = b.GeoRadiusCompleted("k", 1, 2, GeoRadiusQuery{Radius: 1, Unit: KM, StoreDist: "{t}.dst"})
	})
	// GeoRadiusStore 必须有 Store/StoreDist
	expectPanicContains(t, "GeoRadiusStore-Required", "requires", func() {
		_ = b.GeoRadiusStoreCompleted("k", 1, 2, GeoRadiusQuery{Radius: 1, Unit: KM})
	})
}

// ---------------------------------------------------------------------------
// builder_hash.go HPExpire / HPExpireAt 各模式
// ---------------------------------------------------------------------------

func TestBuilder_Hash_HPExpireVariants(t *testing.T) {
	c := freshBuilderClient(t)
	b := c.builder

	cases := []struct {
		name string
		make func() Completed
		want string
	}{
		{"HPExpireNX", func() Completed { return b.HPExpireNXCompleted("k", time.Second, "f") }, "NX"},
		{"HPExpireXX", func() Completed { return b.HPExpireXXCompleted("k", time.Second, "f") }, "XX"},
		{"HPExpireGT", func() Completed { return b.HPExpireGTCompleted("k", time.Second, "f") }, "GT"},
		{"HPExpireLT", func() Completed { return b.HPExpireLTCompleted("k", time.Second, "f") }, "LT"},
		{"HPExpireAtNX", func() Completed { return b.HPExpireAtNXCompleted("k", time.Now().Add(time.Hour), "f") }, "NX"},
		{"HPExpireAtXX", func() Completed { return b.HPExpireAtXXCompleted("k", time.Now().Add(time.Hour), "f") }, "XX"},
		{"HPExpireAtGT", func() Completed { return b.HPExpireAtGTCompleted("k", time.Now().Add(time.Hour), "f") }, "GT"},
		{"HPExpireAtLT", func() Completed { return b.HPExpireAtLTCompleted("k", time.Now().Add(time.Hour), "f") }, "LT"},
		// HExpireAt 的秒级 NX/XX/GT/LT 变体（与 HPExpireAt 毫秒级对偶）
		{"HExpireAtNX", func() Completed { return b.HExpireAtNXCompleted("k", time.Now().Add(time.Hour), "f") }, "NX"},
		{"HExpireAtXX", func() Completed { return b.HExpireAtXXCompleted("k", time.Now().Add(time.Hour), "f") }, "XX"},
		{"HExpireAtGT", func() Completed { return b.HExpireAtGTCompleted("k", time.Now().Add(time.Hour), "f") }, "GT"},
		{"HExpireAtLT", func() Completed { return b.HExpireAtLTCompleted("k", time.Now().Add(time.Hour), "f") }, "LT"},
		{"HStrLen", func() Completed { return b.HStrLenCompleted("k", "f") }, "HSTRLEN"},
		{"HMSetX", func() Completed { return b.HMSetXCompleted("k", "f1", "v1", "f2", "v2") }, "HSET"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			cmd := c.make()
			assertCmdsContains(t, c.name, cmd.Commands(), c.want)
		})
	}
}

// ---------------------------------------------------------------------------
// builder_list.go LMPush / RMPush + LInsert panic
// ---------------------------------------------------------------------------

func TestBuilder_List_Branches(t *testing.T) {
	c := freshBuilderClient(t)
	b := c.builder

	cases := []struct {
		name string
		make func() Completed
		want string
	}{
		{"LMPush", func() Completed { return b.LMPushCompleted("k", "v1", "v2", "v3") }, "LPUSH"},
		{"LMPushX", func() Completed { return b.LMPushXCompleted("k", "v1", "v2") }, "LPUSHX"},
		{"RMPush", func() Completed { return b.RMPushCompleted("k", "v1", "v2", "v3") }, "RPUSH"},
		{"RMPushX", func() Completed { return b.RMPushXCompleted("k", "v1", "v2") }, "RPUSHX"},
		{"LMPop-NoCount", func() Completed { return b.LMPopCompleted(LEFT, 0, "{t}.k") }, "LMPOP"},
		{"LMPop-WithCount", func() Completed { return b.LMPopCompleted(LEFT, 5, "{t}.k") }, "COUNT"},
		{"LPos-NoArgs", func() Completed { return b.LPosCompleted("k", "v", LPosArgs{}) }, "LPOS"},
		{"LPosCount-NoArgs", func() Completed { return b.LPosCountCompleted("k", "v", 3, LPosArgs{}) }, "COUNT"},
		// LPosCount 带 MaxLen（覆盖 if a.MaxLen != 0 分支）
		{"LPosCount-WithMaxLen", func() Completed {
			return b.LPosCountCompleted("k", "v", 3, LPosArgs{Rank: 1, MaxLen: 100})
		}, "MAXLEN"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			cmd := c.make()
			assertCmdsContains(t, c.name, cmd.Commands(), c.want)
		})
	}

	// LInsert 非法 op → panic
	expectPanicContains(t, "LInsert-InvalidOp", "invalid", func() {
		_ = b.LInsertCompleted("k", "INVALID", "p", "v")
	})
	// LInsertBeforeCompleted/AfterCompleted 是显式版本，无 panic
	beforeCmd := b.LInsertBeforeCompleted("k", "p", "v")
	assertCmdsContains(t, "LInsertBefore", beforeCmd.Commands(), "BEFORE")
	afterCmd := b.LInsertAfterCompleted("k", "p", "v")
	assertCmdsContains(t, "LInsertAfter", afterCmd.Commands(), "AFTER")
}

// ---------------------------------------------------------------------------
// builder_generic.go Sort 各分支 + panic
// ---------------------------------------------------------------------------

func TestBuilder_Generic_SortBranches(t *testing.T) {
	c := freshBuilderClient(t)
	b := c.builder

	cases := []struct {
		name string
		make func() Completed
		want string
	}{
		{"Sort-Asc", func() Completed { return b.SortCompleted("k", Sort{Order: "ASC"}) }, "ASC"},
		{"Sort-Desc", func() Completed { return b.SortCompleted("k", Sort{Order: "DESC"}) }, "DESC"},
		{"Sort-Limit", func() Completed { return b.SortCompleted("k", Sort{Offset: 0, Count: 10}) }, "LIMIT"},
		{"Sort-Get", func() Completed { return b.SortCompleted("k", Sort{Get: []string{"obj_*", "obj2_*"}}) }, "GET"},
		{"Sort-Alpha", func() Completed { return b.SortCompleted("k", Sort{Alpha: true}) }, "ALPHA"},
		{"Sort-Empty", func() Completed { return b.SortCompleted("k", Sort{}) }, "SORT"},
		{"Migrate", func() Completed { return b.MigrateCompleted("h", 6379, "k", 1, time.Second) }, "MIGRATE"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			cmd := c.make()
			assertCmdsContains(t, c.name, cmd.Commands(), c.want)
		})
	}

	// Sort 非法 order → panic
	expectPanicContains(t, "Sort-InvalidOrder", "invalid sort", func() {
		_ = b.SortCompleted("k", Sort{Order: "INVALID"})
	})
}

// ---------------------------------------------------------------------------
// builder_sortedset.go ZRangeBy* / ZRev* / ZInterStore / ZUnionStore 分支
// ---------------------------------------------------------------------------

func TestBuilder_SortedSet_Branches(t *testing.T) {
	c := freshBuilderClient(t)
	b := c.builder

	storeWithWeights := ZStore{
		Keys:      []string{"{t}.k1", "{t}.k2"},
		Weights:   []int64{1, 2},
		Aggregate: "MAX",
	}
	by := ZRangeBy{Min: "0", Max: "10", Offset: 0, Count: 5}

	cases := []struct {
		name string
		make func() Completed
		want string
	}{
		// ZInterStore + ZUnionStore Weights/Aggregate
		{"ZInterStore-Weights", func() Completed { return b.ZInterStoreCompleted("{t}.dst", storeWithWeights) }, "WEIGHTS"},
		{"ZInterStore-Aggregate", func() Completed { return b.ZInterStoreCompleted("{t}.dst", storeWithWeights) }, "AGGREGATE"},
		{"ZUnionStore-Weights", func() Completed { return b.ZUnionStoreCompleted("{t}.dst", storeWithWeights) }, "WEIGHTS"},
		{"ZUnion-Weights", func() Completed { return b.ZUnionCompleted(storeWithWeights) }, "WEIGHTS"},
		{"ZInter-Weights", func() Completed { return b.ZInterCompleted(storeWithWeights) }, "WEIGHTS"},
		// ZPopMax/Min 0 args
		{"ZPopMax-NoArgs", func() Completed { return b.ZPopMaxCompleted("k") }, "ZPOPMAX"},
		{"ZPopMin-NoArgs", func() Completed { return b.ZPopMinCompleted("k") }, "ZPOPMIN"},
		// ZRangeByScore/Lex 带 Limit
		{"ZRangeByScore-Limit", func() Completed { return b.ZRangeByScoreCompleted("k", by) }, "LIMIT"},
		{"ZRangeByScore-NoLimit", func() Completed {
			return b.ZRangeByScoreCompleted("k", ZRangeBy{Min: "0", Max: "10"})
		}, "ZRANGEBYSCORE"},
		{"ZRangeByLex-Limit", func() Completed {
			return b.ZRangeByLexCompleted("k", ZRangeBy{Min: "-", Max: "+", Offset: 0, Count: 5})
		}, "LIMIT"},
		{"ZRangeByLex-NoLimit", func() Completed {
			return b.ZRangeByLexCompleted("k", ZRangeBy{Min: "-", Max: "+"})
		}, "ZRANGEBYLEX"},
		{"ZRangeByScoreWithScores-Limit", func() Completed { return b.ZRangeByScoreWithScoresCompleted("k", by) }, "LIMIT"},
		{"ZRangeByScoreWithScores-NoLimit", func() Completed {
			return b.ZRangeByScoreWithScoresCompleted("k", ZRangeBy{Min: "0", Max: "10"})
		}, "WITHSCORES"},
		// ZRev*
		{"ZRevRangeByScore-Limit", func() Completed { return b.ZRevRangeByScoreCompleted("k", by) }, "ZREVRANGEBYSCORE"},
		{"ZRevRangeByScore-NoLimit", func() Completed {
			return b.ZRevRangeByScoreCompleted("k", ZRangeBy{Min: "0", Max: "10"})
		}, "ZREVRANGEBYSCORE"},
		{"ZRevRangeByLex-Limit", func() Completed {
			return b.ZRevRangeByLexCompleted("k", ZRangeBy{Min: "-", Max: "+", Offset: 0, Count: 5})
		}, "ZREVRANGEBYLEX"},
		{"ZRevRangeByLex-NoLimit", func() Completed {
			return b.ZRevRangeByLexCompleted("k", ZRangeBy{Min: "-", Max: "+"})
		}, "ZREVRANGEBYLEX"},
		{"ZRevRangeByScoreWithScores-Limit", func() Completed { return b.ZRevRangeByScoreWithScoresCompleted("k", by) }, "WITHSCORES"},
		{"ZRevRangeByScoreWithScores-NoLimit", func() Completed {
			return b.ZRevRangeByScoreWithScoresCompleted("k", ZRangeBy{Min: "0", Max: "10"})
		}, "WITHSCORES"},
		// zRangeArgs withScores=true / Rev / ByLex
		{"ZRangeArgs-WithScores", func() Completed {
			return b.ZRangeArgsWithScoresCompleted(ZRangeArgs{Key: "k", Start: "0", Stop: "10", ByScore: true})
		}, "WITHSCORES"},
		{"ZRangeArgs-Rev-ByScore", func() Completed {
			return b.ZRangeArgsCompleted(ZRangeArgs{Key: "k", Start: "0", Stop: "10", ByScore: true, Rev: true})
		}, "REV"},
		{"ZRangeArgs-ByLex", func() Completed {
			return b.ZRangeArgsCompleted(ZRangeArgs{Key: "k", Start: "-", Stop: "+", ByLex: true})
		}, "BYLEX"},
		{"ZRangeArgs-Limit", func() Completed {
			return b.ZRangeArgsCompleted(ZRangeArgs{Key: "k", Start: "0", Stop: "10", Offset: 0, Count: 5})
		}, "LIMIT"},
		// ZRangeStore Rev+ByScore (覆盖交换分支)
		{"ZRangeStore-Rev-ByScore", func() Completed {
			return b.ZRangeStoreCompleted("{t}.dst", ZRangeArgs{Key: "{t}.k", Start: "0", Stop: "10", ByScore: true, Rev: true, Offset: 0, Count: 5})
		}, "REV"},
		{"ZRangeStore-ByLex", func() Completed {
			return b.ZRangeStoreCompleted("{t}.dst", ZRangeArgs{Key: "{t}.k", Start: "-", Stop: "+", ByLex: true})
		}, "BYLEX"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			cmd := c.make()
			assertCmdsContains(t, c.name, cmd.Commands(), c.want)
		})
	}

	// ZPopMax/Min too many args → panic
	expectPanicContains(t, "ZPopMax-TooMany", "too many", func() {
		_ = b.ZPopMaxCompleted("k", 1, 2, 3)
	})
	expectPanicContains(t, "ZPopMin-TooMany", "too many", func() {
		_ = b.ZPopMinCompleted("k", 1, 2, 3)
	})
}

// ---------------------------------------------------------------------------
// builder_stream.go XAdd 各分支 + xTrim Approx + Limit
// ---------------------------------------------------------------------------

func TestBuilder_Stream_Branches(t *testing.T) {
	c := freshBuilderClient(t)
	b := c.builder

	cases := []struct {
		name string
		make func() Completed
		want string
	}{
		{"XAdd-MaxLen-NoApprox", func() Completed {
			return b.XAddCompleted(XAddArgs{Stream: "k", MaxLen: 100, Values: map[string]any{"f": "v"}})
		}, "MAXLEN"},
		{"XAdd-MaxLen-Approx", func() Completed {
			return b.XAddCompleted(XAddArgs{Stream: "k", MaxLen: 100, Approx: true, Values: map[string]any{"f": "v"}})
		}, "MAXLEN"},
		{"XAdd-MinID-NoApprox", func() Completed {
			return b.XAddCompleted(XAddArgs{Stream: "k", MinID: "0-0", Values: map[string]any{"f": "v"}})
		}, "MINID"},
		{"XAdd-MinID-Approx", func() Completed {
			return b.XAddCompleted(XAddArgs{Stream: "k", MinID: "0-0", Approx: true, Values: map[string]any{"f": "v"}})
		}, "MINID"},
		{"XAdd-Limit", func() Completed {
			return b.XAddCompleted(XAddArgs{Stream: "k", MaxLen: 100, Limit: 10, Values: map[string]any{"f": "v"}})
		}, "LIMIT"},
		{"XAdd-NoMkStream", func() Completed {
			return b.XAddCompleted(XAddArgs{Stream: "k", NoMkStream: true, Values: map[string]any{"f": "v"}})
		}, "NOMKSTREAM"},
		{"XAdd-WithID", func() Completed {
			return b.XAddCompleted(XAddArgs{Stream: "k", ID: "1-1", Values: map[string]any{"f": "v"}})
		}, "1-1"},
		// XAutoClaim 无 Count
		{"XAutoClaim-NoCount", func() Completed {
			return b.XAutoClaimCompleted(XAutoClaimArgs{Stream: "k", Group: "g", Consumer: "c", Start: "0"})
		}, "XAUTOCLAIM"},
		// XAutoClaim 带 Count（覆盖 if a.Count > 0 分支）
		{"XAutoClaim-WithCount", func() Completed {
			return b.XAutoClaimCompleted(XAutoClaimArgs{Stream: "k", Group: "g", Consumer: "c", Start: "0", Count: 10})
		}, "XAUTOCLAIM"},
		{"XAutoClaimJustID-NoCount", func() Completed {
			return b.XAutoClaimJustIDCompleted(XAutoClaimArgs{Stream: "k", Group: "g", Consumer: "c", Start: "0"})
		}, "JUSTID"},
		{"XAutoClaimJustID-WithCount", func() Completed {
			return b.XAutoClaimJustIDCompleted(XAutoClaimArgs{Stream: "k", Group: "g", Consumer: "c", Start: "0", Count: 10})
		}, "JUSTID"},
		// XPendingExt with Idle/Consumer
		{"XPendingExt-WithIdle", func() Completed {
			return b.XPendingExtCompleted(XPendingExtArgs{Stream: "k", Group: "g", Idle: time.Second, Start: "-", End: "+", Count: 10})
		}, "IDLE"},
		{"XPendingExt-WithConsumer", func() Completed {
			return b.XPendingExtCompleted(XPendingExtArgs{Stream: "k", Group: "g", Start: "-", End: "+", Count: 10, Consumer: "c"})
		}, "XPENDING"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			cmd := c.make()
			assertCmdsContains(t, c.name, cmd.Commands(), c.want)
		})
	}
}

// ---------------------------------------------------------------------------
// builder_server.go CommandList 各 filter + MemoryUsage too many panic
// ---------------------------------------------------------------------------

func TestBuilder_Server_Branches(t *testing.T) {
	c := freshBuilderClient(t)
	b := c.builder

	cases := []struct {
		name string
		make func() Completed
		want string
	}{
		{"CommandList-Pattern", func() Completed { return b.CommandListCompleted(FilterBy{Pattern: "GET*"}) }, "PATTERN"},
		{"CommandList-ACLCat", func() Completed { return b.CommandListCompleted(FilterBy{ACLCat: "@read"}) }, "ACLCAT"},
		{"CommandList-Empty", func() Completed { return b.CommandListCompleted(FilterBy{}) }, "COMMAND"},
		{"MemoryUsage-NoSamples", func() Completed { return b.MemoryUsageCompleted("k") }, "MEMORY"},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			cmd := c.make()
			assertCmdsContains(t, c.name, cmd.Commands(), c.want)
		})
	}

	// MemoryUsage too many → panic
	expectPanicContains(t, "MemoryUsage-TooMany", "too many", func() {
		_ = b.MemoryUsageCompleted("k", 1, 2)
	})
}
