package redisson

import (
	"github.com/redis/rueidis"
	"github.com/redis/rueidis/rueidiscompat"
)

const (
	OK      = "OK"
	EMPTY   = ""
	BYTE    = "BYTE"
	BIT     = "BIT"
	M       = "M"
	KM      = "KM"
	FT      = "FT"
	MI      = "MI"
	XX      = "XX"
	NX      = "NX"
	BEFORE  = "BEFORE"
	AFTER   = "AFTER"
	RIGHT   = "RIGHT"
	LEFT    = "LEFT"
	LADDR   = "LADDR"
	TYPE    = "TYPE"
	ASC     = "ASC"
	DESC    = "DESC"
	KeepTTL = rueidiscompat.KeepTTL
)

type (
	KeyValue                   = rueidiscompat.KeyValue
	CommandInfo                = rueidiscompat.CommandInfo
	SetArgs                    = rueidiscompat.SetArgs
	Sort                       = rueidiscompat.Sort
	RankScore                  = rueidiscompat.RankScore
	LPosArgs                   = rueidiscompat.LPosArgs
	Z                          = rueidiscompat.Z
	ZStore                     = rueidiscompat.ZStore
	ZAddArgs                   = rueidiscompat.ZAddArgs
	ZWithKey                   = rueidiscompat.ZWithKey
	ZRangeBy                   = rueidiscompat.ZRangeBy
	ZRangeArgs                 = rueidiscompat.ZRangeArgs
	XMessage                   = rueidiscompat.XMessage
	XInfoConsumer              = rueidiscompat.XInfoConsumer
	XInfoGroup                 = rueidiscompat.XInfoGroup
	XInfoStream                = rueidiscompat.XInfoStream
	XInfoStreamFull            = rueidiscompat.XInfoStreamFull
	XInfoStreamGroup           = rueidiscompat.XInfoStreamGroup
	XInfoStreamGroupPending    = rueidiscompat.XInfoStreamGroupPending
	XInfoStreamConsumer        = rueidiscompat.XInfoStreamConsumer
	XInfoStreamConsumerPending = rueidiscompat.XInfoStreamConsumerPending
	XPending                   = rueidiscompat.XPending
	XPendingExt                = rueidiscompat.XPendingExt
	XStream                    = rueidiscompat.XStream
	XAddArgs                   = rueidiscompat.XAddArgs
	XAutoClaimArgs             = rueidiscompat.XAutoClaimArgs
	XClaimArgs                 = rueidiscompat.XClaimArgs
	XPendingExtArgs            = rueidiscompat.XPendingExtArgs
	XReadArgs                  = rueidiscompat.XReadArgs
	XReadGroupArgs             = rueidiscompat.XReadGroupArgs
	BitCount                   = rueidiscompat.BitCount
	GeoPos                     = rueidiscompat.GeoPos
	GeoLocation                = rueidiscompat.GeoLocation
	GeoSearchQuery             = rueidiscompat.GeoSearchQuery
	GeoSearchLocationQuery     = rueidiscompat.GeoSearchLocationQuery
	GeoSearchStoreQuery        = rueidiscompat.GeoSearchStoreQuery
	GeoRadiusQuery             = rueidiscompat.GeoRadiusQuery
	ClusterNode                = rueidiscompat.ClusterNode
	ClusterSlot                = rueidiscompat.ClusterSlot
	ClusterShard               = rueidiscompat.ClusterShard
	Library                    = rueidiscompat.Library
	FunctionListQuery          = rueidiscompat.FunctionListQuery
	FilterBy                   = rueidiscompat.FilterBy
	KeyFlags                   = rueidiscompat.KeyFlags
	Message                    = rueidis.PubSubMessage
	Completed                  = rueidis.Completed
	Builder                    = rueidis.Builder
	RedisResult                = rueidis.RedisResult
	KeyValues                  = rueidis.KeyValues
)

type baseCmd[T any] struct {
	err error
	val T
}

func (c *baseCmd[T]) SetVal(val T) { c.val = val }
func (c *baseCmd[T]) Val() T {
	return c.val
}
func (c *baseCmd[T]) SetErr(err error) { c.err = err }
func (c *baseCmd[T]) Err() error {
	return c.err
}
func (c *baseCmd[T]) Result() (T, error) {
	return c.Val(), c.Err()
}
