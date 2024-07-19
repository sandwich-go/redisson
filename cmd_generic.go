package redisson

import (
	"context"
	"time"
)

type GenericCmdable interface {
	GenericWriter
	GenericReader
}

type GenericWriter interface {
	// Copy
	// Available since: 6.2.0
	// Time complexity: O(N) worst case for collections, where N is the number of nested items. O(1) for string values.
	// ACL categories: @keyspace, @write, @slow
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- Integer reply: 1 if source was copied.
	//		- Integer reply: 0 if source was not copied.
	Copy(ctx context.Context, sourceKey string, destKey string, db int64, replace bool) IntCmd

	// Del
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the number of keys that will be removed. When a key to remove holds a value other than a string,
	//					the individual complexity for this key is O(M) where M is the number of elements in the list, set, sorted set or hash.
	//					Removing a single key that holds a string value is O(1).
	// ACL categories: @keyspace, @write, @slow
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the number of keys that were removed.
	Del(ctx context.Context, keys ...string) IntCmd

	// Expire
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @keyspace @write @fast
	// Options
	// 	- NX -- Set expiry only when the key has no expiry
	// 	- XX -- Set expiry only when the key has an existing expiry
	// 	- GT -- Set expiry only when the new expiry is greater than current one
	// 	- LT -- Set expiry only when the new expiry is less than current one
	// A non-volatile key is treated as an infinite TTL for the purpose of GT and LT. The GT, LT and NX options are mutually exclusive.
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- Integer reply: 0 if the timeout was not set; for example, the key doesn't exist, or the operation was skipped because of the provided arguments.
	//		- Integer reply: 1 if the timeout was set.
	// History:
	//	- Starting with Redis version 7.0.0: Added options: NX, XX, GT and LT.
	Expire(ctx context.Context, key string, seconds time.Duration) BoolCmd
	ExpireNX(ctx context.Context, key string, seconds time.Duration) BoolCmd
	ExpireXX(ctx context.Context, key string, seconds time.Duration) BoolCmd
	ExpireGT(ctx context.Context, key string, seconds time.Duration) BoolCmd
	ExpireLT(ctx context.Context, key string, seconds time.Duration) BoolCmd

	// ExpireAt
	// Available since: 1.2.0
	// Time complexity: O(1)
	// ACL categories: @keyspace @write @fast
	// Options
	// 	- NX -- Set expiry only when the key has no expiry
	// 	- XX -- Set expiry only when the key has an existing expiry
	// 	- GT -- Set expiry only when the new expiry is greater than current one
	// 	- LT -- Set expiry only when the new expiry is less than current one
	// A non-volatile key is treated as an infinite TTL for the purpose of GT and LT. The GT, LT and NX options are mutually exclusive.
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- Integer reply: 0 if the timeout was not set; for example, the key doesn't exist, or the operation was skipped because of the provided arguments.
	//		- Integer reply: 1 if the timeout was set.
	// History:
	//	- Starting with Redis version 7.0.0: Added options: NX, XX, GT and LT.
	ExpireAt(ctx context.Context, key string, tm time.Time) BoolCmd
	ExpireAtNX(ctx context.Context, key string, tm time.Time) BoolCmd
	ExpireAtXX(ctx context.Context, key string, tm time.Time) BoolCmd
	ExpireAtGT(ctx context.Context, key string, tm time.Time) BoolCmd
	ExpireAtLT(ctx context.Context, key string, tm time.Time) BoolCmd

	// Migrate
	// Available since: 2.6.0
	// Time complexity: This command actually executes a DUMP+DEL in the source instance, and a RESTORE in the target instance.
	//					See the pages of these commands for time complexity. Also an O(N) data transfer between the two instances is performed.
	// ACL categories: @keyspace @write @slow @dangerous
	// Options
	// 	- COPY -- Do not remove the key from the local instance.
	// 	- REPLACE -- Replace existing key on the remote instance.
	// 	- KEYS -- If the key argument is an empty string, the command will instead migrate all the keys that follow the KEYS option (see the above section for more info).
	// 	- AUTH -- Authenticate with the given password to the remote instance.
	// 	- AUTH2 -- Authenticate with the given username and password pair (Redis 6 or greater ACL auth style).
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- Simple string reply: OK on success.
	//		- Simple string reply: NOKEY when no keys were found in the source instance.
	// History:
	//	- Starting with Redis version 3.0.0: Added the COPY and REPLACE options.
	//	- Starting with Redis version 3.0.6: Added the KEYS option.
	//	- Starting with Redis version 4.0.7: Added the AUTH option.
	//	- Starting with Redis version 6.0.0: Added the AUTH2 option.
	Migrate(ctx context.Context, host string, port int64, key string, db int64, timeout time.Duration) StatusCmd

	// Move
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @keyspace @write @fast
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- Integer reply: 1 if key was moved.
	//		- Integer reply: 0 if key wasn't moved.
	Move(ctx context.Context, key string, db int64) BoolCmd

	// Persist
	// Available since: 2.2.0
	// Time complexity: O(1)
	// ACL categories: @keyspace @write @fast
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- Integer reply: 0 if key does not exist or does not have an associated timeout.
	//		- Integer reply: 1 if the timeout has been removed.
	Persist(ctx context.Context, key string) BoolCmd

	// PExpire
	// Available since: 2.6.0
	// Time complexity: O(1)
	// ACL categories: @keyspace @write @fast
	// Options
	// 	- NX -- Set expiry only when the key has no expiry
	// 	- XX -- Set expiry only when the key has an existing expiry
	// 	- GT -- Set expiry only when the new expiry is greater than current one
	// 	- LT -- Set expiry only when the new expiry is less than current one
	// A non-volatile key is treated as an infinite TTL for the purpose of GT and LT. The GT, LT and NX options are mutually exclusive.
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- Integer reply: 0if the timeout was not set. For example, if the key doesn't exist, or the operation skipped because of the provided arguments.
	//		- Integer reply: 1 if the timeout was set.
	// History:
	//	- Starting with Redis version 7.0.0: Added options: NX, XX, GT and LT.
	PExpire(ctx context.Context, key string, expiration time.Duration) BoolCmd
	PExpireNX(ctx context.Context, key string, expiration time.Duration) BoolCmd
	PExpireXX(ctx context.Context, key string, expiration time.Duration) BoolCmd
	PExpireGT(ctx context.Context, key string, expiration time.Duration) BoolCmd
	PExpireLT(ctx context.Context, key string, expiration time.Duration) BoolCmd

	// PExpireAt
	// Available since: 2.6.0
	// Time complexity: O(1)
	// ACL categories: @keyspace @write @fast
	// Options
	// 	- NX -- Set expiry only when the key has no expiry
	// 	- XX -- Set expiry only when the key has an existing expiry
	// 	- GT -- Set expiry only when the new expiry is greater than current one
	// 	- LT -- Set expiry only when the new expiry is less than current one
	// A non-volatile key is treated as an infinite TTL for the purpose of GT and LT. The GT, LT and NX options are mutually exclusive.
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- Integer reply: 1 if the timeout was set.
	//		- Integer reply: 0 if the timeout was not set. For example, if the key doesn't exist, or the operation was skipped due to the provided arguments.
	// History:
	//	- Starting with Redis version 7.0.0: Added options: NX, XX, GT and LT.
	PExpireAt(ctx context.Context, key string, tm time.Time) BoolCmd
	PExpireAtNX(ctx context.Context, key string, tm time.Time) BoolCmd
	PExpireAtXX(ctx context.Context, key string, tm time.Time) BoolCmd
	PExpireAtGT(ctx context.Context, key string, tm time.Time) BoolCmd
	PExpireAtLT(ctx context.Context, key string, tm time.Time) BoolCmd

	// Rename
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @keyspace @write @slow
	// RESP2 / RESP3 Reply:
	// 	- Simple string reply: OK.
	Rename(ctx context.Context, key, newkey string) StatusCmd

	// RenameNX
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @keyspace @write @fast
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- Integer reply: 1 if key was renamed to newkey.
	//		- Integer reply: 0 if newkey already exists.
	// History:
	//	- Starting with Redis version 3.2.0: The command no longer returns an error when source and destination names are the same.
	RenameNX(ctx context.Context, key, newkey string) BoolCmd

	// Restore
	// Available since: 2.6.0
	// Time complexity: O(1) to create the new key and additional O(NM) to reconstruct the serialized value,
	//					where N is the number of Redis objects composing the value and M their average size. For small string values the time complexity is thus O(1)+O(1M) where M is small,
	//					so simply O(1). However for sorted set values the complexity is O(NMlog(N)) because inserting values into sorted sets is O(log(N)).
	// ACL categories: @keyspace @write @slow @dangerous
	// RESP2 / RESP3 Reply:
	// 	- Simple string reply: OK.
	// History:
	//	- Starting with Redis version 3.0.0: Added the REPLACE modifier.
	//	- Starting with Redis version 5.0.0: Added the ABSTTL modifier.
	//	- Starting with Redis version 5.0.0: Added the IDLETIME and FREQ options.
	Restore(ctx context.Context, key string, ttl time.Duration, value string) StatusCmd
	RestoreReplace(ctx context.Context, key string, ttl time.Duration, value string) StatusCmd

	// Sort
	// Available since: 1.0.0
	// Time complexity: O(N+M*log(M)) where N is the number of elements in the list or set to sort,
	//					and M the number of returned elements. When the elements are not sorted, complexity is O(N).
	// ACL categories: @write @set @sortedset @list @slow @dangerous
	// RESP2 / RESP3 Reply:
	// 	- Array reply: without passing the STORE option, the command returns a list of sorted elements.
	//	- Integer reply: when the STORE option is specified, the command returns the number of sorted elements in the destination list.
	Sort(ctx context.Context, key string, sort Sort) StringSliceCmd
	SortStore(ctx context.Context, key, store string, sort Sort) IntCmd
	SortInterfaces(ctx context.Context, key string, sort Sort) SliceCmd

	// Unlink
	// Available since: 4.0.0
	// Time complexity: O(1) for each key removed regardless of its size.
	//					Then the command does O(N) work in a different thread in order to reclaim memory, where N is the number of allocations the deleted objects where composed of.
	// ACL categories: @keyspace @write @fast
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the number of keys that were unlinked.
	Unlink(ctx context.Context, keys ...string) IntCmd
}

