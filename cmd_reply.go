package redisson

import (
	"context"
	"fmt"
	goredis "github.com/go-redis/redis/v8"
	"github.com/sandwich-go/rueidis"
	"github.com/sandwich-go/rueidis/rueidiscompat"
	"net"
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

func wrapError(err error) error {
	if err != nil && rueidis.IsRedisNil(err) {
		err = Nil
	}
	return err
}

func newCmd(res rueidis.RedisResult, args ...interface{}) Cmd {
	val, err := res.ToAny()
	cmd := goredis.NewCmd(context.Background(), args...)
	cmd.SetErr(wrapError(err))
	cmd.SetVal(val)
	return cmd
}

type SliceCmd interface {
	BaseCmd
	Val() []interface{}
	Result() ([]interface{}, error)
	Scan(dst interface{}) error
}

// args hmget or other
func newSliceCmd(res rueidis.RedisResult, args ...interface{}) SliceCmd {
	val, err := res.ToArray()
	cmd := goredis.NewSliceCmd(context.Background(), args...)
	if err != nil {
		cmd.SetErr(wrapError(err))
		return cmd
	}
	vals := make([]interface{}, len(val))
	for i, v := range val {
		if s, err := v.ToString(); err == nil {
			vals[i] = s
		}
	}
	cmd.SetVal(vals)
	return cmd
}

func newSliceCmdFromMap(res rueidis.RedisResult, args ...interface{}) SliceCmd {
	val, err := res.AsStrMap()
	cmd := goredis.NewSliceCmd(context.Background(), args...)
	if err != nil {
		cmd.SetErr(wrapError(err))
		return cmd
	}
	vals := make([]interface{}, 0, len(val)*2)
	for k, v := range val {
		vals = append(vals, k, v)
	}
	cmd.SetVal(vals)
	return cmd
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
	cmd.SetErr(wrapError(err))
	return cmd
}

func newStatusCmd(res rueidis.RedisResult, args ...interface{}) StatusCmd {
	val, err := res.ToString()
	cmd := goredis.NewStatusCmd(context.Background(), args...)
	cmd.SetErr(wrapError(err))
	cmd.SetVal(val)
	return cmd
}

type IntCmd interface {
	BaseCmd
	Val() int64
	Result() (int64, error)
	Uint64() (uint64, error)
}

func newIntCmdWithError(err error, args ...interface{}) IntCmd {
	cmd := goredis.NewIntCmd(context.Background(), args...)
	cmd.SetErr(wrapError(err))
	return cmd
}

func newIntCmd(res rueidis.RedisResult, args ...interface{}) IntCmd {
	val, err := res.AsInt64()
	cmd := goredis.NewIntCmd(context.Background(), args...)
	cmd.SetErr(wrapError(err))
	cmd.SetVal(val)
	return cmd
}

type IntSliceCmd interface {
	BaseCmd
	Val() []int64
	Result() ([]int64, error)
}

func newIntSliceCmd(res rueidis.RedisResult, args ...interface{}) IntSliceCmd {
	val, err := res.AsIntSlice()
	cmd := goredis.NewIntSliceCmd(context.Background(), args...)
	cmd.SetErr(wrapError(err))
	cmd.SetVal(val)
	return cmd
}

type DurationCmd interface {
	BaseCmd
	Val() time.Duration
	Result() (time.Duration, error)
}

func newDurationCmd(res rueidis.RedisResult, precision time.Duration, args ...interface{}) DurationCmd {
	val, err := res.AsInt64()
	cmd := goredis.NewDurationCmd(context.Background(), precision, args...)
	if err != nil {
		cmd.SetErr(wrapError(err))
		return cmd
	}
	if val > 0 {
		cmd.SetVal(time.Duration(val) * precision)
	} else {
		cmd.SetVal(time.Duration(val))
	}
	return cmd
}

type TimeCmd interface {
	BaseCmd
	Val() time.Time
	Result() (time.Time, error)
}

func newTimeCmd(res rueidis.RedisResult, args ...interface{}) TimeCmd {
	arr, err := res.ToArray()
	cmd := goredis.NewTimeCmd(context.Background(), args...)
	if err != nil {
		cmd.SetErr(wrapError(err))
		return cmd
	}
	if len(arr) < 2 {
		cmd.SetErr(fmt.Errorf("got %d, wanted 2", len(arr)))
		return cmd
	}
	sec, err0 := arr[0].AsInt64()
	if err0 != nil {
		cmd.SetErr(wrapError(err0))
		return cmd
	}
	microSec, err1 := arr[1].AsInt64()
	if err1 != nil {
		cmd.SetErr(wrapError(err1))
		return cmd
	}
	cmd.SetVal(time.Unix(sec, microSec*1000))
	return cmd
}

type BoolCmd interface {
	BaseCmd
	Val() bool
	Result() (bool, error)
}

func newBoolCmd(res rueidis.RedisResult, args ...interface{}) BoolCmd {
	val, err := res.AsBool()
	// `SET key value NX` returns nil when key already exists. But
	// `SETNX key value` returns bool (0/1). So convert nil to bool.
	if err != nil && rueidis.IsRedisNil(err) {
		val = false
		err = nil
	}
	cmd := goredis.NewBoolCmd(context.Background(), args...)
	cmd.SetErr(wrapError(err))
	cmd.SetVal(val)
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
	Scan(val interface{}) error
}

func newStringCmdFromStringCmd(res *rueidiscompat.StringCmd, args ...interface{}) StringCmd {
	cmd := goredis.NewStringCmd(context.Background(), args...)
	cmd.SetErr(wrapError(res.Err()))
	cmd.SetVal(res.Val())
	return cmd
}

func newStringCmd(res rueidis.RedisResult, args ...interface{}) StringCmd {
	val, err := res.ToString()
	cmd := goredis.NewStringCmd(context.Background(), args...)
	cmd.SetErr(wrapError(err))
	cmd.SetVal(val)
	return cmd
}

type FloatCmd interface {
	BaseCmd
	Val() float64
	Result() (float64, error)
}

func newFloatCmd(res rueidis.RedisResult, args ...interface{}) FloatCmd {
	val, err := res.AsFloat64()
	cmd := goredis.NewFloatCmd(context.Background(), args...)
	cmd.SetErr(wrapError(err))
	cmd.SetVal(val)
	return cmd
}

type FloatSliceCmd interface {
	BaseCmd
	Val() []float64
	Result() ([]float64, error)
}

func newFloatSliceCmd(res rueidis.RedisResult, args ...interface{}) FloatSliceCmd {
	val, err := res.AsFloatSlice()
	cmd := goredis.NewFloatSliceCmd(context.Background(), args...)
	cmd.SetErr(wrapError(err))
	cmd.SetVal(val)
	return cmd
}

type StringSliceCmd interface {
	BaseCmd
	Val() []string
	Result() ([]string, error)
	ScanSlice(container interface{}) error
}

func newStringSliceCmdFromStringSliceCmd(res *rueidiscompat.StringSliceCmd, args ...interface{}) StringSliceCmd {
	cmd := goredis.NewStringSliceCmd(context.Background(), args...)
	cmd.SetErr(wrapError(res.Err()))
	cmd.SetVal(res.Val())
	return cmd
}

func newStringSliceCmd(res rueidis.RedisResult, args ...interface{}) StringSliceCmd {
	val, err := res.AsStrSlice()
	cmd := goredis.NewStringSliceCmd(context.Background(), args...)
	cmd.SetErr(wrapError(err))
	cmd.SetVal(val)
	return cmd
}

func flattenStringSliceCmd(res rueidis.RedisResult, args ...interface{}) StringSliceCmd {
	arr, err := res.ToArray()
	cmd := goredis.NewStringSliceCmd(context.Background(), args...)
	if err != nil {
		cmd.SetErr(wrapError(err))
		return cmd
	}
	val := make([]string, 0, len(arr)*2)
	for _, v := range arr {
		s, err0 := v.AsStrSlice()
		if err0 != nil {
			cmd.SetErr(wrapError(err0))
			return cmd
		}
		val = append(val, s...)
	}
	cmd.SetVal(val)
	return cmd
}

type BoolSliceCmd interface {
	BaseCmd
	Val() []bool
	Result() ([]bool, error)
}

func newBoolSliceCmd(res rueidis.RedisResult, args ...interface{}) BoolSliceCmd {
	ints, err := res.AsIntSlice()
	cmd := goredis.NewBoolSliceCmd(context.Background(), args...)
	if err != nil {
		cmd.SetErr(wrapError(err))
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

func newStringStringMapCmd(res rueidis.RedisResult, args ...interface{}) StringStringMapCmd {
	val, err := res.AsStrMap()
	cmd := goredis.NewStringStringMapCmd(context.Background(), args...)
	cmd.SetErr(wrapError(err))
	cmd.SetVal(val)
	return cmd
}

type StringIntMapCmd interface {
	BaseCmd
	Val() map[string]int64
	Result() (map[string]int64, error)
}

func newStringIntMapCmd(res rueidis.RedisResult, args ...interface{}) StringIntMapCmd {
	val, err := res.AsIntMap()
	cmd := goredis.NewStringIntMapCmd(context.Background(), args...)
	cmd.SetErr(wrapError(err))
	cmd.SetVal(val)
	return cmd
}

type StringStructMapCmd interface {
	BaseCmd
	Val() map[string]struct{}
	Result() (map[string]struct{}, error)
}

func newStringStructMapCmd(res rueidis.RedisResult, args ...interface{}) StringStructMapCmd {
	strSlice, err := res.AsStrSlice()
	cmd := goredis.NewStringStructMapCmd(context.Background(), args...)
	if err != nil {
		cmd.SetErr(wrapError(err))
		return cmd
	}
	val := make(map[string]struct{}, len(strSlice))
	for _, v := range strSlice {
		val[v] = struct{}{}
	}
	cmd.SetVal(val)
	return cmd
}

//------------------------------------------------------------------------------

type XMessage = goredis.XMessage

type XMessageSliceCmd interface {
	BaseCmd
	Val() []XMessage
	Result() ([]XMessage, error)
}

func newXMessageSliceCmd(res rueidis.RedisResult, args ...interface{}) XMessageSliceCmd {
	val, err := res.AsXRange()
	cmd := goredis.NewXMessageSliceCmd(context.Background(), args...)
	if err != nil {
		cmd.SetErr(wrapError(err))
		return cmd
	}
	slice := make([]XMessage, len(val))
	for i, r := range val {
		slice[i] = newXMessage(r)
	}
	cmd.SetVal(slice)
	return cmd
}

func newXMessage(r rueidis.XRangeEntry) XMessage {
	if r.FieldValues == nil {
		return XMessage{ID: r.ID, Values: nil}
	}
	m := XMessage{ID: r.ID, Values: make(map[string]interface{}, len(r.FieldValues))}
	for k, v := range r.FieldValues {
		m.Values[k] = v
	}
	return m
}

//------------------------------------------------------------------------------

type XStream = goredis.XStream

type XStreamSliceCmd interface {
	BaseCmd
	Val() []XStream
	Result() ([]XStream, error)
}

func newXStreamSliceCmd(res rueidis.RedisResult, args ...interface{}) XStreamSliceCmd {
	streams, err := res.AsXRead()
	cmd := goredis.NewXStreamSliceCmd(context.Background(), args...)
	if err != nil {
		cmd.SetErr(wrapError(err))
		return cmd
	}
	val := make([]XStream, 0, len(streams))
	for name, messages := range streams {
		msgs := make([]XMessage, 0, len(messages))
		for _, r := range messages {
			msgs = append(msgs, newXMessage(r))
		}
		val = append(val, XStream{Stream: name, Messages: msgs})
	}
	cmd.SetVal(val)
	return cmd
}

//------------------------------------------------------------------------------

type XPending = goredis.XPending

type XPendingCmd interface {
	BaseCmd
	Val() *XPending
	Result() (*XPending, error)
}

func newXPendingCmd(res rueidis.RedisResult, args ...interface{}) XPendingCmd {
	arr, err := res.ToArray()
	cmd := goredis.NewXPendingCmd(context.Background(), args...)
	if err != nil {
		cmd.SetErr(wrapError(err))
		return cmd
	}
	if len(arr) < 4 {
		cmd.SetErr(fmt.Errorf("got %d, wanted 4", len(arr)))
		return cmd
	}
	count, err0 := arr[0].AsInt64()
	if err0 != nil {
		cmd.SetErr(wrapError(err0))
		return cmd
	}
	lower, err1 := arr[1].ToString()
	if err1 != nil {
		cmd.SetErr(wrapError(err1))
		return cmd
	}
	higher, err2 := arr[2].ToString()
	if err2 != nil {
		cmd.SetErr(wrapError(err2))
		return cmd
	}
	val := &XPending{
		Count:  count,
		Lower:  lower,
		Higher: higher,
	}
	consumerArr, err3 := arr[3].ToArray()
	if err3 != nil {
		cmd.SetErr(wrapError(err3))
		return cmd
	}
	for _, v := range consumerArr {
		consumer, err4 := v.ToArray()
		if err4 != nil {
			cmd.SetErr(wrapError(err4))
			return cmd
		}
		if len(consumer) < 2 {
			cmd.SetErr(fmt.Errorf("got %d, wanted 2", len(arr)))
			return cmd
		}
		consumerName, err5 := consumer[0].ToString()
		if err5 != nil {
			cmd.SetErr(wrapError(err5))
			return cmd
		}
		consumerPending, err6 := consumer[1].AsInt64()
		if err6 != nil {
			cmd.SetErr(wrapError(err6))
			return cmd
		}
		if val.Consumers == nil {
			val.Consumers = make(map[string]int64)
		}
		val.Consumers[consumerName] = consumerPending
	}
	cmd.SetVal(val)
	return cmd
}

//------------------------------------------------------------------------------

type XPendingExt = goredis.XPendingExt

type XPendingExtCmd interface {
	BaseCmd
	Val() []XPendingExt
	Result() ([]XPendingExt, error)
}

func newXPendingExtCmd(res rueidis.RedisResult, args ...interface{}) XPendingExtCmd {
	arrs, err := res.ToArray()
	cmd := goredis.NewXPendingExtCmd(context.Background(), args...)
	if err != nil {
		cmd.SetErr(wrapError(err))
		return cmd
	}
	val := make([]XPendingExt, 0, len(arrs))
	for _, v := range arrs {
		arr, err0 := v.ToArray()
		if err0 != nil {
			cmd.SetErr(wrapError(err0))
			return cmd
		}
		if len(arr) < 4 {
			cmd.SetErr(fmt.Errorf("got %d, wanted 4", len(arr)))
			return cmd
		}
		id, err1 := arr[0].ToString()
		if err1 != nil {
			cmd.SetErr(wrapError(err1))
			return cmd
		}
		consumer, err2 := arr[1].ToString()
		if err2 != nil {
			cmd.SetErr(wrapError(err2))
			return cmd
		}
		idle, err3 := arr[2].AsInt64()
		if err3 != nil {
			cmd.SetErr(wrapError(err3))
			return cmd
		}
		retryCount, err4 := arr[3].AsInt64()
		if err4 != nil {
			cmd.SetErr(wrapError(err4))
			return cmd
		}
		val = append(val, XPendingExt{
			ID:         id,
			Consumer:   consumer,
			Idle:       time.Duration(idle) * time.Millisecond,
			RetryCount: retryCount,
		})
	}
	cmd.SetVal(val)
	return cmd
}

//------------------------------------------------------------------------------

type XAutoClaimCmd interface {
	BaseCmd
	Val() (messages []XMessage, start string)
	Result() (messages []XMessage, start string, err error)
}

func newXAutoClaimCmd(res rueidis.RedisResult, args ...interface{}) XAutoClaimCmd {
	arr, err := res.ToArray()
	cmd := goredis.NewXAutoClaimCmd(context.Background(), args...)
	if err != nil {
		cmd.SetErr(wrapError(err))
		return cmd
	}
	if len(arr) < 2 {
		cmd.SetErr(fmt.Errorf("got %d, wanted 2", len(arr)))
		return cmd
	}
	start, err0 := arr[0].ToString()
	if err0 != nil {
		cmd.SetErr(wrapError(err0))
		return cmd
	}
	ranges, err1 := arr[1].AsXRange()
	if err1 != nil {
		cmd.SetErr(wrapError(err1))
		return cmd
	}
	val := make([]XMessage, 0, len(ranges))
	for _, r := range ranges {
		val = append(val, newXMessage(r))
	}
	cmd.SetVal(val, start)
	return cmd
}

//------------------------------------------------------------------------------

type XAutoClaimJustIDCmd interface {
	BaseCmd
	Val() (ids []string, start string)
	Result() (ids []string, start string, err error)
}

func newXAutoClaimJustIDCmd(res rueidis.RedisResult, args ...interface{}) XAutoClaimJustIDCmd {
	arr, err := res.ToArray()
	cmd := goredis.NewXAutoClaimJustIDCmd(context.Background(), args...)
	if err != nil {
		cmd.SetErr(wrapError(err))
		return cmd
	}
	if len(arr) < 2 {
		cmd.SetErr(fmt.Errorf("got %d, wanted 2", len(arr)))
		return cmd
	}
	start, err0 := arr[0].ToString()
	if err0 != nil {
		cmd.SetErr(wrapError(err0))
		return cmd
	}
	val, err1 := arr[1].AsStrSlice()
	if err1 != nil {
		cmd.SetErr(wrapError(err1))
		return cmd
	}
	cmd.SetVal(val, start)
	return cmd
}

//------------------------------------------------------------------------------

type XInfoConsumer = goredis.XInfoConsumer

type XInfoConsumersCmd interface {
	BaseCmd
	Val() []XInfoConsumer
	Result() ([]XInfoConsumer, error)
}

func newXInfoConsumersCmd(res rueidis.RedisResult, stream string, group string) XInfoConsumersCmd {
	arr, err := res.ToArray()
	cmd := goredis.NewXInfoConsumersCmd(context.Background(), stream, group)
	if err != nil {
		cmd.SetErr(wrapError(err))
		return cmd
	}
	val := make([]XInfoConsumer, 0, len(arr))
	for _, v := range arr {
		info, err0 := v.AsMap()
		if err0 != nil {
			cmd.SetErr(wrapError(err0))
			return cmd
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
			consumer.Idle = idle
		}
		val = append(val, consumer)
	}
	cmd.SetVal(val)
	return cmd
}

//------------------------------------------------------------------------------

type XInfoGroup = goredis.XInfoGroup

type XInfoGroupsCmd interface {
	BaseCmd
	Val() []XInfoGroup
	Result() ([]XInfoGroup, error)
}

func newXInfoGroupsCmd(res rueidis.RedisResult, stream string) XInfoGroupsCmd {
	arr, err := res.ToArray()
	cmd := goredis.NewXInfoGroupsCmd(context.Background(), stream)
	if err != nil {
		cmd.SetErr(wrapError(err))
		return cmd
	}
	groupInfos := make([]XInfoGroup, 0, len(arr))
	for _, v := range arr {
		info, err0 := v.AsMap()
		if err0 != nil {
			cmd.SetErr(wrapError(err0))
			return cmd
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
		//if attr, ok := info["entries-read"]; ok {
		//	group.EntriesRead, _ = attr.AsInt64()
		//}
		//if attr, ok := info["lag"]; ok {
		//	group.Lag, _ = attr.AsInt64()
		//}
		if attr, ok := info["last-delivered-id"]; ok {
			group.LastDeliveredID, _ = attr.ToString()
		}
		groupInfos = append(groupInfos, group)
	}
	cmd.SetVal(groupInfos)
	return cmd
}

//------------------------------------------------------------------------------

type XInfoStream = goredis.XInfoStream

type XInfoStreamCmd interface {
	BaseCmd
	Val() *XInfoStream
	Result() (*XInfoStream, error)
}

func newXInfoStreamCmd(res rueidis.RedisResult, stream string) XInfoStreamCmd {
	kv, err := res.AsMap()
	cmd := goredis.NewXInfoStreamCmd(context.Background(), stream)
	if err != nil {
		cmd.SetErr(wrapError(err))
		return cmd
	}
	var val = new(XInfoStream)
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
	//if v, ok := kv["max-deleted-entry-id"]; ok {
	//	val.MaxDeletedEntryID, _ = v.ToString()
	//}
	//if v, ok := kv["recorded-first-entry-id"]; ok {
	//	val.RecordedFirstEntryID, _ = v.ToString()
	//}
	//if v, ok := kv["entries-added"]; ok {
	//	val.EntriesAdded, _ = v.AsInt64()
	//}
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
	cmd.SetVal(val)
	return cmd
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

func newXInfoStreamFullCmd(res rueidis.RedisResult, args ...interface{}) XInfoStreamFullCmd {
	kv, err := res.AsMap()
	cmd := goredis.NewXInfoStreamFullCmd(context.Background(), args...)
	if err != nil {
		cmd.SetErr(wrapError(err))
		return cmd
	}
	var val = new(XInfoStreamFull)
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
	//if v, ok := kv["entries-added"]; ok {
	//	val.EntriesAdded, _ = v.AsInt64()
	//}
	//if v, ok := kv["max-deleted-entry-id"]; ok {
	//	val.MaxDeletedEntryID, _ = v.ToString()
	//}
	//if v, ok := kv["recorded-first-entry-id"]; ok {
	//	val.RecordedFirstEntryID, _ = v.ToString()
	//}
	if v, ok := kv["groups"]; ok {
		val.Groups, err = readStreamGroups(v)
		if err != nil {
			cmd.SetErr(wrapError(err))
			return cmd
		}
	}
	if v, ok := kv["entries"]; ok {
		ranges, err0 := v.AsXRange()
		if err0 != nil {
			cmd.SetErr(wrapError(err0))
			return cmd
		}
		val.Entries = make([]XMessage, 0, len(ranges))
		for _, r := range ranges {
			val.Entries = append(val.Entries, newXMessage(r))
		}
	}
	cmd.SetVal(val)
	return cmd
}

func readStreamGroups(res rueidis.RedisMessage) ([]XInfoStreamGroup, error) {
	arr, err0 := res.ToArray()
	if err0 != nil {
		return nil, err0
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
		//if attr, ok := info["entries-read"]; ok {
		//	group.EntriesRead, _ = attr.AsInt64()
		//}
		//if attr, ok := info["lag"]; ok {
		//	group.Lag, _ = attr.AsInt64()
		//}
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
	arr, err0 := res.ToArray()
	if err0 != nil {
		return nil, err0
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
		delivery, err1 := info[2].AsInt64()
		if err1 != nil {
			return nil, err1
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
	arr, err0 := res.ToArray()
	if err0 != nil {
		return nil, err0
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
			pending, err1 := attr.ToArray()
			if err1 != nil {
				return nil, err1
			}
			c.Pending = make([]XInfoStreamConsumerPending, 0, len(pending))
			for _, v := range pending {
				pendingInfo, err2 := v.ToArray()
				if err2 != nil {
					return nil, err2
				}
				if len(pendingInfo) < 3 {
					return nil, fmt.Errorf("got %d, wanted 3", len(arr))
				}
				var p XInfoStreamConsumerPending
				p.ID, err = pendingInfo[0].ToString()
				if err != nil {
					return nil, err
				}
				delivery, err3 := pendingInfo[1].AsInt64()
				if err3 != nil {
					return nil, err3
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

//------------------------------------------------------------------------------

type Z = goredis.Z

type ZSliceCmd interface {
	BaseCmd
	Val() []Z
	Result() ([]Z, error)
}

func newZSliceCmd(res rueidis.RedisResult, args ...interface{}) ZSliceCmd {
	scores, err := res.AsZScores()
	cmd := goredis.NewZSliceCmd(context.Background(), args...)
	if err != nil {
		cmd.SetErr(wrapError(err))
		return cmd
	}
	val := make([]Z, 0, len(scores))
	for _, s := range scores {
		val = append(val, Z{Member: s.Member, Score: s.Score})
	}
	cmd.SetVal(val)
	return cmd
}

func newZSliceSingleCmd(res rueidis.RedisResult, args ...interface{}) ZSliceCmd {
	s, err := res.AsZScore()
	cmd := goredis.NewZSliceCmd(context.Background(), args...)
	if err != nil {
		cmd.SetErr(wrapError(err))
		return cmd
	}
	cmd.SetVal([]Z{{
		Member: s.Member,
		Score:  s.Score,
	}})
	return cmd
}

//------------------------------------------------------------------------------

type ZWithKey = goredis.ZWithKey

type ZWithKeyCmd interface {
	BaseCmd
	Val() *ZWithKey
	Result() (*ZWithKey, error)
}

func newZWithKeyCmd(res rueidis.RedisResult, args ...interface{}) ZWithKeyCmd {
	arr, err := res.ToArray()
	cmd := goredis.NewZWithKeyCmd(context.Background(), args...)
	if err != nil {
		cmd.SetErr(wrapError(err))
		return cmd
	}
	if len(arr) < 3 {
		cmd.SetErr(fmt.Errorf("got %d, wanted 3", len(arr)))
		return cmd
	}
	val := &ZWithKey{}
	val.Key, err = arr[0].ToString()
	if err != nil {
		cmd.SetErr(wrapError(err))
		return cmd
	}
	val.Member, err = arr[1].ToString()
	if err != nil {
		cmd.SetErr(wrapError(err))
		return cmd
	}
	val.Score, err = arr[2].AsFloat64()
	if err != nil {
		cmd.SetErr(wrapError(err))
		return cmd
	}
	cmd.SetVal(val)
	return cmd
}

//------------------------------------------------------------------------------

type ScanCmd interface {
	BaseCmd
	Val() (keys []string, cursor uint64)
	Result() (keys []string, cursor uint64, err error)
}

func newScanCmd(res rueidis.RedisResult, args ...interface{}) ScanCmd {
	ret, err := res.ToArray()
	// todo for ScanIterator
	cmd := goredis.NewScanCmd(context.Background(), nil, args...)
	if err != nil {
		cmd.SetErr(wrapError(err))
		return cmd
	}
	cursor, err0 := ret[0].AsInt64()
	if err0 != nil {
		cmd.SetErr(wrapError(err0))
		return cmd
	}
	page, err1 := ret[1].AsStrSlice()
	if err1 != nil {
		cmd.SetErr(wrapError(err1))
		return cmd
	}
	cmd.SetVal(page, uint64(cursor))
	return cmd
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

func newClusterSlotsCmd(res rueidis.RedisResult, args ...interface{}) ClusterSlotsCmd {
	arr, err := res.ToArray()
	cmd := goredis.NewClusterSlotsCmd(context.Background(), args...)
	if err != nil {
		cmd.SetErr(wrapError(err))
		return cmd
	}
	val := make([]ClusterSlot, 0, len(arr))
	for _, v := range arr {
		slots, err0 := v.ToArray()
		if err0 != nil {
			cmd.SetErr(wrapError(err0))
			return cmd
		}
		if len(slots) < 2 {
			cmd.SetErr(fmt.Errorf("got %d, excpected atleast 2", len(slots)))
			return cmd
		}
		start, err1 := slots[0].AsInt64()
		if err1 != nil {
			cmd.SetErr(wrapError(err1))
			return cmd
		}
		end, err2 := slots[1].AsInt64()
		if err2 != nil {
			cmd.SetErr(wrapError(err2))
			return cmd
		}
		nodes := make([]ClusterNode, len(slots)-2)
		for i, j := 2, 0; i < len(nodes); i, j = i+1, j+1 {
			node, err3 := slots[i].ToArray()
			if err3 != nil {
				cmd.SetErr(wrapError(err3))
				return cmd
			}
			if len(node) != 2 && len(node) != 3 {
				cmd.SetErr(fmt.Errorf("got %d, expected 2 or 3", len(node)))
				return cmd
			}
			ip, err4 := node[0].ToString()
			if err4 != nil {
				cmd.SetErr(wrapError(err4))
				return cmd
			}
			port, err5 := node[1].AsInt64()
			if err5 != nil {
				cmd.SetErr(wrapError(err5))
				return cmd
			}
			nodes[j].Addr = net.JoinHostPort(ip, str(port))
			if len(node) == 3 {
				id, err6 := node[2].ToString()
				if err6 != nil {
					cmd.SetErr(wrapError(err6))
					return cmd
				}
				nodes[j].ID = id
			}
		}
		val = append(val, ClusterSlot{
			Start: int(start),
			End:   int(end),
			Nodes: nodes,
		})
	}
	cmd.SetVal(val)
	return cmd
}

//------------------------------------------------------------------------------

type GeoLocation = goredis.GeoLocation

type GeoLocationCmd interface {
	BaseCmd
	Val() []GeoLocation
	Result() ([]GeoLocation, error)
}

type geoLocationCmdSetter interface {
	SetErr(error)
	SetVal([]GeoLocation)
}

func newGeoLocationCmdWithError(err error, args ...interface{}) GeoLocationCmd {
	cmd := goredis.NewGeoLocationCmd(context.Background(), nil, args...)
	cmd.SetErr(wrapError(err))
	return cmd
}

func fillGeoLocationCmd(res rueidis.RedisResult, cmd geoLocationCmdSetter, withDist, withGeoHash, withCoord bool) {
	arr, err := res.ToArray()
	if err != nil {
		cmd.SetErr(wrapError(err))
	}
	val := make([]GeoLocation, 0, len(arr))
	if !withDist && !withGeoHash && !withCoord {
		for _, v := range arr {
			name, err0 := v.ToString()
			if err0 != nil {
				cmd.SetErr(wrapError(err0))
				return
			}
			val = append(val, GeoLocation{Name: name})
		}
		cmd.SetVal(val)
		return
	}
	for _, v := range arr {
		info, err1 := v.ToArray()
		if err1 != nil {
			cmd.SetErr(wrapError(err1))
			return
		}
		var loc GeoLocation
		var i int
		loc.Name, err = info[i].ToString()
		i++
		if err != nil {
			cmd.SetErr(wrapError(err))
			return
		}
		if withDist {
			loc.Dist, err = info[i].AsFloat64()
			i++
			if err != nil {
				cmd.SetErr(wrapError(err))
				return
			}
		}
		if withGeoHash {
			loc.GeoHash, err = info[i].AsInt64()
			i++
			if err != nil {
				cmd.SetErr(wrapError(err))
				return
			}
		}
		if withCoord {
			cord, err2 := info[i].ToArray()
			if err2 != nil {
				cmd.SetErr(wrapError(err2))
				return
			}
			if len(cord) != 2 {
				cmd.SetErr(fmt.Errorf("got %d, expected 2", len(info)))
				return
			}
			loc.Longitude, err = cord[0].AsFloat64()
			if err != nil {
				cmd.SetErr(wrapError(err))
				return
			}
			loc.Latitude, err = cord[1].AsFloat64()
			if err != nil {
				cmd.SetErr(wrapError(err))
				return
			}
		}
		val = append(val, loc)
	}
	cmd.SetVal(val)
}

func newGeoLocationCmd(res rueidis.RedisResult, q goredis.GeoRadiusQuery, args ...interface{}) GeoLocationCmd {
	cmd := goredis.NewGeoLocationCmd(context.Background(), &q, args...)
	fillGeoLocationCmd(res, cmd, q.WithDist, q.WithGeoHash, q.WithCoord)
	return cmd
}

//------------------------------------------------------------------------------

type GeoSearchLocationCmd interface {
	BaseCmd
	Val() []GeoLocation
	Result() ([]GeoLocation, error)
}

func newGeoSearchLocationCmd(res rueidis.RedisResult, q goredis.GeoSearchLocationQuery, args ...interface{}) GeoSearchLocationCmd {
	cmd := goredis.NewGeoSearchLocationCmd(context.Background(), &q, args...)
	fillGeoLocationCmd(res, cmd, q.WithDist, q.WithHash, q.WithCoord)
	return cmd
}

//------------------------------------------------------------------------------

type GeoPos = goredis.GeoPos

type GeoPosCmd interface {
	BaseCmd
	Val() []*GeoPos
	Result() ([]*GeoPos, error)
}

func newGeoPosCmd(res rueidis.RedisResult, args ...interface{}) GeoPosCmd {
	arr, err := res.ToArray()
	cmd := goredis.NewGeoPosCmd(context.Background(), args...)
	if err != nil {
		cmd.SetErr(wrapError(err))
		return cmd
	}
	val := make([]*GeoPos, 0, len(arr))
	for _, v := range arr {
		loc, err0 := v.ToArray()
		if err0 != nil {
			if rueidis.IsRedisNil(err0) {
				val = append(val, nil)
				continue
			}
			cmd.SetErr(wrapError(err0))
			return cmd
		}
		if len(loc) != 2 {
			cmd.SetErr(fmt.Errorf("got %d, expected 2", len(loc)))
			return cmd
		}
		long, err1 := loc[0].AsFloat64()
		if err1 != nil {
			cmd.SetErr(wrapError(err1))
			return cmd
		}
		lat, err2 := loc[1].AsFloat64()
		if err2 != nil {
			cmd.SetErr(wrapError(err2))
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

//------------------------------------------------------------------------------

type CommandInfo = goredis.CommandInfo

type CommandsInfoCmd interface {
	BaseCmd
	Val() map[string]*CommandInfo
	Result() (map[string]*CommandInfo, error)
}

func newCommandsInfoCmd(res rueidis.RedisResult, args ...interface{}) CommandsInfoCmd {
	arr, err := res.ToArray()
	cmd := goredis.NewCommandsInfoCmd(context.Background(), args...)
	if err != nil {
		cmd.SetErr(wrapError(err))
		return cmd
	}
	val := make(map[string]*CommandInfo, len(arr))
	for _, v := range arr {
		info, err0 := v.ToArray()
		if err0 != nil {
			cmd.SetErr(wrapError(err0))
			return cmd
		}
		if len(info) < 6 {
			cmd.SetErr(fmt.Errorf("got %d, wanted at least 6", len(info)))
			return cmd
		}
		var cmdInfo = &CommandInfo{}
		cmdInfo.Name, err = info[0].ToString()
		if err != nil {
			cmd.SetErr(wrapError(err))
			return cmd
		}
		arity, err1 := info[1].AsInt64()
		if err1 != nil {
			cmd.SetErr(wrapError(err1))
			return cmd
		}
		cmdInfo.Arity = int8(arity)
		cmdInfo.Flags, err = info[2].AsStrSlice()
		if err != nil {
			if rueidis.IsRedisNil(err) {
				cmdInfo.Flags = []string{}
			} else {
				cmd.SetErr(wrapError(err))
				return cmd
			}
		}
		firstKeyPos, err2 := info[3].AsInt64()
		if err2 != nil {
			cmd.SetErr(wrapError(err2))
			return cmd
		}
		cmdInfo.FirstKeyPos = int8(firstKeyPos)
		lastKeyPos, err3 := info[4].AsInt64()
		if err3 != nil {
			cmd.SetErr(wrapError(err3))
			return cmd
		}
		cmdInfo.LastKeyPos = int8(lastKeyPos)
		stepCount, err4 := info[5].AsInt64()
		if err4 != nil {
			cmd.SetErr(wrapError(err4))
			return cmd
		}
		cmdInfo.StepCount = int8(stepCount)
		for _, flag := range cmdInfo.Flags {
			if flag == "readonly" {
				cmdInfo.ReadOnly = true
				break
			}
		}
		if len(info) == 6 {
			val[cmdInfo.Name] = cmdInfo
			continue
		}
		cmdInfo.ACLFlags, err = info[6].AsStrSlice()
		if err != nil {
			if rueidis.IsRedisNil(err) {
				cmdInfo.ACLFlags = []string{}
			} else {
				cmd.SetErr(wrapError(err))
				return cmd
			}
		}
		val[cmdInfo.Name] = cmdInfo
	}
	cmd.SetVal(val)
	return cmd
}
