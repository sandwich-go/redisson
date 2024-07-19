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

	// Eval
	// Available since: 2.6.0
	// Time complexity: Depends on the script that is executed.
	// ACL categories: @slow @scripting
	// RESP2 / RESP3 Reply:
	// 	- The return value depends on the script that was executed.
	Eval(ctx context.Context, keys []string, args ...any) Cmd

	// EvalRO
	// Available since: 7.0.0
	// Time complexity: Depends on the script that is executed.
	// ACL categories: @slow @scripting
	// RESP2 / RESP3 Reply:
	// 	- The return value depends on the script that was executed.
	EvalRO(ctx context.Context, keys []string, args ...any) Cmd

	// EvalSha
	// Available since: 2.6.0
	// Time complexity: Depends on the script that is executed.
	// ACL categories: @slow @scripting
	// RESP2 / RESP3 Reply:
	// 	- The return value depends on the script that was executed.
	EvalSha(ctx context.Context, keys []string, args ...any) Cmd

	// EvalShaRO
	// Available since: 7.0.0
	// Time complexity: Depends on the script that is executed.
	// ACL categories: @slow @scripting
	// RESP2 / RESP3 Reply:
	// 	- The return value depends on the script that was executed.
	EvalShaRO(ctx context.Context, keys []string, args ...any) Cmd

	// Run optimistically uses EVALSHA to run the script. If script does not exist
	// it is retried using EVAL.
	Run(ctx context.Context, keys []string, args ...any) Cmd

	// RunRO optimistically uses EVALSHA_RO to run the script. If script does not exist
	// it is retried using EVAL_RO.
	RunRO(ctx context.Context, keys []string, args ...any) Cmd

	// Load
	// Available since: 2.6.0
	// Time complexity: O(N) with N being the length in bytes of the script body.
	// ACL categories: @slow @scripting
	// RESP2 / RESP3 Reply:
	// 	- Bulk string reply: the SHA1 digest of the script added into the script cache.
	Load(ctx context.Context) StringCmd

	// Exists
	// Available since: 2.6.0
	// Time complexity: O(N) with N being the number of scripts to check (so checking a single script is an O(1) operation).
	// ACL categories: @slow @scripting
	Exists(ctx context.Context) BoolSliceCmd
}

