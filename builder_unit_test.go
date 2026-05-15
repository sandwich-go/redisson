//go:build miniredis_test

package redisson

import (
	"strings"
	"testing"
	"time"
)

// builder_unit_test.go 用 miniredis client 拿到的 c.builder 直接逐个调用
// XxxCompleted 函数，断言生成的 Completed.Commands() 至少包含期望的命令名/关键参数。
//
// 这是对 builder_*.go 大量 0 覆盖手写 wrapper 的契约性补齐：
//   - 不需要真发 Redis（builder 只构造命令字符串），但需要 rueidis 内部 Builder
//     实例，必须从 connect 后的 c.cmd.B() 拿；
//   - 通过 build tag miniredis_test 启动 miniredis 取得 c.builder，无网依赖。
//
// 检查策略：调用 XxxCompleted 不 panic + Commands() 返回非空数组。命令具体形态以
// rueidis 为准，本测试仅验证 wrapper 不抛错且产出有效 Completed。

// freshBuilderClient 启动一个独立 miniredis client，返回它的内部 *client。
// 仅用于 builder 单元测试，不发命令；调用方 t.Cleanup 自行 Close。
func freshBuilderClient(t *testing.T) *client {
	t.Helper()
	c := MustNewClient(NewConf(
		WithDevelopment(false),
		WithT(t),
		WithEnableCache(false),
	))
	t.Cleanup(func() { _ = c.Close() })
	return c.(*client)
}

// assertCmdsContains 断言 Completed.Commands() 至少含一条命令且首段大写关键字命中。
func assertCmdsContains(t *testing.T, name string, cmds []string, want string) {
	t.Helper()
	if len(cmds) == 0 {
		t.Fatalf("%s: empty Commands() output", name)
		return
	}
	upper := strings.ToUpper(strings.Join(cmds, " "))
	if !strings.Contains(upper, want) {
		t.Fatalf("%s: expected to contain %q, got %v", name, want, cmds)
	}
}

// ---------------------------------------------------------------------------
// builder_string.go
// ---------------------------------------------------------------------------

func TestBuilder_String(t *testing.T) {
	c := freshBuilderClient(t)
	b := c.builder

	cases := []struct {
		name string
		make func() Completed
		want string
	}{
		{"Append", func() Completed { return b.AppendCompleted("k", "v") }, "APPEND"},
		{"Decr", func() Completed { return b.DecrCompleted("k") }, "DECR"},
		{"DecrBy", func() Completed { return b.DecrByCompleted("k", 2) }, "DECRBY"},
		{"Get", func() Completed { return b.GetCompleted("k") }, "GET"},
		{"GetDel", func() Completed { return b.GetDelCompleted("k") }, "GETDEL"},
		{"GetEx-Ex", func() Completed { return b.GetExCompleted("k", 5*time.Second) }, "GETEX"},
		{"GetEx-Px", func() Completed { return b.GetExCompleted("k", 500*time.Millisecond) }, "PX"},
		{"GetEx-NoExpire", func() Completed { return b.GetExCompleted("k", 0) }, "GETEX"},
		{"GetRange", func() Completed { return b.GetRangeCompleted("k", 0, 10) }, "GETRANGE"},
		{"GetSet", func() Completed { return b.GetSetCompleted("k", "v") }, "GETSET"},
		{"Incr", func() Completed { return b.IncrCompleted("k") }, "INCR"},
		{"IncrBy", func() Completed { return b.IncrByCompleted("k", 3) }, "INCRBY"},
		{"IncrByFloat", func() Completed { return b.IncrByFloatCompleted("k", 1.5) }, "INCRBYFLOAT"},
		{"MGet", func() Completed { return b.MGetCompleted("{t}.k1", "{t}.k2") }, "MGET"},
		{"MSet", func() Completed { return b.MSetCompleted("{t}.k1", "v1", "{t}.k2", "v2") }, "MSET"},
		{"MSetNX", func() Completed { return b.MSetNXCompleted("{t}.k1", "v1") }, "MSETNX"},
		{"Set", func() Completed { return b.SetCompleted("k", "v", 5*time.Second) }, "SET"},
		{"Set-Px", func() Completed { return b.SetCompleted("k", "v", 500*time.Millisecond) }, "PX"},
		{"Set-NoExpire", func() Completed { return b.SetCompleted("k", "v", 0) }, "SET"},
		{"Set-KeepTTL", func() Completed { return b.SetKeepTTLCompleted("k", "v") }, "KEEPTTL"},
		{"SetEX", func() Completed { return b.SetEXCompleted("k", "v", 5*time.Second) }, "SETEX"},
		{"SetNX-NoExpire", func() Completed { return b.SetNXCompleted("k", "v", 0) }, "SETNX"},
		{"SetNX-Ex", func() Completed { return b.SetNXCompleted("k", "v", 5*time.Second) }, "EX"},
		{"SetNX-Px", func() Completed { return b.SetNXCompleted("k", "v", 500*time.Millisecond) }, "PX"},
		{"SetNX-KeepTTL", func() Completed { return b.SetNXCompleted("k", "v", KeepTTL) }, "KEEPTTL"},
		{"SetRange", func() Completed { return b.SetRangeCompleted("k", 0, "v") }, "SETRANGE"},
		{"StrLen", func() Completed { return b.StrLenCompleted("k") }, "STRLEN"},
		{"SetArgsXX", func() Completed { return b.SetArgsCompleted("k", "v", SetArgs{Mode: XX}) }, "XX"},
		{"SetArgsNX", func() Completed { return b.SetArgsCompleted("k", "v", SetArgs{Mode: NX}) }, "NX"},
		{"SetArgsTTL-Ex", func() Completed { return b.SetArgsCompleted("k", "v", SetArgs{TTL: 5 * time.Second}) }, "EX"},
		{"SetArgsTTL-Px", func() Completed {
			return b.SetArgsCompleted("k", "v", SetArgs{TTL: 500 * time.Millisecond})
		}, "PX"},
		{"SetArgsKeepTTL", func() Completed {
			return b.SetArgsCompleted("k", "v", SetArgs{KeepTTL: true})
		}, "KEEPTTL"},
		{"SetArgsExpireAt", func() Completed {
			return b.SetArgsCompleted("k", "v", SetArgs{ExpireAt: time.Now().Add(5 * time.Second)})
		}, "EXAT"},
		{"SetArgsGet", func() Completed { return b.SetArgsCompleted("k", "v", SetArgs{Get: true}) }, "GET"},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			cmd := c.make()
			cmds := cmd.Commands()
			assertCmdsContains(t, c.name, cmds, c.want)
		})
	}
}