type GenericReader interface {
	// ExpireTime
	// Available since: 7.0.0
	// Time complexity: O(1)
	// ACL categories: @keyspace @read @fast
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- Integer reply: the expiration Unix timestamp in seconds.
	//		- Integer reply: -1 if the key exists but has no associated expiration time.
	//		- Integer reply: -2 if the key does not exist.
	ExpireTime(ctx context.Context, key string) DurationCmd

	// PExpireTime
	// Available since: 7.0.0
	// Time complexity: O(1)
	// ACL categories: @keyspace @read @fast
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- Integer reply: Expiration Unix timestamp in milliseconds.
	//		- Integer reply: -1 if the key exists but has no associated expiration time.
	//		- Integer reply: -2 if the key does not exist.
	PExpireTime(ctx context.Context, key string) DurationCmd

	// Dump
	// Available since: 2.6.0
	// Time complexity: O(1) to access the key and additional O(NM) to serialize it,
	//					where N is the number of Redis objects composing the value and M their average size.
	//					For small string values the time complexity is thus O(1)+O(1M) where M is small, so simply O(1).
	// ACL categories: @keyspace @read @slow
	// RESP2 Reply:
	//	One of the following:
	//		- Bulk string reply: The serialized value of the key.
	//		- Nil reply: the key does not exist.
	// RESP3 Reply:
	//	One of the following:
	//		- Bulk string reply: the serialized value of the key.
	//		- Null reply: the key does not exist.
	Dump(ctx context.Context, key string) StringCmd

	// Exists
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the number of keys to check.
	// ACL categories: @keyspace @read @fast
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the number of keys that exist from those specified as arguments.
	// History:
	//	- Starting with Redis version 3.0.3: Accepts multiple key arguments.
	Exists(ctx context.Context, keys ...string) IntCmd

	// Keys
	// Available since: 1.0.0
	// Time complexity: O(N) with N being the number of keys in the database, under the assumption that the key names in the database and the given pattern have limited length.
	// ACL categories: @keyspace @read @slow @dangerous
	// Returns all keys matching pattern.
	// RESP2 / RESP3 Reply:
	// 	- Array reply: a list of keys matching pattern.
	Keys(ctx context.Context, pattern string) StringSliceCmd

	// ObjectRefCount
	// Available since: 2.2.3
	// Time complexity: O(1)
	// ACL categories: @keyspace @read @slow
	// RESP2 Reply:
	//	One of the following:
	//		- Integer reply: the number of references.
	//		- Nil reply: if key doesn't exist
	// RESP3 Reply:
	//	One of the following:
	//		- Integer reply: the number of references.
	//		- Null reply: if key doesn't exist.
	ObjectRefCount(ctx context.Context, key string) IntCmd

	// ObjectEncoding
	// Available since: 2.2.3
	// Time complexity: O(1)
	// ACL categories: @keyspace @read @slow
	// RESP2 Reply:
	//	One of the following:
	//		- Nil reply: if the key doesn't exist.
	//		- Bulk string reply: the encoding of the object.
	// RESP3 Reply:
	//	One of the following:
	//		- Null reply: if the key doesn't exist.
	//		- Bulk string reply: the encoding of the object.
	ObjectEncoding(ctx context.Context, key string) StringCmd

	// ObjectIdleTime
	// Available since: 2.2.3
	// Time complexity: O(1)
	// ACL categories: @keyspace @read @slow
	// RESP2 Reply:
	//	One of the following:
	//		- Integer reply: the idle time in seconds.
	//		- Nil reply: if key doesn't exist
	// RESP3 Reply:
	//	One of the following:
	//		- Integer reply: the idle time in seconds.
	//		- Null reply: if key doesn't exist.
	ObjectIdleTime(ctx context.Context, key string) DurationCmd

	// RandomKey
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @keyspace @read @slow
	// RESP2 Reply:
	//	One of the following:
	//		- Nil reply: when the database is empty.
	//		- Bulk string reply: a random key in the database.
	// RESP3 Reply:
	//	One of the following:
	//		- Null reply: when the database is empty.
	//		- Bulk string reply: a random key in the database.
	RandomKey(ctx context.Context) StringCmd

	// Scan
	// Available since: 2.8.0
	// Time complexity: O(1) for every call. O(N) for a complete iteration, including enough command calls for the cursor to return back to 0.
	//					N is the number of elements inside the collection.
	// ACL categories: @keyspace @read @slow
	// RESP2 / RESP3 Reply:
	//	Array reply: specifically, an array with two elements.
	//		- The first element is a Bulk string reply that represents an unsigned 64-bit number, the cursor.
	//		- The second element is an Array reply with the names of scanned keys.
	// History:
	//	- Starting with Redis version 6.0.0: Added the TYPE subcommand.
	Scan(ctx context.Context, cursor uint64, match string, count int64) ScanCmd
	ScanType(ctx context.Context, cursor uint64, match string, count int64, keyType string) ScanCmd

	// Touch
	// Available since: 3.2.1
	// Time complexity: O(N) where N is the number of keys that will be touched.
	// ACL categories: @keyspace @read @fast
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the number of touched keys.
	Touch(ctx context.Context, keys ...string) IntCmd

	// SortRO
	// Available since: 7.0.0
	// Time complexity: O(N+M*log(M)) where N is the number of elements in the list or set to sort,
	//					and M the number of returned elements. When the elements are not sorted, complexity is O(N).
	// ACL categories: @read, @set, @sortedset, @list, @slow, @dangerous
	// RESP2 / RESP3 Reply:
	// 	- Array reply: a list of sorted elements.
	SortRO(ctx context.Context, key string, sort Sort) StringSliceCmd
}

