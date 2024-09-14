package redisson

import (
	"fmt"
	"github.com/redis/rueidis"
	"github.com/redis/rueidis/rueidiscompat"
	"strconv"
	"time"
)

const (
	OK      = "OK"
	EMPTY   = ""
	BYTE    = "BYTE"
	BIT     = "BIT"
	M       = "M"
	KM      = "KM"
	FT      = "FT"
	MI      = "MI"
	XX      = "XX"
	NX      = "NX"
	BEFORE  = "BEFORE"
	AFTER   = "AFTER"
	RIGHT   = "RIGHT"
	LEFT    = "LEFT"
	LADDR   = "LADDR"
	TYPE    = "TYPE"
	KeepTTL = rueidiscompat.KeepTTL
)

type (
	KeyValue                   = rueidiscompat.KeyValue
	CommandInfo                = rueidiscompat.CommandInfo
	SetArgs                    = rueidiscompat.SetArgs
	Sort                       = rueidiscompat.Sort
	RankScore                  = rueidiscompat.RankScore
	LPosArgs                   = rueidiscompat.LPosArgs
	Z                          = rueidiscompat.Z
	ZStore                     = rueidiscompat.ZStore
	ZAddArgs                   = rueidiscompat.ZAddArgs
	ZWithKey                   = rueidiscompat.ZWithKey
	ZRangeBy                   = rueidiscompat.ZRangeBy
	ZRangeArgs                 = rueidiscompat.ZRangeArgs
	XMessage                   = rueidiscompat.XMessage
	XInfoConsumer              = rueidiscompat.XInfoConsumer
	XInfoGroup                 = rueidiscompat.XInfoGroup
	XInfoStream                = rueidiscompat.XInfoStream
	XInfoStreamFull            = rueidiscompat.XInfoStreamFull
	XInfoStreamGroup           = rueidiscompat.XInfoStreamGroup
	XInfoStreamGroupPending    = rueidiscompat.XInfoStreamGroupPending
	XInfoStreamConsumer        = rueidiscompat.XInfoStreamConsumer
	XInfoStreamConsumerPending = rueidiscompat.XInfoStreamConsumerPending
	XPending                   = rueidiscompat.XPending
	XPendingExt                = rueidiscompat.XPendingExt
	XStream                    = rueidiscompat.XStream
	XAddArgs                   = rueidiscompat.XAddArgs
	XAutoClaimArgs             = rueidiscompat.XAutoClaimArgs
	XClaimArgs                 = rueidiscompat.XClaimArgs
	XPendingExtArgs            = rueidiscompat.XPendingExtArgs
	XReadArgs                  = rueidiscompat.XReadArgs
	XReadGroupArgs             = rueidiscompat.XReadGroupArgs
	BitCount                   = rueidiscompat.BitCount
	GeoPos                     = rueidiscompat.GeoPos
	GeoLocation                = rueidiscompat.GeoLocation
	GeoSearchQuery             = rueidiscompat.GeoSearchQuery
	GeoSearchLocationQuery     = rueidiscompat.GeoSearchLocationQuery
	GeoSearchStoreQuery        = rueidiscompat.GeoSearchStoreQuery
	GeoRadiusQuery             = rueidiscompat.GeoRadiusQuery
	ClusterNode                = rueidiscompat.ClusterNode
	ClusterSlot                = rueidiscompat.ClusterSlot
	ClusterShard               = rueidiscompat.ClusterShard
	Library                    = rueidiscompat.Library
	FunctionListQuery          = rueidiscompat.FunctionListQuery
	FilterBy                   = rueidiscompat.FilterBy
	KeyFlags                   = rueidiscompat.KeyFlags
	Message                    = rueidis.PubSubMessage
	Completed                  = rueidis.Completed
	Builder                    = rueidis.Builder
	RedisResult                = rueidis.RedisResult
	KeyValues                  = rueidis.KeyValues
)

type baseCmd[T any] struct {
	err error
	val T
}

func (c *baseCmd[T]) SetVal(val T) { c.val = val }
func (c *baseCmd[T]) Val() T {
	return c.val
}
func (c *baseCmd[T]) SetErr(err error) { c.err = err }
func (c *baseCmd[T]) Err() error {
	return c.err
}
func (c *baseCmd[T]) Result() (T, error) {
	return c.Val(), c.Err()
}

type BaseCmd interface {
	Err() error
}

type CompletedResult interface {
	BaseCmd

	from(rueidis.RedisResult)
}

type Cmd interface {
	BaseCmd

	Val() any
	Result() (any, error)
	Text() (string, error)
	Int() (int, error)
	Int64() (int64, error)
	Uint64() (uint64, error)
	Float32() (float32, error)
	Float64() (float64, error)
	Bool() (bool, error)
	Slice() ([]any, error)
	StringSlice() ([]string, error)
	Int64Slice() ([]int64, error)
	Uint64Slice() ([]uint64, error)
	Float32Slice() ([]float32, error)
	Float64Slice() ([]float64, error)
	BoolSlice() ([]bool, error)
}

type IntCmd interface {
	BaseCmd
	Val() int64
	Result() (int64, error)
	Uint64() (uint64, error)
}

type intCmd struct {
	baseCmd[int64]
}

func newIntCmd(res rueidis.RedisResult) IntCmd {
	cmd := &intCmd{}
	cmd.from(res)
	return cmd
}

func (c *intCmd) Uint64() (uint64, error) {
	return uint64(c.val), c.err
}
func (c *intCmd) from(res rueidis.RedisResult) {
	val, err := res.AsInt64()
	c.SetErr(err)
	c.SetVal(val)
}

type BoolCmd interface {
	BaseCmd
	Val() bool
	Result() (bool, error)
}

type boolCmd struct {
	baseCmd[bool]
}

func newBoolCmd(res rueidis.RedisResult) BoolCmd {
	cmd := &boolCmd{}
	cmd.from(res)
	return cmd
}

func (c *boolCmd) from(res rueidis.RedisResult) {
	val, err := res.AsBool()
	if rueidis.IsRedisNil(err) {
		val = false
		err = nil
	}
	c.SetVal(val)
	c.SetErr(err)
}

