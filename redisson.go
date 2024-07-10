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

func (c *client) getBitCountCompleted(key string, bitCount *BitCount) rueidis.Completed {
	var bitCountKey = c.cmd.B().Bitcount().Key(key)
	if bitCount != nil {
		return bitCountKey.Start(bitCount.Start).End(bitCount.End).Build()
	}
	return bitCountKey.Build()
}

func (c *client) getBitPosCompleted(key string, bit int64, pos ...int64) rueidis.Completed {
	var completed rueidis.Completed
	var bitposBit = c.cmd.B().Bitpos().Key(key).Bit(bit)
	switch len(pos) {
	case 0:
		completed = bitposBit.Build()
	case 1:
		completed = bitposBit.Start(pos[0]).Build()
	case 2:
		completed = bitposBit.Start(pos[0]).End(pos[1]).Build()
	default:
		panic(errTooManyArguments)
	}
	return completed
}

func (c *client) getBitCompleted(key string, offset int64) rueidis.Completed {
	return c.cmd.B().Getbit().Key(key).Offset(offset).Build()
}

func (c *client) copy(ctx context.Context, sourceKey string, destKey string, db int, replace bool) IntCmd {
	var completed rueidis.Completed
	var cmd = c.cmd.B().Copy().Source(sourceKey).Destination(destKey).Db(int64(db))
	if replace {
		completed = cmd.Replace().Build()
	} else {
		completed = cmd.Build()
	}
	return newIntCmdFromResult(c.cmd.Do(ctx, completed))
}

func (c *client) getExistsCompleted(keys ...string) rueidis.Completed {
	return c.cmd.B().Exists().Key(keys...).Build()
}

func (c *client) migrate(ctx context.Context, host, port, key string, db int, timeout time.Duration) StatusCmd {
	iport, err := parseInt(port)
	if err != nil {
		return newStatusCmdWithError(err)
	}
	var migratePort = c.cmd.B().Migrate().Host(host).Port(iport)
	return newStatusCmdFromResult(c.cmd.Do(ctx, migratePort.Key(key).DestinationDb(int64(db)).Timeout(formatSec(timeout)).Build()))
}

