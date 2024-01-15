package redisson

import (
	"context"
	"errors"
	"strconv"
	"time"
)

var ErrLeakyBucketOverflow = errors.New("leak bucket Overflow")

const luaFunnel = `
redis.replicate_commands()

local function now()
    local ts = redis.call('TIME')
    return tostring(ts[1] + ts[2] / 1000000)
end

local Funnel = {}

function Funnel:new(o, capacity, operations, seconds, left_quota, leaking_ts)
    o = o or {}
    setmetatable(o, self)
    self.__index = self
    self.capacity = capacity
    self.operations = operations
    self.seconds = seconds
    self.left_quota = left_quota
    self.leaking_ts = leaking_ts
    self.leaking_rate = operations / seconds
    return o
end

function Funnel:make_space(quota)
    local now_ts = now()
    local delta_ts = now_ts - self.leaking_ts
    local delta_quota = delta_ts * self.leaking_rate
    if (self.left_quota + delta_quota) < quota then
        return
    else
        self.left_quota = self.left_quota + delta_quota
        if self.left_quota > self.capacity then
            self.left_quota = self.capacity
        end
        self.leaking_ts = now_ts
    end
end

function Funnel:watering(quota)
    self:make_space(quota)
    if self.left_quota >= quota then
        self.left_quota = self.left_quota - quota
        return
            0,
            self.capacity,
            self.left_quota,
            tostring(-1.0),
            tostring((self.capacity - self.left_quota) / self.leaking_rate)
    else
        return
            1,
            self.capacity,
            self.left_quota,
            tostring(quota / self.leaking_rate),
            tostring((self.capacity - self.left_quota) / self.leaking_rate)
    end
end

local key =  KEYS[1]
local capacity = tonumber(ARGV[1])
local operations = tonumber(ARGV[2])
local seconds = tonumber(ARGV[3])
local quota = tonumber(ARGV[4])

local left_quota
local leaking_ts
local cache = redis.call('HMGET', key, 'left_quota', 'leaking_ts')
if cache[1] ~= false then
    left_quota = tonumber(cache[1])
    if left_quota > capacity then
        left_quota = capacity
    end
    leaking_ts = cache[2]
else
    left_quota = 0
    leaking_ts = now()
end

local funnel = Funnel:new(nil, capacity, operations, seconds, left_quota, leaking_ts)
local ready, capacity, left_quota, interval, empty_time = funnel:watering(quota)

redis.call('HMSET', key,
    'left_quota', funnel.left_quota,
    'leaking_ts', funnel.leaking_ts,
    'capacity', funnel.capacity,
    'operations', funnel.operations,
    'seconds', funnel.seconds
)
redis.call('SADD', 'funnel:keys', key)

return {ready, capacity, left_quota, interval, empty_time}`

type LeakyBucketState struct {
	Ready     bool          // return True if there has enough left quota, else False
	Capacity  int64         // funnel capacity
	LeftQuota int64         // funnel left quota after watering
	Interval  time.Duration // -1 if ret[0] is True, else waiting time until there have enough left quota to watering
	EmptyTime time.Duration // waiting time until the funnel is empty
}

type Funnel interface {
	Watering(ctx context.Context, quota int64) (LeakyBucketState, error)
}

type funnel struct {
	keys                 []string
	key                  string
	capacity, operations int64
	seconds              float64
	s                    Scripter
}

func newFunnel(c Cmdable, key string, capacity, operations int64, seconds time.Duration) funnel {
	sec := seconds.Seconds()
	if sec <= 0 {
		sec = 1
	}
	f := funnel{key: key, capacity: capacity, operations: operations, seconds: sec}
	f.keys = []string{f.key}
	f.s = c.CreateScript(luaFunnel)
	return f
}

func (f funnel) Watering(ctx context.Context, quota int64) (LeakyBucketState, error) {
	var state = LeakyBucketState{}
	res, err := f.s.Run(ctx, f.keys, f.capacity, f.operations, f.seconds, quota).Result()
	if err != nil {
		return state, err
	}
	state.Ready = res.([]interface{})[0].(int64) == 0
	state.Capacity = res.([]interface{})[1].(int64)
	state.LeftQuota = res.([]interface{})[2].(int64)
	interval, _ := strconv.ParseFloat(res.([]interface{})[3].(string), 64)
	emptyTime, _ := strconv.ParseFloat(res.([]interface{})[4].(string), 64)
	state.Interval = time.Duration(interval * float64(time.Second))
	state.EmptyTime = time.Duration(emptyTime * float64(time.Second))
	return state, err
}

func (r *resp3) NewFunnel(string, int64, int64, time.Duration) Funnel { return nil }
func (r *resp2) NewFunnel(string, int64, int64, time.Duration) Funnel { return nil }
func (c *client) NewFunnel(key string, capacity, operations int64, seconds time.Duration) Funnel {
	return newFunnel(c, key, capacity, operations, seconds)
}
