package redisson

import (
	"context"
	"time"
)

type SortedSetCmdable interface {
	SortedSetWriter
	SortedSetReader
}

type SortedSetWriter interface {
	// BZMPop
	// Available since: 7.0.0
	// Time complexity: O(K) + O(M*log(N)) where K is the number of provided keys, N being the number of elements in the sorted set, and M being the number of elements popped.
	// ACL categories: @write, @sortedset, @slow, @blocking
	BZMPop(ctx context.Context, timeout time.Duration, order string, count int64, keys ...string) ZSliceWithKeyCmd

	// BZPopMax
	// Available since: 5.0.0
	// Time complexity: O(log(N)) with N being the number of elements in the sorted set.
	// ACL categories: @write @sortedset @fast @blocking
	BZPopMax(ctx context.Context, timeout time.Duration, keys ...string) ZWithKeyCmd

	// BZPopMin
	// Available since: 5.0.0
	// Time complexity: O(log(N)) with N being the number of elements in the sorted set.
	// ACL categories: @write @sortedset @fast @blocking
	BZPopMin(ctx context.Context, timeout time.Duration, keys ...string) ZWithKeyCmd

	// ZAdd
	// Available since: 1.2.0
	// Time complexity: O(log(N)) for each item added, where N is the number of elements in the sorted set.
	// ACL categories: @write @sortedset @fast
	ZAdd(ctx context.Context, key string, members ...Z) IntCmd

	// ZAddNX
	// Available since: 3.0.2
	// Time complexity: O(log(N)) for each item added, where N is the number of elements in the sorted set.
	// ACL categories: @write @sortedset @fast
	ZAddNX(ctx context.Context, key string, members ...Z) IntCmd

	// ZAddXX
	// Available since: 3.0.2
	// Time complexity: O(log(N)) for each item added, where N is the number of elements in the sorted set.
	// ACL categories: @write @sortedset @fast
	ZAddXX(ctx context.Context, key string, members ...Z) IntCmd

	// ZAddLT
	// Available since:3.0.2
	// Time complexity: O(log(N)) for each item added, where N is the number of elements in the sorted set.
	// ACL categories: @write @sortedset @fast
	ZAddLT(ctx context.Context, key string, members ...Z) IntCmd

	// ZAddGT
	// Available since: 3.0.2
	// Time complexity: O(log(N)) for each item added, where N is the number of elements in the sorted set.
	// ACL categories: @write @sortedset @fast
	ZAddGT(ctx context.Context, key string, members ...Z) IntCmd

	// ZAddArgs
	// Available since: 3.0.2
	// Time complexity: O(log(N)) for each item added, where N is the number of elements in the sorted set.
	// ACL categories: @write @sortedset @fast
	// Starting with Redis version 6.2.0: Added the GT and LT options.
	ZAddArgs(ctx context.Context, key string, args ZAddArgs) IntCmd

	// ZAddArgsIncr
	// Available since: 3.0.2
	// Time complexity: O(log(N)) for each item added, where N is the number of elements in the sorted set.
	// ACL categories: @write @sortedset @fast
	// Starting with Redis version 6.2.0: Added the GT and LT options.
	ZAddArgsIncr(ctx context.Context, key string, args ZAddArgs) FloatCmd

	// ZDiffStore
	// Available since: 6.2.0
	// Time complexity: O(L + (N-K)log(N)) worst case where L is the total number of elements in all the sets, N is the size of the first set, and K is the size of the result set.
	// ACL categories: @write @sortedset @slow
	ZDiffStore(ctx context.Context, destination string, keys ...string) IntCmd

	// ZIncrBy
	// Available since: 1.2.0
	// Time complexity: O(log(N)) where N is the number of elements in the sorted set.
	// ACL categories: @write @sortedset @fast
	ZIncrBy(ctx context.Context, key string, increment float64, member string) FloatCmd

	// ZInterStore
	// Available since: 2.0.0
	// Time complexity: O(NK)+O(Mlog(M)) worst case with N being the smallest input sorted set, K being the number of input sorted sets and M being the number of elements in the resulting sorted set.
	// ACL categories: @write @sortedset @slow
	ZInterStore(ctx context.Context, destination string, store ZStore) IntCmd

	// ZMPop
	// Available since: 7.0.0
	// Time complexity: O(K) + O(M*log(N)) where K is the number of provided keys, N being the number of elements in the sorted set, and M being the number of elements popped.
	// ACL categories: @write @sortedset @slow
	ZMPop(ctx context.Context, order string, count int64, keys ...string) ZSliceWithKeyCmd

	// ZPopMax
	// Available since: 5.0.0
	// Time complexity: O(log(N)*M) with N being the number of elements in the sorted set, and M being the number of elements popped.
	// ACL categories: @write @sortedset @fast
	ZPopMax(ctx context.Context, key string, count ...int64) ZSliceCmd

	// ZPopMin
	// Available since: 5.0.0
	// Time complexity: O(log(N)*M) with N being the number of elements in the sorted set, and M being the number of elements popped.
	// ACL categories: @write @sortedset @fast
	ZPopMin(ctx context.Context, key string, count ...int64) ZSliceCmd

	// ZRangeStore
	// Available since: 6.2.0
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements stored into the destination key.
	// ACL categories: @write @sortedset @slow
	ZRangeStore(ctx context.Context, dst string, z ZRangeArgs) IntCmd

	// ZRem
	// Available since: 1.2.0
	// Time complexity: O(M*log(N)) with N being the number of elements in the sorted set and M the number of elements to be removed.
	// ACL categories: @write @sortedset @fast
	ZRem(ctx context.Context, key string, members ...any) IntCmd

	// ZRemRangeByLex
	// Available since: 2.8.9
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements removed by the operation.
	// ACL categories: @write @sortedset @slow
	ZRemRangeByLex(ctx context.Context, key, min, max string) IntCmd

	// ZRemRangeByRank
	// Available since: 2.0.0
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements removed by the operation.
	// ACL categories: @write @sortedset @slow
	ZRemRangeByRank(ctx context.Context, key string, start, stop int64) IntCmd

	// ZRemRangeByScore
	// Available since: 1.2.0
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements removed by the operation.
	// ACL categories: @write @sortedset @slow
	ZRemRangeByScore(ctx context.Context, key, min, max string) IntCmd

	// ZUnionStore
	// Available since: 2.0.0
	// Time complexity: O(N)+O(M log(M)) with N being the sum of the sizes of the input sorted sets, and M being the number of elements in the resulting sorted set.
	// ACL categories: @write @sortedset @slow
	ZUnionStore(ctx context.Context, dest string, store ZStore) IntCmd
}