// ---------------------------------------------------------------------------
// builder_bitmap.go
// ---------------------------------------------------------------------------

func TestBuilder_Bitmap(t *testing.T) {
	c := freshBuilderClient(t)
	b := c.builder

	cases := []struct {
		name string
		make func() Completed
		want string
	}{
		{"GetBit", func() Completed { return b.GetBitCompleted("k", 0) }, "GETBIT"},
		{"SetBit", func() Completed { return b.SetBitCompleted("k", 0, 1) }, "SETBIT"},
		{"BitCount-Empty", func() Completed { return b.BitCountCompleted("k", nil) }, "BITCOUNT"},
		{"BitCount-Range", func() Completed { return b.BitCountCompleted("k", &BitCount{Start: 0, End: 10}) }, "BITCOUNT"},
		{"BitCount-Byte", func() Completed { return b.BitCountCompleted("k", &BitCount{Start: 0, End: 10, Unit: BYTE}) }, "BYTE"},
		{"BitCount-Bit", func() Completed { return b.BitCountCompleted("k", &BitCount{Start: 0, End: 10, Unit: BIT}) }, "BIT"},
		{"BitOpAnd", func() Completed { return b.BitOpAndCompleted("{t}.dest", "{t}.k1", "{t}.k2") }, "BITOP"},
		{"BitOpOr", func() Completed { return b.BitOpOrCompleted("{t}.dest", "{t}.k1", "{t}.k2") }, "BITOP"},
		{"BitOpXor", func() Completed { return b.BitOpXorCompleted("{t}.dest", "{t}.k1", "{t}.k2") }, "BITOP"},
		{"BitOpNot", func() Completed { return b.BitOpNotCompleted("{t}.dest", "{t}.k") }, "BITOP"},
		{"BitPos-NoArg", func() Completed { return b.BitPosCompleted("k", 1) }, "BITPOS"},
		{"BitPos-Start", func() Completed { return b.BitPosCompleted("k", 1, 0) }, "BITPOS"},
		{"BitPos-StartEnd", func() Completed { return b.BitPosCompleted("k", 1, 0, 10) }, "BITPOS"},
		{"BitPosSpan-Byte", func() Completed { return b.BitPosSpanCompleted("k", 1, 0, 10, BYTE) }, "BYTE"},
		{"BitPosSpan-Bit", func() Completed { return b.BitPosSpanCompleted("k", 1, 0, 10, BIT) }, "BIT"},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			cmd := c.make()
			cmds := cmd.Commands()
			assertCmdsContains(t, c.name, cmds, c.want)
		})
	}
}

// ---------------------------------------------------------------------------
// builder_list.go
// ---------------------------------------------------------------------------

func TestBuilder_List(t *testing.T) {
	c := freshBuilderClient(t)
	b := c.builder

	cases := []struct {
		name string
		make func() Completed
		want string
	}{
		{"BLPop", func() Completed { return b.BLPopCompleted(time.Second, "k") }, "BLPOP"},
		{"BRPop", func() Completed { return b.BRPopCompleted(time.Second, "k") }, "BRPOP"},
		{"BRPopLPush", func() Completed { return b.BRPopLPushCompleted("{t}.src", "{t}.dst", time.Second) }, "BRPOPLPUSH"},
		{"LIndex", func() Completed { return b.LIndexCompleted("k", 0) }, "LINDEX"},
		{"LLen", func() Completed { return b.LLenCompleted("k") }, "LLEN"},
		{"LPop", func() Completed { return b.LPopCompleted("k") }, "LPOP"},
		{"LPopCount", func() Completed { return b.LPopCountCompleted("k", 3) }, "LPOP"},
		{"LPush", func() Completed { return b.LPushCompleted("k", "v1") }, "LPUSH"},
		{"LPushX", func() Completed { return b.LPushXCompleted("k", "v1") }, "LPUSHX"},
		{"LRange", func() Completed { return b.LRangeCompleted("k", 0, 10) }, "LRANGE"},
		{"LRem", func() Completed { return b.LRemCompleted("k", 1, "v") }, "LREM"},
		{"LSet", func() Completed { return b.LSetCompleted("k", 0, "v") }, "LSET"},
		{"LTrim", func() Completed { return b.LTrimCompleted("k", 0, 10) }, "LTRIM"},
		{"RPop", func() Completed { return b.RPopCompleted("k") }, "RPOP"},
		{"RPopCount", func() Completed { return b.RPopCountCompleted("k", 3) }, "RPOP"},
		{"RPopLPush", func() Completed { return b.RPopLPushCompleted("{t}.src", "{t}.dst") }, "RPOPLPUSH"},
		{"RPush", func() Completed { return b.RPushCompleted("k", "v1") }, "RPUSH"},
		{"RPushX", func() Completed { return b.RPushXCompleted("k", "v1") }, "RPUSHX"},
		{"LInsertBefore", func() Completed { return b.LInsertCompleted("k", BEFORE, "pivot", "v") }, "BEFORE"},
		{"LInsertAfter", func() Completed { return b.LInsertCompleted("k", AFTER, "pivot", "v") }, "AFTER"},
		{"LMove", func() Completed { return b.LMoveCompleted("{t}.src", "{t}.dst", LEFT, RIGHT) }, "LMOVE"},
		{"BLMove", func() Completed { return b.BLMoveCompleted("{t}.src", "{t}.dst", LEFT, RIGHT, time.Second) }, "BLMOVE"},
		{"LPos", func() Completed { return b.LPosCompleted("k", "v", LPosArgs{Rank: 1, MaxLen: 100}) }, "LPOS"},
		{"LPosCount", func() Completed { return b.LPosCountCompleted("k", "v", 3, LPosArgs{Rank: 1}) }, "COUNT"},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			cmd := c.make()
			cmds := cmd.Commands()
			assertCmdsContains(t, c.name, cmds, c.want)
		})
	}
}

