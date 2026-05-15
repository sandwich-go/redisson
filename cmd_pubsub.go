package redisson

import (
	"context"
	"fmt"
	"strings"

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
	msgCh   *unboundedChan[Message]
	handler handler

	ctx    context.Context
	closed AtomicInt32
	cancel context.CancelFunc
}

func newPubSub(ctx context.Context, client *client, handler handler) PubSub {
	p := &pubSub{client: client, msgCh: newUnboundedChan[Message](ctx, client.v.GetPubSubChanSize()), handler: handler}
	p.ctx, p.cancel = context.WithCancel(ctx)
	return p
}

func (p *pubSub) isClosed() bool { return p.closed.Get() == 1 }
func (p *pubSub) Close() error {
	if p.closed.CompareAndSwap(0, 1) {
		p.cancel()
	}
	return nil
}

// 订阅类方法的 err 永远是同步路径上的 nil：rueidis.Receive 是阻塞调用，
// 必须放到独立 goroutine 后台跑；调用方拿到的 err 仅反映"启动是否成功"，
// 真正的接收错误通过日志输出（启动错误几乎不可能，因为没有立即同步操作）。
//
// 注意：旧版本中曾把 err 同时被 goroutine 内写、主路径读，造成 data race；
// 新版去掉无效的写读，明确"同步路径恒返回 nil"。
//
// Close 路径：cancel(p.ctx) 后 rueidis.Receive 会返回，回调不再被调用；
// 但极端情况下 Close 之后仍可能有少量回调正在执行（rueidis 内部调度），
// 此时 forwardMessage 用 select 在 ctx.Done 与 In 之间二选一，
// 避免"UnboundedChan process goroutine 已退出 → In 缓冲打满 → 回调永久阻塞"
// 导致 rueidis 连接池无法释放与 goroutine 泄漏。

// forwardMessage 把 PubSubMessage 投递到 msgCh.In；
// 若 pubSub 已 Close（ctx canceled）则丢弃并打 warning，绝不阻塞。
func (p *pubSub) forwardMessage(kind string, names []string, m rueidis.PubSubMessage) {
	if p.isClosed() {
		warning(fmt.Sprintf("%s, channel closed, %s: %s,", kind, kind, strings.Join(names, " ")))
		return
	}
	select {
	case p.msgCh.In <- m:
	case <-p.ctx.Done():
		warning(fmt.Sprintf("%s, channel closed during send, %s: %s,", kind, kind, strings.Join(names, " ")))
	}
}

func (p *pubSub) PSubscribe(ctx context.Context, patterns ...string) error {
	ctx = p.handler.before(ctx, CommandPSubscribe)
	go func() {
		err := p.client.cmd.Receive(p.ctx, p.client.cmd.B().Psubscribe().Pattern(patterns...).Build(), func(m rueidis.PubSubMessage) {
			p.forwardMessage("psubscribe", patterns, m)
		})
		if err != nil {
			e(fmt.Sprintf("psubscribe failed, patterns: %s, err: %s", strings.Join(patterns, ", "), err.Error()))
		}
	}()
	p.handler.after(ctx, nil)
	return nil
}

func (p *pubSub) Subscribe(ctx context.Context, channels ...string) error {
	ctx = p.handler.before(ctx, CommandSubscribe)
	go func() {
		err := p.client.cmd.Receive(p.ctx, p.client.cmd.B().Subscribe().Channel(channels...).Build(), func(m rueidis.PubSubMessage) {
			p.forwardMessage("subscribe", channels, m)
		})
		if err != nil {
			e(fmt.Sprintf("subscribe failed, channels: %s, err: %s", strings.Join(channels, ", "), err.Error()))
		}
	}()
	p.handler.after(ctx, nil)
	return nil
}

func (p *pubSub) SSubscribe(ctx context.Context, channels ...string) error {
	ctx = p.handler.before(ctx, CommandSSubscribe)
	go func() {
		err := p.client.cmd.Receive(p.ctx, p.client.cmd.B().Ssubscribe().Channel(channels...).Build(), func(m rueidis.PubSubMessage) {
			p.forwardMessage("ssubscribe", channels, m)
		})
		if err != nil {
			e(fmt.Sprintf("ssubscribe failed, channels: %s, err: %s", strings.Join(channels, ", "), err.Error()))
		}
	}()
	p.handler.after(ctx, nil)
	return nil
}

func (p *pubSub) SUnsubscribe(ctx context.Context, channels ...string) error {
	ctx = p.handler.before(ctx, CommandSUnsubscribe)
	err := p.client.cmd.Do(ctx, p.client.cmd.B().Sunsubscribe().Channel(channels...).Build()).Error()
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
	return p.msgCh.Out
}
