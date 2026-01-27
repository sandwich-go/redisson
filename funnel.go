package redisson

import (
	"context"
	"github.com/sandwich-go/funnel"
	"time"
)

type funnelScriptBuilder struct{ c Cmdable }
type funnelScript struct{ s Scripter }

func (s funnelScript) Run(ctx context.Context, keys []string, args ...interface{}) ([]interface{}, error) {
	return s.s.Run(ctx, keys, args...).Slice()
}

func (s funnelScriptBuilder) Build(name, src string) funnel.RedisScript {
	return funnelScript{s: s.c.CreateScriptWithName(name, src)}
}

func (c *client) NewFunnel(key string, capacity, operations int64, seconds time.Duration) funnel.Funnel {
	return funnel.NewRedis(funnelScriptBuilder{c}, key, capacity, funnel.WithOperations(operations), funnel.WithSeconds(seconds))
}
