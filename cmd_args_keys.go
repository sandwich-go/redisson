package redisson

import (
	"github.com/redis/rueidis"
)

type KeyValuesCmd interface {
	Val() (string, []string)
	Err() error
	Result() (string, []string, error)
}

type keyValuesCmd struct {
	err error
	val rueidis.KeyValues
}

// 注：newKeyValuesCmd 工厂已废弃。Pipeliner 路径通过 cmd_gen.go 直接构造
// `&keyValuesCmd{}`，再调用 from(res) 填充；非 Pipeliner 路径（rueidiscompat）已绕开此类型。

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

// 注：newKeyFlagsCmd 工厂已废弃。Pipeliner 路径通过 cmd_gen.go 直接构造空对象。

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

// 注：newScanCmd 工厂已废弃。Pipeliner 路径通过 cmd_gen.go 直接构造空对象。

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
