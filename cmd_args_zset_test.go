package redisson

import (
	"errors"
	"testing"
)

func TestZSliceWithKeyCmd_SetVal(t *testing.T) {
	c := &zSliceWithKeyCmd{}
	c.SetVal("zk", []Z{{Member: "m1", Score: 1.5}, {Member: "m2", Score: 2.5}})
	k, vs := c.Val()
	if k != "zk" {
		t.Errorf("key=%q", k)
	}
	if len(vs) != 2 || vs[0].Member != "m1" {
		t.Errorf("vals=%v", vs)
	}
	k2, vs2, err := c.Result()
	if err != nil {
		t.Fatal(err)
	}
	if k2 != "zk" || len(vs2) != 2 {
		t.Errorf("Result mismatch")
	}
}

func TestZSliceWithKeyCmd_SetErr(t *testing.T) {
	c := &zSliceWithKeyCmd{}
	c.SetErr(errors.New("e"))
	if c.Err() == nil {
		t.Error("Err should propagate")
	}
}

func TestZSliceCmd_BaseAccess(t *testing.T) {
	c := &zSliceCmd{}
	c.SetVal([]Z{{Member: "x", Score: 9}})
	if v := c.Val(); len(v) != 1 || v[0].Score != 9 {
		t.Errorf("got %+v", v)
	}
}

func TestRankWithScoreCmd_BaseAccess(t *testing.T) {
	c := &rankWithScoreCmd{}
	c.SetVal(RankScore{Rank: 3, Score: 1.23})
	v := c.Val()
	if v.Rank != 3 || v.Score < 1.22 || v.Score > 1.24 {
		t.Errorf("got %+v", v)
	}
}
