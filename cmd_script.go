package redisson

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"strings"
)

type Scripter interface {
	// Hash
	// Return SHA1 digest of script.
	Hash() string

	// Load
	// Available since: 2.6.0
	// Time complexity: O(N) with N being the length in bytes of the script body.
	// ACL categories: @slow @scripting
	// Load a script into the scripts cache, without executing it. After the specified command is loaded into the script cache it will be callable using EVALSHA with the correct SHA1 digest of the script, exactly like after the first successful invocation of EVAL.
	// The script is guaranteed to stay in the script cache forever (unless SCRIPT FLUSH is called).
	// The command works in the same way even if the script was already present in the script cache.
	// For more information about EVAL scripts please refer to Introduction to Eval Scripts.
	// Return:
	//	Bulk string reply This command returns the SHA1 digest of the script added into the script cache.
	Load(ctx context.Context) StringCmd

	// Exists
	// Available since: 2.6.0
	// Time complexity: O(N) with N being the number of scripts to check (so checking a single script is an O(1) operation).
	// ACL categories: @slow @scripting
	// Returns information about the existence of the scripts in the script cache.
	// This command accepts one or more SHA1 digests and returns a list of ones or zeros to signal if the scripts are already defined or not inside the script cache. This can be useful before a pipelining operation to ensure that scripts are loaded (and if not, to load them using SCRIPT LOAD) so that the pipelining operation can be performed solely using EVALSHA instead of EVAL to save bandwidth.
	// For more information about EVAL scripts please refer to Introduction to Eval Scripts.
	// Return:
	//	Array reply The command returns an array of integers that correspond to the specified SHA1 digest arguments. For every corresponding SHA1 digest of a script that actually exists in the script cache, a 1 is returned, otherwise 0 is returned.
	Exists(ctx context.Context) BoolSliceCmd

	// Eval
	// Available since: 2.6.0
	// Time complexity: Depends on the script that is executed.
	// ACL categories: @slow @scripting
	// Invoke the execution of a server-side Lua script.
	// The first argument is the script's source code. Scripts are written in Lua and executed by the embedded Lua 5.1 interpreter in Redis.
	// The second argument is the number of input key name arguments, followed by all the keys accessed by the script. These names of input keys are available to the script as the KEYS global runtime variable Any additional input arguments should not represent names of keys.
	// Important: to ensure the correct execution of scripts, both in standalone and clustered deployments, all names of keys that a script accesses must be explicitly provided as input key arguments. The script should only access keys whose names are given as input arguments. Scripts should never access keys with programmatically-generated names or based on the contents of data structures stored in the database.
	// Please refer to the Redis Programmability and Introduction to Eval Scripts for more information about Lua scripts.
	Eval(ctx context.Context, keys []string, args ...interface{}) Cmd

	// EvalSha
	// Available since: 2.6.0
	// Time complexity: Depends on the script that is executed.
	// ACL categories: @slow @scripting
	// Evaluate a script from the server's cache by its SHA1 digest.
	// The server caches scripts by using the SCRIPT LOAD command. The command is otherwise identical to EVAL.
	// Please refer to the Redis Programmability and Introduction to Eval Scripts for more information about Lua scripts.
	EvalSha(ctx context.Context, keys []string, args ...interface{}) Cmd

	// Run optimistically uses EVALSHA to run the script. If script does not exist
	// it is retried using EVAL.
	Run(ctx context.Context, keys []string, args ...interface{}) Cmd
}

