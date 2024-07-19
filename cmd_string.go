package redisson

import (
	"context"
	"strings"
	"sync"
	"time"
)

type StringCmdable interface {
	StringWriter
	StringReader
}

type StringWriter interface {
	// Append
	// Available since: 2.0.0
	// Time complexity: O(1). The amortized time complexity is O(1) assuming the appended value is small and the already present value is of any size,
	//					since the dynamic string library used by Redis will double the free space available on every reallocation.
	// ACL categories: @write @string @fast
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the length of the string after the append operation.
	Append(ctx context.Context, key, value string) IntCmd

	// Decr
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @write @string @fast
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the value of the key after decrementing it.
	Decr(ctx context.Context, key string) IntCmd

	// DecrBy
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @write @string @fast
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the value of the key after decrementing it.
	DecrBy(ctx context.Context, key string, decrement int64) IntCmd

	// GetDel
	// Available since: 6.2.0
	// Time complexity: O(1)
	// ACL categories: @write @string @fast
	// RESP2 Reply:
	//	One of the following:
	//		- Bulk string reply: the value of the key.
	//		- Nil reply: if the key does not exist or if the key's value type is not a string.
	// RESP3 Reply:
	//	One of the following:
	//		- Bulk string reply: the value of the key.
	//		- Null reply: if the key does not exist or if the key's value type is not a string.
	GetDel(ctx context.Context, key string) StringCmd

	// GetEx
	// Available since: 6.2.0
	// Time complexity: O(1)
	// ACL categories: @write @string @fast
	// Options
	// 	- EX seconds -- Set the specified expire time, in seconds.
	// 	- PX milliseconds -- Set the specified expire time, in milliseconds.
	// 	- EXAT timestamp-seconds -- Set the specified Unix time at which the key will expire, in seconds.
	// 	- PXAT timestamp-milliseconds -- Set the specified Unix time at which the key will expire, in milliseconds.
	// 	- PERSIST -- Remove the time to live associated with the key.
	// RESP2 Reply:
	//	One of the following:
	//		- Bulk string reply: the value of the key.
	//		- Nil reply: if key does not exist.
	// RESP3 Reply:
	//	One of the following:
	//		- Bulk string reply: the value of the key.
	//		- Null reply: if key does not exist.
	GetEx(ctx context.Context, key string, expiration time.Duration) StringCmd

	// GetSet
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @write @string @fast
	// RESP2 Reply:
	//	One of the following:
	//		- Bulk string reply: the old value stored at the key.
	//		- Nil reply: if key does not exist.
	// RESP3 Reply:
	//	One of the following:
	//		- Bulk string reply: the old value stored at the key.
	//		- Null reply: if key does not exist.
	GetSet(ctx context.Context, key string, value any) StringCmd

	// Incr
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @write @string @fast
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the value of the key after the increment.
	Incr(ctx context.Context, key string) IntCmd

	// IncrBy
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @write @string @fast
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the value of the key after the increment.
	IncrBy(ctx context.Context, key string, value int64) IntCmd

	// IncrByFloat
	// Available since: 2.6.0
	// Time complexity: O(1)
	// ACL categories: @write @string @fast
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the value of the key after the increment.
	IncrByFloat(ctx context.Context, key string, value float64) FloatCmd

	// MSet
	// Available since: 1.0.1
	// Time complexity: O(N) where N is the number of keys to set.
	// ACL categories: @write @string @slow
	// RESP2 / RESP3 Reply:
	// 	- Simple string reply: always OK because MSET can't fail.
	MSet(ctx context.Context, values ...any) StatusCmd

	// MSetNX
	// Available since: 1.0.1
	// Time complexity: O(N) where N is the number of keys to set.
	// ACL categories: @write @string @slow
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- Integer reply: 0 if no key was set (at least one key already existed).
	//		- Integer reply: 1 if all the keys were set.
	MSetNX(ctx context.Context, values ...any) BoolCmd

	// Set
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @write @string @slow
	// Options
	// 	- EX seconds -- Set the specified expire time, in seconds (a positive integer).
	// 	- PX milliseconds -- Set the specified expire time, in milliseconds (a positive integer).
	// 	- EXAT timestamp-seconds -- Set the specified Unix time at which the key will expire, in seconds (a positive integer).
	// 	- PXAT timestamp-milliseconds -- Set the specified Unix time at which the key will expire, in milliseconds (a positive integer).
	// 	- NX -- Only set the key if it does not already exist.
	// 	- XX -- Only set the key if it already exists.
	// 	- KEEPTTL -- Retain the time to live associated with the key.
	// 	- GET -- Return the old string stored at key, or nil if key did not exist. An error is returned and SET aborted if the value stored at key is not a string.
	// Note: Since the SET command options can replace SETNX, SETEX, PSETEX, GETSET, it is possible that in future versions of Redis these commands will be deprecated and finally removed.
	// RESP2 Reply:
	//	Any of the following:
	//		- Nil reply: GET not given: Operation was aborted (conflict with one of the XX/NX options).
	//		- Simple string reply: OK. GET not given: The key was set.
	//		- Nil reply: GET given: The key didn't exist before the SET.
	//		- Bulk string reply: GET given: The previous value of the key.
	// RESP3 Reply:
	//	Any of the following:
	//		- Null reply: GET not given: Operation was aborted (conflict with one of the XX/NX options).
	//		- Simple string reply: OK. GET not given: The key was set.
	//		- Null reply: GET given: The key didn't exist before the SET.
	//		- Bulk string reply: GET given: The previous value of the key.
	// History:
	//	- Starting with Redis version 2.6.12: Added the EX, PX, NX and XX options.
	//	- Starting with Redis version 6.0.0: Added the KEEPTTL option.
	//	- Starting with Redis version 6.2.0: Added the GET, EXAT and PXAT option.
	//	- Starting with Redis version 7.0.0: Allowed the NX and GET options to be used together.
	Set(ctx context.Context, key string, value any, expiration time.Duration) StatusCmd
	SetXX(ctx context.Context, key string, value any, expiration time.Duration) BoolCmd
	SetArgs(ctx context.Context, key string, value any, a SetArgs) StatusCmd

	// SetEX
	// Available since: 2.0.0
	// Time complexity: O(1)
	// ACL categories: @write @string @slow
	// RESP2 / RESP3 Reply:
	// 	- Simple string reply: OK.
	SetEX(ctx context.Context, key string, value any, expiration time.Duration) StatusCmd

	// SetNX
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @write @string @fast
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- Integer reply: 0 if the key was not set.
	//		- Integer reply: 1 if the key was set.
	SetNX(ctx context.Context, key string, value any, expiration time.Duration) BoolCmd

	// SetRange
	// Available since: 2.2.0
	// Time complexity: O(1), not counting the time taken to copy the new string in place. Usually, this string is very small so the amortized complexity is O(1).
	//					Otherwise, complexity is O(M) with M being the length of the value argument.
	// ACL categories: @write @string @slow
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the length of the string after it was modified by the command.
	SetRange(ctx context.Context, key string, offset int64, value string) IntCmd
}

