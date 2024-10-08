package redisson

import (
	"context"
	"net"
	"strings"
	"sync"
	"time"

	goredis "github.com/go-redis/redis/v8"
)

type resp2 struct {
	v       ConfVisitor
	cmd     goredis.UniversalClient
	handler handler
}

func connectResp2(v ConfVisitor, h handler) (*resp2, error) {
	var opts = &goredis.UniversalOptions{
		Addrs:        v.GetAddrs(),
		DB:           v.GetDB(),
		Username:     v.GetUsername(),
		Password:     v.GetPassword(),
		ReadTimeout:  v.GetReadTimeout(),
		WriteTimeout: v.GetWriteTimeout(),
		PoolSize:     v.GetConnPoolSize(),
		MinIdleConns: v.GetMinIdleConns(),
		MaxConnAge:   v.GetConnMaxAge(),
		IdleTimeout:  v.GetIdleConnTimeout(),
		PoolTimeout:  v.GetConnPoolTimeout(),
		MasterName:   v.GetMasterName(),
	}
	if strings.ToLower(v.GetNet()) == "unix" {
		opts.Dialer = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.Dial(network, addr)
		}
	}
	var cmd goredis.UniversalClient
	if v.GetCluster() {
		cmd = goredis.NewClusterClient(opts.Cluster())
	} else {
		cmd = goredis.NewUniversalClient(opts)
	}
	return &resp2{cmd: cmd, v: v, handler: h}, nil
}

func (r *resp2) PoolStats() PoolStats                    { return *r.cmd.PoolStats() }
func (r *resp2) Close() error                            { return r.cmd.Close() }
func (r *resp2) RegisterCollector(RegisterCollectorFunc) {}
func (r *resp2) Cache(_ time.Duration) CacheCmdable      { return r }
func (r *resp2) Options() ConfVisitor                    { return r.v }
func (r *resp2) IsCluster() bool                         { return r.handler.isCluster() }
func (r *resp2) ForEachNodes(ctx context.Context, f func(context.Context, Cmdable) error) error {
	return r.cmd.(*goredis.ClusterClient).ForEachMaster(ctx, func(ctx context.Context, client *goredis.Client) error {
		return f(ctx, &resp2{cmd: client, v: r.v, handler: r.handler})
	})
}
func (r *resp2) BitCount(ctx context.Context, key string, bitCount *BitCount) IntCmd {
	return r.cmd.BitCount(ctx, key, bitCount)
}

func (r *resp2) BitField(ctx context.Context, key string, args ...interface{}) IntSliceCmd {
	return r.cmd.BitField(ctx, key, args...)
}

func (r *resp2) BitOpAnd(ctx context.Context, destKey string, keys ...string) IntCmd {
	return r.cmd.BitOpAnd(ctx, destKey, keys...)
}

func (r *resp2) BitOpOr(ctx context.Context, destKey string, keys ...string) IntCmd {
	return r.cmd.BitOpOr(ctx, destKey, keys...)
}

func (r *resp2) BitOpXor(ctx context.Context, destKey string, keys ...string) IntCmd {
	return r.cmd.BitOpXor(ctx, destKey, keys...)
}

func (r *resp2) BitOpNot(ctx context.Context, destKey string, key string) IntCmd {
	return r.cmd.BitOpNot(ctx, destKey, key)
}

func (r *resp2) BitPos(ctx context.Context, key string, bit int64, pos ...int64) IntCmd {
	return r.cmd.BitPos(ctx, key, bit, pos...)
}

func (r *resp2) GetBit(ctx context.Context, key string, offset int64) IntCmd {
	return r.cmd.GetBit(ctx, key, offset)
}

func (r *resp2) SetBit(ctx context.Context, key string, offset int64, value int) IntCmd {
	return r.cmd.SetBit(ctx, key, offset, value)
}

func (r *resp2) ClusterAddSlots(ctx context.Context, slots ...int) StatusCmd {
	return r.cmd.ClusterAddSlots(ctx, slots...)
}

func (r *resp2) ClusterAddSlotsRange(ctx context.Context, min, max int) StatusCmd {
	return r.cmd.ClusterAddSlotsRange(ctx, min, max)
}

func (r *resp2) ClusterCountFailureReports(ctx context.Context, nodeID string) IntCmd {
	return r.cmd.ClusterCountFailureReports(ctx, nodeID)
}

func (r *resp2) ClusterCountKeysInSlot(ctx context.Context, slot int) IntCmd {
	return r.cmd.ClusterCountKeysInSlot(ctx, slot)
}

func (r *resp2) ClusterDelSlots(ctx context.Context, slots ...int) StatusCmd {
	return r.cmd.ClusterDelSlots(ctx, slots...)
}

func (r *resp2) ClusterDelSlotsRange(ctx context.Context, min, max int) StatusCmd {
	return r.cmd.ClusterDelSlotsRange(ctx, min, max)
}

func (r *resp2) ClusterFailover(ctx context.Context) StatusCmd {
	return r.cmd.ClusterFailover(ctx)
}

func (r *resp2) ClusterForget(ctx context.Context, nodeID string) StatusCmd {
	return r.cmd.ClusterForget(ctx, nodeID)
}

func (r *resp2) ClusterGetKeysInSlot(ctx context.Context, slot int, count int) StringSliceCmd {
	return r.cmd.ClusterGetKeysInSlot(ctx, slot, count)
}

func (r *resp2) ClusterInfo(ctx context.Context) StringCmd {
	return r.cmd.ClusterInfo(ctx)
}

func (r *resp2) ClusterKeySlot(ctx context.Context, key string) IntCmd {
	return r.cmd.ClusterKeySlot(ctx, key)
}

func (r *resp2) ClusterMeet(ctx context.Context, host, port string) StatusCmd {
	return r.cmd.ClusterMeet(ctx, host, port)
}

func (r *resp2) ClusterNodes(ctx context.Context) StringCmd {
	return r.cmd.ClusterNodes(ctx)
}

func (r *resp2) ClusterReplicate(ctx context.Context, nodeID string) StatusCmd {
	return r.cmd.ClusterReplicate(ctx, nodeID)
}

func (r *resp2) ClusterResetSoft(ctx context.Context) StatusCmd {
	return r.cmd.ClusterResetSoft(ctx)
}

func (r *resp2) ClusterResetHard(ctx context.Context) StatusCmd {
	return r.cmd.ClusterResetHard(ctx)
}

func (r *resp2) ClusterSaveConfig(ctx context.Context) StatusCmd {
	return r.cmd.ClusterSaveConfig(ctx)
}

func (r *resp2) ClusterSlaves(ctx context.Context, nodeID string) StringSliceCmd {
	return r.cmd.ClusterSlaves(ctx, nodeID)
}

