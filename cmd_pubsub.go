package redisson

import (
	"context"
	"github.com/redis/rueidis"
)

type PubSub interface {
	// Subscribe
	// Available since: 2.0.0
	// Time complexity: O(N) where N is the number of channels to subscribe to.
	// ACL categories: @pubsub @slow
	// RESP2 / RESP3 Reply:
	// 	- When successful, this command doesn't return anything.
	//		Instead, for each pattern, one message with the first element being the string psubscribe is pushed as a confirmation that the command succeeded.
	Subscribe(ctx context.Context, channels ...string) error

	// Unsubscribe
	// Available since: 2.0.0
	// Time complexity: O(N) where N is the number of channels to unsubscribe.
	// ACL categories: @pubsub @slow
	// RESP2 / RESP3 Reply:
	// 	- When successful, this command doesn't return anything.
	//		Instead, for each channel, one message with the first element being the string unsubscribe is pushed as a confirmation that the command succeeded.
	Unsubscribe(ctx context.Context, channels ...string) error

	// PSubscribe
	// Available since: 2.0.0
	// Time complexity: O(N) where N is the number of patterns the client is already subscribed to.
	// ACL categories: @pubsub @slow
	// RESP2 / RESP3 Reply:
	// 	- When successful, this command doesn't return anything.
	//		Instead, for each pattern, one message with the first element being the string psubscribe is pushed as a confirmation that the command succeeded.
	PSubscribe(ctx context.Context, patterns ...string) error

	// PUnsubscribe
	// Available since: 2.0.0
	// Time complexity: O(N) where N is the number of patterns to unsubscribe.
	// ACL categories: @pubsub @slow
	// RESP2 / RESP3 Reply:
	// 	- When successful, this command doesn't return anything.
	//		Instead, for each pattern, one message with the first element being the string punsubscribe is pushed as a confirmation that the command succeeded.
	PUnsubscribe(ctx context.Context, patterns ...string) error

	// Channel
	// Receive Message by chan
	Channel() <-chan Message

	// Close
	// Release the hold connection
	Close() error
}

type PubSubCmdable interface {
	// Publish
	// Available since: 2.0.0
	// Time complexity: O(N+M) where N is the number of clients subscribed to the receiving channel and M is the total number of subscribed patterns (by any client).
	// ACL categories: @pubsub @fast
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the number of clients that the message was sent to.
	//		Note that in a Redis Cluster, only clients that are connected to the same node as the publishing client are included in the count.
	Publish(ctx context.Context, channel string, message any) IntCmd

	// SPublish
	// Available since: 7.0.0
	// Time complexity: O(N) where N is the number of clients subscribed to the receiving shard channel.
	// ACL categories: @pubsub @fast
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the number of clients that the message was sent to. Note that in a Redis Cluster,
	//		only clients that are connected to the same node as the publishing client are included in the count
	SPublish(ctx context.Context, channel string, message any) IntCmd

	// PubSubChannels
	// Available since: 2.8.0
	// Time complexity: O(N) where N is the number of active channels, and assuming constant time pattern matching (relatively short channels and patterns)
	// ACL categories: @pubsub @slow
	// RESP2 / RESP3 Reply:
	// 	- Array reply: a list of active channels, optionally matching the specified pattern.
	PubSubChannels(ctx context.Context, pattern string) StringSliceCmd

	// PubSubNumPat
	// Available since: 2.8.0
	// Time complexity: O(1)
	// ACL categories: @pubsub @slow
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the number of patterns all the clients are subscribed to.
	PubSubNumPat(ctx context.Context) IntCmd

	// PubSubNumSub
	// Available since: 2.8.0
	// Time complexity: O(N) for the NUMSUB subcommand, where N is the number of requested channels
	// ACL categories: @pubsub @slow
	// RESP2 / RESP3 Reply:
	// 	- Array reply: the number of subscribers per channel, each even element (including the 0th) is channel name, each odd element is the number of subscribers
	PubSubNumSub(ctx context.Context, channels ...string) StringIntMapCmd

	// Subscribe
	// Available since: 2.0.0
	// Time complexity: O(N) where N is the number of channels to subscribe to.
	// ACL categories: @pubsub @slow
	// RESP2 / RESP3 Reply:
	// 	- When successful, this command doesn't return anything.
	//		Instead, for each channel, one message with the first element being the string subscribe is pushed as a confirmation that the command succeeded.
	Subscribe(ctx context.Context, channels ...string) PubSub

	// PubSubShardChannels
	// Available since: 7.0.0
	// Time complexity: O(N) where N is the number of active shard channels, and assuming constant time pattern matching (relatively short shard channels).
	// ACL categories: @pubsub @slow
	// RESP2 / RESP3 Reply:
	// 	- Array reply: a list of active channels, optionally matching the specified pattern.
	PubSubShardChannels(ctx context.Context, pattern string) StringSliceCmd

	// PubSubShardNumSub
	// Available since: 7.0.0
	// Time complexity: O(N) for the SHARDNUMSUB subcommand, where N is the number of requested shard channels
	// ACL categories: @pubsub @slow
	// RESP2 / RESP3 Reply:
	// 	- Array reply: the number of subscribers per shard channel, each even element (including the 0th) is channel name, each odd element is the number of subscribers.
	PubSubShardNumSub(ctx context.Context, channels ...string) StringIntMapCmd
}