type StringCmd interface {
	BaseCmd
	Val() string
	Result() (string, error)
	Bytes() ([]byte, error)
	Bool() (bool, error)
	Int() (int, error)
	Int64() (int64, error)
	Uint64() (uint64, error)
	Float32() (float32, error)
	Float64() (float64, error)
	Time() (time.Time, error)
}

type stringCmd struct {
	baseCmd[string]
}

func (c *stringCmd) Bytes() ([]byte, error) {
	return []byte(c.val), c.err
}
func (c *stringCmd) Bool() (bool, error) {
	return c.val != "", c.err
}

func (c *stringCmd) Int() (int, error) {
	if c.err != nil {
		return 0, c.err
	}
	return strconv.Atoi(c.Val())
}

func (c *stringCmd) Int64() (int64, error) {
	if c.err != nil {
		return 0, c.err
	}
	return strconv.ParseInt(c.Val(), 10, 64)
}

func (c *stringCmd) Uint64() (uint64, error) {
	if c.err != nil {
		return 0, c.err
	}
	return strconv.ParseUint(c.Val(), 10, 64)
}

func (c *stringCmd) Float32() (float32, error) {
	if c.err != nil {
		return 0, c.err
	}
	v, err := toFloat32(c.Val())
	if err != nil {
		return 0, err
	}
	return v, nil
}

func (c *stringCmd) Float64() (float64, error) {
	if c.err != nil {
		return 0, c.err
	}
	return toFloat64(c.Val())
}

func (c *stringCmd) Time() (time.Time, error) {
	if c.err != nil {
		return time.Time{}, c.err
	}
	return time.Parse(time.RFC3339Nano, c.Val())
}

func (c *stringCmd) String() string {
	return c.val
}

func newStringCmd(res rueidis.RedisResult) StringCmd {
	cmd := &stringCmd{}
	cmd.from(res)
	return cmd
}

func (c *stringCmd) from(res rueidis.RedisResult) {
	val, err := res.ToString()
	c.SetErr(err)
	c.SetVal(val)
}

type TimeCmd interface {
	BaseCmd
	Val() time.Time
	Result() (time.Time, error)
}

type timeCmd struct {
	baseCmd[time.Time]
}

func newTimeCmd(res rueidis.RedisResult) *timeCmd {
	cmd := &timeCmd{}
	cmd.from(res)
	return cmd
}

func (c *timeCmd) from(res rueidis.RedisResult) {
	arr, err := res.ToArray()
	if err != nil {
		c.SetErr(err)
		return
	}
	if len(arr) < 2 {
		c.SetErr(fmt.Errorf("got %d, wanted 2", len(arr)))
		return
	}
	sec, err := arr[0].AsInt64()
	if err != nil {
		c.SetErr(err)
		return
	}
	microSec, err := arr[1].AsInt64()
	if err != nil {
		c.SetErr(err)
		return
	}
	c.SetVal(time.Unix(sec, microSec*1000))
}

type StatusCmd interface {
	BaseCmd
	Val() string
	Result() (string, error)
}

type statusCmd = stringCmd

func newStatusCmd(res rueidis.RedisResult) StatusCmd {
	cmd := &statusCmd{}
	cmd.from(res)
	return cmd
}

func newOKStatusCmdr() StatusCmd {
	cmd := &statusCmd{}
	cmd.SetVal(OK)
	return cmd
}

type FloatCmd interface {
	BaseCmd
	Val() float64
	Result() (float64, error)
}

type floatCmd struct {
	baseCmd[float64]
}

func newFloatCmd(res rueidis.RedisResult) FloatCmd {
	cmd := &floatCmd{}
	cmd.from(res)
	return cmd
}

func (c *floatCmd) from(res rueidis.RedisResult) {
	val, err := res.AsFloat64()
	c.SetErr(err)
	c.SetVal(val)
}

type DurationCmd interface {
	BaseCmd
	Val() time.Duration
	Result() (time.Duration, error)
}

type durationCmd struct {
	baseCmd[time.Duration]
	precision time.Duration
}

func newDurationCmd(res rueidis.RedisResult, precision time.Duration) DurationCmd {
	cmd := &durationCmd{precision: precision}
	cmd.from(res)
	return cmd
}

func (c *durationCmd) from(res rueidis.RedisResult) {
	val, err := res.AsInt64()
	c.SetErr(err)
	if val > 0 {
		c.SetVal(time.Duration(val) * c.precision)
	} else {
		c.SetVal(time.Duration(val))
	}
}

type KeyValuesCmd interface {
	Val() (string, []string)
	Err() error
	Result() (string, []string, error)
}

type keyValuesCmd struct {
	err error
	val rueidis.KeyValues
}

func newKeyValuesCmd(res rueidis.RedisResult) *keyValuesCmd {
	ret := &keyValuesCmd{}
	ret.from(res)
	return ret
}

func (c *keyValuesCmd) from(res rueidis.RedisResult) {
	c.val, c.err = res.AsLMPop()
}

func (c *keyValuesCmd) SetVal(key string, val []string) {
	c.val.Key = key
	c.val.Values = val
}

func (c *keyValuesCmd) SetErr(err error) { c.err = err }
func (c *keyValuesCmd) Val() (string, []string) {
	return c.val.Key, c.val.Values
}
func (c *keyValuesCmd) Err() error {
	return c.err
}
func (c *keyValuesCmd) Result() (string, []string, error) {
	return c.val.Key, c.val.Values, c.err
}

type FunctionListCmd interface {
	BaseCmd
	Val() []Library
	Err() error
	Result() ([]Library, error)
	First() (*Library, error)
}

type KeyFlagsCmd interface {
	BaseCmd
	Val() []KeyFlags
	Err() error
	Result() ([]KeyFlags, error)
}

type keyFlagsCmd struct {
	baseCmd[[]KeyFlags]
}

func newKeyFlagsCmd(res rueidis.RedisResult) *keyFlagsCmd {
	ret := &keyFlagsCmd{}
	ret.from(res)
	return ret
}

func (c *keyFlagsCmd) from(res rueidis.RedisResult) {
	if c.err = res.Error(); c.err == nil {
		kfs, _ := res.ToArray()
		c.val = make([]KeyFlags, len(kfs))
		for i := 0; i < len(kfs); i++ {
			if kf, _ := kfs[i].ToArray(); len(kf) >= 2 {
				c.val[i].Key, _ = kf[0].ToString()
				c.val[i].Flags, _ = kf[1].AsStrSlice()
			}
		}
	}
}

