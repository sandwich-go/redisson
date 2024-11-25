package redisson

import (
	"context"
	"github.com/coreos/go-semver/semver"
	"github.com/sandwich-go/funnel"
	"time"
)

type XCmdable interface {
	SafeCmdable
	RegisterCollector(RegisterCollectorFunc)
	Cache(ttl time.Duration) CacheCmdable
	NewLocker(opts ...LockerOption) (Locker, error)
	NewFunnel(key string, capacity, operations int64, seconds time.Duration) funnel.Funnel
	NewBloomFilter(name string, expectedNumberOfItems uint, falsePositiveRate float64, opts ...BloomOption) (BloomFilter, error)
	NewDelayQueue(name string, f func([]byte) error, opts ...DelayOption) (DelayQueue, error)
	Close() error
	IsCluster() bool
	Options() ConfVisitor
	ForEachNodes(context.Context, func(context.Context, Cmdable) error) error
	Receive(ctx context.Context, cb func(Message), channels ...string) error
	PReceive(ctx context.Context, cb func(Message), patterns ...string) error
	Do(ctx context.Context, completed Completed) RedisResult
	Version() *semver.Version
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
