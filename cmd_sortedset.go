package sandwich_redis

import (
	"context"
	"time"
)

type SortedSetCmdable interface {
	SortedSetWriter
	SortedSetReader
}

type SortedSetWriter interface {
	// BZPopMax
	// Available since: 5.0.0
	// Time complexity: O(log(N)) with N being the number of elements in the sorted set.
	// ACL categories: @write @sortedset @fast @blocking
	// BZPOPMAX is the blocking variant of the sorted set ZPOPMAX primitive.
	// It is the blocking version because it blocks the connection when there are no members to pop from any of the given sorted sets. A member with the highest score is popped from first sorted set that is non-empty, with the given keys being checked in the order that they are given.
	// The timeout argument is interpreted as a double value specifying the maximum number of seconds to block. A timeout of zero can be used to block indefinitely.
	// See the BZPOPMIN documentation for the exact semantics, since BZPOPMAX is identical to BZPOPMIN with the only difference being that it pops members with the highest scores instead of popping the ones with the lowest scores.
	// Return:
	// Array reply: specifically:
	// 	A nil multi-bulk when no element could be popped and the timeout expired.
	// 	A three-element multi-bulk with the first element being the name of the key where a member was popped, the second element is the popped member itself, and the third element is the score of the popped element.
	BZPopMax(ctx context.Context, timeout time.Duration, keys ...string) ZWithKeyCmd

	// BZPopMin
	// Available since: 5.0.0
	// Time complexity: O(log(N)) with N being the number of elements in the sorted set.
	// ACL categories: @write @sortedset @fast @blocking
	// BZPOPMIN is the blocking variant of the sorted set ZPOPMIN primitive.
	// It is the blocking version because it blocks the connection when there are no members to pop from any of the given sorted sets. A member with the lowest score is popped from first sorted set that is non-empty, with the given keys being checked in the order that they are given.
	// The timeout argument is interpreted as a double value specifying the maximum number of seconds to block. A timeout of zero can be used to block indefinitely.
	// See the BLPOP documentation for the exact semantics, since BZPOPMIN is identical to BLPOP with the only difference being the data structure being popped from.
	// Return:
	// Array reply: specifically:
	//	A nil multi-bulk when no element could be popped and the timeout expired.
	//	A three-element multi-bulk with the first element being the name of the key where a member was popped, the second element is the popped member itself, and the third element is the score of the popped element.
	BZPopMin(ctx context.Context, timeout time.Duration, keys ...string) ZWithKeyCmd

	// ZAdd
	// Available since: 1.2.0
	// Time complexity: O(log(N)) for each item added, where N is the number of elements in the sorted set.
	// ACL categories: @write @sortedset @fast
	// Adds all the specified members with the specified scores to the sorted set stored at key. It is possible to specify multiple score / member pairs. If a specified member is already a member of the sorted set, the score is updated and the element reinserted at the right position to ensure the correct ordering.
	// If key does not exist, a new sorted set with the specified members as sole members is created, like if the sorted set was empty. If the key exists but does not hold a sorted set, an error is returned.
	// The score values should be the string representation of a double precision floating point number. +inf and -inf values are valid values as well.
	// Return:
	// Integer reply, specifically:
	// When used without optional arguments, the number of elements added to the sorted set (excluding score updates).
	// 	If the CH option is specified, the number of elements that were changed (added or updated).
	// 	If the INCR option is specified, the return value will be Bulk string reply:
	// 	The new score of member (a double precision floating point number) represented as string, or nil if the operation was aborted (when called with either the XX or the NX option).
	// See https://redis.io/commands/zadd/
	ZAdd(ctx context.Context, key string, members ...Z) IntCmd

	// ZAddNX
	// Available since: 3.0.2
	// Time complexity: O(log(N)) for each item added, where N is the number of elements in the sorted set.
	// ACL categories: @write @sortedset @fast
	// See https://redis.io/commands/zadd/
	ZAddNX(ctx context.Context, key string, members ...Z) IntCmd

	// ZAddXX
	// Available since: 3.0.2
	// Time complexity: O(log(N)) for each item added, where N is the number of elements in the sorted set.
	// ACL categories: @write @sortedset @fast
	// See https://redis.io/commands/zadd/
	ZAddXX(ctx context.Context, key string, members ...Z) IntCmd

	// ZAddCh
	// Available since:3.0.2
	// Time complexity: O(log(N)) for each item added, where N is the number of elements in the sorted set.
	// ACL categories: @write @sortedset @fast
	// See https://redis.io/commands/zadd/
	ZAddCh(ctx context.Context, key string, members ...Z) IntCmd

	// ZAddNXCh
	// Available since: 3.0.2
	// Time complexity: O(log(N)) for each item added, where N is the number of elements in the sorted set.
	// ACL categories: @write @sortedset @fast
	// See https://redis.io/commands/zadd/
	ZAddNXCh(ctx context.Context, key string, members ...Z) IntCmd

	// ZAddXXCh
	// Available since: 3.0.2
	// Time complexity: O(log(N)) for each item added, where N is the number of elements in the sorted set.
	// ACL categories: @write @sortedset @fast
	// See https://redis.io/commands/zadd/
	ZAddXXCh(ctx context.Context, key string, members ...Z) IntCmd

	// ZAddArgs
	// Available since: 3.0.2
	// Time complexity: O(log(N)) for each item added, where N is the number of elements in the sorted set.
	// ACL categories: @write @sortedset @fast
	// Starting with Redis version 6.2.0: Added the GT and LT options.
	// See https://redis.io/commands/zadd/
	ZAddArgs(ctx context.Context, key string, args ZAddArgs) IntCmd

	// ZAddArgsIncr
	// Available since: 3.0.2
	// Time complexity: O(log(N)) for each item added, where N is the number of elements in the sorted set.
	// ACL categories: @write @sortedset @fast
	// Starting with Redis version 6.2.0: Added the GT and LT options.
	// See https://redis.io/commands/zadd/
	ZAddArgsIncr(ctx context.Context, key string, args ZAddArgs) FloatCmd

	// ZIncr
	// Available since: 3.0.2
	// Time complexity: O(log(N)) for each item added, where N is the number of elements in the sorted set.
	// ACL categories: @write @sortedset @fast
	// Starting with Redis version 6.2.0: Added the GT and LT options.
	// See https://redis.io/commands/zadd/
	ZIncr(ctx context.Context, key string, member Z) FloatCmd

	// ZIncrNX
	// Available since: 3.0.2
	// Time complexity: O(log(N)) for each item added, where N is the number of elements in the sorted set.
	// ACL categories: @write @sortedset @fast
	// Starting with Redis version 6.2.0: Added the GT and LT options.
	// See https://redis.io/commands/zadd/
	ZIncrNX(ctx context.Context, key string, member Z) FloatCmd

	// ZIncrXX
	// Available since: 3.0.2
	// Time complexity: O(log(N)) for each item added, where N is the number of elements in the sorted set.
	// ACL categories: @write @sortedset @fast
	// Starting with Redis version 6.2.0: Added the GT and LT options.
	// See https://redis.io/commands/zadd/
	ZIncrXX(ctx context.Context, key string, member Z) FloatCmd

	// ZDiffStore
	// Available since: 6.2.0
	// Time complexity: O(L + (N-K)log(N)) worst case where L is the total number of elements in all the sets, N is the size of the first set, and K is the size of the result set.
	// ACL categories: @write @sortedset @slow
	// Computes the difference between the first and all successive input sorted sets and stores the result in destination. The total number of input keys is specified by numkeys.
	// Keys that do not exist are considered to be empty sets.
	// If destination already exists, it is overwritten.
	// Return:
	//	Integer reply: the number of elements in the resulting sorted set at destination.
	ZDiffStore(ctx context.Context, destination string, keys ...string) IntCmd

	// ZIncrBy
	// Available since: 1.2.0
	// Time complexity: O(log(N)) where N is the number of elements in the sorted set.
	// ACL categories: @write @sortedset @fast
	// Increments the score of member in the sorted set stored at key by increment. If member does not exist in the sorted set, it is added with increment as its score (as if its previous score was 0.0). If key does not exist, a new sorted set with the specified member as its sole member is created.
	// An error is returned when key exists but does not hold a sorted set.
	// The score value should be the string representation of a numeric value, and accepts double precision floating point numbers. It is possible to provide a negative value to decrement the score.
	// Return:
	// 	Bulk string reply: the new score of member (a double precision floating point number), represented as string.
	ZIncrBy(ctx context.Context, key string, increment float64, member string) FloatCmd

	// ZInterStore
	// Available since: 2.0.0
	// Time complexity: O(NK)+O(Mlog(M)) worst case with N being the smallest input sorted set, K being the number of input sorted sets and M being the number of elements in the resulting sorted set.
	// ACL categories: @write @sortedset @slow
	// Computes the intersection of numkeys sorted sets given by the specified keys, and stores the result in destination. It is mandatory to provide the number of input keys (numkeys) before passing the input keys and the other (optional) arguments.
	// By default, the resulting score of an element is the sum of its scores in the sorted sets where it exists. Because intersection requires an element to be a member of every given sorted set, this results in the score of every element in the resulting sorted set to be equal to the number of input sorted sets.
	// For a description of the WEIGHTS and AGGREGATE options, see ZUNIONSTORE.
	// If destination already exists, it is overwritten.
	// Return:
	//	Integer reply: the number of elements in the resulting sorted set at destination.
	ZInterStore(ctx context.Context, destination string, store ZStore) IntCmd

	// ZPopMax
	// Available since: 5.0.0
	// Time complexity: O(log(N)*M) with N being the number of elements in the sorted set, and M being the number of elements popped.
	// ACL categories: @write @sortedset @fast
	// Removes and returns up to count members with the highest scores in the sorted set stored at key.
	// When left unspecified, the default value for count is 1. Specifying a count value that is higher than the sorted set's cardinality will not produce an error. When returning multiple elements, the one with the highest score will be the first, followed by the elements with lower scores.
	// Return:
	// 	Array reply: list of popped elements and scores.
	ZPopMax(ctx context.Context, key string, count ...int64) ZSliceCmd

	// ZPopMin
	// Available since: 5.0.0
	// Time complexity: O(log(N)*M) with N being the number of elements in the sorted set, and M being the number of elements popped.
	// ACL categories: @write @sortedset @fast
	// Removes and returns up to count members with the lowest scores in the sorted set stored at key.
	// When left unspecified, the default value for count is 1. Specifying a count value that is higher than the sorted set's cardinality will not produce an error. When returning multiple elements, the one with the lowest score will be the first, followed by the elements with greater scores.
	// Return:
	//	Array reply: list of popped elements and scores.
	ZPopMin(ctx context.Context, key string, count ...int64) ZSliceCmd

	// ZRangeStore
	// Available since: 6.2.0
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements stored into the destination key.
	// ACL categories: @write @sortedset @slow
	// This command is like ZRANGE, but stores the result in the <dst> destination key.
	// Return:
	//	Integer reply: the number of elements in the resulting sorted set.
	ZRangeStore(ctx context.Context, dst string, z ZRangeArgs) IntCmd

	// ZRem
	// Available since: 1.2.0
	// Time complexity: O(M*log(N)) with N being the number of elements in the sorted set and M the number of elements to be removed.
	// ACL categories: @write @sortedset @fast
	// Removes the specified members from the sorted set stored at key. Non existing members are ignored.
	// An error is returned when key exists and does not hold a sorted set.
	// Return:
	// 	Integer reply, specifically:
	// 	The number of members removed from the sorted set, not including non existing members.
	ZRem(ctx context.Context, key string, members ...interface{}) IntCmd

	// ZRemRangeByLex
	// Available since: 2.8.9
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements removed by the operation.
	// ACL categories: @write @sortedset @slow
	// When all the elements in a sorted set are inserted with the same score, in order to force lexicographical ordering, this command removes all elements in the sorted set stored at key between the lexicographical range specified by min and max.
	// The meaning of min and max are the same of the ZRANGEBYLEX command. Similarly, this command actually removes the same elements that ZRANGEBYLEX would return if called with the same min and max arguments.
	// Return:
	// 	Integer reply: the number of elements removed.
	ZRemRangeByLex(ctx context.Context, key, min, max string) IntCmd

	// ZRemRangeByRank
	// Available since: 2.0.0
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements removed by the operation.
	// ACL categories: @write @sortedset @slow
	// Removes all elements in the sorted set stored at key with rank between start and stop. Both start and stop are 0 -based indexes with 0 being the element with the lowest score. These indexes can be negative numbers, where they indicate offsets starting at the element with the highest score. For example: -1 is the element with the highest score, -2 the element with the second highest score and so forth.
	// Return:
	// 	Integer reply: the number of elements removed.
	ZRemRangeByRank(ctx context.Context, key string, start, stop int64) IntCmd

	// ZRemRangeByScore
	// Available since: 1.2.0
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements removed by the operation.
	// ACL categories: @write @sortedset @slow
	// Removes all elements in the sorted set stored at key with a score between min and max (inclusive).
	// Return:
	// 	Integer reply: the number of elements removed.
	ZRemRangeByScore(ctx context.Context, key, min, max string) IntCmd

	// ZUnionStore
	// Available since: 2.0.0
	// Time complexity: O(N)+O(M log(M)) with N being the sum of the sizes of the input sorted sets, and M being the number of elements in the resulting sorted set.
	// ACL categories: @write @sortedset @slow
	// Computes the union of numkeys sorted sets given by the specified keys, and stores the result in destination. It is mandatory to provide the number of input keys (numkeys) before passing the input keys and the other (optional) arguments.
	// By default, the resulting score of an element is the sum of its scores in the sorted sets where it exists.
	// Using the WEIGHTS option, it is possible to specify a multiplication factor for each input sorted set. This means that the score of every element in every input sorted set is multiplied by this factor before being passed to the aggregation function. When WEIGHTS is not given, the multiplication factors default to 1.
	// With the AGGREGATE option, it is possible to specify how the results of the union are aggregated. This option defaults to SUM, where the score of an element is summed across the inputs where it exists. When this option is set to either MIN or MAX, the resulting set will contain the minimum or maximum score of an element across the inputs where it exists.
	// If destination already exists, it is overwritten.
	// Return:
	//	Integer reply: the number of elements in the resulting sorted set at destination.
	ZUnionStore(ctx context.Context, dest string, store ZStore) IntCmd
}

