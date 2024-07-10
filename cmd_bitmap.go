package redisson

import (
	"context"
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
	// See https://redis.io/commands/bitfield/
	BitField(ctx context.Context, key string, args ...interface{}) IntSliceCmd

	// BitOpAnd
	// Available since: 2.6.0
	// Time complexity: O(N)
	// ACL categories: @write @bitmap @slow
	// Perform a bitwise operation between multiple keys (containing string values) and store the result in the destination key.
	// The BITOP command supports four bitwise operations: AND, OR, XOR and NOT, thus the valid forms to call the command are:
	//	BITOP AND destkey srckey1 srckey2 srckey3 ... srckeyN
	//	BITOP OR destkey srckey1 srckey2 srckey3 ... srckeyN
	//	BITOP XOR destkey srckey1 srckey2 srckey3 ... srckeyN
	//	BITOP NOT destkey srckey
	// As you can see NOT is special as it only takes an input key, because it performs inversion of bits so it only makes sense as a unary operator.
	// The result of the operation is always stored at destkey.
	// Handling of strings with different lengths
	// When an operation is performed between strings having different lengths, all the strings shorter than the longest string in the set are treated as if they were zero-padded up to the length of the longest string.
	// The same holds true for non-existent keys, that are considered as a stream of zero bytes up to the length of the longest string.
	// Return:
	// Integer reply
	// The size of the string stored in the destination key, that is equal to the size of the longest input string.
	BitOpAnd(ctx context.Context, destKey string, keys ...string) IntCmd

	// BitOpOr
	// Available since: 2.6.0
	// Time complexity: O(N)
	// ACL categories: @write @bitmap @slow
	BitOpOr(ctx context.Context, destKey string, keys ...string) IntCmd

	// BitOpXor
	// Available since: 2.6.0
	// Time complexity: O(N)
	// ACL categories: @write @bitmap @slow
	BitOpXor(ctx context.Context, destKey string, keys ...string) IntCmd

	// BitOpNot
	// Available since: 2.6.0
	// Time complexity: O(N)
	// ACL categories: @write @bitmap @slow
	BitOpNot(ctx context.Context, destKey string, key string) IntCmd

	// SetBit
	// Available since: 2.2.0
	// Time complexity: O(1)
	// ACL categories: @write @bitmap @slow
	// Sets or clears the bit at offset in the string value stored at key.
	// The bit is either set or cleared depending on value, which can be either 0 or 1.
	// When key does not exist, a new string value is created. The string is grown to make sure it can hold a bit at offset. The offset argument is required to be greater than or equal to 0, and smaller than 2^32 (this limits bitmaps to 512MB). When the string at key is grown, added bits are set to 0.
	// Warning: When setting the last possible bit (offset equal to 2^32 -1) and the string value stored at key does not yet hold a string value, or holds a small string value, Redis needs to allocate all intermediate memory which can block the server for some time. On a 2010 MacBook Pro, setting bit number 2^32 -1 (512MB allocation) takes ~300ms, setting bit number 2^30 -1 (128MB allocation) takes ~80ms, setting bit number 2^28 -1 (32MB allocation) takes ~30ms and setting bit number 2^26 -1 (8MB allocation) takes ~8ms. Note that once this first allocation is done, subsequent calls to SETBIT for the same key will not have the allocation overhead.
	// Return:
	// Integer reply: the original bit value stored at offset.
	SetBit(ctx context.Context, key string, offset int64, value int) IntCmd
}

type BitmapReader interface{}

