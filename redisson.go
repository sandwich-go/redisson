package redisson

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coreos/go-semver/semver"
	"github.com/redis/rueidis"
	"github.com/redis/rueidis/rueidiscompat"
)

type RESP = string

const (
	RESP2 RESP = "RESP2"
	RESP3 RESP = "RESP3"
)

var Nil = rueidis.Nil

func IsNil(err error) bool { return errors.Is(err, Nil) }

type client struct {
	v ConfInterface
	// version / isCluster 在 Connect.reviseVersion / reviseCluster 时由
	// 后台 goroutine 写入；命令路径并发读。改 atomic 防 reconnect 路径
	// 与其他派生 client 命令读形成 data race。
	// 派生 client（ForEachNodes / Cache）通过 cloneClientFrom 复制 atomic 值。
	version   atomic.Pointer[semver.Version]
	isCluster atomic.Bool

	handler     handler
	cmd         rueidis.Client
	adapter     rueidiscompat.Cmdable
	ttl         time.Duration
	builder     builder
	maxp        int
	delayQueues sync.Map
	// delayWorkerWG 跟踪所有 delay queue 在途 worker（callback + ack 阶段）。
	// q.Close 不等 worker（避免 callback 内死锁），但在 client.Close 时
	// 必须等所有 worker 收尾后才能释放底层连接，否则 worker 中的脚本调用
	// 会与 c.cmd=nil 写 race。
	delayWorkerWG sync.WaitGroup

	once sync.Once
}

// cloneClientFrom 派生新 client 时复制 src 的 atomic 字段（version / isCluster）。
// 不是浅拷贝整个 struct（避免 atomic 字段被复制 header → race detector 告警）。
func cloneClientFrom(src *client) *client {
	dst := &client{
		v:       src.v,
		handler: src.handler,
		ttl:     src.ttl,
		builder: src.builder,
		maxp:    src.maxp,
	}
	if v := src.version.Load(); v != nil {
		dst.version.Store(v)
	}
	dst.isCluster.Store(src.isCluster.Load())
	return dst
}

func MustNewClient(v ConfInterface) Cmdable {
	cmd, err := Connect(v)
	if err != nil {
		panic(err)
	}
	return cmd
}

func (c *client) Options() ConfVisitor { return c.v }
func (c *client) IsCluster() bool      { return c.isCluster.Load() }
func (c *client) ForEachNodes(ctx context.Context, f func(context.Context, Cmdable) error) error {
	if !c.isCluster.Load() {
		return f(ctx, c)
	}
	var errs Errors
	for _, v := range c.cmd.Nodes() {
		dst := cloneClientFrom(c)
		dst.cmd = v
		dst.adapter = rueidiscompat.NewAdapter(v)
		err := f(ctx, dst)
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
	cp := cloneClientFrom(c)
	cp.cmd = c.cmd
	cp.adapter = c.adapter
	cp.ttl = ttl
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
