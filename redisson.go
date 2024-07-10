package redisson

import (
	"context"
	"errors"
	"fmt"
	"github.com/coreos/go-semver/semver"
	"github.com/redis/rueidis"
	"github.com/redis/rueidis/rueidiscompat"
	"strconv"
	"strings"
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
	}
	cp.ttl = ttl
	return cp
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
			return newSliceCmdFromSlice(nil, err)
		}
		_values := ret.Val()
		for _i, _key := range slot2Keys[i] {
			res[keyIndexes[_key]] = _values[_i]
		}
	}
	return newSliceCmdFromSlice(res, nil)
}

func (c *client) Do(ctx context.Context, completed rueidis.Completed) rueidis.RedisResult {
	if c.ttl == 0 {
		return c.cmd.Do(ctx, completed)
	}
	rsp := c.cmd.DoCache(ctx, rueidis.Cacheable(completed), c.ttl)
	c.handler.cache(ctx, rsp.IsCacheHit())
	return rsp
}

func (c *client) sort(command, key string, sort Sort) rueidis.Completed {
	cmd := c.cmd.B().Arbitrary(command).Keys(key)
	if sort.By != "" {
		cmd = cmd.Args("BY", sort.By)
	}
	if sort.Offset != 0 || sort.Count != 0 {
		cmd = cmd.Args("LIMIT", strconv.FormatInt(sort.Offset, 10), strconv.FormatInt(sort.Count, 10))
	}
	for _, get := range sort.Get {
		cmd = cmd.Args("GET").Args(get)
	}
	switch order := strings.ToUpper(sort.Order); order {
	case "ASC", "DESC":
		cmd = cmd.Args(order)
	case "":
	default:
		panic(fmt.Sprintf("invalid sort order %s", sort.Order))
	}
	if sort.Alpha {
		cmd = cmd.Args("ALPHA")
	}
	return cmd.Build()
}

func (c *client) getLPosCompleted(key string, value string, count int64, args LPosArgs) rueidis.Completed {
	arbitrary := c.cmd.B().Arbitrary(LPOS).Keys(key).Args(value)
	if count >= 0 {
		arbitrary = arbitrary.Args(COUNT, str(count))
	}
	if args.Rank != 0 {
		arbitrary = arbitrary.Args(RANK, str(args.Rank))
	}
	if args.MaxLen != 0 {
		arbitrary = arbitrary.Args(MAXLEN, str(args.MaxLen))
	}
	return arbitrary.Build()
}

func (c *client) zAddArgs(ctx context.Context, key string, incr bool, args ZAddArgs, members ...Z) rueidis.RedisResult {
	cmd := c.cmd.B().Arbitrary(ZADD).Keys(key)
	if args.NX {
		cmd = cmd.Args(NX)
	} else {
		if args.XX {
			cmd = cmd.Args(XX)
		}
		if args.GT {
			cmd = cmd.Args(GT)
		} else if args.LT {
			cmd = cmd.Args(LT)
		}
	}
	if args.Ch {
		cmd = cmd.Args(CH)
	}
	if incr {
		cmd = cmd.Args(INCR)
	}
	for _, v := range members {
		cmd = cmd.Args(str(v.Score), str(v.Member))
	}
	return c.cmd.Do(ctx, cmd.Build())
}

func (c *client) getZRangeByScoreCompleted(key string, withScore bool, opt ZRangeBy) rueidis.Completed {
	var completed rueidis.Completed
	var zrangebyscoreMax = c.cmd.B().Zrangebyscore().Key(key).Min(opt.Min).Max(opt.Max)
	if opt.Offset != 0 || opt.Count != 0 {
		if withScore {
			completed = zrangebyscoreMax.Withscores().Limit(opt.Offset, opt.Count).Build()
		} else {
			completed = zrangebyscoreMax.Limit(opt.Offset, opt.Count).Build()
		}
	} else {
		if withScore {
			completed = zrangebyscoreMax.Withscores().Build()
		} else {
			completed = zrangebyscoreMax.Build()
		}
	}
	return completed
}

