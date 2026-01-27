package redisson

import (
	"context"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

type FunnelName string

func (f FunnelName) String() string {
	return string(f)
}

func TestFunnel(t *testing.T) {
	doTestUnits(t, func() []TestUnit {
		return []TestUnit{{FunnelName("funnel"), testFunnel}}
	})
}

func testFunnel(ctx context.Context, c Cmdable) []string {
	var capacity int64 = 1
	var key = "mykey"
	var operations int64 = 1
	var seconds = 1 * time.Second
	f := c.NewFunnel(key, capacity, operations, seconds)
	s, err := f.Watering(context.Background(), 1)
	So(err, ShouldBeNil)
	So(s.Ready, ShouldBeTrue)
	s, err = f.Watering(context.Background(), 1)
	So(err, ShouldBeNil)
	So(s.Ready, ShouldBeFalse)
	time.Sleep(s.Interval)
	s, err = f.Watering(context.Background(), 1)
	So(err, ShouldBeNil)
	So(s.Ready, ShouldBeTrue)
	return []string{key}
}