type StringReader interface {
	// MGet
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the number of keys to retrieve.
	// ACL categories: @read @string @fast
	// RESP2 / RESP3 Reply:
	// 	- Array reply: a list of values at the specified keys.
	MGet(ctx context.Context, keys ...string) SliceCmd
}

type StringCacheCmdable interface {
	// Get
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @read @string @fast
	// RESP2 Reply:
	//	One of the following:
	//		- Bulk string reply: the value of the key.
	//		- Nil reply: if the key does not exist.
	// RESP3 Reply:
	//	One of the following:
	//		- Bulk string reply: the value of the key.
	//		- Null reply: if the key does not exist.
	Get(ctx context.Context, key string) StringCmd

	// GetRange
	// Available since: 2.4.0
	// Time complexity: O(N) where N is the length of the returned string. The complexity is ultimately determined by the returned length,
	//					but because creating a substring from an existing string is very cheap, it can be considered O(1) for small strings.
	// ACL categories: @read @string @slow
	// RESP2 / RESP3 Reply:
	// 	- Bulk string reply: The substring of the string value stored at key, determined by the offsets start and end (both are inclusive).
	GetRange(ctx context.Context, key string, start, end int64) StringCmd

	// StrLen
	// Available since: 2.2.0
	// Time complexity: O(1)
	// ACL categories: @read @string @fast
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the length of the string stored at key, or 0 when the key does not exist.
	StrLen(ctx context.Context, key string) IntCmd
}