type ScanCmd interface {
	BaseCmd
	Val() (keys []string, cursor uint64)
	Result() (keys []string, cursor uint64, err error)
}

type scanCmd struct {
	err    error
	keys   []string
	cursor uint64
}

func newScanCmd(res rueidis.RedisResult) *scanCmd {
	c := &scanCmd{}
	c.from(res)
	return c
}

func (c *scanCmd) SetVal(keys []string, cursor uint64) {
	c.keys = keys
	c.cursor = cursor
}

func (c *scanCmd) Val() (keys []string, cursor uint64) {
	return c.keys, c.cursor
}
func (c *scanCmd) Err() error {
	return c.err
}
func (c *scanCmd) Result() (keys []string, cursor uint64, err error) {
	return c.keys, c.cursor, c.err
}
func (c *scanCmd) from(res rueidis.RedisResult) {
	r, err := res.AsScanEntry()
	c.cursor = r.Cursor
	c.keys = r.Elements
	c.err = err
}

type SliceCmd interface {
	BaseCmd
	Val() []any
	Result() ([]any, error)
	Scan(dst any) error
}

type sliceCmd struct {
	baseCmd[[]any]
	keys          []string
	isJSONObjKeys bool
}

// newSliceCmd returns SliceCmd according to input arguments, if the caller is JSONObjKeys,
// set isJSONObjKeys to true.
func newSliceCmd(res rueidis.RedisResult, isJSONObjKeys bool, keys ...string) SliceCmd {
	cmd := &sliceCmd{keys: keys, isJSONObjKeys: isJSONObjKeys}
	cmd.from(res)
	return cmd
}

func newSliceCmdFromSlice(res []any, err error, keys ...string) *sliceCmd {
	cmd := &sliceCmd{keys: keys}
	if err != nil {
		cmd.SetErr(err)
		return cmd
	}
	cmd.SetVal(res)
	return cmd
}

func (c *sliceCmd) from(res rueidis.RedisResult) {
	arr, err := res.ToArray()
	if err != nil {
		c.SetErr(err)
		return
	}
	vals := make([]any, len(arr))
	if c.isJSONObjKeys {
		for i, v := range arr {
			// for JSON.OBJKEYS
			if v.IsNil() {
				continue
			}
			// convert to any which underlying type is []any
			arr, err := v.ToAny()
			if err != nil {
				c.SetErr(err)
				return
			}
			vals[i] = arr
		}
		c.SetVal(vals)
		return
	}
	for i, v := range arr {
		// keep the old behavior the same as before (don't handle error while parsing v as string)
		if s, err := v.ToString(); err == nil {
			vals[i] = s
		}
	}
	c.SetVal(vals)
}

func (c *sliceCmd) Scan(dst any) error {
	if c.err != nil {
		return c.err
	}
	return rueidiscompat.Scan(dst, c.keys, c.val)
}

type IntSliceCmd interface {
	BaseCmd
	Val() []int64
	Result() ([]int64, error)
}

type intSliceCmd struct {
	baseCmd[[]int64]
}

func newIntSliceCmd(res rueidis.RedisResult) IntSliceCmd {
	cmd := &intSliceCmd{}
	cmd.from(res)
	return cmd
}

func (c *intSliceCmd) from(res rueidis.RedisResult) {
	val, err := res.AsIntSlice()
	c.SetErr(err)
	c.SetVal(val)
}

type FloatSliceCmd interface {
	BaseCmd
	Val() []float64
	Result() ([]float64, error)
}

type floatSliceCmd struct {
	baseCmd[[]float64]
}

func newFloatSliceCmd(res rueidis.RedisResult) FloatSliceCmd {
	cmd := &floatSliceCmd{}
	cmd.from(res)
	return cmd
}

func (c *floatSliceCmd) from(res rueidis.RedisResult) {
	val, err := res.AsFloatSlice()
	c.SetErr(err)
	c.SetVal(val)
}

type StringSliceCmd interface {
	BaseCmd
	Val() []string
	Result() ([]string, error)
}

type stringSliceCmd struct {
	baseCmd[[]string]
}

func newStringSliceCmd(res rueidis.RedisResult) StringSliceCmd {
	cmd := &stringSliceCmd{}
	cmd.from(res)
	return cmd
}

func (c *stringSliceCmd) from(res rueidis.RedisResult) {
	val, err := res.AsStrSlice()
	c.SetVal(val)
	c.SetErr(err)
}

type DurationSliceCmd interface {
	BaseCmd
	Val() []time.Duration
	Result() ([]time.Duration, error)
}

type durationSliceCmd struct {
	precision time.Duration
	baseCmd[[]time.Duration]
}

func newDurationSliceCmd(res rueidis.RedisResult, precision time.Duration) DurationSliceCmd {
	cmd := &durationSliceCmd{precision: precision}
	cmd.from(res)
	return cmd
}

func (c *durationSliceCmd) from(res rueidis.RedisResult) {
	ints, err := res.AsIntSlice()
	if err != nil {
		c.SetErr(err)
		return
	}
	val := make([]time.Duration, 0, len(ints))
	for _, i := range ints {
		if i > 0 {
			val = append(val, time.Duration(i)*c.precision)
		} else {
			val = append(val, time.Duration(i))
		}
	}
	c.SetVal(val)
}

type BoolSliceCmd interface {
	BaseCmd
	Val() []bool
	Result() ([]bool, error)
}

type boolSliceCmd struct {
	baseCmd[[]bool]
}

func newBoolSliceCmd(res rueidis.RedisResult) BoolSliceCmd {
	cmd := &boolSliceCmd{}
	cmd.from(res)
	return cmd
}

func (c *boolSliceCmd) from(res rueidis.RedisResult) {
	ints, err := res.AsIntSlice()
	if err != nil {
		c.SetErr(err)
		return
	}
	val := make([]bool, 0, len(ints))
	for _, i := range ints {
		val = append(val, i == 1)
	}
	c.SetVal(val)
}

type StringStringMapCmd interface {
	BaseCmd
	Val() map[string]string
	Result() (map[string]string, error)
	Scan(dest any) error
}

