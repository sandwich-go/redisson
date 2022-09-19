package redisson

import (
	"context"
	"fmt"
	"github.com/sandwich-go/rueidis"
	"github.com/sandwich-go/rueidis/rueidiscompat"
	"strings"
	"sync"
	"time"
)

type resp3 struct {
	v       ConfVisitor
	cmd     rueidis.Client
	adapter rueidiscompat.Cmdable
	handler handler
}

type resp3Cache struct {
	ttl  time.Duration
	resp *resp3
}

func connectResp3(v ConfVisitor, h handler) (*resp3, error) {
	cmd, err := rueidis.NewClient(rueidis.ClientOption{
		Username:          v.GetUsername(),
		Password:          v.GetPassword(),
		InitAddress:       v.GetAddrs(),
		SelectDB:          v.GetDB(),
		CacheSizeEachConn: v.GetCacheSizeEachConn(),
		RingScaleEachConn: v.GetRingScaleEachConn(),
		BlockingPoolSize:  v.GetConnPoolSize(),
		ConnWriteTimeout:  v.GetWriteTimeout(),
		ShuffleInit:       true,
		Sentinel: rueidis.SentinelOption{
			Username:   v.GetUsername(),
			Password:   v.GetPassword(),
			ClientName: v.GetName(),
			MasterSet:  v.GetMasterName(),
		},
	})
	if err != nil {
		return nil, err
	}
	return &resp3{cmd: cmd, v: v, handler: h, adapter: rueidiscompat.NewAdapter(cmd)}, nil
}

func (r *resp3) PoolStats() PoolStats                    { return PoolStats{} }
func (r *resp3) Close() error                            { r.cmd.Close(); return nil }
func (r *resp3) RegisterCollector(RegisterCollectorFunc) {}
func (r *resp3) Cache(ttl time.Duration) CacheCmdable {
	if r.v.GetEnableCache() {
		return &resp3Cache{resp: r, ttl: ttl}
	}
	return r
}
func (r *resp3Cache) Do(ctx context.Context, completed rueidis.Completed) rueidis.RedisResult {
	rsp := r.resp.cmd.DoCache(ctx, rueidis.Cacheable(completed), r.ttl)
	r.resp.handler.cache(ctx, rsp.IsCacheHit())
	return rsp
}

func (r *resp3) getBitCountCompleted(key string, bitCount *BitCount) rueidis.Completed {
	var bitCountKey = r.cmd.B().Bitcount().Key(key)
	if bitCount != nil {
		return bitCountKey.Start(bitCount.Start).End(bitCount.End).Build()
	}
	return bitCountKey.Build()
}

func (r *resp3) BitCount(ctx context.Context, key string, bitCount *BitCount) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.getBitCountCompleted(key, bitCount)))
}

func (r *resp3Cache) BitCount(ctx context.Context, key string, bitCount *BitCount) IntCmd {
	return newIntCmd(r.Do(ctx, r.resp.getBitCountCompleted(key, bitCount)))
}

func (r *resp3) BitField(ctx context.Context, key string, args ...interface{}) IntSliceCmd {
	return newIntSliceCmd(r.cmd.Do(ctx, r.cmd.B().Arbitrary(BITFIELD).Keys(key).Args(argsToSlice(args)...).Build()))
}

func (r *resp3) bitOp(ctx context.Context, token, destKey string, keys ...string) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Bitop().Operation(token).Destkey(destKey).Key(keys...).Build()))
}

func (r *resp3) BitOpAnd(ctx context.Context, destKey string, keys ...string) IntCmd {
	return r.bitOp(ctx, AND, destKey, keys...)
}

func (r *resp3) BitOpOr(ctx context.Context, destKey string, keys ...string) IntCmd {
	return r.bitOp(ctx, OR, destKey, keys...)
}

func (r *resp3) BitOpXor(ctx context.Context, destKey string, keys ...string) IntCmd {
	return r.bitOp(ctx, XOR, destKey, keys...)
}

func (r *resp3) BitOpNot(ctx context.Context, destKey string, key string) IntCmd {
	return r.bitOp(ctx, NOT, destKey, key)
}

func (r *resp3) getBitPosCompleted(key string, bit int64, pos ...int64) rueidis.Completed {
	var completed rueidis.Completed
	var bitposBit = r.cmd.B().Bitpos().Key(key).Bit(bit)
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

func (r *resp3) BitPos(ctx context.Context, key string, bit int64, pos ...int64) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.getBitPosCompleted(key, bit, pos...)))
}

func (r *resp3Cache) BitPos(ctx context.Context, key string, bit int64, pos ...int64) IntCmd {
	return newIntCmd(r.Do(ctx, r.resp.getBitPosCompleted(key, bit, pos...)))
}

func (r *resp3) getBitCompleted(key string, offset int64) rueidis.Completed {
	return r.cmd.B().Getbit().Key(key).Offset(offset).Build()
}

func (r *resp3) GetBit(ctx context.Context, key string, offset int64) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.getBitCompleted(key, offset)))
}

func (r *resp3Cache) GetBit(ctx context.Context, key string, offset int64) IntCmd {
	return newIntCmd(r.Do(ctx, r.resp.getBitCompleted(key, offset)))
}

func (r *resp3) SetBit(ctx context.Context, key string, offset int64, value int) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Setbit().Key(key).Offset(offset).Value(int64(value)).Build()))
}

func (r *resp3) ClusterAddSlots(ctx context.Context, slots ...int) StatusCmd {
	return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().ClusterAddslots().Slot(intSliceToInt64ToSlice(slots)...).Build()))
}

func (r *resp3) ClusterAddSlotsRange(ctx context.Context, min, max int) StatusCmd {
	return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().ClusterAddslotsrange().StartSlotEndSlot().StartSlotEndSlot(int64(min), int64(max)).Build()))
}

func (r *resp3) ClusterCountFailureReports(ctx context.Context, nodeID string) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().ClusterCountFailureReports().NodeId(nodeID).Build()))
}

func (r *resp3) ClusterCountKeysInSlot(ctx context.Context, slot int) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().ClusterCountkeysinslot().Slot(int64(slot)).Build()))
}

func (r *resp3) ClusterDelSlots(ctx context.Context, slots ...int) StatusCmd {
	return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().ClusterDelslots().Slot(intSliceToInt64ToSlice(slots)...).Build()))
}

func (r *resp3) ClusterDelSlotsRange(ctx context.Context, min, max int) StatusCmd {
	return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().ClusterDelslotsrange().StartSlotEndSlot().StartSlotEndSlot(int64(min), int64(max)).Build()))
}

func (r *resp3) ClusterFailover(ctx context.Context) StatusCmd {
	return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().ClusterFailover().Build()))
}

func (r *resp3) ClusterForget(ctx context.Context, nodeID string) StatusCmd {
	return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().ClusterForget().NodeId(nodeID).Build()))
}

func (r *resp3) ClusterGetKeysInSlot(ctx context.Context, slot int, count int) StringSliceCmd {
	return newStringSliceCmd(r.cmd.Do(ctx, r.cmd.B().ClusterGetkeysinslot().Slot(int64(slot)).Count(int64(count)).Build()))
}

func (r *resp3) ClusterInfo(ctx context.Context) StringCmd {
	return newStringCmd(r.cmd.Do(ctx, r.cmd.B().ClusterInfo().Build()))
}

func (r *resp3) ClusterKeySlot(ctx context.Context, key string) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().ClusterKeyslot().Key(key).Build()))
}

func (r *resp3) ClusterMeet(ctx context.Context, host, port string) StatusCmd {
	iport, err := parseInt(port)
	if err != nil {
		return newStatusCmdWithError(err)
	}
	return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().ClusterMeet().Ip(host).Port(iport).Build()))
}

func (r *resp3) ClusterNodes(ctx context.Context) StringCmd {
	return newStringCmd(r.cmd.Do(ctx, r.cmd.B().ClusterNodes().Build()))
}

func (r *resp3) ClusterReplicate(ctx context.Context, nodeID string) StatusCmd {
	return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().ClusterReplicate().NodeId(nodeID).Build()))
}

func (r *resp3) ClusterResetSoft(ctx context.Context) StatusCmd {
	//return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().ClusterReset().Soft().Build()))
	return r.adapter.ClusterResetSoft(ctx)
}

func (r *resp3) ClusterResetHard(ctx context.Context) StatusCmd {
	//return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().ClusterReset().Hard().Build()))
	return r.adapter.ClusterResetHard(ctx)
}

func (r *resp3) ClusterSaveConfig(ctx context.Context) StatusCmd {
	//return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().ClusterSaveconfig().Build()))
	return r.adapter.ClusterSaveConfig(ctx)
}

func (r *resp3) ClusterSlaves(ctx context.Context, nodeID string) StringSliceCmd {
	return newStringSliceCmd(r.cmd.Do(ctx, r.cmd.B().ClusterSlaves().NodeId(nodeID).Build()))
}

func (r *resp3) ClusterSlots(ctx context.Context) ClusterSlotsCmd {
	return newClusterSlotsCmd(r.cmd.Do(ctx, r.cmd.B().ClusterSlots().Build()))
}

func (r *resp3) ReadOnly(ctx context.Context) StatusCmd {
	return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().Readonly().Build()))
}

func (r *resp3) ReadWrite(ctx context.Context) StatusCmd {
	return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().Readwrite().Build()))
}

func (r *resp3) Select(ctx context.Context, index int) StatusCmd {
	return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().Select().Index(int64(index)).Build()))
}

func (r *resp3) ClientGetName(ctx context.Context) StringCmd {
	return newStringCmd(r.cmd.Do(ctx, r.cmd.B().ClientGetname().Build()))
}

func (r *resp3) ClientID(ctx context.Context) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().ClientId().Build()))
}

func (r *resp3) ClientKill(ctx context.Context, ipPort string) StatusCmd {
	//return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().Arbitrary(CLIENT).Args(KILL).Args(ipPort).Build()))
	return r.adapter.ClientKill(ctx, ipPort)
}

func (r *resp3) ClientKillByFilter(ctx context.Context, keys ...string) IntCmd {
	//return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Arbitrary(CLIENT).Args(KILL).Args(keys...).Build()))
	return r.adapter.ClientKillByFilter(ctx, keys...)
}

func (r *resp3) ClientList(ctx context.Context) StringCmd {
	return newStringCmd(r.cmd.Do(ctx, r.cmd.B().ClientList().Build()))
}

func (r *resp3) ClientPause(ctx context.Context, dur time.Duration) BoolCmd {
	//return newBoolCmd(r.cmd.Do(ctx, r.cmd.B().ClientPause().Timeout(formatSec(dur)).Build()))
	return r.adapter.ClientPause(ctx, dur)
}

func (r *resp3) Echo(ctx context.Context, message interface{}) StringCmd {
	return newStringCmd(r.cmd.Do(ctx, r.cmd.B().Echo().Message(str(message)).Build()))
}

func (r *resp3) Ping(ctx context.Context) StatusCmd {
	return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().Ping().Build()))
}

func (r *resp3) Quit(ctx context.Context) StatusCmd {
	return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().Quit().Build()))
}

func (r *resp3) Copy(ctx context.Context, sourceKey string, destKey string, db int, replace bool) IntCmd {
	var completed rueidis.Completed
	var cmd = r.cmd.B().Copy().Source(sourceKey).Destination(destKey).Db(int64(db))
	if replace {
		completed = cmd.Replace().Build()
	} else {
		completed = cmd.Build()
	}
	return newIntCmd(r.cmd.Do(ctx, completed))
}

func (r *resp3) Del(ctx context.Context, keys ...string) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Del().Key(keys...).Build()))
}

func (r *resp3) Dump(ctx context.Context, key string) StringCmd {
	return newStringCmd(r.cmd.Do(ctx, r.cmd.B().Dump().Key(key).Build()))
}

func (r *resp3) getExistsCompleted(keys ...string) rueidis.Completed {
	return r.cmd.B().Exists().Key(keys...).Build()
}

func (r *resp3) Exists(ctx context.Context, keys ...string) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.getExistsCompleted(keys...)))
}

func (r *resp3Cache) Exists(ctx context.Context, keys ...string) IntCmd {
	return newIntCmd(r.Do(ctx, r.resp.getExistsCompleted(keys...)))
}