// ---------------------------------------------------------------------------
// builder_hash.go
// ---------------------------------------------------------------------------

func TestBuilder_Hash(t *testing.T) {
	c := freshBuilderClient(t)
	b := c.builder

	cases := []struct {
		name string
		make func() Completed
		want string
	}{
		{"HDel", func() Completed { return b.HDelCompleted("k", "f1", "f2") }, "HDEL"},
		{"HExists", func() Completed { return b.HExistsCompleted("k", "f") }, "HEXISTS"},
		{"HGet", func() Completed { return b.HGetCompleted("k", "f") }, "HGET"},
		{"HGetAll", func() Completed { return b.HGetAllCompleted("k") }, "HGETALL"},
		{"HIncrBy", func() Completed { return b.HIncrByCompleted("k", "f", 2) }, "HINCRBY"},
		{"HIncrByFloat", func() Completed { return b.HIncrByFloatCompleted("k", "f", 1.5) }, "HINCRBYFLOAT"},
		{"HKeys", func() Completed { return b.HKeysCompleted("k") }, "HKEYS"},
		{"HLen", func() Completed { return b.HLenCompleted("k") }, "HLEN"},
		{"HMGet", func() Completed { return b.HMGetCompleted("k", "f1", "f2") }, "HMGET"},
		{"HMSet", func() Completed { return b.HMSetCompleted("k", "f1", "v1", "f2", "v2") }, "HSET"},
		{"HSet", func() Completed { return b.HSetCompleted("k", "f", "v") }, "HSET"},
		{"HSetNX", func() Completed { return b.HSetNXCompleted("k", "f", "v") }, "HSETNX"},
		{"HVals", func() Completed { return b.HValsCompleted("k") }, "HVALS"},
		{"HRandField", func() Completed { return b.HRandFieldCompleted("k", 3) }, "HRANDFIELD"},
		{"HRandFieldWithValues", func() Completed { return b.HRandFieldWithValuesCompleted("k", 3) }, "WITHVALUES"},
		{"HScan", func() Completed { return b.HScanCompleted("k", 0, "*", 10) }, "HSCAN"},
		{"HExpire", func() Completed { return b.HExpireCompleted("k", time.Second, "f") }, "HEXPIRE"},
		{"HExpireNX", func() Completed { return b.HExpireNXCompleted("k", time.Second, "f") }, "NX"},
		{"HExpireXX", func() Completed { return b.HExpireXXCompleted("k", time.Second, "f") }, "XX"},
		{"HExpireGT", func() Completed { return b.HExpireGTCompleted("k", time.Second, "f") }, "GT"},
		{"HExpireLT", func() Completed { return b.HExpireLTCompleted("k", time.Second, "f") }, "LT"},
		{"HPExpire", func() Completed { return b.HPExpireCompleted("k", time.Second, "f") }, "HPEXPIRE"},
		{"HPersist", func() Completed { return b.HPersistCompleted("k", "f") }, "HPERSIST"},
		{"HTTL", func() Completed { return b.HTTLCompleted("k", "f") }, "HTTL"},
		{"HPTTL", func() Completed { return b.HPTTLCompleted("k", "f") }, "HPTTL"},
		{"HExpireAt", func() Completed { return b.HExpireAtCompleted("k", time.Now().Add(time.Hour), "f") }, "HEXPIREAT"},
		{"HExpireTime", func() Completed { return b.HExpireTimeCompleted("k", "f") }, "HEXPIRETIME"},
		{"HPExpireAt", func() Completed { return b.HPExpireAtCompleted("k", time.Now().Add(time.Hour), "f") }, "HPEXPIREAT"},
		{"HPExpireTime", func() Completed { return b.HPExpireTimeCompleted("k", "f") }, "HPEXPIRETIME"},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			cmd := c.make()
			cmds := cmd.Commands()
			assertCmdsContains(t, c.name, cmds, c.want)
		})
	}
}

// ---------------------------------------------------------------------------
// builder_sortedset.go
// ---------------------------------------------------------------------------