type stringStringMapCmd struct {
	baseCmd[map[string]string]
}

func newStringStringMapCmd(res rueidis.RedisResult) StringStringMapCmd {
	cmd := &stringStringMapCmd{}
	cmd.from(res)
	return cmd
}

func (c *stringStringMapCmd) from(res rueidis.RedisResult) {
	val, err := res.AsStrMap()
	c.SetErr(err)
	c.SetVal(val)
}

// Scan scans the results from the map into a destination struct. The map keys
// are matched in the Redis struct fields by the `redis:"field"` tag.
func (c *stringStringMapCmd) Scan(dest any) error {
	if c.Err() != nil {
		return c.Err()
	}

	strct, err := rueidiscompat.Struct(dest)
	if err != nil {
		return err
	}

	for k, v := range c.val {
		if err = strct.Scan(k, v); err != nil {
			return err
		}
	}

	return nil
}

type StringIntMapCmd interface {
	BaseCmd
	Val() map[string]int64
	Result() (map[string]int64, error)
}

type stringIntMapCmd struct {
	baseCmd[map[string]int64]
}

func newStringIntMapCmd(res rueidis.RedisResult) *stringIntMapCmd {
	cmd := &stringIntMapCmd{}
	cmd.from(res)
	return cmd
}

func (c *stringIntMapCmd) from(res rueidis.RedisResult) {
	val, err := res.AsIntMap()
	c.SetErr(err)
	c.SetVal(val)
}

type StringStructMapCmd interface {
	BaseCmd
	Val() map[string]struct{}
	Result() (map[string]struct{}, error)
}

type ClusterSlotsCmd interface {
	BaseCmd
	Val() []ClusterSlot
	Result() ([]ClusterSlot, error)
}

type ClusterShardsCmd interface {
	BaseCmd
	Val() []ClusterShard
	Result() ([]ClusterShard, error)
}

type GeoPosCmd interface {
	BaseCmd
	Val() []*GeoPos
	Result() ([]*GeoPos, error)
}

type geoPosCmd struct {
	baseCmd[[]*GeoPos]
}

func newGeoPosCmd(res rueidis.RedisResult) GeoPosCmd {
	cmd := &geoPosCmd{}
	cmd.from(res)
	return cmd
}

func (c *geoPosCmd) from(res rueidis.RedisResult) {
	arr, err := res.ToArray()
	if err != nil {
		c.SetErr(err)
		return
	}
	val := make([]*GeoPos, 0, len(arr))
	for _, v := range arr {
		loc, err := v.ToArray()
		if err != nil {
			if rueidis.IsRedisNil(err) {
				val = append(val, nil)
				continue
			}
			c.SetErr(err)
			return
		}
		if len(loc) != 2 {
			c.SetErr(fmt.Errorf("got %d, expected 2", len(loc)))
			return
		}
		long, err := loc[0].AsFloat64()
		if err != nil {
			c.SetErr(err)
			return
		}
		lat, err := loc[1].AsFloat64()
		if err != nil {
			c.SetErr(err)
			return
		}
		val = append(val, &GeoPos{
			Longitude: long,
			Latitude:  lat,
		})
	}
	c.SetVal(val)
}

type GeoLocationCmd interface {
	BaseCmd
	Val() []GeoLocation
	Result() ([]GeoLocation, error)
}

type geoLocationCmd struct {
	baseCmd[[]rueidis.GeoLocation]
}

func newGeoLocationCmd(res rueidis.RedisResult) GeoLocationCmd {
	cmd := &geoLocationCmd{}
	cmd.from(res)
	return cmd
}

func (c *geoLocationCmd) from(res rueidis.RedisResult) {
	c.val, c.err = res.AsGeosearch()
}

type KeyValueSliceCmd interface {
	BaseCmd
	Val() []KeyValue
	Result() ([]KeyValue, error)
}

type keyValueSliceCmd struct {
	baseCmd[[]KeyValue]
}

func newKeyValueSliceCmd(res rueidis.RedisResult) *keyValueSliceCmd {
	cmd := &keyValueSliceCmd{}
	cmd.from(res)
	return cmd
}

func (c *keyValueSliceCmd) from(res rueidis.RedisResult) {
	arr, err := res.ToArray()
	for _, a := range arr {
		kv, _ := a.AsStrSlice()
		for i := 0; i < len(kv); i += 2 {
			c.val = append(c.val, KeyValue{Key: kv[i], Value: kv[i+1]})
		}
	}
	c.SetErr(err)
}

type CommandsInfoCmd interface {
	BaseCmd
	Val() map[string]CommandInfo
	Result() (map[string]CommandInfo, error)
}

type commandsInfoCmd struct {
	baseCmd[map[string]CommandInfo]
}

func newCommandsInfoCmd(res rueidis.RedisResult) *commandsInfoCmd {
	cmd := &commandsInfoCmd{}
	cmd.from(res)
	return cmd
}

func (c *commandsInfoCmd) from(res rueidis.RedisResult) {
	arr, err := res.ToArray()
	if err != nil {
		c.SetErr(err)
		return
	}
	val := make(map[string]CommandInfo, len(arr))
	for _, v := range arr {
		info, err := v.ToArray()
		if err != nil {
			c.SetErr(err)
			return
		}
		if len(info) < 6 {
			c.SetErr(fmt.Errorf("got %d, wanted at least 6", len(info)))
			return
		}
		var _cmd CommandInfo
		_cmd.Name, err = info[0].ToString()
		if err != nil {
			c.SetErr(err)
			return
		}
		_cmd.Arity, err = info[1].AsInt64()
		if err != nil {
			c.SetErr(err)
			return
		}
		_cmd.Flags, err = info[2].AsStrSlice()
		if err != nil {
			if rueidis.IsRedisNil(err) {
				_cmd.Flags = []string{}
			} else {
				c.SetErr(err)
				return
			}
		}
		_cmd.FirstKeyPos, err = info[3].AsInt64()
		if err != nil {
			c.SetErr(err)
			return
		}
		_cmd.LastKeyPos, err = info[4].AsInt64()
		if err != nil {
			c.SetErr(err)
			return
		}
		_cmd.StepCount, err = info[5].AsInt64()
		if err != nil {
			c.SetErr(err)
			return
		}
		for _, flag := range _cmd.Flags {
			if flag == "readonly" {
				_cmd.ReadOnly = true
				break
			}
		}
		if len(info) == 6 {
			val[_cmd.Name] = _cmd
			continue
		}
		_cmd.ACLFlags, err = info[6].AsStrSlice()
		if err != nil {
			if rueidis.IsRedisNil(err) {
				_cmd.ACLFlags = []string{}
			} else {
				c.SetErr(err)
				return
			}
		}
		val[_cmd.Name] = _cmd
	}
	c.SetVal(val)
}

