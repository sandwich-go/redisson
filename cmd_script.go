package redisson

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"io"
)

type Scripter interface {
	SetName(string)

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
	// RESP2 / RESP3 Reply:
	// 	- Array reply: an array of integers that correspond to the specified SHA1 digest arguments.
	Exists(ctx context.Context) BoolSliceCmd
}

type ScriptCmdable interface {
	CreateScript(src string) Scripter
	CreateScriptWithName(name, src string) Scripter

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
	// Time complexity: Depends on the function that is executed.
	// ACL categories: @slow @scripting
	// RESP2 / RESP3 Reply:
	// 	- The return value depends on the function that was executed.
	FCall(ctx context.Context, function string, keys []string, args ...any) Cmd

	// FCallRO
	// Available since: 7.0.0
	// Time complexity: Depends on the function that is executed.
	// ACL categories: @slow @scripting
	// RESP2 / RESP3 Reply:
	// 	- The return value depends on the function that was executed.
	FCallRO(ctx context.Context, function string, keys []string, args ...any) Cmd

	// FunctionDelete
	// Available since: 7.0.0
	// Time complexity: O(1)
	// ACL categories: @write, @slow, @scripting
	// RESP2 / RESP3 Reply:
	// 	- Simple string reply: OK.
	FunctionDelete(ctx context.Context, libName string) StringCmd

	// FunctionDump
	// Available since: 7.0.0
	// Time complexity: O(N) where N is the number of functions
	// ACL categories: @slow @scripting
	// RESP2 / RESP3 Reply:
	// 	- Bulk string reply: the serialized payload
	FunctionDump(ctx context.Context) StringCmd

	// FunctionFlush
	// Available since: 7.0.0
	// Time complexity: O(N) where N is the number of functions deleted
	// ACL categories: @write, @slow, @scripting
	// RESP2 / RESP3 Reply:
	// 	- Simple string reply: OK.
	FunctionFlush(ctx context.Context) StringCmd
	FunctionFlushAsync(ctx context.Context) StringCmd

	// FunctionKill
	// Available since: 7.0.0
	// Time complexity: O(1)
	// ACL categories: @slow @scripting
	// RESP2 / RESP3 Reply:
	// 	- Simple string reply: OK.
	FunctionKill(ctx context.Context) StringCmd

	// FunctionList
	// Available since: 7.0.0
	// Time complexity: O(N) where N is the number of functions
	// ACL categories: @slow @scripting
	// RESP2 / RESP3 Reply:
	// 	- Array reply: information about functions and libraries.
	FunctionList(ctx context.Context, q FunctionListQuery) FunctionListCmd

	// FunctionLoad
	// Available since: 7.0.0
	// Time complexity: O(1) (considering compilation time is redundant)
	// ACL categories: @write, @slow, @scripting
	// RESP2 / RESP3 Reply:
	// 	- Bulk string reply: the library name that was loaded.
	FunctionLoad(ctx context.Context, code string) StringCmd
	FunctionLoadReplace(ctx context.Context, code string) StringCmd

	// FunctionRestore
	// Available since: 7.0.0
	// Time complexity: O(N) where N is the number of functions on the payload
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

func (c *client) CreateScriptWithName(name, src string) Scripter {
	s := c.CreateScript(src)
	s.SetName(name)
	return s
}
func (c *client) CreateScript(src string) Scripter { return newScript(c, src) }

func (c *client) eval(ctx context.Context, name, script string, keys []string, args ...any) Cmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandEval, func() []string { return keys })
	r := c.adapter.Eval(ctx, script, keys, args...)
	c.handler.after(WithSubCommandName(ctx, name), r.Err())
	return r
}

func (c *client) evalRO(ctx context.Context, name, script string, keys []string, args ...any) Cmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandEvalRO, func() []string { return keys })
	r := c.adapter.EvalRO(ctx, script, keys, args...)
	c.handler.after(WithSubCommandName(ctx, name), r.Err())
	return r
}

func (c *client) evalSha(ctx context.Context, name, sha1 string, keys []string, args ...any) Cmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandEvalSha, func() []string { return keys })
	r := c.adapter.EvalSha(ctx, sha1, keys, args...)
	c.handler.after(WithSubCommandName(ctx, name), r.Err())
	return r
}

