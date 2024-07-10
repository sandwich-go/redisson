package redisson

import (
	"context"
)

type ServerCmdable interface {
	// BgRewriteAOF
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @admin @slow @dangerous
	// Instruct Redis to start an Append Only File rewrite process. The rewrite will create a small optimized version of the current Append Only File.
	// If BGREWRITEAOF fails, no data gets lost as the old AOF will be untouched.
	// The rewrite will be only triggered by Redis if there is not already a background process doing persistence.
	// Specifically:
	//	If a Redis child is creating a snapshot on disk, the AOF rewrite is scheduled but not started until the saving child producing the RDB file terminates. In this case the BGREWRITEAOF will still return a positive status reply, but with an appropriate message. You can check if an AOF rewrite is scheduled looking at the INFO command as of Redis 2.6 or successive versions.
	//	If an AOF rewrite is already in progress the command returns an error and no AOF rewrite will be scheduled for a later time.
	//	If the AOF rewrite could start, but the attempt at starting it fails (for instance because of an error in creating the child process), an error is returned to the caller.
	// Since Redis 2.4 the AOF rewrite is automatically triggered by Redis, however the BGREWRITEAOF command can be used to trigger a rewrite at any time.
	// Please refer to the persistence documentation for detailed information.
	// Return:
	// Simple string reply: A simple string reply indicating that the rewriting started or is about to start ASAP, when the call is executed with success.
	// The command may reply with an error in certain cases, as documented above.
	BgRewriteAOF(ctx context.Context) StatusCmd

	// BgSave
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @admin @slow @dangerous
	// Save the DB in background.
	// Normally the OK code is immediately returned. Redis forks, the parent continues to serve the clients, the child saves the DB on disk then exits.
	// An error is returned if there is already a background save running or if there is another non-background-save process running, specifically an in-progress AOF rewrite.
	// If BGSAVE SCHEDULE is used, the command will immediately return OK when an AOF rewrite is in progress and schedule the background save to run at the next opportunity.
	// A client may be able to check if the operation succeeded using the LASTSAVE command.
	// Please refer to the persistence documentation for detailed information.
	// Return:
	// Simple string reply: Background saving started if BGSAVE started correctly or Background saving scheduled when used with the SCHEDULE subcommand.
	BgSave(ctx context.Context) StatusCmd

	// Command
	// Available since: 2.8.13
	// Time complexity: O(N) where N is the total number of Redis commands
	// ACL categories: @slow @connection
	// Return an array with details about every Redis command.
	// The COMMAND command is introspective. Its reply describes all commands that the server can process. Redis clients can call it to obtain the server's runtime capabilities during the handshake.
	// COMMAND also has several subcommands. Please refer to its subcommands for further details.
	// See https://redis.io/commands/command/
	Command(ctx context.Context) CommandsInfoCmd

	// ConfigGet
	// Available since: 2.0.0
	// Time complexity: O(N) when N is the number of configuration parameters provided
	// ACL categories: @admin @slow @dangerous
	// The CONFIG GET command is used to read the configuration parameters of a running Redis server. Not all the configuration parameters are supported in Redis 2.4, while Redis 2.6 can read the whole configuration of a server using this command.
	// The symmetric command used to alter the configuration at run time is CONFIG SET.
	// CONFIG GET takes multiple arguments, which are glob-style patterns. Any configuration parameter matching any of the patterns are reported as a list of key-value pairs. Example:
	//	redis> config get *max-*-entries* maxmemory
	// 	1) "maxmemory"
	// 	2) "0"
	// 	3) "hash-max-listpack-entries"
	// 	4) "512"
	// 	5) "hash-max-ziplist-entries"
	// 	6) "512"
	// 	7) "set-max-intset-entries"
	// 	8) "512"
	// 	9) "zset-max-listpack-entries"
	//	10) "128"
	//	11) "zset-max-ziplist-entries"
	//	12) "128"
	// You can obtain a list of all the supported configuration parameters by typing CONFIG GET * in an open redis-cli prompt.
	// All the supported parameters have the same meaning of the equivalent configuration parameter used in the redis.conf file:
	// Note that you should look at the redis.conf file relevant to the version you're working with as configuration options might change between versions. The link above is to the latest development version.
	// Return:
	//	The return type of the command is a Array reply.
	ConfigGet(ctx context.Context, parameter string) SliceCmd

	// ConfigResetStat
	// Available since: 2.0.0
	// Time complexity: O(1)
	// ACL categories: @admin @slow @dangerous
	// Resets the statistics reported by Redis using the INFO command.
	// These are the counters that are reset:
	// 	Keyspace hits
	//	Keyspace misses
	//	Number of commands processed
	//	Number of connections received
	//	Number of expired keys
	//	Number of rejected connections
	//	Latest fork(2) time
	//	The aof_delayed_fsync counter
	// Return:
	// Simple string reply: always OK.
	ConfigResetStat(ctx context.Context) StatusCmd

	// ConfigRewrite
	// Available since: 2.8.0
	// Time complexity: O(1)
	// ACL categories: @admin @slow @dangerous
	// The CONFIG REWRITE command rewrites the redis.conf file the server was started with, applying the minimal changes needed to make it reflect the configuration currently used by the server, which may be different compared to the original one because of the use of the CONFIG SET command.
	// The rewrite is performed in a very conservative way:
	//	Comments and the overall structure of the original redis.conf are preserved as much as possible.
	//	If an option already exists in the old redis.conf file, it will be rewritten at the same position (line number).
	//	If an option was not already present, but it is set to its default value, it is not added by the rewrite process.
	//	If an option was not already present, but it is set to a non-default value, it is appended at the end of the file.
	//	Non used lines are blanked. For instance if you used to have multiple save directives, but the current configuration has fewer or none as you disabled RDB persistence, all the lines will be blanked.
	// CONFIG REWRITE is also able to rewrite the configuration file from scratch if the original one no longer exists for some reason. However if the server was started without a configuration file at all, the CONFIG REWRITE will just return an error.
	// Atomic rewrite process
	// In order to make sure the redis.conf file is always consistent, that is, on errors or crashes you always end with the old file, or the new one, the rewrite is performed with a single write(2) call that has enough content to be at least as big as the old file. Sometimes additional padding in the form of comments is added in order to make sure the resulting file is big enough, and later the file gets truncated to remove the padding at the end.
	// Return:
	// Simple string reply: OK when the configuration was rewritten properly. Otherwise an error is returned.
	ConfigRewrite(ctx context.Context) StatusCmd

	// ConfigSet
	// Available since: 2.0.0
	// Time complexity: O(N) when N is the number of configuration parameters provided
	// ACL categories: @admin @slow @dangerous
	// The CONFIG SET command is used in order to reconfigure the server at run time without the need to restart Redis. You can change both trivial parameters or switch from one to another persistence option using this command.
	// The list of configuration parameters supported by CONFIG SET can be obtained issuing a CONFIG GET * command, that is the symmetrical command used to obtain information about the configuration of a running Redis instance.
	// All the configuration parameters set using CONFIG SET are immediately loaded by Redis and will take effect starting with the next command executed.
	// All the supported parameters have the same meaning of the equivalent configuration parameter used in the redis.conf file.
	// Note that you should look at the redis.conf file relevant to the version you're working with as configuration options might change between versions. The link above is to the latest development version.
	// It is possible to switch persistence from RDB snapshotting to append-only file (and the other way around) using the CONFIG SET command. For more information about how to do that please check the persistence page.
	// In general what you should know is that setting the appendonly parameter to yes will start a background process to save the initial append-only file (obtained from the in memory data set), and will append all the subsequent commands on the append-only file, thus obtaining exactly the same effect of a Redis server that started with AOF turned on since the start.
	// You can have both the AOF enabled with RDB snapshotting if you want, the two options are not mutually exclusive.
	// Return:
	// Simple string reply: OK when the configuration was set properly. Otherwise an error is returned.
	ConfigSet(ctx context.Context, parameter, value string) StatusCmd

	// DBSize
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @keyspace @read @fast
	// Return the number of keys in the currently-selected database.
	// Return:
	//	Integer reply
	DBSize(ctx context.Context) IntCmd

	// FlushAll
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the total number of keys in all databases
	// ACL categories: @keyspace @write @slow @dangerous
	// Delete all the keys of all the existing databases, not just the currently selected one. This command never fails.
	// By default, FLUSHALL will synchronously flush all the databases. Starting with Redis 6.2, setting the lazyfree-lazy-user-flush configuration directive to "yes" changes the default flush mode to asynchronous.
	// It is possible to use one of the following modifiers to dictate the flushing mode explicitly:
	//	ASYNC: flushes the databases asynchronously
	//	SYNC: flushes the databases synchronously
	// Note: an asynchronous FLUSHALL command only deletes keys that were present at the time the command was invoked. Keys created during an asynchronous flush will be unaffected.
	// Return:
	//	Simple string reply
	FlushAll(ctx context.Context) StatusCmd

	// FlushAllAsync
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the total number of keys in all databases
	// ACL categories: @keyspace @write @slow @dangerous
	// See FlushAll
	FlushAllAsync(ctx context.Context) StatusCmd

	// FlushDB
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the number of keys in the selected database
	// ACL categories: @keyspace @write @slow @dangerous
	// Delete all the keys of the currently selected DB. This command never fails.
	// By default, FLUSHDB will synchronously flush all keys from the database. Starting with Redis 6.2, setting the lazyfree-lazy-user-flush configuration directive to "yes" changes the default flush mode to asynchronous.
	// It is possible to use one of the following modifiers to dictate the flushing mode explicitly:
	//	ASYNC: flushes the database asynchronously
	//	SYNC: flushes the database synchronously
	// Note: an asynchronous FLUSHDB command only deletes keys that were present at the time the command was invoked. Keys created during an asynchronous flush will be unaffected.
	// Return:
	// 	Simple string reply
	FlushDB(ctx context.Context) StatusCmd

	// FlushDBAsync
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the number of keys in the selected database
	// ACL categories: @keyspace @write @slow @dangerous
	// See FlushDB
	FlushDBAsync(ctx context.Context) StatusCmd

	// Info
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @slow @dangerous
	// The INFO command returns information and statistics about the server in a format that is simple to parse by computers and easy to read by humans.
	// The optional parameter can be used to select a specific section of information:
	//	server: General information about the Redis server
	//	clients: Client connections section
	//	memory: Memory consumption related information
	//	persistence: RDB and AOF related information
	//	stats: General statistics
	//	replication: Master/replica replication information
	//	cpu: CPU consumption statistics
	//	commandstats: Redis command statistics
	//	latencystats: Redis command latency percentile distribution statistics
	//	cluster: Redis Cluster section
	//	modules: Modules section
	//	keyspace: Database related statistics
	//	modules: Module related sections
	//	errorstats: Redis error statistics
	// It can also take the following values:
	//	all: Return all sections (excluding module generated ones)
	//	default: Return only the default set of sections
	//	everything: Includes all and modules
	// When no parameter is provided, the default option is assumed.
	// Return:
	//	Bulk string reply: as a collection of text lines.
	//	Lines can contain a section name (starting with a # character) or a property. All the properties are in the form of field:value terminated by \r\n.
	Info(ctx context.Context, section ...string) StringCmd

	// LastSave
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @admin @fast @dangerous
	// Return the UNIX TIME of the last DB save executed with success. A client may check if a BGSAVE command succeeded reading the LASTSAVE value, then issuing a BGSAVE command and checking at regular intervals every N seconds if LASTSAVE changed.
	// Return:
	//	Integer reply: an UNIX time stamp.
	LastSave(ctx context.Context) IntCmd

	// MemoryUsage
	// Available since: 4.0.0
	// Time complexity: O(N) where N is the number of samples.
	// ACL categories: @read @slow
	// The MEMORY USAGE command reports the number of bytes that a key and its value require to be stored in RAM.
	// The reported usage is the total of memory allocations for data and administrative overheads that a key its value require.
	// For nested data types, the optional SAMPLES option can be provided, where count is the number of sampled nested values. By default, this option is set to 5. To sample the all of the nested values, use SAMPLES 0.
	// Return:
	// 	Integer reply: the memory usage in bytes, or nil when the key does not exist.
	MemoryUsage(ctx context.Context, key string, samples ...int) IntCmd

	// Save
	// Available since: 1.0.0
	// Time complexity: O(N) where N is the total number of keys in all databases
	// ACL categories: @admin @slow @dangerous
	// The SAVE commands performs a synchronous save of the dataset producing a point in time snapshot of all the data inside the Redis instance, in the form of an RDB file.
	// You almost never want to call SAVE in production environments where it will block all the other clients. Instead usually BGSAVE is used. However in case of issues preventing Redis to create the background saving child (for instance errors in the fork(2) system call), the SAVE command can be a good last resort to perform the dump of the latest dataset.
	// Please refer to the persistence documentation for detailed information.
	// Return:
	//	Simple string reply: The commands returns OK on success.
	Save(ctx context.Context) StatusCmd

	// Shutdown
	// Available since: 1.0.0
	// Time complexity: O(N) when saving, where N is the total number of keys in all databases when saving data, otherwise O(1)
	// ACL categories: @admin @slow @dangerous
	// The command behavior is the following:
	//	If there are any replicas lagging behind in replication:
	//		Pause clients attempting to write by performing a CLIENT PAUSE with the WRITE option.
	//		Wait up to the configured shutdown-timeout (default 10 seconds) for replicas to catch up the replication offset.
	//	Stop all the clients.
	//	Perform a blocking SAVE if at least one save point is configured.
	//	Flush the Append Only File if AOF is enabled.
	//	Quit the server.
	// If persistence is enabled this commands makes sure that Redis is switched off without any data loss.
	// Note: A Redis instance that is configured for not persisting on disk (no AOF configured, nor "save" directive) will not dump the RDB file on SHUTDOWN, as usually you don't want Redis instances used only for caching to block on when shutting down.
	// Also note: If Redis receives one of the signals SIGTERM and SIGINT, the same shutdown sequence is performed. See also Signal Handling.
	// Return:
	//	Simple string reply: OK if ABORT was specified and shutdown was aborted. On successful shutdown, nothing is returned since the server quits and the connection is closed. On failure, an error is returned.
	Shutdown(ctx context.Context) StatusCmd

	// ShutdownSave
	// Available since: 1.0.0
	// Time complexity: O(N) when saving, where N is the total number of keys in all databases when saving data, otherwise O(1)
	// ACL categories: @admin @slow @dangerous
	// See Shutdown
	ShutdownSave(ctx context.Context) StatusCmd

	// ShutdownNoSave
	// Available since: 1.0.0
	// Time complexity: O(N) when saving, where N is the total number of keys in all databases when saving data, otherwise O(1)
	// ACL categories: @admin @slow @dangerous
	// See Shutdown
	ShutdownNoSave(ctx context.Context) StatusCmd

	// SlaveOf
	// Available since: 1.0.0
	// Time complexity: O(1)
	// ACL categories: @admin @slow @dangerous
	// As of Redis version 5.0.0, this command is regarded as deprecated.
	// It can be replaced by REPLICAOF when migrating or writing new code.
	// A note about the word slave used in this man page and command name: starting with Redis version 5, if not for backward compatibility, the Redis project no longer uses the word slave. Please use the new command REPLICAOF. The command SLAVEOF will continue to work for backward compatibility.
	// The SLAVEOF command can change the replication settings of a replica on the fly. If a Redis server is already acting as replica, the command SLAVEOF NO ONE will turn off the replication, turning the Redis server into a MASTER. In the proper form SLAVEOF hostname port will make the server a replica of another server listening at the specified hostname and port.
	// If a server is already a replica of some master, SLAVEOF hostname port will stop the replication against the old server and start the synchronization against the new one, discarding the old dataset.
	// The form SLAVEOF NO ONE will stop replication, turning the server into a MASTER, but will not discard the replication. So, if the old master stops working, it is possible to turn the replica into a master and set the application to use this new master in read/write. Later when the other Redis server is fixed, it can be reconfigured to work as a replica.
	// Return
	//	Simple string reply
	SlaveOf(ctx context.Context, host, port string) StatusCmd

	// Time
	// Available since: 2.6.0
	// Time complexity: O(1)
	// ACL categories: @fast
	// The TIME command returns the current server time as a two items lists: a Unix timestamp and the amount of microseconds already elapsed in the current second. Basically the interface is very similar to the one of the gettimeofday system call.
	// Return:
	// Array reply, specifically:
	// A multi bulk reply containing two elements:
	//	unix time in seconds.
	//	microseconds.
	Time(ctx context.Context) TimeCmd

	// DebugObject
	// Available since: 1.0.0
	// Time complexity: Depends on subcommand.
	// ACL categories: @admin @slow @dangerous
	// The DEBUG command is an internal command. It is meant to be used for developing and testing Redis.
	DebugObject(ctx context.Context, key string) StringCmd
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
	r := newCommandsInfoCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Command().Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ConfigGet(ctx context.Context, parameter string) SliceCmd {
	ctx = c.handler.before(ctx, CommandConfigGet)
	r := newSliceCmdFromMapResult(c.cmd.Do(ctx, c.cmd.B().ConfigGet().Parameter(parameter).Build()))
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
		ctx = c.handler.before(ctx, CommandInfoMultiple)
	} else {
		ctx = c.handler.before(ctx, CommandInfos)
	}
	r := newStringCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Info().Section(section...).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) LastSave(ctx context.Context) IntCmd {
	ctx = c.handler.before(ctx, CommandLastSave)
	r := newIntCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Lastsave().Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) MemoryUsage(ctx context.Context, key string, samples ...int) IntCmd {
	ctx = c.handler.before(ctx, CommandMemoryUsage)
	r := c.memoryUsage(ctx, key, samples...)
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

func (c *client) SlaveOf(ctx context.Context, host, port string) StatusCmd {
	ctx = c.handler.before(ctx, CommandSlaveOf)
	r := newStatusCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Arbitrary(SLAVEOF).Args(host, port).Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) Time(ctx context.Context) TimeCmd {
	ctx = c.handler.before(ctx, CommandTime)
	r := newTimeCmdFromResult(c.cmd.Do(ctx, c.cmd.B().Time().Build()))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) DebugObject(ctx context.Context, key string) StringCmd {
	ctx = c.handler.before(ctx, CommandDebug)
	r := newStringCmdFromResult(c.cmd.Do(ctx, c.cmd.B().DebugObject().Key(key).Build()))
	c.handler.after(ctx, r.Err())
	return r
}
