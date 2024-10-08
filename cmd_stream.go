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
	// The XACK command removes one or multiple messages from the Pending Entries List (PEL) of a stream consumer group. A message is pending, and as such stored inside the PEL, when it was delivered to some consumer, normally as a side effect of calling XREADGROUP, or when a consumer took ownership of a message calling XCLAIM. The pending message was delivered to some consumer but the server is yet not sure it was processed at least once. So new calls to XREADGROUP to grab the messages history for a consumer (for instance using an ID of 0), will return such message. Similarly the pending message will be listed by the XPENDING command, that inspects the PEL.
	// Once a consumer successfully processes a message, it should call XACK so that such message does not get processed again, and as a side effect, the PEL entry about this message is also purged, releasing memory from the Redis server.
	// Return:
	//	Integer reply, specifically:
	//	The command returns the number of messages successfully acknowledged. Certain message IDs may no longer be part of the PEL (for example because they have already been acknowledged), and XACK will not count them as successfully acknowledged.
	XAck(ctx context.Context, stream, group string, ids ...string) IntCmd

	// XAdd
	// Available since: 5.0.0
	// Time complexity: O(1) when adding a new entry, O(N) when trimming where N being the number of entries evicted.
	// ACL categories: @write @stream @fast
	// Appends the specified stream entry to the stream at the specified key. If the key does not exist, as a side effect of running this command the key is created with a stream value. The creation of stream's key can be disabled with the NOMKSTREAM option.
	// An entry is composed of a list of field-value pairs. The field-value pairs are stored in the same order they are given by the user. Commands that read the stream, such as XRANGE or XREAD, are guaranteed to return the fields and values exactly in the same order they were added by XADD.
	// XADD is the only Redis command that can add data to a stream, but there are other commands, such as XDEL and XTRIM, that are able to remove data from a stream.
	// Return
	//	Bulk string reply, specifically:
	//	The command returns the ID of the added entry. The ID is the one auto-generated if * is passed as ID argument, otherwise the command just returns the same ID specified by the user during insertion.
	//	The command returns a Null reply when used with the NOMKSTREAM option and the key doesn't exist.
	// See https://redis.io/commands/xadd/
	XAdd(ctx context.Context, a XAddArgs) StringCmd

	// XAutoClaim
	// Available since: 6.2.0
	// Time complexity: O(1) if COUNT is small.
	// ACL categories: @write @stream @fast
	// See https://redis.io/commands/xautoclaim/
	XAutoClaim(ctx context.Context, a XAutoClaimArgs) XAutoClaimCmd

	// XAutoClaimJustID
	// Available since: 6.2.0
	// Time complexity: O(1) if COUNT is small.
	// ACL categories: @write @stream @fast
	// See https://redis.io/commands/xautoclaim/
	XAutoClaimJustID(ctx context.Context, a XAutoClaimArgs) XAutoClaimJustIDCmd

	// XClaim
	// Available since: 5.0.0
	// Time complexity: O(log N) with N being the number of messages in the PEL of the consumer group.
	// ACL categories: @write @stream @fast
	// In the context of a stream consumer group, this command changes the
	// See https://redis.io/commands/xclaim/
	XClaim(ctx context.Context, a XClaimArgs) XMessageSliceCmd

	// XClaimJustID
	// Available since: 5.0.0
	// Time complexity: O(log N) with N being the number of messages in the PEL of the consumer group.
	// ACL categories: @write @stream @fast
	// In the context of a stream consumer group, this command changes the
	// See https://redis.io/commands/xclaim/
	XClaimJustID(ctx context.Context, a XClaimArgs) StringSliceCmd

	// XDel
	// Available since: 5.0.0
	// Time complexity: O(1) for each single item to delete in the stream, regardless of the stream size.
	// ACL categories: @write @stream @fast
	// Removes the specified entries from a stream, and returns the number of entries deleted. This number may be less than the number of IDs passed to the command in the case where some of the specified IDs do not exist in the stream.
	// Normally you may think at a Redis stream as an append-only data structure, however Redis streams are represented in memory, so we are also able to delete entries. This may be useful, for instance, in order to comply with certain privacy policies.
	// Return:
	// 	Integer reply: the number of entries actually deleted.
	XDel(ctx context.Context, stream string, ids ...string) IntCmd

	// XGroupCreate
	// Available since: 5.0.0
	// Time complexity: O(1)
	// ACL categories: @write @stream @slow
	// This command creates a new consumer group uniquely identified by <groupname> for the stream stored at <key>.
	// Every group has a unique name in a given stream. When a consumer group with the same name already exists, the command returns a -BUSYGROUP error.
	// The command's <id> argument specifies the last delivered entry in the stream from the new group's perspective. The special ID $ means the ID of the last entry in the stream, but you can provide any valid ID instead. For example, if you want the group's consumers to fetch the entire stream from the beginning, use zero as the starting ID for the consumer group:
	// XGROUP CREATE mystream mygroup 0
	// By default, the XGROUP CREATE command insists that the target stream exists and returns an error when it doesn't. However, you can use the optional MKSTREAM subcommand as the last argument after the <id> to automatically create the stream (with length of 0) if it doesn't exist:
	// XGROUP CREATE mystream mygroup $ MKSTREAM
	// The optional entries_read named argument can be specified to enable consumer group lag tracking for an arbitrary ID. An arbitrary ID is any ID that isn't the ID of the stream's first entry, its last entry or the zero ("0-0") ID. This can be useful you know exactly how many entries are between the arbitrary ID (excluding it) and the stream's last entry. In such cases, the entries_read can be set to the stream's entries_added subtracted with the number of entries.
	// Return:
	// 	Simple string reply: OK on success.
	XGroupCreate(ctx context.Context, stream, group, start string) StatusCmd

	// XGroupCreateMkStream
	// Available since: 5.0.0
	// Time complexity: O(1)
	// ACL categories: @write @stream @slow
	// This command creates a new consumer group uniquely identified by <groupname> for the stream stored at <key>.
	// Every group has a unique name in a given stream. When a consumer group with the same name already exists, the command returns a -BUSYGROUP error.
	// The command's <id> argument specifies the last delivered entry in the stream from the new group's perspective. The special ID $ means the ID of the last entry in the stream, but you can provide any valid ID instead. For example, if you want the group's consumers to fetch the entire stream from the beginning, use zero as the starting ID for the consumer group:
	// XGROUP CREATE mystream mygroup 0
	// By default, the XGROUP CREATE command insists that the target stream exists and returns an error when it doesn't. However, you can use the optional MKSTREAM subcommand as the last argument after the <id> to automatically create the stream (with length of 0) if it doesn't exist:
	// XGROUP CREATE mystream mygroup $ MKSTREAM
	// The optional entries_read named argument can be specified to enable consumer group lag tracking for an arbitrary ID. An arbitrary ID is any ID that isn't the ID of the stream's first entry, its last entry or the zero ("0-0") ID. This can be useful you know exactly how many entries are between the arbitrary ID (excluding it) and the stream's last entry. In such cases, the entries_read can be set to the stream's entries_added subtracted with the number of entries.
	// Return:
	// 	Simple string reply: OK on success.
	XGroupCreateMkStream(ctx context.Context, stream, group, start string) StatusCmd

	// XGroupCreateConsumer
	// Available since: 6.2.0
	// Time complexity: O(1)
	// ACL categories: @write @stream @slow
	// Create a consumer named <consumername> in the consumer group <groupname> of the stream that's stored at <key>.
	// Consumers are also created automatically whenever an operation, such as XREADGROUP, references a consumer that doesn't exist.
	// Return:
	// 	Integer reply: the number of created consumers (0 or 1)
	XGroupCreateConsumer(ctx context.Context, stream, group, consumer string) IntCmd

	// XGroupDelConsumer
	// Available since: 5.0.0
	// Time complexity: O(1)
	// ACL categories: @write @stream @slow
	// The XGROUP DELCONSUMER command deletes a consumer from the consumer group.
	// Sometimes it may be useful to remove old consumers since they are no longer used.
	// Note, however, that any pending messages that the consumer had will become unclaimable after it was deleted. It is strongly recommended, therefore, that any pending messages are claimed or acknowledged prior to deleting the consumer from the group.
	// Return:
	//	Integer reply: the number of pending messages that the consumer had before it was deleted
	XGroupDelConsumer(ctx context.Context, stream, group, consumer string) IntCmd

	// XGroupDestroy
	// Available since: 5.0.0
	// Time complexity: O(N) where N is the number of entries in the group's pending entries list (PEL).
	// ACL categories: @write @stream @slow
	// The XGROUP DESTROY command completely destroys a consumer group.
	// The consumer group will be destroyed even if there are active consumers, and pending messages, so make sure to call this command only when really needed.
	// Return:
	// 	Integer reply: the number of destroyed consumer groups (0 or 1)
	XGroupDestroy(ctx context.Context, stream, group string) IntCmd

	// XGroupSetID
	// Available since: 5.0.0
	// Time complexity: O(1)
	// ACL categories: @write @stream @slow
	// Set the last delivered ID for a consumer group.
	// Normally, a consumer group's last delivered ID is set when the group is created with XGROUP CREATE. The XGROUP SETID command allows modifying the group's last delivered ID, without having to delete and recreate the group. For instance if you want the consumers in a consumer group to re-process all the messages in a stream, you may want to set its next ID to 0:
	// XGROUP SETID mystream mygroup 0
	// The optional entries_read argument can be specified to enable consumer group lag tracking for an arbitrary ID. An arbitrary ID is any ID that isn't the ID of the stream's first entry, its last entry or the zero ("0-0") ID. This can be useful you know exactly how many entries are between the arbitrary ID (excluding it) and the stream's last entry. In such cases, the entries_read can be set to the stream's entries_added subtracted with the number of entries.
	// Return:
	// 	Simple string reply: OK on success.
	XGroupSetID(ctx context.Context, stream, group, start string) StatusCmd

	// XReadGroup
	// Available since: 5.0.0
	// Time complexity: For each stream mentioned: O(M) with M being the number of elements returned. If M is constant (e.g. always asking for the first 10 elements with COUNT), you can consider it O(1). On the other side when XREADGROUP blocks, XADD will pay the O(N) time in order to serve the N clients blocked on the stream getting new data.
	// ACL categories: @write @stream @slow @blocking
	// See https://redis.io/commands/xreadgroup/
	XReadGroup(ctx context.Context, a XReadGroupArgs) XStreamSliceCmd

	// XTrim
	// Available since: 5.0.0
	// Time complexity: O(N), with N being the number of evicted entries. Constant times are very small however, since entries are organized in macro nodes containing multiple entries that can be released with a single deallocation.
	// ACL categories: @write @stream @slow
	// See https://redis.io/commands/xtrim/
	XTrim(ctx context.Context, key string, maxLen int64) IntCmd

	// XTrimApprox
	// Available since: 5.0.0
	// Time complexity: O(N), with N being the number of evicted entries. Constant times are very small however, since entries are organized in macro nodes containing multiple entries that can be released with a single deallocation.
	// ACL categories: @write @stream @slow
	// See https://redis.io/commands/xtrim/
	XTrimApprox(ctx context.Context, key string, maxLen int64) IntCmd

	// XTrimMaxLen
	// Available since: 5.0.0
	// Time complexity: O(N), with N being the number of evicted entries. Constant times are very small however, since entries are organized in macro nodes containing multiple entries that can be released with a single deallocation.
	// ACL categories: @write @stream @slow
	// See https://redis.io/commands/xtrim/
	XTrimMaxLen(ctx context.Context, key string, maxLen int64) IntCmd

	// XTrimMaxLenApprox
	// Available since: 6.2.0
	// Time complexity: O(N), with N being the number of evicted entries. Constant times are very small however, since entries are organized in macro nodes containing multiple entries that can be released with a single deallocation.
	// ACL categories: @write @stream @slow
	// See https://redis.io/commands/xtrim/
	XTrimMaxLenApprox(ctx context.Context, key string, maxLen, limit int64) IntCmd

	// XTrimMinID
	// Available since: 6.2.0
	// Time complexity: O(N), with N being the number of evicted entries. Constant times are very small however, since entries are organized in macro nodes containing multiple entries that can be released with a single deallocation.
	// ACL categories: @write @stream @slow
	// See https://redis.io/commands/xtrim/
	XTrimMinID(ctx context.Context, key string, minID string) IntCmd

	// XTrimMinIDApprox
	// Available since: 6.2.0
	// Time complexity: O(N), with N being the number of evicted entries. Constant times are very small however, since entries are organized in macro nodes containing multiple entries that can be released with a single deallocation.
	// ACL categories: @write @stream @slow
	// See https://redis.io/commands/xtrim/
	XTrimMinIDApprox(ctx context.Context, key string, minID string, limit int64) IntCmd
}

