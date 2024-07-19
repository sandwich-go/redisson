package redisson

import (
	"context"
	"strings"
	"time"
)

type ConnectionCmdable interface {
	// ClientGetName
	// Available since: 2.6.9
	// Time complexity: O(1)
	// ACL categories: @slow @connection
	ClientGetName(ctx context.Context) StringCmd

	// ClientID
	// Available since: 5.0.0
	// Time complexity: O(1)
	// ACL categories: @slow @connection
	ClientID(ctx context.Context) IntCmd

	// ClientKill
	// Available since: 2.4.0
	// Time complexity: O(N) where N is the number of client connections
	// ACL categories: @admin @slow @dangerous @connection
	ClientKill(ctx context.Context, ipPort string) StatusCmd

	// ClientKillByFilter
	// Available since: 2.4.0
	// Time complexity: O(N) where N is the number of client connections
	// ACL categories: @admin @slow @dangerous @connection
	ClientKillByFilter(ctx context.Context, keys ...string) IntCmd

	// ClientList
	// Available since: 2.4.0
	// Time complexity: O(N) where N is the number of client connections
	// ACL categories: @admin @slow @dangerous @connection
	ClientList(ctx context.Context) StringCmd

	// ClientPause
	// Available since: 2.9.50
	// Time complexity: O(1)
	// ACL categories: @admin @slow @dangerous @connection
	ClientPause(ctx context.Context, dur time.Duration) BoolCmd

	// ClientUnpause
	// Available since: 6.2.0
	// Time complexity: O(N) Where N is the number of paused clients
	// ACL categories: @admin @slow @dangerous @connection
	ClientUnpause(ctx context.Context) BoolCmd

	// ClientUnblock
	// Available since: 5.0.0
	// Time complexity: O(N) Where N is the number of paused clients
	// ACL categories: @admin @slow @dangerous @connection
	ClientUnblock(ctx context.Context, id int64) IntCmd
	ClientUnblockWithError(ctx context.Context, id int64) IntCmd

	// Echo
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @fast @connection
	Echo(ctx context.Context, message any) StringCmd

	// Ping
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @fast @connection
	Ping(ctx context.Context) StatusCmd

	// Quit
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @fast @connection
	Quit(ctx context.Context) StatusCmd
}

func (c *client) ClientGetName(ctx context.Context) StringCmd {
	ctx = c.handler.before(ctx, CommandClientGetName)
	r := c.adapter.ClientGetName(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ClientID(ctx context.Context) IntCmd {
	ctx = c.handler.before(ctx, CommandClientID)
	r := c.adapter.ClientID(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ClientKill(ctx context.Context, ipPort string) StatusCmd {
	ctx = c.handler.before(ctx, CommandClientKill)
	r := c.adapter.ClientKill(ctx, ipPort)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ClientKillByFilter(ctx context.Context, keys ...string) IntCmd {
	var opt = make(map[string]struct{})
	for i := 0; i < len(keys); i += 2 {
		opt[strings.ToUpper(keys[i])] = struct{}{}
	}
	if _, ok := opt[TYPE]; ok {
		ctx = c.handler.before(ctx, CommandClientKillByFilterWithType)
	} else if _, ok = opt[LADDR]; ok {
		ctx = c.handler.before(ctx, CommandClientKillByFilterWithLADDR)
	} else {
		ctx = c.handler.before(ctx, CommandClientKillByFilter)
	}
	r := c.adapter.ClientKillByFilter(ctx, keys...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ClientList(ctx context.Context) StringCmd {
	ctx = c.handler.before(ctx, CommandClientList)
	r := c.adapter.ClientList(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ClientPause(ctx context.Context, dur time.Duration) BoolCmd {
	ctx = c.handler.before(ctx, CommandClientPause)
	r := c.adapter.ClientPause(ctx, dur)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ClientUnpause(ctx context.Context) BoolCmd {
	ctx = c.handler.before(ctx, CommandClientUnpause)
	r := c.adapter.ClientUnpause(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ClientUnblock(ctx context.Context, id int64) IntCmd {
	ctx = c.handler.before(ctx, CommandClientUnblock)
	r := c.adapter.ClientUnblock(ctx, id)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ClientUnblockWithError(ctx context.Context, id int64) IntCmd {
	ctx = c.handler.before(ctx, CommandClientUnblockWithError)
	r := c.adapter.ClientUnblockWithError(ctx, id)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Echo(ctx context.Context, message any) StringCmd {
	ctx = c.handler.before(ctx, CommandEcho)
	r := c.adapter.Echo(ctx, message)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Ping(ctx context.Context) StatusCmd {
	ctx = c.handler.before(ctx, CommandPing)
	r := c.adapter.Ping(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Quit(ctx context.Context) StatusCmd {
	ctx = c.handler.before(ctx, CommandQuit)
	r := c.adapter.Quit(ctx)
	c.handler.after(ctx, r.Err())
	return r
}