func (r *resp3) Expire(ctx context.Context, key string, expiration time.Duration) BoolCmd {
	return newBoolCmd(r.cmd.Do(ctx, r.cmd.B().Expire().Key(key).Seconds(formatSec(expiration)).Build()))
}

func (r *resp3) ExpireAt(ctx context.Context, key string, tm time.Time) BoolCmd {
	return newBoolCmd(r.cmd.Do(ctx, r.cmd.B().Expireat().Key(key).Timestamp(tm.Unix()).Build()))
}

func (r *resp3) Keys(ctx context.Context, pattern string) StringSliceCmd {
	//return newStringSliceCmd(r.cmd.Do(ctx, r.cmd.B().Keys().Pattern(pattern).Build()))
	return newStringSliceCmdFromStringSliceCmd(r.adapter.Keys(ctx, pattern))
}

func (r *resp3) Migrate(ctx context.Context, host, port, key string, db int, timeout time.Duration) StatusCmd {
	iport, err := parseInt(port)
	if err != nil {
		return newStatusCmdWithError(err)
	}
	var migratePort = r.cmd.B().Migrate().Host(host).Port(iport)
	if len(key) > 0 {
		return newStatusCmd(r.cmd.Do(ctx, migratePort.Key().DestinationDb(int64(db)).Timeout(formatSec(timeout)).Build()))
	}
	return newStatusCmd(r.cmd.Do(ctx, migratePort.Empty().DestinationDb(int64(db)).Timeout(formatSec(timeout)).Build()))
}

func (r *resp3) Move(ctx context.Context, key string, db int) BoolCmd {
	return newBoolCmd(r.cmd.Do(ctx, r.cmd.B().Move().Key(key).Db(int64(db)).Build()))
}

func (r *resp3) ObjectRefCount(ctx context.Context, key string) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().ObjectRefcount().Key(key).Build()))
}

func (r *resp3) ObjectEncoding(ctx context.Context, key string) StringCmd {
	return newStringCmd(r.cmd.Do(ctx, r.cmd.B().ObjectEncoding().Key(key).Build()))
}

func (r *resp3) ObjectIdleTime(ctx context.Context, key string) DurationCmd {
	return newDurationCmd(r.cmd.Do(ctx, r.cmd.B().ObjectIdletime().Key(key).Build()), time.Second)
}

func (r *resp3) Persist(ctx context.Context, key string) BoolCmd {
	return newBoolCmd(r.cmd.Do(ctx, r.cmd.B().Persist().Key(key).Build()))
}

func (r *resp3) PExpire(ctx context.Context, key string, expiration time.Duration) BoolCmd {
	return newBoolCmd(r.cmd.Do(ctx, r.cmd.B().Pexpire().Key(key).Milliseconds(formatMs(expiration)).Build()))
}

func (r *resp3) PExpireAt(ctx context.Context, key string, tm time.Time) BoolCmd {
	return newBoolCmd(r.cmd.Do(ctx, r.cmd.B().Pexpireat().Key(key).MillisecondsTimestamp(tm.UnixNano()/int64(time.Millisecond)).Build()))
}

func (r *resp3) getPTTLCompleted(key string) rueidis.Completed {
	return r.cmd.B().Pttl().Key(key).Build()
}

func (r *resp3) PTTL(ctx context.Context, key string) DurationCmd {
	return newDurationCmd(r.cmd.Do(ctx, r.getPTTLCompleted(key)), time.Millisecond)
}

func (r *resp3Cache) PTTL(ctx context.Context, key string) DurationCmd {
	return newDurationCmd(r.Do(ctx, r.resp.getPTTLCompleted(key)), time.Millisecond)
}

func (r *resp3) RandomKey(ctx context.Context) StringCmd {
	//	return newStringCmd(r.cmd.Do(ctx, r.cmd.B().Randomkey().Build()))
	return newStringCmdFromStringCmd(r.adapter.RandomKey(ctx))
}

func (r *resp3) Rename(ctx context.Context, key, newkey string) StatusCmd {
	return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().Rename().Key(key).Newkey(newkey).Build()))
}

func (r *resp3) RenameNX(ctx context.Context, key, newkey string) BoolCmd {
	return newBoolCmd(r.cmd.Do(ctx, r.cmd.B().Renamenx().Key(key).Newkey(newkey).Build()))
}

func (r *resp3) Restore(ctx context.Context, key string, ttl time.Duration, value string) StatusCmd {
	return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().Restore().Key(key).Ttl(formatMs(ttl)).SerializedValue(value).Build()))
}

func (r *resp3) RestoreReplace(ctx context.Context, key string, ttl time.Duration, value string) StatusCmd {
	return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().Restore().Key(key).Ttl(formatMs(ttl)).SerializedValue(value).Replace().Build()))
}

func (r *resp3) getScanArgs(cursor uint64, match string, count int64) rueidis.Arbitrary {
	cmd := r.cmd.B().Arbitrary(SCAN, str(cursor))
	if len(match) > 0 {
		cmd = cmd.Args(MATCH, match)
	}
	if count > 0 {
		cmd = cmd.Args(COUNT, str(count))
	}
	return cmd
}

func (r *resp3) Scan(ctx context.Context, cursor uint64, match string, count int64) ScanCmd {
	return newScanCmd(r.cmd.Do(ctx, r.getScanArgs(cursor, match, count).ReadOnly()))
}

func (r *resp3) ScanType(ctx context.Context, cursor uint64, match string, count int64, keyType string) ScanCmd {
	args := r.getScanArgs(cursor, match, count)
	if len(keyType) > 0 {
		args = args.Args(TYPE, keyType)
	}
	return newScanCmd(r.cmd.Do(ctx, args.ReadOnly()))
}

func (r *resp3) getSortArgs(key string, sort Sort) rueidis.Arbitrary {
	arbitrary := r.cmd.B().Arbitrary(SORT).Keys(key)
	if len(sort.By) > 0 {
		arbitrary = arbitrary.Args(BY, sort.By)
	}
	if sort.Offset != 0 || sort.Count != 0 {
		arbitrary = arbitrary.Args(LIMIT, str(sort.Offset), str(sort.Count))
	}
	for _, g := range sort.Get {
		arbitrary = arbitrary.Args(GET, g)
	}
	if len(sort.Order) > 0 {
		arbitrary = arbitrary.Args(sort.Order)
	}
	if sort.Alpha {
		arbitrary = arbitrary.Args(ALPHA)
	}
	return arbitrary
}

func (r *resp3) Sort(ctx context.Context, key string, sort Sort) StringSliceCmd {
	return newStringSliceCmd(r.cmd.Do(ctx, r.getSortArgs(key, sort).Build()))
}

func (r *resp3Cache) Sort(ctx context.Context, key string, sort Sort) StringSliceCmd {
	return newStringSliceCmd(r.Do(ctx, r.resp.getSortArgs(key, sort).Build()))
}

func (r *resp3) SortStore(ctx context.Context, key, store string, sort Sort) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.getSortArgs(key, sort).Args(STORE, store).Build()))
}

func (r *resp3) SortInterfaces(ctx context.Context, key string, sort Sort) SliceCmd {
	return newSliceCmd(r.cmd.Do(ctx, r.getSortArgs(key, sort).Build()))
}

func (r *resp3Cache) SortInterfaces(ctx context.Context, key string, sort Sort) SliceCmd {
	return newSliceCmd(r.Do(ctx, r.resp.getSortArgs(key, sort).Build()))
}

func (r *resp3) Touch(ctx context.Context, keys ...string) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Touch().Key(keys...).Build()))
}

func (r *resp3) getTTLCompleted(key string) rueidis.Completed {
	return r.cmd.B().Ttl().Key(key).Build()
}

func (r *resp3) TTL(ctx context.Context, key string) DurationCmd {
	return newDurationCmd(r.cmd.Do(ctx, r.getTTLCompleted(key)), time.Second)
}

func (r *resp3Cache) TTL(ctx context.Context, key string) DurationCmd {
	return newDurationCmd(r.Do(ctx, r.resp.getTTLCompleted(key)), time.Second)
}

func (r *resp3) getTypeCompleted(key string) rueidis.Completed {
	return r.cmd.B().Type().Key(key).Build()
}

func (r *resp3) Type(ctx context.Context, key string) StatusCmd {
	return newStatusCmd(r.cmd.Do(ctx, r.getTypeCompleted(key)))
}

func (r *resp3Cache) Type(ctx context.Context, key string) StatusCmd {
	return newStatusCmd(r.Do(ctx, r.resp.getTypeCompleted(key)))
}

func (r *resp3) Unlink(ctx context.Context, keys ...string) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Unlink().Key(keys...).Build()))
}

func (r *resp3) GeoAdd(ctx context.Context, key string, geoLocation ...GeoLocation) IntCmd {
	cmd := r.cmd.B().Geoadd().Key(key).LongitudeLatitudeMember()
	for _, loc := range geoLocation {
		cmd = cmd.LongitudeLatitudeMember(loc.Longitude, loc.Latitude, loc.Name)
	}
	return newIntCmd(r.cmd.Do(ctx, cmd.Build()))
}