type StreamReader interface {
	// XInfoConsumers
	// Available since: 5.0.0
	// Time complexity: O(1)
	// ACL categories: @read @stream @slow
	// This command returns the list of consumers that belong to the <groupname> consumer group of the stream stored at <key>.
	// The following information is provided for each consumer in the group:
	// 	name: the consumer's name
	// 	pending: the number of pending messages for the client, which are messages that were delivered but are yet to be acknowledged
	// 	idle: the number of milliseconds that have passed since the consumer last interacted with the server
	// Return:
	//	Array reply: a list of consumers.
	XInfoConsumers(ctx context.Context, key string, group string) XInfoConsumersCmd

	// XInfoGroups
	// Available since: 5.0.0
	// Time complexity: O(1)
	// ACL categories: @read @stream @slow
	// See https://redis.io/commands/xinfo-groups/
	XInfoGroups(ctx context.Context, key string) XInfoGroupsCmd

	// XInfoStream
	// Available since: 5.0.0
	// Time complexity: O(1)
	// ACL categories: @read @stream @slow
	//	This command returns information about the stream stored at <key>.
	//	The informative details provided by this command are:
	//	length: the number of entries in the stream (see XLEN)
	//	radix-tree-keys: the number of keys in the underlying radix data structure
	//	radix-tree-nodes: the number of nodes in the underlying radix data structure
	//	groups: the number of consumer groups defined for the stream
	//	last-generated-id: the ID of the least-recently entry that was added to the stream
	//	max-deleted-entry-id: the maximal entry ID that was deleted from the stream
	//	entries-added: the count of all entries added to the stream during its lifetime
	//	first-entry: the ID and field-value tuples of the first entry in the stream
	//	last-entry: the ID and field-value tuples of the last entry in the stream
	// The optional FULL modifier provides a more verbose reply. When provided, the FULL reply includes an entries array that consists of the stream entries (ID and field-value tuples) in ascending order. Furthermore, groups is also an array, and for each of the consumer groups it consists of the information reported by XINFO GROUPS and XINFO CONSUMERS.
	// The COUNT option can be used to limit the number of stream and PEL entries that are returned (The first <count> entries are returned). The default COUNT is 10 and a COUNT of 0 means that all entries will be returned (execution time may be long if the stream has a lot of entries).
	// Return:
	//	Array reply: a list of informational bits
	XInfoStream(ctx context.Context, key string) XInfoStreamCmd

	// XInfoStreamFull
	// Available since: 6.0.0
	// Time complexity: O(1)
	// ACL categories: @read @stream @slow
	// See XInfoStream
	XInfoStreamFull(ctx context.Context, key string, count int) XInfoStreamFullCmd

	// XLen
	// Available since: 5.0.0
	// Time complexity: O(1)
	// ACL categories: @read @stream @fast
	// Returns the number of entries inside a stream. If the specified key does not exist the command returns zero, as if the stream was empty. However note that unlike other Redis types, zero-length streams are possible, so you should call TYPE or EXISTS in order to check if a key exists or not.
	// Streams are not auto-deleted once they have no entries inside (for instance after an XDEL call), because the stream may have consumer groups associated with it.
	// Return:
	//	Integer reply: the number of entries of the stream at key.
	XLen(ctx context.Context, stream string) IntCmd

	// XPending
	// Available since: 5.0.0
	// Time complexity: O(N) with N being the number of elements returned, so asking for a small fixed number of entries per call is O(1). O(M), where M is the total number of entries scanned when used with the IDLE filter. When the command returns just the summary and the list of consumers is small, it runs in O(1) time; otherwise, an additional O(N) time for iterating every consumer.
	// ACL categories: @read @stream @slow
	// See https://redis.io/commands/xpending/
	XPending(ctx context.Context, stream, group string) XPendingCmd

	// XPendingExt
	// Available since: 5.0.0
	// Time complexity: O(N) with N being the number of elements returned, so asking for a small fixed number of entries per call is O(1). O(M), where M is the total number of entries scanned when used with the IDLE filter. When the command returns just the summary and the list of consumers is small, it runs in O(1) time; otherwise, an additional O(N) time for iterating every consumer.
	// ACL categories: @read @stream @slow
	// Starting with Redis version 6.2.0: Added the IDLE option and exclusive range intervals.
	// See https://redis.io/commands/xpending/
	XPendingExt(ctx context.Context, a XPendingExtArgs) XPendingExtCmd

	// XRange
	// Available since: 5.0.0
	// Time complexity: O(N) with N being the number of elements being returned. If N is constant (e.g. always asking for the first 10 elements with COUNT), you can consider it O(1).
	// ACL categories: @read @stream @slow
	// See https://redis.io/commands/xrange/
	XRange(ctx context.Context, stream, start, stop string) XMessageSliceCmd

	// XRangeN
	// Available since: 5.0.0
	// Time complexity: O(N) with N being the number of elements being returned. If N is constant (e.g. always asking for the first 10 elements with COUNT), you can consider it O(1).
	// ACL categories: @read @stream @slow
	// See https://redis.io/commands/xrange/
	XRangeN(ctx context.Context, stream, start, stop string, count int64) XMessageSliceCmd

	// XRead
	// Available since: 5.0.0
	// Time complexity: For each stream mentioned: O(N) with N being the number of elements being returned, it means that XREAD-ing with a fixed COUNT is O(1). Note that when the BLOCK option is used, XADD will pay O(M) time in order to serve the M clients blocked on the stream getting new data.
	// ACL categories: @read @stream @slow @blocking
	// See https://redis.io/commands/xread/
	XRead(ctx context.Context, a XReadArgs) XStreamSliceCmd

	// XReadStreams
	// Available since: 5.0.0
	// Time complexity: For each stream mentioned: O(N) with N being the number of elements being returned, it means that XREAD-ing with a fixed COUNT is O(1). Note that when the BLOCK option is used, XADD will pay O(M) time in order to serve the M clients blocked on the stream getting new data.
	// ACL categories: @read @stream @slow @blocking
	// See https://redis.io/commands/xread/
	XReadStreams(ctx context.Context, streams ...string) XStreamSliceCmd

	// XRevRange
	// Available since: 5.0.0
	// Time complexity: O(N) with N being the number of elements returned. If N is constant (e.g. always asking for the first 10 elements with COUNT), you can consider it O(1).
	// ACL categories: @read @stream @slow
	// This command is exactly like XRANGE, but with the notable difference of returning the entries in reverse order, and also taking the start-end range in reverse order: in XREVRANGE you need to state the end ID and later the start ID, and the command will produce all the element between (or exactly like) the two IDs, starting from the end side.
	// So for instance, to get all the elements from the higher ID to the lower ID one could use:
	// XREVRANGE somestream + -
	// Similarly to get just the last element added into the stream it is enough to send:
	// XREVRANGE somestream + - COUNT 1
	// Return
	//	Array reply, specifically:
	//	The command returns the entries with IDs matching the specified range, from the higher ID to the lower ID matching. The returned entries are complete, that means that the ID and all the fields they are composed are returned. Moreover the entries are returned with their fields and values in the exact same order as XADD added them.
	XRevRange(ctx context.Context, stream string, start, stop string) XMessageSliceCmd

	// XRevRangeN
	// Available since: 5.0.0
	// Time complexity: O(N) with N being the number of elements returned. If N is constant (e.g. always asking for the first 10 elements with COUNT), you can consider it O(1).
	// ACL categories: @read @stream @slow
	// This command is exactly like XRANGE, but with the notable difference of returning the entries in reverse order, and also taking the start-end range in reverse order: in XREVRANGE you need to state the end ID and later the start ID, and the command will produce all the element between (or exactly like) the two IDs, starting from the end side.
	// So for instance, to get all the elements from the higher ID to the lower ID one could use:
	// XREVRANGE somestream + -
	// Similarly to get just the last element added into the stream it is enough to send:
	// XREVRANGE somestream + - COUNT 1
	// Return
	//	Array reply, specifically:
	//	The command returns the entries with IDs matching the specified range, from the higher ID to the lower ID matching. The returned entries are complete, that means that the ID and all the fields they are composed are returned. Moreover the entries are returned with their fields and values in the exact same order as XADD added them.
	XRevRangeN(ctx context.Context, stream string, start, stop string, count int64) XMessageSliceCmd
}