type SortedSetReader interface {
	// ZInter
	// Available since: 6.2.0
	// Time complexity: O(NK)+O(Mlog(M)) worst case with N being the smallest input sorted set, K being the number of input sorted sets and M being the number of elements in the resulting sorted set.
	// ACL categories: @read @sortedset @slow
	ZInter(ctx context.Context, store ZStore) StringSliceCmd

	// ZInterWithScores
	// Available since: 6.2.0
	// Time complexity: O(NK)+O(Mlog(M)) worst case with N being the smallest input sorted set, K being the number of input sorted sets and M being the number of elements in the resulting sorted set.
	// ACL categories: @read @sortedset @slow
	ZInterWithScores(ctx context.Context, store ZStore) ZSliceCmd

	// ZInterCard
	// Available since: 7.0.0
	// Time complexity: O(N*K) worst case with N being the smallest input sorted set, K being the number of input sorted sets.
	// ACL categories: @read @sortedset @slow
	ZInterCard(ctx context.Context, limit int64, keys ...string) IntCmd

	// ZRandMember
	// Available since: 6.2.0
	// Time complexity: O(N) where N is the number of elements returned
	// ACL categories: @read @sortedset @slow
	ZRandMember(ctx context.Context, key string, count int64) StringSliceCmd
	ZRandMemberWithScores(ctx context.Context, key string, count int64) ZSliceCmd

	// ZScan
	// Available since: 2.8.0
	// Time complexity: O(1) for every call. O(N) for a complete iteration, including enough command calls for the cursor to return back to 0. N is the number of elements inside the collection..
	// ACL categories: @read @sortedset @slow
	ZScan(ctx context.Context, key string, cursor uint64, match string, count int64) ScanCmd

	// ZDiff
	// Available since: 6.2.0
	// Time complexity: O(L + (N-K)log(N)) worst case where L is the total number of elements in all the sets, N is the size of the first set, and K is the size of the result set.
	// ACL categories: @read @sortedset @slow
	ZDiff(ctx context.Context, keys ...string) StringSliceCmd

	// ZDiffWithScores
	// Available since: 6.2.0
	// Time complexity: O(L + (N-K)log(N)) worst case where L is the total number of elements in all the sets, N is the size of the first set, and K is the size of the result set.
	// ACL categories: @read @sortedset @slow
	ZDiffWithScores(ctx context.Context, keys ...string) ZSliceCmd

	// ZUnion
	// Available since: 6.2.0
	// Time complexity: O(N)+O(M*log(M)) with N being the sum of the sizes of the input sorted sets, and M being the number of elements in the resulting sorted set.
	// ACL categories: @read @sortedset @slow
	ZUnion(ctx context.Context, store ZStore) StringSliceCmd

	// ZUnionWithScores
	// Available since: 6.2.0
	// Time complexity: O(N)+O(M*log(M)) with N being the sum of the sizes of the input sorted sets, and M being the number of elements in the resulting sorted set.
	// ACL categories: @read @sortedset @slow
	ZUnionWithScores(ctx context.Context, store ZStore) ZSliceCmd
}

