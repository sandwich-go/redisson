package redisson

import (
	"fmt"

	"github.com/redis/rueidis"
)

type ClusterSlotsCmd interface {
	BaseCmd
	Val() []ClusterSlot
	Result() ([]ClusterSlot, error)
}

type ClusterShardsCmd interface {
	BaseCmd
	Val() []ClusterShard
	Result() ([]ClusterShard, error)
}

type GeoPosCmd interface {
	BaseCmd
	Val() []*GeoPos
	Result() ([]*GeoPos, error)
}

type geoPosCmd struct {
	baseCmd[[]*GeoPos]
}

func newGeoPosCmd(res rueidis.RedisResult) GeoPosCmd {
	cmd := &geoPosCmd{}
	cmd.from(res)
	return cmd
}

func (c *geoPosCmd) from(res rueidis.RedisResult) {
	arr, err := res.ToArray()
	if err != nil {
		c.SetErr(err)
		return
	}
	val := make([]*GeoPos, 0, len(arr))
	for _, v := range arr {
		loc, err := v.ToArray()
		if err != nil {
			if rueidis.IsRedisNil(err) {
				val = append(val, nil)
				continue
			}
			c.SetErr(err)
			return
		}
		if len(loc) != 2 {
			c.SetErr(fmt.Errorf("got %d, expected 2", len(loc)))
			return
		}
		long, err := loc[0].AsFloat64()
		if err != nil {
			c.SetErr(err)
			return
		}
		lat, err := loc[1].AsFloat64()
		if err != nil {
			c.SetErr(err)
			return
		}
		val = append(val, &GeoPos{
			Longitude: long,
			Latitude:  lat,
		})
	}
	c.SetVal(val)
}

type GeoLocationCmd interface {
	BaseCmd
	Val() []GeoLocation
	Result() ([]GeoLocation, error)
}

type geoLocationCmd struct {
	baseCmd[[]rueidis.GeoLocation]
}

func newGeoLocationCmd(res rueidis.RedisResult) GeoLocationCmd {
	cmd := &geoLocationCmd{}
	cmd.from(res)
	return cmd
}

func (c *geoLocationCmd) from(res rueidis.RedisResult) {
	c.val, c.err = res.AsGeosearch()
}

type KeyValueSliceCmd interface {
	BaseCmd
	Val() []KeyValue
	Result() ([]KeyValue, error)
}

type keyValueSliceCmd struct {
	baseCmd[[]KeyValue]
}

// 注：newKeyValueSliceCmd 工厂已废弃。Pipeliner 路径通过 cmd_gen.go 直接构造空对象。

func (c *keyValueSliceCmd) from(res rueidis.RedisResult) {
	arr, err := res.ToArray()
	for _, a := range arr {
		kv, _ := a.AsStrSlice()
		for i := 0; i < len(kv); i += 2 {
			c.val = append(c.val, KeyValue{Key: kv[i], Value: kv[i+1]})
		}
	}
	c.SetErr(err)
}

type CommandsInfoCmd interface {
	BaseCmd
	Val() map[string]CommandInfo
	Result() (map[string]CommandInfo, error)
}

type commandsInfoCmd struct {
	baseCmd[map[string]CommandInfo]
}

// 注：newCommandsInfoCmd 工厂已废弃。Pipeliner 路径通过 cmd_gen.go 直接构造空对象。

func (c *commandsInfoCmd) from(res rueidis.RedisResult) {
	arr, err := res.ToArray()
	if err != nil {
		c.SetErr(err)
		return
	}
	val := make(map[string]CommandInfo, len(arr))
	for _, v := range arr {
		info, err := v.ToArray()
		if err != nil {
			c.SetErr(err)
			return
		}
		if len(info) < 6 {
			c.SetErr(fmt.Errorf("got %d, wanted at least 6", len(info)))
			return
		}
		var _cmd CommandInfo
		_cmd.Name, err = info[0].ToString()
		if err != nil {
			c.SetErr(err)
			return
		}
		_cmd.Arity, err = info[1].AsInt64()
		if err != nil {
			c.SetErr(err)
			return
		}
		_cmd.Flags, err = info[2].AsStrSlice()
		if err != nil {
			if rueidis.IsRedisNil(err) {
				_cmd.Flags = []string{}
			} else {
				c.SetErr(err)
				return
			}
		}
		_cmd.FirstKeyPos, err = info[3].AsInt64()
		if err != nil {
			c.SetErr(err)
			return
		}
		_cmd.LastKeyPos, err = info[4].AsInt64()
		if err != nil {
			c.SetErr(err)
			return
		}
		_cmd.StepCount, err = info[5].AsInt64()
		if err != nil {
			c.SetErr(err)
			return
		}
		for _, flag := range _cmd.Flags {
			if flag == "readonly" {
				_cmd.ReadOnly = true
				break
			}
		}
		if len(info) == 6 {
			val[_cmd.Name] = _cmd
			continue
		}
		_cmd.ACLFlags, err = info[6].AsStrSlice()
		if err != nil {
			if rueidis.IsRedisNil(err) {
				_cmd.ACLFlags = []string{}
			} else {
				c.SetErr(err)
				return
			}
		}
		val[_cmd.Name] = _cmd
	}
	c.SetVal(val)
}
