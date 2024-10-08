package redisson

import (
	"context"
	"strings"
	"time"
)

type ConnectionCmdable interface {
	// Select
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @fast @connection
	// Select the Redis logical database having the specified zero-based numeric index. New connections always use the database 0.
	// Selectable Redis databases are a form of namespacing: all databases are still persisted in the same RDB / AOF file. However different databases can have keys with the same name, and commands like FLUSHDB, SWAPDB or RANDOMKEY work on specific databases.
	// In practical terms, Redis databases should be used to separate different keys belonging to the same application (if needed), and not to use a single Redis instance for multiple unrelated applications.
	// When using Redis Cluster, the SELECT command cannot be used, since Redis Cluster only supports database zero. In the case of a Redis Cluster, having multiple databases would be useless and an unnecessary source of complexity. Commands operating atomically on a single database would not be possible with the Redis Cluster design and goals.
	// Since the currently selected database is a property of the connection, clients should track the currently selected database and re-select it on reconnection. While there is no command in order to query the selected database in the current connection, the CLIENT LIST output shows, for each client, the currently selected database.
	// Return:
	//	Simple string reply
	Select(ctx context.Context, index int) StatusCmd

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

	// Echo
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @fast @connection
	// Returns message.
	// Return:
	//	Bulk string reply
	Echo(ctx context.Context, message interface{}) StringCmd

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

func (c *client) Select(ctx context.Context, index int) StatusCmd {
	return do[StatusCmd](ctx, c.handler, CommandSelect, func(ctx context.Context) StatusCmd {
		return c.cmdable.Select(ctx, index)
	})
}

func (c *client) ClientGetName(ctx context.Context) StringCmd {
	return do[StringCmd](ctx, c.handler, CommandClientGetName, func(ctx context.Context) StringCmd {
		return c.cmdable.ClientGetName(ctx)
	})
}

func (c *client) ClientID(ctx context.Context) IntCmd {
	return do[IntCmd](ctx, c.handler, CommandClientID, func(ctx context.Context) IntCmd {
		return c.cmdable.ClientID(ctx)
	})
}

func (c *client) ClientKill(ctx context.Context, ipPort string) StatusCmd {
	return do[StatusCmd](ctx, c.handler, CommandClientKill, func(ctx context.Context) StatusCmd {
		return c.cmdable.ClientKill(ctx, ipPort)
	})
}

func (c *client) ClientKillByFilter(ctx context.Context, keys ...string) IntCmd {
	var opt = make(map[string]struct{})
	for i := 0; i < len(keys); i += 2 {
		opt[strings.ToUpper(keys[i])] = struct{}{}
	}
	var cmd Command
	if _, ok := opt[LADDR]; ok {
		cmd = CommandClientKillByFilterByLAddr
	} else {
		cmd = CommandClientKillByFilter
	}
	return do[IntCmd](ctx, c.handler, cmd, func(ctx context.Context) IntCmd {
		return c.cmdable.ClientKillByFilter(ctx, keys...)
	})
}

func (c *client) ClientList(ctx context.Context) StringCmd {
	return do[StringCmd](ctx, c.handler, CommandClientList, func(ctx context.Context) StringCmd {
		return c.cmdable.ClientList(ctx)
	})
}

func (c *client) ClientPause(ctx context.Context, dur time.Duration) BoolCmd {
	return do[BoolCmd](ctx, c.handler, CommandClientPause, func(ctx context.Context) BoolCmd {
		return c.cmdable.ClientPause(ctx, dur)
	})
}

func (c *client) Echo(ctx context.Context, message interface{}) StringCmd {
	return do[StringCmd](ctx, c.handler, CommandEcho, func(ctx context.Context) StringCmd {
		return c.cmdable.Echo(ctx, message)
	})
}

func (c *client) Ping(ctx context.Context) StatusCmd {
	return do[StatusCmd](ctx, c.handler, CommandPing, func(ctx context.Context) StatusCmd {
		return c.cmdable.Ping(ctx)
	})
}

func (c *client) Quit(ctx context.Context) StatusCmd {
	return do[StatusCmd](ctx, c.handler, CommandQuit, func(ctx context.Context) StatusCmd {
		return c.cmdable.Quit(ctx)
	})
}
