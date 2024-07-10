package redisson

import (
	"context"
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
	IsCluster() bool
	Options() ConfVisitor
	ForEachNodes(context.Context, func(context.Context, Cmdable) error) error
	Receive(ctx context.Context, cb func(Message), channels ...string) error
	PReceive(ctx context.Context, cb func(Message), patterns ...string) error

	// XMGet XMGet，类似 MGet 函数，内部会自动按相同 slot 执行 MGet 命令
	XMGet(ctx context.Context, keys ...string) SliceCmd

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