func (r *resp2) ClusterSlots(ctx context.Context) ClusterSlotsCmd {
	return r.cmd.ClusterSlots(ctx)
}

func (r *resp2) ReadOnly(ctx context.Context) StatusCmd {
	return r.cmd.ReadOnly(ctx)
}

func (r *resp2) ReadWrite(ctx context.Context) StatusCmd {
	return r.cmd.ReadWrite(ctx)
}

func (r *resp2) Select(ctx context.Context, index int) StatusCmd {
	_, err := r.cmd.Pipelined(ctx, func(pip goredis.Pipeliner) error {
		_ = pip.Select(ctx, index)
		return nil
	})
	if err != nil {
		return newStatusCmdWithError(err)
	}
	return newOKStatusCmd()
}

func (r *resp2) ClientGetName(ctx context.Context) StringCmd {
	return r.cmd.ClientGetName(ctx)
}

func (r *resp2) ClientID(ctx context.Context) IntCmd {
	return r.cmd.ClientID(ctx)
}

func (r *resp2) ClientKill(ctx context.Context, ipPort string) StatusCmd {
	return r.cmd.ClientKill(ctx, ipPort)
}

func (r *resp2) ClientKillByFilter(ctx context.Context, keys ...string) IntCmd {
	return r.cmd.ClientKillByFilter(ctx, keys...)
}

func (r *resp2) ClientList(ctx context.Context) StringCmd {
	return r.cmd.ClientList(ctx)
}

func (r *resp2) ClientPause(ctx context.Context, dur time.Duration) BoolCmd {
	return r.cmd.ClientPause(ctx, dur)
}

func (r *resp2) Echo(ctx context.Context, message interface{}) StringCmd {
	return r.cmd.Echo(ctx, message)
}

func (r *resp2) Ping(ctx context.Context) StatusCmd {
	return r.cmd.Ping(ctx)
}

func (r *resp2) Quit(ctx context.Context) StatusCmd {
	return r.cmd.Quit(ctx)
}

func (r *resp2) Copy(ctx context.Context, sourceKey string, destKey string, db int, replace bool) IntCmd {
	return r.cmd.Copy(ctx, sourceKey, destKey, db, replace)
}

func (r *resp2) Del(ctx context.Context, keys ...string) IntCmd {
	return r.cmd.Del(ctx, keys...)
}

func (r *resp2) Dump(ctx context.Context, key string) StringCmd {
	return r.cmd.Dump(ctx, key)
}

func (r *resp2) Exists(ctx context.Context, keys ...string) IntCmd {
	return r.cmd.Exists(ctx, keys...)
}

func (r *resp2) Expire(ctx context.Context, key string, expiration time.Duration) BoolCmd {
	return r.cmd.Expire(ctx, key, expiration)
}

func (r *resp2) ExpireAt(ctx context.Context, key string, tm time.Time) BoolCmd {
	return r.cmd.ExpireAt(ctx, key, tm)
}

func (r *resp2) Keys(ctx context.Context, pattern string) StringSliceCmd {
	return r.cmd.Keys(ctx, pattern)
}

func (r *resp2) Migrate(ctx context.Context, host, port, key string, db int, timeout time.Duration) StatusCmd {
	return r.cmd.Migrate(ctx, host, port, key, db, timeout)
}

func (r *resp2) Move(ctx context.Context, key string, db int) BoolCmd {
	return r.cmd.Move(ctx, key, db)
}

func (r *resp2) ObjectRefCount(ctx context.Context, key string) IntCmd {
	return r.cmd.ObjectRefCount(ctx, key)
}

func (r *resp2) ObjectEncoding(ctx context.Context, key string) StringCmd {
	return r.cmd.ObjectEncoding(ctx, key)
}

func (r *resp2) ObjectIdleTime(ctx context.Context, key string) DurationCmd {
	return r.cmd.ObjectIdleTime(ctx, key)
}

func (r *resp2) Persist(ctx context.Context, key string) BoolCmd {
	return r.cmd.Persist(ctx, key)
}

func (r *resp2) PExpire(ctx context.Context, key string, expiration time.Duration) BoolCmd {
	return r.cmd.PExpire(ctx, key, expiration)
}

func (r *resp2) PExpireAt(ctx context.Context, key string, tm time.Time) BoolCmd {
	return r.cmd.PExpireAt(ctx, key, tm)
}

func (r *resp2) PTTL(ctx context.Context, key string) DurationCmd {
	return r.cmd.PTTL(ctx, key)
}

func (r *resp2) RandomKey(ctx context.Context) StringCmd {
	return r.cmd.RandomKey(ctx)
}

func (r *resp2) Rename(ctx context.Context, key, newkey string) StatusCmd {
	return r.cmd.Rename(ctx, key, newkey)
}

func (r *resp2) RenameNX(ctx context.Context, key, newkey string) BoolCmd {
	return r.cmd.RenameNX(ctx, key, newkey)
}

func (r *resp2) Restore(ctx context.Context, key string, ttl time.Duration, value string) StatusCmd {
	return r.cmd.Restore(ctx, key, ttl, value)
}

func (r *resp2) RestoreReplace(ctx context.Context, key string, ttl time.Duration, value string) StatusCmd {
	return r.cmd.RestoreReplace(ctx, key, ttl, value)
}

func (r *resp2) Scan(ctx context.Context, cursor uint64, match string, count int64) ScanCmd {
	return r.cmd.Scan(ctx, cursor, match, count)
}

func (r *resp2) ScanType(ctx context.Context, cursor uint64, match string, count int64, keyType string) ScanCmd {
	return r.cmd.ScanType(ctx, cursor, match, count, keyType)
}

func (r *resp2) Sort(ctx context.Context, key string, sort Sort) StringSliceCmd {
	return r.cmd.Sort(ctx, key, &sort)
}

func (r *resp2) SortStore(ctx context.Context, key, store string, sort Sort) IntCmd {
	return r.cmd.SortStore(ctx, key, store, &sort)
}

func (r *resp2) SortInterfaces(ctx context.Context, key string, sort Sort) SliceCmd {
	return r.cmd.SortInterfaces(ctx, key, &sort)
}

func (r *resp2) Touch(ctx context.Context, keys ...string) IntCmd {
	return r.cmd.Touch(ctx, keys...)
}

func (r *resp2) TTL(ctx context.Context, key string) DurationCmd {
	return r.cmd.TTL(ctx, key)
}

func (r *resp2) Type(ctx context.Context, key string) StatusCmd {
	return r.cmd.Type(ctx, key)
}

