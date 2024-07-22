package redisson

import (
	"context"
	"strings"
)

type BitmapCmdable interface {
	BitmapWriter
	BitmapReader
}

type BitmapWriter interface {
	// BitField
	// Available since: 3.2.0
	// Time complexity: O(1) for each subcommand specified
	// ACL categories: @write @bitmap @slow
	// RESP2 Reply:
	//	One of the following:
	//		- Array reply: each entry being the corresponding result of the sub-command given at the same position.
	//		- Nil reply: if OVERFLOW FAIL was given and overflows or underflows are detected.
	// RESP3 Reply:
	//	One of the following:
	//		- Array reply: each entry being the corresponding result of the sub-command given at the same position.
	//		- Null reply: if OVERFLOW FAIL was given and overflows or underflows are detected.
	BitField(ctx context.Context, key string, args ...any) IntSliceCmd

	// BitOpAnd
	// Available since: 2.6.0
	// Time complexity: O(N)
	// ACL categories: @write @bitmap @slow
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the size of the string stored in the destination key is equal to the size of the longest input string.
	BitOpAnd(ctx context.Context, destKey string, keys ...string) IntCmd
	BitOpOr(ctx context.Context, destKey string, keys ...string) IntCmd
	BitOpXor(ctx context.Context, destKey string, keys ...string) IntCmd
	BitOpNot(ctx context.Context, destKey string, key string) IntCmd

	// SetBit
	// Available since: 2.2.0
	// Time complexity: O(1)
	// ACL categories: @write @bitmap @slow
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the original bit value stored at offset.
	SetBit(ctx context.Context, key string, offset int64, value int64) IntCmd
}

type BitmapReader any

type BitmapCacheCmdable interface {
	// BitCount
	// Available since: 2.6.0
	// Time complexity: O(N)
	// ACL categories: @read @bitmap @slow
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the number of bits set to 1.
	// History:
	//	- Starting with Redis version 7.0.0: Added the BYTE|BIT option.
	BitCount(ctx context.Context, key string, bitCount *BitCount) IntCmd

	// BitPos
	// Available since: 2.8.7
	// Time complexity: O(N)
	// ACL categories: @read @bitmap @slow
	// RESP2 / RESP3 Reply:
	//	One of the following:
	// 		- Integer reply: the position of the first bit set to 1 or 0 according to the request
	//		- Integer reply: -1. In case the bit argument is 1 and the string is empty or composed of just zero bytes
	// History:
	//	- Starting with Redis version 7.0.0: Added the BYTE|BIT option.
	BitPos(ctx context.Context, key string, bit int64, pos ...int64) IntCmd
	BitPosSpan(ctx context.Context, key string, bit, start, end int64, span string) IntCmd

	// GetBit
	// Available since: 2.2.0
	// Time complexity: O(1)
	// ACL categories: @read @bitmap @fast
	// RESP2 / RESP3 Reply:
	//	The bit value stored at offset, one of the following:
	// 		- Integer reply: 0.
	//		- Integer reply: 1.
	GetBit(ctx context.Context, key string, offset int64) IntCmd
}

func (c *client) BitCount(ctx context.Context, key string, bc *BitCount) IntCmd {
	if bc == nil || bc.Unit == "" {
		ctx = c.handler.before(ctx, CommandBitCount)
	} else {
		switch strings.ToUpper(bc.Unit) {
		case BitCountIndexByte:
			ctx = c.handler.before(ctx, CommandBitCountByte)
		case BitCountIndexBit:
			ctx = c.handler.before(ctx, CommandBitCountBit)
		default:
			ctx = c.handler.before(ctx, CommandBitCount)
		}
	}
	var r IntCmd
	if c.ttl > 0 {
		r = newIntCmd(c.Do(ctx, c.builder.BitCountCompleted(key, bc)))
	} else {
		r = c.adapter.BitCount(ctx, key, bc)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) BitField(ctx context.Context, key string, args ...any) IntSliceCmd {
	ctx = c.handler.before(ctx, CommandBitField)
	r := c.adapter.BitField(ctx, key, args...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) BitOpAnd(ctx context.Context, destKey string, keys ...string) IntCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandBitOpAnd, func() []string { return appendString(destKey, keys...) })
	r := c.adapter.BitOpAnd(ctx, destKey, keys...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) BitOpOr(ctx context.Context, destKey string, keys ...string) IntCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandBitOpOr, func() []string { return appendString(destKey, keys...) })
	r := c.adapter.BitOpOr(ctx, destKey, keys...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) BitOpXor(ctx context.Context, destKey string, keys ...string) IntCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandBitOpXor, func() []string { return appendString(destKey, keys...) })
	r := c.adapter.BitOpXor(ctx, destKey, keys...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) BitOpNot(ctx context.Context, destKey string, key string) IntCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandBitOpNot, func() []string { return appendString(destKey, key) })
	r := c.adapter.BitOpNot(ctx, destKey, key)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) BitPos(ctx context.Context, key string, bit int64, pos ...int64) IntCmd {
	ctx = c.handler.before(ctx, CommandBitPos)
	var r IntCmd
	if c.ttl > 0 {
		r = newIntCmd(c.Do(ctx, c.builder.BitPosCompleted(key, bit, pos...)))
	} else {
		r = c.adapter.BitPos(ctx, key, bit, pos...)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) BitPosSpan(ctx context.Context, key string, bit, start, end int64, span string) IntCmd {
	ctx = c.handler.before(ctx, CommandBitPos)
	var r IntCmd
	if c.ttl > 0 {
		r = newIntCmd(c.Do(ctx, c.builder.BitPosSpanCompleted(key, bit, start, end, span)))
	} else {
		r = c.adapter.BitPosSpan(ctx, key, bit, start, end, span)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) GetBit(ctx context.Context, key string, offset int64) IntCmd {
	ctx = c.handler.before(ctx, CommandGetBit)
	var r IntCmd
	if c.ttl > 0 {
		r = newIntCmd(c.Do(ctx, c.builder.GetBitCompleted(key, offset)))
	} else {
		r = c.adapter.GetBit(ctx, key, offset)
	}
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SetBit(ctx context.Context, key string, offset int64, value int64) IntCmd {
	ctx = c.handler.before(ctx, CommandSetBit)
	r := c.adapter.SetBit(ctx, key, offset, value)
	c.handler.after(ctx, r.Err())
	return r
}
