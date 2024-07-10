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
	// ACL categories: @keyspace @write @slow
	// This command copies the value stored at the source key to the destination key.
	// By default, the destination key is created in the logical database used by the connection. The DB option allows specifying an alternative logical database index for the destination key.
	// The command returns an error when the destination key already exists. The REPLACE option removes the destination key before copying the value to it.
	// Return:
	// Integer reply, specifically:
	//	1 if source was copied.
	//	0 if source was not copied.
	Copy(ctx context.Context, sourceKey string, destKey string, db int, replace bool) IntCmd

	// Del
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the number of keys that will be removed. When a key to remove holds a value other than a string, the individual complexity for this key is O(M) where M is the number of elements in the list, set, sorted set or hash. Removing a single key that holds a string value is O(1).
	// ACL categories: @keyspace @write @slow
	// Removes the specified keys. A key is ignored if it does not exist.
	// Return:
	// 	Integer reply: The number of keys that were removed.
	Del(ctx context.Context, keys ...string) IntCmd

	// Expire
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @keyspace @write @fast
	// Set a timeout on key. After the timeout has expired, the key will automatically be deleted. A key with an associated timeout is often said to be volatile in Redis terminology.
	// The timeout will only be cleared by commands that delete or overwrite the contents of the key, including DEL, SET, GETSET and all the *STORE commands. This means that all the operations that conceptually alter the value stored at the key without replacing it with a new one will leave the timeout untouched. For instance, incrementing the value of a key with INCR, pushing a new value into a list with LPUSH, or altering the field value of a hash with HSET are all operations that will leave the timeout untouched.
	// The timeout can also be cleared, turning the key back into a persistent key, using the PERSIST command.
	// If a key is renamed with RENAME, the associated time to live is transferred to the new key name.
	// If a key is overwritten by RENAME, like in the case of an existing key Key_A that is overwritten by a call like RENAME Key_B Key_A, it does not matter if the original Key_A had a timeout associated or not, the new key Key_A will inherit all the characteristics of Key_B.
	// Note that calling EXPIRE/PEXPIRE with a non-positive timeout or EXPIREAT/PEXPIREAT with a time in the past will result in the key being deleted rather than expired (accordingly, the emitted key event will be del, not expired).
	// Options
	// The EXPIRE command supports a set of options since Redis 7.0:
	// 	NX -- Set expiry only when the key has no expiry
	// 	XX -- Set expiry only when the key has an existing expiry
	// 	GT -- Set expiry only when the new expiry is greater than current one
	// 	LT -- Set expiry only when the new expiry is less than current one
	// A non-volatile key is treated as an infinite TTL for the purpose of GT and LT. The GT, LT and NX options are mutually exclusive.
	// Refreshing expires
	// It is possible to call EXPIRE using as argument a key that already has an existing expire set. In this case the time to live of a key is updated to the new value. There are many useful applications for this, an example is documented in the Navigation session pattern section below.
	// Differences in Redis prior 2.1.3
	// In Redis versions prior 2.1.3 altering a key with an expire set using a command altering its value had the effect of removing the key entirely. This semantics was needed because of limitations in the replication layer that are now fixed.
	// EXPIRE would return 0 and not alter the timeout for a key with a timeout set.
	// Return:
	// Integer reply, specifically:
	//	1 if the timeout was set.
	//	0 if the timeout was not set. e.g. key doesn't exist, or operation skipped due to the provided arguments.
	Expire(ctx context.Context, key string, expiration time.Duration) BoolCmd

	// ExpireAt
	// Available since: 1.2.0
	// Time complexity: O(1)
	// ACL categories: @keyspace @write @fast
	// EXPIREAT has the same effect and semantic as EXPIRE, but instead of specifying the number of seconds representing the TTL (time to live), it takes an absolute Unix timestamp (seconds since January 1, 1970). A timestamp in the past will delete the key immediately.
	// Please for the specific semantics of the command refer to the documentation of EXPIRE.
	// Background
	// EXPIREAT was introduced in order to convert relative timeouts to absolute timeouts for the AOF persistence mode. Of course, it can be used directly to specify that a given key should expire at a given time in the future.
	// Options
	// The EXPIREAT command supports a set of options since Redis 7.0:
	//	NX -- Set expiry only when the key has no expiry
	//	XX -- Set expiry only when the key has an existing expiry
	//	GT -- Set expiry only when the new expiry is greater than current one
	//	LT -- Set expiry only when the new expiry is less than current one
	// A non-volatile key is treated as an infinite TTL for the purpose of GT and LT. The GT, LT and NX options are mutually exclusive.
	// Return:
	// Integer reply, specifically:
	//	1 if the timeout was set.
	//	0 if the timeout was not set. e.g. key doesn't exist, or operation skipped due to the provided arguments.
	ExpireAt(ctx context.Context, key string, tm time.Time) BoolCmd

	// Migrate
	// Available since: 2.6.0
	// Time complexity: This command actually executes a DUMP+DEL in the source instance, and a RESTORE in the target instance. See the pages of these commands for time complexity. Also an O(N) data transfer between the two instances is performed.
	// ACL categories: @keyspace @write @slow @dangerous
	// See https://redis.io/commands/migrate/
	Migrate(ctx context.Context, host, port, key string, db int, timeout time.Duration) StatusCmd

	// Move
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @keyspace @write @fast
	// Move key from the currently selected database (see SELECT) to the specified destination database. When key already exists in the destination database, or it does not exist in the source database, it does nothing. It is possible to use MOVE as a locking primitive because of this.
	// Return:
	// Integer reply, specifically:
	//	1 if key was moved.
	//	0 if key was not moved.
	Move(ctx context.Context, key string, db int) BoolCmd

	// Persist
	// Available since: 2.2.0
	// Time complexity: O(1)
	// ACL categories: @keyspace @write @fast
	// Remove the existing timeout on key, turning the key from volatile (a key with an expire set) to persistent (a key that will never expire as no timeout is associated).
	// Return:
	// Integer reply, specifically:
	//	1 if the timeout was removed.
	//	0 if key does not exist or does not have an associated timeout.
	Persist(ctx context.Context, key string) BoolCmd

	// PExpire
	// Available since: 2.6.0
	// Time complexity: O(1)
	// ACL categories: @keyspace @write @fast
	// This command works exactly like EXPIRE but the time to live of the key is specified in milliseconds instead of seconds.
	// Options
	// The PEXPIRE command supports a set of options since Redis 7.0:
	//	NX -- Set expiry only when the key has no expiry
	//	XX -- Set expiry only when the key has an existing expiry
	//	GT -- Set expiry only when the new expiry is greater than current one
	//	LT -- Set expiry only when the new expiry is less than current one
	// A non-volatile key is treated as an infinite TTL for the purpose of GT and LT. The GT, LT and NX options are mutually exclusive.
	// Return:
	// Integer reply, specifically:
	//	1 if the timeout was set.
	//	0 if the timeout was not set. e.g. key doesn't exist, or operation skipped due to the provided arguments.
	PExpire(ctx context.Context, key string, expiration time.Duration) BoolCmd

	// PExpireAt
	// Available since: 2.6.0
	// Time complexity: O(1)
	// ACL categories: @keyspace @write @fast
	// PEXPIREAT has the same effect and semantic as EXPIREAT, but the Unix time at which the key will expire is specified in milliseconds instead of seconds.
	// Options
	// The PEXPIREAT command supports a set of options since Redis 7.0:
	//	NX -- Set expiry only when the key has no expiry
	//	XX -- Set expiry only when the key has an existing expiry
	//	GT -- Set expiry only when the new expiry is greater than current one
	//	LT -- Set expiry only when the new expiry is less than current one
	// A non-volatile key is treated as an infinite TTL for the purpose of GT and LT. The GT, LT and NX options are mutually exclusive.
	// Return:
	// Integer reply, specifically:
	//	1 if the timeout was set.
	//	0 if the timeout was not set. e.g. key doesn't exist, or operation skipped due to the provided arguments.
	PExpireAt(ctx context.Context, key string, tm time.Time) BoolCmd

	// Rename
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @keyspace @write @slow
	// Renames key to newkey. It returns an error when key does not exist. If newkey already exists it is overwritten, when this happens RENAME executes an implicit DEL operation, so if the deleted key contains a very big value it may cause high latency even if RENAME itself is usually a constant-time operation.
	// In Cluster mode, both key and newkey must be in the same hash slot, meaning that in practice only keys that have the same hash tag can be reliably renamed in cluster.
	// Return:
	// 	Simple string reply
	Rename(ctx context.Context, key, newkey string) StatusCmd

	// RenameNX
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @keyspace @write @fast
	// Renames key to newkey if newkey does not yet exist. It returns an error when key does not exist.
	// In Cluster mode, both key and newkey must be in the same hash slot, meaning that in practice only keys that have the same hash tag can be reliably renamed in cluster.
	// Return:
	// Integer reply, specifically:
	//	1 if key was renamed to newkey.
	//	0 if newkey already exists.
	RenameNX(ctx context.Context, key, newkey string) BoolCmd

	// Restore
	// Available since: 2.6.0
	// Time complexity: O(1) to create the new key and additional O(NM) to reconstruct the serialized value, where N is the number of Redis objects composing the value and M their average size. For small string values the time complexity is thus O(1)+O(1M) where M is small, so simply O(1). However for sorted set values the complexity is O(NMlog(N)) because inserting values into sorted sets is O(log(N)).
	// ACL categories: @keyspace @write @slow @dangerous
	// Create a key associated with a value that is obtained by deserializing the provided serialized value (obtained via DUMP).
	// If ttl is 0 the key is created without any expire, otherwise the specified expire time (in milliseconds) is set.
	// If the ABSTTL modifier was used, ttl should represent an absolute Unix timestamp (in milliseconds) in which the key will expire.
	// For eviction purposes, you may use the IDLETIME or FREQ modifiers. See OBJECT for more information.
	// RESTORE will return a "Target key name is busy" error when key already exists unless you use the REPLACE modifier.
	// RESTORE checks the RDB version and data checksum. If they don't match an error is returned.
	// Return:
	//	Simple string reply: The command returns OK on success.
	Restore(ctx context.Context, key string, ttl time.Duration, value string) StatusCmd

	// RestoreReplace
	// Available since: 2.6.0
	// Time complexity: O(1) to create the new key and additional O(NM) to reconstruct the serialized value, where N is the number of Redis objects composing the value and M their average size. For small string values the time complexity is thus O(1)+O(1M) where M is small, so simply O(1). However for sorted set values the complexity is O(NMlog(N)) because inserting values into sorted sets is O(log(N)).
	// ACL categories: @keyspace @write @slow @dangerous
	RestoreReplace(ctx context.Context, key string, ttl time.Duration, value string) StatusCmd

	// SortStore
	// Available since: 1.0.0
	// Time complexity: O(N+M*log(M)) where N is the number of elements in the list or set to sort, and M the number of returned elements. When the elements are not sorted, complexity is O(N).
	// ACL categories: @write @set @sortedset @list @slow @dangerous
	// See https://redis.io/commands/sort/
	SortStore(ctx context.Context, key, store string, sort Sort) IntCmd

	// Unlink
	// Available since: 4.0.0
	// Time complexity: O(1) for each key removed regardless of its size. Then the command does O(N) work in a different thread in order to reclaim memory, where N is the number of allocations the deleted objects where composed of.
	// ACL categories: @keyspace @write @fast
	// This command is very similar to DEL: it removes the specified keys. Just like DEL a key is ignored if it does not exist. However the command performs the actual memory reclaiming in a different thread, so it is not blocking, while DEL is. This is where the command name comes from: the command just unlinks the keys from the keyspace. The actual removal will happen later asynchronously.
	// Return:
	//	Integer reply: The number of keys that were unlinked.
	Unlink(ctx context.Context, keys ...string) IntCmd
}

