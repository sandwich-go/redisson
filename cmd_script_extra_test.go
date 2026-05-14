//go:build integration

package redisson

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// cmd_script_extra_test.go 补 cmd_script.go 中 EvalRO/EvalShaRO/RunRO/FCall*/Function*
// 这些原 cmd_script_test.go 没覆盖的路径，把 cmd_script.go 覆盖率从 ~31% 拉到 60%+。

// testScriptEvalRO 验证只读脚本 EVAL_RO 路径。
// 注意：脚本本身不能写命令，否则 server 会报错。
func testScriptEvalRO(ctx context.Context, c Cmdable) []string {
	cmd := c.EvalRO(ctx, testScript, []string{"k"}, "ro-value")
	So(cmd.Err(), ShouldBeNil)
	So(cmd.Val(), ShouldEqual, "ro-value")
	return nil
}

// testScriptEvalShaRO EVALSHA_RO 路径。
func testScriptEvalShaRO(ctx context.Context, c Cmdable) []string {
	l := c.ScriptLoad(ctx, testScript)
	So(l.Err(), ShouldBeNil)

	cmd := c.EvalShaRO(ctx, l.Val(), []string{"k"}, "ro-sha")
	So(cmd.Err(), ShouldBeNil)
	So(cmd.Val(), ShouldEqual, "ro-sha")

	c.ScriptFlush(ctx)
	return nil
}

// testScriptRunRO Scripter.RunRO 在脚本未缓存时通过 EVAL_RO 兜底，
// 缓存后通过 EVALSHA_RO 加速；这里测两次调用以触发两条路径。
func testScriptRunRO(ctx context.Context, c Cmdable) []string {
	c.ScriptFlush(ctx)
	s := c.CreateScript(testScript)

	// 第一次：EvalSha 失败 NOSCRIPT → 自动 fallback 到 Eval
	cmd := s.RunRO(ctx, nil, "first")
	So(cmd.Err(), ShouldBeNil)
	So(cmd.Val(), ShouldEqual, "first")

	// 第二次：脚本已缓存 → 直接走 EvalSha
	cmd = s.RunRO(ctx, nil, "second")
	So(cmd.Err(), ShouldBeNil)
	So(cmd.Val(), ShouldEqual, "second")

	c.ScriptFlush(ctx)
	return nil
}

// testScriptCreateWithName SetName/Hash 不变式。
func testScriptCreateWithName(ctx context.Context, c Cmdable) []string {
	s := c.CreateScriptWithName("named-script", testScript)
	if s.Hash() == "" {
		So("hash empty", ShouldEqual, "")
	}
	// SetName 二次设置不应崩溃
	s.SetName("renamed")
	// Run 仍正常
	cmd := s.Run(ctx, nil, "ok")
	So(cmd.Err(), ShouldBeNil)
	So(cmd.Val(), ShouldEqual, "ok")
	c.ScriptFlush(ctx)
	return nil
}

// ----- Function 系列（Redis 7.0+） -----

const testFunctionLib = `#!lua name=mylib
redis.register_function{function_name='myfunc', callback=function(keys, args) return args[1] end, flags={'no-writes'}}`

// testFunctionLoad / FunctionList / FunctionDelete / FunctionFlush 链路。
func testFunctionLifecycle(ctx context.Context, c Cmdable) []string {
	// 干净起点
	c.FunctionFlush(ctx)

	loadRes := c.FunctionLoad(ctx, testFunctionLib)
	So(loadRes.Err(), ShouldBeNil)
	So(loadRes.Val(), ShouldEqual, "mylib")

	// 重复 Load 同 lib 报错；用 LoadReplace 覆盖
	loadAgain := c.FunctionLoad(ctx, testFunctionLib)
	So(loadAgain.Err(), ShouldNotBeNil)

	replaceRes := c.FunctionLoadReplace(ctx, testFunctionLib)
	So(replaceRes.Err(), ShouldBeNil)
	So(replaceRes.Val(), ShouldEqual, "mylib")

	// FunctionList
	listRes := c.FunctionList(ctx, FunctionListQuery{LibraryNamePattern: "*"})
	So(listRes.Err(), ShouldBeNil)
	libs, _ := listRes.Result()
	if len(libs) == 0 {
		So("FunctionList returned empty", ShouldEqual, "")
	}

	// FCall：调用注册的函数
	fcall := c.FCall(ctx, "myfunc", nil, "hello")
	So(fcall.Err(), ShouldBeNil)
	So(fcall.Val(), ShouldEqual, "hello")

	// FCallRO：只读路径
	fcallRO := c.FCallRO(ctx, "myfunc", nil, "ro-hello")
	So(fcallRO.Err(), ShouldBeNil)
	So(fcallRO.Val(), ShouldEqual, "ro-hello")

	// FunctionDelete
	delRes := c.FunctionDelete(ctx, "mylib")
	So(delRes.Err(), ShouldBeNil)

	// 删除后再 List 应为空
	listAfter := c.FunctionList(ctx, FunctionListQuery{LibraryNamePattern: "mylib"})
	So(listAfter.Err(), ShouldBeNil)
	libsAfter, _ := listAfter.Result()
	So(len(libsAfter), ShouldEqual, 0)

	// FunctionFlush idempotent
	flushRes := c.FunctionFlush(ctx)
	So(flushRes.Err(), ShouldBeNil)

	flushAsync := c.FunctionFlushAsync(ctx)
	So(flushAsync.Err(), ShouldBeNil)

	return nil
}

// testFunctionDumpRestore Dump 后 Restore 回去。
func testFunctionDumpRestore(ctx context.Context, c Cmdable) []string {
	c.FunctionFlush(ctx)

	loadRes := c.FunctionLoad(ctx, testFunctionLib)
	So(loadRes.Err(), ShouldBeNil)

	dumpRes := c.FunctionDump(ctx)
	So(dumpRes.Err(), ShouldBeNil)
	dumpVal := dumpRes.Val()
	if dumpVal == "" {
		So("dump empty", ShouldEqual, "")
	}

	c.FunctionFlush(ctx)

	restoreRes := c.FunctionRestore(ctx, dumpVal)
	So(restoreRes.Err(), ShouldBeNil)

	c.FunctionFlush(ctx)
	return nil
}

// testFunctionKill Kill 在没运行时返回 NOTBUSY。
func testFunctionKill(ctx context.Context, c Cmdable) []string {
	res := c.FunctionKill(ctx)
	So(res.Err(), ShouldNotBeNil)
	// "NOTBUSY No scripts in execution right now." 或 "NOTBUSY No functions ..."
	return nil
}

func scriptExtraTestUnits() []TestUnit {
	return []TestUnit{
		{CommandEval, testScriptEvalRO},
		{CommandEvalSha, testScriptEvalShaRO},
		{CommandEvalSha, testScriptRunRO},
		{CommandEval, testScriptCreateWithName},
		{CommandFunctionLoad, testFunctionLifecycle},
		{CommandFunctionDump, testFunctionDumpRestore},
		{CommandFunctionKill, testFunctionKill},
	}
}

// TestClient_ScriptExtra 不调 t.Parallel：脚本相关命令（ScriptFlush/FunctionFlush）
// 是 server 全局的，与 TestClient_Script 并发会相互干扰。串行跑稳定且总时长不变。
func TestClient_ScriptExtra(t *testing.T) {
	c := MustNewClient(NewConf(WithDevelopment(false), WithDB(nextDB())))
	_doTestUnits(t, c, scriptExtraTestUnits)
}
