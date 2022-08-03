package sandwich_redis

import (
	"context"
)

type PubSub interface {
	// Subscribe
	// Available since: 2.0.0
	// Time complexity: O(N) where N is the number of channels to subscribe to.
	// ACL categories: @pubsub @slow
	// Subscribes the client to the specified channels.
	// Once the client enters the subscribed state it is not supposed to issue any other commands, except for additional SUBSCRIBE, SSUBSCRIBE, PSUBSCRIBE, UNSUBSCRIBE, SUNSUBSCRIBE, PUNSUBSCRIBE, PING, RESET and QUIT commands.
	Subscribe(ctx context.Context, channels ...string) error

	// Unsubscribe
	// Available since: 2.0.0
	// Time complexity: O(N) where N is the number of clients already subscribed to a channel.
	// ACL categories: @pubsub @slow
	// Unsubscribes the client from the given channels, or from all of them if none is given.
	// When no channels are specified, the client is unsubscribed from all the previously subscribed channels. In this case, a message for every unsubscribed channel will be sent to the client.
	Unsubscribe(ctx context.Context, channels ...string) error

	// PSubscribe
	// Available since: 2.0.0
	// Time complexity: O(N) where N is the number of patterns the client is already subscribed to.
	// ACL categories: @pubsub @slow
	// Subscribes the client to the given patterns.
	// Supported glob-style patterns:
	//	h?llo subscribes to hello, hallo and hxllo
	//	h*llo subscribes to hllo and heeeello
	//	h[ae]llo subscribes to hello and hallo, but not hillo
	// Use \ to escape special characters if you want to match them verbatim.
	PSubscribe(ctx context.Context, patterns ...string) error

	// PUnsubscribe
	// Available since: 2.0.0
	// Time complexity: O(N+M) where N is the number of patterns the client is already subscribed and M is the number of total patterns subscribed in the system (by any client).
	// ACL categories: @pubsub @slow
	// Unsubscribes the client from the given patterns, or from all of them if none is given.
	// When no patterns are specified, the client is unsubscribed from all the previously subscribed patterns. In this case, a message for every unsubscribed pattern will be sent to the client.
	PUnsubscribe(ctx context.Context, patterns ...string) error

	// Channel
	// Receive Message by chan
	Channel() <-chan *Message

	// Close
	// Release the hold connection
	Close() error
}

type PubSubCmdable interface {
	// Publish
	// Available since: 2.0.0
	// Time complexity: O(N+M) where N is the number of clients subscribed to the receiving channel and M is the total number of subscribed patterns (by any client).
	// ACL categories: @pubsub @fast
	// Posts a message to the given channel.
	// In a Redis Cluster clients can publish to every node. The cluster makes sure that published messages are forwarded as needed, so clients can subscribe to any channel by connecting to any one of the nodes.
	// Return:
	// 	Integer reply: the number of clients that received the message. Note that in a Redis Cluster, only clients that are connected to the same node as the publishing client are included in the count.
	Publish(ctx context.Context, channel string, message interface{}) IntCmd

	// PubSubChannels
	// Available since: 2.8.0
	// Time complexity: O(N) where N is the number of active channels, and assuming constant time pattern matching (relatively short channels and patterns)
	// ACL categories: @pubsub @slow
	// Lists the currently active channels.
	// An active channel is a Pub/Sub channel with one or more subscribers (excluding clients subscribed to patterns).
	// If no pattern is specified, all the channels are listed, otherwise if pattern is specified only channels matching the specified glob-style pattern are listed.
	// Cluster note: in a Redis Cluster clients can subscribe to every node, and can also publish to every other node. The cluster will make sure that published messages are forwarded as needed. That said, PUBSUB's replies in a cluster only report information from the node's Pub/Sub context, rather than the entire cluster.
	// Return:
	//	Array reply: a list of active channels, optionally matching the specified pattern.
	PubSubChannels(ctx context.Context, pattern string) StringSliceCmd

	// PubSubNumPat
	// Available since: 2.8.0
	// Time complexity: O(1)
	// ACL categories: @pubsub @slow
	// Returns the number of unique patterns that are subscribed to by clients (that are performed using the PSUBSCRIBE command).
	// Note that this isn't the count of clients subscribed to patterns, but the total number of unique patterns all the clients are subscribed to.
	// Cluster note: in a Redis Cluster clients can subscribe to every node, and can also publish to every other node. The cluster will make sure that published messages are forwarded as needed. That said, PUBSUB's replies in a cluster only report information from the node's Pub/Sub context, rather than the entire cluster.
	// Return:
	//	Integer reply: the number of patterns all the clients are subscribed to.
	PubSubNumPat(ctx context.Context) IntCmd

	// PubSubNumSub
	// Available since: 2.8.0
	// Time complexity: O(N) for the NUMSUB subcommand, where N is the number of requested channels
	// ACL categories: @pubsub @slow
	// Returns the number of subscribers (exclusive of clients subscribed to patterns) for the specified channels.
	// Note that it is valid to call this command without channels. In this case it will just return an empty list.
	// Cluster note: in a Redis Cluster clients can subscribe to every node, and can also publish to every other node. The cluster will make sure that published messages are forwarded as needed. That said, PUBSUB's replies in a cluster only report information from the node's Pub/Sub context, rather than the entire cluster.
	// Return:
	//	Array reply: a list of channels and number of subscribers for every channel.
	//	The format is channel, count, channel, count, ..., so the list is flat. The order in which the channels are listed is the same as the order of the channels specified in the command call.
	PubSubNumSub(ctx context.Context, channels ...string) StringIntMapCmd

	// Subscribe
	// Available since: 2.0.0
	// Time complexity: O(N) where N is the number of channels to subscribe to.
	// ACL categories: @pubsub @slow
	// Subscribes the client to the specified channels.
	// Once the client enters the subscribed state it is not supposed to issue any other commands, except for additional SUBSCRIBE, SSUBSCRIBE, PSUBSCRIBE, UNSUBSCRIBE, SUNSUBSCRIBE, PUNSUBSCRIBE, PING, RESET and QUIT commands.
	Subscribe(ctx context.Context, channels ...string) PubSub
}

func (c *client) Publish(ctx context.Context, channel string, message interface{}) IntCmd {
	ctx = c.handler.before(ctx, CommandPublish)
	r := c.cmdable.Publish(ctx, channel, message)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) PubSubChannels(ctx context.Context, pattern string) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandPubSubChannels)
	r := c.cmdable.PubSubChannels(ctx, pattern)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) PubSubNumPat(ctx context.Context) IntCmd {
	ctx = c.handler.before(ctx, CommandPubSubNumPat)
	r := c.cmdable.PubSubNumPat(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) PubSubNumSub(ctx context.Context, channels ...string) StringIntMapCmd {
	ctx = c.handler.before(ctx, CommandPubSubNumSub)
	r := c.cmdable.PubSubNumSub(ctx, channels...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Subscribe(ctx context.Context, channels ...string) PubSub {
	ctx = c.handler.before(ctx, CommandSubscribe)
	r := c.cmdable.Subscribe(ctx, channels...)
	c.handler.after(ctx, nil)
	return r
}
