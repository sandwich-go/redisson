package redisson

import (
	"context"
	goredis "github.com/go-redis/redis/v8"
	"time"
)

//------------------------------------------------------------------------------

type PoolStats = goredis.PoolStats

//------------------------------------------------------------------------------

type Message = goredis.Message

//------------------------------------------------------------------------------

type BaseCmd interface {
	Err() error
}

type Cmd interface {
	BaseCmd
	String() string
	Result() (interface{}, error)
	Val() interface{}
	Text() (string, error)
	Int() (int, error)
	Int64() (int64, error)
	Uint64() (uint64, error)
	Float32() (float32, error)
	Float64() (float64, error)
	Bool() (bool, error)
	Slice() ([]interface{}, error)
	StringSlice() ([]string, error)
	Int64Slice() ([]int64, error)
	Uint64Slice() ([]uint64, error)
	Float32Slice() ([]float32, error)
	Float64Slice() ([]float64, error)
	BoolSlice() ([]bool, error)
}

type SliceCmd interface {
	BaseCmd
	Val() []interface{}
	Result() ([]interface{}, error)
	Scan(dst interface{}) error
}

type StatusCmd interface {
	BaseCmd
	Val() string
	Result() (string, error)
}

func newOKStatusCmd(args ...interface{}) StatusCmd {
	cmd := goredis.NewStatusCmd(context.Background(), args...)
	cmd.SetVal(OK)
	return cmd
}

func newStatusCmdWithError(err error, args ...interface{}) StatusCmd {
	cmd := goredis.NewStatusCmd(context.Background(), args...)
	cmd.SetErr(err)
	return cmd
}

type IntCmd interface {
	BaseCmd
	Val() int64
	Result() (int64, error)
	Uint64() (uint64, error)
}

type IntSliceCmd interface {
	BaseCmd
	Val() []int64
	Result() ([]int64, error)
}

type DurationCmd interface {
	BaseCmd
	Val() time.Duration
	Result() (time.Duration, error)
}

type TimeCmd interface {
	BaseCmd
	Val() time.Time
	Result() (time.Time, error)
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
	Scan(val interface{}) error
}

type FloatCmd interface {
	BaseCmd
	Val() float64
	Result() (float64, error)
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
	ScanSlice(container interface{}) error
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

//------------------------------------------------------------------------------

type XMessage = goredis.XMessage

type XMessageSliceCmd interface {
	BaseCmd
	Val() []XMessage
	Result() ([]XMessage, error)
}

//------------------------------------------------------------------------------

type XStream = goredis.XStream

type XStreamSliceCmd interface {
	BaseCmd
	Val() []XStream
	Result() ([]XStream, error)
}

//------------------------------------------------------------------------------

type XPending = goredis.XPending

type XPendingCmd interface {
	BaseCmd
	Val() *XPending
	Result() (*XPending, error)
}

//------------------------------------------------------------------------------

type XPendingExt = goredis.XPendingExt

type XPendingExtCmd interface {
	BaseCmd
	Val() []XPendingExt
	Result() ([]XPendingExt, error)
}

//------------------------------------------------------------------------------

type XAutoClaimCmd interface {
	BaseCmd
	Val() (messages []XMessage, start string)
	Result() (messages []XMessage, start string, err error)
}

//------------------------------------------------------------------------------

type XAutoClaimJustIDCmd interface {
	BaseCmd
	Val() (ids []string, start string)
	Result() (ids []string, start string, err error)
}

//------------------------------------------------------------------------------

type XInfoConsumer = goredis.XInfoConsumer

type XInfoConsumersCmd interface {
	BaseCmd
	Val() []XInfoConsumer
	Result() ([]XInfoConsumer, error)
}

//------------------------------------------------------------------------------

type XInfoGroup = goredis.XInfoGroup

type XInfoGroupsCmd interface {
	BaseCmd
	Val() []XInfoGroup
	Result() ([]XInfoGroup, error)
}

//------------------------------------------------------------------------------

type XInfoStream = goredis.XInfoStream

type XInfoStreamCmd interface {
	BaseCmd
	Val() *XInfoStream
	Result() (*XInfoStream, error)
}

//------------------------------------------------------------------------------

type (
	XInfoStreamFull            = goredis.XInfoStreamFull
	XInfoStreamGroup           = goredis.XInfoStreamGroup
	XInfoStreamGroupPending    = goredis.XInfoStreamGroupPending
	XInfoStreamConsumer        = goredis.XInfoStreamConsumer
	XInfoStreamConsumerPending = goredis.XInfoStreamConsumerPending
)

type XInfoStreamFullCmd interface {
	BaseCmd
	Val() *XInfoStreamFull
	Result() (*XInfoStreamFull, error)
}

//------------------------------------------------------------------------------

type Z = goredis.Z

type ZSliceCmd interface {
	BaseCmd
	Val() []Z
	Result() ([]Z, error)
}

//------------------------------------------------------------------------------

type ZWithKey = goredis.ZWithKey

type ZWithKeyCmd interface {
	BaseCmd
	Val() *ZWithKey
	Result() (*ZWithKey, error)
}

//------------------------------------------------------------------------------

type ScanCmd interface {
	BaseCmd
	Val() (keys []string, cursor uint64)
	Result() (keys []string, cursor uint64, err error)
}

//------------------------------------------------------------------------------

type (
	ClusterNode = goredis.ClusterNode
	ClusterSlot = goredis.ClusterSlot
)

type ClusterSlotsCmd interface {
	BaseCmd
	Val() []ClusterSlot
	Result() ([]ClusterSlot, error)
}

//------------------------------------------------------------------------------

type GeoLocation = goredis.GeoLocation

type GeoLocationCmd interface {
	BaseCmd
	Val() []GeoLocation
	Result() ([]GeoLocation, error)
}

//------------------------------------------------------------------------------

type GeoSearchLocationCmd interface {
	BaseCmd
	Val() []GeoLocation
	Result() ([]GeoLocation, error)
}

//------------------------------------------------------------------------------

type GeoPos = goredis.GeoPos

type GeoPosCmd interface {
	BaseCmd
	Val() []*GeoPos
	Result() ([]*GeoPos, error)
}

//------------------------------------------------------------------------------

type CommandInfo = goredis.CommandInfo

type CommandsInfoCmd interface {
	BaseCmd
	Val() map[string]*CommandInfo
	Result() (map[string]*CommandInfo, error)
}