func (r *resp2) Unlink(ctx context.Context, keys ...string) IntCmd {
	return r.cmd.Unlink(ctx, keys...)
}

func (r *resp2) GeoAdd(ctx context.Context, key string, geoLocation ...GeoLocation) IntCmd {
	locs := make([]*GeoLocation, 0, len(geoLocation))
	for _, v := range geoLocation {
		locs = append(locs, &GeoLocation{
			Name:      v.Name,
			Longitude: v.Longitude, Latitude: v.Latitude, Dist: v.Dist,
			GeoHash: v.GeoHash,
		})
	}
	return r.cmd.GeoAdd(ctx, key, locs...)
}

func (r *resp2) GeoDist(ctx context.Context, key string, member1, member2, unit string) FloatCmd {
	return r.cmd.GeoDist(ctx, key, member1, member2, unit)
}

func (r *resp2) GeoHash(ctx context.Context, key string, members ...string) StringSliceCmd {
	return r.cmd.GeoHash(ctx, key, members...)
}

func (r *resp2) GeoPos(ctx context.Context, key string, members ...string) GeoPosCmd {
	return r.cmd.GeoPos(ctx, key, members...)
}

func (r *resp2) GeoRadius(ctx context.Context, key string, longitude, latitude float64, q GeoRadiusQuery) GeoLocationCmd {
	return r.cmd.GeoRadius(ctx, key, longitude, latitude, &q)
}

func (r *resp2) GeoRadiusStore(ctx context.Context, key string, longitude, latitude float64, q GeoRadiusQuery) IntCmd {
	return r.cmd.GeoRadiusStore(ctx, key, longitude, latitude, &q)
}

func (r *resp2) GeoRadiusByMember(ctx context.Context, key, member string, q GeoRadiusQuery) GeoLocationCmd {
	return r.cmd.GeoRadiusByMember(ctx, key, member, &q)
}

func (r *resp2) GeoRadiusByMemberStore(ctx context.Context, key, member string, q GeoRadiusQuery) IntCmd {
	return r.cmd.GeoRadiusByMemberStore(ctx, key, member, &q)
}

func (r *resp2) GeoSearch(ctx context.Context, key string, q GeoSearchQuery) StringSliceCmd {
	return r.cmd.GeoSearch(ctx, key, &q)
}

func (r *resp2) GeoSearchLocation(ctx context.Context, key string, q GeoSearchLocationQuery) GeoSearchLocationCmd {
	return r.cmd.GeoSearchLocation(ctx, key, &q)
}

func (r *resp2) GeoSearchStore(ctx context.Context, src, dest string, q GeoSearchStoreQuery) IntCmd {
	return r.cmd.GeoSearchStore(ctx, src, dest, &q)
}

func (r *resp2) HDel(ctx context.Context, key string, fields ...string) IntCmd {
	return r.cmd.HDel(ctx, key, fields...)
}

func (r *resp2) HExists(ctx context.Context, key, field string) BoolCmd {
	return r.cmd.HExists(ctx, key, field)
}

func (r *resp2) HGet(ctx context.Context, key, field string) StringCmd {
	return r.cmd.HGet(ctx, key, field)
}

func (r *resp2) HGetAll(ctx context.Context, key string) StringStringMapCmd {
	return r.cmd.HGetAll(ctx, key)
}

func (r *resp2) HIncrBy(ctx context.Context, key, field string, incr int64) IntCmd {
	return r.cmd.HIncrBy(ctx, key, field, incr)
}

func (r *resp2) HIncrByFloat(ctx context.Context, key, field string, incr float64) FloatCmd {
	return r.cmd.HIncrByFloat(ctx, key, field, incr)
}

func (r *resp2) HKeys(ctx context.Context, key string) StringSliceCmd {
	return r.cmd.HKeys(ctx, key)
}

func (r *resp2) HLen(ctx context.Context, key string) IntCmd {
	return r.cmd.HLen(ctx, key)
}

func (r *resp2) HMGet(ctx context.Context, key string, fields ...string) SliceCmd {
	return r.cmd.HMGet(ctx, key, fields...)
}

func (r *resp2) HMSet(ctx context.Context, key string, values ...interface{}) BoolCmd {
	return r.cmd.HMSet(ctx, key, values...)
}

func (r *resp2) HRandField(ctx context.Context, key string, count int, withValues bool) StringSliceCmd {
	return r.cmd.HRandField(ctx, key, count, withValues)
}

func (r *resp2) HScan(ctx context.Context, key string, cursor uint64, match string, count int64) ScanCmd {
	return r.cmd.HScan(ctx, key, cursor, match, count)
}

func (r *resp2) HSet(ctx context.Context, key string, values ...interface{}) IntCmd {
	return r.cmd.HSet(ctx, key, values...)
}

func (r *resp2) HSetNX(ctx context.Context, key, field string, value interface{}) BoolCmd {
	return r.cmd.HSetNX(ctx, key, field, value)
}

func (r *resp2) HVals(ctx context.Context, key string) StringSliceCmd {
	return r.cmd.HVals(ctx, key)
}

func (r *resp2) PFAdd(ctx context.Context, key string, els ...interface{}) IntCmd {
	return r.cmd.PFAdd(ctx, key, els...)
}

func (r *resp2) PFCount(ctx context.Context, keys ...string) IntCmd {
	return r.cmd.PFCount(ctx, keys...)
}

func (r *resp2) PFMerge(ctx context.Context, dest string, keys ...string) StatusCmd {
	return r.cmd.PFMerge(ctx, dest, keys...)
}

func (r *resp2) BLMove(ctx context.Context, source, destination, srcpos, destpos string, timeout time.Duration) StringCmd {
	return r.cmd.BLMove(ctx, source, destination, srcpos, destpos, timeout)
}

func (r *resp2) BLPop(ctx context.Context, timeout time.Duration, keys ...string) StringSliceCmd {
	return r.cmd.BLPop(ctx, timeout, keys...)
}

func (r *resp2) BRPop(ctx context.Context, timeout time.Duration, keys ...string) StringSliceCmd {
	return r.cmd.BRPop(ctx, timeout, keys...)
}

func (r *resp2) BRPopLPush(ctx context.Context, source, destination string, timeout time.Duration) StringCmd {
	return r.cmd.BRPopLPush(ctx, source, destination, timeout)
}

func (r *resp2) LIndex(ctx context.Context, key string, index int64) StringCmd {
	return r.cmd.LIndex(ctx, key, index)
}

func (r *resp2) LInsert(ctx context.Context, key, op string, pivot, value interface{}) IntCmd {
	return r.cmd.LInsert(ctx, key, op, pivot, value)
}

