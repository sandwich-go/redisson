package redisson

import (
	"context"
)

type SetCmdable interface {
	SetWriter
	SetReader
}

type SetWriter interface {
	// SAdd
	// Available since: 1.0.0
	// Time complexity: O(1) for each element added, so O(N) to add N elements when the command is called with multiple arguments.
	// ACL categories: @write @set @fast
	// Add the specified members to the set stored at key. Specified members that are already a member of this set are ignored. If key does not exist, a new set is created before adding the specified members.
	// An error is returned when the value stored at key is not a set.
	// Return:
	// 	Integer reply: the number of elements that were added to the set, not including all the elements already present in the set.
	SAdd(ctx context.Context, key string, members ...interface{}) IntCmd

	// SDiffStore
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the total number of elements in all given sets.
	// ACL categories: @write @set @slow
	// This command is equal to SDIFF, but instead of returning the resulting set, it is stored in destination.
	// If destination already exists, it is overwritten.
	// Return:
	//	Integer reply: the number of elements in the resulting set.
	SDiffStore(ctx context.Context, destination string, keys ...string) IntCmd

	// SInterStore
	// Available since: 1.0.0
	// Time complexity: O(N*M) worst case where N is the cardinality of the smallest set and M is the number of sets.
	// ACL categories: @write @set @slow
	// This command is equal to SINTER, but instead of returning the resulting set, it is stored in destination.
	// If destination already exists, it is overwritten.
	// Return:
	//	Integer reply: the number of elements in the resulting set.
	SInterStore(ctx context.Context, destination string, keys ...string) IntCmd

	// SMove
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @write @set @fast
	// Move member from the set at source to the set at destination. This operation is atomic. In every given moment the element will appear to be a member of source or destination for other clients.
	// If the source set does not exist or does not contain the specified element, no operation is performed and 0 is returned. Otherwise, the element is removed from the source set and added to the destination set. When the specified element already exists in the destination set, it is only removed from the source set.
	// An error is returned if source or destination does not hold a set value.
	// Return:
	// Integer reply, specifically:
	//	1 if the element is moved.
	//	0 if the element is not a member of source and no operation was performed.
	SMove(ctx context.Context, source, destination string, member interface{}) BoolCmd

	// SPop
	// Available since: 1.0.0
	// Time complexity: Without the count argument O(1), otherwise O(N) where N is the value of the passed count.
	// ACL categories: @write @set @fast
	// Removes and returns one or more random members from the set value store at key.
	// This operation is similar to SRANDMEMBER, that returns one or more random elements from a set but does not remove it.
	// By default, the command pops a single member from the set. When provided with the optional count argument, the reply will consist of up to count members, depending on the set's cardinality.
	// Return:
	// When called without the count argument:
	// Bulk string reply: the removed member, or nil when key does not exist.
	// When called with the count argument:
	// 	Array reply: the removed members, or an empty array when key does not exist.
	SPop(ctx context.Context, key string) StringCmd

	// SPopN
	// Available since: 3.2.0
	// Time complexity: Without the count argument O(1), otherwise O(N) where N is the value of the passed count.
	// ACL categories: @write @set @fast
	// see SPop
	SPopN(ctx context.Context, key string, count int64) StringSliceCmd

	// SRem
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the number of members to be removed.
	// ACL categories: @write @set @fast
	// Remove the specified members from the set stored at key. Specified members that are not a member of this set are ignored. If key does not exist, it is treated as an empty set and this command returns 0.
	// An error is returned when the value stored at key is not a set.
	// Return:
	// 	Integer reply: the number of members that were removed from the set, not including non existing members.
	SRem(ctx context.Context, key string, members ...interface{}) IntCmd

	// SUnionStore
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the total number of elements in all given sets.
	// ACL categories: @write @set @slow
	// This command is equal to SUNION, but instead of returning the resulting set, it is stored in destination.
	// If destination already exists, it is overwritten.
	// Return:
	// Integer reply: the number of elements in the resulting set.
	SUnionStore(ctx context.Context, destination string, keys ...string) IntCmd
}