type ZSliceWithKeyCmd interface {
	BaseCmd
	Val() (string, []Z)
	Result() (string, []Z, error)
}

type zSliceWithKeyCmd struct {
	err error
	key string
	val []Z
}

func newZSliceWithKeyCmd(res rueidis.RedisResult) *zSliceWithKeyCmd {
	c := &zSliceWithKeyCmd{}
	c.from(res)
	return c
}

func (c *zSliceWithKeyCmd) from(res rueidis.RedisResult) {
	v, err := res.AsZMPop()
	if err != nil {
		c.err = err
		return
	}
	val := make([]Z, 0, len(v.Values))
	for _, s := range v.Values {
		val = append(val, Z{Member: s.Member, Score: s.Score})
	}
	c.val = val
	c.key = v.Key
}

func (c *zSliceWithKeyCmd) SetVal(key string, val []Z) {
	c.key = key
	c.val = val
}

func (c *zSliceWithKeyCmd) SetErr(err error) { c.err = err }
func (c *zSliceWithKeyCmd) Val() (string, []Z) {
	return c.key, c.val
}
func (c *zSliceWithKeyCmd) Err() error {
	return c.err
}
func (c *zSliceWithKeyCmd) Result() (string, []Z, error) {
	return c.key, c.val, c.err
}

type ZWithKeyCmd interface {
	BaseCmd
	Val() ZWithKey
	Result() (ZWithKey, error)
}

type ZSliceCmd interface {
	BaseCmd
	Val() []Z
	Result() ([]Z, error)
}

type zSliceCmd struct {
	baseCmd[[]Z]
}

func newZSliceCmd(res rueidis.RedisResult) ZSliceCmd {
	cmd := &zSliceCmd{}
	cmd.from(res)
	return cmd
}

func (c *zSliceCmd) from(res rueidis.RedisResult) {
	scores, err := res.AsZScores()
	if err != nil {
		c.SetErr(err)
		return
	}
	val := make([]Z, 0, len(scores))
	for _, s := range scores {
		val = append(val, Z{Member: s.Member, Score: s.Score})
	}
	c.SetVal(val)
}

type RankWithScoreCmd interface {
	BaseCmd
	Val() RankScore
	Result() (RankScore, error)
}

type rankWithScoreCmd struct {
	baseCmd[RankScore]
}

func newRankWithScoreCmd(res rueidis.RedisResult) RankWithScoreCmd {
	ret := &rankWithScoreCmd{}
	ret.from(res)
	return ret
}

func (c *rankWithScoreCmd) from(res rueidis.RedisResult) {
	if c.err = res.Error(); c.err == nil {
		vs, _ := res.ToArray()
		if len(vs) >= 2 {
			c.val.Rank, _ = vs[0].AsInt64()
			c.val.Score, _ = vs[1].AsFloat64()
		}
	}
}

type XMessageSliceCmd interface {
	BaseCmd
	Val() []XMessage
	Result() ([]XMessage, error)
}

type xMessageSliceCmd struct {
	baseCmd[[]XMessage]
}

func newXMessageSliceCmd(res rueidis.RedisResult) *xMessageSliceCmd {
	cmd := &xMessageSliceCmd{}
	cmd.from(res)
	return cmd
}

func (c *xMessageSliceCmd) from(res rueidis.RedisResult) {
	val, err := res.AsXRange()
	c.SetErr(err)
	c.val = make([]XMessage, len(val))
	for i, r := range val {
		c.val[i] = newXMessage(r)
	}
}

type XAutoClaimCmd interface {
	BaseCmd
	Val() (messages []XMessage, start string)
	Result() (messages []XMessage, start string, err error)
}

type xAutoClaimCmd struct {
	err   error
	start string
	val   []XMessage
}

func newXMessage(r rueidis.XRangeEntry) XMessage {
	if r.FieldValues == nil {
		return XMessage{ID: r.ID, Values: nil}
	}
	m := XMessage{ID: r.ID, Values: make(map[string]any, len(r.FieldValues))}
	for k, v := range r.FieldValues {
		m.Values[k] = v
	}
	return m
}

func newXAutoClaimCmd(res rueidis.RedisResult) *xAutoClaimCmd {
	c := &xAutoClaimCmd{}
	c.from(res)
	return c
}

func (c *xAutoClaimCmd) from(res rueidis.RedisResult) {
	arr, err := res.ToArray()
	if err != nil {
		c.err = err
		return
	}
	if len(arr) < 2 {
		c.err = fmt.Errorf("got %d, wanted 2", len(arr))
		return
	}
	start, err := arr[0].ToString()
	if err != nil {
		c.err = err
		return
	}
	ranges, err := arr[1].AsXRange()
	if err != nil {
		c.err = err
		return
	}
	val := make([]XMessage, 0, len(ranges))
	for _, r := range ranges {
		val = append(val, newXMessage(r))
	}
	c.val = val
	c.start = start
	c.err = err
}

func (c *xAutoClaimCmd) SetVal(val []XMessage, start string) {
	c.val = val
	c.start = start
}

func (c *xAutoClaimCmd) SetErr(err error) { c.err = err }
func (c *xAutoClaimCmd) Val() (messages []XMessage, start string) {
	return c.val, c.start
}
func (c *xAutoClaimCmd) Err() error {
	return c.err
}
func (c *xAutoClaimCmd) Result() (messages []XMessage, start string, err error) {
	return c.val, c.start, c.err
}

type XInfoConsumersCmd interface {
	BaseCmd
	Val() []XInfoConsumer
	Result() ([]XInfoConsumer, error)
}

