package redisson

import (
	"github.com/redis/rueidis"
)

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

// 注：newZSliceWithKeyCmd 工厂已废弃。Pipeliner 路径通过 cmd_gen.go 直接构造空对象。

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
