package redisson

import (
	"fmt"
	"strconv"
	"time"

	"github.com/redis/rueidis"
)

type XMessageSliceCmd interface {
	BaseCmd
	Val() []XMessage
	Result() ([]XMessage, error)
}

type xMessageSliceCmd struct {
	baseCmd[[]XMessage]
}

// 注：newXMessageSliceCmd 工厂已废弃。Pipeliner 路径通过 cmd_gen.go 直接构造空对象。

func (c *xMessageSliceCmd) from(res rueidis.RedisResult) {
	val, err := res.AsXRange()
	c.SetErr(err)
	c.val = make([]XMessage, len(val))
	for i, r := range val {
		c.val[i] = newXMessage(r)
	}
}

type XAutoClaimCmd interface {
	BaseCmd
	Val() (messages []XMessage, start string)
	Result() (messages []XMessage, start string, err error)
}

type xAutoClaimCmd struct {
	err   error
	start string
	val   []XMessage
}

func newXMessage(r rueidis.XRangeEntry) XMessage {
	if r.FieldValues == nil {
		return XMessage{ID: r.ID, Values: nil}
	}
	m := XMessage{ID: r.ID, Values: make(map[string]any, len(r.FieldValues))}
	for k, v := range r.FieldValues {
		m.Values[k] = v
	}
	return m
}

// 注：newXAutoClaimCmd 工厂已废弃。Pipeliner 路径通过 cmd_gen.go 直接构造空对象。

func (c *xAutoClaimCmd) from(res rueidis.RedisResult) {
	arr, err := res.ToArray()
	if err != nil {
		c.err = err
		return
	}
	if len(arr) < 2 {
		c.err = fmt.Errorf("got %d, wanted 2", len(arr))
		return
	}
	start, err := arr[0].ToString()
	if err != nil {
		c.err = err
		return
	}
	ranges, err := arr[1].AsXRange()
	if err != nil {
		c.err = err
		return
	}
	val := make([]XMessage, 0, len(ranges))
	for _, r := range ranges {
		val = append(val, newXMessage(r))
	}
	c.val = val
	c.start = start
	c.err = err
}

func (c *xAutoClaimCmd) SetVal(val []XMessage, start string) {
	c.val = val
	c.start = start
}

func (c *xAutoClaimCmd) SetErr(err error) { c.err = err }
func (c *xAutoClaimCmd) Val() (messages []XMessage, start string) {
	return c.val, c.start
}
func (c *xAutoClaimCmd) Err() error {
	return c.err
}
func (c *xAutoClaimCmd) Result() (messages []XMessage, start string, err error) {
	return c.val, c.start, c.err
}

type XInfoConsumersCmd interface {
	BaseCmd
	Val() []XInfoConsumer
	Result() ([]XInfoConsumer, error)
}

type xInfoConsumersCmd struct {
	baseCmd[[]XInfoConsumer]
}

// 注：newXInfoConsumersCmd 工厂已废弃。Pipeliner 路径通过 cmd_gen.go 直接构造空对象。

func (c *xInfoConsumersCmd) from(res rueidis.RedisResult) {
	arr, err := res.ToArray()
	if err != nil {
		c.SetErr(err)
		return
	}
	val := make([]XInfoConsumer, 0, len(arr))
	for _, v := range arr {
		info, err := v.AsMap()
		if err != nil {
			c.SetErr(err)
			return
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
			consumer.Idle = time.Duration(idle) * time.Millisecond
		}
		val = append(val, consumer)
	}
	c.SetVal(val)
}

type XInfoGroupsCmd interface {
	BaseCmd
	Val() []XInfoGroup
	Result() ([]XInfoGroup, error)
}

type xInfoGroupsCmd struct {
	baseCmd[[]XInfoGroup]
}

// 注：newXInfoGroupsCmd 工厂已废弃。Pipeliner 路径通过 cmd_gen.go 直接构造空对象。