type SortedSetReader interface {
	SortedSetCacheCmdable

	// ZInter
	// Available since: 6.2.0
	// Time complexity: O(NK)+O(Mlog(M)) worst case with N being the smallest input sorted set, K being the number of input sorted sets and M being the number of elements in the resulting sorted set.
	// ACL categories: @read @sortedset @slow
	// This command is similar to ZINTERSTORE, but instead of storing the resulting sorted set, it is returned to the client.
	// For a description of the WEIGHTS and AGGREGATE options, see ZUNIONSTORE.
	// Return:
	//	Array reply: the result of intersection (optionally with their scores, in case the WITHSCORES option is given).
	ZInter(ctx context.Context, store ZStore) StringSliceCmd

	// ZInterWithScores
	// Available since: 6.2.0
	// Time complexity: O(NK)+O(Mlog(M)) worst case with N being the smallest input sorted set, K being the number of input sorted sets and M being the number of elements in the resulting sorted set.
	// ACL categories: @read @sortedset @slow
	ZInterWithScores(ctx context.Context, store ZStore) ZSliceCmd

	// ZRandMember
	// Available since: 6.2.0
	// Time complexity: O(N) where N is the number of elements returned
	// ACL categories: @read @sortedset @slow
	// When called with just the key argument, return a random element from the sorted set value stored at key.
	// If the provided count argument is positive, return an array of distinct elements. The array's length is either count or the sorted set's cardinality (ZCARD), whichever is lower.
	// If called with a negative count, the behavior changes and the command is allowed to return the same element multiple times. In this case, the number of returned elements is the absolute value of the specified count.
	// The optional WITHSCORES modifier changes the reply so it includes the respective scores of the randomly selected elements from the sorted set.
	// Return:
	// Bulk string reply: without the additional count argument, the command returns a Bulk Reply with the randomly selected element, or nil when key does not exist.
	//	Array reply: when the additional count argument is passed, the command returns an array of elements, or an empty array when key does not exist. If the WITHSCORES modifier is used, the reply is a list elements and their scores from the sorted set.
	ZRandMember(ctx context.Context, key string, count int, withScores bool) StringSliceCmd

	// ZScan
	// Available since: 2.8.0
	// Time complexity: O(1) for every call. O(N) for a complete iteration, including enough command calls for the cursor to return back to 0. N is the number of elements inside the collection..
	// ACL categories: @read @sortedset @slow
	// See https://redis.io/commands/scan/
	ZScan(ctx context.Context, key string, cursor uint64, match string, count int64) ScanCmd

	// ZDiff
	// Available since: 6.2.0
	// Time complexity: O(L + (N-K)log(N)) worst case where L is the total number of elements in all the sets, N is the size of the first set, and K is the size of the result set.
	// ACL categories: @read @sortedset @slow
	// This command is similar to ZDIFFSTORE, but instead of storing the resulting sorted set, it is returned to the client.
	// Return:
	// 	Array reply: the result of the difference (optionally with their scores, in case the WITHSCORES option is given).
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
	// This command is similar to ZUNIONSTORE, but instead of storing the resulting sorted set, it is returned to the client.
	// For a description of the WEIGHTS and AGGREGATE options, see ZUNIONSTORE.
	// Return:
	//	Array reply: the result of union (optionally with their scores, in case the WITHSCORES option is given).
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
	// Returns the sorted set cardinality (number of elements) of the sorted set stored at key.
	// Return:
	// 	Integer reply: the cardinality (number of elements) of the sorted set, or 0 if key does not exist.
	ZCard(ctx context.Context, key string) IntCmd

	// ZCount
	// Available since: 2.0.0
	// Time complexity: O(log(N)) with N being the number of elements in the sorted set.
	// ACL categories: @read @sortedset @fast
	// Returns the number of elements in the sorted set at key with a score between min and max.
	// The min and max arguments have the same semantic as described for ZRANGEBYSCORE.
	// Note: the command has a complexity of just O(log(N)) because it uses elements ranks (see ZRANK) to get an idea of the range. Because of this there is no need to do a work proportional to the size of the range.
	// Return:
	// 	Integer reply: the number of elements in the specified score range.
	ZCount(ctx context.Context, key, min, max string) IntCmd

	// ZLexCount
	// Available since: 2.8.9
	// Time complexity: O(log(N)) with N being the number of elements in the sorted set.
	// ACL categories: @read @sortedset @fast
	// When all the elements in a sorted set are inserted with the same score, in order to force lexicographical ordering, this command returns the number of elements in the sorted set at key with a value between min and max.
	// The min and max arguments have the same meaning as described for ZRANGEBYLEX.
	// Note: the command has a complexity of just O(log(N)) because it uses elements ranks (see ZRANK) to get an idea of the range. Because of this there is no need to do a work proportional to the size of the range.
	// Return:
	// 	Integer reply: the number of elements in the specified score range.
	ZLexCount(ctx context.Context, key, min, max string) IntCmd

	// ZMScore
	// Available since: 6.2.0
	// Time complexity: O(N) where N is the number of members being requested.
	// ACL categories: @read @sortedset @fast
	// Returns the scores associated with the specified members in the sorted set stored at key.
	// For every member that does not exist in the sorted set, a nil value is returned.
	// Return:
	// 	Array reply: list of scores or nil associated with the specified member values (a double precision floating point number), represented as strings.
	ZMScore(ctx context.Context, key string, members ...string) FloatSliceCmd

	// ZRange
	// Available since: 1.2.0
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements returned.
	// ACL categories: @read @sortedset @slow
	// Returns the specified range of elements in the sorted set stored at <key>.
	// ZRANGE can perform different types of range queries: by index (rank), by the score, or by lexicographical order.
	// Array reply: list of elements in the specified range (optionally with their scores, in case the WITHSCORES option is given).
	// See https://redis.io/commands/zrange/
	ZRange(ctx context.Context, key string, start, stop int64) StringSliceCmd

	// ZRangeWithScores
	// Available since: 1.2.0
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements returned.
	// ACL categories: @read @sortedset @slow
	// Returns the specified range of elements in the sorted set stored at <key>.
	// ZRANGE can perform different types of range queries: by index (rank), by the score, or by lexicographical order.
	// Array reply: list of elements in the specified range (optionally with their scores, in case the WITHSCORES option is given).
	// See https://redis.io/commands/zrange/
	ZRangeWithScores(ctx context.Context, key string, start, stop int64) ZSliceCmd

	// ZRangeArgs
	// Available since: 1.2.0
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements returned.
	// ACL categories: @read @sortedset @slow
	// Returns the specified range of elements in the sorted set stored at <key>.
	// ZRANGE can perform different types of range queries: by index (rank), by the score, or by lexicographical order.
	// Array reply: list of elements in the specified range (optionally with their scores, in case the WITHSCORES option is given).
	// Starting with Redis version 6.2.0: Added the REV, BYSCORE, BYLEX and LIMIT options.
	// See https://redis.io/commands/zrange/
	ZRangeArgs(ctx context.Context, z ZRangeArgs) StringSliceCmd

	// ZRangeArgsWithScores
	// Available since: 1.2.0
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements returned.
	// ACL categories: @read @sortedset @slow
	// Returns the specified range of elements in the sorted set stored at <key>.
	// ZRANGE can perform different types of range queries: by index (rank), by the score, or by lexicographical order.
	// Array reply: list of elements in the specified range (optionally with their scores, in case the WITHSCORES option is given).
	// Starting with Redis version 6.2.0: Added the REV, BYSCORE, BYLEX and LIMIT options.
	// See https://redis.io/commands/zrange/
	ZRangeArgsWithScores(ctx context.Context, z ZRangeArgs) ZSliceCmd

	// ZRangeByLex
	// Available since: 2.8.9
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements being returned. If M is constant (e.g. always asking for the first 10 elements with LIMIT), you can consider it O(log(N)).
	// ACL categories: @read @sortedset @slow
	// Return:
	// Array reply: list of elements in the specified score range.
	// See https://redis.io/commands/zrangebylex/
	ZRangeByLex(ctx context.Context, key string, opt ZRangeBy) StringSliceCmd

	// ZRangeByScore
	// Available since: 1.0.5
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements being returned. If M is constant (e.g. always asking for the first 10 elements with LIMIT), you can consider it O(log(N)).
	// ACL categories: @read @sortedset @slow
	// Return:
	// Array reply: list of elements in the specified score range (optionally with their scores).
	// See https://redis.io/commands/zrangebyscore/
	ZRangeByScore(ctx context.Context, key string, opt ZRangeBy) StringSliceCmd

	// ZRangeByScoreWithScores
	// Available since: 1.0.5
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements being returned. If M is constant (e.g. always asking for the first 10 elements with LIMIT), you can consider it O(log(N)).
	// ACL categories: @read @sortedset @slow
	// Return:
	// Array reply: list of elements in the specified score range (optionally with their scores).
	// See https://redis.io/commands/zrangebyscore/
	ZRangeByScoreWithScores(ctx context.Context, key string, opt ZRangeBy) ZSliceCmd

	// ZRank
	// Available since: 2.0.0
	// Time complexity: O(log(N))
	// ACL categories: @read @sortedset @fast
	// Returns the rank of member in the sorted set stored at key, with the scores ordered from low to high. The rank (or index) is 0-based, which means that the member with the lowest score has rank 0.
	// Use ZREVRANK to get the rank of an element with the scores ordered from high to low.
	// Return:
	// 	If member exists in the sorted set, Integer reply: the rank of member.
	// 	If member does not exist in the sorted set or key does not exist, Bulk string reply: nil.
	ZRank(ctx context.Context, key, member string) IntCmd

	// ZRevRange
	// Available since: 1.2.0
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements returned.
	// ACL categories: @read @sortedset @slow
	// Returns the specified range of elements in the sorted set stored at key. The elements are considered to be ordered from the highest to the lowest score. Descending lexicographical order is used for elements with equal score.
	// Apart from the reversed ordering, ZREVRANGE is similar to ZRANGE.
	// Return:
	// 	Array reply: list of elements in the specified range (optionally with their scores).
	ZRevRange(ctx context.Context, key string, start, stop int64) StringSliceCmd

	// ZRevRangeWithScores
	// Available since: 1.2.0
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements returned.
	// ACL categories: @read @sortedset @slow
	// Returns the specified range of elements in the sorted set stored at key. The elements are considered to be ordered from the highest to the lowest score. Descending lexicographical order is used for elements with equal score.
	// Apart from the reversed ordering, ZREVRANGE is similar to ZRANGE.
	// Return:
	// 	Array reply: list of elements in the specified range (optionally with their scores).
	ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) ZSliceCmd

	// ZRevRangeByLex
	// Available since: 2.8.9
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements being returned. If M is constant (e.g. always asking for the first 10 elements with LIMIT), you can consider it O(log(N)).
	// ACL categories: @read @sortedset @slow
	// When all the elements in a sorted set are inserted with the same score, in order to force lexicographical ordering, this command returns all the elements in the sorted set at key with a value between max and min.
	// Apart from the reversed ordering, ZREVRANGEBYLEX is similar to ZRANGEBYLEX.
	// Return:
	// 	Array reply: list of elements in the specified score range.
	ZRevRangeByLex(ctx context.Context, key string, opt ZRangeBy) StringSliceCmd

	// ZRevRangeByScore
	// Available since: 2.2.0
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements being returned. If M is constant (e.g. always asking for the first 10 elements with LIMIT), you can consider it O(log(N)).
	// ACL categories: @read @sortedset @slow
	// Returns all the elements in the sorted set at key with a score between max and min (including elements with score equal to max or min). In contrary to the default ordering of sorted sets, for this command the elements are considered to be ordered from high to low scores.
	// The elements having the same score are returned in reverse lexicographical order.
	// Apart from the reversed ordering, ZREVRANGEBYSCORE is similar to ZRANGEBYSCORE.
	// Return:
	// 	Array reply: list of elements in the specified score range (optionally with their scores).
	// Starting with Redis version 2.1.6: min and max can be exclusive.
	ZRevRangeByScore(ctx context.Context, key string, opt ZRangeBy) StringSliceCmd

	// ZRevRangeByScoreWithScores
	// Available since: 2.2.0
	// Time complexity: O(log(N)+M) with N being the number of elements in the sorted set and M the number of elements being returned. If M is constant (e.g. always asking for the first 10 elements with LIMIT), you can consider it O(log(N)).
	// ACL categories: @read @sortedset @slow
	// Returns all the elements in the sorted set at key with a score between max and min (including elements with score equal to max or min). In contrary to the default ordering of sorted sets, for this command the elements are considered to be ordered from high to low scores.
	// The elements having the same score are returned in reverse lexicographical order.
	// Apart from the reversed ordering, ZREVRANGEBYSCORE is similar to ZRANGEBYSCORE.
	// Return:
	// 	Array reply: list of elements in the specified score range (optionally with their scores).
	// Starting with Redis version 2.1.6: min and max can be exclusive.
	ZRevRangeByScoreWithScores(ctx context.Context, key string, opt ZRangeBy) ZSliceCmd

	// ZRevRank
	// Available since: 2.0.0
	// Time complexity: O(log(N))
	// ACL categories: @read @sortedset @fast
	// Returns the rank of member in the sorted set stored at key, with the scores ordered from high to low. The rank (or index) is 0-based, which means that the member with the highest score has rank 0.
	// Use ZRANK to get the rank of an element with the scores ordered from low to high.
	// Return:
	// 	If member exists in the sorted set, Integer reply: the rank of member.
	// 	If member does not exist in the sorted set or key does not exist, Bulk string reply: nil.
	ZRevRank(ctx context.Context, key, member string) IntCmd

	// ZScore
	// Available since: 1.2.0
	// Time complexity: O(1)
	// ACL categories: @read @sortedset @fast
	// Returns the score of member in the sorted set at key.
	// If member does not exist in the sorted set, or key does not exist, nil is returned.
	// Return:
	// 	Bulk string reply: the score of member (a double precision floating point number), represented as string.
	ZScore(ctx context.Context, key, member string) FloatCmd
}

