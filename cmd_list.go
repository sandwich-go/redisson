package redisson

import (
	"context"
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
	// BLMOVE is the blocking variant of LMOVE. When source contains elements, this command behaves exactly like LMOVE. When used inside a MULTI/EXEC block, this command behaves exactly like LMOVE. When source is empty, Redis will block the connection until another client pushes to it or until timeout (a double value specifying the maximum number of seconds to block) is reached. A timeout of zero can be used to block indefinitely.
	// This command comes in place of the now deprecated BRPOPLPUSH. Doing BLMOVE RIGHT LEFT is equivalent.
	// See LMOVE for more information.
	// Return:
	//	Bulk string reply: the element being popped from source and pushed to destination. If timeout is reached, a Null reply is returned.
	BLMove(ctx context.Context, source, destination, srcpos, destpos string, timeout time.Duration) StringCmd

	// BLMPop
	// Available since: 7.0.0
	// Time complexity: O(N+M) where N is the number of provided keys and M is the number of elements returned.
	// ACL categories: @write, @list, @slow, @blocking
	// BLMPOP is the blocking variant of LMPOP.
	// When any of the lists contains elements, this command behaves exactly like LMPOP. When used inside a MULTI/EXEC block, this command behaves exactly like LMPOP. When all lists are empty, Redis will block the connection until another client pushes to it or until the timeout (a double value specifying the maximum number of seconds to block) elapses. A timeout of zero can be used to block indefinitely.
	BLMPop(ctx context.Context, timeout time.Duration, direction string, count int64, keys ...string) KeyValuesCmd

	// BLPop
	// Available since: 2.0.0
	// Time complexity: O(N) where N is the number of provided keys.
	// ACL categories: @write @list @slow @blocking
	// BLPOP is a blocking list pop primitive. It is the blocking version of LPOP because it blocks the connection when there are no elements to pop from any of the given lists. An element is popped from the head of the first list that is non-empty, with the given keys being checked in the order that they are given.
	// See https://redis.io/commands/blpop/
	BLPop(ctx context.Context, timeout time.Duration, keys ...string) StringSliceCmd

	// BRPop
	// Available since: 2.0.0
	// Time complexity: O(N) where N is the number of provided keys.
	// ACL categories: @write @list @slow @blocking
	// BRPOP is a blocking list pop primitive. It is the blocking version of RPOP because it blocks the connection when there are no elements to pop from any of the given lists. An element is popped from the tail of the first list that is non-empty, with the given keys being checked in the order that they are given.
	// See the BLPOP documentation for the exact semantics, since BRPOP is identical to BLPOP with the only difference being that it pops elements from the tail of a list instead of popping from the head.
	// Return:
	// Array reply: specifically:
	// 	A nil multi-bulk when no element could be popped and the timeout expired.
	// 	A two-element multi-bulk with the first element being the name of the key where an element was popped and the second element being the value of the popped element.
	BRPop(ctx context.Context, timeout time.Duration, keys ...string) StringSliceCmd

	// BRPopLPush
	// Available since: 2.2.0
	// Time complexity: O(1)
	// ACL categories: @write @list @slow @blocking
	// As of Redis version 6.2.0, this command is regarded as deprecated.
	// It can be replaced by BLMOVE with the RIGHT and LEFT arguments when migrating or writing new code.
	// BRPOPLPUSH is the blocking variant of RPOPLPUSH. When source contains elements, this command behaves exactly like RPOPLPUSH. When used inside a MULTI/EXEC block, this command behaves exactly like RPOPLPUSH. When source is empty, Redis will block the connection until another client pushes to it or until timeout is reached. A timeout of zero can be used to block indefinitely.
	// See RPOPLPUSH for more information.
	BRPopLPush(ctx context.Context, source, destination string, timeout time.Duration) StringCmd

	// LInsert
	// Available since: 2.2.0
	// Time complexity: O(N) where N is the number of elements to traverse before seeing the value pivot. This means that inserting somewhere on the left end on the list (head) can be considered O(1) and inserting somewhere on the right end (tail) is O(N).
	// ACL categories: @write @list @slow
	// Inserts element in the list stored at key either before or after the reference value pivot.
	// When key does not exist, it is considered an empty list and no operation is performed.
	// An error is returned when key exists but does not hold a list value.
	// Return:
	// 	Integer reply: the length of the list after the insert operation, or -1 when the value pivot was not found.
	LInsert(ctx context.Context, key, op string, pivot, value any) IntCmd

	// LInsertBefore
	// Available since: 2.2.0
	// Time complexity: O(N) where N is the number of elements to traverse before seeing the value pivot. This means that inserting somewhere on the left end on the list (head) can be considered O(1) and inserting somewhere on the right end (tail) is O(N).
	// ACL categories: @write @list @slow
	LInsertBefore(ctx context.Context, key string, pivot, value any) IntCmd

	// LInsertAfter
	// Available since: 2.2.0
	// Time complexity: O(N) where N is the number of elements to traverse before seeing the value pivot. This means that inserting somewhere on the left end on the list (head) can be considered O(1) and inserting somewhere on the right end (tail) is O(N).
	// ACL categories: @write @list @slow
	LInsertAfter(ctx context.Context, key string, pivot, value any) IntCmd

	// LMove
	// Available since: 6.2.0
	// Time complexity: O(1)
	// ACL categories: @write @list @slow
	// Atomically returns and removes the first/last element (head/tail depending on the wherefrom argument) of the list stored at source, and pushes the element at the first/last element (head/tail depending on the whereto argument) of the list stored at destination.
	// For example: consider source holding the list a,b,c, and destination holding the list x,y,z. Executing LMOVE source destination RIGHT LEFT results in source holding a,b and destination holding c,x,y,z.
	// If source does not exist, the value nil is returned and no operation is performed. If source and destination are the same, the operation is equivalent to removing the first/last element from the list and pushing it as first/last element of the list, so it can be considered as a list rotation command (or a no-op if wherefrom is the same as whereto).
	// This command comes in place of the now deprecated RPOPLPUSH. Doing LMOVE RIGHT LEFT is equivalent.
	// Return
	//	Bulk string reply: the element being popped and pushed.
	LMove(ctx context.Context, source, destination, srcpos, destpos string) StringCmd

	// LPop
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the number of elements returned
	// ACL categories: @write @list @fast
	// Removes and returns the first elements of the list stored at key.
	// By default, the command pops a single element from the beginning of the list. When provided with the optional count argument, the reply will consist of up to count elements, depending on the list's length.
	// Return:
	// When called without the count argument:
	// 	Bulk string reply: the value of the first element, or nil when key does not exist.
	// When called with the count argument:
	// 	Array reply: list of popped elements, or nil when key does not exist.
	LPop(ctx context.Context, key string) StringCmd

	// LMPop
	// Available since: 7.0.0
	// Time complexity: O(N+M) where N is the number of provided keys and M is the number of elements returned.
	// ACL categories: @write, @list, @slow
	LMPop(ctx context.Context, direction string, count int64, keys ...string) KeyValuesCmd

	// LPopCount
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the number of elements returned
	// ACL categories: @write @list @fast
	// Removes and returns the first elements of the list stored at key.
	// By default, the command pops a single element from the beginning of the list. When provided with the optional count argument, the reply will consist of up to count elements, depending on the list's length.
	// Return:
	// When called without the count argument:
	// 	Bulk string reply: the value of the first element, or nil when key does not exist.
	// When called with the count argument:
	// 	Array reply: list of popped elements, or nil when key does not exist.
	LPopCount(ctx context.Context, key string, count int64) StringSliceCmd

	// LPush
	// Available since: 1.0.0
	// Time complexity: O(1) for each element added, so O(N) to add N elements when the command is called with multiple arguments.
	// ACL categories: @write @list @fast
	// Insert all the specified values at the head of the list stored at key. If key does not exist, it is created as empty list before performing the push operations. When key holds a value that is not a list, an error is returned.
	// It is possible to push multiple elements using a single command call just specifying multiple arguments at the end of the command. Elements are inserted one after the other to the head of the list, from the leftmost element to the rightmost element. So for instance the command LPUSH mylist a b c will result into a list containing c as first element, b as second element and a as third element.
	// Return:
	// 	Integer reply: the length of the list after the push operations.
	LPush(ctx context.Context, key string, values ...any) IntCmd

	// LPushX
	// Available since: 2.2.0
	// Time complexity: O(1) for each element added, so O(N) to add N elements when the command is called with multiple arguments.
	// ACL categories: @write @list @fast
	// Inserts specified values at the head of the list stored at key, only if key already exists and holds a list. In contrary to LPUSH, no operation will be performed when key does not yet exist.
	// Return:
	// 	Integer reply: the length of the list after the push operation.
	LPushX(ctx context.Context, key string, values ...any) IntCmd

	// LRem
	// Available since: 1.0.0
	// Time complexity: O(N+M) where N is the length of the list and M is the number of elements removed.
	// ACL categories: @write @list @slow
	// Removes the first count occurrences of elements equal to element from the list stored at key. The count argument influences the operation in the following ways:
	// count > 0: Remove elements equal to element moving from head to tail.
	// count < 0: Remove elements equal to element moving from tail to head.
	// count = 0: Remove all elements equal to element.
	// For example, LREM list -2 "hello" will remove the last two occurrences of "hello" in the list stored at list.
	// Note that non-existing keys are treated like empty lists, so when key does not exist, the command will always return 0.
	// Return:
	// 	Integer reply: the number of removed elements.
	LRem(ctx context.Context, key string, count int64, value any) IntCmd

	// LSet
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the length of the list. Setting either the first or the last element of the list is O(1).
	// ACL categories: @write @list @slow
	// Sets the list element at index to element. For more information on the index argument, see LINDEX.
	// An error is returned for out of range indexes.
	// Return:
	// 	Simple string reply
	LSet(ctx context.Context, key string, index int64, value any) StatusCmd

	// LTrim
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the number of elements to be removed by the operation.
	// ACL categories: @write @list @slow
	// Trim an existing list so that it will contain only the specified range of elements specified. Both start and stop are zero-based indexes, where 0 is the first element of the list (the head), 1 the next element and so on.
	// For example: LTRIM foobar 0 2 will modify the list stored at foobar so that only the first three elements of the list will remain.
	// start and end can also be negative numbers indicating offsets from the end of the list, where -1 is the last element of the list, -2 the penultimate element and so on.
	// Out of range indexes will not produce an error: if start is larger than the end of the list, or start > end, the result will be an empty list (which causes key to be removed). If end is larger than the end of the list, Redis will treat it like the last element of the list.
	// A common use of LTRIM is together with LPUSH / RPUSH. For example:
	// 	LPUSH mylist someelement
	// 	LTRIM mylist 0 99
	// This pair of commands will push a new element on the list, while making sure that the list will not grow larger than 100 elements. This is very useful when using Redis to store logs for example. It is important to note that when used in this way LTRIM is an O(1) operation because in the average case just one element is removed from the tail of the list.
	// Return:
	// 	Simple string reply
	LTrim(ctx context.Context, key string, start, stop int64) StatusCmd

	// RPop
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the number of elements returned
	// ACL categories: @write @list @fast
	// Removes and returns the last elements of the list stored at key.
	// By default, the command pops a single element from the end of the list. When provided with the optional count argument, the reply will consist of up to count elements, depending on the list's length.
	// Return:
	// When called without the count argument:
	// 	Bulk string reply: the value of the last element, or nil when key does not exist.
	// When called with the count argument:
	// 	Array reply: list of popped elements, or nil when key does not exist.
	RPop(ctx context.Context, key string) StringCmd

	// RPopCount
	// Available since: 6.2.0
	// Time complexity: O(N) where N is the number of elements returned
	// ACL categories: @write @list @fast
	// See RPop
	RPopCount(ctx context.Context, key string, count int64) StringSliceCmd

	// RPopLPush
	// Available since: 1.2.0
	// Time complexity: O(1)
	// ACL categories: @write @list @slow
	// As of Redis version 6.2.0, this command is regarded as deprecated.
	// It can be replaced by LMOVE with the RIGHT and LEFT arguments when migrating or writing new code.
	// Atomically returns and removes the last element (tail) of the list stored at source, and pushes the element at the first element (head) of the list stored at destination.
	// For example: consider source holding the list a,b,c, and destination holding the list x,y,z. Executing RPOPLPUSH results in source holding a,b and destination holding c,x,y,z.
	// If source does not exist, the value nil is returned and no operation is performed. If source and destination are the same, the operation is equivalent to removing the last element from the list and pushing it as first element of the list, so it can be considered as a list rotation command.
	// Return
	//	Bulk string reply: the element being popped and pushed.
	RPopLPush(ctx context.Context, source, destination string) StringCmd

	// RPush
	// Available since: 1.0.0
	// Time complexity: O(1) for each element added, so O(N) to add N elements when the command is called with multiple arguments.
	// ACL categories: @write @list @fast
	// Insert all the specified values at the tail of the list stored at key. If key does not exist, it is created as empty list before performing the push operation. When key holds a value that is not a list, an error is returned.
	// It is possible to push multiple elements using a single command call just specifying multiple arguments at the end of the command. Elements are inserted one after the other to the tail of the list, from the leftmost element to the rightmost element. So for instance the command RPUSH mylist a b c will result into a list containing a as first element, b as second element and c as third element.
	// Return:
	// 	Integer reply: the length of the list after the push operation.
	RPush(ctx context.Context, key string, values ...any) IntCmd

	// RPushX
	// Available since: 2.2.0
	// Time complexity: O(1) for each element added, so O(N) to add N elements when the command is called with multiple arguments.
	// ACL categories: @write @list @fast
	// Inserts specified values at the tail of the list stored at key, only if key already exists and holds a list. In contrary to RPUSH, no operation will be performed when key does not yet exist.
	// Return:
	// 	Integer reply: the length of the list after the push operation.
	RPushX(ctx context.Context, key string, values ...any) IntCmd
}