func (c *client) XAck(ctx context.Context, stream, group string, ids ...string) IntCmd {
	return do[IntCmd](ctx, c.handler, CommandXAck, func(ctx context.Context) IntCmd {
		return c.cmdable.XAck(ctx, stream, group, ids...)
	})
}

func (c *client) XAdd(ctx context.Context, a XAddArgs) StringCmd {
	var cmd Command
	if a.NoMkStream {
		cmd = CommandXAddNoMkStream
	} else if len(a.MinID) > 0 {
		cmd = CommandXAddMinId
	} else if a.Limit > 0 {
		cmd = CommandXAddLimit
	} else {
		cmd = CommandXAdd
	}
	return do[StringCmd](ctx, c.handler, cmd, func(ctx context.Context) StringCmd {
		return c.cmdable.XAdd(ctx, a)
	})
}

func (c *client) XAutoClaim(ctx context.Context, a XAutoClaimArgs) XAutoClaimCmd {
	return do[XAutoClaimCmd](ctx, c.handler, CommandXAutoClaim, func(ctx context.Context) XAutoClaimCmd {
		return c.cmdable.XAutoClaim(ctx, a)
	})
}

func (c *client) XAutoClaimJustID(ctx context.Context, a XAutoClaimArgs) XAutoClaimJustIDCmd {
	return do[XAutoClaimJustIDCmd](ctx, c.handler, CommandXAutoClaim, func(ctx context.Context) XAutoClaimJustIDCmd {
		return c.cmdable.XAutoClaimJustID(ctx, a)
	})
}