type GenericCacheCmdable interface {
	// Type
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @keyspace @read @fast
	// RESP2 / RESP3 Reply:
	// 	- Simple string reply: the type of key, or none when key doesn't exist.
	Type(ctx context.Context, key string) StatusCmd

	// TTL
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @keyspace @read @fast
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- Integer reply: TTL in seconds.
	//		- Integer reply: -1 if the key exists but has no associated expiration.
	//		- Integer reply: -2 if the key does not exist.
	// History:
	//	- Starting with Redis version 2.8.0: Added the -2 reply.
	TTL(ctx context.Context, key string) DurationCmd

	// PTTL
	// Available since: 2.6.0
	// Time complexity: O(1)
	// ACL categories: @keyspace @read @fast
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- Integer reply: TTL in milliseconds.
	//		- Integer reply: -1 if the key exists but has no associated expiration.
	//		- Integer reply: -2 if the key does not exist.
	// History:
	//	- Starting with Redis version 2.8.0: Added the -2 reply.
	PTTL(ctx context.Context, key string) DurationCmd
}

func (c *client) Copy(ctx context.Context, sourceKey string, destKey string, db int64, replace bool) IntCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandCopy, func() []string { return appendString(sourceKey, destKey) })
	r := c.adapter.Copy(ctx, sourceKey, destKey, db, replace)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Del(ctx context.Context, keys ...string) IntCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandDel, func() []string { return keys })
	r := c.adapter.Del(ctx, keys...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Dump(ctx context.Context, key string) StringCmd {
	ctx = c.handler.before(ctx, CommandDump)
	r := c.adapter.Dump(ctx, key)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Exists(ctx context.Context, keys ...string) IntCmd {
	if len(keys) > 1 {
		ctx = c.handler.beforeWithKeys(ctx, CommandMExists, func() []string { return keys })
	} else {
		ctx = c.handler.before(ctx, CommandExists)
	}
	r := c.adapter.Exists(ctx, keys...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Expire(ctx context.Context, key string, seconds time.Duration) BoolCmd {
	ctx = c.handler.before(ctx, CommandExpire)
	r := c.adapter.Expire(ctx, key, seconds)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ExpireNX(ctx context.Context, key string, seconds time.Duration) BoolCmd {
	ctx = c.handler.before(ctx, CommandExpireNX)
	r := c.adapter.ExpireNX(ctx, key, seconds)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ExpireXX(ctx context.Context, key string, seconds time.Duration) BoolCmd {
	ctx = c.handler.before(ctx, CommandExpireXX)
	r := c.adapter.ExpireXX(ctx, key, seconds)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ExpireGT(ctx context.Context, key string, seconds time.Duration) BoolCmd {
	ctx = c.handler.before(ctx, CommandExpireGT)
	r := c.adapter.ExpireGT(ctx, key, seconds)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ExpireLT(ctx context.Context, key string, seconds time.Duration) BoolCmd {
	ctx = c.handler.before(ctx, CommandExpireLT)
	r := c.adapter.ExpireLT(ctx, key, seconds)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ExpireAt(ctx context.Context, key string, tm time.Time) BoolCmd {
	ctx = c.handler.before(ctx, CommandExpireAt)
	r := c.adapter.ExpireAt(ctx, key, tm)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ExpireAtNX(ctx context.Context, key string, tm time.Time) BoolCmd {
	ctx = c.handler.before(ctx, CommandExpireAtNX)
	r := newBoolCmd(c.Do(ctx, c.builder.ExpireAtNXCompleted(key, tm)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ExpireAtXX(ctx context.Context, key string, tm time.Time) BoolCmd {
	ctx = c.handler.before(ctx, CommandExpireAtXX)
	r := newBoolCmd(c.Do(ctx, c.builder.ExpireAtXXCompleted(key, tm)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ExpireAtGT(ctx context.Context, key string, tm time.Time) BoolCmd {
	ctx = c.handler.before(ctx, CommandExpireAtGT)
	r := newBoolCmd(c.Do(ctx, c.builder.ExpireAtGTCompleted(key, tm)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ExpireAtLT(ctx context.Context, key string, tm time.Time) BoolCmd {
	ctx = c.handler.before(ctx, CommandExpireAtLT)
	r := newBoolCmd(c.Do(ctx, c.builder.ExpireAtLTCompleted(key, tm)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ExpireTime(ctx context.Context, key string) DurationCmd {
	ctx = c.handler.before(ctx, CommandExpireTime)
	r := c.adapter.ExpireTime(ctx, key)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Keys(ctx context.Context, pattern string) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandKeys)
	r := c.adapter.Keys(ctx, pattern)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Migrate(ctx context.Context, host string, port int64, key string, db int64, timeout time.Duration) StatusCmd {
	ctx = c.handler.before(ctx, CommandMigrate)
	r := c.adapter.Migrate(ctx, host, port, key, db, timeout)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Move(ctx context.Context, key string, db int64) BoolCmd {
	ctx = c.handler.before(ctx, CommandMove)
	r := c.adapter.Move(ctx, key, db)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ObjectRefCount(ctx context.Context, key string) IntCmd {
	ctx = c.handler.before(ctx, CommandObjectRefCount)
	r := c.adapter.ObjectRefCount(ctx, key)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ObjectEncoding(ctx context.Context, key string) StringCmd {
	ctx = c.handler.before(ctx, CommandObjectEncoding)
	r := c.adapter.ObjectEncoding(ctx, key)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ObjectIdleTime(ctx context.Context, key string) DurationCmd {
	ctx = c.handler.before(ctx, CommandObjectIdleTime)
	r := c.adapter.ObjectIdleTime(ctx, key)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Persist(ctx context.Context, key string) BoolCmd {
	ctx = c.handler.before(ctx, CommandPersist)
	r := c.adapter.Persist(ctx, key)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) PExpire(ctx context.Context, key string, milliseconds time.Duration) BoolCmd {
	ctx = c.handler.before(ctx, CommandPExpire)
	r := c.adapter.PExpire(ctx, key, milliseconds)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) PExpireNX(ctx context.Context, key string, milliseconds time.Duration) BoolCmd {
	ctx = c.handler.before(ctx, CommandPExpireNX)
	r := newBoolCmd(c.Do(ctx, c.builder.PExpireNXCompleted(key, milliseconds)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) PExpireXX(ctx context.Context, key string, milliseconds time.Duration) BoolCmd {
	ctx = c.handler.before(ctx, CommandPExpireXX)
	r := newBoolCmd(c.Do(ctx, c.builder.PExpireXXCompleted(key, milliseconds)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) PExpireGT(ctx context.Context, key string, milliseconds time.Duration) BoolCmd {
	ctx = c.handler.before(ctx, CommandPExpireGT)
	r := newBoolCmd(c.Do(ctx, c.builder.PExpireGTCompleted(key, milliseconds)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) PExpireLT(ctx context.Context, key string, milliseconds time.Duration) BoolCmd {
	ctx = c.handler.before(ctx, CommandPExpireLT)
	r := newBoolCmd(c.Do(ctx, c.builder.PExpireLTCompleted(key, milliseconds)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) PExpireAt(ctx context.Context, key string, tm time.Time) BoolCmd {
	ctx = c.handler.before(ctx, CommandPExpireAt)
	r := c.adapter.PExpireAt(ctx, key, tm)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) PExpireAtNX(ctx context.Context, key string, tm time.Time) BoolCmd {
	ctx = c.handler.before(ctx, CommandPExpireAtNX)
	r := newBoolCmd(c.Do(ctx, c.builder.PExpireAtNXCompleted(key, tm)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) PExpireAtXX(ctx context.Context, key string, tm time.Time) BoolCmd {
	ctx = c.handler.before(ctx, CommandPExpireAtXX)
	r := newBoolCmd(c.Do(ctx, c.builder.PExpireAtXXCompleted(key, tm)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) PExpireAtGT(ctx context.Context, key string, tm time.Time) BoolCmd {
	ctx = c.handler.before(ctx, CommandPExpireAtGT)
	r := newBoolCmd(c.Do(ctx, c.builder.PExpireAtGTCompleted(key, tm)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) PExpireAtLT(ctx context.Context, key string, tm time.Time) BoolCmd {
	ctx = c.handler.before(ctx, CommandPExpireAtLT)
	r := newBoolCmd(c.Do(ctx, c.builder.PExpireAtLTCompleted(key, tm)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) PExpireTime(ctx context.Context, key string) DurationCmd {
	ctx = c.handler.before(ctx, CommandPExpireAt)
	r := c.adapter.PExpireTime(ctx, key)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) PTTL(ctx context.Context, key string) DurationCmd {
	ctx = c.handler.before(ctx, CommandPTTL)
	var r DurationCmd
	if c.ttl > 0 {
		r = newDurationCmd(c.Do(ctx, c.builder.PTTLCompleted(key)), time.Millisecond)
	} else {
		r = c.adapter.PTTL(ctx, key)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Rename(ctx context.Context, key, newkey string) StatusCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandRename, func() []string { return appendString(key, newkey) })
	r := c.adapter.Rename(ctx, key, newkey)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) RenameNX(ctx context.Context, key, newkey string) BoolCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandRenameNX, func() []string { return appendString(key, newkey) })
	r := c.adapter.RenameNX(ctx, key, newkey)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) RandomKey(ctx context.Context) StringCmd {
	ctx = c.handler.before(ctx, CommandRandomKey)
	r := c.adapter.RandomKey(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Restore(ctx context.Context, key string, ttl time.Duration, value string) StatusCmd {
	ctx = c.handler.before(ctx, CommandRestore)
	r := c.adapter.Restore(ctx, key, ttl, value)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) RestoreReplace(ctx context.Context, key string, ttl time.Duration, value string) StatusCmd {
	ctx = c.handler.before(ctx, CommandRestoreReplace)
	r := c.adapter.RestoreReplace(ctx, key, ttl, value)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Scan(ctx context.Context, cursor uint64, match string, count int64) ScanCmd {
	ctx = c.handler.before(ctx, CommandScan)
	r := c.adapter.Scan(ctx, cursor, match, count)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ScanType(ctx context.Context, cursor uint64, match string, count int64, keyType string) ScanCmd {
	ctx = c.handler.before(ctx, CommandScanType)
	r := c.adapter.ScanType(ctx, cursor, match, count, keyType)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Sort(ctx context.Context, key string, sort Sort) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandSort)
	r := c.adapter.Sort(ctx, key, sort)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SortStore(ctx context.Context, key, store string, sort Sort) IntCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandSortStore, func() []string { return appendString(key, store) })
	r := c.adapter.SortStore(ctx, key, store, sort)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SortInterfaces(ctx context.Context, key string, sort Sort) SliceCmd {
	ctx = c.handler.before(ctx, CommandSort)
	r := c.adapter.SortInterfaces(ctx, key, sort)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SortRO(ctx context.Context, key string, sort Sort) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandSortRO)
	r := c.adapter.SortRO(ctx, key, sort)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Touch(ctx context.Context, keys ...string) IntCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandTouch, func() []string { return keys })
	r := c.adapter.Touch(ctx, keys...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) TTL(ctx context.Context, key string) DurationCmd {
	ctx = c.handler.before(ctx, CommandTTL)
	var r DurationCmd
	if c.ttl > 0 {
		r = newDurationCmd(c.Do(ctx, c.builder.TTLCompleted(key)), time.Second)
	} else {
		r = c.adapter.TTL(ctx, key)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Type(ctx context.Context, key string) StatusCmd {
	ctx = c.handler.before(ctx, CommandType)
	var r StatusCmd
	if c.ttl > 0 {
		r = newStatusCmd(c.Do(ctx, c.builder.TypeCompleted(key)))
	} else {
		r = c.adapter.Type(ctx, key)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Unlink(ctx context.Context, keys ...string) IntCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandUnlink, func() []string { return keys })
	r := c.adapter.Unlink(ctx, keys...)
	c.handler.after(ctx, r.Err())
	return r
}
