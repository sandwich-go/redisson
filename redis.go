package redisson

import (
	"context"
	"github.com/coreos/go-semver/semver"
	"time"
)

type XCmdable interface {
	SafeCmdable
	RegisterCollector(RegisterCollectorFunc)
	Cache(ttl time.Duration) CacheCmdable
	NewLocker(opts ...LockerOption) (Locker, error)
	NewBloomFilter(name string, expectedNumberOfItems uint, falsePositiveRate float64, opts ...BloomOption) (BloomFilter, error)
	NewRateLimiter(opts ...RateLimiterOption) (RateLimiter, error)
	NewDelayQueue(name string, f func([]byte) error, opts ...DelayOption) (DelayQueue, error)
	Close() error
	IsCluster() bool
	Options() ConfVisitor
	ForEachNodes(context.Context, func(context.Context, Cmdable) error) error
	Receive(ctx context.Context, cb func(Message), channels ...string) error
	PReceive(ctx context.Context, cb func(Message), patterns ...string) error
	SReceive(ctx context.Context, cb func(Message), channels ...string) error
	Do(ctx context.Context, completed Completed) RedisResult
	Version() *semver.Version
	// NewVirtualPubSubHub 创建一个进程级 PubSub 多路复用 Hub。详见 cmd_pubsub_virtual.go 文档注释。
	// 通过 Hub 创建的 VirtualPubSub 实例会共享底层少量真实 dedicated 连接，
	// 解决"大量 topic 各自 Subscribe 占用连接超出上限"的问题。
	NewVirtualPubSubHub() VirtualPubSubHub
	// NewVirtualStreamHub 创建一个进程级 Stream 多路复用 Hub。详见 cmd_stream_virtual.go 文档注释。
	// 通过 Hub 创建的 VirtualStream 实例会共享底层少量 blocking 连接做 XREAD 合并读取，
	// 解决"大量 stream 各起 goroutine 占用 BlockingPool 连接超上限"的问题。
	NewVirtualStreamHub() VirtualStreamHub
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