func (c *client) XClaim(ctx context.Context, a XClaimArgs) XMessageSliceCmd {
	return do[XMessageSliceCmd](ctx, c.handler, CommandXClaim, func(ctx context.Context) XMessageSliceCmd {
		return c.cmdable.XClaim(ctx, a)
	})
}

func (c *client) XClaimJustID(ctx context.Context, a XClaimArgs) StringSliceCmd {
	return do[StringSliceCmd](ctx, c.handler, CommandXClaim, func(ctx context.Context) StringSliceCmd {
		return c.cmdable.XClaimJustID(ctx, a)
	})
}

func (c *client) XDel(ctx context.Context, stream string, ids ...string) IntCmd {
	return do[IntCmd](ctx, c.handler, CommandXDel, func(ctx context.Context) IntCmd {
		return c.cmdable.XDel(ctx, stream, ids...)
	})
}

func (c *client) XGroupCreate(ctx context.Context, stream, group, start string) StatusCmd {
	return do[StatusCmd](ctx, c.handler, CommandXGroupCreate, func(ctx context.Context) StatusCmd {
		return c.cmdable.XGroupCreate(ctx, stream, group, start)
	})
}

func (c *client) XGroupCreateMkStream(ctx context.Context, stream, group, start string) StatusCmd {
	return do[StatusCmd](ctx, c.handler, CommandXGroupCreate, func(ctx context.Context) StatusCmd {
		return c.cmdable.XGroupCreateMkStream(ctx, stream, group, start)
	})
}

