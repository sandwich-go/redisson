package redisson

import (
	"fmt"
	"strconv"
	"time"

	"github.com/redis/rueidis"
	"github.com/redis/rueidis/rueidiscompat"
)

type BaseCmd interface {
	Err() error
}

type fromRedisResult interface {
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
	Scan(val interface{}) error
}

type stringCmd struct {
	baseCmd[string]
}

func wrapStringCmd(cmd *rueidiscompat.StringCmd) *stringCmd {
	c := &stringCmd{}
	c.SetVal(cmd.Val())
	c.SetErr(cmd.Err())
	return c
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

func (c *stringCmd) Scan(val interface{}) error {
	if c.err != nil {
		return c.err
	}
	return scan([]byte(c.val), val)
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