type GenericReader interface {
	// Dump
	// Available since: 2.6.0
	// Time complexity: O(1) to access the key and additional O(NM) to serialize it, where N is the number of Redis objects composing the value and M their average size. For small string values the time complexity is thus O(1)+O(1M) where M is small, so simply O(1).
	// ACL categories: @keyspace @read @slow
	// Serialize the value stored at key in a Redis-specific format and return it to the user. The returned value can be synthesized back into a Redis key using the RESTORE command.
	// The serialization format is opaque and non-standard, however it has a few semantic characteristics:
	// It contains a 64-bit checksum that is used to make sure errors will be detected. The RESTORE command makes sure to check the checksum before synthesizing a key using the serialized value.
	// Values are encoded in the same format used by RDB.
	// An RDB version is encoded inside the serialized value, so that different Redis versions with incompatible RDB formats will refuse to process the serialized value.
	// The serialized value does NOT contain expire information. In order to capture the time to live of the current value the PTTL command should be used.
	// If key does not exist a nil bulk reply is returned.
	// Return:
	//	Bulk string reply: the serialized value.
	Dump(ctx context.Context, key string) StringCmd

	// Keys
	// Available since: 1.0.0
	// Time complexity: O(N) with N being the number of keys in the database, under the assumption that the key names in the database and the given pattern have limited length.
	// ACL categories: @keyspace @read @slow @dangerous
	// Returns all keys matching pattern.
	// While the time complexity for this operation is O(N), the constant times are fairly low. For example, Redis running on an entry level laptop can scan a 1 million key database in 40 milliseconds.
	// Warning: consider KEYS as a command that should only be used in production environments with extreme care. It may ruin performance when it is executed against large databases. This command is intended for debugging and special operations, such as changing your keyspace layout. Don't use KEYS in your regular application code. If you're looking for a way to find keys in a subset of your keyspace, consider using SCAN or sets.
	// Supported glob-style patterns:
	// h?llo matches hello, hallo and hxllo
	// h*llo matches hllo and heeeello
	// h[ae]llo matches hello and hallo, but not hillo
	// h[^e]llo matches hallo, hbllo, ... but not hello
	// h[a-b]llo matches hallo and hbllo
	// Use \ to escape special characters if you want to match them verbatim.
	// Return
	//	Array reply: list of keys matching pattern.
	Keys(ctx context.Context, pattern string) StringSliceCmd

	// ObjectRefCount
	// Available since: 2.2.3
	// Time complexity: O(1)
	// ACL categories: @keyspace @read @slow
	// This command returns the reference count of the stored at <key>.
	// Return:
	//	Integer reply
	//	The number of references.
	ObjectRefCount(ctx context.Context, key string) IntCmd

	// ObjectEncoding
	// Available since: 2.2.3
	// Time complexity: O(1)
	// ACL categories: @keyspace @read @slow
	// Returns the internal encoding for the Redis object stored at <key>
	// Redis objects can be encoded in different ways:
	// Strings can be encoded as:
	// raw, normal string encoding.
	// int, strings representing integers in a 64-bit signed interval, encoded in this way to save space.
	// embstr, an embedded string, which is an object where the internal simple dynamic string, sds, is an unmodifiable string allocated in the same chuck as the object itself. embstr can be strings with lengths up to the hardcoded limit of OBJ_ENCODING_EMBSTR_SIZE_LIMIT or 44 bytes.
	// Lists can be encoded as ziplist or linkedlist. The ziplist is the special representation that is used to save space for small lists.
	// Sets can be encoded as intset or hashtable. The intset is a special encoding used for small sets composed solely of integers.
	// Hashes can be encoded as ziplist or hashtable. The ziplist is a special encoding used for small hashes.
	// Sorted Sets can be encoded as ziplist or skiplist format. As for the List type small sorted sets can be specially encoded using ziplist, while the skiplist encoding is the one that works with sorted sets of any size.
	// All the specially encoded types are automatically converted to the general type once you perform an operation that makes it impossible for Redis to retain the space saving encoding.
	// Return:
	//	Bulk string reply: the encoding of the object, or nil if the key doesn't exist
	ObjectEncoding(ctx context.Context, key string) StringCmd

	// ObjectIdleTime
	// Available since: 2.2.3
	// Time complexity: O(1)
	// ACL categories: @keyspace @read @slow
	// This command returns the time in seconds since the last access to the value stored at <key>.
	// The command is only available when the maxmemory-policy configuration directive is not set to one of the LFU policies.
	// Return:
	//	Integer reply
	//	The idle time in seconds.
	ObjectIdleTime(ctx context.Context, key string) DurationCmd

	// RandomKey
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @keyspace @read @slow
	// Return a random key from the currently selected database.
	// Return:
	//	Bulk string reply: the random key, or nil when the database is empty.
	RandomKey(ctx context.Context) StringCmd

	// Scan
	// Available since: 2.8.0
	// Time complexity: O(1) for every call. O(N) for a complete iteration, including enough command calls for the cursor to return back to 0. N is the number of elements inside the collection.
	// ACL categories: @keyspace @read @slow
	// See https://redis.io/commands/scan/
	Scan(ctx context.Context, cursor uint64, match string, count int64) ScanCmd

	// ScanType
	// Available since: 6.0.0
	// Time complexity: O(1) for every call. O(N) for a complete iteration, including enough command calls for the cursor to return back to 0. N is the number of elements inside the collection.
	// ACL categories: @keyspace @read @slow
	// See https://redis.io/commands/scan/
	ScanType(ctx context.Context, cursor uint64, match string, count int64, keyType string) ScanCmd

	// Touch
	// Available since: 3.2.1
	// Time complexity: O(N) where N is the number of keys that will be touched.
	// ACL categories: @keyspace @read @fast
	// Alters the last access time of a key(s). A key is ignored if it does not exist.
	// Return:
	//	Integer reply: The number of keys that were touched.
	Touch(ctx context.Context, keys ...string) IntCmd

	// TTL
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @keyspace @read @fast
	// Returns the remaining time to live of a key that has a timeout. This introspection capability allows a Redis client to check how many seconds a given key will continue to be part of the dataset.
	// In Redis 2.6 or older the command returns -1 if the key does not exist or if the key exist but has no associated expire.
	// Starting with Redis 2.8 the return value in case of error changed:
	// The command returns -2 if the key does not exist.
	// The command returns -1 if the key exists but has no associated expire.
	// See also the PTTL command that returns the same information with milliseconds resolution (Only available in Redis 2.6 or greater).
	// Return:
	// 	Integer reply: TTL in seconds, or a negative value in order to signal an error (see the description above).
	TTL(ctx context.Context, key string) DurationCmd

	// PTTL
	// Available since: 2.6.0
	// Time complexity: O(1)
	// ACL categories: @keyspace @read @fast
	// Like TTL this command returns the remaining time to live of a key that has an expire set, with the sole difference that TTL returns the amount of remaining time in seconds while PTTL returns it in milliseconds.
	// In Redis 2.6 or older the command returns -1 if the key does not exist or if the key exist but has no associated expire.
	// Starting with Redis 2.8 the return value in case of error changed:
	// The command returns -2 if the key does not exist.
	// The command returns -1 if the key exists but has no associated expire.
	// Return:
	// 	Integer reply: TTL in milliseconds, or a negative value in order to signal an error (see the description above).
	PTTL(ctx context.Context, key string) DurationCmd

	// Sort
	// Available since: 1.0.0
	// Time complexity: O(N+M*log(M)) where N is the number of elements in the list or set to sort, and M the number of returned elements. When the elements are not sorted, complexity is O(N).
	// ACL categories: @write @set @sortedset @list @slow @dangerous
	// See https://redis.io/commands/sort/
	Sort(ctx context.Context, key string, sort Sort) StringSliceCmd

	// SortInterfaces
	// Available since: 1.0.0
	// Time complexity: O(N+M*log(M)) where N is the number of elements in the list or set to sort, and M the number of returned elements. When the elements are not sorted, complexity is O(N).
	// ACL categories: @write @set @sortedset @list @slow @dangerous
	// See https://redis.io/commands/sort/
	SortInterfaces(ctx context.Context, key string, sort Sort) SliceCmd
}