func (c *client) XGroupCreateConsumer(ctx context.Context, stream, group, consumer string) IntCmd {
	return do[IntCmd](ctx, c.handler, CommandXGroupCreateConsumer, func(ctx context.Context) IntCmd {
		return c.cmdable.XGroupCreateConsumer(ctx, stream, group, consumer)
	})
}

func (c *client) XGroupDelConsumer(ctx context.Context, stream, group, consumer string) IntCmd {
	return do[IntCmd](ctx, c.handler, CommandXGroupDelConsumer, func(ctx context.Context) IntCmd {
		return c.cmdable.XGroupDelConsumer(ctx, stream, group, consumer)
	})
}

func (c *client) XGroupDestroy(ctx context.Context, stream, group string) IntCmd {
	return do[IntCmd](ctx, c.handler, CommandXGroupDestroy, func(ctx context.Context) IntCmd {
		return c.cmdable.XGroupDestroy(ctx, stream, group)
	})
}

func (c *client) XGroupSetID(ctx context.Context, stream, group, start string) StatusCmd {
	return do[StatusCmd](ctx, c.handler, CommandXGroupSetID, func(ctx context.Context) StatusCmd {
		return c.cmdable.XGroupSetID(ctx, stream, group, start)
	})
}

func (c *client) XInfoConsumers(ctx context.Context, key string, group string) XInfoConsumersCmd {
	return do[XInfoConsumersCmd](ctx, c.handler, CommandXInfoConsumers, func(ctx context.Context) XInfoConsumersCmd {
		return c.cmdable.XInfoConsumers(ctx, key, group)
	})
}

