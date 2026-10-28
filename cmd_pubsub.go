package redisson

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

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

	// SSubscribe
	// Available since: 7.0.0
	// Time complexity: O(N) where N is the number of shard channels to subscribe to.
	// ACL categories: @pubsub @slow
	// Active channels are automatically resubscribed after slot migration.
	// Messages published during resubscription are not replayed.
	SSubscribe(ctx context.Context, channels ...string) error

	// SUnsubscribe
	// Available since: 7.0.0
	// Time complexity: O(N) where N is the number of shard channels to unsubscribe
	// ACL categories: @pubsub @slow
	// RESP2 / RESP3 Reply:
	// 	- When successful, this command doesn't return anything. Instead, for each shard channel,
	//		one message with the first element being the string sunsubscribe is pushed as a confirmation that the command succeeded.
	SUnsubscribe(ctx context.Context, channels ...string) error

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

	// SSubscribe
	// Available since: 7.0.0
	// Time complexity: O(N) where N is the number of shard channels to subscribe to.
	// ACL categories: @pubsub @slow
	SSubscribe(ctx context.Context, channels ...string) PubSub

	// PSubscribe
	// Available since: 2.0.0
	// Time complexity: O(N) where N is the number of patterns the client is already subscribed to.
	// ACL categories: @pubsub @slow
	// RESP2 / RESP3 Reply:
	// 	- When successful, this command doesn't return anything.
	//		Instead, for each pattern, one message with the first element being the string psubscribe is pushed as a confirmation that the command succeeded.
	PSubscribe(ctx context.Context, patterns ...string) PubSub

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
	r := newStringSliceCmd(c.Do(ctx, c.builder.PubSubChannelsCompleted(pattern)))
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
	r := newStringSliceCmd(c.Do(ctx, c.builder.PubSubShardChannelsCompleted(pattern)))
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
	r := newPubSub(ctx, c, c.handler)

	var err error
	if len(channels) > 0 {
		err = r.Subscribe(ctx, channels...)
	}

	c.handler.after(ctx, err)
	return r
}

func (c *client) SSubscribe(ctx context.Context, channels ...string) PubSub {
	ctx = c.handler.before(ctx, CommandSSubscribe)
	r := newPubSub(ctx, c, c.handler)

	var err error
	if len(channels) > 0 {
		err = r.SSubscribe(ctx, channels...)
	}

	c.handler.after(ctx, err)
	return r
}

func (c *client) PSubscribe(ctx context.Context, patterns ...string) PubSub {
	ctx = c.handler.before(ctx, CommandPSubscribe)
	r := newPubSub(ctx, c, c.handler)

	var err error
	if len(patterns) > 0 {
		err = r.PSubscribe(ctx, patterns...)
	}

	c.handler.after(ctx, err)
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
	handler handler

	ctx    context.Context
	closed AtomicInt32
	cancel context.CancelFunc

	shardMu sync.Mutex
	shards  map[string]*shardSubscription
	monitor *pubSubMonitor
}

type shardSubscription struct {
	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
	err    error // Cleanup result, read only after done is closed.
}

func newPubSub(ctx context.Context, client *client, handler handler) PubSub {
	p := &pubSub{client: client, handler: handler}
	p.ctx, p.cancel = context.WithCancel(ctx)
	p.monitor = newPubSubMonitor(p.ctx, client.v.GetPubSubChanSize())
	if p.monitor.queue == nil {
		p.monitor.defaultReceive = func(message Message) { p.monitor.forward(p.ctx, message) }
	}
	return p
}

func (p *pubSub) isClosed() bool { return p.closed.Get() == 1 }
func (p *pubSub) Close() error {
	if p.closed.CompareAndSwap(0, 1) {
		p.cancel()
	}
	return nil
}

func (p *pubSub) PSubscribe(ctx context.Context, patterns ...string) error {
	ctx = p.handler.before(ctx, CommandPSubscribe)
	cmd := p.client.cmd
	go func() {
		receiveCtx, receive, finish := p.monitor.observeReceiver(p.ctx, "psubscribe", patterns)
		err := cmd.Receive(receiveCtx, cmd.B().Psubscribe().Pattern(patterns...).Build(), func(m rueidis.PubSubMessage) {
			if !p.isClosed() {
				receive(m)
			} else {
				warning(fmt.Sprintf("psubscribe, channel closed, patterns: %s,", strings.Join(patterns, " ")))
			}
		})
		finish(err)
		if err != nil {
			e(fmt.Sprintf("psubscribe failed, patterns: %s, err: %s", strings.Join(patterns, ", "), err.Error()))
		}
	}()
	p.handler.after(ctx, nil)
	return nil
}

