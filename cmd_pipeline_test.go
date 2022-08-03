package sandwich_redis

import (
	"context"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type pipeline string

func (pipeline) String() string { return "Pipeline" }

func testPipeline(ctx context.Context, c Cmdable) []string {
	var key1, key2, value1, value2 = "key1", "key2", "value1", "value2"
	pip := c.Pipeline()
	err := pip.Put(ctx, CommandSet, []string{key1}, value1)
	So(err, ShouldBeNil)

	err = pip.Put(ctx, CommandSet, []string{key2}, value2)
	So(err, ShouldBeNil)

	err = pip.Put(ctx, CommandGet, []string{key1})
	So(err, ShouldBeNil)

	err = pip.Put(ctx, CommandGet, []string{key2})
	So(err, ShouldBeNil)

	var res []interface{}
	res, err = pip.Exec(ctx)
	So(err, ShouldBeNil)
	So(res, ShouldNotBeNil)
	So(len(res), ShouldEqual, 4)
	So(res[0], ShouldEqual, OK)
	So(res[1], ShouldEqual, OK)
	So(res[2], ShouldEqual, value1)
	So(res[3], ShouldEqual, value2)

	return []string{key1, key2}
}

func pipelineTestUnits() []TestUnit {
	return []TestUnit{
		{new(pipeline), testPipeline},
	}
}

func TestResp2Client_Pipeline(t *testing.T) { doTestUnits(t, RESP2, pipelineTestUnits) }