type ScriptCmdable interface {
	CreateScript(src string) Scripter

	// Eval
	// Available since: 2.6.0
	// Time complexity: Depends on the script that is executed.
	// ACL categories: @slow @scripting
	// RESP2 / RESP3 Reply:
	// 	- The return value depends on the script that was executed.
	Eval(ctx context.Context, script string, keys []string, args ...any) Cmd

	// EvalRO
	// Available since: 7.0.0
	// Time complexity: Depends on the script that is executed.
	// ACL categories: @slow @scripting
	// RESP2 / RESP3 Reply:
	// 	- The return value depends on the script that was executed.
	EvalRO(ctx context.Context, script string, keys []string, args ...any) Cmd

	// EvalSha
	// Available since: 2.6.0
	// Time complexity: Depends on the script that is executed.
	// ACL categories: @slow @scripting
	// RESP2 / RESP3 Reply:
	// 	- The return value depends on the script that was executed.
	EvalSha(ctx context.Context, sha1 string, keys []string, args ...any) Cmd

	// EvalShaRO
	// Available since: 7.0.0
	// Time complexity: Depends on the script that is executed.
	// ACL categories: @slow @scripting
	// RESP2 / RESP3 Reply:
	// 	- The return value depends on the script that was executed.
	EvalShaRO(ctx context.Context, sha1 string, keys []string, args ...any) Cmd

	// FCall
	// Available since: 7.0.0
	// Time complexity: Depends on the script that is executed.
	// ACL categories: @slow @scripting
	// RESP2 / RESP3 Reply:
	// 	- The return value depends on the script that was executed.
	FCall(ctx context.Context, function string, keys []string, args ...any) Cmd

	// FCallRO
	// Available since: 7.0.0
	// Time complexity: Depends on the script that is executed.
	// ACL categories: @slow @scripting
	// RESP2 / RESP3 Reply:
	// 	- The return value depends on the script that was executed.
	FCallRO(ctx context.Context, function string, keys []string, args ...any) Cmd

	// FunctionDelete
	// Available since: 7.0.0
	// Time complexity: Depends on the script that is executed.
	// ACL categories: @write, @slow, @scripting
	// RESP2 / RESP3 Reply:
	// 	- Simple string reply: OK.
	FunctionDelete(ctx context.Context, libName string) StringCmd

	// FunctionDump
	// Available since: 7.0.0
	// Time complexity: Depends on the script that is executed.
	// ACL categories: @slow @scripting
	// RESP2 / RESP3 Reply:
	// 	- Bulk string reply: the serialized payload
	FunctionDump(ctx context.Context) StringCmd

	// FunctionFlush
	// Available since: 7.0.0
	// Time complexity: Depends on the script that is executed.
	// ACL categories: @write, @slow, @scripting
	// RESP2 / RESP3 Reply:
	// 	- Simple string reply: OK.
	FunctionFlush(ctx context.Context) StringCmd
	FunctionFlushAsync(ctx context.Context) StringCmd

	// FunctionKill
	// Available since: 7.0.0
	// Time complexity: Depends on the script that is executed.
	// ACL categories: @slow @scripting
	// RESP2 / RESP3 Reply:
	// 	- Simple string reply: OK.
	FunctionKill(ctx context.Context) StringCmd

	// FunctionList
	// Available since: 7.0.0
	// Time complexity: Depends on the script that is executed.
	// ACL categories: @slow @scripting
	// RESP2 / RESP3 Reply:
	// 	- Array reply: information about functions and libraries.
	FunctionList(ctx context.Context, q FunctionListQuery) FunctionListCmd

	// FunctionLoad
	// Available since: 7.0.0
	// Time complexity: Depends on the script that is executed.
	// ACL categories: @write, @slow, @scripting
	// RESP2 / RESP3 Reply:
	// 	- Bulk string reply: the library name that was loaded.
	FunctionLoad(ctx context.Context, code string) StringCmd
	FunctionLoadReplace(ctx context.Context, code string) StringCmd

	// FunctionRestore
	// Available since: 7.0.0
	// Time complexity: Depends on the script that is executed.
	// ACL categories: @write, @slow, @scripting
	// RESP2 / RESP3 Reply:
	// 	- Simple string reply: OK.
	FunctionRestore(ctx context.Context, libDump string) StringCmd

	// ScriptExists
	// Available since: 2.6.0
	// Time complexity: O(N) with N being the number of scripts to check (so checking a single script is an O(1) operation).
	// ACL categories: @slow @scripting
	// RESP2 / RESP3 Reply:
	// 	- Array reply: an array of integers that correspond to the specified SHA1 digest arguments.
	ScriptExists(ctx context.Context, hashes ...string) BoolSliceCmd

	// ScriptFlush
	// Available since: 2.6.0
	// Time complexity: O(N) with N being the number of scripts in cache
	// ACL categories: @slow @scripting
	// RESP2 / RESP3 Reply:
	// 	- Simple string reply: OK.
	// History:
	//	- Starting with Redis version 6.2.0: Added the ASYNC and SYNC flushing mode modifiers
	ScriptFlush(ctx context.Context) StatusCmd

	// ScriptKill
	// Available since: 2.6.0
	// Time complexity: O(1)
	// ACL categories: @slow @scripting
	// RESP2 / RESP3 Reply:
	// 	- Simple string reply: OK.
	ScriptKill(ctx context.Context) StatusCmd

	// ScriptLoad
	// Available since: 2.6.0
	// Time complexity: O(N) with N being the length in bytes of the script body.
	// ACL categories: @slow @scripting
	// RESP2 / RESP3 Reply:
	// 	- Bulk string reply: the SHA1 digest of the script added into the script cache.
	ScriptLoad(ctx context.Context, script string) StringCmd
}

func (c *client) CreateScript(src string) Scripter { return newScript(c, src) }