func (c *client) XInfoGroups(ctx context.Context, key string) XInfoGroupsCmd {
	return do[XInfoGroupsCmd](ctx, c.handler, CommandXInfoGroups, func(ctx context.Context) XInfoGroupsCmd {
		return c.cmdable.XInfoGroups(ctx, key)
	})
}

func (c *client) XInfoStream(ctx context.Context, key string) XInfoStreamCmd {
	return do[XInfoStreamCmd](ctx, c.handler, CommandXInfoStream, func(ctx context.Context) XInfoStreamCmd {
		return c.cmdable.XInfoStream(ctx, key)
	})
}

func (c *client) XInfoStreamFull(ctx context.Context, key string, count int) XInfoStreamFullCmd {
	return do[XInfoStreamFullCmd](ctx, c.handler, CommandXInfoStreamFull, func(ctx context.Context) XInfoStreamFullCmd {
		return c.cmdable.XInfoStreamFull(ctx, key, count)
	})
}

func (c *client) XLen(ctx context.Context, stream string) IntCmd {
	return do[IntCmd](ctx, c.handler, CommandXLen, func(ctx context.Context) IntCmd {
		return c.cmdable.XLen(ctx, stream)
	})
}

func (c *client) XPending(ctx context.Context, stream, group string) XPendingCmd {
	return do[XPendingCmd](ctx, c.handler, CommandXPending, func(ctx context.Context) XPendingCmd {
		return c.cmdable.XPending(ctx, stream, group)
	})
}

