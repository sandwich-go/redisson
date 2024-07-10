package redisson

import (
	"context"
	"strings"
	"time"
)

type StringCmdable interface {
	StringWriter
	StringReader
}

type StringWriter interface {
	// Append
	// Available since: 2.0.0
	// Time complexity: O(1). The amortized time complexity is O(1) assuming the appended value is small and the already present value is of any size, since the dynamic string library used by Redis will double the free space available on every reallocation.
	// ACL categories: @write @string @fast
	// If key already exists and is a string, this command appends the value at the end of the string. If key does not exist it is created and set as an empty string, so APPEND will be similar to SET in this special case.
	// Return:
	// 	Integer reply: the length of the string after the append operation.
	Append(ctx context.Context, key, value string) IntCmd

	// Decr
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @write @string @fast
	// Decrements the number stored at key by one. If the key does not exist, it is set to 0 before performing the operation. An error is returned if the key contains a value of the wrong type or contains a string that can not be represented as integer. This operation is limited to 64 bit signed integers.
	// See INCR for extra information on increment/decrement operations.
	// Return:
	//	Integer reply: the value of key after the decrement
	Decr(ctx context.Context, key string) IntCmd

	// DecrBy
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @write @string @fast
	// Decrements the number stored at key by decrement. If the key does not exist, it is set to 0 before performing the operation. An error is returned if the key contains a value of the wrong type or contains a string that can not be represented as integer. This operation is limited to 64 bit signed integers.
	// See INCR for extra information on increment/decrement operations.
	// Return:
	//	Integer reply: the value of key after the decrement
	DecrBy(ctx context.Context, key string, decrement int64) IntCmd

	// GetDel
	// Available since: 6.2.0
	// Time complexity: O(1)
	// ACL categories: @write @string @fast
	// Get the value of key and delete the key. This command is similar to GET, except for the fact that it also deletes the key on success (if and only if the key's value type is a string).
	// Return:
	// 	Bulk string reply: the value of key, nil when key does not exist, or an error if the key's value type isn't a string.
	GetDel(ctx context.Context, key string) StringCmd

	// GetEx
	// Available since: 6.2.0
	// Time complexity: O(1)
	// ACL categories: @write @string @fast
	// Get the value of key and optionally set its expiration. GETEX is similar to GET, but is a write command with additional options.
	// Options
	// The GETEX command supports a set of options that modify its behavior:
	//	EX seconds -- Set the specified expire time, in seconds.
	//	PX milliseconds -- Set the specified expire time, in milliseconds.
	// 	EXAT timestamp-seconds -- Set the specified Unix time at which the key will expire, in seconds.
	// 	PXAT timestamp-milliseconds -- Set the specified Unix time at which the key will expire, in milliseconds.
	// 	PERSIST -- Remove the time to live associated with the key.
	// Return:
	//	Bulk string reply: the value of key, or nil when key does not exist.
	GetEx(ctx context.Context, key string, expiration time.Duration) StringCmd

	// GetSet
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @write @string @fast
	// Atomically sets key to value and returns the old value stored at key. Returns an error when key exists but does not hold a string value. Any previous time to live associated with the key is discarded on successful SET operation.
	// Return:
	//	Bulk string reply: the old value stored at key, or nil when key did not exist.
	// As of Redis version 6.2.0, this command is regarded as deprecated.
	// It can be replaced by SET with the GET argument when migrating or writing new code.
	GetSet(ctx context.Context, key string, value interface{}) StringCmd

	// Incr
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @write @string @fast
	// Increments the number stored at key by one. If the key does not exist, it is set to 0 before performing the operation. An error is returned if the key contains a value of the wrong type or contains a string that can not be represented as integer. This operation is limited to 64 bit signed integers.
	// Note: this is a string operation because Redis does not have a dedicated integer type. The string stored at the key is interpreted as a base-10 64 bit signed integer to execute the operation.
	// Redis stores integers in their integer representation, so for string values that actually hold an integer, there is no overhead for storing the string representation of the integer.
	// Return:
	//	Integer reply: the value of key after the increment
	Incr(ctx context.Context, key string) IntCmd

	// IncrBy
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @write @string @fast
	// Increments the number stored at key by increment. If the key does not exist, it is set to 0 before performing the operation. An error is returned if the key contains a value of the wrong type or contains a string that can not be represented as integer. This operation is limited to 64 bit signed integers.
	// See INCR for extra information on increment/decrement operations.
	// Return:
	//	Integer reply: the value of key after the increment
	IncrBy(ctx context.Context, key string, value int64) IntCmd

	// IncrByFloat
	// Available since: 2.6.0
	// Time complexity: O(1)
	// ACL categories: @write @string @fast
	// Increment the string representing a floating point number stored at key by the specified increment. By using a negative increment value, the result is that the value stored at the key is decremented (by the obvious properties of addition). If the key does not exist, it is set to 0 before performing the operation. An error is returned if one of the following conditions occur:
	// The key contains a value of the wrong type (not a string).
	// The current key content or the specified increment are not parsable as a double precision floating point number.
	// If the command is successful the new incremented value is stored as the new value of the key (replacing the old one), and returned to the caller as a string.
	// Both the value already contained in the string key and the increment argument can be optionally provided in exponential notation, however the value computed after the increment is stored consistently in the same format, that is, an integer number followed (if needed) by a dot, and a variable number of digits representing the decimal part of the number. Trailing zeroes are always removed.
	// The precision of the output is fixed at 17 digits after the decimal point regardless of the actual internal precision of the computation.
	// Return:
	//	Bulk string reply: the value of key after the increment.
	IncrByFloat(ctx context.Context, key string, value float64) FloatCmd

	// MSet
	// Available since: 1.0.1
	// Time complexity: O(N) where N is the number of keys to set.
	// ACL categories: @write @string @slow
	// Sets the given keys to their respective values. MSET replaces existing values with new values, just as regular SET. See MSETNX if you don't want to overwrite existing values.
	// MSET is atomic, so all given keys are set at once. It is not possible for clients to see that some of the keys were updated while others are unchanged.
	// Return:
	//	Simple string reply: always OK since MSET can't fail.
	MSet(ctx context.Context, values ...interface{}) StatusCmd

	// MSetNX
	// Available since: 1.0.1
	//Time complexity: O(N) where N is the number of keys to set.
	//ACL categories: @write @string @slow
	//Sets the given keys to their respective values. MSETNX will not perform any operation at all even if just a single key already exists.
	//Because of this semantic MSETNX can be used in order to set different keys representing different fields of a unique logic object in a way that ensures that either all the fields or none at all are set.
	// MSETNX is atomic, so all given keys are set at once. It is not possible for clients to see that some of the keys were updated while others are unchanged.
	// Return:
	// Integer reply, specifically:
	//	1 if the all the keys were set.
	//	0 if no key was set (at least one key already existed).
	MSetNX(ctx context.Context, values ...interface{}) BoolCmd

	// Set
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @write @string @slow
	// Set key to hold the string value. If key already holds a value, it is overwritten, regardless of its type. Any previous time to live associated with the key is discarded on successful SET operation.
	// Options
	// The SET command supports a set of options that modify its behavior:
	//	EX seconds -- Set the specified expire time, in seconds. since Redis 2.6.12.
	//	PX milliseconds -- Set the specified expire time, in milliseconds. since Redis 2.6.12.
	//	EXAT timestamp-seconds -- Set the specified Unix time at which the key will expire, in seconds. since Redis 6.2.0.
	//	PXAT timestamp-milliseconds -- Set the specified Unix time at which the key will expire, in milliseconds. since Redis 6.2.0.
	//	NX -- Only set the key if it does not already exist. since Redis 2.6.12.
	//	XX -- Only set the key if it already exist. since Redis 2.6.12.
	//	KEEPTTL -- Retain the time to live associated with the key. since Redis 6.0.0.
	//	GET -- Return the old string stored at key, or nil if key did not exist. An error is returned and SET aborted if the value stored at key is not a string. since Redis 6.2.0.
	// Note: Since the SET command options can replace SETNX, SETEX, PSETEX, GETSET, it is possible that in future versions of Redis these commands will be deprecated and finally removed.
	// Return:
	// Simple string reply: OK if SET was executed correctly.
	// Null reply: (nil) if the SET operation was not performed because the user specified the NX or XX option but the condition was not met.
	// If the command is issued with the GET option, the above does not apply. It will instead reply as follows, regardless if the SET was actually performed:
	// Bulk string reply: the old string value stored at key.
	// Null reply: (nil) if the key did not exist.
	// Starting with Redis version 7.0.0: Allowed the NX and GET options to be used together.
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) StatusCmd

	// SetEX
	// Available since: 2.0.0
	// Time complexity: O(1)
	// ACL categories: @write @string @slow
	// Set key to hold the string value and set key to timeout after a given number of seconds. This command is equivalent to executing the following commands:
	// SET mykey value
	// EXPIRE mykey seconds
	// SETEX is atomic, and can be reproduced by using the previous two commands inside an MULTI / EXEC block. It is provided as a faster alternative to the given sequence of operations, because this operation is very common when Redis is used as a cache.
	// An error is returned when seconds is invalid.
	// Return:
	//	Simple string reply
	SetEX(ctx context.Context, key string, value interface{}, expiration time.Duration) StatusCmd

	// SetNX
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @write @string @fast
	// Set key to hold string value if key does not exist. In that case, it is equal to SET. When key already holds a value, no operation is performed. SETNX is short for "SET if Not eXists".
	// Return:
	// Integer reply, specifically:
	//	1 if the key was set
	//	0 if the key was not set
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) BoolCmd

	// SetXX
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @write @string @fast
	// Only set the key if it already exist.
	SetXX(ctx context.Context, key string, value interface{}, expiration time.Duration) BoolCmd

	// SetArgs
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @write @string @fast
	// See Set
	SetArgs(ctx context.Context, key string, value interface{}, a SetArgs) StatusCmd

	// SetRange
	// Available since: 2.2.0
	// Time complexity: O(1), not counting the time taken to copy the new string in place. Usually, this string is very small so the amortized complexity is O(1). Otherwise, complexity is O(M) with M being the length of the value argument.
	// ACL categories: @write @string @slow
	// Overwrites part of the string stored at key, starting at the specified offset, for the entire length of value. If the offset is larger than the current length of the string at key, the string is padded with zero-bytes to make offset fit. Non-existing keys are considered as empty strings, so this command will make sure it holds a string large enough to be able to set value at offset.
	// Note that the maximum offset that you can set is 2^29 -1 (536870911), as Redis Strings are limited to 512 megabytes. If you need to grow beyond this size, you can use multiple keys.
	// Warning: When setting the last possible byte and the string value stored at key does not yet hold a string value, or holds a small string value, Redis needs to allocate all intermediate memory which can block the server for some time. On a 2010 MacBook Pro, setting byte number 536870911 (512MB allocation) takes ~300ms, setting byte number 134217728 (128MB allocation) takes ~80ms, setting bit number 33554432 (32MB allocation) takes ~30ms and setting bit number 8388608 (8MB allocation) takes ~8ms. Note that once this first allocation is done, subsequent calls to SETRANGE for the same key will not have the allocation overhead.
	// Patterns
	// Thanks to SETRANGE and the analogous GETRANGE commands, you can use Redis strings as a linear array with O(1) random access. This is a very fast and efficient storage in many real world use cases.
	// Return:
	//	Integer reply: the length of the string after it was modified by the command.
	SetRange(ctx context.Context, key string, offset int64, value string) IntCmd
}