func (r *resp2) LInsertBefore(ctx context.Context, key string, pivot, value interface{}) IntCmd {
	return r.cmd.LInsertBefore(ctx, key, pivot, value)
}

func (r *resp2) LInsertAfter(ctx context.Context, key string, pivot, value interface{}) IntCmd {
	return r.cmd.LInsertAfter(ctx, key, pivot, value)
}

func (r *resp2) LLen(ctx context.Context, key string) IntCmd {
	return r.cmd.LLen(ctx, key)
}

func (r *resp2) LMove(ctx context.Context, source, destination, srcpos, destpos string) StringCmd {
	return r.cmd.LMove(ctx, source, destination, srcpos, destpos)
}

func (r *resp2) LPop(ctx context.Context, key string) StringCmd {
	return r.cmd.LPop(ctx, key)
}

func (r *resp2) LPopCount(ctx context.Context, key string, count int) StringSliceCmd {
	return r.cmd.LPopCount(ctx, key, count)
}

func (r *resp2) LPos(ctx context.Context, key string, value string, args LPosArgs) IntCmd {
	return r.cmd.LPos(ctx, key, value, args)
}

func (r *resp2) LPosCount(ctx context.Context, key string, value string, count int64, args LPosArgs) IntSliceCmd {
	return r.cmd.LPosCount(ctx, key, value, count, args)
}

func (r *resp2) LPush(ctx context.Context, key string, values ...interface{}) IntCmd {
	return r.cmd.LPush(ctx, key, values...)
}

func (r *resp2) LPushX(ctx context.Context, key string, values ...interface{}) IntCmd {
	return r.cmd.LPushX(ctx, key, values...)
}

func (r *resp2) LRange(ctx context.Context, key string, start, stop int64) StringSliceCmd {
	return r.cmd.LRange(ctx, key, start, stop)
}

func (r *resp2) LRem(ctx context.Context, key string, count int64, value interface{}) IntCmd {
	return r.cmd.LRem(ctx, key, count, value)
}

func (r *resp2) LSet(ctx context.Context, key string, index int64, value interface{}) StatusCmd {
	return r.cmd.LSet(ctx, key, index, value)
}

func (r *resp2) LTrim(ctx context.Context, key string, start, stop int64) StatusCmd {
	return r.cmd.LTrim(ctx, key, start, stop)
}

func (r *resp2) RPop(ctx context.Context, key string) StringCmd {
	return r.cmd.RPop(ctx, key)
}

func (r *resp2) RPopCount(ctx context.Context, key string, count int) StringSliceCmd {
	return r.cmd.RPopCount(ctx, key, count)
}

func (r *resp2) RPopLPush(ctx context.Context, source, destination string) StringCmd {
	return r.cmd.RPopLPush(ctx, source, destination)
}

func (r *resp2) RPush(ctx context.Context, key string, values ...interface{}) IntCmd {
	return r.cmd.RPush(ctx, key, values...)
}

func (r *resp2) RPushX(ctx context.Context, key string, values ...interface{}) IntCmd {
	return r.cmd.RPushX(ctx, key, values...)
}

type pipeCommand struct {
	cmd  []string
	keys []string
	args []interface{}
}

type pipelineResp2 struct {
	resp     *resp2
	commands []pipeCommand
	mx       sync.RWMutex
}

type pipelineCommand struct{}

func (pipelineCommand) String() string         { return "PIPELINE" }
func (pipelineCommand) Class() string          { return "Pipeline" }
func (pipelineCommand) RequireVersion() string { return "0.0.0" }
func (pipelineCommand) Forbid() bool           { return false }
func (pipelineCommand) WarnVersion() string    { return "" }
func (pipelineCommand) Warning() string        { return "" }
func (pipelineCommand) Cmd() []string          { return nil }

var pipelineCmd = &pipelineCommand{}

func (r *resp2) Pipeline() Pipeliner { return &pipelineResp2{resp: r} }

func (p *pipelineResp2) Put(_ context.Context, cmd Command, keys []string, args ...interface{}) (err error) {
	p.mx.Lock()
	p.commands = append(p.commands, pipeCommand{cmd: cmd.Cmd(), keys: keys, args: args})
	p.mx.Unlock()
	return
}

func (p *pipelineResp2) Exec(ctx context.Context) ([]interface{}, error) {
	var cancel context.CancelFunc
	ctx, cancel = p.resp.handler.before(ctx, pipelineCmd)
	defer cancel()
	res, err := p.resp.cmd.Pipelined(ctx, func(pip goredis.Pipeliner) error {
		p.mx.RLock()
		defer p.mx.RUnlock()
		for _, cmd := range p.commands {
			args := make([]interface{}, 0, 1+len(cmd.cmd)+len(cmd.keys)+len(cmd.args))
			for _, k := range cmd.cmd {
				args = append(args, k)
			}
			for _, k := range cmd.keys {
				args = append(args, k)
			}
			args = append(args, cmd.args...)
			_ = pip.Do(ctx, args...)
		}
		return nil
	})
	result := make([]interface{}, len(res))
	for i, j := range res {
		switch jj := j.(type) {
		case nil:
			result[i] = nil
		case error:
			result[i] = j
		case *goredis.Cmd:
			result[i] = jj.Val()
		default:
			result[i] = j
		}
	}
	p.resp.handler.after(ctx, err)
	return result, err
}

func (r *resp2) RawCmdable() interface{} { return r.cmd }

func (r *resp2) Publish(ctx context.Context, channel string, message interface{}) IntCmd {
	return r.cmd.Publish(ctx, channel, message)
}

func (r *resp2) PubSubChannels(ctx context.Context, pattern string) StringSliceCmd {
	return r.cmd.PubSubChannels(ctx, pattern)
}

func (r *resp2) PubSubNumPat(ctx context.Context) IntCmd {
	return r.cmd.PubSubNumPat(ctx)
}

func (r *resp2) PubSubNumSub(ctx context.Context, channels ...string) StringIntMapCmd {
	return r.cmd.PubSubNumSub(ctx, channels...)
}

func (r *resp2) Subscribe(ctx context.Context, channels ...string) PubSub {
	return newPubSubResp2(r.cmd.Subscribe(ctx, channels...), r.handler)
}

type pubSubResp2 struct {
	cmd     *goredis.PubSub
	handler handler
}

func newPubSubResp2(cmd *goredis.PubSub, handler handler) PubSub {
	return &pubSubResp2{cmd: cmd, handler: handler}
}

func (p *pubSubResp2) Close() error { return p.cmd.Close() }

func (p *pubSubResp2) Subscribe(ctx context.Context, channels ...string) error {
	var cancel context.CancelFunc
	ctx, cancel = p.handler.before(ctx, CommandSubscribe)
	err := p.cmd.Subscribe(ctx, channels...)
	p.handler.after(ctx, err)
	cancel()
	return err
}