type ScriptCmdable interface {
	CreateScript(src string) Scripter

	// Eval
	// Available since: 2.6.0
	// Time complexity: Depends on the script that is executed.
	// ACL categories: @slow @scripting
	// Invoke the execution of a server-side Lua script.
	// The first argument is the script's source code. Scripts are written in Lua and executed by the embedded Lua 5.1 interpreter in Redis.
	// The second argument is the number of input key name arguments, followed by all the keys accessed by the script. These names of input keys are available to the script as the KEYS global runtime variable Any additional input arguments should not represent names of keys.
	// Important: to ensure the correct execution of scripts, both in standalone and clustered deployments, all names of keys that a script accesses must be explicitly provided as input key arguments. The script should only access keys whose names are given as input arguments. Scripts should never access keys with programmatically-generated names or based on the contents of data structures stored in the database.
	// Please refer to the Redis Programmability and Introduction to Eval Scripts for more information about Lua scripts.
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) Cmd

	// EvalSha
	// Available since: 2.6.0
	// Time complexity: Depends on the script that is executed.
	// ACL categories: @slow @scripting
	// Evaluate a script from the server's cache by its SHA1 digest.
	// The server caches scripts by using the SCRIPT LOAD command. The command is otherwise identical to EVAL.
	// Please refer to the Redis Programmability and Introduction to Eval Scripts for more information about Lua scripts.
	EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) Cmd

	// ScriptExists
	// Available since: 2.6.0
	// Time complexity: O(N) with N being the number of scripts to check (so checking a single script is an O(1) operation).
	// ACL categories: @slow @scripting
	// Returns information about the existence of the scripts in the script cache.
	// This command accepts one or more SHA1 digests and returns a list of ones or zeros to signal if the scripts are already defined or not inside the script cache. This can be useful before a pipelining operation to ensure that scripts are loaded (and if not, to load them using SCRIPT LOAD) so that the pipelining operation can be performed solely using EVALSHA instead of EVAL to save bandwidth.
	// For more information about EVAL scripts please refer to Introduction to Eval Scripts.
	// Return:
	//	Array reply The command returns an array of integers that correspond to the specified SHA1 digest arguments. For every corresponding SHA1 digest of a script that actually exists in the script cache, a 1 is returned, otherwise 0 is returned.
	ScriptExists(ctx context.Context, hashes ...string) BoolSliceCmd

	// ScriptFlush
	// Available since: 2.6.0
	// Time complexity: O(N) with N being the number of scripts in cache
	// ACL categories: @slow @scripting
	// Flush the Lua scripts cache.
	// By default, SCRIPT FLUSH will synchronously flush the cache. Starting with Redis 6.2, setting the lazyfree-lazy-user-flush configuration directive to "yes" changes the default flush mode to asynchronous.
	// It is possible to use one of the following modifiers to dictate the flushing mode explicitly:
	//	ASYNC: flushes the cache asynchronously
	//	SYNC: flushes the cache synchronously
	// For more information about EVAL scripts please refer to Introduction to Eval Scripts.
	// Return:
	//	Simple string reply
	ScriptFlush(ctx context.Context) StatusCmd

	// ScriptKill
	// Available since: 2.6.0
	// Time complexity: O(1)
	// ACL categories: @slow @scripting
	// Kills the currently executing EVAL script, assuming no write operation was yet performed by the script.
	// This command is mainly useful to kill a script that is running for too much time(for instance, because it entered an infinite loop because of a bug). The script will be killed, and the client currently blocked into EVAL will see the command returning with an error.
	// If the script has already performed write operations, it can not be killed in this way because it would violate Lua's script atomicity contract. In such a case, only SHUTDOWN NOSAVE can kill the script, killing the Redis process in a hard way and preventing it from persisting with half-written information.
	// For more information about EVAL scripts please refer to Introduction to Eval Scripts.
	// Return:
	//	Simple string reply
	ScriptKill(ctx context.Context) StatusCmd

	// ScriptLoad
	// Available since: 2.6.0
	// Time complexity: O(N) with N being the length in bytes of the script body.
	// ACL categories: @slow @scripting
	// Load a script into the scripts cache, without executing it. After the specified command is loaded into the script cache it will be callable using EVALSHA with the correct SHA1 digest of the script, exactly like after the first successful invocation of EVAL.
	// The script is guaranteed to stay in the script cache forever (unless SCRIPT FLUSH is called).
	// The command works in the same way even if the script was already present in the script cache.
	// For more information about EVAL scripts please refer to Introduction to Eval Scripts.
	// Return:
	//	Bulk string reply This command returns the SHA1 digest of the script added into the script cache.
	ScriptLoad(ctx context.Context, script string) StringCmd
}

func (c *client) CreateScript(src string) Scripter { return newScript(c, src) }

func (c *client) Eval(ctx context.Context, script string, keys []string, args ...interface{}) Cmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandEval, func() []string { return keys })
	r := c.adapter.Eval(ctx, script, keys, args...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) Cmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandEvalSha, func() []string { return keys })
	r := c.adapter.EvalSha(ctx, sha1, keys, args...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ScriptExists(ctx context.Context, hashes ...string) BoolSliceCmd {
	ctx = c.handler.before(ctx, CommandScriptExists)
	r := c.adapter.ScriptExists(ctx, hashes...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ScriptFlush(ctx context.Context) StatusCmd {
	ctx = c.handler.before(ctx, CommandScriptFlush)
	r := c.adapter.ScriptFlush(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ScriptKill(ctx context.Context) StatusCmd {
	ctx = c.handler.before(ctx, CommandScriptKill)
	r := c.adapter.ScriptKill(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) ScriptLoad(ctx context.Context, script string) StringCmd {
	ctx = c.handler.before(ctx, CommandScriptLoad)
	r := c.adapter.ScriptLoad(ctx, script)
	c.handler.after(ctx, r.Err())
	return r
}

type script struct {
	ScriptCmdable
	src, hash string
}

func newScript(c ScriptCmdable, src string) Scripter {
	h := sha1.New()
	_, _ = io.WriteString(h, src)
	return &script{
		ScriptCmdable: c,
		src:           src,
		hash:          hex.EncodeToString(h.Sum(nil)),
	}
}

func (s *script) Hash() string                            { return s.hash }
func (s *script) Load(ctx context.Context) StringCmd      { return s.ScriptLoad(ctx, s.src) }
func (s *script) Exists(ctx context.Context) BoolSliceCmd { return s.ScriptExists(ctx, s.hash) }
func (s *script) Eval(ctx context.Context, keys []string, args ...interface{}) Cmd {
	return s.ScriptCmdable.Eval(ctx, s.src, keys, args...)
}

func (s *script) EvalSha(ctx context.Context, keys []string, args ...interface{}) Cmd {
	return s.ScriptCmdable.EvalSha(ctx, s.hash, keys, args...)
}

func (s *script) Run(ctx context.Context, keys []string, args ...interface{}) Cmd {
	r := s.EvalSha(ctx, keys, args...)
	if err := r.Err(); err != nil && strings.HasPrefix(err.Error(), "NOSCRIPT ") {
		return s.Eval(ctx, keys, args...)
	}
	return r
}