type SetReader interface {
	// SDiff
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the total number of elements in all given sets.
	// ACL categories: @read @set @slow
	// Returns the members of the set resulting from the difference between the first set and all the successive sets.
	// For example:
	// key1 = {a,b,c,d}
	// key2 = {c}
	// key3 = {a,c,e}
	// SDIFF key1 key2 key3 = {b,d}
	// Keys that do not exist are considered to be empty sets.
	// Return:
	// 	Array reply: list with members of the resulting set.
	SDiff(ctx context.Context, keys ...string) StringSliceCmd

	// SInter
	// Available since: 1.0.0
	// Time complexity: O(N*M) worst case where N is the cardinality of the smallest set and M is the number of sets.
	// ACL categories: @read @set @slow
	// Returns the members of the set resulting from the intersection of all the given sets.
	// For example:
	// key1 = {a,b,c,d}
	// key2 = {c}
	// key3 = {a,c,e}
	// SINTER key1 key2 key3 = {c}
	// Keys that do not exist are considered to be empty sets. With one of the keys being an empty set, the resulting set is also empty (since set intersection with an empty set always results in an empty set).
	// Return:
	//	Array reply: list with members of the resulting set.
	SInter(ctx context.Context, keys ...string) StringSliceCmd

	// SRandMember
	// Available since: 1.0.0
	// Time complexity: Without the count argument O(1), otherwise O(N) where N is the absolute value of the passed count.
	// ACL categories: @read @set @slow
	// When called with just the key argument, return a random element from the set value stored at key.
	// If the provided count argument is positive, return an array of distinct elements. The array's length is either count or the set's cardinality (SCARD), whichever is lower.
	// If called with a negative count, the behavior changes and the command is allowed to return the same element multiple times. In this case, the number of returned elements is the absolute value of the specified count.
	// Return:
	// 	Bulk string reply: without the additional count argument, the command returns a Bulk Reply with the randomly selected element, or nil when key does not exist.
	// 	Array reply: when the additional count argument is passed, the command returns an array of elements, or an empty array when key does not exist.
	SRandMember(ctx context.Context, key string) StringCmd

	// SRandMemberN
	// Available since: 2.6.0
	// Time complexity: Without the count argument O(1), otherwise O(N) where N is the absolute value of the passed count.
	// ACL categories: @read @set @slow
	// see SRandMember
	SRandMemberN(ctx context.Context, key string, count int64) StringSliceCmd

	// SScan
	// Available since: 2.8.0
	// Time complexity: O(1) for every call. O(N) for a complete iteration, including enough command calls for the cursor to return back to 0. N is the number of elements inside the collection..
	// ACL categories: @read @set @slow
	// See https://redis.io/commands/scan/
	SScan(ctx context.Context, key string, cursor uint64, match string, count int64) ScanCmd

	// SUnion
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the total number of elements in all given sets.
	// ACL categories: @read @set @slow
	// Returns the members of the set resulting from the union of all the given sets.
	// For example:
	// key1 = {a,b,c,d}
	// key2 = {c}
	// key3 = {a,c,e}
	// SUNION key1 key2 key3 = {a,b,c,d,e}
	// Keys that do not exist are considered to be empty sets.
	// Return:
	//	Array reply: list with members of the resulting set.
	SUnion(ctx context.Context, keys ...string) StringSliceCmd
}

type SetCacheCmdable interface {
	// SCard
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @read @set @fast
	// Returns the set cardinality (number of elements) of the set stored at key.
	// Return:
	// 	Integer reply: the cardinality (number of elements) of the set, or 0 if key does not exist.
	SCard(ctx context.Context, key string) IntCmd

	// SIsMember
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @read @set @fast
	// Returns if member is a member of the set stored at key.
	// Return:
	// Integer reply, specifically:
	//	1 if the element is a member of the set.
	//	0 if the element is not a member of the set, or if key does not exist.
	SIsMember(ctx context.Context, key string, member interface{}) BoolCmd

	// SMIsMember
	// Available since: 6.2.0
	// Time complexity: O(N) where N is the number of elements being checked for membership
	// ACL categories: @read @set @fast
	// Returns whether each member is a member of the set stored at key.
	// For every member, 1 is returned if the value is a member of the set, or 0 if the element is not a member of the set or if key does not exist.
	// Return:
	// 	Array reply: list representing the membership of the given elements, in the same order as they are requested.
	SMIsMember(ctx context.Context, key string, members ...interface{}) BoolSliceCmd

	// SMembers
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the set cardinality.
	// ACL categories: @read @set @slow
	// Returns all the members of the set value stored at key.
	// This has the same effect as running SINTER with one argument key.
	// Return:
	// 	Array reply: all elements of the set.
	SMembers(ctx context.Context, key string) StringSliceCmd

	// SMembersMap
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the set cardinality.
	// ACL categories: @read @set @slow
	// Returns all the members of the set value stored at key.
	// This has the same effect as running SINTER with one argument key.
	// Return:
	// 	Array reply: all elements of the set.
	SMembersMap(ctx context.Context, key string) StringStructMapCmd
}