func (p *pubSubResp2) PSubscribe(ctx context.Context, patterns ...string) error {
	var cancel context.CancelFunc
	ctx, cancel = p.handler.before(ctx, CommandPSubscribe)
	err := p.cmd.PSubscribe(ctx, patterns...)
	p.handler.after(ctx, err)
	cancel()
	return err
}

func (p *pubSubResp2) Unsubscribe(ctx context.Context, channels ...string) error {
	var cancel context.CancelFunc
	ctx, cancel = p.handler.before(ctx, CommandUnsubscribe)
	err := p.cmd.Unsubscribe(ctx, channels...)
	p.handler.after(ctx, err)
	cancel()
	return err
}

func (p *pubSubResp2) PUnsubscribe(ctx context.Context, patterns ...string) error {
	var cancel context.CancelFunc
	ctx, cancel = p.handler.before(ctx, CommandPUnsubscribe)
	err := p.cmd.PUnsubscribe(ctx, patterns...)
	p.handler.after(ctx, err)
	cancel()
	return err
}

func (p *pubSubResp2) Channel() <-chan *Message {
	return p.cmd.Channel()
}

func (r *resp2) CreateScript(string) Scripter { return nil }

func (r *resp2) Eval(ctx context.Context, script string, keys []string, args ...interface{}) Cmd {
	return r.cmd.Eval(ctx, script, keys, args...)
}

func (r *resp2) EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) Cmd {
	return r.cmd.EvalSha(ctx, sha1, keys, args...)
}

func (r *resp2) ScriptExists(ctx context.Context, hashes ...string) BoolSliceCmd {
	return r.cmd.ScriptExists(ctx, hashes...)
}

func (r *resp2) ScriptFlush(ctx context.Context) StatusCmd {
	return r.cmd.ScriptFlush(ctx)
}

func (r *resp2) ScriptKill(ctx context.Context) StatusCmd {
	return r.cmd.ScriptKill(ctx)
}

func (r *resp2) ScriptLoad(ctx context.Context, script string) StringCmd {
	return r.cmd.ScriptLoad(ctx, script)
}

func (r *resp2) BgRewriteAOF(ctx context.Context) StatusCmd {
	return r.cmd.BgRewriteAOF(ctx)
}

func (r *resp2) BgSave(ctx context.Context) StatusCmd {
	return r.cmd.BgSave(ctx)
}

func (r *resp2) Command(ctx context.Context) CommandsInfoCmd {
	return r.cmd.Command(ctx)
}

func (r *resp2) ConfigGet(ctx context.Context, parameter string) SliceCmd {
	return r.cmd.ConfigGet(ctx, parameter)
}

func (r *resp2) ConfigResetStat(ctx context.Context) StatusCmd {
	return r.cmd.ConfigResetStat(ctx)
}

func (r *resp2) ConfigRewrite(ctx context.Context) StatusCmd {
	return r.cmd.ConfigResetStat(ctx)
}

func (r *resp2) ConfigSet(ctx context.Context, parameter, value string) StatusCmd {
	return r.cmd.ConfigSet(ctx, parameter, value)
}

func (r *resp2) DBSize(ctx context.Context) IntCmd {
	return r.cmd.DBSize(ctx)
}

func (r *resp2) FlushAll(ctx context.Context) StatusCmd {
	return r.cmd.FlushAll(ctx)
}

func (r *resp2) FlushAllAsync(ctx context.Context) StatusCmd {
	return r.cmd.FlushAllAsync(ctx)
}

func (r *resp2) FlushDB(ctx context.Context) StatusCmd {
	return r.cmd.FlushDB(ctx)
}

func (r *resp2) FlushDBAsync(ctx context.Context) StatusCmd {
	return r.cmd.FlushDBAsync(ctx)
}

func (r *resp2) Info(ctx context.Context, section ...string) StringCmd {
	return r.cmd.Info(ctx, section...)
}

func (r *resp2) LastSave(ctx context.Context) IntCmd {
	return r.cmd.LastSave(ctx)
}

func (r *resp2) MemoryUsage(ctx context.Context, key string, samples ...int) IntCmd {
	return r.cmd.MemoryUsage(ctx, key, samples...)
}

func (r *resp2) Save(ctx context.Context) StatusCmd {
	return r.cmd.Save(ctx)
}

func (r *resp2) Shutdown(ctx context.Context) StatusCmd {
	return r.cmd.Shutdown(ctx)
}

func (r *resp2) ShutdownSave(ctx context.Context) StatusCmd {
	return r.cmd.ShutdownSave(ctx)
}

func (r *resp2) ShutdownNoSave(ctx context.Context) StatusCmd {
	return r.cmd.ShutdownNoSave(ctx)
}

func (r *resp2) SlaveOf(ctx context.Context, host, port string) StatusCmd {
	return r.cmd.SlaveOf(ctx, host, port)
}

func (r *resp2) Time(ctx context.Context) TimeCmd {
	return r.cmd.Time(ctx)
}

func (r *resp2) DebugObject(ctx context.Context, key string) StringCmd {
	return r.cmd.DebugObject(ctx, key)
}

func (r *resp2) SAdd(ctx context.Context, key string, members ...interface{}) IntCmd {
	return r.cmd.SAdd(ctx, key, members...)
}

func (r *resp2) SCard(ctx context.Context, key string) IntCmd {
	return r.cmd.SCard(ctx, key)
}

func (r *resp2) SDiff(ctx context.Context, keys ...string) StringSliceCmd {
	return r.cmd.SDiff(ctx, keys...)
}

func (r *resp2) SDiffStore(ctx context.Context, destination string, keys ...string) IntCmd {
	return r.cmd.SDiffStore(ctx, destination, keys...)
}

func (r *resp2) SInter(ctx context.Context, keys ...string) StringSliceCmd {
	return r.cmd.SInter(ctx, keys...)
}

func (r *resp2) SInterStore(ctx context.Context, destination string, keys ...string) IntCmd {
	return r.cmd.SInterStore(ctx, destination, keys...)
}

func (r *resp2) SIsMember(ctx context.Context, key string, member interface{}) BoolCmd {
	return r.cmd.SIsMember(ctx, key, member)
}

func (r *resp2) SMIsMember(ctx context.Context, key string, members ...interface{}) BoolSliceCmd {
	return r.cmd.SMIsMember(ctx, key, members...)
}

func (r *resp2) SMembers(ctx context.Context, key string) StringSliceCmd {
	return r.cmd.SMembers(ctx, key)
}