func (c *client) evalShaRO(ctx context.Context, name, sha1 string, keys []string, args ...any) Cmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandEvalShaRO, func() []string { return keys })
	r := c.adapter.EvalShaRO(ctx, sha1, keys, args...)
	c.handler.after(WithSubCommandName(ctx, name), r.Err())
	return r
}

func (c *client) fCall(ctx context.Context, function string, keys []string, args ...any) Cmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandFCall, func() []string { return keys })
	r := c.adapter.FCall(ctx, function, keys, args...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) fCallRO(ctx context.Context, function string, keys []string, args ...any) Cmd {
	ctx = c.handler.beforeWithKeys(ctx, CommandFCallRO, func() []string { return keys })
	r := c.adapter.FCallRO(ctx, function, keys, args...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) functionDelete(ctx context.Context, libName string) StringCmd {
	ctx = c.handler.before(ctx, CommandFunctionDelete)
	r := wrapStringCmd(c.adapter.FunctionDelete(ctx, libName))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) functionDump(ctx context.Context) StringCmd {
	ctx = c.handler.before(ctx, CommandFunctionDump)
	r := wrapStringCmd(c.adapter.FunctionDump(ctx))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) functionFlush(ctx context.Context) StringCmd {
	ctx = c.handler.before(ctx, CommandFunctionFlush)
	r := wrapStringCmd(c.adapter.FunctionFlush(ctx))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) functionFlushAsync(ctx context.Context) StringCmd {
	ctx = c.handler.before(ctx, CommandFunctionFlushAsync)
	r := wrapStringCmd(c.adapter.FunctionFlushAsync(ctx))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) functionKill(ctx context.Context) StringCmd {
	ctx = c.handler.before(ctx, CommandFunctionKill)
	r := wrapStringCmd(c.adapter.FunctionKill(ctx))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) functionList(ctx context.Context, q FunctionListQuery) FunctionListCmd {
	ctx = c.handler.before(ctx, CommandFunctionList)
	r := c.adapter.FunctionList(ctx, q)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) functionLoad(ctx context.Context, code string) StringCmd {
	ctx = c.handler.before(ctx, CommandFunctionLoad)
	r := wrapStringCmd(c.adapter.FunctionLoad(ctx, code))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) functionLoadReplace(ctx context.Context, code string) StringCmd {
	ctx = c.handler.before(ctx, CommandFunctionLoadReplace)
	r := wrapStringCmd(c.adapter.FunctionLoadReplace(ctx, code))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) functionRestore(ctx context.Context, libDump string) StringCmd {
	ctx = c.handler.before(ctx, CommandFunctionRestore)
	r := wrapStringCmd(c.adapter.FunctionRestore(ctx, libDump))
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) scriptExists(ctx context.Context, hashes ...string) BoolSliceCmd {
	ctx = c.handler.before(ctx, CommandScriptExists)
	r := c.adapter.ScriptExists(ctx, hashes...)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) scriptFlush(ctx context.Context) StatusCmd {
	ctx = c.handler.before(ctx, CommandScriptFlush)
	r := c.adapter.ScriptFlush(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) scriptKill(ctx context.Context) StatusCmd {
	ctx = c.handler.before(ctx, CommandScriptKill)
	r := c.adapter.ScriptKill(ctx)
	c.handler.after(ctx, r.Err())
	return r
}

func (c *client) scriptLoad(ctx context.Context, name, script string) StringCmd {
	ctx = c.handler.before(ctx, CommandScriptLoad)
	r := wrapStringCmd(c.adapter.ScriptLoad(ctx, script))
	c.handler.after(WithSubCommandName(ctx, name), r.Err())
	return r
}

func (c *client) Eval(ctx context.Context, script string, keys []string, args ...any) Cmd {
	return c.eval(ctx, "", script, keys, args...)
}

func (c *client) EvalRO(ctx context.Context, script string, keys []string, args ...any) Cmd {
	return c.evalRO(ctx, "", script, keys, args...)
}

func (c *client) EvalSha(ctx context.Context, sha1 string, keys []string, args ...any) Cmd {
	return c.evalSha(ctx, "", sha1, keys, args...)
}

func (c *client) EvalShaRO(ctx context.Context, sha1 string, keys []string, args ...any) Cmd {
	return c.evalShaRO(ctx, "", sha1, keys, args...)
}

func (c *client) FCall(ctx context.Context, function string, keys []string, args ...any) Cmd {
	return c.fCall(ctx, function, keys, args...)
}

func (c *client) FCallRO(ctx context.Context, function string, keys []string, args ...any) Cmd {
	return c.fCallRO(ctx, function, keys, args...)
}

func (c *client) FunctionDelete(ctx context.Context, libName string) StringCmd {
	return c.functionDelete(ctx, libName)
}

func (c *client) FunctionDump(ctx context.Context) StringCmd       { return c.functionDump(ctx) }
func (c *client) FunctionFlush(ctx context.Context) StringCmd      { return c.functionFlush(ctx) }
func (c *client) FunctionFlushAsync(ctx context.Context) StringCmd { return c.functionFlushAsync(ctx) }
func (c *client) FunctionKill(ctx context.Context) StringCmd       { return c.functionKill(ctx) }

func (c *client) FunctionList(ctx context.Context, q FunctionListQuery) FunctionListCmd {
	return c.functionList(ctx, q)
}

func (c *client) FunctionLoad(ctx context.Context, code string) StringCmd {
	return c.functionLoad(ctx, code)
}

func (c *client) FunctionLoadReplace(ctx context.Context, code string) StringCmd {
	return c.functionLoadReplace(ctx, code)
}

func (c *client) FunctionRestore(ctx context.Context, libDump string) StringCmd {
	return c.functionRestore(ctx, libDump)
}

func (c *client) ScriptExists(ctx context.Context, hashes ...string) BoolSliceCmd {
	return c.scriptExists(ctx, hashes...)
}

func (c *client) ScriptFlush(ctx context.Context) StatusCmd { return c.scriptFlush(ctx) }
func (c *client) ScriptKill(ctx context.Context) StatusCmd  { return c.scriptKill(ctx) }

func (c *client) ScriptLoad(ctx context.Context, script string) StringCmd {
	return c.scriptLoad(ctx, "", script)
}

type script struct {
	*client
	name      string
	src, hash string
}

func newScript(c *client, src string) Scripter {
	h := sha1.New()
	_, _ = io.WriteString(h, src)
	return &script{
		client: c,
		src:    src,
		hash:   hex.EncodeToString(h.Sum(nil)),
	}
}

func (s *script) SetName(name string)                     { s.name = name }
func (s *script) Hash() string                            { return s.hash }
func (s *script) Load(ctx context.Context) StringCmd      { return s.scriptLoad(ctx, s.name, s.src) }
func (s *script) Exists(ctx context.Context) BoolSliceCmd { return s.scriptExists(ctx, s.hash) }
func (s *script) Eval(ctx context.Context, keys []string, args ...any) Cmd {
	return s.client.eval(ctx, s.name, s.src, keys, args...)
}
func (s *script) EvalRO(ctx context.Context, keys []string, args ...any) Cmd {
	return s.client.evalRO(ctx, s.name, s.src, keys, args...)
}

func (s *script) EvalSha(ctx context.Context, keys []string, args ...any) Cmd {
	return s.client.evalSha(ctx, s.name, s.hash, keys, args...)
}
func (s *script) EvalShaRO(ctx context.Context, keys []string, args ...any) Cmd {
	return s.client.evalShaRO(ctx, s.name, s.hash, keys, args...)
}

func (s *script) Run(ctx context.Context, keys []string, args ...any) Cmd {
	r := s.client.evalSha(ctx, s.name, s.hash, keys, args...)
	if isNoScriptError(r.Err()) {
		return s.client.eval(ctx, s.name, s.src, keys, args...)
	}
	return r
}

func (s *script) RunRO(ctx context.Context, keys []string, args ...any) Cmd {
	r := s.client.evalShaRO(ctx, s.name, s.hash, keys, args...)
	if isNoScriptError(r.Err()) {
		return s.client.evalRO(ctx, s.name, s.src, keys, args...)
	}
	return r
}
