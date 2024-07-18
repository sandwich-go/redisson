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
	// Removes the specified fields from the hash stored at key. Specified fields that do not exist within this hash are ignored. If key does not exist, it is treated as an empty hash and this command returns 0.
	// Return:
	// 	Integer reply: the number of fields that were removed from the hash, not including specified but non existing fields.
	HDel(ctx context.Context, key string, fields ...string) IntCmd

	// HIncrBy
	// Available since: 2.0.0
	// Time complexity: O(1)
	// ACL categories: @write @hash @fast
	// Increments the number stored at field in the hash stored at key by increment. If key does not exist, a new key holding a hash is created. If field does not exist the value is set to 0 before the operation is performed.
	// The range of values supported by HINCRBY is limited to 64 bit signed integers.
	// Return:
	// 	Integer reply: the value at field after the increment operation.
	HIncrBy(ctx context.Context, key, field string, incr int64) IntCmd

	// HIncrByFloat
	// Available since: 2.6.0
	// Time complexity: O(1)
	// ACL categories: @write @hash @fast
	// Increment the specified field of a hash stored at key, and representing a floating point number, by the specified increment. If the increment value is negative, the result is to have the hash field value decremented instead of incremented. If the field does not exist, it is set to 0 before performing the operation. An error is returned if one of the following conditions occur:
	// The field contains a value of the wrong type (not a string).
	// The current field content or the specified increment are not parsable as a double precision floating point number.
	// The exact behavior of this command is identical to the one of the INCRBYFLOAT command, please refer to the documentation of INCRBYFLOAT for further information.
	// Return:
	// 	Bulk string reply: the value of field after the increment.
	HIncrByFloat(ctx context.Context, key, field string, incr float64) FloatCmd

	// HMSet
	// Available since: 2.0.0
	// Time complexity: O(N) where N is the number of fields being set.
	// ACL categories: @write @hash @fast
	// As of Redis version 4.0.0, this command is regarded as deprecated.
	// It can be replaced by HSET with multiple field-value pairs when migrating or writing new code.
	// Sets the specified fields to their respective values in the hash stored at key. This command overwrites any specified fields already existing in the hash. If key does not exist, a new key holding a hash is created.
	// Return:
	//	Simple string reply
	HMSet(ctx context.Context, key string, values ...any) BoolCmd

	// HSet
	// Available since: 2.0.0
	// Time complexity: O(1) for each field/value pair added, so O(N) to add N field/value pairs when the command is called with multiple field/value pairs.
	// ACL categories: @write @hash @fast
	// Sets field in the hash stored at key to value. If key does not exist, a new key holding a hash is created. If field already exists in the hash, it is overwritten.
	// Return:
	// 	Integer reply: The number of fields that were added.
	HSet(ctx context.Context, key string, values ...any) IntCmd

	// HSetNX
	// Available since: 2.0.0
	// Time complexity: O(1)
	// ACL categories: @write @hash @fast
	// Sets field in the hash stored at key to value, only if field does not yet exist. If key does not exist, a new key holding a hash is created. If field already exists, this operation has no effect.
	// Return:
	// Integer reply, specifically:
	//	1 if field is a new field in the hash and value was set.
	//	0 if field already exists in the hash and no operation was performed.
	HSetNX(ctx context.Context, key, field string, value any) BoolCmd

	// HExpire
	// Available since: 7.4.0
	// Time complexity: O(N) where N is the number of arguments to the command
	// ACL categories: @write @hash @fast
	// Set an expiration (TTL or time to live) on one or more fields of a given hash key. You must specify at least one field. Field(s) will automatically be deleted from the hash key when their TTLs expire.
	HExpire(ctx context.Context, key string, seconds time.Duration, fields ...string) IntSliceCmd
	HExpireNX(ctx context.Context, key string, seconds time.Duration, fields ...string) IntSliceCmd
	HExpireXX(ctx context.Context, key string, seconds time.Duration, fields ...string) IntSliceCmd
	HExpireGT(ctx context.Context, key string, seconds time.Duration, fields ...string) IntSliceCmd
	HExpireLT(ctx context.Context, key string, seconds time.Duration, fields ...string) IntSliceCmd

	// HExpireAt
	// Available since: 7.4.0
	// Time complexity: O(N) where N is the number of arguments to the command
	// ACL categories: @write @hash @fast
	// HEXPIREAT has the same effect and semantics as HEXPIRE, but instead of specifying the number of seconds for the TTL (time to live),
	// it takes an absolute Unix timestamp in seconds since Unix epoch. A timestamp in the past will delete the field immediately.
	HExpireAt(ctx context.Context, key string, tm time.Time, fields ...string) IntSliceCmd
	HExpireAtNX(ctx context.Context, key string, tm time.Time, fields ...string) IntSliceCmd
	HExpireAtXX(ctx context.Context, key string, tm time.Time, fields ...string) IntSliceCmd
	HExpireAtGT(ctx context.Context, key string, tm time.Time, fields ...string) IntSliceCmd
	HExpireAtLT(ctx context.Context, key string, tm time.Time, fields ...string) IntSliceCmd

	// HPExpire
	// Available since: 7.4.0
	// Time complexity: O(N) where N is the number of arguments to the command
	// ACL categories: @write @hash @fast
	// This command works like HEXPIRE, but the expiration of a field is specified in milliseconds instead of seconds.
	HPExpire(ctx context.Context, key string, seconds time.Duration, fields ...string) IntSliceCmd
	HPExpireNX(ctx context.Context, key string, seconds time.Duration, fields ...string) IntSliceCmd
	HPExpireXX(ctx context.Context, key string, seconds time.Duration, fields ...string) IntSliceCmd
	HPExpireGT(ctx context.Context, key string, seconds time.Duration, fields ...string) IntSliceCmd
	HPExpireLT(ctx context.Context, key string, seconds time.Duration, fields ...string) IntSliceCmd

	// HPExpireAt
	// Available since: 7.4.0
	// Time complexity: O(N) where N is the number of arguments to the command
	// ACL categories: @write @hash @fast
	// HPEXPIREAT has the same effect and semantics as HEXPIREAT, but the Unix time at which the field will expire is specified in milliseconds since Unix epoch instead of seconds.
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
	// When called with just the key argument, return a random field from the hash value stored at key.
	// If the provided count argument is positive, return an array of distinct fields. The array's length is either count or the hash's number of fields (HLEN), whichever is lower.
	// If called with a negative count, the behavior changes and the command is allowed to return the same field multiple times. In this case, the number of returned fields is the absolute value of the specified count.
	// The optional WITHVALUES modifier changes the reply so it includes the respective values of the randomly selected hash fields.
	// Return:
	// 	Bulk string reply: without the additional count argument, the command returns a Bulk Reply with the randomly selected field, or nil when key does not exist.
	// 	Array reply: when the additional count argument is passed, the command returns an array of fields, or an empty array when key does not exist. If the WITHVALUES modifier is used, the reply is a list fields and their values from the hash.
	HRandField(ctx context.Context, key string, count int64) StringSliceCmd
	HRandFieldWithValues(ctx context.Context, key string, count int64) KeyValueSliceCmd

	// HScan
	// Available since: 2.8.0
	// Time complexity: O(1) for every call. O(N) for a complete iteration, including enough command calls for the cursor to return back to 0. N is the number of elements inside the collection..
	// ACL categories: @read @hash @slow
	// See https://redis.io/commands/scan/
	HScan(ctx context.Context, key string, cursor uint64, match string, count int64) ScanCmd

	// HExpireTime
	// Available since: 7.4.0
	// Time complexity: O(N) where N is the number of arguments to the command
	// ACL categories: @read, @hash, @fast
	// Returns the absolute Unix timestamp in seconds since Unix epoch at which the given key's field(s) will expire.
	HExpireTime(ctx context.Context, key string, fields ...string) DurationSliceCmd

	// HPersist
	// Available since: 7.4.0
	// Time complexity: O(N) where N is the number of arguments to the command
	// ACL categories: @read, @hash, @fast
	// Remove the existing expiration on a hash key's field(s),
	// turning the field(s) from volatile (a field with expiration set) to persistent (a field that will never expire as no TTL (time to live) is associated).
	HPersist(ctx context.Context, key string, fields ...string) IntSliceCmd

	// HPExpireTime
	// Available since: 7.4.0
	// Time complexity: O(N) where N is the number of arguments to the command
	// ACL categories: @read, @hash, @fast
	// HPEXPIRETIME has the same semantics as HEXPIRETIME, but returns the absolute Unix expiration timestamp in milliseconds since Unix epoch instead of seconds.
	HPExpireTime(ctx context.Context, key string, fields ...string) DurationSliceCmd

	// HTTL
	// Available since: 7.4.0
	// Time complexity: O(N) where N is the number of arguments to the command
	// ACL categories: @read, @hash, @fast
	// Returns the remaining TTL (time to live) of a hash key's field(s) that have a set expiration.
	// This introspection capability allows you to check how many seconds a given hash field will continue to be part of the hash key.
	HTTL(ctx context.Context, key string, fields ...string) DurationSliceCmd

	// HPTTL
	// Available since: 7.4.0
	// Time complexity: O(N) where N is the number of arguments to the command
	// ACL categories: @read, @hash, @fast
	// Like HTTL, this command returns the remaining TTL (time to live) of a field that has an expiration set, but in milliseconds instead of seconds.
	HPTTL(ctx context.Context, key string, fields ...string) DurationSliceCmd
}

