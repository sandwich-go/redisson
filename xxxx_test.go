package redisson

import (
	"context"
	"testing"
)

func TestNewPubOptions(t *testing.T) {
	c := MustNewClient(NewConf(WithAddrs("10.21.66.4:6379")))
	err := c.Info(context.Background(), "server").Err()
	if err != nil {
		t.Error(err)
	}
}
