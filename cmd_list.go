package redisson

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type ListCmdable interface {
	ListWriter
	ListReader
}

type ListWriter interface {
	// BLMove
	// Available since: 6.2.0
	// Time complexity: O(1)
	// ACL categories: @write @list @slow @blocking
	// RESP2 Reply:
	//	One of the following:
	//		- Bulk string reply: the element being popped from the source and pushed to the destination.
	//		- Nil reply: the operation timed-out
	// RESP3 Reply:
	//	One of the following:
	//		- Bulk string reply: the element being popped from the source and pushed to the destination.
	//		- Null reply: the operation timed-out
	BLMove(ctx context.Context, source, destination, srcpos, destpos string, timeout time.Duration) StringCmd

	// BLMPop
	// Available since: 7.0.0
	// Time complexity: O(N+M) where N is the number of provided keys and M is the number of elements returned.
	// ACL categories: @write, @list, @slow, @blocking
	// RESP2 Reply:
	//	One of the following:
	//		- Nil reply: when no element could be popped and the timeout is reached.
	//		- Array reply: a two-element array with the first element being the name of the key from which elements were popped, and the second element being an array of the popped elements.
	// RESP3 Reply:
	//	One of the following:
	//		- Null reply: when no element could be popped and the timeout is reached.
	//		- Array reply: a two-element array with the first element being the name of the key from which elements were popped, and the second element being an array of the popped elements.
	BLMPop(ctx context.Context, timeout time.Duration, direction string, count int64, keys ...string) KeyValuesCmd

	// BLPop
	// Available since: 2.0.0
	// Time complexity: O(N) where N is the number of provided keys.
	// ACL categories: @write @list @slow @blocking
	// RESP2 Reply:
	//	One of the following:
	//		- Nil reply: no element could be popped and the timeout expired
	//		- Array reply: the key from which the element was popped and the value of the popped element.
	// RESP3 Reply:
	//	One of the following:
	//		- Null reply: no element could be popped and the timeout expired
	//		- Array reply: the key from which the element was popped and the value of the popped element.
	// History:
	//	- Starting with Redis version 6.0.0: timeout is interpreted as a double instead of an integer.
	BLPop(ctx context.Context, timeout time.Duration, keys ...string) StringSliceCmd

	// BRPop
	// Available since: 2.0.0
	// Time complexity: O(N) where N is the number of provided keys.
	// ACL categories: @write @list @slow @blocking
	// RESP2 Reply:
	//	One of the following:
	//		- Nil reply: no element could be popped and the timeout expired.
	//		- Array reply: the key from which the element was popped and the value of the popped element
	// RESP3 Reply:
	//	One of the following:
	//		- Null reply: no element could be popped and the timeout expired.
	//		- Array reply: the key from which the element was popped and the value of the popped element
	// History:
	//	- Starting with Redis version 6.0.0: timeout is interpreted as a double instead of an integer.
	BRPop(ctx context.Context, timeout time.Duration, keys ...string) StringSliceCmd

	// BRPopLPush
	// Available since: 2.2.0
	// Time complexity: O(1)
	// ACL categories: @write @list @slow @blocking
	// RESP2 Reply:
	//	One of the following:
	//		- Bulk string reply: the element being popped from source and pushed to destination.
	//		- Nil reply: the timeout is reached.
	// RESP3 Reply:
	//	One of the following:
	//		- Bulk string reply: the element being popped from source and pushed to destination.
	//		- Null reply: the timeout is reached.
	// History:
	//	- Starting with Redis version 6.0.0: timeout is interpreted as a double instead of an integer.
	BRPopLPush(ctx context.Context, source, destination string, timeout time.Duration) StringCmd

	// LInsert
	// Available since: 2.2.0
	// Time complexity: O(N) where N is the number of elements to traverse before seeing the value pivot.
	//					This means that inserting somewhere on the left end on the list (head) can be considered O(1) and inserting somewhere on the right end (tail) is O(N).
	// ACL categories: @write @list @slow
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- Integer reply: the list length after a successful insert operation.
	//		- Integer reply: 0 when the key doesn't exist.
	//		- Integer reply: -1 when the pivot wasn't found.
	LInsert(ctx context.Context, key, op string, pivot, value any) IntCmd
	LInsertBefore(ctx context.Context, key string, pivot, value any) IntCmd
	LInsertAfter(ctx context.Context, key string, pivot, value any) IntCmd

	// LMove
	// Available since: 6.2.0
	// Time complexity: O(1)
	// ACL categories: @write @list @slow
	// RESP2 / RESP3 Reply:
	// 	- Bulk string reply: the element being popped and pushed.
	LMove(ctx context.Context, source, destination, srcpos, destpos string) StringCmd

	// LPop
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the number of elements returned
	// ACL categories: @write @list @fast
	// RESP2 Reply:
	//	One of the following:
	//		- Nil reply: if the key does not exist.
	//		- Bulk string reply: when called without the count argument, the value of the first element.
	//		- Array reply: when called with the count argument, a list of popped elements.
	// RESP3 Reply:
	//	One of the following:
	//		- Null reply: if the key does not exist.
	//		- Bulk string reply: when called without the count argument, the value of the first element.
	//		- Array reply: when called with the count argument, a list of popped elements.
	// History:
	//	- Starting with Redis version 6.2.0: Added the count argument.
	LPop(ctx context.Context, key string) StringCmd
	LPopCount(ctx context.Context, key string, count int64) StringSliceCmd

	// LMPop
	// Available since: 7.0.0
	// Time complexity: O(N+M) where N is the number of provided keys and M is the number of elements returned.
	// ACL categories: @write, @list, @slow
	// RESP2 Reply:
	//	One of the following:
	//		- Nil reply: if no element could be popped.
	//		- Array reply: a two-element array with the first element being the name of the key from which elements were popped and the second element being an array of elements.
	// RESP3 Reply:
	//	One of the following:
	//		- Null reply: if no element could be popped.
	//		- Array reply: a two-element array with the first element being the name of the key from which elements were popped and the second element being an array of elements.
	LMPop(ctx context.Context, direction string, count int64, keys ...string) KeyValuesCmd

	// LPush
	// Available since: 1.0.0
	// Time complexity: O(1) for each element added, so O(N) to add N elements when the command is called with multiple arguments.
	// ACL categories: @write @list @fast
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the length of the list after the push operation.
	// History:
	//	- Starting with Redis version 2.4.0: Accepts multiple element arguments.
	LPush(ctx context.Context, key string, values ...any) IntCmd

	// LPushX
	// Available since: 2.2.0
	// Time complexity: O(1) for each element added, so O(N) to add N elements when the command is called with multiple arguments.
	// ACL categories: @write @list @fast
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the length of the list after the push operation.
	// History:
	//	- Starting with Redis version 4.0.0: Accepts multiple element arguments.
	LPushX(ctx context.Context, key string, values ...any) IntCmd

	// LRem
	// Available since: 1.0.0
	// Time complexity: O(N+M) where N is the length of the list and M is the number of elements removed.
	// ACL categories: @write @list @slow
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the number of removed elements.
	LRem(ctx context.Context, key string, count int64, value any) IntCmd

	// LSet
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the length of the list. Setting either the first or the last element of the list is O(1).
	// ACL categories: @write @list @slow
	// RESP2 / RESP3 Reply:
	// 	- Simple string reply: OK.
	LSet(ctx context.Context, key string, index int64, value any) StatusCmd

	// LTrim
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the number of elements to be removed by the operation.
	// ACL categories: @write @list @slow
	// RESP2 / RESP3 Reply:
	// 	- Simple string reply: OK.
	LTrim(ctx context.Context, key string, start, stop int64) StatusCmd

	// RPop
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the number of elements returned
	// ACL categories: @write @list @fast
	// RESP2 Reply:
	//	One of the following:
	//		- Nil reply: if the key does not exist.
	//		- Bulk string reply: when called without the count argument, the value of the last element.
	//		- Array reply: when called with the count argument, a list of popped elements.
	// RESP3 Reply:
	//	One of the following:
	//		- Null reply: if the key does not exist.
	//		- Bulk string reply: when called without the count argument, the value of the last element.
	//		- Array reply: when called with the count argument, a list of popped elements.
	// History:
	//	- Starting with Redis version 6.2.0: Added the count argument.
	RPop(ctx context.Context, key string) StringCmd
	RPopCount(ctx context.Context, key string, count int64) StringSliceCmd

	// RPopLPush
	// Available since: 1.2.0
	// Time complexity: O(1)
	// ACL categories: @write @list @slow
	// RESP2 Reply:
	//	One of the following:
	//		- Bulk string reply: the element being popped and pushed.
	//		- Nil reply: if the source list is empty.
	// RESP3 Reply:
	//	One of the following:
	//		- Bulk string reply: the element being popped and pushed.
	//		- Null reply: if the source list is empty.
	RPopLPush(ctx context.Context, source, destination string) StringCmd

	// RPush
	// Available since: 1.0.0
	// Time complexity: O(1) for each element added, so O(N) to add N elements when the command is called with multiple arguments.
	// ACL categories: @write @list @fast
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the length of the list after the push operation.
	// History:
	//	- Starting with Redis version 2.4.0: Accepts multiple element arguments.
	RPush(ctx context.Context, key string, values ...any) IntCmd

	// RPushX
	// Available since: 2.2.0
	// Time complexity: O(1) for each element added, so O(N) to add N elements when the command is called with multiple arguments.
	// ACL categories: @write @list @fast
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the length of the list after the push operation.
	// History:
	//	- Starting with Redis version 4.0.0: Accepts multiple element arguments.
	RPushX(ctx context.Context, key string, values ...any) IntCmd
}