func (c *client) SAdd(ctx context.Context, key string, members ...interface{}) IntCmd {
	if len(members) > 1 {
		ctx = c.handler.before(ctx, CommandSAddMultiple)
	} else {
		ctx = c.handler.before(ctx, CommandSAdd)
	}
	cmd := c.cmd.B().Sadd().Key(key).Member()
	for _, m := range argsToSlice(members) {
		cmd = cmd.Member(str(m))
	}
	r := newIntCmdFromResult(c.cmd.Do(ctx, cmd.Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SCard(ctx context.Context, key string) IntCmd {
	ctx = c.handler.before(ctx, CommandSCard)
	r := newIntCmdFromResult(c.Do(ctx, c.cmd.B().Scard().Key(key).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SDiff(ctx context.Context, keys ...string) StringSliceCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandSDiff, func() []string { return keys })
	r := newStringSliceCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Sdiff().Key(keys...).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SDiffStore(ctx context.Context, destination string, keys ...string) IntCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandSDiffStore, func() []string { return appendString(destination, keys...) })
	r := newIntCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Sdiffstore().Destination(destination).Key(keys...).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SInter(ctx context.Context, keys ...string) StringSliceCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandSInter, func() []string { return keys })
	r := newStringSliceCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Sinter().Key(keys...).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SInterStore(ctx context.Context, destination string, keys ...string) IntCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandSInterStore, func() []string { return appendString(destination, keys...) })
	r := newIntCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Sinterstore().Destination(destination).Key(keys...).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SIsMember(ctx context.Context, key string, member interface{}) BoolCmd {
	ctx = c.handler.before(ctx, CommandSIsMember)
	r := newBoolCmdFromResult(c.Do(ctx, c.cmd.B().Sismember().Key(key).Member(str(member)).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SMIsMember(ctx context.Context, key string, members ...interface{}) BoolSliceCmd {
	ctx = c.handler.before(ctx, CommandSMIsMember)
	r := newBoolSliceCmdFromResult(c.Do(ctx, c.cmd.B().Smismember().Key(key).Member(argsToSlice(members)...).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SMembers(ctx context.Context, key string) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandSMembers)
	r := newStringSliceCmdFromResult(c.Do(ctx, c.cmd.B().Smembers().Key(key).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SMembersMap(ctx context.Context, key string) StringStructMapCmd {
	ctx = c.handler.before(ctx, CommandSMembers)
	r := newStringStructMapCmdFromResult(c.Do(ctx, c.cmd.B().Smembers().Key(key).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SMove(ctx context.Context, source, destination string, member interface{}) BoolCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandSMove, func() []string { return appendString(source, destination) })
	r := newBoolCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Smove().Source(source).Destination(destination).Member(str(member)).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SPop(ctx context.Context, key string) StringCmd {
	ctx = c.handler.before(ctx, CommandSPop)
	r := newStringCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Spop().Key(key).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SPopN(ctx context.Context, key string, count int64) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandSPopN)
	r := newStringSliceCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Spop().Key(key).Count(count).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SRandMember(ctx context.Context, key string) StringCmd {
	ctx = c.handler.before(ctx, CommandSRandMember)
	r := newStringCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Srandmember().Key(key).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SRandMemberN(ctx context.Context, key string, count int64) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandSRandMemberN)
	r := newStringSliceCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Srandmember().Key(key).Count(count).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SRem(ctx context.Context, key string, members ...interface{}) IntCmd {
	if len(members) > 1 {
		ctx = c.handler.before(ctx, CommandSRemMultiple)
	} else {
		ctx = c.handler.before(ctx, CommandSRem)
	}
	r := newIntCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Srem().Key(key).Member(argsToSlice(members)...).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SScan(ctx context.Context, key string, cursor uint64, match string, count int64) ScanCmd {
	ctx = c.handler.before(ctx, CommandSScan)
	cmd := c.cmd.B().Arbitrary(SSCAN).Keys(key).Args(str(int64(cursor)))
	if match != "" {
		cmd = cmd.Args(MATCH, match)
	}
	if count > 0 {
		cmd = cmd.Args(COUNT, str(count))
	}
	r := newScanCmdFromResult(c.cmd.Do(ctx, cmd.ReadOnly()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SUnion(ctx context.Context, keys ...string) StringSliceCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandSUnion, func() []string { return keys })
	r := newStringSliceCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Sunion().Key(keys...).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SUnionStore(ctx context.Context, destination string, keys ...string) IntCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandSUnionStore, func() []string { return appendString(destination, keys...) })
	r := newIntCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Sunionstore().Destination(destination).Key(keys...).Build()))
	c.handler.after(ctx, r.Err())
	return r
}