func TestBuilder_SortedSet(t *testing.T) {
	c := freshBuilderClient(t)
	b := c.builder

	z := []Z{{Score: 1.0, Member: "m1"}, {Score: 2.0, Member: "m2"}}
	store := ZStore{Keys: []string{"{t}.k1", "{t}.k2"}}

	cases := []struct {
		name string
		make func() Completed
		want string
	}{
		{"ZAdd", func() Completed { return b.ZAddCompleted("k", z...) }, "ZADD"},
		{"ZAddNX", func() Completed { return b.ZAddNXCompleted("k", z...) }, "NX"},
		{"ZAddXX", func() Completed { return b.ZAddXXCompleted("k", z...) }, "XX"},
		{"ZAddLT", func() Completed { return b.ZAddLTCompleted("k", z...) }, "LT"},
		{"ZAddGT", func() Completed { return b.ZAddGTCompleted("k", z...) }, "GT"},
		{"ZAddCh", func() Completed { return b.ZAddChCompleted("k", z...) }, "CH"},
		{"ZAddArgs", func() Completed {
			return b.ZAddArgsCompleted("k", ZAddArgs{NX: true, Members: z})
		}, "NX"},
		{"ZAddArgsIncr", func() Completed {
			return b.ZAddArgsIncrCompleted("k", ZAddArgs{Members: z[:1]})
		}, "INCR"},
		{"ZCard", func() Completed { return b.ZCardCompleted("k") }, "ZCARD"},
		{"ZCount", func() Completed { return b.ZCountCompleted("k", "0", "10") }, "ZCOUNT"},
		{"ZLexCount", func() Completed { return b.ZLexCountCompleted("k", "-", "+") }, "ZLEXCOUNT"},
		{"ZIncrBy", func() Completed { return b.ZIncrByCompleted("k", 1.5, "m") }, "ZINCRBY"},
		{"ZInter", func() Completed { return b.ZInterCompleted(store) }, "ZINTER"},
		{"ZInterWithScores", func() Completed { return b.ZInterWithScoresCompleted(store) }, "WITHSCORES"},
		{"ZInterCard", func() Completed { return b.ZInterCardCompleted(0, "{t}.k1", "{t}.k2") }, "ZINTERCARD"},
		{"ZInterStore", func() Completed { return b.ZInterStoreCompleted("{t}.dst", store) }, "ZINTERSTORE"},
		{"ZMPop", func() Completed { return b.ZMPopCompleted("MIN", 1, "{t}.k") }, "ZMPOP"},
		{"ZMScore", func() Completed { return b.ZMScoreCompleted("k", "m1", "m2") }, "ZMSCORE"},
		{"ZPopMax", func() Completed { return b.ZPopMaxCompleted("k", 1) }, "ZPOPMAX"},
		{"ZPopMin", func() Completed { return b.ZPopMinCompleted("k", 1) }, "ZPOPMIN"},
		{"ZRange", func() Completed { return b.ZRangeCompleted("k", 0, 10) }, "ZRANGE"},
		{"ZRangeWithScores", func() Completed { return b.ZRangeWithScoresCompleted("k", 0, 10) }, "WITHSCORES"},
		{"ZRangeByScore", func() Completed {
			return b.ZRangeByScoreCompleted("k", ZRangeBy{Min: "0", Max: "10"})
		}, "BYSCORE"},
		{"ZRangeByLex", func() Completed {
			return b.ZRangeByLexCompleted("k", ZRangeBy{Min: "-", Max: "+"})
		}, "BYLEX"},
		{"ZRangeByScoreWithScores", func() Completed {
			return b.ZRangeByScoreWithScoresCompleted("k", ZRangeBy{Min: "0", Max: "10"})
		}, "WITHSCORES"},
		{"ZRangeArgs", func() Completed {
			return b.ZRangeArgsCompleted(ZRangeArgs{Key: "k", Start: "0", Stop: "10"})
		}, "ZRANGE"},
		{"ZRangeArgsWithScores", func() Completed {
			return b.ZRangeArgsWithScoresCompleted(ZRangeArgs{Key: "k", Start: "0", Stop: "10"})
		}, "WITHSCORES"},
		{"ZRangeStore", func() Completed {
			return b.ZRangeStoreCompleted("{t}.dst", ZRangeArgs{Key: "{t}.k", Start: "0", Stop: "10"})
		}, "ZRANGESTORE"},
		{"ZRank", func() Completed { return b.ZRankCompleted("k", "m") }, "ZRANK"},
		{"ZRankWithScore", func() Completed { return b.ZRankWithScoreCompleted("k", "m") }, "WITHSCORE"},
		{"ZRem", func() Completed { return b.ZRemCompleted("k", "m1", "m2") }, "ZREM"},
		{"ZRemRangeByRank", func() Completed { return b.ZRemRangeByRankCompleted("k", 0, 10) }, "ZREMRANGEBYRANK"},
		{"ZRemRangeByScore", func() Completed { return b.ZRemRangeByScoreCompleted("k", "0", "10") }, "ZREMRANGEBYSCORE"},
		{"ZRemRangeByLex", func() Completed { return b.ZRemRangeByLexCompleted("k", "-", "+") }, "ZREMRANGEBYLEX"},
		{"ZRevRange", func() Completed { return b.ZRevRangeCompleted("k", 0, 10) }, "ZREVRANGE"},
		{"ZRevRangeWithScores", func() Completed { return b.ZRevRangeWithScoresCompleted("k", 0, 10) }, "WITHSCORES"},
		{"ZRevRangeByScore", func() Completed {
			return b.ZRevRangeByScoreCompleted("k", ZRangeBy{Min: "0", Max: "10"})
		}, "ZREVRANGEBYSCORE"},
		{"ZRevRangeByLex", func() Completed {
			return b.ZRevRangeByLexCompleted("k", ZRangeBy{Min: "-", Max: "+"})
		}, "ZREVRANGEBYLEX"},
		{"ZRevRangeByScoreWithScores", func() Completed {
			return b.ZRevRangeByScoreWithScoresCompleted("k", ZRangeBy{Min: "0", Max: "10"})
		}, "WITHSCORES"},
		{"ZRevRank", func() Completed { return b.ZRevRankCompleted("k", "m") }, "ZREVRANK"},
		{"ZRevRankWithScore", func() Completed { return b.ZRevRankWithScoreCompleted("k", "m") }, "WITHSCORE"},
		{"ZScore", func() Completed { return b.ZScoreCompleted("k", "m") }, "ZSCORE"},
		{"ZUnion", func() Completed { return b.ZUnionCompleted(store) }, "ZUNION"},
		{"ZUnionWithScores", func() Completed { return b.ZUnionWithScoresCompleted(store) }, "WITHSCORES"},
		{"ZUnionStore", func() Completed { return b.ZUnionStoreCompleted("{t}.dst", store) }, "ZUNIONSTORE"},
		{"ZRandMember", func() Completed { return b.ZRandMemberCompleted("k", 3) }, "ZRANDMEMBER"},
		{"ZRandMemberWithScores", func() Completed { return b.ZRandMemberWithScoresCompleted("k", 3) }, "WITHSCORES"},
		{"ZDiff", func() Completed { return b.ZDiffCompleted("{t}.k1", "{t}.k2") }, "ZDIFF"},
		{"ZDiffWithScores", func() Completed { return b.ZDiffWithScoresCompleted("{t}.k1", "{t}.k2") }, "WITHSCORES"},
		{"ZDiffStore", func() Completed { return b.ZDiffStoreCompleted("{t}.dst", "{t}.k1", "{t}.k2") }, "ZDIFFSTORE"},
		{"ZScan", func() Completed { return b.ZScanCompleted("k", 0, "*", 10) }, "ZSCAN"},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			cmd := c.make()
			cmds := cmd.Commands()
			assertCmdsContains(t, c.name, cmds, c.want)
		})
	}
}