func (c *client) Eval(ctx context.Context, script string, keys []string, args ...any) Cmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandEval, func() []string { return keys })
	r := c.adapter.Eval(ctx, script, keys, args...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) EvalRO(ctx context.Context, script string, keys []string, args ...any) Cmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandEvalRO, func() []string { return keys })
	r := c.adapter.EvalRO(ctx, script, keys, args...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) EvalSha(ctx context.Context, sha1 string, keys []string, args ...any) Cmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandEvalSha, func() []string { return keys })
	r := c.adapter.EvalSha(ctx, sha1, keys, args...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) EvalShaRO(ctx context.Context, sha1 string, keys []string, args ...any) Cmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandEvalShaRO, func() []string { return keys })
	r := c.adapter.EvalShaRO(ctx, sha1, keys, args...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) FCall(ctx context.Context, function string, keys []string, args ...any) Cmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandFCall, func() []string { return keys })
	r := c.adapter.FCall(ctx, function, keys, args...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) FCallRO(ctx context.Context, function string, keys []string, args ...any) Cmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandFCallRO, func() []string { return keys })
	r := c.adapter.FCallRO(ctx, function, keys, args...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) FunctionDelete(ctx context.Context, libName string) StringCmd {
	ctx = c.handler.before(ctx, CommandFunctionDelete)
	r := c.adapter.FunctionDelete(ctx, libName)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) FunctionDump(ctx context.Context) StringCmd {
	ctx = c.handler.before(ctx, CommandFunctionDump)
	r := c.adapter.FunctionDump(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) FunctionFlush(ctx context.Context) StringCmd {
	ctx = c.handler.before(ctx, CommandFunctionFlush)
	r := c.adapter.FunctionFlush(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) FunctionFlushAsync(ctx context.Context) StringCmd {
	ctx = c.handler.before(ctx, CommandFunctionFlushAsync)
	r := c.adapter.FunctionFlushAsync(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) FunctionKill(ctx context.Context) StringCmd {
	ctx = c.handler.before(ctx, CommandFunctionKill)
	r := c.adapter.FunctionKill(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) FunctionList(ctx context.Context, q FunctionListQuery) FunctionListCmd {
	ctx = c.handler.before(ctx, CommandFunctionList)
	r := c.adapter.FunctionList(ctx, q)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) FunctionLoad(ctx context.Context, code string) StringCmd {
	ctx = c.handler.before(ctx, CommandFunctionLoad)
	r := c.adapter.FunctionLoad(ctx, code)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) FunctionLoadReplace(ctx context.Context, code string) StringCmd {
	ctx = c.handler.before(ctx, CommandFunctionLoadReplace)
	r := c.adapter.FunctionLoadReplace(ctx, code)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) FunctionRestore(ctx context.Context, libDump string) StringCmd {
	ctx = c.handler.before(ctx, CommandFunctionRestore)
	r := c.adapter.FunctionRestore(ctx, libDump)
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
func (s *script) Eval(ctx context.Context, keys []string, args ...any) Cmd {
	return s.ScriptCmdable.Eval(ctx, s.src, keys, args...)
}
func (s *script) EvalRO(ctx context.Context, keys []string, args ...any) Cmd {
	return s.ScriptCmdable.EvalRO(ctx, s.src, keys, args...)
}

func (s *script) EvalSha(ctx context.Context, keys []string, args ...any) Cmd {
	return s.ScriptCmdable.EvalSha(ctx, s.hash, keys, args...)
}
func (s *script) EvalShaRO(ctx context.Context, keys []string, args ...any) Cmd {
	return s.ScriptCmdable.EvalShaRO(ctx, s.src, keys, args...)
}

func (s *script) Run(ctx context.Context, keys []string, args ...any) Cmd {
	r := s.EvalSha(ctx, keys, args...)
	if err := r.Err(); err != nil && strings.HasPrefix(err.Error(), "NOSCRIPT ") {
		return s.Eval(ctx, keys, args...)
	}
	return r
}

func (s *script) RunRO(ctx context.Context, keys []string, args ...any) Cmd {
	r := s.EvalShaRO(ctx, keys, args...)
	if err := r.Err(); err != nil && strings.HasPrefix(err.Error(), "NOSCRIPT ") {
		return s.EvalRO(ctx, keys, args...)
	}
	return r
}