func (r *resp2) SMembersMap(ctx context.Context, key string) StringStructMapCmd {
	return r.cmd.SMembersMap(ctx, key)
}

func (r *resp2) SMove(ctx context.Context, source, destination string, member interface{}) BoolCmd {
	return r.cmd.SMove(ctx, source, destination, member)
}

func (r *resp2) SPop(ctx context.Context, key string) StringCmd {
	return r.cmd.SPop(ctx, key)
}

func (r *resp2) SPopN(ctx context.Context, key string, count int64) StringSliceCmd {
	return r.cmd.SPopN(ctx, key, count)
}

func (r *resp2) SRandMember(ctx context.Context, key string) StringCmd {
	return r.cmd.SRandMember(ctx, key)
}

func (r *resp2) SRandMemberN(ctx context.Context, key string, count int64) StringSliceCmd {
	return r.cmd.SRandMemberN(ctx, key, count)
}

func (r *resp2) SRem(ctx context.Context, key string, members ...interface{}) IntCmd {
	return r.cmd.SRem(ctx, key, members...)
}

func (r *resp2) SScan(ctx context.Context, key string, cursor uint64, match string, count int64) ScanCmd {
	return r.cmd.SScan(ctx, key, cursor, match, count)
}

func (r *resp2) SUnion(ctx context.Context, keys ...string) StringSliceCmd {
	return r.cmd.SUnion(ctx, keys...)
}

func (r *resp2) SUnionStore(ctx context.Context, destination string, keys ...string) IntCmd {
	return r.cmd.SUnionStore(ctx, destination, keys...)
}

func (r *resp2) BZPopMax(ctx context.Context, timeout time.Duration, keys ...string) ZWithKeyCmd {
	return r.cmd.BZPopMax(ctx, timeout, keys...)
}

func (r *resp2) BZPopMin(ctx context.Context, timeout time.Duration, keys ...string) ZWithKeyCmd {
	return r.cmd.BZPopMin(ctx, timeout, keys...)
}

func (r *resp2) toZs(members ...Z) []*Z {
	zs := make([]*Z, 0, len(members))
	for _, v := range members {
		zs = append(zs, &Z{Score: v.Score, Member: v.Member})
	}
	return zs
}

func (r *resp2) ZAdd(ctx context.Context, key string, members ...Z) IntCmd {
	return r.cmd.ZAdd(ctx, key, r.toZs(members...)...)
}

func (r *resp2) ZAddNX(ctx context.Context, key string, members ...Z) IntCmd {
	return r.cmd.ZAddNX(ctx, key, r.toZs(members...)...)
}

func (r *resp2) ZAddXX(ctx context.Context, key string, members ...Z) IntCmd {
	return r.cmd.ZAddXX(ctx, key, r.toZs(members...)...)
}

func (r *resp2) ZAddCh(ctx context.Context, key string, members ...Z) IntCmd {
	return r.cmd.ZAddCh(ctx, key, r.toZs(members...)...)
}

func (r *resp2) ZAddNXCh(ctx context.Context, key string, members ...Z) IntCmd {
	return r.cmd.ZAddNXCh(ctx, key, r.toZs(members...)...)
}

func (r *resp2) ZAddXXCh(ctx context.Context, key string, members ...Z) IntCmd {
	return r.cmd.ZAddXXCh(ctx, key, r.toZs(members...)...)
}

func (r *resp2) ZAddArgs(ctx context.Context, key string, args ZAddArgs) IntCmd {
	return r.cmd.ZAddArgs(ctx, key, args)
}

func (r *resp2) ZAddArgsIncr(ctx context.Context, key string, args ZAddArgs) FloatCmd {
	return r.cmd.ZAddArgsIncr(ctx, key, args)
}

func (r *resp2) ZCard(ctx context.Context, key string) IntCmd {
	return r.cmd.ZCard(ctx, key)
}

func (r *resp2) ZCount(ctx context.Context, key, min, max string) IntCmd {
	return r.cmd.ZCount(ctx, key, min, max)
}

func (r *resp2) ZDiff(ctx context.Context, keys ...string) StringSliceCmd {
	return r.cmd.ZDiff(ctx, keys...)
}

func (r *resp2) ZDiffWithScores(ctx context.Context, keys ...string) ZSliceCmd {
	return r.cmd.ZDiffWithScores(ctx, keys...)
}

func (r *resp2) ZDiffStore(ctx context.Context, destination string, keys ...string) IntCmd {
	return r.cmd.ZDiffStore(ctx, destination, keys...)
}

func (r *resp2) ZIncr(ctx context.Context, key string, member Z) FloatCmd {
	return r.cmd.ZIncr(ctx, key, &member)
}

func (r *resp2) ZIncrNX(ctx context.Context, key string, member Z) FloatCmd {
	return r.cmd.ZIncrNX(ctx, key, &member)
}

func (r *resp2) ZIncrXX(ctx context.Context, key string, member Z) FloatCmd {
	return r.cmd.ZIncrXX(ctx, key, &member)
}

func (r *resp2) ZIncrBy(ctx context.Context, key string, increment float64, member string) FloatCmd {
	return r.cmd.ZIncrBy(ctx, key, increment, member)
}

func (r *resp2) ZInter(ctx context.Context, store ZStore) StringSliceCmd {
	return r.cmd.ZInter(ctx, &store)
}

func (r *resp2) ZInterWithScores(ctx context.Context, store ZStore) ZSliceCmd {
	return r.cmd.ZInterWithScores(ctx, &store)
}

func (r *resp2) ZInterStore(ctx context.Context, destination string, store ZStore) IntCmd {
	return r.cmd.ZInterStore(ctx, destination, &store)
}

func (r *resp2) ZLexCount(ctx context.Context, key, min, max string) IntCmd {
	return r.cmd.ZLexCount(ctx, key, min, max)
}

func (r *resp2) ZMScore(ctx context.Context, key string, members ...string) FloatSliceCmd {
	return r.cmd.ZMScore(ctx, key, members...)
}

func (r *resp2) ZPopMax(ctx context.Context, key string, count ...int64) ZSliceCmd {
	return r.cmd.ZPopMax(ctx, key, count...)
}

func (r *resp2) ZPopMin(ctx context.Context, key string, count ...int64) ZSliceCmd {
	return r.cmd.ZPopMin(ctx, key, count...)
}

func (r *resp2) ZRandMember(ctx context.Context, key string, count int, withScores bool) StringSliceCmd {
	return r.cmd.ZRandMember(ctx, key, count, withScores)
}

func (r *resp2) ZRange(ctx context.Context, key string, start, stop int64) StringSliceCmd {
	return r.cmd.ZRange(ctx, key, start, stop)
}

