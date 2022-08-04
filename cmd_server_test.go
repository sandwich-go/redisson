package redisson

import (
	"context"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func testBgRewriteAOF(ctx context.Context, c Cmdable) []string {
	val, err := c.BgRewriteAOF(ctx).Result()
	if err != nil {
		So(err.Error(), ShouldContainSubstring, "Background append only file rewriting")
	} else {
		So(val, ShouldContainSubstring, "Background append only file rewriting")
	}
	return nil
}

func testBgSave(ctx context.Context, c Cmdable) []string {
	err := c.BgSave(ctx).Err()
	So(err, ShouldNotBeNil)
	So(err.Error(), ShouldContainSubstring, "ERR Another child process is active")

	return nil
}

func testCommand(context.Context, Cmdable) []string { return nil }

func testConfigGet(ctx context.Context, c Cmdable) []string {
	val, err := c.ConfigGet(ctx, "*").Result()
	So(err, ShouldBeNil)
	So(val, ShouldNotBeEmpty)

	return nil
}

func testConfigResetStat(ctx context.Context, c Cmdable) []string {
	r := c.ConfigResetStat(ctx)
	So(r.Err(), ShouldBeNil)
	So(r.Val(), ShouldEqual, OK)

	return nil
}

func testConfigRewrite(ctx context.Context, c Cmdable) []string {
	configRewrite := c.ConfigRewrite(ctx)
	So(configRewrite.Err(), ShouldBeNil)
	So(configRewrite.Val(), ShouldEqual, OK)

	return nil
}

func testConfigSet(ctx context.Context, c Cmdable) []string {
	configGet := c.ConfigGet(ctx, "maxmemory")
	So(configGet.Err(), ShouldBeNil)
	So(len(configGet.Val()), ShouldEqual, 2)
	So(configGet.Val()[0], ShouldEqual, "maxmemory")

	configSet := c.ConfigSet(ctx, "maxmemory", configGet.Val()[1].(string))
	So(configSet.Err(), ShouldBeNil)
	So(configSet.Val(), ShouldEqual, OK)

	return nil
}

func testDBSize(ctx context.Context, c Cmdable) []string {
	dbSize := c.DBSize(ctx)
	So(dbSize.Err(), ShouldBeNil)
	So(dbSize.Val(), ShouldBeZeroValue)
	return nil
}

func testFlushAll(context.Context, Cmdable) []string      { return nil }
func testFlushAllAsync(context.Context, Cmdable) []string { return nil }
func testFlushDB(context.Context, Cmdable) []string       { return nil }
func testFlushDBAsync(context.Context, Cmdable) []string  { return nil }

func testInfo(ctx context.Context, c Cmdable) []string {
	info := c.Info(ctx)
	So(info.Err(), ShouldBeNil)
	So(info.Val(), ShouldNotBeEmpty)

	return nil
}

func testLastSave(ctx context.Context, c Cmdable) []string {
	lastSave := c.LastSave(ctx)
	So(lastSave.Err(), ShouldBeNil)
	So(lastSave.Val(), ShouldBeGreaterThan, 0)

	return nil
}

func testMemoryUsage(ctx context.Context, c Cmdable) []string {
	var key = "key"

	err := c.MemoryUsage(ctx, key).Err()
	So(err, ShouldNotBeNil)
	So(IsNil(err), ShouldBeTrue)

	err = c.Set(ctx, key, "bar", 0).Err()
	So(err, ShouldBeNil)

	n, err1 := c.MemoryUsage(ctx, key).Result()
	So(err1, ShouldBeNil)
	So(n, ShouldBeGreaterThan, 0)

	n, err = c.MemoryUsage(ctx, key, 0).Result()
	So(err, ShouldBeNil)
	So(n, ShouldBeGreaterThan, 0)

	return []string{key}
}

func testSave(ctx context.Context, c Cmdable) []string {
	err := c.Save(ctx).Err()
	So(err, ShouldBeNil)

	return nil
}

func testShutdown(context.Context, Cmdable) []string       { return nil }
func testShutdownSave(context.Context, Cmdable) []string   { return nil }
func testShutdownNoSave(context.Context, Cmdable) []string { return nil }

func testSlaveOf(ctx context.Context, c Cmdable) []string {
	slaveOf := c.SlaveOf(ctx, "localhost", "8888")
	So(slaveOf.Err(), ShouldBeNil)
	So(slaveOf.Val(), ShouldEqual, OK)

	slaveOf = c.SlaveOf(ctx, "NO", "ONE")
	So(slaveOf.Err(), ShouldBeNil)
	So(slaveOf.Val(), ShouldEqual, OK)

	return nil
}

func testTime(ctx context.Context, c Cmdable) []string {
	tm, err := c.Time(ctx).Result()
	So(err, ShouldBeNil)
	So(tm, ShouldNotEqual, time.Now().Add(3*time.Second))

	return nil
}

func testDebugObject(context.Context, Cmdable) []string { return nil }

func serverTestUnits() []TestUnit {
	return []TestUnit{
		{CommandBgRewriteAOF, testBgRewriteAOF},
		{CommandBgSave, testBgSave},
		{CommandCommand, testCommand},
		{CommandConfigGet, testConfigGet},
		{CommandConfigResetStat, testConfigResetStat},
		{CommandConfigRewrite, testConfigRewrite},
		{CommandConfigSet, testConfigSet},
		{CommandDBSize, testDBSize},
		{CommandFlushAll, testFlushAll},
		{CommandFlushAllAsync, testFlushAllAsync},
		{CommandFlushDB, testFlushDB},
		{CommandFlushDBAsync, testFlushDBAsync},
		{CommandInfos, testInfo},
		{CommandLastSave, testLastSave},
		{CommandMemoryUsage, testMemoryUsage},
		{CommandSave, testSave},
		{CommandShutdown, testShutdown},
		{CommandShutdownSave, testShutdownSave},
		{CommandShutdownNoSave, testShutdownNoSave},
		{CommandSlaveOf, testSlaveOf},
		{CommandTime, testTime},
		{CommandDebug, testDebugObject},
	}
}

func TestResp2Client_Server(t *testing.T) { doTestUnits(t, RESP2, serverTestUnits) }
