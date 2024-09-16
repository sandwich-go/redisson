package redisson

import (
	"context"
)

type StreamCmdable interface {
	StreamWriter
	StreamReader
}

type StreamWriter interface {
	// XAck
	// Available since: 5.0.0
	// Time complexity: O(1) for each message ID processed.
	// ACL categories: @write @stream @fast
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: The command returns the number of messages successfully acknowledged.
	//		Certain message IDs may no longer be part of the PEL (for example because they have already been acknowledged),
	//		and XACK will not count them as successfully acknowledged.
	XAck(ctx context.Context, stream, group string, ids ...string) IntCmd

	// XAdd
	// Available since: 5.0.0
	// Time complexity: O(1) when adding a new entry, O(N) when trimming where N being the number of entries evicted.
	// ACL categories: @write @stream @fast
	// RESP2 Reply:
	//	One of the following:
	//		- Bulk string reply: The ID of the added entry. The ID is the one automatically generated if an asterisk (*) is passed as the id argument,
	//			otherwise the command just returns the same ID specified by the user during insertion.
	//		- Nil reply: if the NOMKSTREAM option is given and the key doesn't exist.
	// RESP3 Reply:
	//	One of the following:
	//		- Bulk string reply: The ID of the added entry. The ID is the one automatically generated if an asterisk (*) is passed as the id argument,
	//			otherwise the command just returns the same ID specified by the user during insertion.
	//		- Null reply: if the NOMKSTREAM option is given and the key doesn't exist.
	// History:
	//	- Starting with Redis version 6.2.0: Added the NOMKSTREAM option, MINID trimming strategy and the LIMIT option.
	//	- Starting with Redis version 7.0.0: Added support for the <ms>-* explicit ID form.
	XAdd(ctx context.Context, a XAddArgs) StringCmd

	// XAutoClaim
	// Available since: 6.2.0
	// Time complexity: O(1) if COUNT is small.
	// ACL categories: @write @stream @fast
	// RESP2 / RESP3 Reply:
	//	Array reply, specifically, an array with three elements:
	//		1. A stream ID to be used as the start argument for the next call to XAUTOCLAIM.
	//		2. An Array reply containing all the successfully claimed messages in the same format as XRANGE.
	//		3. An Array reply containing message IDs that no longer exist in the stream, and were deleted from the PEL in which they were found.
	// History:
	//	- Starting with Redis version 7.0.0: Added an element to the reply array, containing deleted entries the command cleared from the PEL
	XAutoClaim(ctx context.Context, a XAutoClaimArgs) XAutoClaimCmd
	XAutoClaimJustID(ctx context.Context, a XAutoClaimArgs) XAutoClaimJustIDCmd

	// XClaim
	// Available since: 5.0.0
	// Time complexity: O(log N) with N being the number of messages in the PEL of the consumer group.
	// ACL categories: @write @stream @fast
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- Array reply: when the JUSTID option is specified, an array of IDs of messages successfully claimed.
	//		- Array reply: an array of stream entries, each of which contains an array of two elements, the entry ID and the entry data itself.
	XClaim(ctx context.Context, a XClaimArgs) XMessageSliceCmd
	XClaimJustID(ctx context.Context, a XClaimArgs) StringSliceCmd

	// XDel
	// Available since: 5.0.0
	// Time complexity: O(1) for each single item to delete in the stream, regardless of the stream size.
	// ACL categories: @write @stream @fast
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the number of entries that were deleted.
	XDel(ctx context.Context, stream string, ids ...string) IntCmd

	// XGroupCreate
	// Available since: 5.0.0
	// Time complexity: O(1)
	// ACL categories: @write @stream @slow
	// RESP2 / RESP3 Reply:
	// 	- Simple string reply: OK.
	// History:
	//	- Starting with Redis version 7.0.0: Added the entries_read named argument.
	XGroupCreate(ctx context.Context, stream, group, start string) StatusCmd
	XGroupCreateMkStream(ctx context.Context, stream, group, start string) StatusCmd

	// XGroupCreateConsumer
	// Available since: 6.2.0
	// Time complexity: O(1)
	// ACL categories: @write @stream @slow
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the number of created consumers, either 0 or 1.
	XGroupCreateConsumer(ctx context.Context, stream, group, consumer string) IntCmd

	// XGroupDelConsumer
	// Available since: 5.0.0
	// Time complexity: O(1)
	// ACL categories: @write @stream @slow
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the number of pending messages the consumer had before it was deleted.
	XGroupDelConsumer(ctx context.Context, stream, group, consumer string) IntCmd

	// XGroupDestroy
	// Available since: 5.0.0
	// Time complexity: O(N) where N is the number of entries in the group's pending entries list (PEL).
	// ACL categories: @write @stream @slow
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the number of destroyed consumer groups, either 0 or 1.
	XGroupDestroy(ctx context.Context, stream, group string) IntCmd

	// XGroupSetID
	// Available since: 5.0.0
	// Time complexity: O(1)
	// ACL categories: @write @stream @slow
	// RESP2 / RESP3 Reply:
	// 	- Simple string reply: OK.
	// History:
	//	- Starting with Redis version 7.0.0: Added the entries_read named argument.
	XGroupSetID(ctx context.Context, stream, group, start string) StatusCmd

	// XReadGroup
	// Available since: 5.0.0
	// Time complexity: For each stream mentioned: O(M) with M being the number of elements returned.
	//					If M is constant (e.g. always asking for the first 10 elements with COUNT), you can consider it O(1).
	//					On the other side when XREADGROUP blocks, XADD will pay the O(N) time in order to serve the N clients blocked on the stream getting new data.
	// ACL categories: @write @stream @slow @blocking
	// RESP2 Reply:
	//	One of the following:
	//		- Array reply: an array where each element is an array composed of a two elements containing the key name and the entries reported for that key.
	//			The entries reported are full stream entries, having IDs and the list of all the fields and values. Field and values are guaranteed to be reported
	//			in the same order they were added by XADD.
	//		- Nil reply: if the BLOCK option is given and a timeout occurs, or if there is no stream that can be served.
	// RESP3 Reply:
	//	One of the following:
	//		- Map reply: A map of key-value elements where each element is composed of the key name and the entries reported for that key.
	//			The entries reported are full stream entries, having IDs and the list of all the fields and values. Field and values are guaranteed to be reported
	//			in the same order they were added by XADD.
	//		- Null reply: if the BLOCK option is given and a timeout occurs, or if there is no stream that can be served.
	XReadGroup(ctx context.Context, a XReadGroupArgs) XStreamSliceCmd

	// XTrim
	// Available since: 5.0.0
	// Time complexity: O(N), with N being the number of evicted entries. Constant times are very small however,
	//					since entries are organized in macro nodes containing multiple entries that can be released with a single deallocation.
	// ACL categories: @write @stream @slow
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: The number of entries deleted from the stream.
	// History:
	//	- Starting with Redis version 6.2.0: Added the MINID trimming strategy and the LIMIT option.
	XTrim(ctx context.Context, key string, maxLen int64) IntCmd
	XTrimApprox(ctx context.Context, key string, maxLen int64) IntCmd
	XTrimMaxLen(ctx context.Context, key string, maxLen int64) IntCmd
	XTrimMaxLenApprox(ctx context.Context, key string, maxLen, limit int64) IntCmd
	XTrimMinID(ctx context.Context, key string, minID string) IntCmd
	XTrimMinIDApprox(ctx context.Context, key string, minID string, limit int64) IntCmd
}