type SortedSetCacheCmdable interface {
	// ZCard
	// Available since: 1.2.0
	// Time complexity: O(1)
	// ACL categories: @read @sortedset @fast
	ZCard(ctx context.Context, key string) IntCmd

	// ZCount
	// Available since: 2.0.0
	// Time complexity: O(log(N)) with N being the number of elements in the sorted set.
	// ACL categories: @read @sortedset @fast
	ZCount(ctx context.Context, key, min, max string) IntCmd

	// ZLexCount
	// Available since: 2.8.9
	// Time complexity: O(log(N)) with N being the number of elements in the sorted set.
	// ACL categories: @read @sortedset @fast
	ZLexCount(ctx context.Context, key, min, max string) IntCmd

	// ZMScore
	// Available since: 6.2.0
	// Time complexity: O(N) where N is the number of members being requested.
	// ACL categories: @read @sortedset @fast
	ZMScore(ctx context.Context, key string, members ...string) FloatSliceCmd

	// ZRange
	// Available since: 1.2.0
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements returned.
	// ACL categories: @read @sortedset @slow
	ZRange(ctx context.Context, key string, start, stop int64) StringSliceCmd

	// ZRangeArgs
	// Available since: 1.2.0
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements returned.
	// ACL categories: @read @sortedset @slow
	ZRangeArgs(ctx context.Context, z ZRangeArgs) StringSliceCmd

	// ZRangeArgsWithScores
	// Available since: 1.2.0
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements returned.
	// ACL categories: @read @sortedset @slow
	ZRangeArgsWithScores(ctx context.Context, z ZRangeArgs) ZSliceCmd

	// ZRangeWithScores
	// Available since: 1.2.0
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements returned.
	// ACL categories: @read @sortedset @slow
	ZRangeWithScores(ctx context.Context, key string, start, stop int64) ZSliceCmd

	// ZRangeByLex
	// Available since: 2.8.9
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements being returned. If M is constant (e.g. always asking for the first 10 elements with LIMIT), you can consider it O(log(N)).
	// ACL categories: @read @sortedset @slow
	ZRangeByLex(ctx context.Context, key string, opt ZRangeBy) StringSliceCmd

	// ZRangeByScore
	// Available since: 1.0.5
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements being returned. If M is constant (e.g. always asking for the first 10 elements with LIMIT), you can consider it O(log(N)).
	// ACL categories: @read @sortedset @slow
	ZRangeByScore(ctx context.Context, key string, opt ZRangeBy) StringSliceCmd

	// ZRangeByScoreWithScores
	// Available since: 1.0.5
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements being returned. If M is constant (e.g. always asking for the first 10 elements with LIMIT), you can consider it O(log(N)).
	// ACL categories: @read @sortedset @slow
	ZRangeByScoreWithScores(ctx context.Context, key string, opt ZRangeBy) ZSliceCmd

	// ZRank
	// Available since: 2.0.0
	// Time complexity: O(log(N))
	// ACL categories: @read @sortedset @fast
	ZRank(ctx context.Context, key, member string) IntCmd
	ZRankWithScore(ctx context.Context, key, member string) RankWithScoreCmd

	// ZRevRange
	// Available since: 1.2.0
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements returned.
	// ACL categories: @read @sortedset @slow
	ZRevRange(ctx context.Context, key string, start, stop int64) StringSliceCmd

	// ZRevRangeWithScores
	// Available since: 1.2.0
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements returned.
	// ACL categories: @read @sortedset @slow
	ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) ZSliceCmd

	// ZRevRangeByLex
	// Available since: 2.8.9
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements being returned. If M is constant (e.g. always asking for the first 10 elements with LIMIT), you can consider it O(log(N)).
	// ACL categories: @read @sortedset @slow
	ZRevRangeByLex(ctx context.Context, key string, opt ZRangeBy) StringSliceCmd

	// ZRevRangeByScore
	// Available since: 2.2.0
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements being returned. If M is constant (e.g. always asking for the first 10 elements with LIMIT), you can consider it O(log(N)).
	// ACL categories: @read @sortedset @slow
	ZRevRangeByScore(ctx context.Context, key string, opt ZRangeBy) StringSliceCmd

	// ZRevRangeByScoreWithScores
	// Available since: 2.2.0
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements being returned. If M is constant (e.g. always asking for the first 10 elements with LIMIT), you can consider it O(log(N)).
	// ACL categories: @read @sortedset @slow
	ZRevRangeByScoreWithScores(ctx context.Context, key string, opt ZRangeBy) ZSliceCmd

	// ZRevRank
	// Available since: 2.0.0
	// Time complexity: O(log(N))
	// ACL categories: @read @sortedset @fast
	ZRevRank(ctx context.Context, key, member string) IntCmd
	ZRevRankWithScore(ctx context.Context, key, member string) RankWithScoreCmd

	// ZScore
	// Available since: 1.2.0
	// Time complexity: O(1)
	// ACL categories: @read @sortedset @fast
	ZScore(ctx context.Context, key, member string) FloatCmd
}