func (c *client) XPendingExt(ctx context.Context, a XPendingExtArgs) XPendingExtCmd {
	var cmd Command
	if a.Idle != 0 {
		cmd = CommandXPendingIdle
	} else {
		cmd = CommandXPending
	}
	return do[XPendingExtCmd](ctx, c.handler, cmd, func(ctx context.Context) XPendingExtCmd {
		return c.cmdable.XPendingExt(ctx, a)
	})
}

func (c *client) XRange(ctx context.Context, stream, start, stop string) XMessageSliceCmd {
	return do[XMessageSliceCmd](ctx, c.handler, CommandXRange, func(ctx context.Context) XMessageSliceCmd {
		return c.cmdable.XRange(ctx, stream, start, stop)
	})
}

func (c *client) XRangeN(ctx context.Context, stream, start, stop string, count int64) XMessageSliceCmd {
	return do[XMessageSliceCmd](ctx, c.handler, CommandXRange, func(ctx context.Context) XMessageSliceCmd {
		return c.cmdable.XRangeN(ctx, stream, start, stop, count)
	})
}

func (c *client) XRead(ctx context.Context, a XReadArgs) XStreamSliceCmd {
	return do[XStreamSliceCmd](ctx, c.handler, CommandXRead, func(ctx context.Context) XStreamSliceCmd {
		return c.cmdable.XRead(ctx, a)
	})
}