func (r *resp3) getGeoDistCompleted(key string, member1, member2, unit string) rueidis.Completed {
	var completed rueidis.Completed
	var geodistMember2 = r.cmd.B().Geodist().Key(key).Member1(member1).Member2(member2)
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

func (r *resp3) GeoDist(ctx context.Context, key string, member1, member2, unit string) FloatCmd {
	return newFloatCmd(r.cmd.Do(ctx, r.getGeoDistCompleted(key, member1, member2, unit)))
}

func (r *resp3Cache) GeoDist(ctx context.Context, key string, member1, member2, unit string) FloatCmd {
	return newFloatCmd(r.Do(ctx, r.resp.getGeoDistCompleted(key, member1, member2, unit)))
}

func (r *resp3) getGeoHashCompleted(key string, members ...string) rueidis.Completed {
	return r.cmd.B().Geohash().Key(key).Member(members...).Build()
}

func (r *resp3) GeoHash(ctx context.Context, key string, members ...string) StringSliceCmd {
	return newStringSliceCmd(r.cmd.Do(ctx, r.getGeoHashCompleted(key, members...)))
}

func (r *resp3Cache) GeoHash(ctx context.Context, key string, members ...string) StringSliceCmd {
	return newStringSliceCmd(r.Do(ctx, r.resp.getGeoHashCompleted(key, members...)))
}

func (r *resp3) getGeoPosCompleted(key string, members ...string) rueidis.Completed {
	return r.cmd.B().Geopos().Key(key).Member(members...).Build()
}

func (r *resp3) GeoPos(ctx context.Context, key string, members ...string) GeoPosCmd {
	return newGeoPosCmd(r.cmd.Do(ctx, r.getGeoPosCompleted(key, members...)))
}

func (r *resp3Cache) GeoPos(ctx context.Context, key string, members ...string) GeoPosCmd {
	return newGeoPosCmd(r.Do(ctx, r.resp.getGeoPosCompleted(key, members...)))
}

func (r *resp3) getGeoRadiusCompleted(key string, longitude, latitude float64, q GeoRadiusQuery) rueidis.Completed {
	return r.cmd.B().Arbitrary(GEORADIUS_RO).Keys(key).Args(str(longitude), str(latitude)).Args(getGeoRadiusQueryArgs(q)...).Build()
}

func (r *resp3) GeoRadius(ctx context.Context, key string, longitude, latitude float64, q GeoRadiusQuery) GeoLocationCmd {
	if len(q.Store) > 0 || len(q.StoreDist) > 0 {
		return newGeoLocationCmdWithError(errGeoRadiusNotSupportStore)
	}
	return newGeoLocationCmd(r.cmd.Do(ctx, r.getGeoRadiusCompleted(key, longitude, latitude, q)), q)
}

func (r *resp3Cache) GeoRadius(ctx context.Context, key string, longitude, latitude float64, q GeoRadiusQuery) GeoLocationCmd {
	if len(q.Store) > 0 || len(q.StoreDist) > 0 {
		return newGeoLocationCmdWithError(errGeoRadiusNotSupportStore)
	}
	return newGeoLocationCmd(r.Do(ctx, r.resp.getGeoRadiusCompleted(key, longitude, latitude, q)), q)
}

func (r *resp3) GeoRadiusStore(ctx context.Context, key string, longitude, latitude float64, q GeoRadiusQuery) IntCmd {
	cmd := r.cmd.B().Arbitrary(GEORADIUS).Keys(key).Args(str(longitude), str(latitude))
	if len(q.Store) == 0 && len(q.StoreDist) == 0 {
		return newIntCmdWithError(errGeoRadiusStoreRequiresStore)
	}
	return newIntCmd(r.cmd.Do(ctx, cmd.Args(getGeoRadiusQueryArgs(q)...).Build()))
}

func (r *resp3) getGeoRadiusByMemberCompleted(key, member string, q GeoRadiusQuery) rueidis.Completed {
	return r.cmd.B().Arbitrary(GEORADIUSBYMEMBER_RO).Keys(key).Args(member).Args(getGeoRadiusQueryArgs(q)...).Build()
}

func (r *resp3) GeoRadiusByMember(ctx context.Context, key, member string, q GeoRadiusQuery) GeoLocationCmd {
	if len(q.Store) > 0 || len(q.StoreDist) > 0 {
		return newGeoLocationCmdWithError(errGeoRadiusByMemberNotSupportStore)
	}
	return newGeoLocationCmd(r.cmd.Do(ctx, r.getGeoRadiusByMemberCompleted(key, member, q)), q)
}

func (r *resp3Cache) GeoRadiusByMember(ctx context.Context, key, member string, q GeoRadiusQuery) GeoLocationCmd {
	if len(q.Store) > 0 || len(q.StoreDist) > 0 {
		return newGeoLocationCmdWithError(errGeoRadiusByMemberNotSupportStore)
	}
	return newGeoLocationCmd(r.Do(ctx, r.resp.getGeoRadiusByMemberCompleted(key, member, q)), q)
}

func (r *resp3) GeoRadiusByMemberStore(ctx context.Context, key, member string, q GeoRadiusQuery) IntCmd {
	cmd := r.cmd.B().Arbitrary(GEORADIUSBYMEMBER).Keys(key).Args(member)
	if len(q.Store) == 0 && len(q.StoreDist) == 0 {
		return newIntCmdWithError(errGeoRadiusByMemberStoreRequiresStore)
	}
	return newIntCmd(r.cmd.Do(ctx, cmd.Args(getGeoRadiusQueryArgs(q)...).Build()))
}

func (r *resp3) getGeoSearchCompleted(key string, q GeoSearchQuery) rueidis.Completed {
	return r.cmd.B().Arbitrary(GEOSEARCH).Keys(key).Args(getGeoSearchQueryArgs(q)...).Build()
}

func (r *resp3) GeoSearch(ctx context.Context, key string, q GeoSearchQuery) StringSliceCmd {
	return newStringSliceCmd(r.cmd.Do(ctx, r.getGeoSearchCompleted(key, q)))
}

func (r *resp3Cache) GeoSearch(ctx context.Context, key string, q GeoSearchQuery) StringSliceCmd {
	return newStringSliceCmd(r.Do(ctx, r.resp.getGeoSearchCompleted(key, q)))
}

func (r *resp3) getGeoSearchLocationCompleted(key string, q GeoSearchLocationQuery) rueidis.Completed {
	return r.cmd.B().Arbitrary(GEOSEARCH).Keys(key).Args(getGeoSearchLocationQueryArgs(q)...).Build()
}

func (r *resp3) GeoSearchLocation(ctx context.Context, key string, q GeoSearchLocationQuery) GeoSearchLocationCmd {
	return newGeoSearchLocationCmd(r.cmd.Do(ctx, r.getGeoSearchLocationCompleted(key, q)), q)
}

func (r *resp3Cache) GeoSearchLocation(ctx context.Context, key string, q GeoSearchLocationQuery) GeoSearchLocationCmd {
	return newGeoSearchLocationCmd(r.Do(ctx, r.resp.getGeoSearchLocationCompleted(key, q)), q)
}

func (r *resp3) GeoSearchStore(ctx context.Context, src, dest string, q GeoSearchStoreQuery) IntCmd {
	cmd := r.cmd.B().Arbitrary(GEOSEARCHSTORE).Keys(dest, src).Args(getGeoSearchQueryArgs(q.GeoSearchQuery)...)
	if q.StoreDist {
		cmd = cmd.Args(STOREDIST)
	}
	return newIntCmd(r.cmd.Do(ctx, cmd.Build()))
}

func (r *resp3) HDel(ctx context.Context, key string, fields ...string) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Hdel().Key(key).Field(fields...).Build()))
}

func (r *resp3) getHExistsCompleted(key, field string) rueidis.Completed {
	return r.cmd.B().Hexists().Key(key).Field(field).Build()
}

func (r *resp3) HExists(ctx context.Context, key, field string) BoolCmd {
	return newBoolCmd(r.cmd.Do(ctx, r.getHExistsCompleted(key, field)))
}

func (r *resp3Cache) HExists(ctx context.Context, key, field string) BoolCmd {
	return newBoolCmd(r.Do(ctx, r.resp.getHExistsCompleted(key, field)))
}

func (r *resp3) getHGetCompleted(key, field string) rueidis.Completed {
	return r.cmd.B().Hget().Key(key).Field(field).Build()
}

func (r *resp3) HGet(ctx context.Context, key, field string) StringCmd {
	return newStringCmd(r.cmd.Do(ctx, r.getHGetCompleted(key, field)))
}

func (r *resp3Cache) HGet(ctx context.Context, key, field string) StringCmd {
	return newStringCmd(r.Do(ctx, r.resp.getHGetCompleted(key, field)))
}

func (r *resp3) getHGetAllCompleted(key string) rueidis.Completed {
	return r.cmd.B().Hgetall().Key(key).Build()
}

func (r *resp3) HGetAll(ctx context.Context, key string) StringStringMapCmd {
	return newStringStringMapCmd(r.cmd.Do(ctx, r.getHGetAllCompleted(key)))
}

func (r *resp3Cache) HGetAll(ctx context.Context, key string) StringStringMapCmd {
	return newStringStringMapCmd(r.Do(ctx, r.resp.getHGetAllCompleted(key)))
}

func (r *resp3) HIncrBy(ctx context.Context, key, field string, incr int64) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Hincrby().Key(key).Field(field).Increment(incr).Build()))
}

func (r *resp3) HIncrByFloat(ctx context.Context, key, field string, incr float64) FloatCmd {
	return newFloatCmd(r.cmd.Do(ctx, r.cmd.B().Hincrbyfloat().Key(key).Field(field).Increment(incr).Build()))
}

func (r *resp3) getHKeysCompleted(key string) rueidis.Completed {
	return r.cmd.B().Hkeys().Key(key).Build()
}

func (r *resp3) HKeys(ctx context.Context, key string) StringSliceCmd {
	return newStringSliceCmd(r.cmd.Do(ctx, r.getHKeysCompleted(key)))
}

func (r *resp3Cache) HKeys(ctx context.Context, key string) StringSliceCmd {
	return newStringSliceCmd(r.Do(ctx, r.resp.getHKeysCompleted(key)))
}

func (r *resp3) getHLenCompleted(key string) rueidis.Completed {
	return r.cmd.B().Hlen().Key(key).Build()
}

func (r *resp3) HLen(ctx context.Context, key string) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.getHLenCompleted(key)))
}

func (r *resp3Cache) HLen(ctx context.Context, key string) IntCmd {
	return newIntCmd(r.Do(ctx, r.resp.getHLenCompleted(key)))
}

func (r *resp3) getHMGetCompleted(key string, fields ...string) rueidis.Completed {
	return r.cmd.B().Hmget().Key(key).Field(fields...).Build()
}

func (r *resp3) HMGet(ctx context.Context, key string, fields ...string) SliceCmd {
	return newSliceCmd(r.cmd.Do(ctx, r.getHMGetCompleted(key, fields...)), HMGET)
}

func (r *resp3Cache) HMGet(ctx context.Context, key string, fields ...string) SliceCmd {
	return newSliceCmd(r.Do(ctx, r.resp.getHMGetCompleted(key, fields...)), HMGET)
}

func (r *resp3) HMSet(ctx context.Context, key string, values ...interface{}) BoolCmd {
	fv := r.cmd.B().Hset().Key(key).FieldValue()
	args := argsToSlice(values)
	for i := 0; i < len(args); i += 2 {
		fv = fv.FieldValue(args[i], args[i+1])
	}
	return newBoolCmd(r.cmd.Do(ctx, fv.Build()))
}

func (r *resp3) HRandField(ctx context.Context, key string, count int, withValues bool) StringSliceCmd {
	h := r.cmd.B().Hrandfield().Key(key).Count(int64(count))
	if withValues {
		return flattenStringSliceCmd(r.cmd.Do(ctx, h.Withvalues().Build()))
	}
	return newStringSliceCmd(r.cmd.Do(ctx, h.Build()))
}

func (r *resp3) HScan(ctx context.Context, key string, cursor uint64, match string, count int64) ScanCmd {
	cmd := r.cmd.B().Arbitrary(HSCAN).Keys(key).Args(str(int64(cursor)))
	if match != "" {
		cmd = cmd.Args(MATCH, match)
	}
	if count > 0 {
		cmd = cmd.Args(COUNT, str(count))
	}
	return newScanCmd(r.cmd.Do(ctx, cmd.ReadOnly()))
}

func (r *resp3) HSet(ctx context.Context, key string, values ...interface{}) IntCmd {
	fv := r.cmd.B().Hset().Key(key).FieldValue()
	args := argsToSlice(values)
	for i := 0; i < len(args); i += 2 {
		fv = fv.FieldValue(args[i], args[i+1])
	}
	return newIntCmd(r.cmd.Do(ctx, fv.Build()))
}

func (r *resp3) HSetNX(ctx context.Context, key, field string, value interface{}) BoolCmd {
	return newBoolCmd(r.cmd.Do(ctx, r.cmd.B().Hsetnx().Key(key).Field(field).Value(str(value)).Build()))
}

func (r *resp3) getHValsCompleted(key string) rueidis.Completed {
	return r.cmd.B().Hvals().Key(key).Build()
}

func (r *resp3) HVals(ctx context.Context, key string) StringSliceCmd {
	return newStringSliceCmd(r.cmd.Do(ctx, r.getHValsCompleted(key)))
}

func (r *resp3Cache) HVals(ctx context.Context, key string) StringSliceCmd {
	return newStringSliceCmd(r.Do(ctx, r.resp.getHValsCompleted(key)))
}

func (r *resp3) PFAdd(ctx context.Context, key string, els ...interface{}) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Pfadd().Key(key).Element(argsToSlice(els)...).Build()))
}

func (r *resp3) PFCount(ctx context.Context, keys ...string) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Pfcount().Key(keys...).Build()))
}

func (r *resp3) PFMerge(ctx context.Context, dest string, keys ...string) StatusCmd {
	return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().Pfmerge().Destkey(dest).Sourcekey(keys...).Build()))
}

func (r *resp3) BLMove(ctx context.Context, source, destination, srcpos, destpos string, timeout time.Duration) StringCmd {
	return newStringCmd(r.cmd.Do(ctx, r.cmd.B().Arbitrary(BLMOVE).Keys(source, destination).
		Args(srcpos, destpos, str(float64(formatSec(timeout)))).Blocking()))
}

func (r *resp3) BLPop(ctx context.Context, timeout time.Duration, keys ...string) StringSliceCmd {
	return newStringSliceCmd(r.cmd.Do(ctx, r.cmd.B().Blpop().Key(keys...).Timeout(float64(formatSec(timeout))).Build()))
}

func (r *resp3) BRPop(ctx context.Context, timeout time.Duration, keys ...string) StringSliceCmd {
	return newStringSliceCmd(r.cmd.Do(ctx, r.cmd.B().Brpop().Key(keys...).Timeout(float64(formatSec(timeout))).Build()))
}

func (r *resp3) BRPopLPush(ctx context.Context, source, destination string, timeout time.Duration) StringCmd {
	return newStringCmd(r.cmd.Do(ctx, r.cmd.B().Brpoplpush().Source(source).Destination(destination).Timeout(float64(formatSec(timeout))).Build()))
}

func (r *resp3) getLIndexCompleted(key string, index int64) rueidis.Completed {
	return r.cmd.B().Lindex().Key(key).Index(index).Build()
}

func (r *resp3) LIndex(ctx context.Context, key string, index int64) StringCmd {
	return newStringCmd(r.cmd.Do(ctx, r.getLIndexCompleted(key, index)))
}