type ListReader interface {
	// LPosCount
	// vailable since: 6.0.6
	// Time complexity: O(N) where N is the number of elements in the list, for the average case. When searching for elements near the head or the tail of the list, or when the MAXLEN option is provided, the command may run in constant time.
	// ACL categories: @read @list @slow
	// See https://redis.io/commands/lpos/
	LPosCount(ctx context.Context, key string, value string, count int64, args LPosArgs) IntSliceCmd
}

type ListCacheCmdable interface {
	// LIndex
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the number of elements to traverse to get to the element at index. This makes asking for the first or the last element of the list O(1).
	// ACL categories: @read @list @slow
	// Returns the element at index index in the list stored at key. The index is zero-based, so 0 means the first element, 1 the second element and so on. Negative indices can be used to designate elements starting at the tail of the list. Here, -1 means the last element, -2 means the penultimate and so forth.
	// When the value at key is not a list, an error is returned.
	// Return:
	// 	Bulk string reply: the requested element, or nil when index is out of range.
	LIndex(ctx context.Context, key string, index int64) StringCmd

	// LLen
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @read @list @fast
	// Returns the length of the list stored at key. If key does not exist, it is interpreted as an empty list and 0 is returned. An error is returned when the value stored at key is not a list.
	// Return:
	// 	Integer reply: the length of the list at key.
	LLen(ctx context.Context, key string) IntCmd

	// LRange
	// Available since: 1.0.0
	// Time complexity: O(S+N) where S is the distance of start offset from HEAD for small lists, from nearest end (HEAD or TAIL) for large lists; and N is the number of elements in the specified range.
	// ACL categories: @read @list @slow
	// Returns the specified elements of the list stored at key. The offsets start and stop are zero-based indexes, with 0 being the first element of the list (the head of the list), 1 being the next element and so on.
	// These offsets can also be negative numbers indicating offsets starting at the end of the list. For example, -1 is the last element of the list, -2 the penultimate, and so on.
	// Consistency with range functions in various programming languages
	// Note that if you have a list of numbers from 0 to 100, LRANGE list 0 10 will return 11 elements, that is, the rightmost item is included. This may or may not be consistent with behavior of range-related functions in your programming language of choice (think Ruby's Range.new, Array#slice or Python's range() function).
	// Out-of-range indexes
	// Out of range indexes will not produce an error. If start is larger than the end of the list, an empty list is returned. If stop is larger than the actual end of the list, Redis will treat it like the last element of the list.
	// Return:
	// 	Array reply: list of elements in the specified range.
	LRange(ctx context.Context, key string, start, stop int64) StringSliceCmd

	// LPos
	// vailable since: 6.0.6
	// Time complexity: O(N) where N is the number of elements in the list, for the average case. When searching for elements near the head or the tail of the list, or when the MAXLEN option is provided, the command may run in constant time.
	// ACL categories: @read @list @slow
	// See https://redis.io/commands/lpos/
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
	ctx = c.handler.before(ctx, CommandLInsert)
	r := c.adapter.LInsert(ctx, key, op, pivot, value)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) LInsertBefore(ctx context.Context, key string, pivot, value any) IntCmd {
	ctx = c.handler.before(ctx, CommandLInsert)
	r := c.adapter.LInsertBefore(ctx, key, pivot, value)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) LInsertAfter(ctx context.Context, key string, pivot, value any) IntCmd {
	ctx = c.handler.before(ctx, CommandLInsert)
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
