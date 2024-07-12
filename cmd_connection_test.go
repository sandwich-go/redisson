package redisson

import (
	"context"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func testClientGetName(ctx context.Context, c Cmdable) []string {
	pipe := c.Pipeline()
	_ = pipe.Put(ctx, CommandClientSetName, nil, "theclientname")
	_, err := pipe.Exec(ctx)
	So(err, ShouldBeNil)

	get := c.ClientGetName(ctx)
	So(get.Err(), ShouldBeNil)
	So(get.Val(), ShouldEqual, "theclientname")

	return nil
}

func testClientID(ctx context.Context, c Cmdable) []string {
	cc := c.ClientID(ctx)
	So(cc.Err(), ShouldBeNil)
	So(cc.Val(), ShouldBeGreaterThan, 0)

	return nil
}

func testClientKill(ctx context.Context, c Cmdable) []string {
	r := c.ClientKill(ctx, "1.1.1.1:1111")
	So(r.Err(), ShouldNotBeNil)
	So(r.Err().Error(), ShouldContainSubstring, "No such client")

	return nil
}

func testClientKillByFilter(ctx context.Context, c Cmdable) []string {
	r := c.ClientKillByFilter(ctx, "TYPE", "test")
	So(r.Err(), ShouldNotBeNil)
	So(r.Err().Error(), ShouldContainSubstring, "Unknown client type 'test'")

	return nil
}

func testClientList(ctx context.Context, c Cmdable) []string {
	clientList := c.ClientList(ctx)
	So(clientList.Err(), ShouldBeNil)
	So(len(clientList.Val()), ShouldBeGreaterThan, 0)

	return nil
}

func testClientPause(ctx context.Context, c Cmdable) []string {
	err := c.ClientPause(ctx, time.Second).Err()
	So(err, ShouldBeNil)

	start := time.Now()
	err = c.Ping(ctx).Err()
	So(err, ShouldBeNil)
	So(time.Now(), ShouldNotEqual, start.Add(time.Second))

	return nil
}

func testEcho(ctx context.Context, c Cmdable) []string {
	echo := c.Echo(ctx, "hello")
	So(echo.Err(), ShouldBeNil)
	So(echo.Val(), ShouldEqual, "hello")

	return nil
}

func testPing(ctx context.Context, c Cmdable) []string {
	echo := c.Ping(ctx)
	So(echo.Err(), ShouldBeNil)
	So(echo.Val(), ShouldEqual, "PONG")

	return nil
}

func testQuit(_ context.Context, _ Cmdable) []string {
	return nil
}

func connectionTestUnits() []TestUnit {
	return []TestUnit{
		{CommandClientGetName, testClientGetName},
		{CommandClientID, testClientID},
		{CommandClientKill, testClientKill},
		{CommandClientKillByFilter, testClientKillByFilter},
		{CommandClientList, testClientList},
		{CommandClientPause, testClientPause},
		{CommandEcho, testEcho},
		{CommandPing, testPing},
		{CommandQuit, testQuit},
	}
}

func TestClient_Connection(t *testing.T) { doTestUnits(t, connectionTestUnits) }
