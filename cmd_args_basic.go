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

// anyCmd 把 rueidis 原生 RedisResult 拼装为满足 Cmd 接口的具体实现。风格与
// 周围 intCmd / boolCmd / stringCmd / floatCmd 一致（baseCmd[any] + from），
// 转换逻辑逐条对齐 rueidiscompat/command.go 的 Cmd 实现（rueidis v1.0.75）,
// 便于以后手动同步。使用场景:需要绕开 rueidiscompat.Adapter 走 rueidis 原生
// c.cmd.Do 时（如 EVAL 的 Retryable 变体）,用 newAnyCmd 把结果拼回 Cmd。
type anyCmd struct {
	baseCmd[any]
}

func newAnyCmd(res rueidis.RedisResult) Cmd {
	cmd := &anyCmd{}
	cmd.from(res)
	return cmd
}

// from 对齐 rueidiscompat.Cmd.from:res.ToAny + SetVal/SetErr,错误不区分 Redis
// Nil（Nil 也保留为 err）,与 adapter.Eval 路径完全一致。
func (c *anyCmd) from(res rueidis.RedisResult) {
	val, err := res.ToAny()
	if err != nil {
		c.SetErr(err)
		return
	}
	c.SetVal(val)
}

// 以下 15 个方法逐条对齐 rueidiscompat/command.go 的 (*Cmd).Xxx 实现。
// 复用本包已有的 toFloat32 / toFloat64（见 util.go），其余 helper（toStringAny /
// toInt64Any / toUint64Any / toBoolAny）就近定义在文件末尾。

func (c *anyCmd) Text() (string, error) {
	if c.err != nil {
		return "", c.err
	}
	return toStringAny(c.val)
}

func (c *anyCmd) Int() (int, error) {
	if c.err != nil {
		return 0, c.err
	}
	switch v := c.val.(type) {
	case int64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("redis: unexpected type=%T for Int", v)
	}
}

func (c *anyCmd) Int64() (int64, error) {
	if c.err != nil {
		return 0, c.err
	}
	return toInt64Any(c.val)
}

func (c *anyCmd) Uint64() (uint64, error) {
	if c.err != nil {
		return 0, c.err
	}
	return toUint64Any(c.val)
}

func (c *anyCmd) Float32() (float32, error) {
	if c.err != nil {
		return 0, c.err
	}
	return toFloat32(c.val)
}

func (c *anyCmd) Float64() (float64, error) {
	if c.err != nil {
		return 0, c.err
	}
	return toFloat64(c.val)
}

func (c *anyCmd) Bool() (bool, error) {
	if c.err != nil {
		return false, c.err
	}
	return toBoolAny(c.val)
}

func (c *anyCmd) Slice() ([]any, error) {
	if c.err != nil {
		return nil, c.err
	}
	switch v := c.val.(type) {
	case []any:
		return v, nil
	default:
		return nil, fmt.Errorf("redis: unexpected type=%T for Slice", v)
	}
}

func (c *anyCmd) StringSlice() ([]string, error) {
	slice, err := c.Slice()
	if err != nil {
		return nil, err
	}
	ss := make([]string, len(slice))
	for i, iface := range slice {
		val, err := toStringAny(iface)
		if err != nil {
			return nil, err
		}
		ss[i] = val
	}
	return ss, nil
}

func (c *anyCmd) Int64Slice() ([]int64, error) {
	slice, err := c.Slice()
	if err != nil {
		return nil, err
	}
	nums := make([]int64, len(slice))
	for i, iface := range slice {
		val, err := toInt64Any(iface)
		if err != nil {
			return nil, err
		}
		nums[i] = val
	}
	return nums, nil
}

func (c *anyCmd) Uint64Slice() ([]uint64, error) {
	slice, err := c.Slice()
	if err != nil {
		return nil, err
	}
	nums := make([]uint64, len(slice))
	for i, iface := range slice {
		val, err := toUint64Any(iface)
		if err != nil {
			return nil, err
		}
		nums[i] = val
	}
	return nums, nil
}

func (c *anyCmd) Float32Slice() ([]float32, error) {
	slice, err := c.Slice()
	if err != nil {
		return nil, err
	}
	floats := make([]float32, len(slice))
	for i, iface := range slice {
		val, err := toFloat32(iface)
		if err != nil {
			return nil, err
		}
		floats[i] = val
	}
	return floats, nil
}

func (c *anyCmd) Float64Slice() ([]float64, error) {
	slice, err := c.Slice()
	if err != nil {
		return nil, err
	}
	floats := make([]float64, len(slice))
	for i, iface := range slice {
		val, err := toFloat64(iface)
		if err != nil {
			return nil, err
		}
		floats[i] = val
	}
	return floats, nil
}

func (c *anyCmd) BoolSlice() ([]bool, error) {
	slice, err := c.Slice()
	if err != nil {
		return nil, err
	}
	bools := make([]bool, len(slice))
	for i, iface := range slice {
		val, err := toBoolAny(iface)
		if err != nil {
			return nil, err
		}
		bools[i] = val
	}
	return bools, nil
}

// toStringAny / toInt64Any / toUint64Any / toBoolAny 供 anyCmd 使用。签名/语义
// 逐条对齐 rueidiscompat/command.go 里的 toString / toInt64 / toUint64 / toBool
// （rueidis v1.0.75）。命名加 Any 后缀避免与本包已有 toString（见 delay.go,语
// 义为 Lua 兜底无损转 string，不同）冲突。
func toStringAny(val any) (string, error) {
	switch v := val.(type) {
	case string:
		return v, nil
	default:
		return "", fmt.Errorf("redis: unexpected type=%T for String", v)
	}
}

func toInt64Any(val any) (int64, error) {
	switch v := val.(type) {
	case int64:
		return v, nil
	case string:
		return strconv.ParseInt(v, 10, 64)
	default:
		return 0, fmt.Errorf("redis: unexpected type=%T for Int64", v)
	}
}

func toUint64Any(val any) (uint64, error) {
	switch v := val.(type) {
	case int64:
		return uint64(v), nil
	case string:
		return strconv.ParseUint(v, 10, 64)
	default:
		return 0, fmt.Errorf("redis: unexpected type=%T for Uint64", v)
	}
}

func toBoolAny(val any) (bool, error) {
	switch v := val.(type) {
	case int64:
		return v != 0, nil
	case string:
		return strconv.ParseBool(v)
	default:
		return false, fmt.Errorf("redis: unexpected type=%T for Bool", v)
	}
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
	// Redis PTTL/TTL 用 -2 表示 key 不存在、-1 表示无过期时间。统一乘以 precision，
	// sentinel 值符号语义保留（调用方判 val < 0 即可）。
	c.SetVal(time.Duration(val) * c.precision)
}