func (c *client) BZMPop(ctx context.Context, timeout time.Duration, order string, count int64, keys ...string) ZSliceWithKeyCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandBZMPop, func() []string { return keys })
	r := c.adapter.BZMPop(ctx, timeout, order, count, keys...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) BZPopMax(ctx context.Context, timeout time.Duration, keys ...string) ZWithKeyCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandBZPopMax, func() []string { return keys })
	r := c.adapter.BZPopMax(ctx, timeout, keys...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) BZPopMin(ctx context.Context, timeout time.Duration, keys ...string) ZWithKeyCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandBZPopMin, func() []string { return keys })
	r := c.adapter.BZPopMin(ctx, timeout, keys...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZAdd(ctx context.Context, key string, members ...Z) IntCmd {
	if len(members) > 1 {
		ctx = c.handler.before(ctx, CommandZMAdd)
	} else {
		ctx = c.handler.before(ctx, CommandZAdd)
	}
	r := c.adapter.ZAdd(ctx, key, members...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZAddLT(ctx context.Context, key string, members ...Z) IntCmd {
	ctx = c.handler.before(ctx, CommandZAddLT)
	r := c.adapter.ZAddLT(ctx, key, members...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZAddGT(ctx context.Context, key string, members ...Z) IntCmd {
	ctx = c.handler.before(ctx, CommandZAddGT)
	r := c.adapter.ZAddGT(ctx, key, members...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZAddNX(ctx context.Context, key string, members ...Z) IntCmd {
	ctx = c.handler.before(ctx, CommandZAddNX)
	r := c.adapter.ZAddNX(ctx, key, members...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZAddXX(ctx context.Context, key string, members ...Z) IntCmd {
	ctx = c.handler.before(ctx, CommandZAddXX)
	r := c.adapter.ZAddXX(ctx, key, members...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZAddArgs(ctx context.Context, key string, args ZAddArgs) IntCmd {
	if args.GT {
		ctx = c.handler.before(ctx, CommandZAddGT)
	} else if args.LT {
		ctx = c.handler.before(ctx, CommandZAddLT)
	} else if args.Ch {
		ctx = c.handler.before(ctx, CommandZAddCh)
	} else if args.NX {
		ctx = c.handler.before(ctx, CommandZAddNX)
	} else if args.XX {
		ctx = c.handler.before(ctx, CommandZAddXX)
	} else if len(args.Members) > 1 {
		ctx = c.handler.before(ctx, CommandZMAdd)
	} else {
		ctx = c.handler.before(ctx, CommandZAdd)
	}
	r := c.adapter.ZAddArgs(ctx, key, args)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZAddArgsIncr(ctx context.Context, key string, args ZAddArgs) FloatCmd {
	if args.GT {
		ctx = c.handler.before(ctx, CommandZAddGT)
	} else if args.LT {
		ctx = c.handler.before(ctx, CommandZAddLT)
	} else {
		ctx = c.handler.before(ctx, CommandZAddINCR)
	}
	r := c.adapter.ZAddArgsIncr(ctx, key, args)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZCard(ctx context.Context, key string) IntCmd {
	ctx = c.handler.before(ctx, CommandZCard)
	var r IntCmd
	if c.ttl > 0 {
		r = newIntCmd(c.Do(ctx, c.builder.ZCardCompleted(key)))
	} else {
		r = c.adapter.ZCard(ctx, key)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZCount(ctx context.Context, key, min, max string) IntCmd {
	ctx = c.handler.before(ctx, CommandZCount)
	var r IntCmd
	if c.ttl > 0 {
		r = newIntCmd(c.Do(ctx, c.builder.ZCountCompleted(key, min, max)))
	} else {
		r = c.adapter.ZCount(ctx, key, min, max)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZDiff(ctx context.Context, keys ...string) StringSliceCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandZDiff, func() []string { return keys })
	r := c.adapter.ZDiff(ctx, keys...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZDiffWithScores(ctx context.Context, keys ...string) ZSliceCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandZDiffWithScores, func() []string { return keys })
	r := c.adapter.ZDiffWithScores(ctx, keys...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZDiffStore(ctx context.Context, destination string, keys ...string) IntCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandZDiffStore, func() []string { return appendString(destination, keys...) })
	r := c.adapter.ZDiffStore(ctx, destination, keys...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZIncrBy(ctx context.Context, key string, increment float64, member string) FloatCmd {
	ctx = c.handler.before(ctx, CommandZIncrBy)
	r := c.adapter.ZIncrBy(ctx, key, increment, member)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZInter(ctx context.Context, store ZStore) StringSliceCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandZInter, func() []string { return store.Keys })
	r := c.adapter.ZInter(ctx, store)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZInterWithScores(ctx context.Context, store ZStore) ZSliceCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandZInter, func() []string { return store.Keys })
	r := c.adapter.ZInterWithScores(ctx, store)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZInterStore(ctx context.Context, destination string, store ZStore) IntCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandZInterStore, func() []string { return appendString(destination, store.Keys...) })
	r := c.adapter.ZInterStore(ctx, destination, store)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZInterCard(ctx context.Context, limit int64, keys ...string) IntCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandZInterCard, func() []string { return keys })
	r := c.adapter.ZInterCard(ctx, limit, keys...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZLexCount(ctx context.Context, key, min, max string) IntCmd {
	ctx = c.handler.before(ctx, CommandZLexCount)
	var r IntCmd
	if c.ttl > 0 {
		r = newIntCmd(c.Do(ctx, c.builder.ZLexCountCompleted(key, min, max)))
	} else {
		r = c.adapter.ZLexCount(ctx, key, min, max)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZMPop(ctx context.Context, order string, count int64, keys ...string) ZSliceWithKeyCmd {
	ctx = c.handler.before(ctx, CommandZMPop)
	r := c.adapter.ZMPop(ctx, order, count, keys...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZMScore(ctx context.Context, key string, members ...string) FloatSliceCmd {
	ctx = c.handler.before(ctx, CommandZMScore)
	var r FloatSliceCmd
	if c.ttl > 0 {
		r = newFloatSliceCmd(c.Do(ctx, c.builder.ZMScoreCompleted(key, members...)))
	} else {
		r = c.adapter.ZMScore(ctx, key, members...)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZPopMax(ctx context.Context, key string, count ...int64) ZSliceCmd {
	ctx = c.handler.before(ctx, CommandZPopMax)
	r := c.adapter.ZPopMax(ctx, key, count...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZPopMin(ctx context.Context, key string, count ...int64) ZSliceCmd {
	ctx = c.handler.before(ctx, CommandZPopMin)
	r := c.adapter.ZPopMin(ctx, key, count...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRandMember(ctx context.Context, key string, count int64) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandZRandMember)
	r := c.adapter.ZRandMember(ctx, key, count)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRandMemberWithScores(ctx context.Context, key string, count int64) ZSliceCmd {
	ctx = c.handler.before(ctx, CommandZRandMemberWithScores)
	r := c.adapter.ZRandMemberWithScores(ctx, key, count)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRange(ctx context.Context, key string, start, stop int64) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandZRange)
	var r StringSliceCmd
	if c.ttl > 0 {
		r = newStringSliceCmd(c.Do(ctx, c.builder.ZRangeCompleted(key, start, stop)))
	} else {
		r = c.adapter.ZRange(ctx, key, start, stop)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ZSliceCmd {
	ctx = c.handler.before(ctx, CommandZRange)
	var r ZSliceCmd
	if c.ttl > 0 {
		r = newZSliceCmd(c.Do(ctx, c.builder.ZRangeWithScoresCompleted(key, start, stop)))
	} else {
		r = c.adapter.ZRangeWithScores(ctx, key, start, stop)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRangeArgs(ctx context.Context, z ZRangeArgs) StringSliceCmd {
	if z.Rev || z.ByScore || z.ByLex || z.Offset != 0 || z.Count != 0 {
		ctx = c.handler.before(ctx, CommandZRangeArgsWithOption)
	} else {
		ctx = c.handler.before(ctx, CommandZRangeArgs)
	}
	var r StringSliceCmd
	if c.ttl > 0 {
		r = newStringSliceCmd(c.Do(ctx, c.builder.ZRangeArgsCompleted(z)))
	} else {
		r = c.adapter.ZRangeArgs(ctx, z)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRangeArgsWithScores(ctx context.Context, z ZRangeArgs) ZSliceCmd {
	if z.Rev || z.ByScore || z.ByLex || z.Offset != 0 || z.Count != 0 {
		ctx = c.handler.before(ctx, CommandZRangeArgsWithScoresWithOption)
	} else {
		ctx = c.handler.before(ctx, CommandZRangeArgsWithScores)
	}
	var r ZSliceCmd
	if c.ttl > 0 {
		r = newZSliceCmd(c.Do(ctx, c.builder.ZRangeArgsWithScoresCompleted(z)))
	} else {
		r = c.adapter.ZRangeArgsWithScores(ctx, z)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRangeByLex(ctx context.Context, key string, opt ZRangeBy) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandZRangeByLex)
	var r StringSliceCmd
	if c.ttl > 0 {
		r = newStringSliceCmd(c.Do(ctx, c.builder.ZRangeByLexCompleted(key, opt)))
	} else {
		r = c.adapter.ZRangeByLex(ctx, key, opt)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRangeByScore(ctx context.Context, key string, opt ZRangeBy) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandZRangeByScore)
	var r StringSliceCmd
	if c.ttl > 0 {
		r = newStringSliceCmd(c.Do(ctx, c.builder.ZRangeByScoreCompleted(key, opt)))
	} else {
		r = c.adapter.ZRangeByScore(ctx, key, opt)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRangeByScoreWithScores(ctx context.Context, key string, opt ZRangeBy) ZSliceCmd {
	ctx = c.handler.before(ctx, CommandZRangeByScoreWithScores)
	var r ZSliceCmd
	if c.ttl > 0 {
		r = newZSliceCmd(c.Do(ctx, c.builder.ZRangeByScoreWithScoresCompleted(key, opt)))
	} else {
		r = c.adapter.ZRangeByScoreWithScores(ctx, key, opt)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRangeStore(ctx context.Context, dst string, z ZRangeArgs) IntCmd {
	ctx = c.handler.before(ctx, CommandZRangeStore)
	r := c.adapter.ZRangeStore(ctx, dst, z)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRank(ctx context.Context, key, member string) IntCmd {
	ctx = c.handler.before(ctx, CommandZRank)
	var r IntCmd
	if c.ttl > 0 {
		r = newIntCmd(c.Do(ctx, c.builder.ZRankCompleted(key, member)))
	} else {
		r = c.adapter.ZRank(ctx, key, member)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRankWithScore(ctx context.Context, key, member string) RankWithScoreCmd {
	ctx = c.handler.before(ctx, CommandZRank)
	var r RankWithScoreCmd
	if c.ttl > 0 {
		r = newRankWithScoreCmd(c.Do(ctx, c.builder.ZRankWithScoreCompleted(key, member)))
	} else {
		r = c.adapter.ZRankWithScore(ctx, key, member)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRem(ctx context.Context, key string, members ...any) IntCmd {
	if len(members) > 1 {
		ctx = c.handler.before(ctx, CommandZMRem)
	} else {
		ctx = c.handler.before(ctx, CommandZRem)
	}
	r := c.adapter.ZRem(ctx, key, members...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRemRangeByLex(ctx context.Context, key, min, max string) IntCmd {
	ctx = c.handler.before(ctx, CommandZRemRangeByLex)
	r := c.adapter.ZRemRangeByLex(ctx, key, min, max)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRemRangeByRank(ctx context.Context, key string, start, stop int64) IntCmd {
	ctx = c.handler.before(ctx, CommandZRemRangeByRank)
	r := c.adapter.ZRemRangeByRank(ctx, key, start, stop)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRemRangeByScore(ctx context.Context, key, min, max string) IntCmd {
	ctx = c.handler.before(ctx, CommandZRemRangeByScore)
	r := c.adapter.ZRemRangeByScore(ctx, key, min, max)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRevRange(ctx context.Context, key string, start, stop int64) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandZRevRange)
	var r StringSliceCmd
	if c.ttl > 0 {
		r = newStringSliceCmd(c.Do(ctx, c.builder.ZRevRangeCompleted(key, start, stop)))
	} else {
		r = c.adapter.ZRevRange(ctx, key, start, stop)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) ZSliceCmd {
	ctx = c.handler.before(ctx, CommandZRevRange)
	var r ZSliceCmd
	if c.ttl > 0 {
		r = newZSliceCmd(c.Do(ctx, c.builder.ZRevRangeWithScoresCompleted(key, start, stop)))
	} else {
		r = c.adapter.ZRevRangeWithScores(ctx, key, start, stop)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRevRangeByLex(ctx context.Context, key string, opt ZRangeBy) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandZRevRangeByLex)
	var r StringSliceCmd
	if c.ttl > 0 {
		r = newStringSliceCmd(c.Do(ctx, c.builder.ZRevRangeByLexCompleted(key, opt)))
	} else {
		r = c.adapter.ZRevRangeByLex(ctx, key, opt)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRevRangeByScore(ctx context.Context, key string, opt ZRangeBy) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandZRevRangeByScore)
	var r StringSliceCmd
	if c.ttl > 0 {
		r = newStringSliceCmd(c.Do(ctx, c.builder.ZRevRangeByScoreCompleted(key, opt)))
	} else {
		r = c.adapter.ZRevRangeByScore(ctx, key, opt)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRevRangeByScoreWithScores(ctx context.Context, key string, opt ZRangeBy) ZSliceCmd {
	ctx = c.handler.before(ctx, CommandZRevRangeByScore)
	var r ZSliceCmd
	if c.ttl > 0 {
		r = newZSliceCmd(c.Do(ctx, c.builder.ZRevRangeByScoreWithScoresCompleted(key, opt)))
	} else {
		r = c.adapter.ZRevRangeByScoreWithScores(ctx, key, opt)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRevRank(ctx context.Context, key, member string) IntCmd {
	ctx = c.handler.before(ctx, CommandZRevRank)
	var r IntCmd
	if c.ttl > 0 {
		r = newIntCmd(c.Do(ctx, c.builder.ZRevRankCompleted(key, member)))
	} else {
		r = c.adapter.ZRevRank(ctx, key, member)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRevRankWithScore(ctx context.Context, key, member string) RankWithScoreCmd {
	ctx = c.handler.before(ctx, CommandZRevRank)
	var r RankWithScoreCmd
	if c.ttl > 0 {
		r = newRankWithScoreCmd(c.Do(ctx, c.builder.ZRevRankWithScoreCompleted(key, member)))
	} else {
		r = c.adapter.ZRevRankWithScore(ctx, key, member)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZScan(ctx context.Context, key string, cursor uint64, match string, count int64) ScanCmd {
	ctx = c.handler.before(ctx, CommandZScan)
	r := c.adapter.ZScan(ctx, key, cursor, match, count)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZScore(ctx context.Context, key, member string) FloatCmd {
	ctx = c.handler.before(ctx, CommandZScore)
	var r FloatCmd
	if c.ttl > 0 {
		r = newFloatCmd(c.Do(ctx, c.builder.ZScoreCompleted(key, member)))
	} else {
		r = c.adapter.ZScore(ctx, key, member)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZUnion(ctx context.Context, store ZStore) StringSliceCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandZUnion, func() []string { return store.Keys })
	r := c.adapter.ZUnion(ctx, store)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZUnionWithScores(ctx context.Context, store ZStore) ZSliceCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandZUnion, func() []string { return store.Keys })
	r := c.adapter.ZUnionWithScores(ctx, store)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZUnionStore(ctx context.Context, dest string, store ZStore) IntCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandZUnionStore, func() []string { return appendString(dest, store.Keys...) })
	r := c.adapter.ZUnionStore(ctx, dest, store)
	c.handler.after(ctx, r.Err())
	return r
}
