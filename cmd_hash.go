package redisson

import (
	"context"
	"time"
)

type HashCmdable interface {
	HashWriter
	HashReader
}

type HashWriter interface {
	// HDel
	// Available since: 2.0.0
	// Time complexity: O(N) where N is the number of fields to be removed.
	// ACL categories: @write @hash @fast
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the number of fields that were removed from the hash, excluding any specified but non-existing fields.
	// History:
	//	- Starting with Redis version 2.4.0: Accepts multiple field arguments.
	HDel(ctx context.Context, key string, fields ...string) IntCmd

	// HIncrBy
	// Available since: 2.0.0
	// Time complexity: O(1)
	// ACL categories: @write @hash @fast
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the value of the field after the increment operation.
	HIncrBy(ctx context.Context, key, field string, incr int64) IntCmd

	// HIncrByFloat
	// Available since: 2.6.0
	// Time complexity: O(1)
	// ACL categories: @write @hash @fast
	// RESP2 / RESP3 Reply:
	// 	- Bulk string reply: the value of the field after the increment operation.
	HIncrByFloat(ctx context.Context, key, field string, incr float64) FloatCmd

	// HMSet
	// Available since: 2.0.0
	// Time complexity: O(N) where N is the number of fields being set.
	// ACL categories: @write @hash @fast
	// RESP2 / RESP3 Reply:
	// 	- Simple string reply: OK.
	HMSet(ctx context.Context, key string, values ...any) BoolCmd

	// HSet
	// Available since: 2.0.0
	// Time complexity: O(1) for each field/value pair added, so O(N) to add N field/value pairs when the command is called with multiple field/value pairs.
	// ACL categories: @write @hash @fast
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the number of fields that were added.
	// History:
	//	- Starting with Redis version 4.0.0: Accepts multiple field and value arguments.
	HSet(ctx context.Context, key string, values ...any) IntCmd

	// HSetNX
	// Available since: 2.0.0
	// Time complexity: O(1)
	// ACL categories: @write @hash @fast
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- Integer reply: 0 if the field already exists in the hash and no operation was performed.
	//		- Integer reply: 1 if the field is a new field in the hash and the value was set.
	HSetNX(ctx context.Context, key, field string, value any) BoolCmd

	// HExpire
	// Available since: 7.4.0
	// Time complexity: O(N) where N is the number of arguments to the command
	// ACL categories: @write @hash @fast
	// Options
	// 	- NX -- For each specified field, set expiration only when the field has no expiration.
	// 	- XX -- For each specified field, set expiration only when the field has an existing expiration.
	// 	- GT -- For each specified field, set expiration only when the new expiration is greater than current one.
	// 	- LT -- For each specified field, set expiration only when the new expiration is less than current one.
	// A non-volatile field is treated as an infinite TTL for the purpose of GT and LT. The NX, XX, GT, and LT options are mutually exclusive.
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- Array reply. For each field:
	//			- Integer reply: -2 if no such field exists in the provided hash key, or the provided key does not exist.
	//			- Integer reply: 0 if the specified NX | XX | GT | LT condition has not been met.
	//			- Integer reply: 1 if the expiration time was set/updated.
	//			- Integer reply: 2 when HEXPIRE/HPEXPIRE is called with 0 seconds/milliseconds or when HEXPIREAT/HPEXPIREAT is called with a past Unix time in seconds/milliseconds.
	//		- Simple error reply:
	//			- if parsing failed, mandatory arguments are missing, unknown arguments are specified, or argument values are of the wrong type or out of range.
	//			- if the provided key exists but is not a hash.
	HExpire(ctx context.Context, key string, seconds time.Duration, fields ...string) IntSliceCmd
	HExpireNX(ctx context.Context, key string, seconds time.Duration, fields ...string) IntSliceCmd
	HExpireXX(ctx context.Context, key string, seconds time.Duration, fields ...string) IntSliceCmd
	HExpireGT(ctx context.Context, key string, seconds time.Duration, fields ...string) IntSliceCmd
	HExpireLT(ctx context.Context, key string, seconds time.Duration, fields ...string) IntSliceCmd

	// HExpireAt
	// Available since: 7.4.0
	// Time complexity: O(N) where N is the number of arguments to the command
	// ACL categories: @write @hash @fast
	// Options
	// 	- NX -- For each specified field, set expiration only when the field has no expiration.
	// 	- XX -- For each specified field, set expiration only when the field has an existing expiration.
	// 	- GT -- For each specified field, set expiration only when the new expiration is greater than current one.
	// 	- LT -- For each specified field, set expiration only when the new expiration is less than current one.
	// A non-volatile key is treated as an infinite TTL for the purposes of GT and LT. The NX, XX, GT, and LT options are mutually exclusive.
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- Array reply. For each field:
	//			- Integer reply: -2 if no such field exists in the provided hash key, or the provided key does not exist.
	//			- Integer reply: 0 if the specified NX, XX, GT, or LT condition has not been met.
	//			- Integer reply: 1 if the expiration time was set/updated.
	//			- Integer reply: 2 when HEXPIRE or HPEXPIRE is called with 0 seconds or milliseconds, or when HEXPIREAT or HPEXPIREAT is called with a past Unix time in seconds or milliseconds.
	//		- Simple error reply:
	//			- if parsing failed, mandatory arguments are missing, unknown arguments are specified, or argument values are of the wrong type or out of range.
	//			- if the provided key exists but is not a hash.
	HExpireAt(ctx context.Context, key string, tm time.Time, fields ...string) IntSliceCmd
	HExpireAtNX(ctx context.Context, key string, tm time.Time, fields ...string) IntSliceCmd
	HExpireAtXX(ctx context.Context, key string, tm time.Time, fields ...string) IntSliceCmd
	HExpireAtGT(ctx context.Context, key string, tm time.Time, fields ...string) IntSliceCmd
	HExpireAtLT(ctx context.Context, key string, tm time.Time, fields ...string) IntSliceCmd

	// HPExpire
	// Available since: 7.4.0
	// Time complexity: O(N) where N is the number of arguments to the command
	// ACL categories: @write @hash @fast
	// Options
	// 	- NX -- For each specified field, set expiration only when the field has no expiration.
	// 	- XX -- For each specified field, set expiration only when the field has an existing expiration.
	// 	- GT -- For each specified field, set expiration only when the new expiration is greater than current one.
	// 	- LT -- For each specified field, set expiration only when the new expiration is less than current one.
	// A non-volatile key is treated as an infinite TTL for the purposes of GT and LT. The NX, XX, GT, and LT options are mutually exclusive.
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- Array reply. For each field:
	//			- Integer reply: -2 if no such field exists in the provided hash key, or the provided key does not exist.
	//			- Integer reply: 0 if the specified NX, XX, GT, or LT condition has not been met.
	//			- Integer reply: 1 if the expiration time was set/updated.
	//			- Integer reply: 2 when HEXPIRE or HPEXPIRE is called with 0 seconds or milliseconds, or when HEXPIREAT or HPEXPIREAT is called with a past Unix time in seconds or milliseconds.
	//		- Simple error reply:
	//			- if parsing failed, mandatory arguments are missing, unknown arguments are specified, or argument values are of the wrong type or out of range.
	//			- if the provided key exists but is not a hash.
	HPExpire(ctx context.Context, key string, seconds time.Duration, fields ...string) IntSliceCmd
	HPExpireNX(ctx context.Context, key string, seconds time.Duration, fields ...string) IntSliceCmd
	HPExpireXX(ctx context.Context, key string, seconds time.Duration, fields ...string) IntSliceCmd
	HPExpireGT(ctx context.Context, key string, seconds time.Duration, fields ...string) IntSliceCmd
	HPExpireLT(ctx context.Context, key string, seconds time.Duration, fields ...string) IntSliceCmd

	// HPExpireAt
	// Available since: 7.4.0
	// Time complexity: O(N) where N is the number of arguments to the command
	// ACL categories: @write @hash @fast
	// Options
	// 	- NX -- For each specified field, set expiration only when the field has no expiration.
	// 	- XX -- For each specified field, set expiration only when the field has an existing expiration.
	// 	- GT -- For each specified field, set expiration only when the new expiration is greater than current one.
	// 	- LT -- For each specified field, set expiration only when the new expiration is less than current one.
	// A non-volatile key is treated as an infinite TTL for the purposes of GT and LT. The NX, XX, GT, and LT options are mutually exclusive.
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- Array reply. For each field:
	//			- Integer reply: -2 if no such field exists in the provided hash key, or the provided key does not exist.
	//			- Integer reply: 0 if the specified NX, XX, GT, or LT condition has not been met.
	//			- Integer reply: 1 if the expiration time was set/updated.
	//			- Integer reply: 2 when HEXPIRE or HPEXPIRE is called with 0 seconds or milliseconds, or when HEXPIREAT or HPEXPIREAT is called with a past Unix time in seconds or milliseconds.
	//		- Simple error reply:
	//			- if parsing failed, mandatory arguments are missing, unknown arguments are specified, or argument values are of the wrong type or out of range.
	//			- if the provided key exists but is not a hash.
	HPExpireAt(ctx context.Context, key string, tm time.Time, fields ...string) IntSliceCmd
	HPExpireAtNX(ctx context.Context, key string, tm time.Time, fields ...string) IntSliceCmd
	HPExpireAtXX(ctx context.Context, key string, tm time.Time, fields ...string) IntSliceCmd
	HPExpireAtGT(ctx context.Context, key string, tm time.Time, fields ...string) IntSliceCmd
	HPExpireAtLT(ctx context.Context, key string, tm time.Time, fields ...string) IntSliceCmd
}