type StreamReader interface {
	// XInfoConsumers
	// Available since: 5.0.0
	// Time complexity: O(1)
	// ACL categories: @read @stream @slow
	// RESP2 / RESP3 Reply:
	// 	- Array reply: a list of consumers and their attributes.
	// History:
	//	- Starting with Redis version 7.2.0: Added the inactive field.
	XInfoConsumers(ctx context.Context, key string, group string) XInfoConsumersCmd

	// XInfoGroups
	// Available since: 5.0.0
	// Time complexity: O(1)
	// ACL categories: @read @stream @slow
	// RESP2 / RESP3 Reply:
	// 	- Array reply: a list of consumer groups.
	// History:
	//	- Starting with Redis version 7.0.0: Added the entries-read and lag fields
	XInfoGroups(ctx context.Context, key string) XInfoGroupsCmd

	// XInfoStream
	// Available since: 5.0.0
	// Time complexity: O(1)
	// ACL categories: @read @stream @slow
	// RESP2 Reply:
	//	One of the following:
	//		- Array reply: when the FULL argument is used, a list of information about a stream in summary form.
	//		- Array reply: when the FULL argument is used, a list of information about a stream in extended form.
	// RESP3 Reply:
	//	One of the following:
	//		- Map reply: when the FULL argument was not given, a list of information about a stream in summary form.
	//		- Map reply: when the FULL argument was given, a list of information about a stream in extended form.
	// History:
	//	- Starting with Redis version 6.0.0: Added the FULL modifier.
	//	- Starting with Redis version 7.0.0: Added the max-deleted-entry-id, entries-added, recorded-first-entry-id, entries-read and lag fields
	//	- Starting with Redis version 7.2.0: Added the active-time field, and changed the meaning of seen-time.
	XInfoStream(ctx context.Context, key string) XInfoStreamCmd
	XInfoStreamFull(ctx context.Context, key string, count int64) XInfoStreamFullCmd

	// XLen
	// Available since: 5.0.0
	// Time complexity: O(1)
	// ACL categories: @read @stream @fast
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the number of entries of the stream at key.
	XLen(ctx context.Context, stream string) IntCmd

	// XPending
	// Available since: 5.0.0
	// Time complexity: O(N) with N being the number of elements returned, so asking for a small fixed number of entries per call is O(1).
	//					O(M), where M is the total number of entries scanned when used with the IDLE filter. When the command returns just the summary and the list of consumers is small,
	//					it runs in O(1) time; otherwise, an additional O(N) time for iterating every consumer.
	// ACL categories: @read @stream @slow
	// RESP2 / RESP3 Reply:
	// 	- Array reply: different data depending on the way XPENDING is called, as explained on this page.
	// History:
	//	- Starting with Redis version 6.2.0: Added the IDLE option and exclusive range intervals.
	XPending(ctx context.Context, stream, group string) XPendingCmd
	XPendingExt(ctx context.Context, a XPendingExtArgs) XPendingExtCmd

	// XRange
	// Available since: 5.0.0
	// Time complexity: O(N) with N being the number of elements being returned. If N is constant (e.g. always asking for the first 10 elements with COUNT), you can consider it O(1).
	// ACL categories: @read @stream @slow
	// RESP2 / RESP3 Reply:
	// 	- Array reply: a list of stream entries with IDs matching the specified range.
	// History:
	//	- Starting with Redis version 6.2.0: Added exclusive ranges.
	XRange(ctx context.Context, stream, start, stop string) XMessageSliceCmd
	XRangeN(ctx context.Context, stream, start, stop string, count int64) XMessageSliceCmd

	// XRead
	// Available since: 5.0.0
	// Time complexity: For each stream mentioned: O(N) with N being the number of elements being returned, it means that XREAD-ing with a fixed COUNT is O(1).
	//					Note that when the BLOCK option is used, XADD will pay O(M) time in order to serve the M clients blocked on the stream getting new data.
	// ACL categories: @read @stream @slow @blocking
	// RESP2 Reply:
	//	One of the following:
	//		- Array reply: an array where each element is an array composed of a two elements containing the key name and the entries reported for that key.
	//			The entries reported are full stream entries, having IDs and the list of all the fields and values. Field and values are guaranteed to be reported
	//			in the same order they were added by XADD.
	//		- Nil reply: if the BLOCK option is given and a timeout occurs, or if there is no stream that can be served.
	// RESP3 Reply:
	//	One of the following:
	//		- Map reply: A map of key-value elements where each element is composed of the key name and the entries reported for that key.
	//			The entries reported are full stream entries, having IDs and the list of all the fields and values. Field and values are guaranteed to be reported
	//			in the same order they were added by XADD.
	//		- Null reply: if the BLOCK option is given and a timeout occurs, or if there is no stream that can be served.
	XRead(ctx context.Context, a XReadArgs) XStreamSliceCmd
	XReadStreams(ctx context.Context, streams ...string) XStreamSliceCmd

	// XRevRange
	// Available since: 5.0.0
	// Time complexity: O(N) with N being the number of elements returned. If N is constant (e.g. always asking for the first 10 elements with COUNT), you can consider it O(1).
	// ACL categories: @read @stream @slow
	// RESP2 / RESP3 Reply:
	// 	- Array reply: The command returns the entries with IDs matching the specified range.
	//		The returned entries are complete, which means that the ID and all the fields they are composed of are returned.
	//		Moreover, the entries are returned with their fields and values in the same order as XADD added them.
	// History:
	//	- Starting with Redis version 6.2.0: Added exclusive ranges.
	XRevRange(ctx context.Context, stream string, start, stop string) XMessageSliceCmd
	XRevRangeN(ctx context.Context, stream string, start, stop string, count int64) XMessageSliceCmd
}

