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
	//
	// The command treats a Redis string as an array of bits, and is capable of addressing specific integer fields of
	// varying bit widths and arbitrary non (necessary) aligned offset. In practical terms using this command you can set,
	// for example, a signed 5 bits integer at bit offset 1234 to a specific value, retrieve a 31 bit unsigned integer from offset 4567.
	// Similarly, the command handles increments and decrements of the specified integers, providing guaranteed and well specified
	// overflow and underflow behavior that the user can configure.
	//
	// BITFIELD is able to operate with multiple bit fields in the same command call. It takes a list of operations to perform,
	// and returns an array of replies, where each array matches the corresponding operation in the list of arguments.
	//
	// Performance considerations
	//	Usually BITFIELD is a fast command, however note that addressing far bits of currently short strings will trigger an allocation that
	// 	may be more costly than executing the command on bits already existing.
	//
	// RESP2 Reply
	// 	One of the following:
	//		Array reply: each entry being the corresponding result of the sub-command given at the same position.
	//		Nil reply: if OVERFLOW FAIL was given and overflows or underflows are detected.
	//
	// RESP3 Reply
	// 	One of the following:
	// 		Array reply: each entry being the corresponding result of the sub-command given at the same position.
	//		Null reply: if OVERFLOW FAIL was given and overflows or underflows are detected.
	BitField(ctx context.Context, key string, args ...any) IntSliceCmd

	// BitOpAnd
	// Available since: 2.6.0
	// Time complexity: O(N)
	// ACL categories: @write @bitmap @slow
	//
	// Perform a bitwise operation between multiple keys (containing string values) and store the result in the destination key.
	//
	// The BITOP command supports four bitwise operations: AND, OR, XOR and NOT, thus the valid forms to call the command are:
	//	BITOP AND destkey srckey1 srckey2 srckey3 ... srckeyN
	//	BITOP OR destkey srckey1 srckey2 srckey3 ... srckeyN
	//	BITOP XOR destkey srckey1 srckey2 srckey3 ... srckeyN
	//	BITOP NOT destkey srckey
	//
	// As you can see NOT is special as it only takes an input key, because it performs inversion of bits so it only makes sense as a unary operator.
	//
	// The result of the operation is always stored at destkey.
	//
	// Performance considerations
	// 	BITOP is a potentially slow command as it runs in O(N) time. Care should be taken when running it against long input strings.
	// 	For real-time metrics and statistics involving large inputs a good approach is to use a replica (with replica-read-only option enabled)
	//	where the bit-wise operations are performed to avoid blocking the master instance.
	//
	// RESP2/RESP3 Reply
	// 	Integer reply: the size of the string stored in the destination key is equal to the size of the longest input string.
	BitOpAnd(ctx context.Context, destKey string, keys ...string) IntCmd
	BitOpOr(ctx context.Context, destKey string, keys ...string) IntCmd
	BitOpXor(ctx context.Context, destKey string, keys ...string) IntCmd
	BitOpNot(ctx context.Context, destKey string, key string) IntCmd

	// SetBit
	// Available since: 2.2.0
	// Time complexity: O(1)
	// ACL categories: @write @bitmap @slow
	//
	// Sets or clears the bit at offset in the string value stored at key.
	//
	// The bit is either set or cleared depending on value, which can be either 0 or 1.
	//
	// When key does not exist, a new string value is created. The string is grown to make sure it can hold a bit at offset.
	// The offset argument is required to be greater than or equal to 0, and smaller than 2^32 (this limits bitmaps to 512MB).
	// When the string at key is grown, added bits are set to 0.
	//
	// Warning: When setting the last possible bit (offset equal to 2^32 -1) and the string value stored at key does not
	// yet hold a string value, or holds a small string value, Redis needs to allocate all intermediate memory which can block
	// the server for some time. On a 2010 MacBook Pro, setting bit number 2^32 -1 (512MB allocation) takes ~300ms,
	// setting bit number 2^30 -1 (128MB allocation) takes ~80ms, setting bit number 2^28 -1 (32MB allocation) takes ~30ms and
	// setting bit number 2^26 -1 (8MB allocation) takes ~8ms. Note that once this first allocation is done, subsequent calls
	// to SETBIT for the same key will not have the allocation overhead.
	//
	// RESP2/RESP3 Reply
	// 	Integer reply: the original bit value stored at offset.
	SetBit(ctx context.Context, key string, offset int64, value int64) IntCmd
}

