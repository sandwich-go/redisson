package redisson

import (
	"context"
	"errors"
	"github.com/coreos/go-semver/semver"
	"github.com/redis/rueidis"
	"github.com/redis/rueidis/rueidiscompat"
	"sync"
	"time"
)

type RESP = string

const (
	RESP2 RESP = "RESP2"
	RESP3 RESP = "RESP3"
)

var Nil = rueidis.Nil

func IsNil(err error) bool { return errors.Is(err, Nil) }

type client struct {
	v         ConfInterface
	version   semver.Version
	handler   handler
	isCluster bool
	cmd       rueidis.Client
	adapter   rueidiscompat.Cmdable
	ttl       time.Duration
	builder   builder
	maxp      int

	once sync.Once
}

func MustNewClient(v ConfInterface) Cmdable {
	cmd, err := Connect(v)
	if err != nil {
		panic(err)
	}
	return cmd
}

func (c *client) Options() ConfVisitor { return c.v }
func (c *client) IsCluster() bool      { return c.isCluster }
func (c *client) ForEachNodes(ctx context.Context, f func(context.Context, Cmdable) error) error {
	if !c.isCluster {
		return f(ctx, c)
	}
	var errs Errors
	for _, v := range c.cmd.Nodes() {
		err := f(ctx, &client{
			v:         c.v,
			version:   c.version,
			handler:   c.handler,
			isCluster: c.isCluster,
			cmd:       v,
			adapter:   rueidiscompat.NewAdapter(v),
			builder:   c.builder,
			maxp:      c.maxp,
		})
		if err != nil {
			errs.Push(err)
		}
	}
	return errs.Err()
}

func (c *client) Cache(ttl time.Duration) CacheCmdable {
	if !c.v.GetEnableCache() || c.ttl == ttl {
		return c
	}
	cp := &client{
		v:         c.v,
		version:   c.version,
		handler:   c.handler,
		isCluster: c.isCluster,
		cmd:       c.cmd,
		adapter:   c.adapter,
		ttl:       ttl,
		builder:   c.builder,
		maxp:      c.maxp,
	}
	return cp
}

func (c *client) Do(ctx context.Context, completed Completed) RedisResult {
	if c.ttl <= 0 {
		return c.cmd.Do(ctx, completed)
	}
	resp := c.cmd.DoCache(ctx, rueidis.Cacheable(completed), c.ttl)
	c.handler.cache(ctx, resp.IsCacheHit())
	return resp
}