func (p *pubSub) Subscribe(ctx context.Context, channels ...string) error {
	ctx = p.handler.before(ctx, CommandSubscribe)
	cmd := p.client.cmd
	go func() {
		receiveCtx, receive, finish := p.monitor.observeReceiver(p.ctx, "subscribe", channels)
		err := cmd.Receive(receiveCtx, cmd.B().Subscribe().Channel(channels...).Build(), func(m rueidis.PubSubMessage) {
			if !p.isClosed() {
				receive(m)
			} else {
				warning(fmt.Sprintf("subscribe, channel closed, channels: %s,", strings.Join(channels, " ")))
			}
		})
		finish(err)
		if err != nil {
			e(fmt.Sprintf("subscribe failed, channels: %s, err: %s", strings.Join(channels, ", "), err.Error()))
		}
	}()
	p.handler.after(ctx, nil)
	return nil
}

func (p *pubSub) SSubscribe(ctx context.Context, channels ...string) error {
	ctx = p.handler.before(ctx, CommandSSubscribe)
	err := p.sSubscribe(ctx, channels)
	p.handler.after(ctx, err)
	return err
}

func (p *pubSub) sSubscribe(ctx context.Context, channels []string) error {
	p.shardMu.Lock()
	defer p.shardMu.Unlock()
	if err := p.ctx.Err(); err != nil {
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if p.shards == nil {
		p.shards = make(map[string]*shardSubscription)
	}
	for _, channel := range channels {
		if prev := p.shards[channel]; prev != nil {
			if prev.ctx.Err() == nil {
				select {
				case <-prev.done:
				default:
					continue // Already subscribed.
				}
			}
			// A previous unsubscribe must finish before subscribing again.
			select {
			case <-prev.done:
			case <-ctx.Done():
				return ctx.Err()
			case <-p.ctx.Done():
				return p.ctx.Err()
			}
		}
		subCtx, cancel := context.WithCancel(p.ctx)
		sub := &shardSubscription{ctx: subCtx, cancel: cancel, done: make(chan struct{})}
		p.shards[channel] = sub
		go p.receiveShard(p.client.cmd, channel, sub)
	}
	return nil
}

func (p *pubSub) receiveShard(cmd rueidis.Client, channel string, sub *shardSubscription) {
	defer close(sub.done)
	defer sub.cancel()
	defer func() {
		// Canceling Receive removes the local listener, not the server subscription.
		ctx, cancel := context.WithTimeout(context.Background(), p.client.v.GetWriteTimeout())
		defer cancel()
		sub.err = cmd.Do(ctx, cmd.B().Sunsubscribe().Channel(channel).Build()).Error()
	}()
	receiveCtx, receive, finish := p.monitor.observeReceiver(sub.ctx, "ssubscribe", []string{channel})
	var receiveError error
	defer func() { finish(receiveError) }()
	for sub.ctx.Err() == nil {
		// Receive's command is recycled on success; rebuild it on every attempt.
		err := cmd.Receive(receiveCtx, cmd.B().Ssubscribe().Channel(channel).Build(), receive)
		receiveError = err
		if err != nil {
			if sub.ctx.Err() == nil && err != rueidis.ErrClosing {
				e(fmt.Sprintf("ssubscribe failed, channel: %s, err: %s", channel, err))
			}
			return
		}
		if sub.ctx.Err() == nil {
			p.monitor.observeRetry("ssubscribe", channel)
		}
		// A slot migration sends SUNSUBSCRIBE without closing the connection.
		// Retry through the cluster client so MOVED refreshes the destination.
		timer := time.NewTimer(50 * time.Millisecond)
		select {
		case <-sub.ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

func (p *pubSub) SUnsubscribe(ctx context.Context, channels ...string) error {
	ctx = p.handler.before(ctx, CommandSUnsubscribe)
	err := p.sUnsubscribe(ctx, channels)
	p.handler.after(ctx, err)
	return err
}

func (p *pubSub) sUnsubscribe(ctx context.Context, channels []string) error {
	p.shardMu.Lock()
	defer p.shardMu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(channels) == 0 {
		for channel := range p.shards {
			channels = append(channels, channel)
		}
	}
	for _, channel := range channels {
		if sub := p.shards[channel]; sub != nil {
			sub.cancel()
		}
	}
	var err error
	for _, channel := range channels {
		if sub := p.shards[channel]; sub != nil {
			select {
			case <-sub.done:
				if err == nil {
					err = sub.err
				}
				delete(p.shards, channel)
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
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
	return p.monitor.channel()
}
