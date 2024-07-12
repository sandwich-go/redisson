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

var (
	errTooManyArguments                    = errors.New("too many arguments")
	errGeoRadiusByMemberNotSupportStore    = errors.New("GeoRadiusByMember does not support Store or StoreDist")
	errGeoRadiusNotSupportStore            = errors.New("GeoRadius does not support Store or StoreDist")
	errGeoRadiusStoreRequiresStore         = errors.New("GeoRadiusStore requires Store or StoreDist")
	errGeoRadiusByMemberStoreRequiresStore = errors.New("GeoRadiusByMemberStore requires Store or StoreDist")
	errMemoryUsageArgsCount                = errors.New("MemoryUsage expects single sample count")
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

func (c *client) doCache(ctx context.Context, cacheable rueidis.Cacheable) rueidis.RedisResult {
	resp := c.cmd.DoCache(ctx, cacheable, c.ttl)
	c.handler.cache(ctx, resp.IsCacheHit())
	return resp
}

func (c *client) XMGet(ctx context.Context, keys ...string) SliceCmd {
	if len(keys) <= 1 {
		return c.MGet(ctx, keys...)
	}
	var slot2Keys = make(map[uint16][]string)
	var keyIndexes = make(map[string]int)
	for i, key := range keys {
		keySlot := slot(key)
		slot2Keys[keySlot] = append(slot2Keys[keySlot], key)
		keyIndexes[key] = i
	}
	if len(slot2Keys) == 1 {
		return c.MGet(ctx, keys...)
	}
	var wg sync.WaitGroup
	var mx sync.Mutex
	var scs = make(map[uint16]SliceCmd)
	wg.Add(len(slot2Keys))
	for i, sameSlotKeys := range slot2Keys {
		go func(_i uint16, _keys []string) {
			ret := c.MGet(context.Background(), _keys...)
			mx.Lock()
			scs[_i] = ret
			mx.Unlock()
			wg.Done()
		}(i, sameSlotKeys)
	}
	wg.Wait()

	var res = make([]interface{}, len(keys))
	for i, ret := range scs {
		if err := ret.Err(); err != nil {
			return newSliceCmdFromSlice(nil, err, keys...)
		}
		_values := ret.Val()
		for _i, _key := range slot2Keys[i] {
			res[keyIndexes[_key]] = _values[_i]
		}
	}
	return newSliceCmdFromSlice(res, nil, keys...)
}

func (c *client) Do(ctx context.Context, completed rueidis.Completed) rueidis.RedisResult {
	if c.ttl == 0 {
		return c.cmd.Do(ctx, completed)
	}
	return c.doCache(ctx, rueidis.Cacheable(completed))
}