func (r *resp3Cache) LIndex(ctx context.Context, key string, index int64) StringCmd {
	return newStringCmd(r.Do(ctx, r.resp.getLIndexCompleted(key, index)))
}

func (r *resp3) LInsert(ctx context.Context, key, op string, pivot, value interface{}) IntCmd {
	var linsertKey = r.cmd.B().Linsert().Key(key)
	switch strings.ToUpper(op) {
	case BEFORE:
		return newIntCmd(r.cmd.Do(ctx, linsertKey.Before().Pivot(str(pivot)).Element(str(value)).Build()))
	case AFTER:
		return newIntCmd(r.cmd.Do(ctx, linsertKey.After().Pivot(str(pivot)).Element(str(value)).Build()))
	default:
		panic(fmt.Sprintf("Invalid op argument value: %s", op))
	}
}

func (r *resp3) LInsertBefore(ctx context.Context, key string, pivot, value interface{}) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Linsert().Key(key).Before().Pivot(str(pivot)).Element(str(value)).Build()))
}

func (r *resp3) LInsertAfter(ctx context.Context, key string, pivot, value interface{}) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Linsert().Key(key).After().Pivot(str(pivot)).Element(str(value)).Build()))
}

func (r *resp3) getLLenCompleted(key string) rueidis.Completed {
	return r.cmd.B().Llen().Key(key).Build()
}

func (r *resp3) LLen(ctx context.Context, key string) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.getLLenCompleted(key)))
}

func (r *resp3Cache) LLen(ctx context.Context, key string) IntCmd {
	return newIntCmd(r.Do(ctx, r.resp.getLLenCompleted(key)))
}

func (r *resp3) LMove(ctx context.Context, source, destination, srcpos, destpos string) StringCmd {
	return newStringCmd(r.cmd.Do(ctx, r.cmd.B().Arbitrary(LMOVE).Keys(source, destination).Args(srcpos, destpos).Build()))
}

func (r *resp3) LPop(ctx context.Context, key string) StringCmd {
	return newStringCmd(r.cmd.Do(ctx, r.cmd.B().Lpop().Key(key).Build()))
}

func (r *resp3) LPopCount(ctx context.Context, key string, count int) StringSliceCmd {
	return newStringSliceCmd(r.cmd.Do(ctx, r.cmd.B().Lpop().Key(key).Count(int64(count)).Build()))
}

func (r *resp3) getLPosCompleted(key string, value string, count int64, args LPosArgs) rueidis.Completed {
	arbitrary := r.cmd.B().Arbitrary(LPOS).Keys(key).Args(value)
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

func (r *resp3) LPos(ctx context.Context, key string, value string, args LPosArgs) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.getLPosCompleted(key, value, -1, args)))
}

func (r *resp3Cache) LPos(ctx context.Context, key string, value string, args LPosArgs) IntCmd {
	return newIntCmd(r.Do(ctx, r.resp.getLPosCompleted(key, value, -1, args)))
}

func (r *resp3) LPosCount(ctx context.Context, key string, value string, count int64, args LPosArgs) IntSliceCmd {
	return newIntSliceCmd(r.cmd.Do(ctx, r.getLPosCompleted(key, value, count, args)))
}

func (r *resp3Cache) LPosCount(ctx context.Context, key string, value string, count int64, args LPosArgs) IntSliceCmd {
	return newIntSliceCmd(r.Do(ctx, r.resp.getLPosCompleted(key, value, count, args)))
}

func (r *resp3) LPush(ctx context.Context, key string, values ...interface{}) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Lpush().Key(key).Element(argsToSlice(values)...).Build()))
}

func (r *resp3) LPushX(ctx context.Context, key string, values ...interface{}) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Lpushx().Key(key).Element(argsToSlice(values)...).Build()))
}

func (r *resp3) getLRangeCompleted(key string, start, stop int64) rueidis.Completed {
	return r.cmd.B().Lrange().Key(key).Start(start).Stop(stop).Build()
}

func (r *resp3) LRange(ctx context.Context, key string, start, stop int64) StringSliceCmd {
	return newStringSliceCmd(r.cmd.Do(ctx, r.getLRangeCompleted(key, start, stop)))
}

func (r *resp3Cache) LRange(ctx context.Context, key string, start, stop int64) StringSliceCmd {
	return newStringSliceCmd(r.Do(ctx, r.resp.getLRangeCompleted(key, start, stop)))
}

func (r *resp3) LRem(ctx context.Context, key string, count int64, value interface{}) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Lrem().Key(key).Count(count).Element(str(value)).Build()))
}

func (r *resp3) LSet(ctx context.Context, key string, index int64, value interface{}) StatusCmd {
	return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().Lset().Key(key).Index(index).Element(str(value)).Build()))
}

func (r *resp3) LTrim(ctx context.Context, key string, start, stop int64) StatusCmd {
	return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().Ltrim().Key(key).Start(start).Stop(stop).Build()))
}

func (r *resp3) RPop(ctx context.Context, key string) StringCmd {
	return newStringCmd(r.cmd.Do(ctx, r.cmd.B().Rpop().Key(key).Build()))
}

func (r *resp3) RPopCount(ctx context.Context, key string, count int) StringSliceCmd {
	return newStringSliceCmd(r.cmd.Do(ctx, r.cmd.B().Rpop().Key(key).Count(int64(count)).Build()))
}

func (r *resp3) RPopLPush(ctx context.Context, source, destination string) StringCmd {
	return newStringCmd(r.cmd.Do(ctx, r.cmd.B().Rpoplpush().Source(source).Destination(destination).Build()))
}

func (r *resp3) RPush(ctx context.Context, key string, values ...interface{}) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Rpush().Key(key).Element(argsToSlice(values)...).Build()))
}

func (r *resp3) RPushX(ctx context.Context, key string, values ...interface{}) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Rpushx().Key(key).Element(argsToSlice(values)...).Build()))
}

type pipelineResp3 struct {
	resp       *resp3
	mx         sync.Mutex
	firstError error
	res        []interface{}
}

func (r *resp3) Pipeline() Pipeliner { return &pipelineResp3{resp: r} }

func (p *pipelineResp3) Put(ctx context.Context, cmd Command, keys []string, args ...interface{}) (err error) {
	ctx = p.resp.handler.before(ctx, cmd)
	var r interface{}
	r, err = p.resp.cmd.Do(ctx, p.resp.cmd.B().Arbitrary(cmd.Cmd()...).Keys(keys...).Args(argsToSlice(args)...).Build()).ToAny()
	p.mx.Lock()
	if err != nil {
		p.res = append(p.res, err)
		if p.firstError == nil {
			p.firstError = err
		}
	} else {
		p.res = append(p.res, r)
	}
	p.mx.Unlock()
	p.resp.handler.after(ctx, err)
	return err
}

func (p *pipelineResp3) Exec(_ context.Context) ([]interface{}, error) {
	return p.res, p.firstError
}

func (r *resp3) Publish(ctx context.Context, channel string, message interface{}) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Publish().Channel(channel).Message(str(message)).Build()))
}

func (r *resp3) PubSubChannels(ctx context.Context, pattern string) StringSliceCmd {
	return newStringSliceCmd(r.cmd.Do(ctx, r.cmd.B().PubsubChannels().Pattern(pattern).Build()))
}

func (r *resp3) PubSubNumPat(ctx context.Context) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().PubsubNumpat().Build()))
}

func (r *resp3) PubSubNumSub(ctx context.Context, channels ...string) StringIntMapCmd {
	return newStringIntMapCmd(r.cmd.Do(ctx, r.cmd.B().PubsubNumsub().Channel(channels...).Build()))
}

func (r *resp3) Subscribe(ctx context.Context, channels ...string) PubSub {
	return newPubSubResp3(ctx, r.cmd, r.handler, channels...)
}

type pubSubResp3 struct {
	cmd     rueidis.DedicatedClient
	msgCh   chan *Message
	handler handler
	cancel  context.CancelFunc
}

func newPubSubResp3(ctx context.Context, cmd rueidis.Client, handler handler, channels ...string) PubSub {
	// chan size todo, use goredis.ChannelOption?
	p := &pubSubResp3{msgCh: make(chan *Message, 100), handler: handler}
	p.cmd, p.cancel = cmd.Dedicate()
	p.cmd.SetPubSubHooks(rueidis.PubSubHooks{
		OnMessage: func(m rueidis.PubSubMessage) {
			p.msgCh <- &Message{
				Channel: m.Channel,
				Pattern: m.Pattern,
				Payload: m.Message,
			}
		},
	})
	if len(channels) > 0 {
		_ = p.Subscribe(ctx, channels...)
	}
	return p
}

func (p *pubSubResp3) Close() error {
	close(p.msgCh)
	p.cancel()
	return nil
}

func (p *pubSubResp3) PSubscribe(ctx context.Context, patterns ...string) error {
	ctx = p.handler.before(ctx, CommandPSubscribe)
	err := p.cmd.Do(ctx, p.cmd.B().Psubscribe().Pattern(patterns...).Build()).Error()
	p.handler.after(ctx, err)
	return err
}

func (p *pubSubResp3) Subscribe(ctx context.Context, channels ...string) error {
	ctx = p.handler.before(ctx, CommandSubscribe)
	err := p.cmd.Do(ctx, p.cmd.B().Subscribe().Channel(channels...).Build()).Error()
	p.handler.after(ctx, err)
	return err
}

func (p *pubSubResp3) Unsubscribe(ctx context.Context, channels ...string) error {
	ctx = p.handler.before(ctx, CommandUnsubscribe)
	err := p.cmd.Do(ctx, p.cmd.B().Unsubscribe().Channel(channels...).Build()).Error()
	p.handler.after(ctx, err)
	return err
}

func (p *pubSubResp3) PUnsubscribe(ctx context.Context, patterns ...string) error {
	ctx = p.handler.before(ctx, CommandPUnsubscribe)
	err := p.cmd.Do(ctx, p.cmd.B().Punsubscribe().Pattern(patterns...).Build()).Error()
	p.handler.after(ctx, err)
	return err
}

func (p *pubSubResp3) Channel() <-chan *Message {
	return p.msgCh
}

func (r *resp3) CreateScript(string) Scripter { return nil }

func (r *resp3) Eval(ctx context.Context, script string, keys []string, args ...interface{}) Cmd {
	return newCmd(r.cmd.Do(ctx, r.cmd.B().Eval().Script(script).Numkeys(int64(len(keys))).Key(keys...).Arg(argsToSlice(args)...).Build()))
}

func (r *resp3) EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) Cmd {
	return newCmd(r.cmd.Do(ctx, r.cmd.B().Evalsha().Sha1(sha1).Numkeys(int64(len(keys))).Key(keys...).Arg(argsToSlice(args)...).Build()))
}

func (r *resp3) ScriptExists(ctx context.Context, hashes ...string) BoolSliceCmd {
	//return newBoolSliceCmd(r.cmd.Do(ctx, r.cmd.B().ScriptExists().Sha1(hashes...).Build()))
	return r.adapter.ScriptExists(ctx, hashes...)
}

func (r *resp3) ScriptFlush(ctx context.Context) StatusCmd {
	//return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().ScriptFlush().Build()))
	return r.adapter.ScriptFlush(ctx)
}

func (r *resp3) ScriptKill(ctx context.Context) StatusCmd {
	//return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().ScriptKill().Build()))
	return r.adapter.ScriptKill(ctx)
}

func (r *resp3) ScriptLoad(ctx context.Context, script string) StringCmd {
	//return newStringCmd(r.cmd.Do(ctx, r.cmd.B().ScriptLoad().Script(script).Build()))
	return newStringCmdFromStringCmd(r.adapter.ScriptLoad(ctx, script))
}

func (r *resp3) BgRewriteAOF(ctx context.Context) StatusCmd {
	//return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().Bgrewriteaof().Build()))
	return r.adapter.BgRewriteAOF(ctx)
}

func (r *resp3) BgSave(ctx context.Context) StatusCmd {
	//return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().Bgsave().Build()))
	return r.adapter.BgSave(ctx)
}

func (r *resp3) Command(ctx context.Context) CommandsInfoCmd {
	return newCommandsInfoCmd(r.cmd.Do(ctx, r.cmd.B().Command().Build()))
}

func (r *resp3) ConfigGet(ctx context.Context, parameter string) SliceCmd {
	return newSliceCmdFromMap(r.cmd.Do(ctx, r.cmd.B().ConfigGet().Parameter(parameter).Build()))
}