type HashReader interface {
	// HRandField
	// Available since: 6.2.0
	// Time complexity: O(N) where N is the number of fields returned
	// ACL categories: @read @hash @slow
	// RESP2 Reply:
	//	Any of the following:
	//		- Nil reply: if the key doesn't exist
	//		- Bulk string reply: a single, randomly selected field when the count option is not used
	//		- Array reply: a list containing count fields when the count option is used, or an empty array if the key does not exists.
	//		- Array reply: a list of fields and their values when count and WITHVALUES were both used.
	// RESP3 Reply:
	//	Any of the following:
	//		- Null reply: if the key doesn't exist
	//		- Bulk string reply: a single, randomly selected field when the count option is not used
	//		- Array reply: a list containing count fields when the count option is used, or an empty array if the key does not exists.
	//		- Array reply: a list of fields and their values when count and WITHVALUES were both used.
	HRandField(ctx context.Context, key string, count int64) StringSliceCmd
	HRandFieldWithValues(ctx context.Context, key string, count int64) KeyValueSliceCmd

	// HScan
	// Available since: 2.8.0
	// Time complexity: O(1) for every call. O(N) for a complete iteration, including enough command calls for the cursor to return back to 0. N is the number of elements inside the collection.
	// ACL categories: @read @hash @slow
	// RESP2 / RESP3 Reply:
	// 	Array reply: a two-element array.
	//		- The first element is a Bulk string reply that represents an unsigned 64-bit number, the cursor.
	//		- The second element is an Array reply of field/value pairs that were scanned. When the NOVALUES flag (since Redis 7.4) is used, only the field names are returned.
	HScan(ctx context.Context, key string, cursor uint64, match string, count int64) ScanCmd

	// HExpireTime
	// Available since: 7.4.0
	// Time complexity: O(N) where N is the number of arguments to the command
	// ACL categories: @read, @hash, @fast
	// RESP2 / RESP3 Reply:
	//	- Array reply. For each field:
	//		- Integer reply: -2 if no such field exists in the provided hash key, or the provided key does not exist.
	//		- Integer reply: -1 if the field exists but has no associated expiration set.
	//		- Integer reply: the expiration (Unix timestamp) in seconds.
	HExpireTime(ctx context.Context, key string, fields ...string) DurationSliceCmd

	// HPersist
	// Available since: 7.4.0
	// Time complexity: O(N) where N is the number of arguments to the command
	// ACL categories: @read, @hash, @fast
	// RESP2 / RESP3 Reply:
	//	- Array reply. For each field:
	//		- Integer reply: -2 if no such field exists in the provided hash key, or the provided key does not exist.
	//		- Integer reply: -1 if the field exists but has no associated expiration set.
	//		- Integer reply: 1 the expiration was removed.
	HPersist(ctx context.Context, key string, fields ...string) IntSliceCmd

	// HPExpireTime
	// Available since: 7.4.0
	// Time complexity: O(N) where N is the number of arguments to the command
	// ACL categories: @read, @hash, @fast
	// RESP2 / RESP3 Reply:
	//	- Array reply. For each field:
	//		- Integer reply: -2 if no such field exists in the provided hash key, or the provided key does not exist.
	//		- Integer reply: -1 if the field exists but has no associated expiration set.
	//		- Integer reply: the expiration (Unix timestamp) in milliseconds.
	HPExpireTime(ctx context.Context, key string, fields ...string) DurationSliceCmd

	// HTTL
	// Available since: 7.4.0
	// Time complexity: O(N) where N is the number of arguments to the command
	// ACL categories: @read, @hash, @fast
	// RESP2 / RESP3 Reply:
	//	- Array reply. For each field:
	//		- Integer reply: -2 if no such field exists in the provided hash key, or the provided key does not exist.
	//		- Integer reply: -1 if the field exists but has no associated expiration set.
	//		- Integer reply: the TTL in seconds.
	HTTL(ctx context.Context, key string, fields ...string) DurationSliceCmd

	// HPTTL
	// Available since: 7.4.0
	// Time complexity: O(N) where N is the number of arguments to the command
	// ACL categories: @read, @hash, @fast
	// RESP2 / RESP3 Reply:
	//	- Array reply. For each field:
	//		- Integer reply: -2 if no such field exists in the provided hash key, or the provided key does not exist.
	//		- Integer reply: -1 if the field exists but has no associated expiration set.
	//		- Integer reply: the TTL in milliseconds.
	HPTTL(ctx context.Context, key string, fields ...string) DurationSliceCmd
}

