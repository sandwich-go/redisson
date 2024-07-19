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
	SAdd(ctx context.Context, key string, members ...any) IntCmd

	// SDiffStore
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the total number of elements in all given sets.
	// ACL categories: @write @set @slow
	SDiffStore(ctx context.Context, destination string, keys ...string) IntCmd

	// SInterStore
	// Available since: 1.0.0
	// Time complexity: O(N*M) worst case where N is the cardinality of the smallest set and M is the number of sets.
	// ACL categories: @write @set @slow
	SInterStore(ctx context.Context, destination string, keys ...string) IntCmd

	// SMove
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @write @set @fast
	SMove(ctx context.Context, source, destination string, member any) BoolCmd

	// SPop
	// Available since: 1.0.0
	// Time complexity: Without the count argument O(1), otherwise O(N) where N is the value of the passed count.
	// ACL categories: @write @set @fast
	SPop(ctx context.Context, key string) StringCmd

	// SPopN
	// Available since: 3.2.0
	// Time complexity: Without the count argument O(1), otherwise O(N) where N is the value of the passed count.
	// ACL categories: @write @set @fast
	SPopN(ctx context.Context, key string, count int64) StringSliceCmd

	// SRem
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the number of members to be removed.
	// ACL categories: @write @set @fast
	SRem(ctx context.Context, key string, members ...any) IntCmd

	// SUnionStore
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the total number of elements in all given sets.
	// ACL categories: @write @set @slow
	SUnionStore(ctx context.Context, destination string, keys ...string) IntCmd
}

type SetReader interface {
	// SDiff
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the total number of elements in all given sets.
	// ACL categories: @read @set @slow
	SDiff(ctx context.Context, keys ...string) StringSliceCmd

	// SInter
	// Available since: 1.0.0
	// Time complexity: O(N*M) worst case where N is the cardinality of the smallest set and M is the number of sets.
	// ACL categories: @read @set @slow
	SInter(ctx context.Context, keys ...string) StringSliceCmd

	// SInterCard
	// Available since: 7.0.0
	// Time complexity: O(N*M) worst case where N is the cardinality of the smallest set and M is the number of sets.
	// ACL categories: @read @set @slow
	SInterCard(ctx context.Context, limit int64, keys ...string) IntCmd

	// SRandMember
	// Available since: 1.0.0
	// Time complexity: Without the count argument O(1), otherwise O(N) where N is the absolute value of the passed count.
	// ACL categories: @read @set @slow
	SRandMember(ctx context.Context, key string) StringCmd

	// SRandMemberN
	// Available since: 2.6.0
	// Time complexity: Without the count argument O(1), otherwise O(N) where N is the absolute value of the passed count.
	// ACL categories: @read @set @slow
	SRandMemberN(ctx context.Context, key string, count int64) StringSliceCmd

	// SScan
	// Available since: 2.8.0
	// Time complexity: O(1) for every call. O(N) for a complete iteration, including enough command calls for the cursor to return back to 0. N is the number of elements inside the collection..
	// ACL categories: @read @set @slow
	SScan(ctx context.Context, key string, cursor uint64, match string, count int64) ScanCmd

	// SUnion
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the total number of elements in all given sets.
	// ACL categories: @read @set @slow
	SUnion(ctx context.Context, keys ...string) StringSliceCmd
}

type SetCacheCmdable interface {
	// SCard
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @read @set @fast
	SCard(ctx context.Context, key string) IntCmd

	// SIsMember
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @read @set @fast
	SIsMember(ctx context.Context, key string, member any) BoolCmd

	// SMIsMember
	// Available since: 6.2.0
	// Time complexity: O(N) where N is the number of elements being checked for membership
	// ACL categories: @read @set @fast
	SMIsMember(ctx context.Context, key string, members ...any) BoolSliceCmd

	// SMembers
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the set cardinality.
	// ACL categories: @read @set @slow
	SMembers(ctx context.Context, key string) StringSliceCmd
}