func (r *resp3) ConfigResetStat(ctx context.Context) StatusCmd {
	//return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().ConfigResetstat().Build()))
	return r.adapter.ConfigResetStat(ctx)
}

func (r *resp3) ConfigRewrite(ctx context.Context) StatusCmd {
	//return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().ConfigRewrite().Build()))
	return r.adapter.ConfigRewrite(ctx)
}

func (r *resp3) ConfigSet(ctx context.Context, parameter, value string) StatusCmd {
	//return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().ConfigSet().ParameterValue().ParameterValue(parameter, value).Build()))
	return r.adapter.ConfigSet(ctx, parameter, value)
}

func (r *resp3) DBSize(ctx context.Context) IntCmd {
	//return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Dbsize().Build()))
	return r.adapter.DBSize(ctx)
}

func (r *resp3) FlushAll(ctx context.Context) StatusCmd {
	//return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().Flushall().Build()))
	return r.adapter.FlushAll(ctx)
}

func (r *resp3) FlushAllAsync(ctx context.Context) StatusCmd {
	//return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().Flushall().Async().Build()))
	return r.adapter.FlushAllAsync(ctx)
}

func (r *resp3) FlushDB(ctx context.Context) StatusCmd {
	//return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().Flushdb().Build()))
	return r.adapter.FlushDB(ctx)
}

func (r *resp3) FlushDBAsync(ctx context.Context) StatusCmd {
	//return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().Flushdb().Async().Build()))
	return r.adapter.FlushDBAsync(ctx)
}

func (r *resp3) Info(ctx context.Context, section ...string) StringCmd {
	return newStringCmd(r.cmd.Do(ctx, r.cmd.B().Info().Section(section...).Build()))
}

func (r *resp3) LastSave(ctx context.Context) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Lastsave().Build()))
}

func (r *resp3) MemoryUsage(ctx context.Context, key string, samples ...int) IntCmd {
	var memoryUsageKey = r.cmd.B().MemoryUsage().Key(key)
	switch len(samples) {
	case 0:
		return newIntCmd(r.cmd.Do(ctx, memoryUsageKey.Build()))
	case 1:
		return newIntCmd(r.cmd.Do(ctx, memoryUsageKey.Samples(int64(samples[0])).Build()))
	default:
		panic(errMemoryUsageArgsCount)
	}

}

func (r *resp3) Save(ctx context.Context) StatusCmd {
	//return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().Save().Build()))
	return r.adapter.Save(ctx)
}

func (r *resp3) Shutdown(ctx context.Context) StatusCmd {
	//return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().Shutdown().Build()))
	return r.adapter.Shutdown(ctx)
}

func (r *resp3) ShutdownSave(ctx context.Context) StatusCmd {
	//return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().Shutdown().Save().Build()))
	return r.adapter.ShutdownSave(ctx)
}

func (r *resp3) ShutdownNoSave(ctx context.Context) StatusCmd {
	//return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().Shutdown().Nosave().Build()))
	return r.adapter.ShutdownNoSave(ctx)
}

func (r *resp3) SlaveOf(ctx context.Context, host, port string) StatusCmd {
	return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().Arbitrary(SLAVEOF).Args(host, port).Build()))
}

func (r *resp3) Time(ctx context.Context) TimeCmd {
	return newTimeCmd(r.cmd.Do(ctx, r.cmd.B().Time().Build()))
}

func (r *resp3) DebugObject(ctx context.Context, key string) StringCmd {
	return newStringCmd(r.cmd.Do(ctx, r.cmd.B().DebugObject().Key(key).Build()))
}

func (r *resp3) SAdd(ctx context.Context, key string, members ...interface{}) IntCmd {
	cmd := r.cmd.B().Sadd().Key(key).Member()
	for _, m := range argsToSlice(members) {
		cmd = cmd.Member(str(m))
	}
	return newIntCmd(r.cmd.Do(ctx, cmd.Build()))
}

func (r *resp3) getSCardCompleted(key string) rueidis.Completed {
	return r.cmd.B().Scard().Key(key).Build()
}

func (r *resp3) SCard(ctx context.Context, key string) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.getSCardCompleted(key)))
}

func (r *resp3Cache) SCard(ctx context.Context, key string) IntCmd {
	return newIntCmd(r.Do(ctx, r.resp.getSCardCompleted(key)))
}

func (r *resp3) SDiff(ctx context.Context, keys ...string) StringSliceCmd {
	return newStringSliceCmd(r.cmd.Do(ctx, r.cmd.B().Sdiff().Key(keys...).Build()))
}

func (r *resp3) SDiffStore(ctx context.Context, destination string, keys ...string) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Sdiffstore().Destination(destination).Key(keys...).Build()))
}

func (r *resp3) SInter(ctx context.Context, keys ...string) StringSliceCmd {
	return newStringSliceCmd(r.cmd.Do(ctx, r.cmd.B().Sinter().Key(keys...).Build()))
}

func (r *resp3) SInterStore(ctx context.Context, destination string, keys ...string) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Sinterstore().Destination(destination).Key(keys...).Build()))
}

func (r *resp3) getSIsMemberCompleted(key string, member interface{}) rueidis.Completed {
	return r.cmd.B().Sismember().Key(key).Member(str(member)).Build()
}

func (r *resp3) SIsMember(ctx context.Context, key string, member interface{}) BoolCmd {
	return newBoolCmd(r.cmd.Do(ctx, r.getSIsMemberCompleted(key, member)))
}

func (r *resp3Cache) SIsMember(ctx context.Context, key string, member interface{}) BoolCmd {
	return newBoolCmd(r.Do(ctx, r.resp.getSIsMemberCompleted(key, member)))
}

func (r *resp3) getSMIsMemberCompleted(key string, members ...interface{}) rueidis.Completed {
	return r.cmd.B().Smismember().Key(key).Member(argsToSlice(members)...).Build()
}

func (r *resp3) SMIsMember(ctx context.Context, key string, members ...interface{}) BoolSliceCmd {
	return newBoolSliceCmd(r.cmd.Do(ctx, r.getSMIsMemberCompleted(key, members...)))
}

func (r *resp3Cache) SMIsMember(ctx context.Context, key string, members ...interface{}) BoolSliceCmd {
	return newBoolSliceCmd(r.Do(ctx, r.resp.getSMIsMemberCompleted(key, members...)))
}

func (r *resp3) getSMembersCompleted(key string) rueidis.Completed {
	return r.cmd.B().Smembers().Key(key).Build()
}

func (r *resp3) SMembers(ctx context.Context, key string) StringSliceCmd {
	return newStringSliceCmd(r.cmd.Do(ctx, r.getSMembersCompleted(key)))
}

func (r *resp3Cache) SMembers(ctx context.Context, key string) StringSliceCmd {
	return newStringSliceCmd(r.Do(ctx, r.resp.getSMembersCompleted(key)))
}

func (r *resp3) SMembersMap(ctx context.Context, key string) StringStructMapCmd {
	return newStringStructMapCmd(r.cmd.Do(ctx, r.getSMembersCompleted(key)))
}

func (r *resp3Cache) SMembersMap(ctx context.Context, key string) StringStructMapCmd {
	return newStringStructMapCmd(r.Do(ctx, r.resp.getSMembersCompleted(key)))
}

func (r *resp3) SMove(ctx context.Context, source, destination string, member interface{}) BoolCmd {
	return newBoolCmd(r.cmd.Do(ctx, r.cmd.B().Smove().Source(source).Destination(destination).Member(str(member)).Build()))
}

func (r *resp3) SPop(ctx context.Context, key string) StringCmd {
	return newStringCmd(r.cmd.Do(ctx, r.cmd.B().Spop().Key(key).Build()))
}

func (r *resp3) SPopN(ctx context.Context, key string, count int64) StringSliceCmd {
	return newStringSliceCmd(r.cmd.Do(ctx, r.cmd.B().Spop().Key(key).Count(count).Build()))
}

func (r *resp3) SRandMember(ctx context.Context, key string) StringCmd {
	return newStringCmd(r.cmd.Do(ctx, r.cmd.B().Srandmember().Key(key).Build()))
}

func (r *resp3) SRandMemberN(ctx context.Context, key string, count int64) StringSliceCmd {
	return newStringSliceCmd(r.cmd.Do(ctx, r.cmd.B().Srandmember().Key(key).Count(count).Build()))
}

func (r *resp3) SRem(ctx context.Context, key string, members ...interface{}) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Srem().Key(key).Member(argsToSlice(members)...).Build()))
}

func (r *resp3) SScan(ctx context.Context, key string, cursor uint64, match string, count int64) ScanCmd {
	cmd := r.cmd.B().Arbitrary(SSCAN).Keys(key).Args(str(int64(cursor)))
	if match != "" {
		cmd = cmd.Args(MATCH, match)
	}
	if count > 0 {
		cmd = cmd.Args(COUNT, str(count))
	}
	return newScanCmd(r.cmd.Do(ctx, cmd.ReadOnly()))
}

func (r *resp3) SUnion(ctx context.Context, keys ...string) StringSliceCmd {
	return newStringSliceCmd(r.cmd.Do(ctx, r.cmd.B().Sunion().Key(keys...).Build()))
}

func (r *resp3) SUnionStore(ctx context.Context, destination string, keys ...string) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Sunionstore().Destination(destination).Key(keys...).Build()))
}

func (r *resp3) BZPopMax(ctx context.Context, timeout time.Duration, keys ...string) ZWithKeyCmd {
	return newZWithKeyCmd(r.cmd.Do(ctx, r.cmd.B().Bzpopmax().Key(keys...).Timeout(float64(formatSec(timeout))).Build()))
}

func (r *resp3) BZPopMin(ctx context.Context, timeout time.Duration, keys ...string) ZWithKeyCmd {
	return newZWithKeyCmd(r.cmd.Do(ctx, r.cmd.B().Bzpopmin().Key(keys...).Timeout(float64(formatSec(timeout))).Build()))
}

func (r *resp3) ZAdd(ctx context.Context, key string, members ...Z) IntCmd {
	cmd := r.cmd.B().Zadd().Key(key).ScoreMember()
	for _, v := range members {
		cmd = cmd.ScoreMember(v.Score, str(v.Member))
	}
	return newIntCmd(r.cmd.Do(ctx, cmd.Build()))
}

func (r *resp3) ZAddNX(ctx context.Context, key string, members ...Z) IntCmd {
	cmd := r.cmd.B().Zadd().Key(key).Nx().ScoreMember()
	for _, v := range members {
		cmd = cmd.ScoreMember(v.Score, str(v.Member))
	}
	return newIntCmd(r.cmd.Do(ctx, cmd.Build()))
}

func (r *resp3) ZAddXX(ctx context.Context, key string, members ...Z) IntCmd {
	cmd := r.cmd.B().Zadd().Key(key).Xx().ScoreMember()
	for _, v := range members {
		cmd = cmd.ScoreMember(v.Score, str(v.Member))
	}
	return newIntCmd(r.cmd.Do(ctx, cmd.Build()))
}

func (r *resp3) zAddArgs(ctx context.Context, key string, incr bool, args ZAddArgs, members ...Z) rueidis.RedisResult {
	cmd := r.cmd.B().Arbitrary(ZADD).Keys(key)
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
	return r.cmd.Do(ctx, cmd.Build())
}

func (r *resp3) ZAddCh(ctx context.Context, key string, members ...Z) IntCmd {
	return newIntCmd(r.zAddArgs(ctx, key, false, ZAddArgs{Ch: true}, members...))
}

func (r *resp3) ZAddNXCh(ctx context.Context, key string, members ...Z) IntCmd {
	return newIntCmd(r.zAddArgs(ctx, key, false, ZAddArgs{NX: true, Ch: true}, members...))
}

func (r *resp3) ZAddXXCh(ctx context.Context, key string, members ...Z) IntCmd {
	return newIntCmd(r.zAddArgs(ctx, key, false, ZAddArgs{XX: true, Ch: true}, members...))
}

func (r *resp3) ZAddArgs(ctx context.Context, key string, args ZAddArgs) IntCmd {
	return newIntCmd(r.zAddArgs(ctx, key, false, args, args.Members...))
}

func (r *resp3) ZAddArgsIncr(ctx context.Context, key string, args ZAddArgs) FloatCmd {
	return newFloatCmd(r.zAddArgs(ctx, key, true, args, args.Members...))
}