type BitmapReader any

type BitmapCacheCmdable interface {
	// BitCount
	// Available since: 2.6.0
	// Time complexity: O(N)
	// ACL categories: @read @bitmap @slow
	//
	// Count the number of set bits (population counting) in a string.
	//
	// By default, all the bytes contained in the string are examined.
	// It is possible to specify the counting operation only in an interval passing the additional arguments start and end.
	//
	// Like for the GETRANGE command start and end can contain negative values in order to index bytes starting from the end of the string,
	// where -1 is the last byte, -2 is the penultimate, and so forth.
	//
	// Non-existent keys are treated as empty strings, so the command will return zero.
	//
	// By default, the additional arguments start and end specify a byte index.
	// We can use an additional argument BIT to specify a bit index. So 0 is the first bit, 1 is the second bit, and so forth.
	// For negative values, -1 is the last bit, -2 is the penultimate, and so forth.
	//
	// RESP2/RESP3 Reply
	// 	Integer reply: the number of bits set to 1.
	//
	// History
	// 	Starting with Redis version 7.0.0: Added the BYTE|BIT option.
	BitCount(ctx context.Context, key string, bitCount *BitCount) IntCmd

	// BitPos
	// Available since: 2.8.7
	// Time complexity: O(N)
	// ACL categories: @read @bitmap @slow
	//
	// Return the position of the first bit set to 1 or 0 in a string.
	//
	// The position is returned, thinking of the string as an array of bits from left to right,
	// where the first byte's most significant bit is at position 0, the second byte's most significant bit is at position 8, and so forth.
	//
	// The same bit position convention is followed by GETBIT and SETBIT.
	//
	// By default, all the bytes contained in the string are examined. It is possible to look for bits only in a specified interval passing
	// the additional arguments start and end (it is possible to just pass start, the operation will assume that the end is the last byte of the string.
	// However there are semantic differences as explained later). By default, the range is interpreted as a range of bytes and not a range of bits,
	// so start=0 and end=2 means to look at the first three bytes.
	//
	// You can use the optional BIT modifier to specify that the range should be interpreted as a range of bits. So start=0 and end=2 means to look at the first three bits.
	//
	// Note that bit positions are returned always as absolute values starting from bit zero even when start and end are used to specify a range.
	// Like for the GETRANGE command start and end can contain negative values in order to index bytes starting from the end of the string, where -1 is the last byte,
	// -2 is the penultimate, and so forth. When BIT is specified, -1 is the last bit, -2 is the penultimate, and so forth.
	//
	// Non-existent keys are treated as empty strings.
	//
	// RESP2/RESP3 Reply
	// 	One of the following:
	//		Integer reply: the position of the first bit set to 1 or 0 according to the request
	//		Integer reply: -1. In case the bit argument is 1 and the string is empty or composed of just zero bytes
	//
	// History
	//	Starting with Redis version 7.0.0: Added the BYTE|BIT option.
	BitPos(ctx context.Context, key string, bit int64, pos ...int64) IntCmd
	BitPosSpan(ctx context.Context, key string, bit, start, end int64, span string) IntCmd

	// GetBit
	// Available since: 2.2.0
	// Time complexity: O(1)
	// ACL categories: @read @bitmap @fast
	// Returns the bit value at offset in the string value stored at key.
	//
	// When offset is beyond the string length, the string is assumed to be a contiguous space with 0 bits.
	// When key does not exist it is assumed to be an empty string, so offset is always out of range and the
	// value is also assumed to be a contiguous space with 0 bits.
	//
	// RESP2/RESP3 Reply
	// 	The bit value stored at offset, one of the following:
	//		Integer reply: 0.
	//		Integer reply: 1.
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
