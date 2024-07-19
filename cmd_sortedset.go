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
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- Nil reply: when no element could be popped.
	//		- Array reply: a two-element array with the first element being the name of the key from which elements were popped,
	//			and the second element is an array of the popped elements. Every entry in the elements array is also an array that contains the member and its score.
	BZMPop(ctx context.Context, timeout time.Duration, order string, count int64, keys ...string) ZSliceWithKeyCmd

	// BZPopMax
	// Available since: 5.0.0
	// Time complexity: O(log(N)) with N being the number of elements in the sorted set.
	// ACL categories: @write @sortedset @fast @blocking
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- Nil reply: when no element could be popped and the timeout expired.
	//		- Array reply: the keyname, popped member, and its score.
	// History:
	//	- Starting with Redis version 6.0.0: timeout is interpreted as a double instead of an integer.
	BZPopMax(ctx context.Context, timeout time.Duration, keys ...string) ZWithKeyCmd

	// BZPopMin
	// Available since: 5.0.0
	// Time complexity: O(log(N)) with N being the number of elements in the sorted set.
	// ACL categories: @write @sortedset @fast @blocking
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- Nil reply: when no element could be popped and the timeout expired.
	//		- Array reply: the keyname, popped member, and its score.
	// History:
	//	- Starting with Redis version 6.0.0: timeout is interpreted as a double instead of an integer.
	BZPopMin(ctx context.Context, timeout time.Duration, keys ...string) ZWithKeyCmd

	// ZAdd
	// Available since: 1.2.0
	// Time complexity: O(log(N)) for each item added, where N is the number of elements in the sorted set.
	// ACL categories: @write @sortedset @fast
	// Options
	// 	- XX: Only update elements that already exist. Don't add new elements.
	// 	- NX: Only add new elements. Don't update already existing elements.
	// 	- LT: Only update existing elements if the new score is less than the current score. This flag doesn't prevent adding new elements.
	// 	- GT: Only update existing elements if the new score is greater than the current score. This flag doesn't prevent adding new elements.
	// 	- CH: Modify the return value from the number of new elements added, to the total number of elements changed (CH is an abbreviation of changed).
	//		Changed elements are new elements added and elements already existing for which the score was updated. So elements specified in the command
	//		line having the same score as they had in the past are not counted. Note: normally the return value of ZADD only counts the number of new elements added.
	// 	- INCR: When this option is specified ZADD acts like ZINCRBY. Only one score-element pair can be specified in this mode.
	// RESP2 Reply:
	//	One of the following:
	//		- Nil reply: if the operation was aborted because of a conflict with one of the XX/NX/LT/GT options.
	//		- Integer reply: the number of new members when the CH option is not used.
	//		- Integer reply: the number of new or updated members when the CH option is used.
	//		- Bulk string reply: the updated score of the member when the INCR option is used.
	// RESP3 Reply:
	//	One of the following:
	//		- Null reply: if the operation was aborted because of a conflict with one of the XX/NX/LT/GT options.
	//		- Integer reply: the number of new members when the CH option is not used.
	//		- Integer reply: the number of new or updated members when the CH option is used.
	//		- Double reply: the updated score of the member when the INCR option is used.
	// History:
	//	- Starting with Redis version 2.4.0: Accepts multiple elements.
	//	- Starting with Redis version 3.0.2: Added the XX, NX, CH and INCR options.
	//	- Starting with Redis version 6.2.0: Added the GT and LT options.
	ZAdd(ctx context.Context, key string, members ...Z) IntCmd
	ZAddNX(ctx context.Context, key string, members ...Z) IntCmd
	ZAddXX(ctx context.Context, key string, members ...Z) IntCmd
	ZAddLT(ctx context.Context, key string, members ...Z) IntCmd
	ZAddGT(ctx context.Context, key string, members ...Z) IntCmd
	ZAddArgsIncr(ctx context.Context, key string, args ZAddArgs) FloatCmd
	ZAddArgs(ctx context.Context, key string, args ZAddArgs) IntCmd

	// ZDiffStore
	// Available since: 6.2.0
	// Time complexity: O(L + (N-K)log(N)) worst case where L is the total number of elements in all the sets, N is the size of the first set, and K is the size of the result set.
	// ACL categories: @write @sortedset @slow
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the number of members in the resulting sorted set at destination.
	ZDiffStore(ctx context.Context, destination string, keys ...string) IntCmd

	// ZIncrBy
	// Available since: 1.2.0
	// Time complexity: O(log(N)) where N is the number of elements in the sorted set.
	// ACL categories: @write @sortedset @fast
	// RESP2 Reply:
	// 	- Bulk string reply: the new score of member as a double precision floating point number.
	// RESP3 Reply:
	// 	- Double reply: the new score of member.
	ZIncrBy(ctx context.Context, key string, increment float64, member string) FloatCmd

	// ZInterStore
	// Available since: 2.0.0
	// Time complexity: O(NK)+O(Mlog(M)) worst case with N being the smallest input sorted set, K being the number of input sorted sets and M being the number of elements in the resulting sorted set.
	// ACL categories: @write @sortedset @slow
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the number of members in the resulting sorted set at the destination.
	ZInterStore(ctx context.Context, destination string, store ZStore) IntCmd

	// ZMPop
	// Available since: 7.0.0
	// Time complexity: O(K) + O(M*log(N)) where K is the number of provided keys, N being the number of elements in the sorted set, and M being the number of elements popped.
	// ACL categories: @write @sortedset @slow
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- Nil reply: when no element could be popped.
	//		- Array reply: A two-element array with the first element being the name of the key from which elements were popped,
	//			and the second element is an array of the popped elements. Every entry in the elements array is also an array that contains the member and its score.
	ZMPop(ctx context.Context, order string, count int64, keys ...string) ZSliceWithKeyCmd

	// ZPopMax
	// Available since: 5.0.0
	// Time complexity: O(log(N)*M) with N being the number of elements in the sorted set, and M being the number of elements popped.
	// ACL categories: @write @sortedset @fast
	// RESP2 / RESP3 Reply:
	// 	- Array reply: a list of popped elements and scores.
	ZPopMax(ctx context.Context, key string, count ...int64) ZSliceCmd

	// ZPopMin
	// Available since: 5.0.0
	// Time complexity: O(log(N)*M) with N being the number of elements in the sorted set, and M being the number of elements popped.
	// ACL categories: @write @sortedset @fast
	// RESP2 / RESP3 Reply:
	// 	- Array reply: a list of popped elements and scores.
	ZPopMin(ctx context.Context, key string, count ...int64) ZSliceCmd

	// ZRangeStore
	// Available since: 6.2.0
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements stored into the destination key.
	// ACL categories: @write @sortedset @slow
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the number of elements in the resulting sorted set.
	ZRangeStore(ctx context.Context, dst string, z ZRangeArgs) IntCmd

	// ZRem
	// Available since: 1.2.0
	// Time complexity: O(M*log(N)) with N being the number of elements in the sorted set and M the number of elements to be removed.
	// ACL categories: @write @sortedset @fast
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the number of members removed from the sorted set, not including non-existing members.
	// History:
	//	- Starting with Redis version 2.4.0: Accepts multiple elements.
	ZRem(ctx context.Context, key string, members ...any) IntCmd

	// ZRemRangeByLex
	// Available since: 2.8.9
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements removed by the operation.
	// ACL categories: @write @sortedset @slow
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the number of members removed.
	ZRemRangeByLex(ctx context.Context, key, min, max string) IntCmd

	// ZRemRangeByRank
	// Available since: 2.0.0
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements removed by the operation.
	// ACL categories: @write @sortedset @slow
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the number of members removed.
	ZRemRangeByRank(ctx context.Context, key string, start, stop int64) IntCmd

	// ZRemRangeByScore
	// Available since: 1.2.0
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements removed by the operation.
	// ACL categories: @write @sortedset @slow
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the number of members removed.
	ZRemRangeByScore(ctx context.Context, key, min, max string) IntCmd

	// ZUnionStore
	// Available since: 2.0.0
	// Time complexity: O(N)+O(M log(M)) with N being the sum of the sizes of the input sorted sets, and M being the number of elements in the resulting sorted set.
	// ACL categories: @write @sortedset @slow
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the number of elements in the resulting sorted set.
	ZUnionStore(ctx context.Context, dest string, store ZStore) IntCmd
}