type StringReader interface{}

type StringCacheCmdable interface {
	// Get
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @read @string @fast
	// Get the value of key. If the key does not exist the special value nil is returned. An error is returned if the value stored at key is not a string, because GET only handles string values.
	// Return:
	//	Bulk string reply: the value of key, or nil when key does not exist.
	Get(ctx context.Context, key string) StringCmd

	// MGet
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the number of keys to retrieve.
	// ACL categories: @read @string @fast
	// Returns the values of all specified keys. For every key that does not hold a string value or does not exist, the special value nil is returned. Because of this, the operation never fails.
	// Return:
	//	Array reply: list of values at the specified keys.
	MGet(ctx context.Context, keys ...string) SliceCmd

	// GetRange
	// Available since: 2.4.0
	// Time complexity: O(N) where N is the length of the returned string. The complexity is ultimately determined by the returned length, but because creating a substring from an existing string is very cheap, it can be considered O(1) for small strings.
	// ACL categories: @read @string @slow
	// Returns the substring of the string value stored at key, determined by the offsets start and end (both are inclusive). Negative offsets can be used in order to provide an offset starting from the end of the string. So -1 means the last character, -2 the penultimate and so forth.
	// The function handles out of range requests by limiting the resulting range to the actual length of the string.
	// Return:
	//	Bulk string reply
	GetRange(ctx context.Context, key string, start, end int64) StringCmd

	// StrLen
	// Available since: 2.2.0
	// Time complexity: O(1)
	// ACL categories: @read @string @fast
	// Returns the length of the string value stored at key. An error is returned when key holds a non-string value.
	// Return:
	//	Integer reply: the length of the string at key, or 0 when key does not exist.
	StrLen(ctx context.Context, key string) IntCmd
}