func (c *client) Publish(ctx context.Context, channel string, message any) IntCmd {
	ctx = c.handler.before(ctx, CommandPublish)
	r := c.adapter.Publish(ctx, channel, message)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) SPublish(ctx context.Context, channel string, message any) IntCmd {
	ctx = c.handler.before(ctx, CommandSPublish)
	r := c.adapter.SPublish(ctx, channel, message)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) PubSubChannels(ctx context.Context, pattern string) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandPubSubChannels)
	r := c.adapter.PubSubChannels(ctx, pattern)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) PubSubNumSub(ctx context.Context, channels ...string) StringIntMapCmd {
	ctx = c.handler.before(ctx, CommandPubSubNumSub)
	r := c.adapter.PubSubNumSub(ctx, channels...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) PubSubNumPat(ctx context.Context) IntCmd {
	ctx = c.handler.before(ctx, CommandPubSubNumPat)
	r := c.adapter.PubSubNumPat(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) PubSubShardChannels(ctx context.Context, pattern string) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandPubSubShardChannels)
	r := c.adapter.PubSubShardChannels(ctx, pattern)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) PubSubShardNumSub(ctx context.Context, channels ...string) StringIntMapCmd {
	ctx = c.handler.before(ctx, CommandPubSubShardNumSub)
	r := c.adapter.PubSubShardNumSub(ctx, channels...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Subscribe(ctx context.Context, channels ...string) PubSub {
	ctx = c.handler.before(ctx, CommandSubscribe)
	r := newPubSub(ctx, c, c.handler, channels...)
	c.handler.after(ctx, nil)
	return r
}

func (c *client) Receive(ctx context.Context, cb func(Message), channels ...string) error {
	return c.cmd.Receive(ctx, c.cmd.B().Subscribe().Channel(channels...).Build(), func(msg rueidis.PubSubMessage) {
		cb(msg)
	})
}

func (c *client) PReceive(ctx context.Context, cb func(Message), patterns ...string) error {
	return c.cmd.Receive(ctx, c.cmd.B().Psubscribe().Pattern(patterns...).Build(), func(msg rueidis.PubSubMessage) {
		cb(msg)
	})
}

type pubSub struct {
	client  *client
	msgCh   chan Message
	handler handler

	ctx    context.Context
	cancel context.CancelFunc
}

func newPubSub(ctx context.Context, client *client, handler handler, channels ...string) PubSub {
	// chan size todo, use goredis.ChannelOption?
	p := &pubSub{client: client, msgCh: make(chan Message, 100), handler: handler}
	p.ctx, p.cancel = context.WithCancel(ctx)
	if len(channels) > 0 {
		_ = p.Subscribe(ctx, channels...)
	}
	return p
}

func (p *pubSub) Close() error {
	close(p.msgCh)
	p.cancel()
	return nil
}

func (p *pubSub) PSubscribe(ctx context.Context, patterns ...string) error {
	ctx = p.handler.before(ctx, CommandPSubscribe)
	var err error
	go func() {
		err = p.client.cmd.Receive(p.ctx, p.client.cmd.B().Psubscribe().Pattern(patterns...).Build(), func(m rueidis.PubSubMessage) {
			p.msgCh <- m
		})
	}()
	p.handler.after(ctx, err)
	return err
}

func (p *pubSub) Subscribe(ctx context.Context, channels ...string) error {
	ctx = p.handler.before(ctx, CommandSubscribe)
	var err error
	go func() {
		err = p.client.cmd.Receive(p.ctx, p.client.cmd.B().Subscribe().Channel(channels...).Build(), func(m rueidis.PubSubMessage) {
			p.msgCh <- m
		})
	}()
	p.handler.after(ctx, err)
	return err
}

func (p *pubSub) Unsubscribe(ctx context.Context, channels ...string) error {
	ctx = p.handler.before(ctx, CommandUnsubscribe)
	err := p.client.cmd.Do(ctx, p.client.cmd.B().Unsubscribe().Channel(channels...).Build()).Error()
	p.handler.after(ctx, err)
	return err
}

func (p *pubSub) PUnsubscribe(ctx context.Context, patterns ...string) error {
	ctx = p.handler.before(ctx, CommandPUnsubscribe)
	err := p.client.cmd.Do(ctx, p.client.cmd.B().Punsubscribe().Pattern(patterns...).Build()).Error()
	p.handler.after(ctx, err)
	return err
}

func (p *pubSub) Channel() <-chan Message {
	return p.msgCh
}