// ---------------------------------------------------------------------------
// builder_generic.go
// ---------------------------------------------------------------------------

func TestBuilder_Generic(t *testing.T) {
	c := freshBuilderClient(t)
	b := c.builder

	cases := []struct {
		name string
		make func() Completed
		want string
	}{
		{"Copy", func() Completed { return b.CopyCompleted("{t}.src", "{t}.dst", 0, false) }, "COPY"},
		{"CopyReplace", func() Completed { return b.CopyCompleted("{t}.src", "{t}.dst", 0, true) }, "REPLACE"},
		{"Del", func() Completed { return b.DelCompleted("{t}.k1", "{t}.k2") }, "DEL"},
		{"Dump", func() Completed { return b.DumpCompleted("k") }, "DUMP"},
		{"Exists", func() Completed { return b.ExistsCompleted("{t}.k1", "{t}.k2") }, "EXISTS"},
		{"Expire", func() Completed { return b.ExpireCompleted("k", time.Second) }, "EXPIRE"},
		{"ExpireNX", func() Completed { return b.ExpireNXCompleted("k", time.Second) }, "NX"},
		{"ExpireXX", func() Completed { return b.ExpireXXCompleted("k", time.Second) }, "XX"},
		{"ExpireGT", func() Completed { return b.ExpireGTCompleted("k", time.Second) }, "GT"},
		{"ExpireLT", func() Completed { return b.ExpireLTCompleted("k", time.Second) }, "LT"},
		{"ExpireAt", func() Completed { return b.ExpireAtCompleted("k", time.Now().Add(time.Hour)) }, "EXPIREAT"},
		{"ExpireAtNX", func() Completed { return b.ExpireAtNXCompleted("k", time.Now().Add(time.Hour)) }, "NX"},
		{"ExpireAtXX", func() Completed { return b.ExpireAtXXCompleted("k", time.Now().Add(time.Hour)) }, "XX"},
		{"ExpireAtGT", func() Completed { return b.ExpireAtGTCompleted("k", time.Now().Add(time.Hour)) }, "GT"},
		{"ExpireAtLT", func() Completed { return b.ExpireAtLTCompleted("k", time.Now().Add(time.Hour)) }, "LT"},
		{"ExpireTime", func() Completed { return b.ExpireTimeCompleted("k") }, "EXPIRETIME"},
		{"Keys", func() Completed { return b.KeysCompleted("*") }, "KEYS"},
		{"Move", func() Completed { return b.MoveCompleted("k", 1) }, "MOVE"},
		{"ObjectEncoding", func() Completed { return b.ObjectEncodingCompleted("k") }, "ENCODING"},
		{"ObjectIdleTime", func() Completed { return b.ObjectIdleTimeCompleted("k") }, "IDLETIME"},
		{"ObjectRefCount", func() Completed { return b.ObjectRefCountCompleted("k") }, "REFCOUNT"},
		{"Persist", func() Completed { return b.PersistCompleted("k") }, "PERSIST"},
		{"PExpire", func() Completed { return b.PExpireCompleted("k", time.Second) }, "PEXPIRE"},
		{"PExpireNX", func() Completed { return b.PExpireNXCompleted("k", time.Second) }, "NX"},
		{"PExpireXX", func() Completed { return b.PExpireXXCompleted("k", time.Second) }, "XX"},
		{"PExpireGT", func() Completed { return b.PExpireGTCompleted("k", time.Second) }, "GT"},
		{"PExpireLT", func() Completed { return b.PExpireLTCompleted("k", time.Second) }, "LT"},
		{"PExpireAt", func() Completed { return b.PExpireAtCompleted("k", time.Now().Add(time.Hour)) }, "PEXPIREAT"},
		{"PExpireAtNX", func() Completed { return b.PExpireAtNXCompleted("k", time.Now().Add(time.Hour)) }, "NX"},
		{"PExpireAtXX", func() Completed { return b.PExpireAtXXCompleted("k", time.Now().Add(time.Hour)) }, "XX"},
		{"PExpireAtGT", func() Completed { return b.PExpireAtGTCompleted("k", time.Now().Add(time.Hour)) }, "GT"},
		{"PExpireAtLT", func() Completed { return b.PExpireAtLTCompleted("k", time.Now().Add(time.Hour)) }, "LT"},
		{"PExpireTime", func() Completed { return b.PExpireTimeCompleted("k") }, "PEXPIRETIME"},
		{"PTTL", func() Completed { return b.PTTLCompleted("k") }, "PTTL"},
		{"RandomKey", func() Completed { return b.RandomKeyCompleted() }, "RANDOMKEY"},
		{"Rename", func() Completed { return b.RenameCompleted("{t}.k1", "{t}.k2") }, "RENAME"},
		{"RenameNX", func() Completed { return b.RenameNXCompleted("{t}.k1", "{t}.k2") }, "RENAMENX"},
		{"Restore", func() Completed { return b.RestoreCompleted("k", time.Second, "value") }, "RESTORE"},
		{"RestoreReplace", func() Completed { return b.RestoreReplaceCompleted("k", time.Second, "value") }, "REPLACE"},
		{"Scan", func() Completed { return b.ScanCompleted(0, "*", 10) }, "SCAN"},
		{"ScanType", func() Completed { return b.ScanTypeCompleted(0, "*", 10, "string") }, "TYPE"},
		{"Sort", func() Completed { return b.SortCompleted("k", Sort{By: "byw", Offset: 0, Count: 10}) }, "SORT"},
		{"SortRO", func() Completed { return b.SortROCompleted("k", Sort{By: "byw"}) }, "SORT_RO"},
		{"Touch", func() Completed { return b.TouchCompleted("{t}.k1", "{t}.k2") }, "TOUCH"},
		{"TTL", func() Completed { return b.TTLCompleted("k") }, "TTL"},
		{"Type", func() Completed { return b.TypeCompleted("k") }, "TYPE"},
		{"Unlink", func() Completed { return b.UnlinkCompleted("{t}.k1", "{t}.k2") }, "UNLINK"},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			cmd := c.make()
			cmds := cmd.Commands()
			assertCmdsContains(t, c.name, cmds, c.want)
		})
	}
}