func (r *resp3) getZCardCompleted(key string) rueidis.Completed {
	return r.cmd.B().Zcard().Key(key).Build()
}

func (r *resp3) ZCard(ctx context.Context, key string) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.getZCardCompleted(key)))
}

func (r *resp3Cache) ZCard(ctx context.Context, key string) IntCmd {
	return newIntCmd(r.Do(ctx, r.resp.getZCardCompleted(key)))
}

func (r *resp3) getZCount(key, min, max string) rueidis.Completed {
	return r.cmd.B().Zcount().Key(key).Min(min).Max(max).Build()
}

func (r *resp3) ZCount(ctx context.Context, key, min, max string) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.getZCount(key, min, max)))
}

func (r *resp3Cache) ZCount(ctx context.Context, key, min, max string) IntCmd {
	return newIntCmd(r.Do(ctx, r.resp.getZCount(key, min, max)))
}

func (r *resp3) ZDiff(ctx context.Context, keys ...string) StringSliceCmd {
	return newStringSliceCmd(r.cmd.Do(ctx, r.cmd.B().Zdiff().Numkeys(int64(len(keys))).Key(keys...).Build()))
}

func (r *resp3) ZDiffWithScores(ctx context.Context, keys ...string) ZSliceCmd {
	return newZSliceCmd(r.cmd.Do(ctx, r.cmd.B().Zdiff().Numkeys(int64(len(keys))).Key(keys...).Withscores().Build()))
}

func (r *resp3) ZDiffStore(ctx context.Context, destination string, keys ...string) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Zdiffstore().Destination(destination).Numkeys(int64(len(keys))).Key(keys...).Build()))
}

func (r *resp3) ZIncr(ctx context.Context, key string, member Z) FloatCmd {
	return newFloatCmd(r.zAddArgs(ctx, key, true, ZAddArgs{}, member))
}

func (r *resp3) ZIncrNX(ctx context.Context, key string, member Z) FloatCmd {
	return newFloatCmd(r.zAddArgs(ctx, key, true, ZAddArgs{NX: true}, member))
}

func (r *resp3) ZIncrXX(ctx context.Context, key string, member Z) FloatCmd {
	return newFloatCmd(r.zAddArgs(ctx, key, true, ZAddArgs{XX: true}, member))
}

func (r *resp3) ZIncrBy(ctx context.Context, key string, increment float64, member string) FloatCmd {
	return newFloatCmd(r.cmd.Do(ctx, r.cmd.B().Zincrby().Key(key).Increment(increment).Member(member).Build()))
}

func (r *resp3) fillZInterArbitrary(arbitrary rueidis.Arbitrary, store ZStore) rueidis.Arbitrary {
	arbitrary = arbitrary.Args(str(len(store.Keys))).Keys(store.Keys...)
	if len(store.Weights) > 0 {
		arbitrary = arbitrary.Args(WEIGHTS)
		for _, w := range store.Weights {
			arbitrary = arbitrary.Args(str(w))
		}
	}
	if len(store.Aggregate) > 0 {
		arbitrary = arbitrary.Args(AGGREGATE, store.Aggregate)
	}
	return arbitrary
}

func (r *resp3) ZInter(ctx context.Context, store ZStore) StringSliceCmd {
	return newStringSliceCmd(r.cmd.Do(ctx, r.fillZInterArbitrary(r.cmd.B().Arbitrary(ZINTER), store).ReadOnly()))
}

func (r *resp3) ZInterWithScores(ctx context.Context, store ZStore) ZSliceCmd {
	return newZSliceCmd(r.cmd.Do(ctx, r.fillZInterArbitrary(r.cmd.B().Arbitrary(ZINTER), store).Args(WITHSCORES).ReadOnly()))
}

func (r *resp3) ZInterStore(ctx context.Context, destination string, store ZStore) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.fillZInterArbitrary(r.cmd.B().Arbitrary(ZINTERSTORE).Keys(destination), store).Build()))
}

func (r *resp3) getZLexCountCompleted(key, min, max string) rueidis.Completed {
	return r.cmd.B().Zlexcount().Key(key).Min(min).Max(max).Build()
}

func (r *resp3) ZLexCount(ctx context.Context, key, min, max string) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.getZLexCountCompleted(key, min, max)))
}

func (r *resp3Cache) ZLexCount(ctx context.Context, key, min, max string) IntCmd {
	return newIntCmd(r.Do(ctx, r.resp.getZLexCountCompleted(key, min, max)))
}

func (r *resp3) getZMScoreCompleted(key string, members ...string) rueidis.Completed {
	return r.cmd.B().Zmscore().Key(key).Member(members...).Build()
}

func (r *resp3) ZMScore(ctx context.Context, key string, members ...string) FloatSliceCmd {
	return newFloatSliceCmd(r.cmd.Do(ctx, r.getZMScoreCompleted(key, members...)))
}

func (r *resp3Cache) ZMScore(ctx context.Context, key string, members ...string) FloatSliceCmd {
	return newFloatSliceCmd(r.Do(ctx, r.resp.getZMScoreCompleted(key, members...)))
}

func (r *resp3) ZPopMax(ctx context.Context, key string, count ...int64) ZSliceCmd {
	var resp rueidis.RedisResult
	var zpopmaxKey = r.cmd.B().Zpopmax().Key(key)
	switch len(count) {
	case 0:
		resp = r.cmd.Do(ctx, zpopmaxKey.Build())
	case 1:
		resp = r.cmd.Do(ctx, zpopmaxKey.Count(count[0]).Build())
		if count[0] > 1 {
			return newZSliceCmd(resp)
		}
	default:
		panic(errTooManyArguments)
	}
	return newZSliceSingleCmd(resp)
}

func (r *resp3) ZPopMin(ctx context.Context, key string, count ...int64) ZSliceCmd {
	var resp rueidis.RedisResult
	var zpopminKey = r.cmd.B().Zpopmin().Key(key)
	switch len(count) {
	case 0:
		resp = r.cmd.Do(ctx, zpopminKey.Build())
	case 1:
		resp = r.cmd.Do(ctx, zpopminKey.Count(count[0]).Build())
		if count[0] > 1 {
			return newZSliceCmd(resp)
		}
	default:
		panic(errTooManyArguments)
	}
	return newZSliceSingleCmd(resp)
}

func (r *resp3) ZRandMember(ctx context.Context, key string, count int, withScores bool) StringSliceCmd {
	var zrandmemberOptionsCount = r.cmd.B().Zrandmember().Key(key).Count(int64(count))
	if withScores {
		return flattenStringSliceCmd(r.cmd.Do(ctx, zrandmemberOptionsCount.Withscores().Build()))
	}
	return newStringSliceCmd(r.cmd.Do(ctx, zrandmemberOptionsCount.Build()))
}

func (r *resp3) getZRangeCompleted(key string, start, stop int64) rueidis.Completed {
	return r.cmd.B().Zrange().Key(key).Min(str(start)).Max(str(stop)).Build()
}

func (r *resp3) ZRange(ctx context.Context, key string, start, stop int64) StringSliceCmd {
	return newStringSliceCmd(r.cmd.Do(ctx, r.getZRangeCompleted(key, start, stop)))
}

func (r *resp3Cache) ZRange(ctx context.Context, key string, start, stop int64) StringSliceCmd {
	return newStringSliceCmd(r.Do(ctx, r.resp.getZRangeCompleted(key, start, stop)))
}

func (r *resp3) getZRangeWithScoresCompleted(key string, start, stop int64) rueidis.Completed {
	return r.cmd.B().Zrange().Key(key).Min(str(start)).Max(str(stop)).Withscores().Build()
}

func (r *resp3) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ZSliceCmd {
	return newZSliceCmd(r.cmd.Do(ctx, r.getZRangeWithScoresCompleted(key, start, stop)))
}

func (r *resp3Cache) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ZSliceCmd {
	return newZSliceCmd(r.Do(ctx, r.resp.getZRangeWithScoresCompleted(key, start, stop)))
}

func (r *resp3) getZRangeByLexCompleted(key string, opt ZRangeBy) rueidis.Completed {
	var completed rueidis.Completed
	var zrangebylexMax = r.cmd.B().Zrangebylex().Key(key).Min(opt.Min).Max(opt.Max)
	if opt.Offset != 0 || opt.Count != 0 {
		completed = zrangebylexMax.Limit(opt.Offset, opt.Count).Build()
	} else {
		completed = zrangebylexMax.Build()
	}
	return completed
}

func (r *resp3) ZRangeByLex(ctx context.Context, key string, opt ZRangeBy) StringSliceCmd {
	return newStringSliceCmd(r.cmd.Do(ctx, r.getZRangeByLexCompleted(key, opt)))
}

func (r *resp3Cache) ZRangeByLex(ctx context.Context, key string, opt ZRangeBy) StringSliceCmd {
	return newStringSliceCmd(r.Do(ctx, r.resp.getZRangeByLexCompleted(key, opt)))
}

func (r *resp3) getZRangeByScoreCompleted(key string, withScore bool, opt ZRangeBy) rueidis.Completed {
	var completed rueidis.Completed
	var zrangebyscoreMax = r.cmd.B().Zrangebyscore().Key(key).Min(opt.Min).Max(opt.Max)
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

func (r *resp3) ZRangeByScore(ctx context.Context, key string, opt ZRangeBy) StringSliceCmd {
	return newStringSliceCmd(r.cmd.Do(ctx, r.getZRangeByScoreCompleted(key, false, opt)))
}

func (r *resp3Cache) ZRangeByScore(ctx context.Context, key string, opt ZRangeBy) StringSliceCmd {
	return newStringSliceCmd(r.Do(ctx, r.resp.getZRangeByScoreCompleted(key, false, opt)))
}

func (r *resp3) ZRangeByScoreWithScores(ctx context.Context, key string, opt ZRangeBy) ZSliceCmd {
	return newZSliceCmd(r.cmd.Do(ctx, r.getZRangeByScoreCompleted(key, true, opt)))
}

func (r *resp3Cache) ZRangeByScoreWithScores(ctx context.Context, key string, opt ZRangeBy) ZSliceCmd {
	return newZSliceCmd(r.Do(ctx, r.resp.getZRangeByScoreCompleted(key, true, opt)))
}

func (r *resp3) getZRangeArgsArbitrary(arbitrary rueidis.Arbitrary, withScores bool, z ZRangeArgs) rueidis.Arbitrary {
	if z.Rev && (z.ByScore || z.ByLex) {
		arbitrary = arbitrary.Args(str(z.Stop), str(z.Start))
	} else {
		arbitrary = arbitrary.Args(str(z.Start), str(z.Stop))
	}
	if z.ByScore {
		arbitrary = arbitrary.Args(BYSCORE)
	} else if z.ByLex {
		arbitrary = arbitrary.Args(BYLEX)
	}
	if z.Rev {
		arbitrary = arbitrary.Args(REV)
	}
	if z.Offset != 0 || z.Count != 0 {
		arbitrary = arbitrary.Args(LIMIT, str(z.Offset), str(z.Count))
	}
	if withScores {
		arbitrary = arbitrary.Args(WITHSCORES)
	}
	return arbitrary
}

func (r *resp3) getZRangeArgsCompleted(withScores bool, z ZRangeArgs) rueidis.Completed {
	return r.getZRangeArgsArbitrary(r.cmd.B().Arbitrary(ZRANGE).Keys(z.Key), withScores, z).Build()
}

func (r *resp3) ZRangeArgs(ctx context.Context, z ZRangeArgs) StringSliceCmd {
	return newStringSliceCmd(r.cmd.Do(ctx, r.getZRangeArgsCompleted(false, z)))
}

func (r *resp3Cache) ZRangeArgs(ctx context.Context, z ZRangeArgs) StringSliceCmd {
	return newStringSliceCmd(r.Do(ctx, r.resp.getZRangeArgsCompleted(false, z)))
}

func (r *resp3) ZRangeArgsWithScores(ctx context.Context, z ZRangeArgs) ZSliceCmd {
	return newZSliceCmd(r.cmd.Do(ctx, r.getZRangeArgsCompleted(true, z)))
}

func (r *resp3Cache) ZRangeArgsWithScores(ctx context.Context, z ZRangeArgs) ZSliceCmd {
	return newZSliceCmd(r.Do(ctx, r.resp.getZRangeArgsCompleted(true, z)))
}

func (r *resp3) ZRangeStore(ctx context.Context, dst string, z ZRangeArgs) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.getZRangeArgsArbitrary(r.cmd.B().Arbitrary(ZRANGESTORE).Keys(dst, z.Key), false, z).Build()))
}