type ListReader interface {
	// LPosCount
	// vailable since: 6.0.6
	// Time complexity: O(N) where N is the number of elements in the list, for the average case. When searching for elements near the head or the tail of the list,
	//					or when the MAXLEN option is provided, the command may run in constant time.
	// ACL categories: @read @list @slow
	// RESP2 Reply:
	//	Any of the following:
	//		- nil reply: if there is no matching element.
	//		- Integer reply: an integer representing the matching element.
	//		- Array reply: If the COUNT option is given, an array of integers representing the matching elements (or an empty array if there are no matches).
	// RESP3 Reply:
	//	Any of the following:
	//		- Null reply: if there is no matching element.
	//		- Integer reply: an integer representing the matching element.
	//		- Array reply: If the COUNT option is given, an array of integers representing the matching elements (or an empty array if there are no matches).
	LPosCount(ctx context.Context, key string, value string, count int64, args LPosArgs) IntSliceCmd
}

type ListCacheCmdable interface {
	// LIndex
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the number of elements to traverse to get to the element at index. This makes asking for the first or the last element of the list O(1).
	// ACL categories: @read @list @slow
	// RESP2 Reply:
	//	One of the following:
	//		- Nil reply: when index is out of range.
	//		- Bulk string reply: the requested element.
	// RESP3 Reply:
	//	One of the following:
	//		- Null reply: when index is out of range.
	//		- Bulk string reply: the requested element.
	LIndex(ctx context.Context, key string, index int64) StringCmd

	// LLen
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @read @list @fast
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the length of the list.
	LLen(ctx context.Context, key string) IntCmd

	// LRange
	// Available since: 1.0.0
	// Time complexity: O(S+N) where S is the distance of start offset from HEAD for small lists, from nearest end (HEAD or TAIL) for large lists;
	//					and N is the number of elements in the specified range.
	// ACL categories: @read @list @slow
	// RESP2 / RESP3 Reply:
	// 	- Array reply: a list of elements in the specified range, or an empty array if the key doesn't exist.
	LRange(ctx context.Context, key string, start, stop int64) StringSliceCmd

	// LPos
	// vailable since: 6.0.6
	// Time complexity: O(N) where N is the number of elements in the list, for the average case. When searching for elements near the head or the tail of the list,
	//					or when the MAXLEN option is provided, the command may run in constant time.
	// ACL categories: @read @list @slow
	// RESP2 Reply:
	//	One of the following:
	//		- nil reply: if there is no matching element.
	//		- Integer reply: an integer representing the matching element.
	//		- Array reply: If the COUNT option is given, an array of integers representing the matching elements (or an empty array if there are no matches).
	// RESP3 Reply:
	//	One of the following:
	//		- Null reply: if there is no matching element.
	//		- Integer reply: an integer representing the matching element.
	//		- Array reply: If the COUNT option is given, an array of integers representing the matching elements (or an empty array if there are no matches).
	LPos(ctx context.Context, key string, value string, args LPosArgs) IntCmd
}

