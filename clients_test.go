package redisson

// 静态接口断言：确保 type alias 能让 Cmdable 自动满足细粒度 Client 接口。
// 这是编译期 nil-pointer 断言，零运行开销，任意一项失败将导致编译错误。
//
//nolint:unused // 这些只用于编译期类型检查
var (
	_ StringClient     = (Cmdable)(nil)
	_ HashClient       = (Cmdable)(nil)
	_ ListClient       = (Cmdable)(nil)
	_ SetClient        = (Cmdable)(nil)
	_ SortedSetClient  = (Cmdable)(nil)
	_ StreamClient     = (Cmdable)(nil)
	_ BitmapClient     = (Cmdable)(nil)
	_ GeoClient        = (Cmdable)(nil)
	_ HyperLogClient   = (Cmdable)(nil)
	_ ScriptClient     = (Cmdable)(nil)
	_ ServerClient     = (Cmdable)(nil)
	_ ClusterClient    = (Cmdable)(nil)
	_ ConnectionClient = (Cmdable)(nil)
	_ PubSubClient     = (Cmdable)(nil)
	_ GenericClient    = (Cmdable)(nil)
	_ PipelineClient   = (Cmdable)(nil)
)
