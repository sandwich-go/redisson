package redisson

import (
	"time"

	"github.com/redis/rueidis"
	"github.com/redis/rueidis/rueidiscompat"
)

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
	ScanSlice(container interface{}) error
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

func wrapStringSliceCmd(cmd *rueidiscompat.StringSliceCmd) *stringSliceCmd {
	c := &stringSliceCmd{}
	c.SetVal(cmd.Val())
	c.SetErr(cmd.Err())
	return c
}

func (c *stringSliceCmd) ScanSlice(container interface{}) error {
	return scanSlice(c.Val(), container)
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
