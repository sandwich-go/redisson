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
	// The CLIENT GETNAME returns the name of the current connection as set by CLIENT SETNAME. Since every new connection starts without an associated name, if no name was assigned a null bulk reply is returned.
	// Return:
	//	Bulk string reply: The connection name, or a null bulk reply if no name is set.
	ClientGetName(ctx context.Context) StringCmd

	// ClientID
	// Available since: 5.0.0
	//Time complexity: O(1)
	//ACL categories: @slow @connection
	//The command just returns the ID of the current connection. Every connection ID has certain guarantees:
	// It is never repeated, so if CLIENT ID returns the same number, the caller can be sure that the underlying client did not disconnect and reconnect the connection, but it is still the same connection.
	// The ID is monotonically incremental. If the ID of a connection is greater than the ID of another connection, it is guaranteed that the second connection was established with the server at a later time.
	// This command is especially useful together with CLIENT UNBLOCK which was introduced also in Redis 5 together with CLIENT ID. Check the CLIENT UNBLOCK command page for a pattern involving the two commands.
	// Return:
	// 	Integer reply
	// 	The id of the client.
	ClientID(ctx context.Context) IntCmd

	// ClientKill
	// Available since: 2.4.0
	// Time complexity: O(N) where N is the number of client connections
	// ACL categories: @admin @slow @dangerous @connection
	// See https://redis.io/commands/client-kill/
	ClientKill(ctx context.Context, ipPort string) StatusCmd

	// ClientKillByFilter
	// Available since: 2.4.0
	// Time complexity: O(N) where N is the number of client connections
	// ACL categories: @admin @slow @dangerous @connection
	// See https://redis.io/commands/client-kill/
	ClientKillByFilter(ctx context.Context, keys ...string) IntCmd

	// ClientList
	// Available since: 2.4.0
	// Time complexity: O(N) where N is the number of client connections
	// ACL categories: @admin @slow @dangerous @connection
	// The CLIENT LIST command returns information and statistics about the client connections server in a mostly human readable format.
	// You can use one of the optional subcommands to filter the list. The TYPE type subcommand filters the list by clients' type, where type is one of normal, master, replica, and pubsub. Note that clients blocked by the MONITOR command belong to the normal class.
	// The ID filter only returns entries for clients with IDs matching the client-id arguments.
	// Return:
	//	Bulk string reply: a unique string, formatted as follows:
	//	One client connection per line (separated by LF)
	//	Each line is composed of a succession of property=value fields separated by a space character.
	ClientList(ctx context.Context) StringCmd

	// ClientPause
	// Available since: 2.9.50
	// Time complexity: O(1)
	// ACL categories: @admin @slow @dangerous @connection
	// See https://redis.io/commands/client-pause/
	ClientPause(ctx context.Context, dur time.Duration) BoolCmd

	// ClientUnpause
	// Available since: 6.2.0
	// Time complexity: O(N) Where N is the number of paused clients
	// ACL categories: @admin @slow @dangerous @connection
	// See https://redis.io/commands/client-unpause/
	ClientUnpause(ctx context.Context) BoolCmd

	// ClientUnblock
	// Available since: 5.0.0
	// Time complexity: O(N) Where N is the number of paused clients
	// ACL categories: @admin @slow @dangerous @connection
	// See https://redis.io/commands/client-unblock/
	ClientUnblock(ctx context.Context, id int64) IntCmd
	ClientUnblockWithError(ctx context.Context, id int64) IntCmd

	// Echo
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @fast @connection
	// Returns message.
	// Return:
	//	Bulk string reply
	Echo(ctx context.Context, message any) StringCmd

	// Ping
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @fast @connection
	// Returns PONG if no argument is provided, otherwise return a copy of the argument as a bulk. This command is often used to test if a connection is still alive, or to measure latency.
	// If the client is subscribed to a channel or a pattern, it will instead return a multi-bulk with a "pong" in the first position and an empty bulk in the second position, unless an argument is provided in which case it returns a copy of the argument.
	// Return:
	// 	Simple string reply, and specifically PONG, when no argument is provided.
	// 	Bulk string reply the argument provided, when applicable.
	Ping(ctx context.Context) StatusCmd

	// Quit
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @fast @connection
	// Ask the server to close the connection. The connection is closed as soon as all pending replies have been written to the client.
	// Return:
	//	Simple string reply: always OK.
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
