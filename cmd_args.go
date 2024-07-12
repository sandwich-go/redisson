package redisson

import (
	"fmt"
	"github.com/redis/rueidis"
	"github.com/redis/rueidis/rueidiscompat"
	"strconv"
	"time"
)

const KeepTTL = rueidiscompat.KeepTTL

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
	Message                    = rueidis.PubSubMessage
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

type Cmd interface {
	BaseCmd

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
	val, err := res.AsInt64()
	cmd.SetErr(err)
	cmd.SetVal(val)
	return cmd
}

func (c *intCmd) Uint64() (uint64, error) {
	return uint64(c.val), c.err
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
	val, err := res.AsBool()
	if rueidis.IsRedisNil(err) {
		val = false
		err = nil
	}
	cmd.SetVal(val)
	cmd.SetErr(err)
	return cmd
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
	val, err := res.ToString()
	cmd.SetErr(err)
	cmd.SetVal(val)
	return cmd
}

type TimeCmd interface {
	BaseCmd
	Val() time.Time
	Result() (time.Time, error)
}

type StatusCmd interface {
	BaseCmd
	Val() string
	Result() (string, error)
}

type statusCmd = stringCmd

func newStatusCmd(res rueidis.RedisResult) StatusCmd {
	cmd := &statusCmd{}
	val, err := res.ToString()
	cmd.SetErr(err)
	cmd.SetVal(val)
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
	val, err := res.AsFloat64()
	cmd.SetErr(err)
	cmd.SetVal(val)
	return cmd
}

type DurationCmd interface {
	BaseCmd
	Val() time.Duration
	Result() (time.Duration, error)
}

type durationCmd struct {
	baseCmd[time.Duration]
}

func newDurationCmd(res rueidis.RedisResult, precision time.Duration) DurationCmd {
	cmd := &durationCmd{}
	val, err := res.AsInt64()
	cmd.SetErr(err)
	if val > 0 {
		cmd.SetVal(time.Duration(val) * precision)
		return cmd
	}
	cmd.SetVal(time.Duration(val))
	return cmd
}

type ScanCmd interface {
	BaseCmd
	Val() (keys []string, cursor uint64)
	Result() (keys []string, cursor uint64, err error)
}

type SliceCmd interface {
	BaseCmd
	Val() []any
	Result() ([]any, error)
	Scan(dst any) error
}

type sliceCmd struct {
	baseCmd[[]any]
	keys []string
}

// newSliceCmd returns SliceCmd according to input arguments, if the caller is JSONObjKeys,
// set isJSONObjKeys to true.
func newSliceCmd(res rueidis.RedisResult, isJSONObjKeys bool, keys ...string) SliceCmd {
	cmd := &sliceCmd{keys: keys}
	arr, err := res.ToArray()
	if err != nil {
		cmd.SetErr(err)
		return cmd
	}
	vals := make([]any, len(arr))
	if isJSONObjKeys {
		for i, v := range arr {
			// for JSON.OBJKEYS
			if v.IsNil() {
				continue
			}
			// convert to any which underlying type is []any
			arr, err := v.ToAny()
			if err != nil {
				cmd.SetErr(err)
				return cmd
			}
			vals[i] = arr
		}
		cmd.SetVal(vals)
		return cmd
	}
	for i, v := range arr {
		// keep the old behavior the same as before (don't handle error while parsing v as string)
		if s, err := v.ToString(); err == nil {
			vals[i] = s
		}
	}
	cmd.SetVal(vals)
	return cmd
}

func newSliceCmdFromSlice(res []interface{}, err error, keys ...string) *sliceCmd {
	cmd := &sliceCmd{keys: keys}
	if err != nil {
		cmd.SetErr(err)
		return cmd
	}
	cmd.SetVal(res)
	return cmd
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
	val, err := res.AsFloatSlice()
	cmd.SetErr(err)
	cmd.SetVal(val)
	return cmd
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
	val, err := res.AsStrSlice()
	cmd.SetVal(val)
	cmd.SetErr(err)
	return cmd
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
	ints, err := res.AsIntSlice()
	if err != nil {
		cmd.SetErr(err)
		return cmd
	}
	val := make([]bool, 0, len(ints))
	for _, i := range ints {
		val = append(val, i == 1)
	}
	cmd.SetVal(val)
	return cmd
}

type StringStringMapCmd interface {
	BaseCmd
	Val() map[string]string
	Result() (map[string]string, error)
	Scan(dest interface{}) error
}

type stringStringMapCmd struct {
	baseCmd[map[string]string]
}

func newStringStringMapCmd(res rueidis.RedisResult) StringStringMapCmd {
	cmd := &stringStringMapCmd{}
	val, err := res.AsStrMap()
	cmd.SetErr(err)
	cmd.SetVal(val)
	return cmd
}