type SortedSetReader interface {
	// ZInter
	// Available since: 6.2.0
	// Time complexity: O(NK)+O(Mlog(M)) worst case with N being the smallest input sorted set, K being the number of input sorted sets and M being the number of elements in the resulting sorted set.
	// ACL categories: @read @sortedset @slow
	// RESP2 / RESP3 Reply:
	// 	- Array reply: the result of the intersection including, optionally, scores when the WITHSCORES option is used.
	ZInter(ctx context.Context, store ZStore) StringSliceCmd
	ZInterWithScores(ctx context.Context, store ZStore) ZSliceCmd

	// ZInterCard
	// Available since: 7.0.0
	// Time complexity: O(N*K) worst case with N being the smallest input sorted set, K being the number of input sorted sets.
	// ACL categories: @read @sortedset @slow
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the number of members in the resulting intersection.
	ZInterCard(ctx context.Context, limit int64, keys ...string) IntCmd

	// ZRandMember
	// Available since: 6.2.0
	// Time complexity: O(N) where N is the number of elements returned
	// ACL categories: @read @sortedset @slow
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- Bulk string reply: without the additional count argument, the command returns a randomly selected member
	//		- Null reply when key doesn't exist
	//		- Array reply: when the additional count argument is passed, the command returns an array of members, or an empty array when key doesn't exist.
	//			If the WITHSCORES modifier is used, the reply is a list of members and their scores from the sorted set.
	ZRandMember(ctx context.Context, key string, count int64) StringSliceCmd
	ZRandMemberWithScores(ctx context.Context, key string, count int64) ZSliceCmd

	// ZScan
	// Available since: 2.8.0
	// Time complexity: O(1) for every call. O(N) for a complete iteration, including enough command calls for the cursor to return back to 0.
	//	N is the number of elements inside the collection.
	// ACL categories: @read @sortedset @slow
	// RESP2 / RESP3 Reply:
	// 	- Array reply: cursor and scan response in array form.
	ZScan(ctx context.Context, key string, cursor uint64, match string, count int64) ScanCmd

	// ZDiff
	// Available since: 6.2.0
	// Time complexity: O(L + (N-K)log(N)) worst case where L is the total number of elements in all the sets, N is the size of the first set, and K is the size of the result set.
	// ACL categories: @read @sortedset @slow
	// RESP2 / RESP3 Reply:
	// 	- Array reply: the result of the difference including, optionally, scores when the WITHSCORES option is used.
	ZDiff(ctx context.Context, keys ...string) StringSliceCmd
	ZDiffWithScores(ctx context.Context, keys ...string) ZSliceCmd

	// ZUnion
	// Available since: 6.2.0
	// Time complexity: O(N)+O(M*log(M)) with N being the sum of the sizes of the input sorted sets, and M being the number of elements in the resulting sorted set.
	// ACL categories: @read @sortedset @slow
	// RESP2 / RESP3 Reply:
	// 	- Array reply: the result of the union with, optionally, their scores when WITHSCORES is used.
	ZUnion(ctx context.Context, store ZStore) StringSliceCmd
	ZUnionWithScores(ctx context.Context, store ZStore) ZSliceCmd
}

