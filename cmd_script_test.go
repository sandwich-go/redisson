package redisson

import (
	"context"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

const testScript = `return ARGV[1]`

func testScriptExists(ctx context.Context, c Cmdable) []string {
	scriptExists := c.ScriptExists(ctx, "1")
	So(scriptExists.Err(), ShouldBeNil)
	So(len(scriptExists.Val()), ShouldEqual, 1)
	So(scriptExists.Val()[0], ShouldBeFalse)

	s := c.CreateScript(testScript)
	l := s.Load(ctx)
	So(l.Err(), ShouldBeNil)
	So(l.Val(), ShouldNotBeEmpty)

	scriptExists = c.ScriptExists(ctx, l.Val(), "1")
	So(scriptExists.Err(), ShouldBeNil)
	So(len(scriptExists.Val()), ShouldEqual, 2)
	So(scriptExists.Val()[0], ShouldBeTrue)
	So(scriptExists.Val()[1], ShouldBeFalse)

	sd := c.ScriptFlush(ctx)
	So(sd.Err(), ShouldBeNil)
	So(sd.Val(), ShouldEqual, OK)

	scriptExists = c.ScriptExists(ctx, l.Val())
	So(scriptExists.Err(), ShouldBeNil)
	So(len(scriptExists.Val()), ShouldEqual, 1)
	So(scriptExists.Val()[0], ShouldBeFalse)

	return nil
}

func testScriptFlush(ctx context.Context, c Cmdable) []string {
	s := c.CreateScript(testScript)
	l := s.Load(ctx)
	So(l.Err(), ShouldBeNil)
	So(l.Val(), ShouldNotBeEmpty)

	scriptExists := c.ScriptExists(ctx, l.Val())
	So(scriptExists.Err(), ShouldBeNil)
	So(len(scriptExists.Val()), ShouldEqual, 1)
	So(scriptExists.Val()[0], ShouldBeTrue)

	sd := c.ScriptFlush(ctx)
	So(sd.Err(), ShouldBeNil)
	So(sd.Val(), ShouldEqual, OK)

	scriptExists = c.ScriptExists(ctx, l.Val())
	So(scriptExists.Err(), ShouldBeNil)
	So(len(scriptExists.Val()), ShouldEqual, 1)
	So(scriptExists.Val()[0], ShouldBeFalse)

	return nil
}

func testScriptLoad(ctx context.Context, c Cmdable) []string {
	l := c.ScriptLoad(ctx, testScript)
	So(l.Err(), ShouldBeNil)
	So(l.Val(), ShouldNotBeEmpty)

	scriptExists := c.ScriptExists(ctx, l.Val())
	So(scriptExists.Err(), ShouldBeNil)
	So(len(scriptExists.Val()), ShouldEqual, 1)
	So(scriptExists.Val()[0], ShouldBeTrue)

	sd := c.ScriptFlush(ctx)
	So(sd.Err(), ShouldBeNil)
	So(sd.Val(), ShouldEqual, OK)

	scriptExists = c.ScriptExists(ctx, l.Val())
	So(scriptExists.Err(), ShouldBeNil)
	So(len(scriptExists.Val()), ShouldEqual, 1)
	So(scriptExists.Val()[0], ShouldBeFalse)

	return nil
}

func testScriptKill(ctx context.Context, c Cmdable) []string {
	l := c.ScriptLoad(ctx, testScript)
	So(l.Err(), ShouldBeNil)
	So(l.Val(), ShouldNotBeEmpty)

	sk := c.ScriptKill(ctx)
	So(sk.Err(), ShouldNotBeNil)
	So(sk.Err().Error(), ShouldEqual, "NOTBUSY No scripts in execution right now.")

	sd := c.ScriptFlush(ctx)
	So(sd.Err(), ShouldBeNil)
	So(sd.Val(), ShouldEqual, OK)

	scriptExists := c.ScriptExists(ctx, l.Val())
	So(scriptExists.Err(), ShouldBeNil)
	So(len(scriptExists.Val()), ShouldEqual, 1)
	So(scriptExists.Val()[0], ShouldBeFalse)

	return nil
}

func testScriptEval(ctx context.Context, c Cmdable) []string {
	var key, value = "mykey", "hello"

	cmd := c.Eval(ctx, testScript, []string{key}, value)
	So(cmd.Err(), ShouldBeNil)
	So(cmd.Val(), ShouldEqual, value)

	sd := c.ScriptFlush(ctx)
	So(sd.Err(), ShouldBeNil)
	So(sd.Val(), ShouldEqual, OK)

	return nil
}

func testScriptEvalSha(ctx context.Context, c Cmdable) []string {
	var key, value = "mykey", "hello"

	cmd := c.EvalSha(ctx, testScript, []string{key}, value)
	So(cmd.Err(), ShouldNotBeNil)
	So(cmd.Err().Error(), ShouldEqual, "NOSCRIPT No matching script. Please use EVAL.")

	l := c.ScriptLoad(ctx, testScript)
	So(l.Err(), ShouldBeNil)
	So(l.Val(), ShouldNotBeEmpty)

	cmd = c.EvalSha(ctx, l.Val(), []string{key}, value)
	So(cmd.Err(), ShouldBeNil)
	So(cmd.Val(), ShouldEqual, value)

	sd := c.ScriptFlush(ctx)
	So(sd.Err(), ShouldBeNil)
	So(sd.Val(), ShouldEqual, OK)

	return nil
}

func testScriptRun(ctx context.Context, c Cmdable) []string {
	var value = "hello"

	s := c.CreateScript(testScript)
	l := s.Load(ctx)
	So(l.Err(), ShouldBeNil)
	So(l.Val(), ShouldNotBeEmpty)

	cmd := s.Run(ctx, nil, value)
	So(cmd.Err(), ShouldBeNil)
	So(cmd.Val(), ShouldEqual, value)

	sd := c.ScriptFlush(ctx)
	So(sd.Err(), ShouldBeNil)
	So(sd.Val(), ShouldEqual, OK)

	return nil
}

func scriptTestUnits() []TestUnit {
	return []TestUnit{
		{CommandScriptExists, testScriptExists},
		{CommandScriptFlush, testScriptFlush},
		{CommandScriptLoad, testScriptLoad},
		{CommandScriptKill, testScriptKill},
		{CommandEval, testScriptEval},
		{CommandEvalSha, testScriptEvalSha},
		{CommandEvalSha, testScriptRun},
	}
}

func TestClient_Script(t *testing.T) { doTestUnits(t, scriptTestUnits) }
