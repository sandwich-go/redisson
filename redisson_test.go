package redisson

import (
	"context"
	"errors"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestRedisson(t *testing.T) {
	testAddr := ""
	Convey("redisson should work ok", t, func() {
		for _, v := range []RESP{
			RESP2, RESP3,
		} {
			opts := []ConfOption{WithResp(v), WithCluster(true), WithDevelopment(false)}
			if len(testAddr) > 0 {
				opts = append(opts, WithAddrs(testAddr))
			}
			c := MustNewClient(NewConf(opts...))
			So(c.Ping(context.Background()).Err(), ShouldBeNil)
			So(c.Close(), ShouldBeNil)
		}
	})
}

func TestPubSubReceive(t *testing.T) {
	testAddr := "127.0.0.1:55000"
	opts := []ConfOption{WithResp(RESP3), WithDevelopment(false)}
	opts = append(opts, WithAddrs(testAddr))
	c := MustNewClient(NewConf(opts...))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	err := c.PReceive(ctx, func(message Message) {
		t.Log(message)
	}, "redisson*")
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		t.Error(err)
	}
	cancel()
}
