package redisson

import (
	"context"
)

type ServerCmdable interface {
	// ACLDryRun
	// Available since: 7.0.0
	// Time complexity: O(1)
	// ACL categories: @admin @slow @dangerous
	ACLDryRun(ctx context.Context, username string, command ...any) StringCmd

	// BgRewriteAOF
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @admin @slow @dangerous
	BgRewriteAOF(ctx context.Context) StatusCmd

	// BgSave
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @admin @slow @dangerous
	BgSave(ctx context.Context) StatusCmd

	// Command
	// Available since: 2.8.13
	// Time complexity: O(N) where N is the total number of Redis commands
	// ACL categories: @slow @connection
	Command(ctx context.Context) CommandsInfoCmd

	// CommandList
	// Available since: 7.0.0
	// Time complexity: O(N) where N is the total number of Redis commands
	// ACL categories: @slow @connection
	CommandList(ctx context.Context, filter FilterBy) StringSliceCmd

	// CommandGetKeys
	// Available since: 2.8.13
	// Time complexity: O(N) where N is the total number of Redis commands
	// ACL categories: @slow @connection
	CommandGetKeys(ctx context.Context, commands ...any) StringSliceCmd

	// CommandGetKeysAndFlags
	// Available since: 7.0.0
	// Time complexity: O(N) where N is the total number of Redis commands
	// ACL categories: @slow @connection
	CommandGetKeysAndFlags(ctx context.Context, commands ...any) KeyFlagsCmd

	// ConfigGet
	// Available since: 2.0.0
	// Time complexity: O(N) when N is the number of configuration parameters provided
	// ACL categories: @admin @slow @dangerous
	ConfigGet(ctx context.Context, parameter string) StringStringMapCmd

	// ConfigResetStat
	// Available since: 2.0.0
	// Time complexity: O(1)
	// ACL categories: @admin @slow @dangerous
	ConfigResetStat(ctx context.Context) StatusCmd

	// ConfigRewrite
	// Available since: 2.8.0
	// Time complexity: O(1)
	// ACL categories: @admin @slow @dangerous
	ConfigRewrite(ctx context.Context) StatusCmd

	// ConfigSet
	// Available since: 2.0.0
	// Time complexity: O(N) when N is the number of configuration parameters provided
	// ACL categories: @admin @slow @dangerous
	ConfigSet(ctx context.Context, parameter, value string) StatusCmd

	// DBSize
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @keyspace @read @fast
	DBSize(ctx context.Context) IntCmd

	// FlushAll
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the total number of keys in all databases
	// ACL categories: @keyspace @write @slow @dangerous
	FlushAll(ctx context.Context) StatusCmd

	// FlushAllAsync
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the total number of keys in all databases
	// ACL categories: @keyspace @write @slow @dangerous
	FlushAllAsync(ctx context.Context) StatusCmd

	// FlushDB
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the number of keys in the selected database
	// ACL categories: @keyspace @write @slow @dangerous
	FlushDB(ctx context.Context) StatusCmd

	// FlushDBAsync
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the number of keys in the selected database
	// ACL categories: @keyspace @write @slow @dangerous
	FlushDBAsync(ctx context.Context) StatusCmd

	// Info
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @slow @dangerous
	Info(ctx context.Context, section ...string) StringCmd

	// LastSave
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @admin @fast @dangerous
	LastSave(ctx context.Context) IntCmd

	// MemoryUsage
	// Available since: 4.0.0
	// Time complexity: O(N) where N is the number of samples.
	// ACL categories: @read @slow
	MemoryUsage(ctx context.Context, key string, samples ...int64) IntCmd

	// Save
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the total number of keys in all databases
	// ACL categories: @admin @slow @dangerous
	Save(ctx context.Context) StatusCmd

	// Shutdown
	// Available since: 1.0.0
	// Time complexity: O(N) when saving, where N is the total number of keys in all databases when saving data, otherwise O(1)
	// ACL categories: @admin @slow @dangerous
	Shutdown(ctx context.Context) StatusCmd

	// ShutdownSave
	// Available since: 1.0.0
	// Time complexity: O(N) when saving, where N is the total number of keys in all databases when saving data, otherwise O(1)
	// ACL categories: @admin @slow @dangerous
	ShutdownSave(ctx context.Context) StatusCmd

	// ShutdownNoSave
	// Available since: 1.0.0
	// Time complexity: O(N) when saving, where N is the total number of keys in all databases when saving data, otherwise O(1)
	// ACL categories: @admin @slow @dangerous
	ShutdownNoSave(ctx context.Context) StatusCmd

	// Time
	// Available since: 2.6.0
	// Time complexity: O(1)
	// ACL categories: @fast
	Time(ctx context.Context) TimeCmd

	// DebugObject
	// Available since: 1.0.0
	// Time complexity: Depends on subcommand.
	// ACL categories: @admin @slow @dangerous
	DebugObject(ctx context.Context, key string) StringCmd
}

func (c *client) ACLDryRun(ctx context.Context, username string, command ...any) StringCmd {
	ctx = c.handler.before(ctx, CommandACLDryRun)
	r := c.adapter.ACLDryRun(ctx, username, command...)
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
	r := c.adapter.CommandList(ctx, filter)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) CommandGetKeys(ctx context.Context, commands ...any) StringSliceCmd {
	ctx = c.handler.before(ctx, CommandCommandGetKeys)
	r := c.adapter.CommandGetKeys(ctx, commands...)
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
	r := c.adapter.Info(ctx, section...)
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
	r := c.adapter.DebugObject(ctx, key)
	c.handler.after(ctx, r.Err())
	return r
}