func (c *client) BZPopMax(ctx context.Context, timeout time.Duration, keys ...string) ZWithKeyCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandBZPopMax, func() []string { return keys })
	r := c.cmdable.BZPopMax(ctx, timeout, keys...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) BZPopMin(ctx context.Context, timeout time.Duration, keys ...string) ZWithKeyCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandBZPopMin, func() []string { return keys })
	r := c.cmdable.BZPopMin(ctx, timeout, keys...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZAdd(ctx context.Context, key string, members ...Z) IntCmd {
	if len(members) > 1 {
		ctx = c.handler.before(ctx, CommandZAddMultiple)
	} else {
		ctx = c.handler.before(ctx, CommandZAdd)
	}
	r := c.cmdable.ZAdd(ctx, key, members...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZAddNX(ctx context.Context, key string, members ...Z) IntCmd {
	ctx = c.handler.before(ctx, CommandZAddNX)
	r := c.cmdable.ZAddNX(ctx, key, members...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZAddXX(ctx context.Context, key string, members ...Z) IntCmd {
	ctx = c.handler.before(ctx, CommandZAddXX)
	r := c.cmdable.ZAddXX(ctx, key, members...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZAddCh(ctx context.Context, key string, members ...Z) IntCmd {
	ctx = c.handler.before(ctx, CommandZAddCh)
	r := c.cmdable.ZAddCh(ctx, key, members...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZAddNXCh(ctx context.Context, key string, members ...Z) IntCmd {
	ctx = c.handler.before(ctx, CommandZAddNX)
	r := c.cmdable.ZAddNXCh(ctx, key, members...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZAddXXCh(ctx context.Context, key string, members ...Z) IntCmd {
	ctx = c.handler.before(ctx, CommandZAddXX)
	r := c.cmdable.ZAddXXCh(ctx, key, members...)
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
	} else {
		ctx = c.handler.before(ctx, CommandZAdd)
	}
	r := c.cmdable.ZAddArgs(ctx, key, args)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZAddArgsIncr(ctx context.Context, key string, args ZAddArgs) FloatCmd {
	if args.GT {
		ctx = c.handler.before(ctx, CommandZAddGT)
	} else if args.LT {
		ctx = c.handler.before(ctx, CommandZAddLT)
	} else {
		ctx = c.handler.before(ctx, CommandZAddIncr)
	}
	r := c.cmdable.ZAddArgsIncr(ctx, key, args)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZIncr(ctx context.Context, key string, member Z) FloatCmd {
	ctx = c.handler.before(ctx, CommandZAddIncr)
	r := c.cmdable.ZIncr(ctx, key, member)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZIncrNX(ctx context.Context, key string, member Z) FloatCmd {
	ctx = c.handler.before(ctx, CommandZAddIncr)
	r := c.cmdable.ZIncrNX(ctx, key, member)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZIncrXX(ctx context.Context, key string, member Z) FloatCmd {
	ctx = c.handler.before(ctx, CommandZAddIncr)
	r := c.cmdable.ZIncrXX(ctx, key, member)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZCard(ctx context.Context, key string) IntCmd {
	ctx = c.handler.before(ctx, CommandZCard)
	r := c.cmdable.ZCard(ctx, key)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZCount(ctx context.Context, key, min, max string) IntCmd {
	ctx = c.handler.before(ctx, CommandZCount)
	r := c.cmdable.ZCount(ctx, key, min, max)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZDiff(ctx context.Context, keys ...string) StringSliceCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandZDiff, func() []string { return keys })
	r := c.cmdable.ZDiff(ctx, keys...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZDiffWithScores(ctx context.Context, keys ...string) ZSliceCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandZDiff, func() []string { return keys })
	r := c.cmdable.ZDiffWithScores(ctx, keys...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZDiffStore(ctx context.Context, destination string, keys ...string) IntCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandZDiffStore, func() []string { return appendString(destination, keys...) })
	r := c.cmdable.ZDiffStore(ctx, destination, keys...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZIncrBy(ctx context.Context, key string, increment float64, member string) FloatCmd {
	ctx = c.handler.before(ctx, CommandZIncrBy)
	r := c.cmdable.ZIncrBy(ctx, key, increment, member)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZInter(ctx context.Context, store ZStore) StringSliceCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandZInter, func() []string { return store.Keys })
	r := c.cmdable.ZInter(ctx, store)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZInterWithScores(ctx context.Context, store ZStore) ZSliceCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandZInter, func() []string { return store.Keys })
	r := c.cmdable.ZInterWithScores(ctx, store)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZInterStore(ctx context.Context, destination string, store ZStore) IntCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandZInterStore, func() []string { return appendString(destination, store.Keys...) })
	r := c.cmdable.ZInterStore(ctx, destination, store)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZLexCount(ctx context.Context, key, min, max string) IntCmd {
	ctx = c.handler.before(ctx, CommandZLexCount)
	r := c.cmdable.ZLexCount(ctx, key, min, max)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZMScore(ctx context.Context, key string, members ...string) FloatSliceCmd {
	ctx = c.handler.before(ctx, CommandZMScore)
	r := c.cmdable.ZMScore(ctx, key, members...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZPopMax(ctx context.Context, key string, count ...int64) ZSliceCmd {
	ctx = c.handler.before(ctx, CommandZPopMax)
	r := c.cmdable.ZPopMax(ctx, key, count...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZPopMin(ctx context.Context, key string, count ...int64) ZSliceCmd {
	ctx = c.handler.before(ctx, CommandZPopMin)
	r := c.cmdable.ZPopMin(ctx, key, count...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRandMember(ctx context.Context, key string, count int, withScores bool) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandZRandMember)
	r := c.cmdable.ZRandMember(ctx, key, count, withScores)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRange(ctx context.Context, key string, start, stop int64) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandZRange)
	r := c.cmdable.ZRange(ctx, key, start, stop)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ZSliceCmd {
	ctx = c.handler.before(ctx, CommandZRange)
	r := c.cmdable.ZRangeWithScores(ctx, key, start, stop)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRangeArgs(ctx context.Context, z ZRangeArgs) StringSliceCmd {
	if z.Rev {
		ctx = c.handler.before(ctx, CommandZRangeRev)
	} else if z.ByScore {
		ctx = c.handler.before(ctx, CommandZRangeByScore)
	} else if z.ByLex {
		ctx = c.handler.before(ctx, CommandZRangeByLex)
	} else if z.Offset != 0 || z.Count != 0 {
		ctx = c.handler.before(ctx, CommandZRangeLimit)
	} else {
		ctx = c.handler.before(ctx, CommandZRange)
	}
	r := c.cmdable.ZRangeArgs(ctx, z)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRangeArgsWithScores(ctx context.Context, z ZRangeArgs) ZSliceCmd {
	if z.Rev {
		ctx = c.handler.before(ctx, CommandZRangeRev)
	} else if z.ByScore {
		ctx = c.handler.before(ctx, CommandZRangeByScore)
	} else if z.ByLex {
		ctx = c.handler.before(ctx, CommandZRangeByLex)
	} else if z.Offset != 0 || z.Count != 0 {
		ctx = c.handler.before(ctx, CommandZRangeLimit)
	} else {
		ctx = c.handler.before(ctx, CommandZRange)
	}
	r := c.cmdable.ZRangeArgsWithScores(ctx, z)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRangeByLex(ctx context.Context, key string, opt ZRangeBy) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandZRangebylex)
	r := c.cmdable.ZRangeByLex(ctx, key, opt)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRangeByScore(ctx context.Context, key string, opt ZRangeBy) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandZrangebyscore)
	r := c.cmdable.ZRangeByScore(ctx, key, opt)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRangeByScoreWithScores(ctx context.Context, key string, opt ZRangeBy) ZSliceCmd {
	ctx = c.handler.before(ctx, CommandZrangebyscoreWithScores)
	r := c.cmdable.ZRangeByScoreWithScores(ctx, key, opt)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRangeStore(ctx context.Context, dst string, z ZRangeArgs) IntCmd {
	ctx = c.handler.before(ctx, CommandZRangeStore)
	r := c.cmdable.ZRangeStore(ctx, dst, z)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRank(ctx context.Context, key, member string) IntCmd {
	ctx = c.handler.before(ctx, CommandZRank)
	r := c.cmdable.ZRank(ctx, key, member)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRem(ctx context.Context, key string, members ...interface{}) IntCmd {
	if len(members) > 1 {
		ctx = c.handler.before(ctx, CommandZRemMultiple)
	} else {
		ctx = c.handler.before(ctx, CommandZRem)
	}
	r := c.cmdable.ZRem(ctx, key, members...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRemRangeByLex(ctx context.Context, key, min, max string) IntCmd {
	ctx = c.handler.before(ctx, CommandZRemRangeByLex)
	r := c.cmdable.ZRemRangeByLex(ctx, key, min, max)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRemRangeByRank(ctx context.Context, key string, start, stop int64) IntCmd {
	ctx = c.handler.before(ctx, CommandZRemRangeByRank)
	r := c.cmdable.ZRemRangeByRank(ctx, key, start, stop)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRemRangeByScore(ctx context.Context, key, min, max string) IntCmd {
	ctx = c.handler.before(ctx, CommandZRemRangeByScore)
	r := c.cmdable.ZRemRangeByScore(ctx, key, min, max)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRevRange(ctx context.Context, key string, start, stop int64) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandZRevRange)
	r := c.cmdable.ZRevRange(ctx, key, start, stop)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) ZSliceCmd {
	ctx = c.handler.before(ctx, CommandZRevRange)
	r := c.cmdable.ZRevRangeWithScores(ctx, key, start, stop)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRevRangeByLex(ctx context.Context, key string, opt ZRangeBy) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandZRevRangeByLex)
	r := c.cmdable.ZRevRangeByLex(ctx, key, opt)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRevRangeByScore(ctx context.Context, key string, opt ZRangeBy) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandZRevRangeByScore)
	r := c.cmdable.ZRevRangeByScore(ctx, key, opt)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRevRangeByScoreWithScores(ctx context.Context, key string, opt ZRangeBy) ZSliceCmd {
	ctx = c.handler.before(ctx, CommandZRevRangeByScore)
	r := c.cmdable.ZRevRangeByScoreWithScores(ctx, key, opt)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZRevRank(ctx context.Context, key, member string) IntCmd {
	ctx = c.handler.before(ctx, CommandZRevRank)
	r := c.cmdable.ZRevRank(ctx, key, member)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZScan(ctx context.Context, key string, cursor uint64, match string, count int64) ScanCmd {
	ctx = c.handler.before(ctx, CommandZScan)
	r := c.cmdable.ZScan(ctx, key, cursor, match, count)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZScore(ctx context.Context, key, member string) FloatCmd {
	ctx = c.handler.before(ctx, CommandZScore)
	r := c.cmdable.ZScore(ctx, key, member)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZUnion(ctx context.Context, store ZStore) StringSliceCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandZUnion, func() []string { return store.Keys })
	r := c.cmdable.ZUnion(ctx, store)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZUnionWithScores(ctx context.Context, store ZStore) ZSliceCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandZUnion, func() []string { return store.Keys })
	r := c.cmdable.ZUnionWithScores(ctx, store)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ZUnionStore(ctx context.Context, dest string, store ZStore) IntCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandZUnionStore, func() []string { return appendString(dest, store.Keys...) })
	r := c.cmdable.ZUnionStore(ctx, dest, store)
	c.handler.after(ctx, r.Err())
	return r
}
