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
	// RESP2 Reply:
	//	One of the following:
	//		- Bulk string reply: the connection name of the current connection.
	//		- Nil reply: the connection name was not set.
	// RESP3 Reply:
	//	One of the following:
	//		- Bulk string reply: the connection name of the current connection.
	//		- Null reply: the connection name was not set.
	ClientGetName(ctx context.Context) StringCmd

	// ClientID
	// Available since: 5.0.0
	// Time complexity: O(1)
	// ACL categories: @slow @connection
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the ID of the client.
	ClientID(ctx context.Context) IntCmd

	// ClientKill
	// Available since: 2.4.0
	// Time complexity: O(N) where N is the number of client connections
	// ACL categories: @admin @slow @dangerous @connection
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- Simple string reply: OK when called in 3 argument format and the connection has been closed.
	//		- Integer reply: when called in filter/value format, the number of clients killed.
	// History:
	//	- Starting with Redis version 2.8.12: Added new filter format.
	//	- Starting with Redis version 2.8.12: ID option.
	//	- Starting with Redis version 3.2.0: Added master type in for TYPE option.
	//	- Starting with Redis version 5.0.0: Replaced slave TYPE with replica. slave still supported for backward compatibility.
	//	- Starting with Redis version 6.2.0: LADDR option.
	ClientKill(ctx context.Context, ipPort string) StatusCmd
	ClientKillByFilter(ctx context.Context, keys ...string) IntCmd

	// ClientList
	// Available since: 2.4.0
	// Time complexity: O(N) where N is the number of client connections
	// ACL categories: @admin @slow @dangerous @connection
	// RESP2 / RESP3 Reply:
	// 	- Bulk string reply: information and statistics about client connections.
	// History:
	//	- Starting with Redis version 2.8.12: Added unique client id field.
	//	- Starting with Redis version 5.0.0: Added optional TYPE filter.
	//	- Starting with Redis version 6.0.0: Added user field.
	//	- Starting with Redis version 6.2.0: Added argv-mem, tot-mem, laddr and redir fields and the optional ID filter.
	//	- Starting with Redis version 7.0.0: Added resp, multi-mem, rbs and rbp fields.
	//	- Starting with Redis version 7.0.3: Added ssub field.
	ClientList(ctx context.Context) StringCmd

	// ClientPause
	// Available since: 2.9.50
	// Time complexity: O(1)
	// ACL categories: @admin @slow @dangerous @connection
	// RESP2 / RESP3 Reply:
	// 	- Simple string reply: OK or an error if the timeout is invalid.
	// History:
	//	- Starting with Redis version 6.2.0: CLIENT PAUSE WRITE mode added along with the mode option.
	ClientPause(ctx context.Context, dur time.Duration) BoolCmd

	// ClientUnpause
	// Available since: 6.2.0
	// Time complexity: O(N) Where N is the number of paused clients
	// ACL categories: @admin @slow @dangerous @connection
	// RESP2 / RESP3 Reply:
	// 	- Simple string reply: OK.
	ClientUnpause(ctx context.Context) BoolCmd

	// ClientUnblock
	// Available since: 5.0.0
	// Time complexity: O(N) Where N is the number of paused clients
	// ACL categories: @admin @slow @dangerous @connection
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- Integer reply: 1 if the client was unblocked successfully.
	//		- Integer reply: 0 if the client wasn't unblocked.
	ClientUnblock(ctx context.Context, id int64) IntCmd
	ClientUnblockWithError(ctx context.Context, id int64) IntCmd

	// Echo
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @fast @connection
	// RESP2 / RESP3 Reply:
	// 	- Bulk string reply: the given string.
	Echo(ctx context.Context, message any) StringCmd

	// Ping
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @fast @connection
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- Simple string reply: PONG when no argument is provided.
	//		- Bulk string reply: the provided argument.
	Ping(ctx context.Context) StatusCmd

	// Quit
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @fast @connection
	// RESP2 / RESP3 Reply:
	// 	- Simple string reply: OK.
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
