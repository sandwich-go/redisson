package redisson

import (
	"context"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
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