func (c *client) XAck(ctx context.Context, stream, group string, ids ...string) IntCmd {
	ctx = c.handler.before(ctx, CommandXAck)
	r := c.adapter.XAck(ctx, stream, group, ids...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) XAdd(ctx context.Context, a XAddArgs) StringCmd {
	if a.NoMkStream {
		ctx = c.handler.before(ctx, CommandXAddNoMKStream)
	} else if a.MaxLen > 0 {
		ctx = c.handler.before(ctx, CommandXAddMaxLen)
	} else if len(a.MinID) > 0 {
		ctx = c.handler.before(ctx, CommandXAddMinID)
	} else if a.Limit > 0 {
		ctx = c.handler.before(ctx, CommandXAddLimit)
	} else {
		ctx = c.handler.before(ctx, CommandXAdd)
	}
	r := newStringCmd(c.Do(ctx, c.builder.XAddCompleted(a)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) XAutoClaim(ctx context.Context, a XAutoClaimArgs) XAutoClaimCmd {
	ctx = c.handler.before(ctx, CommandXAutoClaim)
	r := c.adapter.XAutoClaim(ctx, a)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) XAutoClaimJustID(ctx context.Context, a XAutoClaimArgs) XAutoClaimJustIDCmd {
	ctx = c.handler.before(ctx, CommandXAutoClaimJustID)
	r := c.adapter.XAutoClaimJustID(ctx, a)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) XClaim(ctx context.Context, a XClaimArgs) XMessageSliceCmd {
	ctx = c.handler.before(ctx, CommandXClaim)
	r := c.adapter.XClaim(ctx, a)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) XClaimJustID(ctx context.Context, a XClaimArgs) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandXClaimJustID)
	r := newStringSliceCmd(c.Do(ctx, c.builder.XClaimJustIDCompleted(a)))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) XDel(ctx context.Context, stream string, ids ...string) IntCmd {
	ctx = c.handler.before(ctx, CommandXDel)
	r := c.adapter.XDel(ctx, stream, ids...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) XGroupCreate(ctx context.Context, stream, group, start string) StatusCmd {
	ctx = c.handler.before(ctx, CommandXGroupCreate)
	r := c.adapter.XGroupCreate(ctx, stream, group, start)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) XGroupCreateMkStream(ctx context.Context, stream, group, start string) StatusCmd {
	ctx = c.handler.before(ctx, CommandXGroupCreateMkStream)
	r := c.adapter.XGroupCreateMkStream(ctx, stream, group, start)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) XGroupCreateConsumer(ctx context.Context, stream, group, consumer string) IntCmd {
	ctx = c.handler.before(ctx, CommandXGroupCreateConsumer)
	r := c.adapter.XGroupCreateConsumer(ctx, stream, group, consumer)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) XGroupDelConsumer(ctx context.Context, stream, group, consumer string) IntCmd {
	ctx = c.handler.before(ctx, CommandXGroupDelConsumer)
	r := c.adapter.XGroupDelConsumer(ctx, stream, group, consumer)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) XGroupDestroy(ctx context.Context, stream, group string) IntCmd {
	ctx = c.handler.before(ctx, CommandXGroupDestroy)
	r := c.adapter.XGroupDestroy(ctx, stream, group)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) XGroupSetID(ctx context.Context, stream, group, start string) StatusCmd {
	ctx = c.handler.before(ctx, CommandXGroupSetID)
	r := c.adapter.XGroupSetID(ctx, stream, group, start)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) XInfoConsumers(ctx context.Context, key string, group string) XInfoConsumersCmd {
	ctx = c.handler.before(ctx, CommandXInfoConsumers)
	r := c.adapter.XInfoConsumers(ctx, key, group)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) XInfoGroups(ctx context.Context, key string) XInfoGroupsCmd {
	ctx = c.handler.before(ctx, CommandXInfoGroups)
	r := c.adapter.XInfoGroups(ctx, key)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) XInfoStream(ctx context.Context, key string) XInfoStreamCmd {
	ctx = c.handler.before(ctx, CommandXInfoStream)
	r := c.adapter.XInfoStream(ctx, key)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) XInfoStreamFull(ctx context.Context, key string, count int64) XInfoStreamFullCmd {
	ctx = c.handler.before(ctx, CommandXInfoStreamFull)
	r := c.adapter.XInfoStreamFull(ctx, key, count)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) XLen(ctx context.Context, stream string) IntCmd {
	ctx = c.handler.before(ctx, CommandXLen)
	r := c.adapter.XLen(ctx, stream)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) XPending(ctx context.Context, stream, group string) XPendingCmd {
	ctx = c.handler.before(ctx, CommandXPending)
	r := c.adapter.XPending(ctx, stream, group)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) XPendingExt(ctx context.Context, a XPendingExtArgs) XPendingExtCmd {
	ctx = c.handler.before(ctx, CommandXPendingExt)
	r := c.adapter.XPendingExt(ctx, a)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) XRange(ctx context.Context, stream, start, stop string) XMessageSliceCmd {
	ctx = c.handler.before(ctx, CommandXRange)
	r := c.adapter.XRange(ctx, stream, start, stop)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) XRangeN(ctx context.Context, stream, start, stop string, count int64) XMessageSliceCmd {
	ctx = c.handler.before(ctx, CommandXRangeN)
	r := c.adapter.XRangeN(ctx, stream, start, stop, count)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) XRead(ctx context.Context, a XReadArgs) XStreamSliceCmd {
	ctx = c.handler.before(ctx, CommandXRead)
	r := c.adapter.XRead(ctx, a)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) XReadStreams(ctx context.Context, streams ...string) XStreamSliceCmd {
	ctx = c.handler.before(ctx, CommandXRead)
	r := c.adapter.XReadStreams(ctx, streams...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) XReadGroup(ctx context.Context, a XReadGroupArgs) XStreamSliceCmd {
	ctx = c.handler.before(ctx, CommandXReadGroup)
	r := c.adapter.XReadGroup(ctx, a)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) XRevRange(ctx context.Context, stream string, stop, start string) XMessageSliceCmd {
	ctx = c.handler.before(ctx, CommandXRevRange)
	r := c.adapter.XRevRange(ctx, stream, stop, start)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) XRevRangeN(ctx context.Context, stream string, stop, start string, count int64) XMessageSliceCmd {
	ctx = c.handler.before(ctx, CommandXRevRangeN)
	r := c.adapter.XRevRangeN(ctx, stream, stop, start, count)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) XTrim(ctx context.Context, key string, maxLen int64) IntCmd {
	ctx = c.handler.before(ctx, CommandXTrim)
	r := c.adapter.XTrimMaxLen(ctx, key, maxLen)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) XTrimApprox(ctx context.Context, key string, maxLen int64) IntCmd {
	ctx = c.handler.before(ctx, CommandXTrim)
	r := c.adapter.XTrimMaxLenApprox(ctx, key, maxLen, 0)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) XTrimMaxLen(ctx context.Context, key string, maxLen int64) IntCmd {
	ctx = c.handler.before(ctx, CommandXTrim)
	r := c.adapter.XTrimMaxLen(ctx, key, maxLen)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) XTrimMaxLenApprox(ctx context.Context, key string, maxLen, limit int64) IntCmd {
	if limit > 0 {
		ctx = c.handler.before(ctx, CommandXTrimMaxLenApprox)
	} else {
		ctx = c.handler.before(ctx, CommandXTrim)
	}
	r := c.adapter.XTrimMaxLenApprox(ctx, key, maxLen, limit)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) XTrimMinID(ctx context.Context, key string, minID string) IntCmd {
	ctx = c.handler.before(ctx, CommandXTrimMinID)
	r := c.adapter.XTrimMinID(ctx, key, minID)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) XTrimMinIDApprox(ctx context.Context, key string, minID string, limit int64) IntCmd {
	ctx = c.handler.before(ctx, CommandXTrimMinIDApprox)
	r := c.adapter.XTrimMinIDApprox(ctx, key, minID, limit)
	c.handler.after(ctx, r.Err())
	return r
}
