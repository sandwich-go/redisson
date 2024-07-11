package redisson

import (
	"github.com/redis/rueidis"
	"github.com/redis/rueidis/rueidiscompat"
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

type BoolCmd interface {
	BaseCmd
	Val() bool
	Result() (bool, error)
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

type FloatCmd interface {
	BaseCmd
	Val() float64
	Result() (float64, error)
}

type DurationCmd interface {
	BaseCmd
	Val() time.Duration
	Result() (time.Duration, error)
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

type StringSliceCmd interface {
	BaseCmd
	Val() []string
	Result() ([]string, error)
}

type BoolSliceCmd interface {
	BaseCmd
	Val() []bool
	Result() ([]bool, error)
}

type StringStringMapCmd interface {
	BaseCmd
	Val() map[string]string
	Result() (map[string]string, error)
	Scan(dest interface{}) error
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

type GeoLocationCmd interface {
	BaseCmd
	Val() []GeoLocation
	Result() ([]GeoLocation, error)
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

type RankWithScoreCmd interface {
	BaseCmd
	Val() RankScore
	Result() (RankScore, error)
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