func (c *client) XReadStreams(ctx context.Context, streams ...string) XStreamSliceCmd {
	return do[XStreamSliceCmd](ctx, c.handler, CommandXRead, func(ctx context.Context) XStreamSliceCmd {
		return c.cmdable.XReadStreams(ctx, streams...)
	})
}

func (c *client) XReadGroup(ctx context.Context, a XReadGroupArgs) XStreamSliceCmd {
	return do[XStreamSliceCmd](ctx, c.handler, CommandXReadGroup, func(ctx context.Context) XStreamSliceCmd {
		return c.cmdable.XReadGroup(ctx, a)
	})
}

func (c *client) XRevRange(ctx context.Context, stream string, start, stop string) XMessageSliceCmd {
	return do[XMessageSliceCmd](ctx, c.handler, CommandXRevRange, func(ctx context.Context) XMessageSliceCmd {
		return c.cmdable.XRevRange(ctx, stream, start, stop)
	})
}

func (c *client) XRevRangeN(ctx context.Context, stream string, start, stop string, count int64) XMessageSliceCmd {
	return do[XMessageSliceCmd](ctx, c.handler, CommandXRevRange, func(ctx context.Context) XMessageSliceCmd {
		return c.cmdable.XRevRangeN(ctx, stream, start, stop, count)
	})
}

func (c *client) XTrim(ctx context.Context, key string, maxLen int64) IntCmd {
	return do[IntCmd](ctx, c.handler, CommandXTrim, func(ctx context.Context) IntCmd {
		return c.cmdable.XTrim(ctx, key, maxLen)
	})
}

func (c *client) XTrimApprox(ctx context.Context, key string, maxLen int64) IntCmd {
	return do[IntCmd](ctx, c.handler, CommandXTrim, func(ctx context.Context) IntCmd {
		return c.cmdable.XTrimApprox(ctx, key, maxLen)
	})
}

func (c *client) XTrimMaxLen(ctx context.Context, key string, maxLen int64) IntCmd {
	return do[IntCmd](ctx, c.handler, CommandXTrim, func(ctx context.Context) IntCmd {
		return c.cmdable.XTrimMaxLen(ctx, key, maxLen)
	})
}

func (c *client) XTrimMaxLenApprox(ctx context.Context, key string, maxLen, limit int64) IntCmd {
	var cmd Command
	if limit > 0 {
		cmd = CommandXTrimLimit
	} else {
		cmd = CommandXTrim
	}
	return do[IntCmd](ctx, c.handler, cmd, func(ctx context.Context) IntCmd {
		return c.cmdable.XTrimMaxLenApprox(ctx, key, maxLen, limit)
	})
}

func (c *client) XTrimMinID(ctx context.Context, key string, minID string) IntCmd {
	return do[IntCmd](ctx, c.handler, CommandXTrimMinId, func(ctx context.Context) IntCmd {
		return c.cmdable.XTrimMinID(ctx, key, minID)
	})
}

func (c *client) XTrimMinIDApprox(ctx context.Context, key string, minID string, limit int64) IntCmd {
	return do[IntCmd](ctx, c.handler, CommandXTrimMinId, func(ctx context.Context) IntCmd {
		return c.cmdable.XTrimMinIDApprox(ctx, key, minID, limit)
	})
}