func (c *client) zRangeArgs(withScores bool, z ZRangeArgs) rueidis.Completed {
	cmd := c.cmd.B().Arbitrary("ZRANGE").Keys(z.Key)
	if z.Rev && (z.ByScore || z.ByLex) {
		cmd = cmd.Args(str(z.Stop), str(z.Start))
	} else {
		cmd = cmd.Args(str(z.Start), str(z.Stop))
	}
	if z.ByScore {
		cmd = cmd.Args("BYSCORE")
	} else if z.ByLex {
		cmd = cmd.Args("BYLEX")
	}
	if z.Rev {
		cmd = cmd.Args("REV")
	}
	if z.Offset != 0 || z.Count != 0 {
		cmd = cmd.Args("LIMIT", strconv.FormatInt(z.Offset, 10), strconv.FormatInt(z.Count, 10))
	}
	if withScores {
		cmd = cmd.Args("WITHSCORES")
	}
	return cmd.Build()
}

func (c *client) getZRevRangeCompleted(key string, start, stop int64, withScore bool) rueidis.Completed {
	var zrevrangeStop = c.cmd.B().Zrevrange().Key(key).Start(start).Stop(stop)
	if withScore {
		return zrevrangeStop.Withscores().Build()
	}
	return zrevrangeStop.Build()
}

func (c *client) getZRevRangeByScoreCompleted(key string, withScore bool, opt ZRangeBy) rueidis.Completed {
	var completed rueidis.Completed
	var zrevrangebyscoreMin = c.cmd.B().Zrevrangebyscore().Key(key).Max(opt.Max).Min(opt.Min)
	if opt.Offset != 0 || opt.Count != 0 {
		if withScore {
			completed = zrevrangebyscoreMin.Withscores().Limit(opt.Offset, opt.Count).Build()
		} else {
			completed = zrevrangebyscoreMin.Limit(opt.Offset, opt.Count).Build()
		}
	} else {
		if withScore {
			completed = zrevrangebyscoreMin.Withscores().Build()
		} else {
			completed = zrevrangebyscoreMin.Build()
		}
	}
	return completed
}

func (c *client) getXAutoClaimCompleted(a XAutoClaimArgs, justId bool) rueidis.Completed {
	var completed rueidis.Completed
	var xautoclaimStart = c.cmd.B().Xautoclaim().Key(a.Stream).Group(a.Group).Consumer(a.Consumer).MinIdleTime(str(formatMs(a.MinIdle))).Start(a.Start)
	if a.Count > 0 {
		if justId {
			completed = xautoclaimStart.Count(a.Count).Justid().Build()
		} else {
			completed = xautoclaimStart.Count(a.Count).Build()
		}
	} else {
		if justId {
			completed = xautoclaimStart.Justid().Build()
		} else {
			completed = xautoclaimStart.Build()
		}
	}
	return completed
}

func (c *client) getXClaimCompleted(a XClaimArgs, justId bool) rueidis.Completed {
	var xclaimId = c.cmd.B().Xclaim().Key(a.Stream).Group(a.Group).Consumer(a.Consumer).MinIdleTime(str(formatMs(a.MinIdle))).Id(a.Messages...)
	if justId {
		return xclaimId.Justid().Build()
	}
	return xclaimId.Build()
}

func (c *client) xtrim(ctx context.Context, key, strategy string,
	approx bool, threshold string, limit int64) IntCmd {
	cmd := c.cmd.B().Arbitrary(XTRIM).Keys(key).Args(strategy)
	if approx {
		cmd = cmd.Args("~")
	}
	cmd = cmd.Args(threshold)
	if limit > 0 {
		cmd = cmd.Args(LIMIT, str(limit))
	}
	return newIntCmdFromResult(c.cmd.Do(ctx, cmd.Build()))
}