// Scan scans the results from the map into a destination struct. The map keys
// are matched in the Redis struct fields by the `redis:"field"` tag.
func (c *stringStringMapCmd) Scan(dest interface{}) error {
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
	arr, err := res.ToArray()
	if err != nil {
		cmd.SetErr(err)
		return cmd
	}
	val := make([]*GeoPos, 0, len(arr))
	for _, v := range arr {
		loc, err := v.ToArray()
		if err != nil {
			if rueidis.IsRedisNil(err) {
				val = append(val, nil)
				continue
			}
			cmd.SetErr(err)
			return cmd
		}
		if len(loc) != 2 {
			cmd.SetErr(fmt.Errorf("got %d, expected 2", len(loc)))
			return cmd
		}
		long, err := loc[0].AsFloat64()
		if err != nil {
			cmd.SetErr(err)
			return cmd
		}
		lat, err := loc[1].AsFloat64()
		if err != nil {
			cmd.SetErr(err)
			return cmd
		}
		val = append(val, &GeoPos{
			Longitude: long,
			Latitude:  lat,
		})
	}
	cmd.SetVal(val)
	return cmd
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
	ret := &geoLocationCmd{}
	ret.val, ret.err = res.AsGeosearch()
	return ret
}

type GeoSearchLocationCmd interface {
	BaseCmd
	Val() []GeoLocation
	Result() ([]GeoLocation, error)
}

type KeyValueSliceCmd interface {
	BaseCmd
	Val() []KeyValue
	Result() ([]KeyValue, error)
}

type CommandsInfoCmd interface {
	BaseCmd
	Val() map[string]CommandInfo
	Result() (map[string]CommandInfo, error)
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
	scores, err := res.AsZScores()
	if err != nil {
		cmd.SetErr(err)
		return cmd
	}
	val := make([]Z, 0, len(scores))
	for _, s := range scores {
		val = append(val, Z{Member: s.Member, Score: s.Score})
	}
	cmd.SetVal(val)
	return cmd
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
	if ret.err = res.Error(); ret.err == nil {
		vs, _ := res.ToArray()
		if len(vs) >= 2 {
			ret.val.Rank, _ = vs[0].AsInt64()
			ret.val.Score, _ = vs[1].AsFloat64()
		}
	}
	return ret
}

type XMessageSliceCmd interface {
	BaseCmd
	Val() []XMessage
	Result() ([]XMessage, error)
}

type XAutoClaimCmd interface {
	BaseCmd
	Val() (messages []XMessage, start string)
	Result() (messages []XMessage, start string, err error)
}

type XInfoConsumersCmd interface {
	BaseCmd
	Val() []XInfoConsumer
	Result() ([]XInfoConsumer, error)
}

type XInfoGroupsCmd interface {
	BaseCmd
	Val() []XInfoGroup
	Result() ([]XInfoGroup, error)
}

type XInfoStreamCmd interface {
	BaseCmd
	Val() XInfoStream
	Result() (XInfoStream, error)
}

type XInfoStreamFullCmd interface {
	BaseCmd
	Val() XInfoStreamFull
	Result() (XInfoStreamFull, error)
}

type XPendingCmd interface {
	BaseCmd
	Val() XPending
	Result() (XPending, error)
}

type XPendingExtCmd interface {
	BaseCmd
	Val() []XPendingExt
	Result() ([]XPendingExt, error)
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

func geoRadiusQueryArgs(q GeoRadiusQuery) []string {
	args := make([]string, 0, 2)
	args = append(args, strconv.FormatFloat(q.Radius, 'f', -1, 64))
	if q.Unit != "" {
		args = append(args, q.Unit)
	} else {
		args = append(args, KM)
	}
	if q.WithCoord {
		args = append(args, WITHCOORD)
	}
	if q.WithDist {
		args = append(args, WITHDIST)
	}
	if q.WithGeoHash {
		args = append(args, WITHHASH)
	}
	if q.Count > 0 {
		args = append(args, COUNT, strconv.FormatInt(q.Count, 10))
	}
	if q.Sort != "" {
		args = append(args, q.Sort)
	}
	if q.Store != "" {
		args = append(args, STORE)
		args = append(args, q.Store)
	}
	if q.StoreDist != "" {
		args = append(args, STOREDIST)
		args = append(args, q.StoreDist)
	}
	return args
}

func geoSearchQueryArgs(q GeoSearchQuery) []string {
	args := make([]string, 0, 2)
	if q.Member != "" {
		args = append(args, FROMMEMBER, q.Member)
	} else {
		args = append(args, FROMLONLAT, strconv.FormatFloat(q.Longitude, 'f', -1, 64), strconv.FormatFloat(q.Latitude, 'f', -1, 64))
	}
	if q.Radius > 0 {
		if q.RadiusUnit == "" {
			q.RadiusUnit = KM
		}
		args = append(args, BYRADIUS, strconv.FormatFloat(q.Radius, 'f', -1, 64), q.RadiusUnit)
	} else {
		if q.BoxUnit == "" {
			q.BoxUnit = KM
		}
		args = append(args, BYBOX, strconv.FormatFloat(q.BoxWidth, 'f', -1, 64), strconv.FormatFloat(q.BoxHeight, 'f', -1, 64), q.BoxUnit)
	}
	if q.Sort != "" {
		args = append(args, q.Sort)
	}
	if q.Count > 0 {
		args = append(args, COUNT, strconv.FormatInt(q.Count, 10))
		if q.CountAny {
			args = append(args, ANY)
		}
	}
	return args
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
