package redisson

//go:generate go run ./cmd/genbuilder -check
//go:generate go run ./cmd/genbuilder -check=false -export-spec specs/builder.yaml

// builder 是 rueidis Builder 的语义糖,所有命令构造器 (XxxCompleted) 方法按命令族
// 分布在 builder_<class>.go 文件中,与 cmd_gen_<class>.go 一一对应:
//
//	builder_bitmap.go        位图 (GetBit/SetBit/BitCount/BitPos/BitField/BitOp...)
//	builder_cluster_conn.go  Cluster + Connection (Ping/Echo/ClientList/ClusterReplicas)
//	builder_generic.go       通用 key 操作 (Del/Exists/Expire/Sort/Scan/Restore...)
//	builder_geospatial.go    地理空间 (Geo*)
//	builder_hash.go          Hash (H*)
//	builder_hyperlog.go      HyperLogLog (PFAdd/PFCount/PFMerge)
//	builder_list.go          List (L*/R*/B*)
//	builder_pubsub.go        Pub/Sub (Publish/SPublish/PubSub*)
//	builder_script.go        Script + Function + FCall + ACL
//	builder_server.go        Server / Admin (Command/Config/Info/Debug/Memory/Time...)
//	builder_set.go           Set (S*)
//	builder_sortedset.go     SortedSet (Z*)
//	builder_stream.go        Stream (X*)
//	builder_string.go        String (Get/Set/Incr/Decr/Append...)
type builder struct {
	Builder
}
