package redisson

import (
	"context"
	"github.com/sandwich-go/funnel"
	"time"
)

type XCmdable interface {
	Cache(ttl time.Duration) CacheCmdable
	NewLocker(opts ...LockerOption) (Locker, error)
	NewFunnel(key string, capacity, operations int64, seconds time.Duration) funnel.Funnel
	Close() error
	IsCluster() bool
	Options() ConfVisitor
	ForEachNodes(context.Context, func(context.Context, Cmdable) error) error
	Receive(ctx context.Context, cb func(Message), channels ...string) error
	PReceive(ctx context.Context, cb func(Message), patterns ...string) error

	XMGet(ctx context.Context, keys ...string) SliceCmd
}

type Cmdable interface {
	XCmdable
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