type SortedSetCacheCmdable interface {
	// ZCard
	// Available since: 1.2.0
	// Time complexity: O(1)
	// ACL categories: @read @sortedset @fast
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the cardinality (number of members) of the sorted set, or 0 if the key doesn't exist.
	ZCard(ctx context.Context, key string) IntCmd

	// ZCount
	// Available since: 2.0.0
	// Time complexity: O(log(N)) with N being the number of elements in the sorted set.
	// ACL categories: @read @sortedset @fast
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the number of members in the specified score range.
	ZCount(ctx context.Context, key, min, max string) IntCmd

	// ZLexCount
	// Available since: 2.8.9
	// Time complexity: O(log(N)) with N being the number of elements in the sorted set.
	// ACL categories: @read @sortedset @fast
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the number of members in the specified score range.
	ZLexCount(ctx context.Context, key, min, max string) IntCmd

	// ZMScore
	// Available since: 6.2.0
	// Time complexity: O(N) where N is the number of members being requested.
	// ACL categories: @read @sortedset @fast
	// RESP2 Reply:
	// 	- Nil reply: if the member does not exist in the sorted set.
	// 	- Array reply: a list of Bulk string reply member scores as double-precision floating point numbers.
	// RESP3 Reply:
	// 	- Null reply: if the member does not exist in the sorted set.
	// 	- Array reply: a list of Double reply member scores as double-precision floating point numbers.
	ZMScore(ctx context.Context, key string, members ...string) FloatSliceCmd

	// ZRange
	// Available since: 1.2.0
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements returned.
	// ACL categories: @read @sortedset @slow
	// RESP2 / RESP3 Reply:
	// 	- Array reply: a list of members in the specified range with, optionally, their scores when the WITHSCORES option is given.
	// History:
	//	- Starting with Redis version 6.2.0: Added the REV, BYSCORE, BYLEX and LIMIT options.
	ZRange(ctx context.Context, key string, start, stop int64) StringSliceCmd
	ZRangeArgs(ctx context.Context, z ZRangeArgs) StringSliceCmd
	ZRangeArgsWithScores(ctx context.Context, z ZRangeArgs) ZSliceCmd
	ZRangeWithScores(ctx context.Context, key string, start, stop int64) ZSliceCmd

	// ZRangeByLex
	// Available since: 2.8.9
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements being returned.
	//	If M is constant (e.g. always asking for the first 10 elements with LIMIT), you can consider it O(log(N)).
	// ACL categories: @read @sortedset @slow
	// RESP2 / RESP3 Reply:
	// 	- Array reply: a list of elements in the specified score range.
	ZRangeByLex(ctx context.Context, key string, opt ZRangeBy) StringSliceCmd

	// ZRangeByScore
	// Available since: 1.0.5
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements being returned.
	//	If M is constant (e.g. always asking for the first 10 elements with LIMIT), you can consider it O(log(N)).
	// ACL categories: @read @sortedset @slow
	// RESP2 / RESP3 Reply:
	// 	- Array reply: a list of the members with, optionally, their scores in the specified score range.
	// History:
	//	- Starting with Redis version 2.0.0: Added the WITHSCORES modifier.
	ZRangeByScore(ctx context.Context, key string, opt ZRangeBy) StringSliceCmd
	ZRangeByScoreWithScores(ctx context.Context, key string, opt ZRangeBy) ZSliceCmd

	// ZRank
	// Available since: 2.0.0
	// Time complexity: O(log(N))
	// ACL categories: @read @sortedset @fast
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- Null reply: if the key does not exist or the member does not exist in the sorted set.
	//		-Integer reply: the rank of the member when WITHSCORE is not used.
	//		-Array reply: the rank and score of the member when WITHSCORE is used.
	// History:
	//	- Starting with Redis version 7.2.0: Added the optional WITHSCORE argument.
	ZRank(ctx context.Context, key, member string) IntCmd
	ZRankWithScore(ctx context.Context, key, member string) RankWithScoreCmd

	// ZRevRange
	// Available since: 1.2.0
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements returned.
	// ACL categories: @read @sortedset @slow
	// RESP2 / RESP3 Reply:
	// 	- Array reply: a list of members in the specified range, optionally with their scores if WITHSCORE was used.
	ZRevRange(ctx context.Context, key string, start, stop int64) StringSliceCmd

	// ZRevRangeWithScores
	// Available since: 1.2.0
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements returned.
	// ACL categories: @read @sortedset @slow
	ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) ZSliceCmd

	// ZRevRangeByLex
	// Available since: 2.8.9
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements being returned.
	//	If M is constant (e.g. always asking for the first 10 elements with LIMIT), you can consider it O(log(N)).
	// ACL categories: @read @sortedset @slow
	// RESP2 / RESP3 Reply:
	// 	- Array reply: a list of members in the specified score range.
	ZRevRangeByLex(ctx context.Context, key string, opt ZRangeBy) StringSliceCmd

	// ZRevRangeByScore
	// Available since: 2.2.0
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements being returned.
	//	If M is constant (e.g. always asking for the first 10 elements with LIMIT), you can consider it O(log(N)).
	// ACL categories: @read @sortedset @slow
	// RESP2 / RESP3 Reply:
	// 	- Array reply: a list of the members and, optionally, their scores in the specified score range.
	// History:
	//	- Starting with Redis version 2.1.6: min and max can be exclusive.
	ZRevRangeByScore(ctx context.Context, key string, opt ZRangeBy) StringSliceCmd
	ZRevRangeByScoreWithScores(ctx context.Context, key string, opt ZRangeBy) ZSliceCmd

	// ZRevRank
	// Available since: 2.0.0
	// Time complexity: O(log(N))
	// ACL categories: @read @sortedset @fast
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- Null reply: if the key does not exist or the member does not exist in the sorted set.
	//		- Integer reply: The rank of the member when WITHSCORE is not used.
	//		- Array reply: The rank and score of the member when WITHSCORE is used.
	// History:
	//	- Starting with Redis version 7.2.0: Added the optional WITHSCORE argument.
	ZRevRank(ctx context.Context, key, member string) IntCmd
	ZRevRankWithScore(ctx context.Context, key, member string) RankWithScoreCmd

	// ZScore
	// Available since: 1.2.0
	// Time complexity: O(1)
	// ACL categories: @read @sortedset @fast
	// RESP2 Reply:
	//	One of the following:
	//		- Bulk string reply: the score of the member (a double-precision floating point number), represented as a string.
	//		- Nil reply: if member does not exist in the sorted set, or the key does not exist.
	// RESP3 Reply:
	//	One of the following:
	//		- Double reply: the score of the member (a double-precision floating point number).
	//		- Nil reply: if member does not exist in the sorted set, or the key does not exist.
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
