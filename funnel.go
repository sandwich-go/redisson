package redisson

import (
	"context"
	"github.com/sandwich-go/funnel"
	"time"
)

type funnelScriptBuilder struct{ c Cmdable }
type funnelScript struct{ s Scripter }

func (s funnelScriptBuilder) Build(src string) funnel.RedisScript {
	return funnelScript{s: s.c.CreateScript(src)}
}

func (s funnelScript) EvalSha(ctx context.Context, keys []string, args ...interface{}) ([]interface{}, error) {
	return s.s.EvalSha(ctx, keys, args...).Slice()
}

func (s funnelScript) Eval(ctx context.Context, keys []string, args ...interface{}) ([]interface{}, error) {
	return s.s.Eval(ctx, keys, args...).Slice()
}

func newFunnel(c Cmdable, key string, capacity, operations int64, seconds time.Duration) funnel.Funnel {
	return funnel.NewRedisFunnel(funnelScriptBuilder{c}, key, capacity, operations, seconds)
}

func (c *client) NewFunnel(key string, capacity, operations int64, seconds time.Duration) funnel.Funnel {
	return newFunnel(c, key, capacity, operations, seconds)
}
