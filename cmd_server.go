package redisson

import (
	"context"
)

type ServerCmdable interface {
	// ACLDryRun
	// Available since: 7.0.0
	// Time complexity: O(1)
	// ACL categories: @admin @slow @dangerous
	// RESP2 / RESP3 Reply:
	//	Any of the following:
	//		- Simple string reply: OK on success.
	//		- Bulk string reply: an error describing why the user can't execute the command.
	ACLDryRun(ctx context.Context, username string, command ...any) StringCmd

	// BgRewriteAOF
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @admin @slow @dangerous
	// RESP2 / RESP3 Reply:
	// 	- Simple string reply: a simple string reply indicating that the rewriting started or is about to start ASAP when the call is executed with success.
	// 	The command may reply with an error in certain cases, as documented above.
	BgRewriteAOF(ctx context.Context) StatusCmd

	// BgSave
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @admin @slow @dangerous
	// RESP2 / RESP3 Reply:
	//	One of the following:
	//		- Simple string reply: Background saving started.
	//		- Simple string reply: Background saving scheduled.
	// History:
	//	- Starting with Redis version 3.2.2: Added the SCHEDULE option.
	BgSave(ctx context.Context) StatusCmd

	// Command
	// Available since: 2.8.13
	// Time complexity: O(N) where N is the total number of Redis commands
	// ACL categories: @slow @connection
	// RESP2 / RESP3 Reply:
	// 	- Array reply: a nested list of command details. The order of the commands in the array is random.
	Command(ctx context.Context) CommandsInfoCmd

	// CommandList
	// Available since: 7.0.0
	// Time complexity: O(N) where N is the total number of Redis commands
	// ACL categories: @slow @connection
	// Options
	// 	- MODULE module-name: get the commands that belong to the module specified by module-name.
	// 	- ACLCAT category: get the commands in the ACL category specified by category.
	// 	- PATTERN pattern: get the commands that match the given glob-like pattern.
	// RESP2 / RESP3 Reply:
	// 	- Array reply: a list of command names.
	CommandList(ctx context.Context, filter FilterBy) StringSliceCmd

	// CommandGetKeys
	// Available since: 2.8.13
	// Time complexity: O(N) where N is the total number of Redis commands
	// ACL categories: @slow @connection
	// RESP2 / RESP3 Reply:
	// 	- Array reply: a list of keys from the given command.
	CommandGetKeys(ctx context.Context, commands ...any) StringSliceCmd

	// CommandGetKeysAndFlags
	// Available since: 7.0.0
	// Time complexity: O(N) where N is the number of arguments to the command
	// ACL categories: @slow @connection
	// RESP2 / RESP3 Reply:
	// 	- Array reply: a list of keys from the given command and their usage flags.
	CommandGetKeysAndFlags(ctx context.Context, commands ...any) KeyFlagsCmd

	// ConfigGet
	// Available since: 2.0.0
	// Time complexity: O(N) when N is the number of configuration parameters provided
	// ACL categories: @admin @slow @dangerous
	// RESP2 / RESP3 Reply:
	// 	- Array reply: a list of configuration parameters matching the provided arguments.
	// History:
	//	- Starting with Redis version 7.0.0: Added the ability to pass multiple pattern parameters in one call
	ConfigGet(ctx context.Context, parameter string) StringStringMapCmd

	// ConfigResetStat
	// Available since: 2.0.0
	// Time complexity: O(1)
	// ACL categories: @admin @slow @dangerous
	// RESP2 / RESP3 Reply:
	// 	- Simple string reply: OK.
	ConfigResetStat(ctx context.Context) StatusCmd

	// ConfigRewrite
	// Available since: 2.8.0
	// Time complexity: O(1)
	// ACL categories: @admin @slow @dangerous
	// RESP2 / RESP3 Reply:
	// 	- Simple string reply: OK when the configuration was rewritten properly. Otherwise an error is returned.
	ConfigRewrite(ctx context.Context) StatusCmd

	// ConfigSet
	// Available since: 2.0.0
	// Time complexity: O(N) when N is the number of configuration parameters provided
	// ACL categories: @admin @slow @dangerous
	// RESP2 / RESP3 Reply:
	// 	- Simple string reply: OK when the configuration was set properly. Otherwise an error is returned.
	// History:
	//	- Starting with Redis version 7.0.0: Added the ability to set multiple parameters in one call.
	ConfigSet(ctx context.Context, parameter, value string) StatusCmd

	// DBSize
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @keyspace @read @fast
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: the number of keys in the currently-selected database.
	DBSize(ctx context.Context) IntCmd

	// FlushAll
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the total number of keys in all databases
	// ACL categories: @keyspace @write @slow @dangerous
	// RESP2 / RESP3 Reply:
	// 	- Simple string reply: OK.
	// History:
	//	- Starting with Redis version 4.0.0: Added the ASYNC flushing mode modifier.
	//	- Starting with Redis version 6.2.0: Added the SYNC flushing mode modifier.
	FlushAll(ctx context.Context) StatusCmd
	FlushAllAsync(ctx context.Context) StatusCmd

	// FlushDB
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the number of keys in the selected database
	// ACL categories: @keyspace @write @slow @dangerous
	// RESP2 / RESP3 Reply:
	// 	- Simple string reply: OK.
	// History:
	//	- Starting with Redis version 4.0.0: Added the ASYNC flushing mode modifier.
	//	- Starting with Redis version 6.2.0: Added the SYNC flushing mode modifier.
	FlushDB(ctx context.Context) StatusCmd
	FlushDBAsync(ctx context.Context) StatusCmd

	// Info
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @slow @dangerous
	// RESP2 / RESP3 Reply:
	// 	- Bulk string reply: a map of info fields, one field per line in the form of <field>:<value> where the value can be a comma separated map like <key>=<val>.
	//		Also contains section header lines starting with # and blank lines.
	//	Lines can contain a section name (starting with a # character) or a property. All the properties are in the form of field:value terminated by \r\n.
	// History:
	//	- Starting with Redis version 7.0.0: Added support for taking multiple section arguments.
	Info(ctx context.Context, section ...string) StringCmd

	// LastSave
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @admin @fast @dangerous
	// RESP2 / RESP3 Reply:
	// 	- Integer reply: UNIX TIME of the last DB save executed with success.
	LastSave(ctx context.Context) IntCmd

	// MemoryUsage
	// Available since: 4.0.0
	// Time complexity: O(N) where N is the number of samples.
	// ACL categories: @read @slow
	// RESP2 Reply:
	//	One of the following:
	//		- Integer reply: the memory usage in bytes.
	//		- Nil reply: if the key does not exist.
	// RESP3 Reply:
	//	One of the following:
	//		- Integer reply: the memory usage in bytes.
	//		- Null reply: if the key does not exist.
	MemoryUsage(ctx context.Context, key string, samples ...int64) IntCmd

	// Save
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the total number of keys in all databases
	// ACL categories: @admin @slow @dangerous
	// RESP2 / RESP3 Reply:
	// 	- Simple string reply: OK.
	Save(ctx context.Context) StatusCmd

	// Shutdown
	// Available since: 1.0.0
	// Time complexity: O(N) when saving, where N is the total number of keys in all databases when saving data, otherwise O(1)
	// ACL categories: @admin @slow @dangerous
	// Options
	// 	- SAVE will force a DB saving operation even if no save points are configured.
	// 	- NOSAVE will prevent a DB saving operation even if one or more save points are configured.
	// 	- NOW skips waiting for lagging replicas, i.e. it bypasses the first step in the shutdown sequence.
	// 	- FORCE ignores any errors that would normally prevent the server from exiting. For details, see the following section.
	// 	- ABORT cancels an ongoing shutdown and cannot be combined with other flags.
	// RESP2 / RESP3 Reply:
	// 	- Simple string reply: OK if ABORT was specified and shutdown was aborted. On successful shutdown,
	//		nothing is returned because the server quits and the connection is closed. On failure, an error is returned.
	// History:
	//	- Starting with Redis version 7.0.0: Added the NOW, FORCE and ABORT modifiers.
	Shutdown(ctx context.Context) StatusCmd
	ShutdownSave(ctx context.Context) StatusCmd
	ShutdownNoSave(ctx context.Context) StatusCmd

	// Time
	// Available since: 2.6.0
	// Time complexity: O(1)
	// ACL categories: @fast
	// RESP2 / RESP3 Reply:
	// 	- Array reply: specifically, a two-element array consisting of the Unix timestamp in seconds and the microseconds' count.
	Time(ctx context.Context) TimeCmd

	// DebugObject
	// Available since: 1.0.0
	// Time complexity: Depends on subcommand.
	// ACL categories: @admin @slow @dangerous
	DebugObject(ctx context.Context, key string) StringCmd
}