func (c *client) BLMove(ctx context.Context, source, destination, srcpos, destpos string, timeout time.Duration) StringCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandBLMove, func() []string { return appendString(source, destination) })
	r := c.adapter.BLMove(ctx, source, destination, srcpos, destpos, timeout)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) BLMPop(ctx context.Context, timeout time.Duration, direction string, count int64, keys ...string) KeyValuesCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandBLMPop, func() []string { return keys })
	r := c.adapter.BLMPop(ctx, timeout, direction, count, keys...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) BLPop(ctx context.Context, timeout time.Duration, keys ...string) StringSliceCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandBLPop, func() []string { return keys })
	r := c.adapter.BLPop(ctx, timeout, keys...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) BRPop(ctx context.Context, timeout time.Duration, keys ...string) StringSliceCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandBRPop, func() []string { return keys })
	r := c.adapter.BRPop(ctx, timeout, keys...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) BRPopLPush(ctx context.Context, source, destination string, timeout time.Duration) StringCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandBRPopLPush, func() []string { return appendString(source, destination) })
	r := c.adapter.BRPopLPush(ctx, source, destination, timeout)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) LIndex(ctx context.Context, key string, index int64) StringCmd {
	ctx = c.handler.before(ctx, CommandLIndex)
	var r StringCmd
	if c.ttl > 0 {
		r = newStringCmd(c.Do(ctx, c.builder.LIndexCompleted(key, index)))
	} else {
		r = c.adapter.LIndex(ctx, key, index)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) LInsert(ctx context.Context, key, op string, pivot, value any) IntCmd {
	switch strings.ToUpper(op) {
	case BEFORE:
		ctx = c.handler.before(ctx, CommandLInsertBefore)
	case AFTER:
		ctx = c.handler.before(ctx, CommandLInsertAfter)
	default:
		panic(fmt.Sprintf("Invalid op argument value: %s", op))
	}
	r := c.adapter.LInsert(ctx, key, op, pivot, value)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) LInsertBefore(ctx context.Context, key string, pivot, value any) IntCmd {
	ctx = c.handler.before(ctx, CommandLInsertBefore)
	r := c.adapter.LInsertBefore(ctx, key, pivot, value)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) LInsertAfter(ctx context.Context, key string, pivot, value any) IntCmd {
	ctx = c.handler.before(ctx, CommandLInsertAfter)
	r := c.adapter.LInsertAfter(ctx, key, pivot, value)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) LLen(ctx context.Context, key string) IntCmd {
	ctx = c.handler.before(ctx, CommandLLen)
	var r IntCmd
	if c.ttl > 0 {
		r = newIntCmd(c.Do(ctx, c.builder.LLenCompleted(key)))
	} else {
		r = c.adapter.LLen(ctx, key)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) LMove(ctx context.Context, source, destination, srcpos, destpos string) StringCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandLMove, func() []string { return appendString(source, destination) })
	r := c.adapter.LMove(ctx, source, destination, srcpos, destpos)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) LPop(ctx context.Context, key string) StringCmd {
	ctx = c.handler.before(ctx, CommandLPop)
	r := c.adapter.LPop(ctx, key)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) LMPop(ctx context.Context, direction string, count int64, keys ...string) KeyValuesCmd {
	ctx = c.handler.before(ctx, CommandLMPop)
	r := c.adapter.LMPop(ctx, direction, count, keys...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) LPopCount(ctx context.Context, key string, count int64) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandLPopCount)
	r := c.adapter.LPopCount(ctx, key, count)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) LPos(ctx context.Context, key string, value string, args LPosArgs) IntCmd {
	ctx = c.handler.before(ctx, CommandLPos)
	var r IntCmd
	if c.ttl > 0 {
		r = newIntCmd(c.Do(ctx, c.builder.LPosCompleted(key, value, args)))
	} else {
		r = c.adapter.LPos(ctx, key, value, args)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) LPosCount(ctx context.Context, key string, value string, count int64, args LPosArgs) IntSliceCmd {
	ctx = c.handler.before(ctx, CommandLPosCount)
	r := c.adapter.LPosCount(ctx, key, value, count, args)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) LPush(ctx context.Context, key string, values ...any) IntCmd {
	if len(values) > 1 {
		ctx = c.handler.before(ctx, CommandLMPush)
	} else {
		ctx = c.handler.before(ctx, CommandLPush)
	}
	r := c.adapter.LPush(ctx, key, values...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) LPushX(ctx context.Context, key string, values ...any) IntCmd {
	if len(values) > 1 {
		ctx = c.handler.before(ctx, CommandLMPushX)
	} else {
		ctx = c.handler.before(ctx, CommandLPushX)
	}
	r := c.adapter.LPushX(ctx, key, values...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) LRange(ctx context.Context, key string, start, stop int64) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandLRange)
	var r StringSliceCmd
	if c.ttl > 0 {
		r = newStringSliceCmd(c.Do(ctx, c.builder.LRangeCompleted(key, start, stop)))
	} else {
		r = c.adapter.LRange(ctx, key, start, stop)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) LRem(ctx context.Context, key string, count int64, value any) IntCmd {
	ctx = c.handler.before(ctx, CommandLRem)
	r := c.adapter.LRem(ctx, key, count, value)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) LSet(ctx context.Context, key string, index int64, value any) StatusCmd {
	ctx = c.handler.before(ctx, CommandLSet)
	r := c.adapter.LSet(ctx, key, index, value)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) LTrim(ctx context.Context, key string, start, stop int64) StatusCmd {
	ctx = c.handler.before(ctx, CommandLTrim)
	r := c.adapter.LTrim(ctx, key, start, stop)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) RPop(ctx context.Context, key string) StringCmd {
	ctx = c.handler.before(ctx, CommandRPop)
	r := c.adapter.RPop(ctx, key)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) RPopCount(ctx context.Context, key string, count int64) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandRPopCount)
	r := c.adapter.RPopCount(ctx, key, count)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) RPopLPush(ctx context.Context, source, destination string) StringCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandRPopLPush, func() []string { return appendString(source, destination) })
	r := c.adapter.RPopLPush(ctx, source, destination)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) RPush(ctx context.Context, key string, values ...any) IntCmd {
	if len(values) > 1 {
		ctx = c.handler.before(ctx, CommandRMPush)
	} else {
		ctx = c.handler.before(ctx, CommandRPush)
	}
	r := c.adapter.RPush(ctx, key, values...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) RPushX(ctx context.Context, key string, values ...any) IntCmd {
	if len(values) > 1 {
		ctx = c.handler.before(ctx, CommandRMPushX)
	} else {
		ctx = c.handler.before(ctx, CommandRPushX)
	}
	r := c.adapter.RPushX(ctx, key, values...)
	c.handler.after(ctx, r.Err())
	return r
}