func (r *resp2) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ZSliceCmd {
	return r.cmd.ZRangeWithScores(ctx, key, start, stop)
}

func (r *resp2) ZRangeByLex(ctx context.Context, key string, opt ZRangeBy) StringSliceCmd {
	return r.cmd.ZRangeByLex(ctx, key, &opt)
}

func (r *resp2) ZRangeByScore(ctx context.Context, key string, opt ZRangeBy) StringSliceCmd {
	return r.cmd.ZRangeByScore(ctx, key, &opt)
}

func (r *resp2) ZRangeByScoreWithScores(ctx context.Context, key string, opt ZRangeBy) ZSliceCmd {
	return r.cmd.ZRangeByScoreWithScores(ctx, key, &opt)
}

func (r *resp2) ZRangeArgs(ctx context.Context, z ZRangeArgs) StringSliceCmd {
	return r.cmd.ZRangeArgs(ctx, z)
}

func (r *resp2) ZRangeArgsWithScores(ctx context.Context, z ZRangeArgs) ZSliceCmd {
	return r.cmd.ZRangeArgsWithScores(ctx, z)
}

func (r *resp2) ZRangeStore(ctx context.Context, dst string, z ZRangeArgs) IntCmd {
	return r.cmd.ZRangeStore(ctx, dst, z)
}

func (r *resp2) ZRank(ctx context.Context, key, member string) IntCmd {
	return r.cmd.ZRank(ctx, key, member)
}

func (r *resp2) ZRem(ctx context.Context, key string, members ...interface{}) IntCmd {
	return r.cmd.ZRem(ctx, key, members...)
}

func (r *resp2) ZRemRangeByLex(ctx context.Context, key, min, max string) IntCmd {
	return r.cmd.ZRemRangeByLex(ctx, key, min, max)
}

func (r *resp2) ZRemRangeByRank(ctx context.Context, key string, start, stop int64) IntCmd {
	return r.cmd.ZRemRangeByRank(ctx, key, start, stop)
}

func (r *resp2) ZRemRangeByScore(ctx context.Context, key, min, max string) IntCmd {
	return r.cmd.ZRemRangeByScore(ctx, key, min, max)
}

func (r *resp2) ZRevRange(ctx context.Context, key string, start, stop int64) StringSliceCmd {
	return r.cmd.ZRevRange(ctx, key, start, stop)
}

func (r *resp2) ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) ZSliceCmd {
	return r.cmd.ZRevRangeWithScores(ctx, key, start, stop)
}

func (r *resp2) ZRevRangeByLex(ctx context.Context, key string, opt ZRangeBy) StringSliceCmd {
	return r.cmd.ZRevRangeByLex(ctx, key, &opt)
}

func (r *resp2) ZRevRangeByScore(ctx context.Context, key string, opt ZRangeBy) StringSliceCmd {
	return r.cmd.ZRevRangeByScore(ctx, key, &opt)
}

func (r *resp2) ZRevRangeByScoreWithScores(ctx context.Context, key string, opt ZRangeBy) ZSliceCmd {
	return r.cmd.ZRevRangeByScoreWithScores(ctx, key, &opt)
}

func (r *resp2) ZRevRank(ctx context.Context, key, member string) IntCmd {
	return r.cmd.ZRevRank(ctx, key, member)
}

func (r *resp2) ZScan(ctx context.Context, key string, cursor uint64, match string, count int64) ScanCmd {
	return r.cmd.ZScan(ctx, key, cursor, match, count)
}

func (r *resp2) ZScore(ctx context.Context, key, member string) FloatCmd {
	return r.cmd.ZScore(ctx, key, member)
}

func (r *resp2) ZUnion(ctx context.Context, store ZStore) StringSliceCmd {
	return r.cmd.ZUnion(ctx, store)
}

func (r *resp2) ZUnionWithScores(ctx context.Context, store ZStore) ZSliceCmd {
	return r.cmd.ZUnionWithScores(ctx, store)
}

func (r *resp2) ZUnionStore(ctx context.Context, dest string, store ZStore) IntCmd {
	return r.cmd.ZUnionStore(ctx, dest, &store)
}

func (r *resp2) XAck(ctx context.Context, stream, group string, ids ...string) IntCmd {
	return r.cmd.XAck(ctx, stream, group, ids...)
}

func (r *resp2) XAdd(ctx context.Context, a XAddArgs) StringCmd {
	return r.cmd.XAdd(ctx, &a)
}

func (r *resp2) XAutoClaim(ctx context.Context, a XAutoClaimArgs) XAutoClaimCmd {
	return r.cmd.XAutoClaim(ctx, &a)
}

func (r *resp2) XAutoClaimJustID(ctx context.Context, a XAutoClaimArgs) XAutoClaimJustIDCmd {
	return r.cmd.XAutoClaimJustID(ctx, &a)
}

func (r *resp2) XClaim(ctx context.Context, a XClaimArgs) XMessageSliceCmd {
	return r.cmd.XClaim(ctx, &a)
}

func (r *resp2) XClaimJustID(ctx context.Context, a XClaimArgs) StringSliceCmd {
	return r.cmd.XClaimJustID(ctx, &a)
}

func (r *resp2) XDel(ctx context.Context, stream string, ids ...string) IntCmd {
	return r.cmd.XDel(ctx, stream, ids...)
}

func (r *resp2) XGroupCreate(ctx context.Context, stream, group, start string) StatusCmd {
	return r.cmd.XGroupCreate(ctx, stream, group, start)
}

func (r *resp2) XGroupCreateMkStream(ctx context.Context, stream, group, start string) StatusCmd {
	return r.cmd.XGroupCreateMkStream(ctx, stream, group, start)
}

func (r *resp2) XGroupCreateConsumer(ctx context.Context, stream, group, consumer string) IntCmd {
	return r.cmd.XGroupCreateConsumer(ctx, stream, group, consumer)
}

func (r *resp2) XGroupDelConsumer(ctx context.Context, stream, group, consumer string) IntCmd {
	return r.cmd.XGroupDelConsumer(ctx, stream, group, consumer)
}

func (r *resp2) XGroupDestroy(ctx context.Context, stream, group string) IntCmd {
	return r.cmd.XGroupDestroy(ctx, stream, group)
}

func (r *resp2) XGroupSetID(ctx context.Context, stream, group, start string) StatusCmd {
	return r.cmd.XGroupSetID(ctx, stream, group, start)
}

func (r *resp2) XInfoConsumers(ctx context.Context, key string, group string) XInfoConsumersCmd {
	return r.cmd.XInfoConsumers(ctx, key, group)
}

func (r *resp2) XInfoGroups(ctx context.Context, key string) XInfoGroupsCmd {
	return r.cmd.XInfoGroups(ctx, key)
}