type GenericCacheCmdable interface {
	// Type
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @keyspace @read @fast
	// Returns the string representation of the type of the value stored at key. The different types that can be returned are: string, list, set, zset, hash and stream.
	// Return:
	// 	Simple string reply: type of key, or none when key does not exist.
	Type(ctx context.Context, key string) StatusCmd

	// Exists
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the number of keys to check.
	// ACL categories: @keyspace @read @fast
	// Returns if key exists.
	// The user should be aware that if the same existing key is mentioned in the arguments multiple times, it will be counted multiple times. So if somekey exists, EXISTS somekey somekey will return 2.
	// Return:
	// 	Integer reply, specifically the number of keys that exist from those specified as arguments.
	Exists(ctx context.Context, keys ...string) IntCmd
}

func (c *client) Copy(ctx context.Context, sourceKey string, destKey string, db int, replace bool) IntCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandCopy, func() []string { return appendString(sourceKey, destKey) })
	r := c.copy(ctx, sourceKey, destKey, db, replace)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Del(ctx context.Context, keys ...string) IntCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandDel, func() []string { return keys })
	r := newIntCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Del().Key(keys...).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Dump(ctx context.Context, key string) StringCmd {
	ctx = c.handler.before(ctx, CommandDump)
	r := newStringCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Dump().Key(key).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Exists(ctx context.Context, keys ...string) IntCmd {
	if len(keys) > 1 {
		ctx = c.handler.beforeWithKeys(ctx, CommandExistsMultipleKeys, func() []string { return keys })
	} else {
		ctx = c.handler.before(ctx, CommandExists)
	}
	r := newIntCmdFromResult(c.Do(ctx, c.getExistsCompleted(keys...)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Expire(ctx context.Context, key string, expiration time.Duration) BoolCmd {
	ctx = c.handler.before(ctx, CommandExpire)
	r := newBoolCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Expire().Key(key).Seconds(formatSec(expiration)).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ExpireAt(ctx context.Context, key string, tm time.Time) BoolCmd {
	ctx = c.handler.before(ctx, CommandExpireAt)
	r := newBoolCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Expireat().Key(key).Timestamp(tm.Unix()).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Keys(ctx context.Context, pattern string) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandKeys)
	r := newStringSliceCmdFromStringSliceCmd(c.adapter.Keys(ctx, pattern))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Migrate(ctx context.Context, host, port, key string, db int, timeout time.Duration) StatusCmd {
	ctx = c.handler.before(ctx, CommandMigrate)
	r := c.migrate(ctx, host, port, key, db, timeout)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Move(ctx context.Context, key string, db int) BoolCmd {
	ctx = c.handler.before(ctx, CommandMove)
	r := newBoolCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Move().Key(key).Db(int64(db)).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ObjectRefCount(ctx context.Context, key string) IntCmd {
	ctx = c.handler.before(ctx, CommandObjectRefCount)
	r := newIntCmdFromResult(c.cmd.Do(ctx, c.cmd.B().ObjectRefcount().Key(key).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ObjectEncoding(ctx context.Context, key string) StringCmd {
	ctx = c.handler.before(ctx, CommandObjectEncoding)
	r := newStringCmdFromResult(c.cmd.Do(ctx, c.cmd.B().ObjectEncoding().Key(key).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ObjectIdleTime(ctx context.Context, key string) DurationCmd {
	ctx = c.handler.before(ctx, CommandObjectIdleTime)
	r := newDurationCmdFromResult(c.cmd.Do(ctx, c.cmd.B().ObjectIdletime().Key(key).Build()), time.Second)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Persist(ctx context.Context, key string) BoolCmd {
	ctx = c.handler.before(ctx, CommandPersist)
	r := newBoolCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Persist().Key(key).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) PExpire(ctx context.Context, key string, expiration time.Duration) BoolCmd {
	ctx = c.handler.before(ctx, CommandPExpire)
	r := newBoolCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Pexpire().Key(key).Milliseconds(formatMs(expiration)).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) PExpireAt(ctx context.Context, key string, tm time.Time) BoolCmd {
	ctx = c.handler.before(ctx, CommandPExpireAt)
	r := newBoolCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Pexpireat().Key(key).MillisecondsTimestamp(tm.UnixNano()/int64(time.Millisecond)).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) PTTL(ctx context.Context, key string) DurationCmd {
	ctx = c.handler.before(ctx, CommandPTTL)
	r := newDurationCmdFromResult(c.cmd.Do(ctx, c.getPTTLCompleted(key)), time.Millisecond)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Rename(ctx context.Context, key, newkey string) StatusCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandRename, func() []string { return appendString(key, newkey) })
	r := newStatusCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Rename().Key(key).Newkey(newkey).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) RenameNX(ctx context.Context, key, newkey string) BoolCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandRenameNX, func() []string { return appendString(key, newkey) })
	r := newBoolCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Renamenx().Key(key).Newkey(newkey).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) RandomKey(ctx context.Context) StringCmd {
	ctx = c.handler.before(ctx, CommandRandomKey)
	r := newStringCmdFromStringCmd(c.adapter.RandomKey(ctx))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Restore(ctx context.Context, key string, ttl time.Duration, value string) StatusCmd {
	ctx = c.handler.before(ctx, CommandRestore)
	r := newStatusCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Restore().Key(key).Ttl(formatMs(ttl)).SerializedValue(value).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) RestoreReplace(ctx context.Context, key string, ttl time.Duration, value string) StatusCmd {
	ctx = c.handler.before(ctx, CommandRestoreReplace)
	r := newStatusCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Restore().Key(key).Ttl(formatMs(ttl)).SerializedValue(value).Replace().Build()))
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
	if len(keyType) > 0 {
		ctx = c.handler.before(ctx, CommandScanType)
	} else {
		ctx = c.handler.before(ctx, CommandScan)
	}
	r := c.adapter.ScanType(ctx, cursor, match, count, keyType)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Sort(ctx context.Context, key string, sort Sort) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandSort)
	r := newStringSliceCmdFromResult(c.cmd.Do(ctx, c.sort("SORT", key, sort)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SortStore(ctx context.Context, key, store string, sort Sort) IntCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandSort, func() []string { return appendString(key, store) })
	r := newIntCmdFromIntCmd(c.adapter.SortStore(ctx, key, store, toSort(sort)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SortInterfaces(ctx context.Context, key string, sort Sort) SliceCmd {
	ctx = c.handler.before(ctx, CommandSort)
	r := newSliceCmdFromSliceResult(c.cmd.Do(ctx, c.sort("SORT", key, sort)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Touch(ctx context.Context, keys ...string) IntCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandTouch, func() []string { return keys })
	r := newIntCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Touch().Key(keys...).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) TTL(ctx context.Context, key string) DurationCmd {
	ctx = c.handler.before(ctx, CommandTTL)
	r := newDurationCmdFromResult(c.cmd.Do(ctx, c.getTTLCompleted(key)), time.Second)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Type(ctx context.Context, key string) StatusCmd {
	ctx = c.handler.before(ctx, CommandType)
	r := newStatusCmdFromResult(c.Do(ctx, c.getTypeCompleted(key)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Unlink(ctx context.Context, keys ...string) IntCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandUnlink, func() []string { return keys })
	r := newIntCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Unlink().Key(keys...).Build()))
	c.handler.after(ctx, r.Err())
	return r
}
