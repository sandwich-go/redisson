package redisson

//go:generate go run ./cmd/genbuilder -check
//go:generate go run ./cmd/genbuilder -check=false -export-spec specs/builder.yaml

// builder 是 rueidis Builder 的语义糖，所有命令构造器（XxxCompleted）方法已按
// 命令族拆分到 builder_*.go 文件，对应关系：
//
//   builder_bitmap.go        位图命令（GetBit/SetBit/BitCount/BitPos/BitField/BitOp...）
//   builder_cluster_conn.go  Cluster + Connection（Ping/Echo/ClientList/ClusterReplicas）
//   builder_generic.go       通用 key 操作（Del/Exists/Expire/Sort/Scan/Restore...）
//   builder_geospatial.go    地理空间（Geo*）
//   builder_hash.go          Hash（H*）
//   builder_hyperlog.go      HyperLogLog（PFAdd/PFCount/PFMerge）
//   builder_list.go          List（L*/R*/B*）
//   builder_pubsub.go        Pub/Sub（Publish/SPublish/PubSub*）
//   builder_script.go        Script + Function + FCall + ACL
//   builder_server.go        Server / Admin（Command/Config/Info/Debug/Memory/Time...）
//   builder_set.go           Set（S*）
//   builder_sortedset.go     SortedSet（Z*）
//   builder_stream.go        Stream（X*）
//   builder_string.go        String（Get/Set/Incr/Decr/Append...）
//
// 该拆分不改变行为：所有方法仍然挂载在同一 builder 类型上。
type builder struct {
	Builder
}