type HashCacheCmdable interface {
	// HExists
	// Available since: 2.0.0
	// Time complexity: O(1)
	// ACL categories: @read @hash @fast
	// Returns if field is an existing field in the hash stored at key.
	// Return:
	// Integer reply, specifically:
	//	1 if the hash contains field.
	//	0 if the hash does not contain field, or key does not exist.
	HExists(ctx context.Context, key, field string) BoolCmd

	// HGet
	// Available since: 2.0.0
	// Time complexity: O(1)
	// ACL categories: @read @hash @fast
	// Returns the value associated with field in the hash stored at key.
	// Return:
	// 	Bulk string reply: the value associated with field, or nil when field is not present in the hash or key does not exist.
	HGet(ctx context.Context, key, field string) StringCmd

	// HGetAll
	// Available since: 2.0.0
	// Time complexity: O(N) where N is the size of the hash.
	// ACL categories: @read @hash @slow
	// Returns all fields and values of the hash stored at key. In the returned value, every field name is followed by its value, so the length of the reply is twice the size of the hash.
	// Return:
	// 	Array reply: list of fields and their values stored in the hash, or an empty list when key does not exist.
	HGetAll(ctx context.Context, key string) StringStringMapCmd

	// HKeys
	// Available since: 2.0.0
	// Time complexity: O(N) where N is the size of the hash.
	// ACL categories: @read @hash @slow
	// Returns all field names in the hash stored at key.
	// Return:
	// 	Array reply: list of fields in the hash, or an empty list when key does not exist.
	HKeys(ctx context.Context, key string) StringSliceCmd

	// HLen
	// Available since: 2.0.0
	// Time complexity: O(1)
	// ACL categories: @read @hash @fast
	// Returns the number of fields contained in the hash stored at key.
	// Return:
	// 	Integer reply: number of fields in the hash, or 0 when key does not exist.
	HLen(ctx context.Context, key string) IntCmd

	// HMGet
	// Available since: 2.0.0
	// Time complexity: O(N) where N is the number of fields being requested.
	// ACL categories: @read @hash @fast
	// Returns the values associated with the specified fields in the hash stored at key.
	// For every field that does not exist in the hash, a nil value is returned. Because non-existing keys are treated as empty hashes, running HMGET against a non-existing key will return a list of nil values.
	// Return:
	// 	Array reply: list of values associated with the given fields, in the same order as they are requested.
	HMGet(ctx context.Context, key string, fields ...string) SliceCmd

	// HVals
	// Available since: 2.0.0
	// Time complexity: O(N) where N is the size of the hash.
	// ACL categories: @read @hash @slow
	// Returns all values in the hash stored at key.
	// Return:
	// 	Array reply: list of values in the hash, or an empty list when key does not exist.
	HVals(ctx context.Context, key string) StringSliceCmd

	// HStrLen
	// Available since: 3.2.0
	// Time complexity: O(1)
	// ACL categories: @read, @hash, @fast
	// Returns the string length of the value associated with field in the hash stored at key. If the key or the field do not exist, 0 is returned.
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
		r = c.adapter.HGet(ctx, key, field)
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
		r = c.adapter.HKeys(ctx, key)
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
	r := c.adapter.HRandField(ctx, key, count)
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
		r = c.adapter.HVals(ctx, key)
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