func (r *resp3) getZRankCompleted(key, member string) rueidis.Completed {
	return r.cmd.B().Zrank().Key(key).Member(member).Build()
}

func (r *resp3) ZRank(ctx context.Context, key, member string) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.getZRankCompleted(key, member)))
}

func (r *resp3Cache) ZRank(ctx context.Context, key, member string) IntCmd {
	return newIntCmd(r.Do(ctx, r.resp.getZRankCompleted(key, member)))
}

func (r *resp3) ZRem(ctx context.Context, key string, members ...interface{}) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Zrem().Key(key).Member(argsToSlice(members)...).Build()))
}

func (r *resp3) ZRemRangeByLex(ctx context.Context, key, min, max string) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Zremrangebylex().Key(key).Min(min).Max(max).Build()))
}

func (r *resp3) ZRemRangeByRank(ctx context.Context, key string, start, stop int64) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Zremrangebyrank().Key(key).Start(start).Stop(stop).Build()))
}

func (r *resp3) ZRemRangeByScore(ctx context.Context, key, min, max string) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Zremrangebyscore().Key(key).Min(min).Max(max).Build()))
}

func (r *resp3) getZRevRangeCompleted(key string, start, stop int64, withScore bool) rueidis.Completed {
	var zrevrangeStop = r.cmd.B().Zrevrange().Key(key).Start(start).Stop(stop)
	if withScore {
		return zrevrangeStop.Withscores().Build()
	}
	return zrevrangeStop.Build()
}

func (r *resp3) ZRevRange(ctx context.Context, key string, start, stop int64) StringSliceCmd {
	return newStringSliceCmd(r.cmd.Do(ctx, r.getZRevRangeCompleted(key, start, stop, false)))
}

func (r *resp3Cache) ZRevRange(ctx context.Context, key string, start, stop int64) StringSliceCmd {
	return newStringSliceCmd(r.Do(ctx, r.resp.getZRevRangeCompleted(key, start, stop, false)))
}

func (r *resp3) ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) ZSliceCmd {
	return newZSliceCmd(r.cmd.Do(ctx, r.getZRevRangeCompleted(key, start, stop, true)))
}

func (r *resp3Cache) ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) ZSliceCmd {
	return newZSliceCmd(r.Do(ctx, r.resp.getZRevRangeCompleted(key, start, stop, true)))
}

func (r *resp3) getZRevRangeByLexCompleted(key string, opt ZRangeBy) rueidis.Completed {
	var completed rueidis.Completed
	var zrevrangebylexMin = r.cmd.B().Zrevrangebylex().Key(key).Max(opt.Max).Min(opt.Min)
	if opt.Offset != 0 || opt.Count != 0 {
		completed = zrevrangebylexMin.Limit(opt.Offset, opt.Count).Build()
	} else {
		completed = zrevrangebylexMin.Build()
	}
	return completed
}

func (r *resp3) ZRevRangeByLex(ctx context.Context, key string, opt ZRangeBy) StringSliceCmd {
	return newStringSliceCmd(r.cmd.Do(ctx, r.getZRevRangeByLexCompleted(key, opt)))
}

func (r *resp3Cache) ZRevRangeByLex(ctx context.Context, key string, opt ZRangeBy) StringSliceCmd {
	return newStringSliceCmd(r.Do(ctx, r.resp.getZRevRangeByLexCompleted(key, opt)))
}

func (r *resp3) getZRevRangeByScoreCompleted(key string, withScore bool, opt ZRangeBy) rueidis.Completed {
	var completed rueidis.Completed
	var zrevrangebyscoreMin = r.cmd.B().Zrevrangebyscore().Key(key).Max(opt.Max).Min(opt.Min)
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

func (r *resp3) ZRevRangeByScore(ctx context.Context, key string, opt ZRangeBy) StringSliceCmd {
	return newStringSliceCmd(r.cmd.Do(ctx, r.getZRevRangeByScoreCompleted(key, false, opt)))
}

func (r *resp3Cache) ZRevRangeByScore(ctx context.Context, key string, opt ZRangeBy) StringSliceCmd {
	return newStringSliceCmd(r.Do(ctx, r.resp.getZRevRangeByScoreCompleted(key, false, opt)))
}

func (r *resp3) ZRevRangeByScoreWithScores(ctx context.Context, key string, opt ZRangeBy) ZSliceCmd {
	return newZSliceCmd(r.cmd.Do(ctx, r.getZRevRangeByScoreCompleted(key, true, opt)))
}

func (r *resp3Cache) ZRevRangeByScoreWithScores(ctx context.Context, key string, opt ZRangeBy) ZSliceCmd {
	return newZSliceCmd(r.Do(ctx, r.resp.getZRevRangeByScoreCompleted(key, true, opt)))
}

func (r *resp3) getZRevRankCompleted(key, member string) rueidis.Completed {
	return r.cmd.B().Zrevrank().Key(key).Member(member).Build()
}

func (r *resp3) ZRevRank(ctx context.Context, key, member string) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.getZRevRankCompleted(key, member)))
}

func (r *resp3Cache) ZRevRank(ctx context.Context, key, member string) IntCmd {
	return newIntCmd(r.Do(ctx, r.resp.getZRevRankCompleted(key, member)))
}

func (r *resp3) ZScan(ctx context.Context, key string, cursor uint64, match string, count int64) ScanCmd {
	cmd := r.cmd.B().Arbitrary(ZSCAN).Keys(key).Args(str(cursor))
	if match != "" {
		cmd = cmd.Args(MATCH, match)
	}
	if count > 0 {
		cmd = cmd.Args(COUNT, str(count))
	}
	return newScanCmd(r.cmd.Do(ctx, cmd.ReadOnly()))
}

func (r *resp3) getZScoreCompleted(key, member string) rueidis.Completed {
	return r.cmd.B().Zscore().Key(key).Member(member).Build()
}

func (r *resp3) ZScore(ctx context.Context, key, member string) FloatCmd {
	return newFloatCmd(r.cmd.Do(ctx, r.getZScoreCompleted(key, member)))
}

func (r *resp3Cache) ZScore(ctx context.Context, key, member string) FloatCmd {
	return newFloatCmd(r.Do(ctx, r.resp.getZScoreCompleted(key, member)))
}

func (r *resp3) fillZUnionArbitrary(arbitrary rueidis.Arbitrary, withScore bool, store ZStore) rueidis.Arbitrary {
	arbitrary = arbitrary.Args(str(len(store.Keys))).Keys(store.Keys...)
	if len(store.Weights) > 0 {
		arbitrary = arbitrary.Args(WEIGHTS)
		for _, w := range store.Weights {
			arbitrary = arbitrary.Args(str(w))
		}
	}
	if len(store.Aggregate) > 0 {
		arbitrary = arbitrary.Args(AGGREGATE, store.Aggregate)
	}
	if withScore {
		arbitrary = arbitrary.Args(WITHSCORES)
	}
	return arbitrary
}

func (r *resp3) ZUnion(ctx context.Context, store ZStore) StringSliceCmd {
	return newStringSliceCmd(r.cmd.Do(ctx, r.fillZUnionArbitrary(r.cmd.B().Arbitrary(ZUNION), false, store).ReadOnly()))
}

func (r *resp3) ZUnionWithScores(ctx context.Context, store ZStore) ZSliceCmd {
	return newZSliceCmd(r.cmd.Do(ctx, r.fillZUnionArbitrary(r.cmd.B().Arbitrary(ZUNION), true, store).ReadOnly()))
}

func (r *resp3) ZUnionStore(ctx context.Context, dest string, store ZStore) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.fillZUnionArbitrary(r.cmd.B().Arbitrary(ZUNIONSTORE).Keys(dest), false, store).ReadOnly()))
}

func (r *resp3) XAck(ctx context.Context, stream, group string, ids ...string) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Xack().Key(stream).Group(group).Id(ids...).Build()))
}

func (r *resp3) XAdd(ctx context.Context, a XAddArgs) StringCmd {
	cmd := r.cmd.B().Arbitrary(XADD).Keys(a.Stream)
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
	return newStringCmd(r.cmd.Do(ctx, cmd.Args(argToSlice(a.Values)...).Build()))
}

func (r *resp3) getXAutoClaimCompleted(a XAutoClaimArgs, justId bool) rueidis.Completed {
	var completed rueidis.Completed
	var xautoclaimStart = r.cmd.B().Xautoclaim().Key(a.Stream).Group(a.Group).Consumer(a.Consumer).MinIdleTime(str(formatMs(a.MinIdle))).Start(a.Start)
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

func (r *resp3) XAutoClaim(ctx context.Context, a XAutoClaimArgs) XAutoClaimCmd {
	return newXAutoClaimCmd(r.cmd.Do(ctx, r.getXAutoClaimCompleted(a, false)))
}

func (r *resp3) XAutoClaimJustID(ctx context.Context, a XAutoClaimArgs) XAutoClaimJustIDCmd {
	return newXAutoClaimJustIDCmd(r.cmd.Do(ctx, r.getXAutoClaimCompleted(a, true)))
}

func (r *resp3) getXClaimCompleted(a XClaimArgs, justId bool) rueidis.Completed {
	var xclaimId = r.cmd.B().Xclaim().Key(a.Stream).Group(a.Group).Consumer(a.Consumer).MinIdleTime(str(formatMs(a.MinIdle))).Id(a.Messages...)
	if justId {
		return xclaimId.Justid().Build()
	}
	return xclaimId.Build()
}

func (r *resp3) XClaim(ctx context.Context, a XClaimArgs) XMessageSliceCmd {
	return newXMessageSliceCmd(r.cmd.Do(ctx, r.getXClaimCompleted(a, false)))
}

func (r *resp3) XClaimJustID(ctx context.Context, a XClaimArgs) StringSliceCmd {
	return newStringSliceCmd(r.cmd.Do(ctx, r.getXClaimCompleted(a, true)))
}

func (r *resp3) XDel(ctx context.Context, stream string, ids ...string) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Xdel().Key(stream).Id(ids...).Build()))
}

func (r *resp3) XGroupCreate(ctx context.Context, stream, group, start string) StatusCmd {
	return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().XgroupCreate().Key(stream).Groupname(group).Id(start).Build()))
}

func (r *resp3) XGroupCreateMkStream(ctx context.Context, stream, group, start string) StatusCmd {
	return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().XgroupCreate().Key(stream).Groupname(group).Id(start).Mkstream().Build()))
}

func (r *resp3) XGroupCreateConsumer(ctx context.Context, stream, group, consumer string) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().XgroupCreateconsumer().Key(stream).Groupname(group).Consumername(consumer).Build()))
}

func (r *resp3) XGroupDelConsumer(ctx context.Context, stream, group, consumer string) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().XgroupDelconsumer().Key(stream).Groupname(group).Consumername(consumer).Build()))
}

func (r *resp3) XGroupDestroy(ctx context.Context, stream, group string) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().XgroupDestroy().Key(stream).Groupname(group).Build()))
}

func (r *resp3) XGroupSetID(ctx context.Context, stream, group, start string) StatusCmd {
	return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().XgroupSetid().Key(stream).Groupname(group).Id(start).Build()))
}

func (r *resp3) XInfoConsumers(ctx context.Context, key string, group string) XInfoConsumersCmd {
	return newXInfoConsumersCmd(r.cmd.Do(ctx, r.cmd.B().XinfoConsumers().Key(key).Groupname(group).Build()), EMPTY, group)
}

func (r *resp3) XInfoGroups(ctx context.Context, key string) XInfoGroupsCmd {
	return newXInfoGroupsCmd(r.cmd.Do(ctx, r.cmd.B().XinfoGroups().Key(key).Build()), key)
}

func (r *resp3) XInfoStream(ctx context.Context, key string) XInfoStreamCmd {
	return newXInfoStreamCmd(r.cmd.Do(ctx, r.cmd.B().XinfoStream().Key(key).Build()), key)
}

func (r *resp3) XInfoStreamFull(ctx context.Context, key string, count int) XInfoStreamFullCmd {
	return newXInfoStreamFullCmd(r.cmd.Do(ctx, r.cmd.B().XinfoStream().Key(key).Full().Count(int64(count)).Build()))
}