// ---------------------------------------------------------------------------
// builder_geospatial.go
// ---------------------------------------------------------------------------

func TestBuilder_Geospatial(t *testing.T) {
	c := freshBuilderClient(t)
	b := c.builder

	loc := []GeoLocation{{Longitude: 1, Latitude: 2, Name: "n1"}}
	q := GeoSearchQuery{Member: "m", Radius: 100, RadiusUnit: KM, Sort: ASC}

	cases := []struct {
		name string
		make func() Completed
		want string
	}{
		{"GeoAdd", func() Completed { return b.GeoAddCompleted("k", loc...) }, "GEOADD"},
		{"GeoDist", func() Completed { return b.GeoDistCompleted("k", "m1", "m2", "km") }, "GEODIST"},
		{"GeoHash", func() Completed { return b.GeoHashCompleted("k", "m1") }, "GEOHASH"},
		{"GeoPos", func() Completed { return b.GeoPosCompleted("k", "m1") }, "GEOPOS"},
		{"GeoRadiusByMember", func() Completed {
			return b.GeoRadiusByMemberCompleted("k", "m", GeoRadiusQuery{Radius: 1, Unit: KM})
		}, "GEORADIUSBYMEMBER_RO"},
		{"GeoRadiusByMemberStore", func() Completed {
			return b.GeoRadiusByMemberStoreCompleted("k", "m", GeoRadiusQuery{Radius: 1, Unit: KM, Store: "{t}.dst"})
		}, "STORE"},
		{"GeoRadius", func() Completed {
			return b.GeoRadiusCompleted("k", 1, 2, GeoRadiusQuery{Radius: 1, Unit: KM})
		}, "GEORADIUS_RO"},
		{"GeoRadiusStore", func() Completed {
			return b.GeoRadiusStoreCompleted("k", 1, 2, GeoRadiusQuery{Radius: 1, Unit: KM, Store: "{t}.dst"})
		}, "STORE"},
		{"GeoSearch", func() Completed { return b.GeoSearchCompleted("k", q) }, "GEOSEARCH"},
		{"GeoSearchLocation", func() Completed {
			return b.GeoSearchLocationCompleted("k", GeoSearchLocationQuery{GeoSearchQuery: q})
		}, "GEOSEARCH"},
		{"GeoSearchStore", func() Completed {
			return b.GeoSearchStoreCompleted("{t}.dst", "{t}.k", GeoSearchStoreQuery{GeoSearchQuery: q})
		}, "GEOSEARCHSTORE"},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			cmd := c.make()
			cmds := cmd.Commands()
			assertCmdsContains(t, c.name, cmds, c.want)
		})
	}
}

// ---------------------------------------------------------------------------
// builder_stream.go
// ---------------------------------------------------------------------------

