package redisson

import (
	"context"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func testPublish(ctx context.Context, c Cmdable) []string {
	p := c.Publish(ctx, "channel", "one")
	So(p.Err(), ShouldBeNil)
	So(p.Val(), ShouldEqual, 0)

	return nil
}

func testSubscribe(ctx context.Context, c Cmdable) []string {
	s := c.Subscribe(ctx)
	err := s.Subscribe(ctx, "channel")
	So(err, ShouldBeNil)

	p := c.Publish(ctx, "channel", "one")
	So(p.Err(), ShouldBeNil)
	So(p.Val(), ShouldEqual, 1)

	var index int
	for msg := range s.Channel() {
		index++
		So(msg.Payload, ShouldEqual, "one")
		break
	}

	So(index, ShouldBeGreaterThan, 0)

	err = s.Unsubscribe(ctx, "channel")
	So(err, ShouldBeNil)

	err = s.Close()
	So(err, ShouldBeNil)

	return nil
}

func testPubSubChannels(ctx context.Context, c Cmdable) []string {
	s := c.Subscribe(ctx)
	err := s.Subscribe(ctx, "channel1", "channel2")
	So(err, ShouldBeNil)

	pubSubChannels := c.PubSubChannels(ctx, "*")
	So(pubSubChannels.Err(), ShouldBeNil)
	So(stringSliceEqual(pubSubChannels.Val(), []string{"channel1", "channel2"}, false), ShouldBeTrue)

	err = s.Unsubscribe(ctx, "channel1")
	So(err, ShouldBeNil)

	pubSubChannels = c.PubSubChannels(ctx, "*")
	So(pubSubChannels.Err(), ShouldBeNil)
	So(stringSliceEqual(pubSubChannels.Val(), []string{"channel2"}, false), ShouldBeTrue)

	err = s.Close()
	So(err, ShouldBeNil)

	return nil
}

func testPSubscribe(ctx context.Context, c Cmdable) []string {
	s := c.Subscribe(ctx)

	err := s.PSubscribe(ctx, "channel1.*")
	So(err, ShouldBeNil)

	time.Sleep(1 * time.Second)

	pubSubNumPat := c.PubSubNumPat(ctx)
	So(pubSubNumPat.Err(), ShouldBeNil)
	So(pubSubNumPat.Val(), ShouldEqual, 1)

	err = s.Close()
	So(err, ShouldBeNil)

	return nil
}

func testPUnsubscribe(ctx context.Context, c Cmdable) []string {
	s := c.Subscribe(ctx)

	err := s.PSubscribe(ctx, "channel1.*")
	So(err, ShouldBeNil)

	pubSubNumPat := c.PubSubNumPat(ctx)
	So(pubSubNumPat.Err(), ShouldBeNil)
	So(pubSubNumPat.Val(), ShouldEqual, 1)

	err = s.PUnsubscribe(ctx)
	So(err, ShouldBeNil)

	pubSubNumPat = c.PubSubNumPat(ctx)
	So(pubSubNumPat.Err(), ShouldBeNil)
	So(pubSubNumPat.Val(), ShouldEqual, 0)

	err = s.Close()
	So(err, ShouldBeNil)

	return nil
}

func testPubSubNumPat(ctx context.Context, c Cmdable) []string {
	s := c.Subscribe(ctx)

	pubSubNumPat := c.PubSubNumPat(ctx)
	So(pubSubNumPat.Err(), ShouldBeNil)
	So(pubSubNumPat.Val(), ShouldEqual, 0)

	err := s.PSubscribe(ctx, "channel1.*", "channel2.*")
	So(err, ShouldBeNil)

	pubSubNumPat = c.PubSubNumPat(ctx)
	So(pubSubNumPat.Err(), ShouldBeNil)
	So(pubSubNumPat.Val(), ShouldEqual, 2)

	err = s.Close()
	So(err, ShouldBeNil)

	return nil
}

func testPubSubNumSub(ctx context.Context, c Cmdable) []string {
	s := c.Subscribe(ctx)
	err := s.Subscribe(ctx, "channel1", "channel2")
	So(err, ShouldBeNil)

	pubSubNumSub := c.PubSubNumSub(ctx, "channel1", "channel2", "channel3")
	So(pubSubNumSub.Err(), ShouldBeNil)
	So(len(pubSubNumSub.Val()), ShouldEqual, 3)
	So(pubSubNumSub.Val()["channel1"], ShouldEqual, 1)
	So(pubSubNumSub.Val()["channel2"], ShouldEqual, 1)
	So(pubSubNumSub.Val()["channel3"], ShouldEqual, 0)

	pubSubNumSub = c.PubSubNumSub(ctx)
	So(pubSubNumSub.Err(), ShouldBeNil)
	So(pubSubNumSub.Val(), ShouldBeEmpty)

	err = s.Close()
	So(err, ShouldBeNil)

	return nil
}

func pubSubTestUnits() []TestUnit {
	return []TestUnit{
		{CommandPublish, testPublish},
		{CommandSubscribe, testSubscribe},
		{CommandPubSubChannels, testPubSubChannels},
		{CommandPSubscribe, testPSubscribe},
		{CommandPUnsubscribe, testPUnsubscribe},
		{CommandPubSubNumPat, testPubSubNumPat},
		{CommandPubSubNumSub, testPubSubNumSub},
	}
}

func TestResp2Client_PubSub(t *testing.T) { doTestUnits(t, RESP2, pubSubTestUnits) }
func TestResp3Client_PubSub(t *testing.T) { doTestUnits(t, RESP3, pubSubTestUnits) }
