package redisson

import (
	"context"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type _pipeline string

func (_pipeline) String() string { return "Pipeline" }

func testPipeline(ctx context.Context, c Cmdable) []string {
	var key1, key2, value1, value2 = "key1", "key2", "value1", "value2"
	var pip = func() Pipeliner {
		pip := c.Pipeline()
		CommandSet.P(pip).Cmd(key1, value1, 0)
		CommandSet.P(pip).Cmd(key2, value2, 0)
		CommandGet.P(pip).Cmd(key1)
		CommandGet.P(pip).Cmd(key2)
		return pip
	}

	res, err := pip().Exec(ctx)
	So(err, ShouldBeNil)
	So(res, ShouldNotBeNil)
	So(len(res), ShouldEqual, 4)
	So(res[0], ShouldEqual, OK)
	So(res[1], ShouldEqual, OK)
	So(res[2], ShouldEqual, value1)
	So(res[3], ShouldEqual, value2)

	var cmds []BaseCmd
	cmds, err = pip().ExecCmds(ctx)
	So(err, ShouldBeNil)
	So(cmds, ShouldNotBeNil)
	So(len(cmds), ShouldEqual, 4)
	So(CommandSet.PR(cmds[0]).Val(), ShouldEqual, OK)
	So(CommandSet.PR(cmds[1]).Val(), ShouldEqual, OK)
	So(CommandGet.PR(cmds[2]).Val(), ShouldEqual, value1)
	So(CommandGet.PR(cmds[3]).Val(), ShouldEqual, value2)

	return []string{key1, key2}
}

func pipelineTestUnits() []TestUnit {
	return []TestUnit{
		{new(_pipeline), testPipeline},
	}
}

func TestClient_Pipeline(t *testing.T) { doTestUnits(t, pipelineTestUnits) }