type HashCacheCmdable interface {
	// HExists
	// Available since: 2.0.0
	// Time complexity: O(1)
	// ACL categories: @read @hash @fast
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- Integer reply: 0 if the hash does not contain the field, or the key does not exist.
	//		- Integer reply: 1 if the hash contains the field.
	HExists(ctx context.Context, key, field string) BoolCmd

	// HGet
	// Available since: 2.0.0
	// Time complexity: O(1)
	// ACL categories: @read @hash @fast
	// RESP2 Reply:
	//	One of the following:
	//		- Bulk string reply: The value associated with the field.
	//		- Nil reply: If the field is not present in the hash or key does not exist.
	// RESP3 Reply:
	//	One of the following:
	//		- Bulk string reply: The value associated with the field.
	//		- Null reply: If the field is not present in the hash or key does not exist.
	HGet(ctx context.Context, key, field string) StringCmd

	// HGetAll
	// Available since: 2.0.0
	// Time complexity: O(N) where N is the size of the hash.
	// ACL categories: @read @hash @slow
	// RESP2 Reply:
	// 	- Array reply: a list of fields and their values stored in the hash, or an empty list when key does not exist.
	// RESP3 Reply:
	//  - Map reply: a map of fields and their values stored in the hash, or an empty list when key does not exist.
	HGetAll(ctx context.Context, key string) StringStringMapCmd

	// HKeys
	// Available since: 2.0.0
	// Time complexity: O(N) where N is the size of the hash.
	// ACL categories: @read @hash @slow
	// RESP2 / RESP3 Reply:
	// 	- Array reply: a list of fields in the hash, or an empty list when the key does not exist
	HKeys(ctx context.Context, key string) StringSliceCmd

	// HLen
	// Available since: 2.0.0
	// Time complexity: O(1)
	// ACL categories: @read @hash @fast
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the number of the fields in the hash, or 0 when the key does not exist.
	HLen(ctx context.Context, key string) IntCmd

	// HMGet
	// Available since: 2.0.0
	// Time complexity: O(N) where N is the number of fields being requested.
	// ACL categories: @read @hash @fast
	// RESP2 / RESP3 Reply:
	// 	- Array reply: a list of values associated with the given fields, in the same order as they are requested.
	HMGet(ctx context.Context, key string, fields ...string) SliceCmd

	// HVals
	// Available since: 2.0.0
	// Time complexity: O(N) where N is the size of the hash.
	// ACL categories: @read @hash @slow
	// RESP2 / RESP3 Reply:
	// 	- Array reply: a list of values in the hash, or an empty list when the key does not exist
	HVals(ctx context.Context, key string) StringSliceCmd

	// HStrLen
	// Available since: 3.2.0
	// Time complexity: O(1)
	// ACL categories: @read, @hash, @fast
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the string length of the value associated with the field, or zero when the field isn't present in the hash or the key doesn't exist at all.
	HStrLen(ctx context.Context, key, field string) IntCmd
}