func (c *client) Append(ctx context.Context, key, value string) IntCmd {
	ctx = c.handler.before(ctx, CommandAppend)
	r := c.adapter.Append(ctx, key, value)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Decr(ctx context.Context, key string) IntCmd {
	ctx = c.handler.before(ctx, CommandDecr)
	r := c.adapter.Decr(ctx, key)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) DecrBy(ctx context.Context, key string, decrement int64) IntCmd {
	ctx = c.handler.before(ctx, CommandDecrBy)
	r := c.adapter.DecrBy(ctx, key, decrement)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Get(ctx context.Context, key string) StringCmd {
	ctx = c.handler.before(ctx, CommandGet)
	var r StringCmd
	if c.ttl > 0 {
		r = newStringCmd(c.Do(ctx, c.builder.GetCompleted(key)))
	} else {
		r = c.adapter.Get(ctx, key)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) GetDel(ctx context.Context, key string) StringCmd {
	ctx = c.handler.before(ctx, CommandGetDel)
	r := c.adapter.GetDel(ctx, key)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) GetEx(ctx context.Context, key string, expiration time.Duration) StringCmd {
	ctx = c.handler.before(ctx, CommandGetEx)
	r := c.adapter.GetEx(ctx, key, expiration)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) GetRange(ctx context.Context, key string, start, end int64) StringCmd {
	ctx = c.handler.before(ctx, CommandGetRange)
	var r StringCmd
	if c.ttl > 0 {
		r = newStringCmd(c.Do(ctx, c.builder.GetRangeCompleted(key, start, end)))
	} else {
		r = c.adapter.GetRange(ctx, key, start, end)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) GetSet(ctx context.Context, key string, value any) StringCmd {
	ctx = c.handler.before(ctx, CommandGetSet)
	r := c.adapter.GetSet(ctx, key, value)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Incr(ctx context.Context, key string) IntCmd {
	ctx = c.handler.before(ctx, CommandIncr)
	r := c.adapter.Incr(ctx, key)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) IncrBy(ctx context.Context, key string, value int64) IntCmd {
	ctx = c.handler.before(ctx, CommandIncrBy)
	r := c.adapter.IncrBy(ctx, key, value)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) IncrByFloat(ctx context.Context, key string, value float64) FloatCmd {
	ctx = c.handler.before(ctx, CommandIncrByFloat)
	r := c.adapter.IncrByFloat(ctx, key, value)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) MGet(ctx context.Context, keys ...string) SliceCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandMGet, func() []string { return keys })
	r := c.adapter.MGet(ctx, keys...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) MGetIgnoreSlot(ctx context.Context, keys ...string) SliceCmd {
	if len(keys) <= 1 {
		return c.MGet(ctx, keys...)
	}
	var slot2Keys = make(map[uint16][]string)
	var keyIndexes = make(map[string]int)
	for i, key := range keys {
		keySlot := slot(key)
		slot2Keys[keySlot] = append(slot2Keys[keySlot], key)
		keyIndexes[key] = i
	}
	if len(slot2Keys) == 1 {
		return c.MGet(ctx, keys...)
	}
	var wg sync.WaitGroup
	var mx sync.Mutex
	var scs = make(map[uint16]SliceCmd)
	wg.Add(len(slot2Keys))
	for i, sameSlotKeys := range slot2Keys {
		go func(_i uint16, _keys []string) {
			ret := c.MGet(context.Background(), _keys...)
			mx.Lock()
			scs[_i] = ret
			mx.Unlock()
			wg.Done()
		}(i, sameSlotKeys)
	}
	wg.Wait()

	var res = make([]any, len(keys))
	for i, ret := range scs {
		if err := ret.Err(); err != nil {
			return newSliceCmdFromSlice(nil, err, keys...)
		}
		_values := ret.Val()
		for _i, _key := range slot2Keys[i] {
			res[keyIndexes[_key]] = _values[_i]
		}
	}
	return newSliceCmdFromSlice(res, nil, keys...)
}

func (c *client) MSet(ctx context.Context, values ...any) StatusCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandMSet, func() []string { return argsToSliceWithValues(values) })
	r := c.adapter.MSet(ctx, values...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) MSetNX(ctx context.Context, values ...any) BoolCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandMSetNX, func() []string { return argsToSliceWithValues(values) })
	r := c.adapter.MSetNX(ctx, values...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Set(ctx context.Context, key string, value any, expiration time.Duration) StatusCmd {
	if expiration == KeepTTL {
		ctx = c.handler.before(ctx, CommandSetKeepTTL)
	} else {
		ctx = c.handler.before(ctx, CommandSet)
	}
	r := c.adapter.Set(ctx, key, value, expiration)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SetEX(ctx context.Context, key string, value any, expiration time.Duration) StatusCmd {
	ctx = c.handler.before(ctx, CommandSetEX)
	r := c.adapter.SetEX(ctx, key, value, expiration)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SetNX(ctx context.Context, key string, value any, expiration time.Duration) BoolCmd {
	if expiration == KeepTTL {
		ctx = c.handler.before(ctx, CommandSetKeepTTL)
	} else {
		ctx = c.handler.before(ctx, CommandSetNX)
	}
	r := c.adapter.SetNX(ctx, key, value, expiration)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SetXX(ctx context.Context, key string, value any, expiration time.Duration) BoolCmd {
	if expiration == KeepTTL {
		ctx = c.handler.before(ctx, CommandSetKeepTTL)
	} else {
		ctx = c.handler.before(ctx, CommandSetXX)
	}
	r := c.adapter.SetXX(ctx, key, value, expiration)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SetArgs(ctx context.Context, key string, value any, a SetArgs) StatusCmd {
	m := strings.ToUpper(a.Mode)
	if a.Get && m == NX {
		ctx = c.handler.before(ctx, CommandSetNXGet)
	} else if a.Get || !a.ExpireAt.IsZero() {
		ctx = c.handler.before(ctx, CommandSetGet)
	} else if a.KeepTTL {
		ctx = c.handler.before(ctx, CommandSetKeepTTL)
	} else if m == NX {
		ctx = c.handler.before(ctx, CommandSetArgsNX)
	} else if m == XX {
		ctx = c.handler.before(ctx, CommandSetXX)
	} else if a.TTL > 0 {
		ctx = c.handler.before(ctx, CommandSetArgsEX)
	} else {
		ctx = c.handler.before(ctx, CommandSet)
	}
	r := c.adapter.SetArgs(ctx, key, value, a)
	c.handler.after(ctx, r.Err())

	return r
}

func (c *client) SetRange(ctx context.Context, key string, offset int64, value string) IntCmd {
	ctx = c.handler.before(ctx, CommandSetRange)
	r := c.adapter.SetRange(ctx, key, offset, value)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) StrLen(ctx context.Context, key string) IntCmd {
	ctx = c.handler.before(ctx, CommandStrLen)
	var r IntCmd
	if c.ttl > 0 {
		r = newIntCmd(c.Do(ctx, c.builder.StrLenCompleted(key)))
	} else {
		r = c.adapter.StrLen(ctx, key)
	}
	c.handler.after(ctx, r.Err())
	return r
}