func (c *client) getPTTLCompleted(key string) rueidis.Completed {
	return c.cmd.B().Pttl().Key(key).Build()
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

func (c *client) getTTLCompleted(key string) rueidis.Completed {
	return c.cmd.B().Ttl().Key(key).Build()
}

func (c *client) getTypeCompleted(key string) rueidis.Completed {
	return c.cmd.B().Type().Key(key).Build()
}

func (c *client) geoAdd(ctx context.Context, key string, geoLocation ...GeoLocation) IntCmd {
	cmd := c.cmd.B().Geoadd().Key(key).LongitudeLatitudeMember()
	for _, loc := range geoLocation {
		cmd = cmd.LongitudeLatitudeMember(loc.Longitude, loc.Latitude, loc.Name)
	}
	return newIntCmdFromResult(c.cmd.Do(ctx, cmd.Build()))
}

func (c *client) getGeoDistCompleted(key string, member1, member2, unit string) rueidis.Completed {
	var completed rueidis.Completed
	var geodistMember2 = c.cmd.B().Geodist().Key(key).Member1(member1).Member2(member2)
	switch strings.ToUpper(unit) {
	case M:
		completed = geodistMember2.M().Build()
	case MI:
		completed = geodistMember2.Mi().Build()
	case FT:
		completed = geodistMember2.Ft().Build()
	case KM, EMPTY:
		completed = geodistMember2.Km().Build()
	default:
		panic(fmt.Sprintf("invalid unit %s", unit))
	}
	return completed
}

func (c *client) getGeoHashCompleted(key string, members ...string) rueidis.Completed {
	return c.cmd.B().Geohash().Key(key).Member(members...).Build()
}

func (c *client) getGeoPosCompleted(key string, members ...string) rueidis.Completed {
	return c.cmd.B().Geopos().Key(key).Member(members...).Build()
}

func (c *client) getGeoRadiusCompleted(key string, longitude, latitude float64, q GeoRadiusQuery) rueidis.Completed {
	return c.cmd.B().Arbitrary(GEORADIUS_RO).Keys(key).Args(str(longitude), str(latitude)).Args(getGeoRadiusQueryArgs(q)...).Build()
}

func (c *client) geoRadius(ctx context.Context, key string, longitude, latitude float64, q GeoRadiusQuery) GeoLocationCmd {
	if len(q.Store) > 0 || len(q.StoreDist) > 0 {
		return newGeoLocationCmdWithError(errGeoRadiusNotSupportStore)
	}
	return newGeoLocationCmd(c.Do(ctx, c.getGeoRadiusCompleted(key, longitude, latitude, q)), q)
}

func (c *client) geoRadiusStore(ctx context.Context, key string, longitude, latitude float64, q GeoRadiusQuery) IntCmd {
	cmd := c.cmd.B().Arbitrary(GEORADIUS).Keys(key).Args(str(longitude), str(latitude))
	if len(q.Store) == 0 && len(q.StoreDist) == 0 {
		return newIntCmdWithError(errGeoRadiusStoreRequiresStore)
	}
	return newIntCmdFromResult(c.cmd.Do(ctx, cmd.Args(getGeoRadiusQueryArgs(q)...).Build()))
}

func (c *client) getGeoRadiusByMemberCompleted(key, member string, q GeoRadiusQuery) rueidis.Completed {
	return c.cmd.B().Arbitrary(GEORADIUSBYMEMBER_RO).Keys(key).Args(member).Args(getGeoRadiusQueryArgs(q)...).Build()
}

func (c *client) geoRadiusByMember(ctx context.Context, key, member string, q GeoRadiusQuery) GeoLocationCmd {
	if len(q.Store) > 0 || len(q.StoreDist) > 0 {
		return newGeoLocationCmdWithError(errGeoRadiusByMemberNotSupportStore)
	}
	return newGeoLocationCmd(c.Do(ctx, c.getGeoRadiusByMemberCompleted(key, member, q)), q)
}

func (c *client) geoRadiusByMemberStore(ctx context.Context, key, member string, q GeoRadiusQuery) IntCmd {
	cmd := c.cmd.B().Arbitrary(GEORADIUSBYMEMBER).Keys(key).Args(member)
	if len(q.Store) == 0 && len(q.StoreDist) == 0 {
		return newIntCmdWithError(errGeoRadiusByMemberStoreRequiresStore)
	}
	return newIntCmdFromResult(c.cmd.Do(ctx, cmd.Args(getGeoRadiusQueryArgs(q)...).Build()))
}

func (c *client) getGeoSearchCompleted(key string, q GeoSearchQuery) rueidis.Completed {
	return c.cmd.B().Arbitrary(GEOSEARCH).Keys(key).Args(getGeoSearchQueryArgs(q)...).Build()
}

func (c *client) getGeoSearchLocationCompleted(key string, q GeoSearchLocationQuery) rueidis.Completed {
	return c.cmd.B().Arbitrary(GEOSEARCH).Keys(key).Args(getGeoSearchLocationQueryArgs(q)...).Build()
}

func (c *client) geoSearchStore(ctx context.Context, src, dest string, q GeoSearchStoreQuery) IntCmd {
	cmd := c.cmd.B().Arbitrary(GEOSEARCHSTORE).Keys(dest, src).Args(getGeoSearchQueryArgs(q.GeoSearchQuery)...)
	if q.StoreDist {
		cmd = cmd.Args(STOREDIST)
	}
	return newIntCmdFromResult(c.cmd.Do(ctx, cmd.Build()))
}

func (c *client) getHExistsCompleted(key, field string) rueidis.Completed {
	return c.cmd.B().Hexists().Key(key).Field(field).Build()
}

func (c *client) getHGetCompleted(key, field string) rueidis.Completed {
	return c.cmd.B().Hget().Key(key).Field(field).Build()
}

func (c *client) getHGetAllCompleted(key string) rueidis.Completed {
	return c.cmd.B().Hgetall().Key(key).Build()
}

func (c *client) getHKeysCompleted(key string) rueidis.Completed {
	return c.cmd.B().Hkeys().Key(key).Build()
}

func (c *client) getHLenCompleted(key string) rueidis.Completed {
	return c.cmd.B().Hlen().Key(key).Build()
}

func (c *client) getHMGetCompleted(key string, fields ...string) rueidis.Completed {
	return c.cmd.B().Hmget().Key(key).Field(fields...).Build()
}

func (c *client) hmset(ctx context.Context, key string, values ...interface{}) BoolCmd {
	fv := c.cmd.B().Hset().Key(key).FieldValue()
	args := argsToSlice(values)
	for i := 0; i < len(args); i += 2 {
		fv = fv.FieldValue(args[i], args[i+1])
	}
	return newBoolCmdFromResult(c.cmd.Do(ctx, fv.Build()))
}

func (c *client) hrandField(ctx context.Context, key string, count int, withValues bool) StringSliceCmd {
	h := c.cmd.B().Hrandfield().Key(key).Count(int64(count))
	if withValues {
		return flattenStringSliceCmd(c.cmd.Do(ctx, h.Withvalues().Build()))
	}
	return newStringSliceCmdFromResult(c.cmd.Do(ctx, h.Build()))
}

func (c *client) hscan(ctx context.Context, key string, cursor uint64, match string, count int64) ScanCmd {
	cmd := c.cmd.B().Arbitrary(HSCAN).Keys(key).Args(str(int64(cursor)))
	if match != "" {
		cmd = cmd.Args(MATCH, match)
	}
	if count > 0 {
		cmd = cmd.Args(COUNT, str(count))
	}
	return newScanCmdFromResult(c.cmd.Do(ctx, cmd.ReadOnly()))
}

func (c *client) hset(ctx context.Context, key string, values ...interface{}) IntCmd {
	fv := c.cmd.B().Hset().Key(key).FieldValue()
	args := argsToSlice(values)
	for i := 0; i < len(args); i += 2 {
		fv = fv.FieldValue(args[i], args[i+1])
	}
	return newIntCmdFromResult(c.cmd.Do(ctx, fv.Build()))
}

func (c *client) getLIndexCompleted(key string, index int64) rueidis.Completed {
	return c.cmd.B().Lindex().Key(key).Index(index).Build()
}

func (c *client) linsert(ctx context.Context, key, op string, pivot, value interface{}) IntCmd {
	var linsertKey = c.cmd.B().Linsert().Key(key)
	switch strings.ToUpper(op) {
	case BEFORE:
		return newIntCmdFromResult(c.cmd.Do(ctx, linsertKey.Before().Pivot(str(pivot)).Element(str(value)).Build()))
	case AFTER:
		return newIntCmdFromResult(c.cmd.Do(ctx, linsertKey.After().Pivot(str(pivot)).Element(str(value)).Build()))
	default:
		panic(fmt.Sprintf("Invalid op argument value: %s", op))
	}
}

func (c *client) getLLenCompleted(key string) rueidis.Completed {
	return c.cmd.B().Llen().Key(key).Build()
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

func (c *client) memoryUsage(ctx context.Context, key string, samples ...int) IntCmd {
	var memoryUsageKey = c.cmd.B().MemoryUsage().Key(key)
	switch len(samples) {
	case 0:
		return newIntCmdFromResult(c.cmd.Do(ctx, memoryUsageKey.Build()))
	case 1:
		return newIntCmdFromResult(c.cmd.Do(ctx, memoryUsageKey.Samples(int64(samples[0])).Build()))
	default:
		panic(errMemoryUsageArgsCount)
	}
}

func (c *client) sadd(ctx context.Context, key string, members ...interface{}) IntCmd {
	cmd := c.cmd.B().Sadd().Key(key).Member()
	for _, m := range argsToSlice(members) {
		cmd = cmd.Member(str(m))
	}
	return newIntCmdFromResult(c.cmd.Do(ctx, cmd.Build()))
}

func (c *client) getSCardCompleted(key string) rueidis.Completed {
	return c.cmd.B().Scard().Key(key).Build()
}

func (c *client) getSIsMemberCompleted(key string, member interface{}) rueidis.Completed {
	return c.cmd.B().Sismember().Key(key).Member(str(member)).Build()
}

func (c *client) getSMIsMemberCompleted(key string, members ...interface{}) rueidis.Completed {
	return c.cmd.B().Smismember().Key(key).Member(argsToSlice(members)...).Build()
}

func (c *client) getSMembersCompleted(key string) rueidis.Completed {
	return c.cmd.B().Smembers().Key(key).Build()
}

func (c *client) sscan(ctx context.Context, key string, cursor uint64, match string, count int64) ScanCmd {
	cmd := c.cmd.B().Arbitrary(SSCAN).Keys(key).Args(str(int64(cursor)))
	if match != "" {
		cmd = cmd.Args(MATCH, match)
	}
	if count > 0 {
		cmd = cmd.Args(COUNT, str(count))
	}
	return newScanCmdFromResult(c.cmd.Do(ctx, cmd.ReadOnly()))
}

func (c *client) zadd(ctx context.Context, key string, members ...Z) IntCmd {
	cmd := c.cmd.B().Zadd().Key(key).ScoreMember()
	for _, v := range members {
		cmd = cmd.ScoreMember(v.Score, str(v.Member))
	}
	return newIntCmdFromResult(c.cmd.Do(ctx, cmd.Build()))
}

func (c *client) zaddNX(ctx context.Context, key string, members ...Z) IntCmd {
	cmd := c.cmd.B().Zadd().Key(key).Nx().ScoreMember()
	for _, v := range members {
		cmd = cmd.ScoreMember(v.Score, str(v.Member))
	}
	return newIntCmdFromResult(c.cmd.Do(ctx, cmd.Build()))
}

func (c *client) zaddXX(ctx context.Context, key string, members ...Z) IntCmd {
	cmd := c.cmd.B().Zadd().Key(key).Xx().ScoreMember()
	for _, v := range members {
		cmd = cmd.ScoreMember(v.Score, str(v.Member))
	}
	return newIntCmdFromResult(c.cmd.Do(ctx, cmd.Build()))
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

func (c *client) getZCardCompleted(key string) rueidis.Completed {
	return c.cmd.B().Zcard().Key(key).Build()
}

func (c *client) getZCount(key, min, max string) rueidis.Completed {
	return c.cmd.B().Zcount().Key(key).Min(min).Max(max).Build()
}

func (c *client) getZLexCountCompleted(key, min, max string) rueidis.Completed {
	return c.cmd.B().Zlexcount().Key(key).Min(min).Max(max).Build()
}

func (c *client) getZMScoreCompleted(key string, members ...string) rueidis.Completed {
	return c.cmd.B().Zmscore().Key(key).Member(members...).Build()
}

func (c *client) zpopMax(ctx context.Context, key string, count ...int64) ZSliceCmd {
	var resp rueidis.RedisResult
	var zpopmaxKey = c.cmd.B().Zpopmax().Key(key)
	switch len(count) {
	case 0:
		resp = c.cmd.Do(ctx, zpopmaxKey.Build())
	case 1:
		resp = c.cmd.Do(ctx, zpopmaxKey.Count(count[0]).Build())
		if count[0] > 1 {
			return newZSliceCmdFromResult(resp)
		}
	default:
		panic(errTooManyArguments)
	}
	return newZSliceSingleCmdFromResult(resp)
}

func (c *client) zpopMin(ctx context.Context, key string, count ...int64) ZSliceCmd {
	var resp rueidis.RedisResult
	var zpopminKey = c.cmd.B().Zpopmin().Key(key)
	switch len(count) {
	case 0:
		resp = c.cmd.Do(ctx, zpopminKey.Build())
	case 1:
		resp = c.cmd.Do(ctx, zpopminKey.Count(count[0]).Build())
		if count[0] > 1 {
			return newZSliceCmdFromResult(resp)
		}
	default:
		panic(errTooManyArguments)
	}
	return newZSliceSingleCmdFromResult(resp)
}

func (c *client) zrandMember(ctx context.Context, key string, count int, withScores bool) StringSliceCmd {
	var zrandmemberOptionsCount = c.cmd.B().Zrandmember().Key(key).Count(int64(count))
	if withScores {
		return flattenStringSliceCmd(c.cmd.Do(ctx, zrandmemberOptionsCount.Withscores().Build()))
	}
	return newStringSliceCmdFromResult(c.cmd.Do(ctx, zrandmemberOptionsCount.Build()))
}

func (c *client) getZRangeCompleted(key string, start, stop int64) rueidis.Completed {
	return c.cmd.B().Zrange().Key(key).Min(str(start)).Max(str(stop)).Build()
}

func (c *client) getZRangeWithScoresCompleted(key string, start, stop int64) rueidis.Completed {
	return c.cmd.B().Zrange().Key(key).Min(str(start)).Max(str(stop)).Withscores().Build()
}

func (c *client) getZRangeByLexCompleted(key string, opt ZRangeBy) rueidis.Completed {
	var completed rueidis.Completed
	var zrangebylexMax = c.cmd.B().Zrangebylex().Key(key).Min(opt.Min).Max(opt.Max)
	if opt.Offset != 0 || opt.Count != 0 {
		completed = zrangebylexMax.Limit(opt.Offset, opt.Count).Build()
	} else {
		completed = zrangebylexMax.Build()
	}
	return completed
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

func (c *client) getZRankCompleted(key, member string) rueidis.Completed {
	return c.cmd.B().Zrank().Key(key).Member(member).Build()
}

func (c *client) getZRevRangeCompleted(key string, start, stop int64, withScore bool) rueidis.Completed {
	var zrevrangeStop = c.cmd.B().Zrevrange().Key(key).Start(start).Stop(stop)
	if withScore {
		return zrevrangeStop.Withscores().Build()
	}
	return zrevrangeStop.Build()
}

func (c *client) getZRevRangeByLexCompleted(key string, opt ZRangeBy) rueidis.Completed {
	var completed rueidis.Completed
	var zrevrangebylexMin = c.cmd.B().Zrevrangebylex().Key(key).Max(opt.Max).Min(opt.Min)
	if opt.Offset != 0 || opt.Count != 0 {
		completed = zrevrangebylexMin.Limit(opt.Offset, opt.Count).Build()
	} else {
		completed = zrevrangebylexMin.Build()
	}
	return completed
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

func (c *client) getZRevRankCompleted(key, member string) rueidis.Completed {
	return c.cmd.B().Zrevrank().Key(key).Member(member).Build()
}

func (c *client) zscan(ctx context.Context, key string, cursor uint64, match string, count int64) ScanCmd {
	cmd := c.cmd.B().Arbitrary(ZSCAN).Keys(key).Args(str(cursor))
	if match != "" {
		cmd = cmd.Args(MATCH, match)
	}
	if count > 0 {
		cmd = cmd.Args(COUNT, str(count))
	}
	return newScanCmdFromResult(c.cmd.Do(ctx, cmd.ReadOnly()))
}

func (c *client) getZScoreCompleted(key, member string) rueidis.Completed {
	return c.cmd.B().Zscore().Key(key).Member(member).Build()
}

func (c *client) xadd(ctx context.Context, a XAddArgs) StringCmd {
	cmd := c.cmd.B().Arbitrary(XADD).Keys(a.Stream)
	if a.NoMkStream {
		cmd = cmd.Args(NOMKSTREAM)
	}
	switch {
	case a.MaxLen > 0:
		if a.Approx {
			cmd = cmd.Args(MAXLEN, "~", str(a.MaxLen))
		} else {
			cmd = cmd.Args(MAXLEN, str(a.MaxLen))
		}
	case len(a.MinID) > 0:
		if a.Approx {
			cmd = cmd.Args(MINID, "~", a.MinID)
		} else {
			cmd = cmd.Args(MINID, a.MinID)
		}
	}
	if a.Limit > 0 {
		cmd = cmd.Args(LIMIT, str(a.Limit))
	}
	if len(a.ID) > 0 {
		cmd = cmd.Args(a.ID)
	} else {
		cmd = cmd.Args("*")
	}
	return newStringCmdFromResult(c.cmd.Do(ctx, cmd.Args(argToSlice(a.Values)...).Build()))
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

func (c *client) xpendingExt(ctx context.Context, a XPendingExtArgs) XPendingExtCmd {
	cmd := c.cmd.B().Arbitrary(XPENDING).Keys(a.Stream).Args(a.Group)
	if a.Idle != 0 {
		cmd = cmd.Args(IDLE, str(formatMs(a.Idle)))
	}
	cmd = cmd.Args(a.Start, a.End, str(a.Count))
	if len(a.Consumer) > 0 {
		cmd = cmd.Args(a.Consumer)
	}
	return newXPendingExtCmdFromResult(c.cmd.Do(ctx, cmd.Build()))
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

func (c *client) getGetCompleted(key string) rueidis.Completed {
	return c.cmd.B().Get().Key(key).Build()
}

func (c *client) getEx(ctx context.Context, key string, expiration time.Duration) StringCmd {
	var completed rueidis.Completed
	var getexKey = c.cmd.B().Getex().Key(key)
	if expiration > 0 {
		if usePrecise(expiration) {
			completed = getexKey.PxMilliseconds(formatMs(expiration)).Build()
		} else {
			completed = getexKey.ExSeconds(formatSec(expiration)).Build()
		}
	} else if expiration == 0 {
		completed = getexKey.Persist().Build()
	} else {
		completed = getexKey.Build()
	}
	return newStringCmdFromResult(c.cmd.Do(ctx, completed))
}

func (c *client) getGetRangeCompleted(key string, start, end int64) rueidis.Completed {
	return c.cmd.B().Getrange().Key(key).Start(start).End(end).Build()
}

func (c *client) getMGetCompleted(keys ...string) rueidis.Completed {
	return c.cmd.B().Mget().Key(keys...).Build()
}

func (c *client) mset(ctx context.Context, values ...interface{}) StatusCmd {
	kv := c.cmd.B().Mset().KeyValue()
	args := argsToSlice(values)
	for i := 0; i < len(args); i += 2 {
		kv = kv.KeyValue(args[i], args[i+1])
	}
	return newStatusCmdFromResult(c.cmd.Do(ctx, kv.Build()))
}

func (c *client) msetNX(ctx context.Context, values ...interface{}) BoolCmd {
	kv := c.cmd.B().Msetnx().KeyValue()
	args := argsToSlice(values)
	for i := 0; i < len(args); i += 2 {
		kv = kv.KeyValue(args[i], args[i+1])
	}
	return newBoolCmdFromResult(c.cmd.Do(ctx, kv.Build()))
}

func (c *client) set(ctx context.Context, key string, value interface{}, expiration time.Duration) StatusCmd {
	var completed rueidis.Completed
	var setValue = c.cmd.B().Set().Key(key).Value(str(value))
	if expiration > 0 {
		if usePrecise(expiration) {
			completed = setValue.PxMilliseconds(formatMs(expiration)).Build()
		} else {
			completed = setValue.ExSeconds(formatSec(expiration)).Build()
		}
	} else if expiration == KeepTTL {
		completed = setValue.Keepttl().Build()
	} else {
		completed = setValue.Build()
	}
	return newStatusCmdFromResult(c.cmd.Do(ctx, completed))
}

func (c *client) setNX(ctx context.Context, key string, value interface{}, expiration time.Duration) BoolCmd {
	var resp rueidis.RedisResult
	switch expiration {
	case 0:
		resp = c.cmd.Do(ctx, c.cmd.B().Setnx().Key(key).Value(str(value)).Build())
	case KeepTTL:
		resp = c.cmd.Do(ctx, c.cmd.B().Set().Key(key).Value(str(value)).Nx().Keepttl().Build())
	default:
		if usePrecise(expiration) {
			resp = c.cmd.Do(ctx, c.cmd.B().Set().Key(key).Value(str(value)).Nx().PxMilliseconds(formatMs(expiration)).Build())
		} else {
			resp = c.cmd.Do(ctx, c.cmd.B().Set().Key(key).Value(str(value)).Nx().ExSeconds(formatSec(expiration)).Build())
		}
	}
	return newBoolCmdFromResult(resp)
}

func (c *client) setXX(ctx context.Context, key string, value interface{}, expiration time.Duration) BoolCmd {
	var resp rueidis.RedisResult
	switch expiration {
	case 0:
		resp = c.cmd.Do(ctx, c.cmd.B().Set().Key(key).Value(str(value)).Xx().Build())
	case KeepTTL:
		resp = c.cmd.Do(ctx, c.cmd.B().Set().Key(key).Value(str(value)).Xx().Keepttl().Build())
	default:
		if usePrecise(expiration) {
			resp = c.cmd.Do(ctx, c.cmd.B().Set().Key(key).Value(str(value)).Xx().PxMilliseconds(formatMs(expiration)).Build())
		} else {
			resp = c.cmd.Do(ctx, c.cmd.B().Set().Key(key).Value(str(value)).Xx().ExSeconds(formatSec(expiration)).Build())
		}
	}
	return newBoolCmdFromResult(resp)
}

func (c *client) setArgs(ctx context.Context, key string, value interface{}, a SetArgs) StatusCmd {
	cmd := c.cmd.B().Arbitrary(SET).Keys(key).Args(str(value))
	if a.KeepTTL {
		cmd = cmd.Args(KEEPTTL)
	}
	if !a.ExpireAt.IsZero() {
		cmd = cmd.Args(EXAT, str(a.ExpireAt.Unix()))
	}
	if a.TTL > 0 {
		if usePrecise(a.TTL) {
			cmd = cmd.Args(PX, str(formatMs(a.TTL)))
		} else {
			cmd = cmd.Args(EX, str(formatSec(a.TTL)))
		}
	}
	if len(a.Mode) > 0 {
		cmd = cmd.Args(a.Mode)
	}
	if a.Get {
		cmd = cmd.Args(GET)
	}
	return newStatusCmdFromResult(c.cmd.Do(ctx, cmd.Build()))
}

func (c *client) getStrLenCompleted(key string) rueidis.Completed {
	return c.cmd.B().Strlen().Key(key).Build()
}
func (r *client) getHValsCompleted(key string) rueidis.Completed {
	return r.cmd.B().Hvals().Key(key).Build()
}
func (r *client) getLRangeCompleted(key string, start, stop int64) rueidis.Completed {
	return r.cmd.B().Lrange().Key(key).Start(start).Stop(stop).Build()
}