type xInfoConsumersCmd struct {
	baseCmd[[]XInfoConsumer]
}

func newXInfoConsumersCmd(res rueidis.RedisResult) *xInfoConsumersCmd {
	cmd := &xInfoConsumersCmd{}
	cmd.from(res)
	return cmd
}

func (c *xInfoConsumersCmd) from(res rueidis.RedisResult) {
	arr, err := res.ToArray()
	if err != nil {
		c.SetErr(err)
		return
	}
	val := make([]XInfoConsumer, 0, len(arr))
	for _, v := range arr {
		info, err := v.AsMap()
		if err != nil {
			c.SetErr(err)
			return
		}
		var consumer XInfoConsumer
		if attr, ok := info["name"]; ok {
			consumer.Name, _ = attr.ToString()
		}
		if attr, ok := info["pending"]; ok {
			consumer.Pending, _ = attr.AsInt64()
		}
		if attr, ok := info["idle"]; ok {
			idle, _ := attr.AsInt64()
			consumer.Idle = time.Duration(idle) * time.Millisecond
		}
		val = append(val, consumer)
	}
	c.SetVal(val)
}

type XInfoGroupsCmd interface {
	BaseCmd
	Val() []XInfoGroup
	Result() ([]XInfoGroup, error)
}

type xInfoGroupsCmd struct {
	baseCmd[[]XInfoGroup]
}

func newXInfoGroupsCmd(res rueidis.RedisResult) *xInfoGroupsCmd {
	cmd := &xInfoGroupsCmd{}
	cmd.from(res)
	return cmd
}

func (c *xInfoGroupsCmd) from(res rueidis.RedisResult) {
	arr, err := res.ToArray()
	if err != nil {
		c.SetErr(err)
		return
	}
	groupInfos := make([]XInfoGroup, 0, len(arr))
	for _, v := range arr {
		info, err := v.AsMap()
		if err != nil {
			c.SetErr(err)
			return
		}
		var group XInfoGroup
		if attr, ok := info["name"]; ok {
			group.Name, _ = attr.ToString()
		}
		if attr, ok := info["consumers"]; ok {
			group.Consumers, _ = attr.AsInt64()
		}
		if attr, ok := info["pending"]; ok {
			group.Pending, _ = attr.AsInt64()
		}
		if attr, ok := info["entries-read"]; ok {
			group.EntriesRead, _ = attr.AsInt64()
		}
		if attr, ok := info["lag"]; ok {
			group.Lag, _ = attr.AsInt64()
		}
		if attr, ok := info["last-delivered-id"]; ok {
			group.LastDeliveredID, _ = attr.ToString()
		}
		groupInfos = append(groupInfos, group)
	}
	c.SetVal(groupInfos)
}

type XInfoStreamCmd interface {
	BaseCmd
	Val() XInfoStream
	Result() (XInfoStream, error)
}

type xInfoStreamCmd struct {
	baseCmd[XInfoStream]
}

func newXInfoStreamCmd(res rueidis.RedisResult) *xInfoStreamCmd {
	cmd := &xInfoStreamCmd{}
	cmd.from(res)
	return cmd
}

func (c *xInfoStreamCmd) from(res rueidis.RedisResult) {
	kv, err := res.AsMap()
	if err != nil {
		c.SetErr(err)
		return
	}
	var val XInfoStream
	if v, ok := kv["length"]; ok {
		val.Length, _ = v.AsInt64()
	}
	if v, ok := kv["radix-tree-keys"]; ok {
		val.RadixTreeKeys, _ = v.AsInt64()
	}
	if v, ok := kv["radix-tree-nodes"]; ok {
		val.RadixTreeNodes, _ = v.AsInt64()
	}
	if v, ok := kv["groups"]; ok {
		val.Groups, _ = v.AsInt64()
	}
	if v, ok := kv["last-generated-id"]; ok {
		val.LastGeneratedID, _ = v.ToString()
	}
	if v, ok := kv["max-deleted-entry-id"]; ok {
		val.MaxDeletedEntryID, _ = v.ToString()
	}
	if v, ok := kv["recorded-first-entry-id"]; ok {
		val.RecordedFirstEntryID, _ = v.ToString()
	}
	if v, ok := kv["entries-added"]; ok {
		val.EntriesAdded, _ = v.AsInt64()
	}
	if v, ok := kv["first-entry"]; ok {
		if r, err := v.AsXRangeEntry(); err == nil {
			val.FirstEntry = newXMessage(r)
		}
	}
	if v, ok := kv["last-entry"]; ok {
		if r, err := v.AsXRangeEntry(); err == nil {
			val.LastEntry = newXMessage(r)
		}
	}
	c.SetVal(val)
}

type XInfoStreamFullCmd interface {
	BaseCmd
	Val() XInfoStreamFull
	Result() (XInfoStreamFull, error)
}

type xInfoStreamFullCmd struct {
	baseCmd[XInfoStreamFull]
}

func newXInfoStreamFullCmd(res rueidis.RedisResult) *xInfoStreamFullCmd {
	cmd := &xInfoStreamFullCmd{}
	cmd.from(res)
	return cmd
}

func (c *xInfoStreamFullCmd) from(res rueidis.RedisResult) {
	kv, err := res.AsMap()
	if err != nil {
		c.SetErr(err)
		return
	}
	var val XInfoStreamFull
	if v, ok := kv["length"]; ok {
		val.Length, _ = v.AsInt64()
	}
	if v, ok := kv["radix-tree-keys"]; ok {
		val.RadixTreeKeys, _ = v.AsInt64()
	}
	if v, ok := kv["radix-tree-nodes"]; ok {
		val.RadixTreeNodes, _ = v.AsInt64()
	}
	if v, ok := kv["last-generated-id"]; ok {
		val.LastGeneratedID, _ = v.ToString()
	}
	if v, ok := kv["entries-added"]; ok {
		val.EntriesAdded, _ = v.AsInt64()
	}
	if v, ok := kv["max-deleted-entry-id"]; ok {
		val.MaxDeletedEntryID, _ = v.ToString()
	}
	if v, ok := kv["recorded-first-entry-id"]; ok {
		val.RecordedFirstEntryID, _ = v.ToString()
	}
	if v, ok := kv["groups"]; ok {
		val.Groups, err = readStreamGroups(v)
		if err != nil {
			c.SetErr(err)
			return
		}
	}
	if v, ok := kv["entries"]; ok {
		ranges, err := v.AsXRange()
		if err != nil {
			c.SetErr(err)
			return
		}
		val.Entries = make([]XMessage, 0, len(ranges))
		for _, r := range ranges {
			val.Entries = append(val.Entries, newXMessage(r))
		}
	}
	c.SetVal(val)
}