func (c *client) HDel(ctx context.Context, key string, fields ...string) IntCmd {
	if len(fields) > 1 {
		ctx = c.handler.before(ctx, CommandHMDel)
	} else {
		ctx = c.handler.before(ctx, CommandHDel)
	}
	r := c.adapter.HDel(ctx, key, fields...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HExists(ctx context.Context, key, field string) BoolCmd {
	ctx = c.handler.before(ctx, CommandHExists)
	var r BoolCmd
	if c.ttl > 0 {
		r = newBoolCmd(c.Do(ctx, c.builder.HExistsCompleted(key, field)))
	} else {
		r = c.adapter.HExists(ctx, key, field)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HExpire(ctx context.Context, key string, seconds time.Duration, fields ...string) IntSliceCmd {
	ctx = c.handler.before(ctx, CommandHExpire)
	r := newIntSliceCmd(c.Do(ctx, c.builder.HExpireCompleted(key, seconds, fields...)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HExpireNX(ctx context.Context, key string, seconds time.Duration, fields ...string) IntSliceCmd {
	ctx = c.handler.before(ctx, CommandHExpireNX)
	r := newIntSliceCmd(c.Do(ctx, c.builder.HExpireNXCompleted(key, seconds, fields...)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HExpireXX(ctx context.Context, key string, seconds time.Duration, fields ...string) IntSliceCmd {
	ctx = c.handler.before(ctx, CommandHExpireXX)
	r := newIntSliceCmd(c.Do(ctx, c.builder.HExpireXXCompleted(key, seconds, fields...)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HExpireGT(ctx context.Context, key string, seconds time.Duration, fields ...string) IntSliceCmd {
	ctx = c.handler.before(ctx, CommandHExpireGT)
	r := newIntSliceCmd(c.Do(ctx, c.builder.HExpireGTCompleted(key, seconds, fields...)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HExpireLT(ctx context.Context, key string, seconds time.Duration, fields ...string) IntSliceCmd {
	ctx = c.handler.before(ctx, CommandHExpireLT)
	r := newIntSliceCmd(c.Do(ctx, c.builder.HExpireLTCompleted(key, seconds, fields...)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HExpireAt(ctx context.Context, key string, tm time.Time, fields ...string) IntSliceCmd {
	ctx = c.handler.before(ctx, CommandHExpireAt)
	r := newIntSliceCmd(c.Do(ctx, c.builder.HExpireAtCompleted(key, tm, fields...)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HExpireAtNX(ctx context.Context, key string, tm time.Time, fields ...string) IntSliceCmd {
	ctx = c.handler.before(ctx, CommandHExpireAtNX)
	r := newIntSliceCmd(c.Do(ctx, c.builder.HExpireAtNXCompleted(key, tm, fields...)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HExpireAtXX(ctx context.Context, key string, tm time.Time, fields ...string) IntSliceCmd {
	ctx = c.handler.before(ctx, CommandHExpireAtXX)
	r := newIntSliceCmd(c.Do(ctx, c.builder.HExpireAtXXCompleted(key, tm, fields...)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HExpireAtGT(ctx context.Context, key string, tm time.Time, fields ...string) IntSliceCmd {
	ctx = c.handler.before(ctx, CommandHExpireAtGT)
	r := newIntSliceCmd(c.Do(ctx, c.builder.HExpireAtGTCompleted(key, tm, fields...)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HExpireAtLT(ctx context.Context, key string, tm time.Time, fields ...string) IntSliceCmd {
	ctx = c.handler.before(ctx, CommandHExpireAtLT)
	r := newIntSliceCmd(c.Do(ctx, c.builder.HExpireAtLTCompleted(key, tm, fields...)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HExpireTime(ctx context.Context, key string, fields ...string) DurationSliceCmd {
	ctx = c.handler.before(ctx, CommandHExpireTime)
	r := newDurationSliceCmd(c.Do(ctx, c.builder.HExpireTimeCompleted(key, fields...)), time.Second)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HGet(ctx context.Context, key, field string) StringCmd {
	ctx = c.handler.before(ctx, CommandHGet)
	var r StringCmd
	if c.ttl > 0 {
		r = newStringCmd(c.Do(ctx, c.builder.HGetCompleted(key, field)))
	} else {
		r = wrapStringCmd(c.adapter.HGet(ctx, key, field))
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HGetAll(ctx context.Context, key string) StringStringMapCmd {
	ctx = c.handler.before(ctx, CommandHGetAll)
	var r StringStringMapCmd
	if c.ttl > 0 {
		r = newStringStringMapCmd(c.Do(ctx, c.builder.HGetAllCompleted(key)))
	} else {
		r = c.adapter.HGetAll(ctx, key)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HIncrBy(ctx context.Context, key, field string, incr int64) IntCmd {
	ctx = c.handler.before(ctx, CommandHIncrBy)
	r := c.adapter.HIncrBy(ctx, key, field, incr)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HIncrByFloat(ctx context.Context, key, field string, incr float64) FloatCmd {
	ctx = c.handler.before(ctx, CommandHIncrByFloat)
	r := c.adapter.HIncrByFloat(ctx, key, field, incr)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HKeys(ctx context.Context, key string) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandHKeys)
	var r StringSliceCmd
	if c.ttl > 0 {
		r = newStringSliceCmd(c.Do(ctx, c.builder.HKeysCompleted(key)))
	} else {
		r = wrapStringSliceCmd(c.adapter.HKeys(ctx, key))
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HLen(ctx context.Context, key string) IntCmd {
	ctx = c.handler.before(ctx, CommandHLen)
	var r IntCmd
	if c.ttl > 0 {
		r = newIntCmd(c.Do(ctx, c.builder.HLenCompleted(key)))
	} else {
		r = c.adapter.HLen(ctx, key)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HMGet(ctx context.Context, key string, fields ...string) SliceCmd {
	ctx = c.handler.before(ctx, CommandHMGet)
	var r SliceCmd
	if c.ttl > 0 {
		r = newSliceCmd(c.Do(ctx, c.builder.HMGetCompleted(key, fields...)), false, fields...)
	} else {
		r = c.adapter.HMGet(ctx, key, fields...)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HMSet(ctx context.Context, key string, values ...any) BoolCmd {
	ctx = c.handler.before(ctx, CommandHMSet)
	r := c.adapter.HMSet(ctx, key, values...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HPersist(ctx context.Context, key string, fields ...string) IntSliceCmd {
	ctx = c.handler.before(ctx, CommandHPersist)
	r := newIntSliceCmd(c.Do(ctx, c.builder.HPersistCompleted(key, fields...)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HPExpire(ctx context.Context, key string, milliseconds time.Duration, fields ...string) IntSliceCmd {
	ctx = c.handler.before(ctx, CommandHPExpire)
	r := newIntSliceCmd(c.Do(ctx, c.builder.HPExpireCompleted(key, milliseconds, fields...)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HPExpireNX(ctx context.Context, key string, milliseconds time.Duration, fields ...string) IntSliceCmd {
	ctx = c.handler.before(ctx, CommandHPExpireNX)
	r := newIntSliceCmd(c.Do(ctx, c.builder.HPExpireNXCompleted(key, milliseconds, fields...)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HPExpireXX(ctx context.Context, key string, milliseconds time.Duration, fields ...string) IntSliceCmd {
	ctx = c.handler.before(ctx, CommandHPExpireXX)
	r := newIntSliceCmd(c.Do(ctx, c.builder.HPExpireXXCompleted(key, milliseconds, fields...)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HPExpireGT(ctx context.Context, key string, milliseconds time.Duration, fields ...string) IntSliceCmd {
	ctx = c.handler.before(ctx, CommandHPExpireGT)
	r := newIntSliceCmd(c.Do(ctx, c.builder.HPExpireGTCompleted(key, milliseconds, fields...)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HPExpireLT(ctx context.Context, key string, milliseconds time.Duration, fields ...string) IntSliceCmd {
	ctx = c.handler.before(ctx, CommandHPExpireLT)
	r := newIntSliceCmd(c.Do(ctx, c.builder.HPExpireLTCompleted(key, milliseconds, fields...)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HPExpireAt(ctx context.Context, key string, tm time.Time, fields ...string) IntSliceCmd {
	ctx = c.handler.before(ctx, CommandHPExpireAt)
	r := newIntSliceCmd(c.Do(ctx, c.builder.HPExpireAtCompleted(key, tm, fields...)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HPExpireAtNX(ctx context.Context, key string, tm time.Time, fields ...string) IntSliceCmd {
	ctx = c.handler.before(ctx, CommandHPExpireAtNX)
	r := newIntSliceCmd(c.Do(ctx, c.builder.HPExpireAtNXCompleted(key, tm, fields...)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HPExpireAtXX(ctx context.Context, key string, tm time.Time, fields ...string) IntSliceCmd {
	ctx = c.handler.before(ctx, CommandHPExpireAtXX)
	r := newIntSliceCmd(c.Do(ctx, c.builder.HPExpireAtXXCompleted(key, tm, fields...)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HPExpireAtGT(ctx context.Context, key string, tm time.Time, fields ...string) IntSliceCmd {
	ctx = c.handler.before(ctx, CommandHPExpireAtGT)
	r := newIntSliceCmd(c.Do(ctx, c.builder.HPExpireAtGTCompleted(key, tm, fields...)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HPExpireAtLT(ctx context.Context, key string, tm time.Time, fields ...string) IntSliceCmd {
	ctx = c.handler.before(ctx, CommandHPExpireAtLT)
	r := newIntSliceCmd(c.Do(ctx, c.builder.HPExpireAtLTCompleted(key, tm, fields...)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HPExpireTime(ctx context.Context, key string, fields ...string) DurationSliceCmd {
	ctx = c.handler.before(ctx, CommandHPExpireTime)
	r := newDurationSliceCmd(c.Do(ctx, c.builder.HPExpireTimeCompleted(key, fields...)), time.Millisecond)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HTTL(ctx context.Context, key string, fields ...string) DurationSliceCmd {
	ctx = c.handler.before(ctx, CommandHTTL)
	r := newDurationSliceCmd(c.Do(ctx, c.builder.HTTLCompleted(key, fields...)), time.Second)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HPTTL(ctx context.Context, key string, fields ...string) DurationSliceCmd {
	ctx = c.handler.before(ctx, CommandHPTTL)
	r := newDurationSliceCmd(c.Do(ctx, c.builder.HPTTLCompleted(key, fields...)), time.Millisecond)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HRandField(ctx context.Context, key string, count int64) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandHRandField)
	r := wrapStringSliceCmd(c.adapter.HRandField(ctx, key, count))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HRandFieldWithValues(ctx context.Context, key string, count int64) KeyValueSliceCmd {
	ctx = c.handler.before(ctx, CommandHRandFieldWithValues)
	r := c.adapter.HRandFieldWithValues(ctx, key, count)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HScan(ctx context.Context, key string, cursor uint64, match string, count int64) ScanCmd {
	ctx = c.handler.before(ctx, CommandHScan)
	r := c.adapter.HScan(ctx, key, cursor, match, count)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HSet(ctx context.Context, key string, values ...any) IntCmd {
	if len(values) > 2 {
		ctx = c.handler.before(ctx, CommandHMSetX)
	} else {
		ctx = c.handler.before(ctx, CommandHSet)
	}
	r := c.adapter.HSet(ctx, key, values...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HSetNX(ctx context.Context, key, field string, value any) BoolCmd {
	ctx = c.handler.before(ctx, CommandHSetNX)
	r := c.adapter.HSetNX(ctx, key, field, value)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HVals(ctx context.Context, key string) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandHVals)
	var r StringSliceCmd
	if c.ttl > 0 {
		r = newStringSliceCmd(c.Do(ctx, c.builder.HValsCompleted(key)))
	} else {
		r = wrapStringSliceCmd(c.adapter.HVals(ctx, key))
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HStrLen(ctx context.Context, key, field string) IntCmd {
	ctx = c.handler.before(ctx, CommandHStrLen)
	var r IntCmd
	r = newIntCmd(c.Do(ctx, c.builder.HStrLenCompleted(key, field)))
	c.handler.after(ctx, r.Err())
	return r
}
