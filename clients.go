package redisson

// 本文件提供细粒度 Client 类型别名，便于调用方在签名中表达更窄的"我只需要 X 命令族"。
//
// 这是 v2.0 引入的渐进式 API：
//   - 现有 Cmdable 仍是聚合大接口，向后兼容；
//   - 新代码可以用 StringClient/HashClient/... 收窄依赖面，符合接口隔离原则；
//   - 由于 Go 类型别名（=）的特性，*client 自动满足所有这些类型，无需新方法。
//
// 用法示例：
//
//	func DoSomething(s redisson.StringClient) error {
//	    s.Set(ctx, "k", "v", 0)
//	    // 这里只能调用 String 命令族，编译期保证调用粒度
//	}
//
//	// 调用：DoSomething(c) 直接传入 Cmdable 即可
//
// 命令族对应关系（与 cmd_*.go / builder_*.go 一致）:

// StringClient String 命令子集(GET/SET/INCR/...)。
type StringClient = StringCmdable

// HashClient Hash 命令子集(HGET/HSET/HMGET/...)。
type HashClient = HashCmdable

// ListClient List 命令子集(LPUSH/LRANGE/BLMPOP/...)。
type ListClient = ListCmdable

// SetClient Set 命令子集(SADD/SMEMBERS/SDIFF/...)。
type SetClient = SetCmdable

// SortedSetClient SortedSet 命令子集(ZADD/ZRANGE/...)。
type SortedSetClient = SortedSetCmdable

// StreamClient Stream 命令子集(XADD/XRANGE/XGROUP/...)。
type StreamClient = StreamCmdable

// BitmapClient Bitmap 命令子集(GETBIT/SETBIT/BITCOUNT/...)。
type BitmapClient = BitmapCmdable

// GeoClient Geospatial 命令子集(GEOADD/GEOSEARCH/...)。
type GeoClient = GeospatialCmdable

// HyperLogClient HyperLogLog 命令子集(PFADD/PFCOUNT/...)。
type HyperLogClient = HyperLogCmdable

// ScriptClient Lua Script + Function 命令子集(EVAL/EVALSHA/FCALL/...)。
type ScriptClient = ScriptCmdable

// ServerClient Server 管理命令子集(INFO/DEBUG/MEMORY/...)。
type ServerClient = ServerCmdable

// ClusterClient Cluster 管理命令子集(CLUSTER NODES/SLOTS/...)。
type ClusterClient = ClusterCmdable

// ConnectionClient Connection 命令子集(PING/AUTH/CLIENT KILL/...)。
type ConnectionClient = ConnectionCmdable

// PubSubClient Pub/Sub 命令子集(PUBLISH/SUBSCRIBE/...)。
type PubSubClient = PubSubCmdable

// GenericClient 通用 key 命令子集(DEL/EXPIRE/SCAN/...)。
type GenericClient = GenericCmdable

// PipelineClient Pipeline 入口（Pipeline()）。
type PipelineClient = PipelineCmdable