func (c *client) Append(ctx context.Context, key, value string) IntCmd {
	ctx = c.handler.before(ctx, CommandAppend)
	r := newIntCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Append().Key(key).Value(value).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Decr(ctx context.Context, key string) IntCmd {
	ctx = c.handler.before(ctx, CommandDecr)
	r := newIntCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Decr().Key(key).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) DecrBy(ctx context.Context, key string, decrement int64) IntCmd {
	ctx = c.handler.before(ctx, CommandDecrBy)
	r := newIntCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Decrby().Key(key).Decrement(decrement).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Get(ctx context.Context, key string) StringCmd {
	ctx = c.handler.before(ctx, CommandGet)
	r := newStringCmdFromResult(c.Do(ctx, c.cmd.B().Get().Key(key).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) GetDel(ctx context.Context, key string) StringCmd {
	ctx = c.handler.before(ctx, CommandGetDel)
	r := newStringCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Getdel().Key(key).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) GetEx(ctx context.Context, key string, expiration time.Duration) StringCmd {
	ctx = c.handler.before(ctx, CommandGetEX)
	var r StringCmd
	if expiration > 0 {
		if usePrecise(expiration) {
			r = newStringCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Getex().Key(key).PxMilliseconds(formatMs(expiration)).Build()))
		} else {
			r = newStringCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Getex().Key(key).ExSeconds(formatSec(expiration)).Build()))
		}
	} else if expiration == 0 {
		r = newStringCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Getex().Key(key).Persist().Build()))
	} else {
		r = newStringCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Getex().Key(key).Build()))
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) GetRange(ctx context.Context, key string, start, end int64) StringCmd {
	ctx = c.handler.before(ctx, CommandGetRange)
	r := newStringCmdFromResult(c.Do(ctx, c.cmd.B().Getrange().Key(key).Start(start).End(end).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) GetSet(ctx context.Context, key string, value interface{}) StringCmd {
	ctx = c.handler.before(ctx, CommandGetSet)
	r := newStringCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Getset().Key(key).Value(str(value)).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Incr(ctx context.Context, key string) IntCmd {
	ctx = c.handler.before(ctx, CommandIncr)
	r := newIntCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Incr().Key(key).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) IncrBy(ctx context.Context, key string, value int64) IntCmd {
	ctx = c.handler.before(ctx, CommandIncrBy)
	r := newIntCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Incrby().Key(key).Increment(value).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) IncrByFloat(ctx context.Context, key string, value float64) FloatCmd {
	ctx = c.handler.before(ctx, CommandIncrByFloat)
	r := newFloatCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Incrbyfloat().Key(key).Increment(value).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) MGet(ctx context.Context, keys ...string) SliceCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandMGet, func() []string { return keys })
	r := newSliceCmdFromSliceResult(c.Do(ctx, c.cmd.B().Mget().Key(keys...).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) MSet(ctx context.Context, values ...interface{}) StatusCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandMSet, func() []string { return argsToSliceWithValues(values) })
	kv := c.cmd.B().Mset().KeyValue()
	args := argsToSlice(values)
	for i := 0; i < len(args); i += 2 {
		kv = kv.KeyValue(args[i], args[i+1])
	}
	r := newStatusCmdFromResult(c.cmd.Do(ctx, kv.Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) MSetNX(ctx context.Context, values ...interface{}) BoolCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandMSetNX, func() []string { return argsToSliceWithValues(values) })
	kv := c.cmd.B().Msetnx().KeyValue()
	args := argsToSlice(values)
	for i := 0; i < len(args); i += 2 {
		kv = kv.KeyValue(args[i], args[i+1])
	}
	r := newBoolCmdFromResult(c.cmd.Do(ctx, kv.Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) StatusCmd {
	if expiration == KeepTTL {
		ctx = c.handler.before(ctx, CommandSetKeepTTL)
	} else {
		ctx = c.handler.before(ctx, CommandSet)
	}
	var r StatusCmd
	if expiration > 0 {
		if usePrecise(expiration) {
			r = newStatusCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Set().Key(key).Value(str(value)).PxMilliseconds(formatMs(expiration)).Build()))
		} else {
			r = newStatusCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Set().Key(key).Value(str(value)).ExSeconds(formatSec(expiration)).Build()))
		}
	} else if expiration == KeepTTL {
		r = newStatusCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Set().Key(key).Value(str(value)).Keepttl().Build()))
	} else {
		r = newStatusCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Set().Key(key).Value(str(value)).Build()))
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SetEX(ctx context.Context, key string, value interface{}, expiration time.Duration) StatusCmd {
	ctx = c.handler.before(ctx, CommandSetex)
	r := newStatusCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Setex().Key(key).Seconds(formatSec(expiration)).Value(str(value)).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) BoolCmd {
	if expiration == KeepTTL {
		ctx = c.handler.before(ctx, CommandSetKeepTTL)
	} else {
		ctx = c.handler.before(ctx, CommandSetnx)
	}
	var r BoolCmd
	switch expiration {
	case 0:
		r = newBoolCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Setnx().Key(key).Value(str(value)).Build()))
	case KeepTTL:
		r = newBoolCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Set().Key(key).Value(str(value)).Nx().Keepttl().Build()))
	default:
		if usePrecise(expiration) {
			r = newBoolCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Set().Key(key).Value(str(value)).Nx().PxMilliseconds(formatMs(expiration)).Build()))
		} else {
			r = newBoolCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Set().Key(key).Value(str(value)).Nx().ExSeconds(formatSec(expiration)).Build()))
		}
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SetXX(ctx context.Context, key string, value interface{}, expiration time.Duration) BoolCmd {
	if expiration == KeepTTL {
		ctx = c.handler.before(ctx, CommandSetKeepTTL)
	} else {
		ctx = c.handler.before(ctx, CommandSetXX)
	}
	var r BoolCmd
	switch expiration {
	case 0:
		r = newBoolCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Set().Key(key).Value(str(value)).Xx().Build()))
	case KeepTTL:
		r = newBoolCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Set().Key(key).Value(str(value)).Xx().Keepttl().Build()))
	default:
		if usePrecise(expiration) {
			r = newBoolCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Set().Key(key).Value(str(value)).Xx().PxMilliseconds(formatMs(expiration)).Build()))
		} else {
			r = newBoolCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Set().Key(key).Value(str(value)).Xx().ExSeconds(formatSec(expiration)).Build()))
		}
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SetArgs(ctx context.Context, key string, value interface{}, a SetArgs) StatusCmd {
	if a.KeepTTL {
		ctx = c.handler.before(ctx, CommandSetKeepTTL)
	} else if !a.ExpireAt.IsZero() {
		ctx = c.handler.before(ctx, CommandSetEXAT)
	} else if a.Get {
		ctx = c.handler.before(ctx, CommandSetGet)
	} else if len(a.Mode) > 0 {
		if strings.ToUpper(a.Mode) == NX {
			ctx = c.handler.before(ctx, CommandSetNX)
		} else {
			ctx = c.handler.before(ctx, CommandSetXX)
		}
	} else {
		ctx = c.handler.before(ctx, CommandSet)
	}
	cmd := c.cmd.B().Arbitrary(SET).Keys(key).Args(str(value))
	if a.KeepTTL {
		cmd = cmd.Args(KEEPTTL)
	}
	if !a.ExpireAt.IsZero() {
		cmd = cmd.Args(EXAT, str(a.ExpireAt.Unix()))
	}
	if a.TTL > 0 {
		if usePrecise(a.TTL) {
			cmd = cmd.Args(PX, str(formatMs(a.TTL)))
		} else {
			cmd = cmd.Args(EX, str(formatSec(a.TTL)))
		}
	}
	if len(a.Mode) > 0 {
		cmd = cmd.Args(a.Mode)
	}
	if a.Get {
		cmd = cmd.Args(GET)
	}
	r := newStatusCmdFromResult(c.cmd.Do(ctx, cmd.Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SetRange(ctx context.Context, key string, offset int64, value string) IntCmd {
	ctx = c.handler.before(ctx, CommandSetRange)
	r := newIntCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Setrange().Key(key).Offset(offset).Value(value).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) StrLen(ctx context.Context, key string) IntCmd {
	ctx = c.handler.before(ctx, CommandStrLen)
	r := newIntCmdFromResult(c.Do(ctx, c.cmd.B().Strlen().Key(key).Build()))
	c.handler.after(ctx, r.Err())
	return r
}