func (c *client) SAdd(ctx context.Context, key string, members ...any) IntCmd {
	if len(members) > 1 {
		ctx = c.handler.before(ctx, CommandSMAdd)
	} else {
		ctx = c.handler.before(ctx, CommandSAdd)
	}
	r := c.adapter.SAdd(ctx, key, members...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SCard(ctx context.Context, key string) IntCmd {
	ctx = c.handler.before(ctx, CommandSCard)
	var r IntCmd
	if c.ttl > 0 {
		r = newIntCmd(c.Do(ctx, c.builder.SCardCompleted(key)))
	} else {
		r = c.adapter.SCard(ctx, key)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SDiff(ctx context.Context, keys ...string) StringSliceCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandSDiff, func() []string { return keys })
	r := c.adapter.SDiff(ctx, keys...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SDiffStore(ctx context.Context, destination string, keys ...string) IntCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandSDiffStore, func() []string { return appendString(destination, keys...) })
	r := c.adapter.SDiffStore(ctx, destination, keys...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SInter(ctx context.Context, keys ...string) StringSliceCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandSInter, func() []string { return keys })
	r := c.adapter.SInter(ctx, keys...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SInterStore(ctx context.Context, destination string, keys ...string) IntCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandSInterStore, func() []string { return appendString(destination, keys...) })
	r := c.adapter.SInterStore(ctx, destination, keys...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SInterCard(ctx context.Context, limit int64, keys ...string) IntCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandSInterCard, func() []string { return keys })
	r := c.adapter.SInterCard(ctx, limit, keys...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SIsMember(ctx context.Context, key string, member any) BoolCmd {
	ctx = c.handler.before(ctx, CommandSIsMember)
	var r BoolCmd
	if c.ttl > 0 {
		r = newBoolCmd(c.Do(ctx, c.builder.SIsMemberCompleted(key, member)))
	} else {
		r = c.adapter.SIsMember(ctx, key, member)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SMIsMember(ctx context.Context, key string, members ...any) BoolSliceCmd {
	ctx = c.handler.before(ctx, CommandSMIsMember)
	var r BoolSliceCmd
	if c.ttl > 0 {
		r = newBoolSliceCmd(c.Do(ctx, c.builder.SMIsMemberCompleted(key, members...)))
	} else {
		r = c.adapter.SMIsMember(ctx, key, members...)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SMembers(ctx context.Context, key string) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandSMembers)
	var r StringSliceCmd
	if c.ttl > 0 {
		r = newStringSliceCmd(c.Do(ctx, c.builder.SMembersCompleted(key)))
	} else {
		r = c.adapter.SMembers(ctx, key)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SMove(ctx context.Context, source, destination string, member any) BoolCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandSMove, func() []string { return appendString(source, destination) })
	r := c.adapter.SMove(ctx, source, destination, member)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SPop(ctx context.Context, key string) StringCmd {
	ctx = c.handler.before(ctx, CommandSPop)
	r := c.adapter.SPop(ctx, key)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SPopN(ctx context.Context, key string, count int64) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandSPopN)
	r := c.adapter.SPopN(ctx, key, count)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SRandMember(ctx context.Context, key string) StringCmd {
	ctx = c.handler.before(ctx, CommandSRandMember)
	r := c.adapter.SRandMember(ctx, key)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SRandMemberN(ctx context.Context, key string, count int64) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandSRandMemberN)
	r := c.adapter.SRandMemberN(ctx, key, count)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SRem(ctx context.Context, key string, members ...any) IntCmd {
	if len(members) > 1 {
		ctx = c.handler.before(ctx, CommandSMRem)
	} else {
		ctx = c.handler.before(ctx, CommandSRem)
	}
	r := c.adapter.SRem(ctx, key, members...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SScan(ctx context.Context, key string, cursor uint64, match string, count int64) ScanCmd {
	ctx = c.handler.before(ctx, CommandSScan)
	r := c.adapter.SScan(ctx, key, cursor, match, count)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SUnion(ctx context.Context, keys ...string) StringSliceCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandSUnion, func() []string { return keys })
	r := c.adapter.SUnion(ctx, keys...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SUnionStore(ctx context.Context, destination string, keys ...string) IntCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandSUnionStore, func() []string { return appendString(destination, keys...) })
	r := c.adapter.SUnionStore(ctx, destination, keys...)
	c.handler.after(ctx, r.Err())
	return r
}