func (c *xInfoGroupsCmd) from(res rueidis.RedisResult) {
	arr, err := res.ToArray()
	if err != nil {
		c.SetErr(err)
		return
	}
	groupInfos := make([]XInfoGroup, 0, len(arr))
	for _, v := range arr {
		info, err := v.AsMap()
		if err != nil {
			c.SetErr(err)
			return
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
		if attr, ok := info["entries-read"]; ok {
			group.EntriesRead, _ = attr.AsInt64()
		}
		if attr, ok := info["lag"]; ok {
			group.Lag, _ = attr.AsInt64()
		}
		if attr, ok := info["last-delivered-id"]; ok {
			group.LastDeliveredID, _ = attr.ToString()
		}
		groupInfos = append(groupInfos, group)
	}
	c.SetVal(groupInfos)
}

type XInfoStreamCmd interface {
	BaseCmd
	Val() XInfoStream
	Result() (XInfoStream, error)
}

type xInfoStreamCmd struct {
	baseCmd[XInfoStream]
}

// 注：newXInfoStreamCmd 工厂已废弃。Pipeliner 路径通过 cmd_gen.go 直接构造空对象。

func (c *xInfoStreamCmd) from(res rueidis.RedisResult) {
	kv, err := res.AsMap()
	if err != nil {
		c.SetErr(err)
		return
	}
	var val XInfoStream
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
	if v, ok := kv["max-deleted-entry-id"]; ok {
		val.MaxDeletedEntryID, _ = v.ToString()
	}
	if v, ok := kv["recorded-first-entry-id"]; ok {
		val.RecordedFirstEntryID, _ = v.ToString()
	}
	if v, ok := kv["entries-added"]; ok {
		val.EntriesAdded, _ = v.AsInt64()
	}
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
	c.SetVal(val)
}

type XInfoStreamFullCmd interface {
	BaseCmd
	Val() XInfoStreamFull
	Result() (XInfoStreamFull, error)
}

type xInfoStreamFullCmd struct {
	baseCmd[XInfoStreamFull]
}

// 注：newXInfoStreamFullCmd 工厂已废弃。Pipeliner 路径通过 cmd_gen.go 直接构造空对象。

func (c *xInfoStreamFullCmd) from(res rueidis.RedisResult) {
	kv, err := res.AsMap()
	if err != nil {
		c.SetErr(err)
		return
	}
	var val XInfoStreamFull
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
	if v, ok := kv["entries-added"]; ok {
		val.EntriesAdded, _ = v.AsInt64()
	}
	if v, ok := kv["max-deleted-entry-id"]; ok {
		val.MaxDeletedEntryID, _ = v.ToString()
	}
	if v, ok := kv["recorded-first-entry-id"]; ok {
		val.RecordedFirstEntryID, _ = v.ToString()
	}
	if v, ok := kv["groups"]; ok {
		val.Groups, err = readStreamGroups(v)
		if err != nil {
			c.SetErr(err)
			return
		}
	}
	if v, ok := kv["entries"]; ok {
		ranges, err := v.AsXRange()
		if err != nil {
			c.SetErr(err)
			return
		}
		val.Entries = make([]XMessage, 0, len(ranges))
		for _, r := range ranges {
			val.Entries = append(val.Entries, newXMessage(r))
		}
	}
	c.SetVal(val)
}

func readStreamGroups(res rueidis.RedisMessage) ([]XInfoStreamGroup, error) {
	arr, err := res.ToArray()
	if err != nil {
		return nil, err
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
		if attr, ok := info["entries-read"]; ok {
			group.EntriesRead, _ = attr.AsInt64()
		}
		if attr, ok := info["lag"]; ok {
			group.Lag, _ = attr.AsInt64()
		}
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
	arr, err := res.ToArray()
	if err != nil {
		return nil, err
	}
	pending := make([]XInfoStreamGroupPending, 0, len(arr))
	for _, v := range arr {
		info, err := v.ToArray()
		if err != nil {
			return nil, err
		}
		if len(info) < 4 {
			return nil, fmt.Errorf("got %d, wanted 4", len(info))
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
		delivery, err := info[2].AsInt64()
		if err != nil {
			return nil, err
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
	arr, err := res.ToArray()
	if err != nil {
		return nil, err
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
			pending, err := attr.ToArray()
			if err != nil {
				return nil, err
			}
			c.Pending = make([]XInfoStreamConsumerPending, 0, len(pending))
			for _, v := range pending {
				pendingInfo, err := v.ToArray()
				if err != nil {
					return nil, err
				}
				if len(pendingInfo) < 3 {
					return nil, fmt.Errorf("got %d, wanted 3", len(pendingInfo))
				}
				var p XInfoStreamConsumerPending
				p.ID, err = pendingInfo[0].ToString()
				if err != nil {
					return nil, err
				}
				delivery, err := pendingInfo[1].AsInt64()
				if err != nil {
					return nil, err
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

type XPendingCmd interface {
	BaseCmd
	Val() XPending
	Result() (XPending, error)
}

type xPendingCmd struct {
	baseCmd[XPending]
}

// 注：newXPendingCmd 工厂已废弃。Pipeliner 路径通过 cmd_gen.go 直接构造空对象。

func (c *xPendingCmd) from(res rueidis.RedisResult) {
	arr, err := res.ToArray()
	if err != nil {
		c.SetErr(err)
		return
	}
	if len(arr) < 4 {
		c.SetErr(fmt.Errorf("got %d, wanted 4", len(arr)))
		return
	}
	count, err := arr[0].AsInt64()
	if err != nil {
		c.SetErr(err)
		return
	}
	lower, err := arr[1].ToString()
	if err != nil {
		c.SetErr(err)
		return
	}
	higher, err := arr[2].ToString()
	if err != nil {
		c.SetErr(err)
		return
	}
	val := XPending{
		Count:  count,
		Lower:  lower,
		Higher: higher,
	}
	consumerArr, err := arr[3].ToArray()
	if err != nil {
		c.SetErr(err)
		return
	}
	for _, v := range consumerArr {
		consumer, err := v.ToArray()
		if err != nil {
			c.SetErr(err)
			return
		}
		if len(consumer) < 2 {
			c.SetErr(fmt.Errorf("got %d, wanted 2", len(arr)))
			return
		}
		consumerName, err := consumer[0].ToString()
		if err != nil {
			c.SetErr(err)
			return
		}
		consumerPending, err := consumer[1].AsInt64()
		if err != nil {
			c.SetErr(err)
			return
		}
		if val.Consumers == nil {
			val.Consumers = make(map[string]int64)
		}
		val.Consumers[consumerName] = consumerPending
	}
	c.SetVal(val)
}

type XPendingExtCmd interface {
	BaseCmd
	Val() []XPendingExt
	Result() ([]XPendingExt, error)
}

type xPendingExtCmd struct {
	baseCmd[[]XPendingExt]
}

// 注：newXPendingExtCmd 工厂已废弃。Pipeliner 路径通过 cmd_gen.go 直接构造空对象。

func (c *xPendingExtCmd) from(res rueidis.RedisResult) {
	arrs, err := res.ToArray()
	if err != nil {
		c.SetErr(err)
		return
	}
	val := make([]XPendingExt, 0, len(arrs))
	for _, v := range arrs {
		arr, err := v.ToArray()
		if err != nil {
			c.SetErr(err)
			return
		}
		if len(arr) < 4 {
			c.SetErr(fmt.Errorf("got %d, wanted 4", len(arr)))
			return
		}
		id, err := arr[0].ToString()
		if err != nil {
			c.SetErr(err)
			return
		}
		consumer, err := arr[1].ToString()
		if err != nil {
			c.SetErr(err)
			return
		}
		idle, err := arr[2].AsInt64()
		if err != nil {
			c.SetErr(err)
			return
		}
		retryCount, err := arr[3].AsInt64()
		if err != nil {
			c.SetErr(err)
			return
		}
		val = append(val, XPendingExt{
			ID:         id,
			Consumer:   consumer,
			Idle:       time.Duration(idle) * time.Millisecond,
			RetryCount: retryCount,
		})
	}
	c.SetVal(val)
}

type XStreamSliceCmd interface {
	BaseCmd
	Val() []XStream
	Result() ([]XStream, error)
}

type XAutoClaimJustIDCmd interface {
	BaseCmd
	Val() (ids []string, start string)
	Result() (ids []string, start string, err error)
}

type xAutoClaimJustIDCmd struct {
	err   error
	start string
	val   []string
}

// 注：newXAutoClaimJustIDCmd 工厂已废弃。Pipeliner 路径通过 cmd_gen.go 直接构造空对象。

func (c *xAutoClaimJustIDCmd) from(res rueidis.RedisResult) {
	arr, err := res.ToArray()
	if err != nil {
		c.err = err
		return
	}
	if len(arr) < 2 {
		c.err = fmt.Errorf("got %d, wanted 2", len(arr))
		return
	}
	start, err := arr[0].ToString()
	if err != nil {
		c.err = err
		return
	}
	val, err := arr[1].AsStrSlice()
	if err != nil {
		c.err = err
		return
	}
	c.err = err
	c.val = val
	c.start = start
}

func (c *xAutoClaimJustIDCmd) SetVal(val []string, start string) {
	c.val = val
	c.start = start
}

func (c *xAutoClaimJustIDCmd) SetErr(err error) { c.err = err }
func (c *xAutoClaimJustIDCmd) Val() (ids []string, start string) {
	return c.val, c.start
}
func (c *xAutoClaimJustIDCmd) Err() error {
	return c.err
}
func (c *xAutoClaimJustIDCmd) Result() (ids []string, start string, err error) {
	return c.val, c.start, c.err
}

func geoRadiusQueryArgs(q GeoRadiusQuery) []string {
	args := make([]string, 0, 2)
	args = append(args, strconv.FormatFloat(q.Radius, 'f', -1, 64))
	if q.Unit != "" {
		args = append(args, q.Unit)
	} else {
		args = append(args, KM)
	}
	if q.WithCoord {
		args = append(args, KwWithCoord)
	}
	if q.WithDist {
		args = append(args, KwWithDist)
	}
	if q.WithGeoHash {
		args = append(args, KwWithHash)
	}
	if q.Count > 0 {
		args = append(args, KwCount, strconv.FormatInt(q.Count, 10))
	}
	if q.Sort != "" {
		args = append(args, q.Sort)
	}
	if q.Store != "" {
		args = append(args, KwStore)
		args = append(args, q.Store)
	}
	if q.StoreDist != "" {
		args = append(args, KwStoreDist)
		args = append(args, q.StoreDist)
	}
	return args
}

func geoSearchLocationQueryArgs(q GeoSearchLocationQuery) []string {
	args := geoSearchQueryArgs(q.GeoSearchQuery)
	if q.WithCoord {
		args = append(args, KwWithCoord)
	}
	if q.WithDist {
		args = append(args, KwWithDist)
	}
	if q.WithHash {
		args = append(args, KwWithHash)
	}
	return args
}

func geoSearchQueryArgs(q GeoSearchQuery) []string {
	args := make([]string, 0, 2)
	if q.Member != "" {
		args = append(args, KwFromMember, q.Member)
	} else {
		args = append(args, KwFromLonLat, strconv.FormatFloat(q.Longitude, 'f', -1, 64), strconv.FormatFloat(q.Latitude, 'f', -1, 64))
	}
	if q.Radius > 0 {
		if q.RadiusUnit == "" {
			q.RadiusUnit = KM
		}
		args = append(args, KwByRadius, strconv.FormatFloat(q.Radius, 'f', -1, 64), q.RadiusUnit)
	} else {
		if q.BoxUnit == "" {
			q.BoxUnit = KM
		}
		args = append(args, KwByBox, strconv.FormatFloat(q.BoxWidth, 'f', -1, 64), strconv.FormatFloat(q.BoxHeight, 'f', -1, 64), q.BoxUnit)
	}
	if q.Sort != "" {
		args = append(args, q.Sort)
	}
	if q.Count > 0 {
		args = append(args, KwCount, strconv.FormatInt(q.Count, 10))
		if q.CountAny {
			args = append(args, KwAny)
		}
	}
	return args
}
