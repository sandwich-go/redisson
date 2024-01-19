package redisson

import (
	"github.com/sandwich-go/funnel"
	"time"
)

type Cmdable interface {
	Cache(ttl time.Duration) CacheCmdable
	NewLocker(opts ...LockerOption) (Locker, error)
	NewFunnel(key string, capacity, operations int64, seconds time.Duration) funnel.Funnel
	PoolStats() PoolStats
	RegisterCollector(RegisterCollectorFunc)
	Close() error
	RawCmdable() interface{}

	CacheCmdable
	BitmapCmdable
	ClusterCmdable
	ConnectionCmdable
	GenericCmdable
	GeospatialCmdable
	HashCmdable
	HyperLogCmdable
	ListCmdable
	ScriptCmdable
	ServerCmdable
	SetCmdable
	SortedSetCmdable
	StreamCmdable
	StringCmdable
	PubSubCmdable
	PipelineCmdable
}

type CacheCmdable interface {
	BitmapCacheCmdable
	GenericCacheCmdable
	GeospatialCacheCmdable
	HashCacheCmdable
	ListCacheCmdable
	SetCacheCmdable
	SortedSetCacheCmdable
	StringCacheCmdable
}