func (c *client) ACLDryRun(ctx context.Context, username string, command ...any) StringCmd {
	ctx = c.handler.before(ctx, CommandACLDryRun)
	r := wrapStringCmd(c.adapter.ACLDryRun(ctx, username, command...))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) BgRewriteAOF(ctx context.Context) StatusCmd {
	ctx = c.handler.before(ctx, CommandBgRewriteAOF)
	r := c.adapter.BgRewriteAOF(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) BgSave(ctx context.Context) StatusCmd {
	ctx = c.handler.before(ctx, CommandBgSave)
	r := c.adapter.BgSave(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Command(ctx context.Context) CommandsInfoCmd {
	ctx = c.handler.before(ctx, CommandCommand)
	r := c.adapter.Command(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) CommandList(ctx context.Context, filter FilterBy) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandCommandList)
	r := wrapStringSliceCmd(c.adapter.CommandList(ctx, filter))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) CommandGetKeys(ctx context.Context, commands ...any) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandCommandGetKeys)
	r := wrapStringSliceCmd(c.adapter.CommandGetKeys(ctx, commands...))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) CommandGetKeysAndFlags(ctx context.Context, commands ...any) KeyFlagsCmd {
	ctx = c.handler.before(ctx, CommandCommandGetKeysAndFlags)
	r := c.adapter.CommandGetKeysAndFlags(ctx, commands...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ConfigGet(ctx context.Context, parameter string) StringStringMapCmd {
	ctx = c.handler.before(ctx, CommandConfigGet)
	r := c.adapter.ConfigGet(ctx, parameter)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ConfigResetStat(ctx context.Context) StatusCmd {
	ctx = c.handler.before(ctx, CommandConfigResetStat)
	r := c.adapter.ConfigResetStat(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ConfigRewrite(ctx context.Context) StatusCmd {
	ctx = c.handler.before(ctx, CommandConfigRewrite)
	r := c.adapter.ConfigRewrite(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ConfigSet(ctx context.Context, parameter, value string) StatusCmd {
	ctx = c.handler.before(ctx, CommandConfigSet)
	r := c.adapter.ConfigSet(ctx, parameter, value)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) DBSize(ctx context.Context) IntCmd {
	ctx = c.handler.before(ctx, CommandDBSize)
	r := c.adapter.DBSize(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) FlushAll(ctx context.Context) StatusCmd {
	ctx = c.handler.before(ctx, CommandFlushAll)
	r := c.adapter.FlushAll(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) FlushAllAsync(ctx context.Context) StatusCmd {
	ctx = c.handler.before(ctx, CommandFlushAllAsync)
	r := c.adapter.FlushAllAsync(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) FlushDB(ctx context.Context) StatusCmd {
	ctx = c.handler.before(ctx, CommandFlushDB)
	r := c.adapter.FlushDB(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) FlushDBAsync(ctx context.Context) StatusCmd {
	ctx = c.handler.before(ctx, CommandFlushDBAsync)
	r := c.adapter.FlushDBAsync(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Info(ctx context.Context, section ...string) StringCmd {
	if len(section) > 1 {
		ctx = c.handler.before(ctx, CommandMServerInfo)
	} else {
		ctx = c.handler.before(ctx, CommandServerInfo)
	}
	r := wrapStringCmd(c.adapter.Info(ctx, section...))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) LastSave(ctx context.Context) IntCmd {
	ctx = c.handler.before(ctx, CommandLastSave)
	r := c.adapter.LastSave(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) MemoryUsage(ctx context.Context, key string, samples ...int64) IntCmd {
	ctx = c.handler.before(ctx, CommandMemoryUsage)
	r := c.adapter.MemoryUsage(ctx, key, samples...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Save(ctx context.Context) StatusCmd {
	ctx = c.handler.before(ctx, CommandSave)
	r := c.adapter.Save(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Shutdown(ctx context.Context) StatusCmd {
	ctx = c.handler.before(ctx, CommandShutdown)
	r := c.adapter.Shutdown(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ShutdownSave(ctx context.Context) StatusCmd {
	ctx = c.handler.before(ctx, CommandShutdownSave)
	r := c.adapter.ShutdownSave(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ShutdownNoSave(ctx context.Context) StatusCmd {
	ctx = c.handler.before(ctx, CommandShutdownNoSave)
	r := c.adapter.ShutdownNoSave(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Time(ctx context.Context) TimeCmd {
	ctx = c.handler.before(ctx, CommandTime)
	r := c.adapter.Time(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) DebugObject(ctx context.Context, key string) StringCmd {
	ctx = c.handler.before(ctx, CommandDebugObject)
	r := wrapStringCmd(c.adapter.DebugObject(ctx, key))
	c.handler.after(ctx, r.Err())
	return r
}