func (r *resp3) XLen(ctx context.Context, stream string) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Xlen().Key(stream).Build()))
}

func (r *resp3) XPending(ctx context.Context, stream, group string) XPendingCmd {
	return newXPendingCmd(r.cmd.Do(ctx, r.cmd.B().Xpending().Key(stream).Group(group).Build()))
}

func (r *resp3) XPendingExt(ctx context.Context, a XPendingExtArgs) XPendingExtCmd {
	cmd := r.cmd.B().Arbitrary(XPENDING).Keys(a.Stream).Args(a.Group)
	if a.Idle != 0 {
		cmd = cmd.Args(IDLE, str(formatMs(a.Idle)))
	}
	cmd = cmd.Args(a.Start, a.End, str(a.Count))
	if len(a.Consumer) > 0 {
		cmd = cmd.Args(a.Consumer)
	}
	return newXPendingExtCmd(r.cmd.Do(ctx, cmd.Build()))
}

func (r *resp3) XRange(ctx context.Context, stream, start, stop string) XMessageSliceCmd {
	return newXMessageSliceCmd(r.cmd.Do(ctx, r.cmd.B().Xrange().Key(stream).Start(start).End(stop).Build()))
}

func (r *resp3) XRangeN(ctx context.Context, stream, start, stop string, count int64) XMessageSliceCmd {
	return newXMessageSliceCmd(r.cmd.Do(ctx, r.cmd.B().Xrange().Key(stream).Start(start).End(stop).Count(count).Build()))
}

func (r *resp3) xRead(ctx context.Context, arbitrary rueidis.Arbitrary, a XReadGroupArgs) XStreamSliceCmd {
	if a.Count > 0 {
		arbitrary = arbitrary.Args(COUNT, str(a.Count))
	}
	if a.Block >= 0 {
		arbitrary = arbitrary.Args(BLOCK, str(formatMs(a.Block)))
	}
	if a.NoAck {
		arbitrary = arbitrary.Args(NOACK)
	}
	arbitrary = arbitrary.Args(STREAMS).Keys(a.Streams[:len(a.Streams)/2]...).Args(a.Streams[len(a.Streams)/2:]...)
	var com rueidis.Completed
	if a.Block >= 0 {
		com = arbitrary.Blocking()
	} else {
		com = arbitrary.Build()
	}
	return newXStreamSliceCmd(r.cmd.Do(ctx, com))
}

func (r *resp3) XRead(ctx context.Context, a XReadArgs) XStreamSliceCmd {
	return r.xRead(ctx, r.cmd.B().Arbitrary(XREAD), XReadGroupArgs{Count: a.Count, Block: a.Block, Streams: a.Streams})
}

func (r *resp3) XReadStreams(ctx context.Context, streams ...string) XStreamSliceCmd {
	return r.XRead(ctx, XReadArgs{Streams: streams, Block: -1})
}

func (r *resp3) XReadGroup(ctx context.Context, a XReadGroupArgs) XStreamSliceCmd {
	return r.xRead(ctx, r.cmd.B().Arbitrary(XREADGROUP).Args(GROUP, a.Group, a.Consumer), a)
}

func (r *resp3) XRevRange(ctx context.Context, stream string, start, stop string) XMessageSliceCmd {
	return newXMessageSliceCmd(r.cmd.Do(ctx, r.cmd.B().Xrevrange().Key(stream).End(start).Start(stop).Build()))
}

func (r *resp3) XRevRangeN(ctx context.Context, stream string, start, stop string, count int64) XMessageSliceCmd {
	return newXMessageSliceCmd(r.cmd.Do(ctx, r.cmd.B().Xrevrange().Key(stream).End(start).Start(stop).Count(count).Build()))
}

func (r *resp3) xTrim(ctx context.Context, key, strategy string,
	approx bool, threshold string, limit int64) IntCmd {
	cmd := r.cmd.B().Arbitrary(XTRIM).Keys(key).Args(strategy)
	if approx {
		cmd = cmd.Args("~")
	}
	cmd = cmd.Args(threshold)
	if limit > 0 {
		cmd = cmd.Args(LIMIT, str(limit))
	}
	return newIntCmd(r.cmd.Do(ctx, cmd.Build()))
}

func (r *resp3) XTrim(ctx context.Context, key string, maxLen int64) IntCmd {
	return r.xTrim(ctx, key, MAXLEN, false, str(maxLen), 0)
}

func (r *resp3) XTrimApprox(ctx context.Context, key string, maxLen int64) IntCmd {
	return r.xTrim(ctx, key, MAXLEN, true, str(maxLen), 0)
}

func (r *resp3) XTrimMaxLen(ctx context.Context, key string, maxLen int64) IntCmd {
	return r.xTrim(ctx, key, MAXLEN, false, str(maxLen), 0)
}

func (r *resp3) XTrimMinID(ctx context.Context, key string, minID string) IntCmd {
	return r.xTrim(ctx, key, MINID, false, minID, 0)
}

func (r *resp3) XTrimMaxLenApprox(ctx context.Context, key string, maxLen, limit int64) IntCmd {
	return r.xTrim(ctx, key, MAXLEN, true, str(maxLen), limit)
}

func (r *resp3) XTrimMinIDApprox(ctx context.Context, key string, minID string, limit int64) IntCmd {
	return r.xTrim(ctx, key, MINID, true, minID, limit)
}

func (r *resp3) Append(ctx context.Context, key, value string) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Append().Key(key).Value(value).Build()))
}

func (r *resp3) Decr(ctx context.Context, key string) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Decr().Key(key).Build()))
}

func (r *resp3) DecrBy(ctx context.Context, key string, decrement int64) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Decrby().Key(key).Decrement(decrement).Build()))
}

func (r *resp3) getGetCompleted(key string) rueidis.Completed {
	return r.cmd.B().Get().Key(key).Build()
}

func (r *resp3) Get(ctx context.Context, key string) StringCmd {
	return newStringCmd(r.cmd.Do(ctx, r.getGetCompleted(key)))
}

func (r *resp3Cache) Get(ctx context.Context, key string) StringCmd {
	return newStringCmd(r.Do(ctx, r.resp.getGetCompleted(key)))
}

func (r *resp3) GetDel(ctx context.Context, key string) StringCmd {
	return newStringCmd(r.cmd.Do(ctx, r.cmd.B().Getdel().Key(key).Build()))
}

func (r *resp3) GetEx(ctx context.Context, key string, expiration time.Duration) StringCmd {
	var completed rueidis.Completed
	var getexKey = r.cmd.B().Getex().Key(key)
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
	return newStringCmd(r.cmd.Do(ctx, completed))
}

func (r *resp3) getGetRangeCompleted(key string, start, end int64) rueidis.Completed {
	return r.cmd.B().Getrange().Key(key).Start(start).End(end).Build()
}

func (r *resp3) GetRange(ctx context.Context, key string, start, end int64) StringCmd {
	return newStringCmd(r.cmd.Do(ctx, r.getGetRangeCompleted(key, start, end)))
}

func (r *resp3Cache) GetRange(ctx context.Context, key string, start, end int64) StringCmd {
	return newStringCmd(r.Do(ctx, r.resp.getGetRangeCompleted(key, start, end)))
}

func (r *resp3) GetSet(ctx context.Context, key string, value interface{}) StringCmd {
	return newStringCmd(r.cmd.Do(ctx, r.cmd.B().Getset().Key(key).Value(str(value)).Build()))
}

func (r *resp3) Incr(ctx context.Context, key string) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Incr().Key(key).Build()))
}

func (r *resp3) IncrBy(ctx context.Context, key string, value int64) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Incrby().Key(key).Increment(value).Build()))
}

func (r *resp3) IncrByFloat(ctx context.Context, key string, value float64) FloatCmd {
	return newFloatCmd(r.cmd.Do(ctx, r.cmd.B().Incrbyfloat().Key(key).Increment(value).Build()))
}

func (r *resp3) getMGetCompleted(keys ...string) rueidis.Completed {
	return r.cmd.B().Mget().Key(keys...).Build()
}

func (r *resp3) MGet(ctx context.Context, keys ...string) SliceCmd {
	return newSliceCmd(r.cmd.Do(ctx, r.getMGetCompleted(keys...)))
}

func (r *resp3Cache) MGet(ctx context.Context, keys ...string) SliceCmd {
	return newSliceCmd(r.Do(ctx, r.resp.getMGetCompleted(keys...)))
}

func (r *resp3) MSet(ctx context.Context, values ...interface{}) StatusCmd {
	kv := r.cmd.B().Mset().KeyValue()
	args := argsToSlice(values)
	for i := 0; i < len(args); i += 2 {
		kv = kv.KeyValue(args[i], args[i+1])
	}
	return newStatusCmd(r.cmd.Do(ctx, kv.Build()))
}

func (r *resp3) MSetNX(ctx context.Context, values ...interface{}) BoolCmd {
	kv := r.cmd.B().Msetnx().KeyValue()
	args := argsToSlice(values)
	for i := 0; i < len(args); i += 2 {
		kv = kv.KeyValue(args[i], args[i+1])
	}
	return newBoolCmd(r.cmd.Do(ctx, kv.Build()))
}

func (r *resp3) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) StatusCmd {
	var completed rueidis.Completed
	var setValue = r.cmd.B().Set().Key(key).Value(str(value))
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
	return newStatusCmd(r.cmd.Do(ctx, completed))
}

func (r *resp3) SetEX(ctx context.Context, key string, value interface{}, expiration time.Duration) StatusCmd {
	return newStatusCmd(r.cmd.Do(ctx, r.cmd.B().Setex().Key(key).Seconds(formatSec(expiration)).Value(str(value)).Build()))
}

func (r *resp3) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) BoolCmd {
	var resp rueidis.RedisResult
	switch expiration {
	case 0:
		resp = r.cmd.Do(ctx, r.cmd.B().Setnx().Key(key).Value(str(value)).Build())
	case KeepTTL:
		resp = r.cmd.Do(ctx, r.cmd.B().Set().Key(key).Value(str(value)).Nx().Keepttl().Build())
	default:
		if usePrecise(expiration) {
			resp = r.cmd.Do(ctx, r.cmd.B().Set().Key(key).Value(str(value)).Nx().PxMilliseconds(formatMs(expiration)).Build())
		} else {
			resp = r.cmd.Do(ctx, r.cmd.B().Set().Key(key).Value(str(value)).Nx().ExSeconds(formatSec(expiration)).Build())
		}
	}
	return newBoolCmd(resp)
}

func (r *resp3) SetXX(ctx context.Context, key string, value interface{}, expiration time.Duration) BoolCmd {
	var resp rueidis.RedisResult
	switch expiration {
	case 0:
		resp = r.cmd.Do(ctx, r.cmd.B().Set().Key(key).Value(str(value)).Xx().Build())
	case KeepTTL:
		resp = r.cmd.Do(ctx, r.cmd.B().Set().Key(key).Value(str(value)).Xx().Keepttl().Build())
	default:
		if usePrecise(expiration) {
			resp = r.cmd.Do(ctx, r.cmd.B().Set().Key(key).Value(str(value)).Xx().PxMilliseconds(formatMs(expiration)).Build())
		} else {
			resp = r.cmd.Do(ctx, r.cmd.B().Set().Key(key).Value(str(value)).Xx().ExSeconds(formatSec(expiration)).Build())
		}
	}
	return newBoolCmd(resp)
}

func (r *resp3) SetArgs(ctx context.Context, key string, value interface{}, a SetArgs) StatusCmd {
	cmd := r.cmd.B().Arbitrary(SET).Keys(key).Args(str(value))
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
	return newStatusCmd(r.cmd.Do(ctx, cmd.Build()))
}

func (r *resp3) SetRange(ctx context.Context, key string, offset int64, value string) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.cmd.B().Setrange().Key(key).Offset(offset).Value(value).Build()))
}

func (r *resp3) getStrLenCompleted(key string) rueidis.Completed {
	return r.cmd.B().Strlen().Key(key).Build()
}

func (r *resp3) StrLen(ctx context.Context, key string) IntCmd {
	return newIntCmd(r.cmd.Do(ctx, r.getStrLenCompleted(key)))
}

func (r *resp3Cache) StrLen(ctx context.Context, key string) IntCmd {
	return newIntCmd(r.Do(ctx, r.resp.getStrLenCompleted(key)))
}
