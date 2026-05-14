package redisson

import (
	"github.com/redis/rueidis"
	"github.com/redis/rueidis/rueidiscompat"
)

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