func readStreamGroups(res rueidis.RedisMessage) ([]XInfoStreamGroup, error) {
	arr, err := res.ToArray()
	if err != nil {
		return nil, err
	}
	groups := make([]XInfoStreamGroup, 0, len(arr))
	for _, v := range arr {
		info, err := v.AsMap()
		if err != nil {
			return nil, err
		}
		var group XInfoStreamGroup
		if attr, ok := info["name"]; ok {
			group.Name, _ = attr.ToString()
		}
		if attr, ok := info["last-delivered-id"]; ok {
			group.LastDeliveredID, _ = attr.ToString()
		}
		if attr, ok := info["entries-read"]; ok {
			group.EntriesRead, _ = attr.AsInt64()
		}
		if attr, ok := info["lag"]; ok {
			group.Lag, _ = attr.AsInt64()
		}
		if attr, ok := info["pel-count"]; ok {
			group.PelCount, _ = attr.AsInt64()
		}
		if attr, ok := info["pending"]; ok {
			group.Pending, err = readXInfoStreamGroupPending(attr)
			if err != nil {
				return nil, err
			}
		}
		if attr, ok := info["consumers"]; ok {
			group.Consumers, err = readXInfoStreamConsumers(attr)
			if err != nil {
				return nil, err
			}
		}
		groups = append(groups, group)
	}
	return groups, nil
}

func readXInfoStreamGroupPending(res rueidis.RedisMessage) ([]XInfoStreamGroupPending, error) {
	arr, err := res.ToArray()
	if err != nil {
		return nil, err
	}
	pending := make([]XInfoStreamGroupPending, 0, len(arr))
	for _, v := range arr {
		info, err := v.ToArray()
		if err != nil {
			return nil, err
		}
		if len(info) < 4 {
			return nil, fmt.Errorf("got %d, wanted 4", len(arr))
		}
		var p XInfoStreamGroupPending
		p.ID, err = info[0].ToString()
		if err != nil {
			return nil, err
		}
		p.Consumer, err = info[1].ToString()
		if err != nil {
			return nil, err
		}
		delivery, err := info[2].AsInt64()
		if err != nil {
			return nil, err
		}
		p.DeliveryTime = time.Unix(delivery/1000, delivery%1000*int64(time.Millisecond))
		p.DeliveryCount, err = info[3].AsInt64()
		if err != nil {
			return nil, err
		}
		pending = append(pending, p)
	}
	return pending, nil
}

func readXInfoStreamConsumers(res rueidis.RedisMessage) ([]XInfoStreamConsumer, error) {
	arr, err := res.ToArray()
	if err != nil {
		return nil, err
	}
	consumer := make([]XInfoStreamConsumer, 0, len(arr))
	for _, v := range arr {
		info, err := v.AsMap()
		if err != nil {
			return nil, err
		}
		var c XInfoStreamConsumer
		if attr, ok := info["name"]; ok {
			c.Name, _ = attr.ToString()
		}
		if attr, ok := info["seen-time"]; ok {
			seen, _ := attr.AsInt64()
			c.SeenTime = time.Unix(seen/1000, seen%1000*int64(time.Millisecond))
		}
		if attr, ok := info["pel-count"]; ok {
			c.PelCount, _ = attr.AsInt64()
		}
		if attr, ok := info["pending"]; ok {
			pending, err := attr.ToArray()
			if err != nil {
				return nil, err
			}
			c.Pending = make([]XInfoStreamConsumerPending, 0, len(pending))
			for _, v := range pending {
				pendingInfo, err := v.ToArray()
				if err != nil {
					return nil, err
				}
				if len(pendingInfo) < 3 {
					return nil, fmt.Errorf("got %d, wanted 3", len(arr))
				}
				var p XInfoStreamConsumerPending
				p.ID, err = pendingInfo[0].ToString()
				if err != nil {
					return nil, err
				}
				delivery, err := pendingInfo[1].AsInt64()
				if err != nil {
					return nil, err
				}
				p.DeliveryTime = time.Unix(delivery/1000, delivery%1000*int64(time.Millisecond))
				p.DeliveryCount, err = pendingInfo[2].AsInt64()
				if err != nil {
					return nil, err
				}
				c.Pending = append(c.Pending, p)
			}
		}
		consumer = append(consumer, c)
	}
	return consumer, nil
}

type XPendingCmd interface {
	BaseCmd
	Val() XPending
	Result() (XPending, error)
}

type xPendingCmd struct {
	baseCmd[XPending]
}

func newXPendingCmd(res rueidis.RedisResult) *xPendingCmd {
	cmd := &xPendingCmd{}
	cmd.from(res)
	return cmd
}

func (c *xPendingCmd) from(res rueidis.RedisResult) {
	arr, err := res.ToArray()
	if err != nil {
		c.SetErr(err)
		return
	}
	if len(arr) < 4 {
		c.SetErr(fmt.Errorf("got %d, wanted 4", len(arr)))
		return
	}
	count, err := arr[0].AsInt64()
	if err != nil {
		c.SetErr(err)
		return
	}
	lower, err := arr[1].ToString()
	if err != nil {
		c.SetErr(err)
		return
	}
	higher, err := arr[2].ToString()
	if err != nil {
		c.SetErr(err)
		return
	}
	val := XPending{
		Count:  count,
		Lower:  lower,
		Higher: higher,
	}
	consumerArr, err := arr[3].ToArray()
	if err != nil {
		c.SetErr(err)
		return
	}
	for _, v := range consumerArr {
		consumer, err := v.ToArray()
		if err != nil {
			c.SetErr(err)
			return
		}
		if len(consumer) < 2 {
			c.SetErr(fmt.Errorf("got %d, wanted 2", len(arr)))
			return
		}
		consumerName, err := consumer[0].ToString()
		if err != nil {
			c.SetErr(err)
			return
		}
		consumerPending, err := consumer[1].AsInt64()
		if err != nil {
			c.SetErr(err)
			return
		}
		if val.Consumers == nil {
			val.Consumers = make(map[string]int64)
		}
		val.Consumers[consumerName] = consumerPending
	}
	c.SetVal(val)
}

