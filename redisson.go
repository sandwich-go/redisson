package redisson

import (
	"context"
	"errors"
	"github.com/coreos/go-semver/semver"
	"github.com/redis/rueidis"
	"github.com/redis/rueidis/rueidiscompat"
	"strconv"
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
	}
	return cp
}

func (c *client) Do(ctx context.Context, completed rueidis.Completed) rueidis.RedisResult {
	if c.ttl <= 0 {
		return c.cmd.Do(ctx, completed)
	}
	return c.doCache(ctx, rueidis.Cacheable(completed))
}

func (c *client) doCache(ctx context.Context, cacheable rueidis.Cacheable) rueidis.RedisResult {
	resp := c.cmd.DoCache(ctx, cacheable, c.ttl)
	c.handler.cache(ctx, resp.IsCacheHit())
	return resp
}

func (c *client) zRangeArgs(withScores bool, z ZRangeArgs) rueidis.Cacheable {
	cmd := c.cmd.B().Arbitrary(ZRANGE).Keys(z.Key)
	if z.Rev && (z.ByScore || z.ByLex) {
		cmd = cmd.Args(str(z.Stop), str(z.Start))
	} else {
		cmd = cmd.Args(str(z.Start), str(z.Stop))
	}
	if z.ByScore {
		cmd = cmd.Args(BYSCORE)
	} else if z.ByLex {
		cmd = cmd.Args(BYLEX)
	}
	if z.Rev {
		cmd = cmd.Args(REV)
	}
	if z.Offset != 0 || z.Count != 0 {
		cmd = cmd.Args(LIMIT, strconv.FormatInt(z.Offset, 10), strconv.FormatInt(z.Count, 10))
	}
	if withScores {
		cmd = cmd.Args(WITHSCORES)
	}
	return rueidis.Cacheable(cmd.Build())
}