func (r *resp2) XInfoStream(ctx context.Context, key string) XInfoStreamCmd {
	return r.cmd.XInfoStream(ctx, key)
}

func (r *resp2) XInfoStreamFull(ctx context.Context, key string, count int) XInfoStreamFullCmd {
	return r.cmd.XInfoStreamFull(ctx, key, count)
}

func (r *resp2) XLen(ctx context.Context, stream string) IntCmd {
	return r.cmd.XLen(ctx, stream)
}

func (r *resp2) XPending(ctx context.Context, stream, group string) XPendingCmd {
	return r.cmd.XPending(ctx, stream, group)
}

func (r *resp2) XPendingExt(ctx context.Context, a XPendingExtArgs) XPendingExtCmd {
	return r.cmd.XPendingExt(ctx, &a)
}

func (r *resp2) XRange(ctx context.Context, stream, start, stop string) XMessageSliceCmd {
	return r.cmd.XRange(ctx, stream, start, stop)
}

func (r *resp2) XRangeN(ctx context.Context, stream, start, stop string, count int64) XMessageSliceCmd {
	return r.cmd.XRangeN(ctx, stream, start, stop, count)
}

func (r *resp2) XRead(ctx context.Context, a XReadArgs) XStreamSliceCmd {
	return r.cmd.XRead(ctx, &a)
}

func (r *resp2) XReadStreams(ctx context.Context, streams ...string) XStreamSliceCmd {
	return r.cmd.XReadStreams(ctx, streams...)
}

func (r *resp2) XReadGroup(ctx context.Context, a XReadGroupArgs) XStreamSliceCmd {
	return r.cmd.XReadGroup(ctx, &a)
}

func (r *resp2) XRevRange(ctx context.Context, stream string, start, stop string) XMessageSliceCmd {
	return r.cmd.XRevRange(ctx, stream, start, stop)
}

func (r *resp2) XRevRangeN(ctx context.Context, stream string, start, stop string, count int64) XMessageSliceCmd {
	return r.cmd.XRevRangeN(ctx, stream, start, stop, count)
}

func (r *resp2) XTrim(ctx context.Context, key string, maxLen int64) IntCmd {
	return r.cmd.XTrim(ctx, key, maxLen)
}

func (r *resp2) XTrimApprox(ctx context.Context, key string, maxLen int64) IntCmd {
	return r.cmd.XTrimApprox(ctx, key, maxLen)
}

func (r *resp2) XTrimMaxLen(ctx context.Context, key string, maxLen int64) IntCmd {
	return r.cmd.XTrimMaxLen(ctx, key, maxLen)
}

func (r *resp2) XTrimMinID(ctx context.Context, key string, minID string) IntCmd {
	return r.cmd.XTrimMinID(ctx, key, minID)
}

func (r *resp2) XTrimMaxLenApprox(ctx context.Context, key string, maxLen, limit int64) IntCmd {
	return r.cmd.XTrimMaxLenApprox(ctx, key, maxLen, limit)
}

func (r *resp2) XTrimMinIDApprox(ctx context.Context, key string, maxLen string, limit int64) IntCmd {
	return r.cmd.XTrimMinIDApprox(ctx, key, maxLen, limit)
}

func (r *resp2) Append(ctx context.Context, key, value string) IntCmd {
	return r.cmd.Append(ctx, key, value)
}

func (r *resp2) Decr(ctx context.Context, key string) IntCmd {
	return r.cmd.Decr(ctx, key)
}

func (r *resp2) DecrBy(ctx context.Context, key string, decrement int64) IntCmd {
	return r.cmd.DecrBy(ctx, key, decrement)
}

func (r *resp2) Get(ctx context.Context, key string) StringCmd {
	return r.cmd.Get(ctx, key)
}

func (r *resp2) GetDel(ctx context.Context, key string) StringCmd {
	return r.cmd.GetDel(ctx, key)
}

func (r *resp2) GetEx(ctx context.Context, key string, expiration time.Duration) StringCmd {
	return r.cmd.GetEx(ctx, key, expiration)
}

func (r *resp2) GetRange(ctx context.Context, key string, start, end int64) StringCmd {
	return r.cmd.GetRange(ctx, key, start, end)
}

func (r *resp2) GetSet(ctx context.Context, key string, value interface{}) StringCmd {
	return r.cmd.GetSet(ctx, key, value)
}

func (r *resp2) Incr(ctx context.Context, key string) IntCmd {
	return r.cmd.Incr(ctx, key)
}

func (r *resp2) IncrBy(ctx context.Context, key string, value int64) IntCmd {
	return r.cmd.IncrBy(ctx, key, value)
}

func (r *resp2) IncrByFloat(ctx context.Context, key string, value float64) FloatCmd {
	return r.cmd.IncrByFloat(ctx, key, value)
}

func (r *resp2) MGet(ctx context.Context, keys ...string) SliceCmd {
	return r.cmd.MGet(ctx, keys...)
}

func (r *resp2) MSet(ctx context.Context, values ...interface{}) StatusCmd {
	return r.cmd.MSet(ctx, values...)
}

func (r *resp2) MSetNX(ctx context.Context, values ...interface{}) BoolCmd {
	return r.cmd.MSetNX(ctx, values...)
}

func (r *resp2) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) StatusCmd {
	return r.cmd.Set(ctx, key, value, expiration)
}

func (r *resp2) SetEX(ctx context.Context, key string, value interface{}, expiration time.Duration) StatusCmd {
	return r.cmd.SetEX(ctx, key, value, expiration)
}

func (r *resp2) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) BoolCmd {
	return r.cmd.SetNX(ctx, key, value, expiration)
}

func (r *resp2) SetXX(ctx context.Context, key string, value interface{}, expiration time.Duration) BoolCmd {
	return r.cmd.SetXX(ctx, key, value, expiration)
}

func (r *resp2) SetArgs(ctx context.Context, key string, value interface{}, a SetArgs) StatusCmd {
	return r.cmd.SetArgs(ctx, key, value, a)
}

func (r *resp2) SetRange(ctx context.Context, key string, offset int64, value string) IntCmd {
	return r.cmd.SetRange(ctx, key, offset, value)
}

func (r *resp2) StrLen(ctx context.Context, key string) IntCmd {
	return r.cmd.StrLen(ctx, key)
}

func (r *resp2) Receive(context.Context, func(Message), ...string) error {
	panic("not implemented")
}
func (r *resp2) PReceive(context.Context, func(Message), ...string) error {
	panic("not implemented")
}
func (r *resp2) XMGet(context.Context, ...string) SliceCmd { return nil }