func TestBuilder_Stream(t *testing.T) {
	c := freshBuilderClient(t)
	b := c.builder

	cases := []struct {
		name string
		make func() Completed
		want string
	}{
		{"XAck", func() Completed { return b.XAckCompleted("k", "g", "id1") }, "XACK"},
		{"XAdd", func() Completed {
			return b.XAddCompleted(XAddArgs{Stream: "k", Values: map[string]any{"f": "v"}})
		}, "XADD"},
		{"XAutoClaim", func() Completed {
			return b.XAutoClaimCompleted(XAutoClaimArgs{Stream: "k", Group: "g", Consumer: "c", Start: "0"})
		}, "XAUTOCLAIM"},
		{"XAutoClaimJustID", func() Completed {
			return b.XAutoClaimJustIDCompleted(XAutoClaimArgs{Stream: "k", Group: "g", Consumer: "c", Start: "0"})
		}, "JUSTID"},
		{"XClaim", func() Completed {
			return b.XClaimCompleted(XClaimArgs{Stream: "k", Group: "g", Consumer: "c", MinIdle: time.Second, Messages: []string{"id"}})
		}, "XCLAIM"},
		{"XClaimJustID", func() Completed {
			return b.XClaimJustIDCompleted(XClaimArgs{Stream: "k", Group: "g", Consumer: "c", MinIdle: time.Second, Messages: []string{"id"}})
		}, "JUSTID"},
		{"XDel", func() Completed { return b.XDelCompleted("k", "id1") }, "XDEL"},
		{"XGroupCreate", func() Completed { return b.XGroupCreateCompleted("k", "g", "$") }, "XGROUP"},
		{"XGroupCreateMkStream", func() Completed { return b.XGroupCreateMkStreamCompleted("k", "g", "$") }, "MKSTREAM"},
		{"XGroupCreateConsumer", func() Completed { return b.XGroupCreateConsumerCompleted("k", "g", "c") }, "CREATECONSUMER"},
		{"XGroupDelConsumer", func() Completed { return b.XGroupDelConsumerCompleted("k", "g", "c") }, "DELCONSUMER"},
		{"XGroupDestroy", func() Completed { return b.XGroupDestroyCompleted("k", "g") }, "DESTROY"},
		{"XGroupSetID", func() Completed { return b.XGroupSetIDCompleted("k", "g", "$") }, "SETID"},
		{"XInfoConsumers", func() Completed { return b.XInfoConsumersCompleted("k", "g") }, "CONSUMERS"},
		{"XInfoGroups", func() Completed { return b.XInfoGroupsCompleted("k") }, "GROUPS"},
		{"XInfoStream", func() Completed { return b.XInfoStreamCompleted("k") }, "XINFO"},
		{"XInfoStreamFull", func() Completed { return b.XInfoStreamFullCompleted("k", 10) }, "FULL"},
		{"XLen", func() Completed { return b.XLenCompleted("k") }, "XLEN"},
		{"XPending", func() Completed { return b.XPendingCompleted("k", "g") }, "XPENDING"},
		{"XPendingExt", func() Completed {
			return b.XPendingExtCompleted(XPendingExtArgs{Stream: "k", Group: "g", Start: "-", End: "+", Count: 10})
		}, "XPENDING"},
		{"XRange", func() Completed { return b.XRangeCompleted("k", "-", "+") }, "XRANGE"},
		{"XRangeN", func() Completed { return b.XRangeNCompleted("k", "-", "+", 10) }, "COUNT"},
		{"XRevRange", func() Completed { return b.XRevRangeCompleted("k", "+", "-") }, "XREVRANGE"},
		{"XRevRangeN", func() Completed { return b.XRevRangeNCompleted("k", "+", "-", 10) }, "COUNT"},
		{"XTrim-MaxLen", func() Completed { return b.XTrimCompleted("k", 100) }, "MAXLEN"},
		{"XTrim-MaxLenApprox", func() Completed { return b.XTrimMaxLenApproxCompleted("k", 100, 10) }, "MAXLEN"},
		{"XTrim-MinID", func() Completed { return b.XTrimMinIDCompleted("k", "0-1") }, "MINID"},
		{"XTrim-MinIDApprox", func() Completed { return b.XTrimMinIDApproxCompleted("k", "0-1", 10) }, "MINID"},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			cmd := c.make()
			cmds := cmd.Commands()
			assertCmdsContains(t, c.name, cmds, c.want)
		})
	}
}

// ---------------------------------------------------------------------------
// builder_script.go
// ---------------------------------------------------------------------------

func TestBuilder_Script(t *testing.T) {
	c := freshBuilderClient(t)
	b := c.builder

	cases := []struct {
		name string
		make func() Completed
		want string
	}{
		{"Eval", func() Completed { return b.EvalCompleted("return 1", []string{"k"}, "v") }, "EVAL"},
		{"EvalSha", func() Completed { return b.EvalShaCompleted("sha", []string{"k"}, "v") }, "EVALSHA"},
		{"EvalRO", func() Completed { return b.EvalROCompleted("return 1", []string{"k"}, "v") }, "EVAL_RO"},
		{"EvalShaRO", func() Completed { return b.EvalShaROCompleted("sha", []string{"k"}, "v") }, "EVALSHA_RO"},
		{"FunctionList-Empty", func() Completed { return b.FunctionListCompleted(FunctionListQuery{}) }, "FUNCTION"},
		{"FunctionList-Lib", func() Completed {
			return b.FunctionListCompleted(FunctionListQuery{LibraryNamePattern: "*", WithCode: true})
		}, "WITHCODE"},
		{"FunctionDump", func() Completed { return b.FunctionDumpCompleted() }, "DUMP"},
		{"FCall", func() Completed { return b.FCallCompleted("fn", []string{"k"}, "v") }, "FCALL"},
		{"FCallRO", func() Completed { return b.FCallROCompleted("fn", []string{"k"}, "v") }, "FCALL_RO"},
		{"ACLDryRun", func() Completed { return b.ACLDryRunCompleted("u", "GET", "k") }, "DRYRUN"},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			cmd := c.make()
			cmds := cmd.Commands()
			assertCmdsContains(t, c.name, cmds, c.want)
		})
	}
}

// ---------------------------------------------------------------------------
// builder_server.go
// ---------------------------------------------------------------------------

func TestBuilder_Server(t *testing.T) {
	c := freshBuilderClient(t)
	b := c.builder

	cases := []struct {
		name string
		make func() Completed
		want string
	}{
		{"Command", func() Completed { return b.CommandCompleted() }, "COMMAND"},
		{"CommandList", func() Completed { return b.CommandListCompleted(FilterBy{Module: "M"}) }, "FILTERBY"},
		{"CommandGetKeys", func() Completed { return b.CommandGetKeysCompleted("GET", "k") }, "GETKEYS"},
		{"CommandGetKeysAndFlags", func() Completed { return b.CommandGetKeysAndFlagsCompleted("GET", "k") }, "GETKEYSANDFLAGS"},
		{"ConfigGet", func() Completed { return b.ConfigGetCompleted("maxmemory") }, "CONFIG"},
		{"Info", func() Completed { return b.InfoCompleted("server") }, "INFO"},
		{"LastSave", func() Completed { return b.LastSaveCompleted() }, "LASTSAVE"},
		{"DebugObject", func() Completed { return b.DebugObjectCompleted("k") }, "DEBUG"},
		{"MemoryUsage", func() Completed { return b.MemoryUsageCompleted("k", 1) }, "MEMORY"},
		{"Time", func() Completed { return b.TimeCompleted() }, "TIME"},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			cmd := c.make()
			cmds := cmd.Commands()
			assertCmdsContains(t, c.name, cmds, c.want)
		})
	}
}