type XPendingExtCmd interface {
	BaseCmd
	Val() []XPendingExt
	Result() ([]XPendingExt, error)
}

type xPendingExtCmd struct {
	baseCmd[[]XPendingExt]
}

func newXPendingExtCmd(res rueidis.RedisResult) *xPendingExtCmd {
	cmd := &xPendingExtCmd{}
	cmd.from(res)
	return cmd
}

func (c *xPendingExtCmd) from(res rueidis.RedisResult) {
	arrs, err := res.ToArray()
	if err != nil {
		c.SetErr(err)
		return
	}
	val := make([]XPendingExt, 0, len(arrs))
	for _, v := range arrs {
		arr, err := v.ToArray()
		if err != nil {
			c.SetErr(err)
			return
		}
		if len(arr) < 4 {
			c.SetErr(fmt.Errorf("got %d, wanted 4", len(arr)))
			return
		}
		id, err := arr[0].ToString()
		if err != nil {
			c.SetErr(err)
			return
		}
		consumer, err := arr[1].ToString()
		if err != nil {
			c.SetErr(err)
			return
		}
		idle, err := arr[2].AsInt64()
		if err != nil {
			c.SetErr(err)
			return
		}
		retryCount, err := arr[3].AsInt64()
		if err != nil {
			c.SetErr(err)
			return
		}
		val = append(val, XPendingExt{
			ID:         id,
			Consumer:   consumer,
			Idle:       time.Duration(idle) * time.Millisecond,
			RetryCount: retryCount,
		})
	}
	c.SetVal(val)
}

type XStreamSliceCmd interface {
	BaseCmd
	Val() []XStream
	Result() ([]XStream, error)
}

type XAutoClaimJustIDCmd interface {
	BaseCmd
	Val() (ids []string, start string)
	Result() (ids []string, start string, err error)
}

type xAutoClaimJustIDCmd struct {
	err   error
	start string
	val   []string
}

func newXAutoClaimJustIDCmd(res rueidis.RedisResult) *xAutoClaimJustIDCmd {
	c := &xAutoClaimJustIDCmd{}
	c.from(res)
	return c
}

func (c *xAutoClaimJustIDCmd) from(res rueidis.RedisResult) {
	arr, err := res.ToArray()
	if err != nil {
		c.err = err
		return
	}
	if len(arr) < 2 {
		c.err = fmt.Errorf("got %d, wanted 2", len(arr))
		return
	}
	start, err := arr[0].ToString()
	if err != nil {
		c.err = err
		return
	}
	val, err := arr[1].AsStrSlice()
	if err != nil {
		c.err = err
		return
	}
	c.err = err
	c.val = val
	c.start = start
}

func (c *xAutoClaimJustIDCmd) SetVal(val []string, start string) {
	c.val = val
	c.start = start
}

func (c *xAutoClaimJustIDCmd) SetErr(err error) { c.err = err }
func (c *xAutoClaimJustIDCmd) Val() (ids []string, start string) {
	return c.val, c.start
}
func (c *xAutoClaimJustIDCmd) Err() error {
	return c.err
}
func (c *xAutoClaimJustIDCmd) Result() (ids []string, start string, err error) {
	return c.val, c.start, c.err
}

func geoRadiusQueryArgs(q GeoRadiusQuery) []string {
	args := make([]string, 0, 2)
	args = append(args, strconv.FormatFloat(q.Radius, 'f', -1, 64))
	if q.Unit != "" {
		args = append(args, q.Unit)
	} else {
		args = append(args, KM)
	}
	if q.WithCoord {
		args = append(args, XXX_WITHCOORD)
	}
	if q.WithDist {
		args = append(args, XXX_WITHDIST)
	}
	if q.WithGeoHash {
		args = append(args, XXX_WITHHASH)
	}
	if q.Count > 0 {
		args = append(args, XXX_COUNT, strconv.FormatInt(q.Count, 10))
	}
	if q.Sort != "" {
		args = append(args, q.Sort)
	}
	if q.Store != "" {
		args = append(args, XXX_STORE)
		args = append(args, q.Store)
	}
	if q.StoreDist != "" {
		args = append(args, XXX_STOREDIST)
		args = append(args, q.StoreDist)
	}
	return args
}

func geoSearchLocationQueryArgs(q GeoSearchLocationQuery) []string {
	args := geoSearchQueryArgs(q.GeoSearchQuery)
	if q.WithCoord {
		args = append(args, XXX_WITHCOORD)
	}
	if q.WithDist {
		args = append(args, XXX_WITHDIST)
	}
	if q.WithHash {
		args = append(args, XXX_WITHHASH)
	}
	return args
}

func geoSearchQueryArgs(q GeoSearchQuery) []string {
	args := make([]string, 0, 2)
	if q.Member != "" {
		args = append(args, XXX_FROMMEMBER, q.Member)
	} else {
		args = append(args, XXX_FROMLONLAT, strconv.FormatFloat(q.Longitude, 'f', -1, 64), strconv.FormatFloat(q.Latitude, 'f', -1, 64))
	}
	if q.Radius > 0 {
		if q.RadiusUnit == "" {
			q.RadiusUnit = KM
		}
		args = append(args, XXX_BYRADIUS, strconv.FormatFloat(q.Radius, 'f', -1, 64), q.RadiusUnit)
	} else {
		if q.BoxUnit == "" {
			q.BoxUnit = KM
		}
		args = append(args, XXX_BYBOX, strconv.FormatFloat(q.BoxWidth, 'f', -1, 64), strconv.FormatFloat(q.BoxHeight, 'f', -1, 64), q.BoxUnit)
	}
	if q.Sort != "" {
		args = append(args, q.Sort)
	}
	if q.Count > 0 {
		args = append(args, XXX_COUNT, strconv.FormatInt(q.Count, 10))
		if q.CountAny {
			args = append(args, XXX_ANY)
		}
	}
	return args
}
