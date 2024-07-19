package redisson

import (
	"context"
)

type HyperLogCmdable interface {
	HyperLogWriter
	HyperLogReader
}

type HyperLogWriter interface {
	// PFAdd
	// Available since: 2.8.9
	// Time complexity: O(1) to add every element.
	// ACL categories: @write @hyperloglog @fast
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- Integer reply: 1 if at least one HyperLogLog internal register was altered.
	//		- Integer reply: 0 if no HyperLogLog internal registers were altered.
	PFAdd(ctx context.Context, key string, els ...any) IntCmd

	// PFMerge
	// Available since: 2.8.9
	// Time complexity: O(N) to merge N HyperLogLogs, but with high constant times.
	// ACL categories: @write @hyperloglog @slow
	// RESP2 / RESP3 Reply:
	// 	- Simple string reply: OK.
	PFMerge(ctx context.Context, dest string, keys ...string) StatusCmd
}

type HyperLogReader interface {
	// PFCount
	// Available since: 2.8.9
	// Time complexity: O(1) with a very small average constant time when called with a single key.
	//	O(N) with N being the number of keys, and much bigger constant times, when called with multiple keys.
	// ACL categories: @read @hyperloglog @slow
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the approximated number of unique elements observed via PFADD
	PFCount(ctx context.Context, keys ...string) IntCmd
}

func (c *client) PFAdd(ctx context.Context, key string, els ...any) IntCmd {
	ctx = c.handler.before(ctx, CommandPFAdd)
	r := c.adapter.PFAdd(ctx, key, els...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) PFCount(ctx context.Context, keys ...string) IntCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandPFCount, func() []string { return keys })
	r := c.adapter.PFCount(ctx, keys...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) PFMerge(ctx context.Context, dest string, keys ...string) StatusCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandPFMerge, func() []string { return appendString(dest, keys...) })
	r := c.adapter.PFMerge(ctx, dest, keys...)
	c.handler.after(ctx, r.Err())
	return r
}
