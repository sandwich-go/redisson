package redisson

import (
	"context"
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
	HMSet(ctx context.Context, key string, values ...interface{}) BoolCmd

	// HSet
	// Available since: 2.0.0
	// Time complexity: O(1) for each field/value pair added, so O(N) to add N field/value pairs when the command is called with multiple field/value pairs.
	// ACL categories: @write @hash @fast
	// Sets field in the hash stored at key to value. If key does not exist, a new key holding a hash is created. If field already exists in the hash, it is overwritten.
	// Return:
	// 	Integer reply: The number of fields that were added.
	HSet(ctx context.Context, key string, values ...interface{}) IntCmd

	// HSetNX
	// Available since: 2.0.0
	// Time complexity: O(1)
	// ACL categories: @write @hash @fast
	// Sets field in the hash stored at key to value, only if field does not yet exist. If key does not exist, a new key holding a hash is created. If field already exists, this operation has no effect.
	// Return:
	// Integer reply, specifically:
	//	1 if field is a new field in the hash and value was set.
	//	0 if field already exists in the hash and no operation was performed.
	HSetNX(ctx context.Context, key, field string, value interface{}) BoolCmd
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
	HRandField(ctx context.Context, key string, count int, withValues bool) StringSliceCmd

	// HScan
	// Available since: 2.8.0
	// Time complexity: O(1) for every call. O(N) for a complete iteration, including enough command calls for the cursor to return back to 0. N is the number of elements inside the collection..
	// ACL categories: @read @hash @slow
	// See https://redis.io/commands/scan/
	HScan(ctx context.Context, key string, cursor uint64, match string, count int64) ScanCmd
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
}

func (c *client) HDel(ctx context.Context, key string, fields ...string) IntCmd {
	if len(fields) > 1 {
		ctx = c.handler.before(ctx, CommandHDelMultiple)
	} else {
		ctx = c.handler.before(ctx, CommandHDel)
	}
	r := c.cmdable.HDel(ctx, key, fields...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HExists(ctx context.Context, key, field string) BoolCmd {
	ctx = c.handler.before(ctx, CommandHExists)
	r := c.cacheCmdable.HExists(ctx, key, field)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HGet(ctx context.Context, key, field string) StringCmd {
	ctx = c.handler.before(ctx, CommandHGet)
	r := c.cacheCmdable.HGet(ctx, key, field)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HGetAll(ctx context.Context, key string) StringStringMapCmd {
	ctx = c.handler.before(ctx, CommandHGetAll)
	r := c.cacheCmdable.HGetAll(ctx, key)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HIncrBy(ctx context.Context, key, field string, incr int64) IntCmd {
	ctx = c.handler.before(ctx, CommandHIncrBy)
	r := c.cmdable.HIncrBy(ctx, key, field, incr)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HIncrByFloat(ctx context.Context, key, field string, incr float64) FloatCmd {
	ctx = c.handler.before(ctx, CommandHIncrByFloat)
	r := c.cmdable.HIncrByFloat(ctx, key, field, incr)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HKeys(ctx context.Context, key string) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandHKeys)
	r := c.cacheCmdable.HKeys(ctx, key)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HLen(ctx context.Context, key string) IntCmd {
	ctx = c.handler.before(ctx, CommandHLen)
	r := c.cacheCmdable.HLen(ctx, key)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HMGet(ctx context.Context, key string, fields ...string) SliceCmd {
	ctx = c.handler.before(ctx, CommandHMGet)
	r := c.cacheCmdable.HMGet(ctx, key, fields...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HMSet(ctx context.Context, key string, values ...interface{}) BoolCmd {
	ctx = c.handler.before(ctx, CommandHMSet)
	r := c.cmdable.HMSet(ctx, key, values...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HRandField(ctx context.Context, key string, count int, withValues bool) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandHRandField)
	r := c.cmdable.HRandField(ctx, key, count, withValues)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HScan(ctx context.Context, key string, cursor uint64, match string, count int64) ScanCmd {
	ctx = c.handler.before(ctx, CommandHScan)
	r := c.cmdable.HScan(ctx, key, cursor, match, count)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HSet(ctx context.Context, key string, values ...interface{}) IntCmd {
	if len(values) > 2 {
		ctx = c.handler.before(ctx, CommandHSetMultiple)
	} else {
		ctx = c.handler.before(ctx, CommandHSet)
	}
	r := c.cmdable.HSet(ctx, key, values...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HSetNX(ctx context.Context, key, field string, value interface{}) BoolCmd {
	ctx = c.handler.before(ctx, CommandHSetNX)
	r := c.cmdable.HSetNX(ctx, key, field, value)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) HVals(ctx context.Context, key string) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandHVals)
	r := c.cacheCmdable.HVals(ctx, key)
	c.handler.after(ctx, r.Err())
	return r
}