// ---------------------------------------------------------------------------
// builder_pubsub.go / builder_hyperlog.go / builder_cluster_conn.go
// ---------------------------------------------------------------------------

func TestBuilder_PubSubHyperLogClusterConn(t *testing.T) {
	c := freshBuilderClient(t)
	b := c.builder

	cases := []struct {
		name string
		make func() Completed
		want string
	}{
		// pubsub
		{"Publish", func() Completed { return b.PublishCompleted("ch", "msg") }, "PUBLISH"},
		{"SPublish", func() Completed { return b.SPublishCompleted("ch", "msg") }, "SPUBLISH"},
		{"PubSubChannels", func() Completed { return b.PubSubChannelsCompleted("*") }, "CHANNELS"},
		{"PubSubNumSub", func() Completed { return b.PubSubNumSubCompleted("ch") }, "NUMSUB"},
		{"PubSubNumPat", func() Completed { return b.PubSubNumPatCompleted() }, "NUMPAT"},
		{"PubSubShardChannels", func() Completed { return b.PubSubShardChannelsCompleted("*") }, "SHARDCHANNELS"},
		{"PubSubShardNumSub", func() Completed { return b.PubSubShardNumSubCompleted("ch") }, "SHARDNUMSUB"},
		// hyperlog
		{"PFAdd", func() Completed { return b.PFAddCompleted("k", "e1") }, "PFADD"},
		{"PFCount", func() Completed { return b.PFCountCompleted("{t}.k1", "{t}.k2") }, "PFCOUNT"},
		{"PFMerge", func() Completed { return b.PFMergeCompleted("{t}.dst", "{t}.src") }, "PFMERGE"},
		// cluster + connection
		{"ClusterReplicas", func() Completed { return b.ClusterReplicasCompleted("nodeID") }, "REPLICAS"},
		{"ClientGetName", func() Completed { return b.ClientGetNameCompleted() }, "GETNAME"},
		{"ClientList", func() Completed { return b.ClientListCompleted() }, "LIST"},
		{"Echo", func() Completed { return b.EchoCompleted("hello") }, "ECHO"},
		{"Ping", func() Completed { return b.PingCompleted() }, "PING"},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			cmd := c.make()
			cmds := cmd.Commands()
			assertCmdsContains(t, c.name, cmds, c.want)
		})
	}
}

// ---------------------------------------------------------------------------
// builder_set.go
// ---------------------------------------------------------------------------

func TestBuilder_Set(t *testing.T) {
	c := freshBuilderClient(t)
	b := c.builder

	cases := []struct {
		name string
		make func() Completed
		want string
	}{
		{"SAdd", func() Completed { return b.SAddCompleted("k", "m1", "m2") }, "SADD"},
		{"SCard", func() Completed { return b.SCardCompleted("k") }, "SCARD"},
		{"SDiff", func() Completed { return b.SDiffCompleted("{t}.k1", "{t}.k2") }, "SDIFF"},
		{"SDiffStore", func() Completed { return b.SDiffStoreCompleted("{t}.dst", "{t}.k1", "{t}.k2") }, "SDIFFSTORE"},
		{"SInter", func() Completed { return b.SInterCompleted("{t}.k1", "{t}.k2") }, "SINTER"},
		{"SInterStore", func() Completed { return b.SInterStoreCompleted("{t}.dst", "{t}.k1", "{t}.k2") }, "SINTERSTORE"},
		{"SInterCard", func() Completed { return b.SInterCardCompleted(0, "{t}.k1", "{t}.k2") }, "SINTERCARD"},
		{"SIsMember", func() Completed { return b.SIsMemberCompleted("k", "m") }, "SISMEMBER"},
		{"SMIsMember", func() Completed { return b.SMIsMemberCompleted("k", "m1", "m2") }, "SMISMEMBER"},
		{"SMembers", func() Completed { return b.SMembersCompleted("k") }, "SMEMBERS"},
		{"SMove", func() Completed { return b.SMoveCompleted("{t}.src", "{t}.dst", "m") }, "SMOVE"},
		{"SPop", func() Completed { return b.SPopCompleted("k") }, "SPOP"},
		{"SPopN", func() Completed { return b.SPopNCompleted("k", 3) }, "SPOP"},
		{"SRandMember", func() Completed { return b.SRandMemberCompleted("k") }, "SRANDMEMBER"},
		{"SRandMemberN", func() Completed { return b.SRandMemberNCompleted("k", 3) }, "SRANDMEMBER"},
		{"SRem", func() Completed { return b.SRemCompleted("k", "m1", "m2") }, "SREM"},
		{"SUnion", func() Completed { return b.SUnionCompleted("{t}.k1", "{t}.k2") }, "SUNION"},
		{"SUnionStore", func() Completed { return b.SUnionStoreCompleted("{t}.dst", "{t}.k1", "{t}.k2") }, "SUNIONSTORE"},
		{"SScan", func() Completed { return b.SScanCompleted("k", 0, "*", 10) }, "SSCAN"},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			cmd := c.make()
			cmds := cmd.Commands()
			assertCmdsContains(t, c.name, cmds, c.want)
		})
	}
}