type BitmapCacheCmdable interface {
	// BitCount
	// Available since: 2.6.0
	// Time complexity: O(N)
	// ACL categories: @read @bitmap @slow
	// Count the number of set bits (population counting) in a string.
	// By default all the bytes contained in the string are examined. It is possible to specify the counting operation only in an interval passing the additional arguments start and end.
	// Like for the GETRANGE command start and end can contain negative values in order to index bytes starting from the end of the string, where -1 is the last byte, -2 is the penultimate, and so forth.
	// Non-existent keys are treated as empty strings, so the command will return zero.
	// By default, the additional arguments start and end specify a byte index. We can use an additional argument BIT to specify a bit index. So 0 is the first bit, 1 is the second bit, and so forth. For negative values, -1 is the last bit, -2 is the penultimate, and so forth.
	// Return:
	// Integer reply
	//	The number of bits set to 1.
	BitCount(ctx context.Context, key string, bitCount *BitCount) IntCmd

	// BitPos
	// Available since: 2.8.7
	// Time complexity: O(N)
	// ACL categories: @read @bitmap @slow
	// Return the position of the first bit set to 1 or 0 in a string.
	// The position is returned, thinking of the string as an array of bits from left to right, where the first byte's most significant bit is at position 0, the second byte's most significant bit is at position 8, and so forth.
	// The same bit position convention is followed by GETBIT and SETBIT.
	// By default, all the bytes contained in the string are examined. It is possible to look for bits only in a specified interval passing the additional arguments start and end (it is possible to just pass start, the operation will assume that the end is the last byte of the string. However there are semantic differences as explained later). By default, the range is interpreted as a range of bytes and not a range of bits, so start=0 and end=2 means to look at the first three bytes.
	// You can use the optional BIT modifier to specify that the range should be interpreted as a range of bits. So start=0 and end=2 means to look at the first three bits.
	// Note that bit positions are returned always as absolute values starting from bit zero even when start and end are used to specify a range.
	// Like for the GETRANGE command start and end can contain negative values in order to index bytes starting from the end of the string, where -1 is the last byte, -2 is the penultimate, and so forth. When BIT is specified, -1 is the last bit, -2 is the penultimate, and so forth.
	// Non-existent keys are treated as empty strings.
	// Return:
	// Integer reply
	// The command returns the position of the first bit set to 1 or 0 according to the request.
	// If we look for set bits (the bit argument is 1) and the string is empty or composed of just zero bytes, -1 is returned.
	// If we look for clear bits (the bit argument is 0) and the string only contains bit set to 1, the function returns the first bit not part of the string on the right. So if the string is three bytes set to the value 0xff the command BITPOS key 0 will return 24, since up to bit 23 all the bits are 1.
	// Basically, the function considers the right of the string as padded with zeros if you look for clear bits and specify no range or the start argument only.
	// However, this behavior changes if you are looking for clear bits and specify a range with both start and end. If no clear bit is found in the specified range, the function returns -1 as the user specified a clear range and there are no 0 bits in that range.
	BitPos(ctx context.Context, key string, bit int64, pos ...int64) IntCmd

	// GetBit
	// Available since: 2.2.0
	// Time complexity: O(1)
	// ACL categories: @read @bitmap @fast
	// Returns the bit value at offset in the string value stored at key.
	// When offset is beyond the string length, the string is assumed to be a contiguous space with 0 bits. When key does not exist it is assumed to be an empty string, so offset is always out of range and the value is also assumed to be a contiguous space with 0 bits.
	// Return:
	// Integer reply: the bit value stored at offset.
	GetBit(ctx context.Context, key string, offset int64) IntCmd
}

func (c *client) BitCount(ctx context.Context, key string, bc *BitCount) IntCmd {
	ctx = c.handler.before(ctx, CommandBitCount)
	r := newIntCmdFromResult(c.Do(ctx, c.getBitCountCompleted(key, bc)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) BitField(ctx context.Context, key string, args ...interface{}) IntSliceCmd {
	ctx = c.handler.before(ctx, CommandBitField)
	r := newIntSliceCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Arbitrary(BITFIELD).Keys(key).Args(argsToSlice(args)...).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) BitOpAnd(ctx context.Context, destKey string, keys ...string) IntCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandBitOpAnd, func() []string { return appendString(destKey, keys...) })
	r := newIntCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Bitop().And().Destkey(destKey).Key(keys...).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) BitOpOr(ctx context.Context, destKey string, keys ...string) IntCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandBitOpOr, func() []string { return appendString(destKey, keys...) })
	r := newIntCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Bitop().Or().Destkey(destKey).Key(keys...).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) BitOpXor(ctx context.Context, destKey string, keys ...string) IntCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandBitOpXor, func() []string { return appendString(destKey, keys...) })
	r := newIntCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Bitop().Xor().Destkey(destKey).Key(keys...).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) BitOpNot(ctx context.Context, destKey string, key string) IntCmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandBitOpNot, func() []string { return appendString(destKey, key) })
	r := newIntCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Bitop().Not().Destkey(destKey).Key(key).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) BitPos(ctx context.Context, key string, bit int64, pos ...int64) IntCmd {
	ctx = c.handler.before(ctx, CommandBitPos)
	r := newIntCmdFromResult(c.Do(ctx, c.getBitPosCompleted(key, bit, pos...)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) GetBit(ctx context.Context, key string, offset int64) IntCmd {
	ctx = c.handler.before(ctx, CommandGetBit)
	r := newIntCmdFromResult(c.Do(ctx, c.getBitCompleted(key, offset)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SetBit(ctx context.Context, key string, offset int64, value int) IntCmd {
	ctx = c.handler.before(ctx, CommandSetBit)
	r := newIntCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Setbit().Key(key).Offset(offset).Value(int64(value)).Build()))
	c.handler.after(ctx, r.Err())
	return r
}
