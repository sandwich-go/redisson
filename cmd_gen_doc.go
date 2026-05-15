package redisson

//go:generate go run ./cmd/genmeta -check

// cmd_gen_*.go 是 redisson 384 个 Redis 命令的元数据全集 (Class / RequireVersion /
// Forbid / WarnVersion / Warning / Instead / ETC) 与 Pipeliner P()/Cmd() 包装方法,
// 由 cmd/genmeta 从 specs/cmd_gen.yaml 生成,按 Class 拆分到 15 个文件 +
// cmd_gen_const.go 共享常量。
//
// 文件分布与 builder_*.go 对齐:
//
//	cmd_gen_bitmap.go     12 命令  ←→ builder_bitmap.go
//	cmd_gen_cluster.go    23 命令  ←→ builder_cluster_conn.go (含 Connection 部分)
//	cmd_gen_connection.go 14 命令  ←→ builder_cluster_conn.go (含 Cluster 部分)
//	cmd_gen_generic.go    49 命令  ←→ builder_generic.go
//	cmd_gen_geospatial.go 11 命令  ←→ builder_geospatial.go
//	cmd_gen_hash.go       44 命令  ←→ builder_hash.go
//	cmd_gen_hyperlog.go    3 命令  ←→ builder_hyperlog.go
//	cmd_gen_list.go       31 命令  ←→ builder_list.go
//	cmd_gen_pubsub.go     13 命令  ←→ builder_pubsub.go
//	cmd_gen_scripting.go  19 命令  ←→ builder_script.go
//	cmd_gen_server.go     26 命令  ←→ builder_server.go
//	cmd_gen_set.go        21 命令  ←→ builder_set.go
//	cmd_gen_sortedset.go  59 命令  ←→ builder_sortedset.go
//	cmd_gen_stream.go     34 命令  ←→ builder_stream.go
//	cmd_gen_string.go     25 命令  ←→ builder_string.go
//
// 三位一体闭环:
//
//	specs/cmd_gen.yaml   ← extract ←   cmd_gen_<class>.go 源代码
//	                     → generate →  (生成器 cmd/genmeta 实现双向)
//	cmd_gen_meta_test.go (采样 41) + cmd_gen_full_test.go (全量 384)
//	     ↑ 双向校验: yaml ↔ runtime CommandXxx ↔ specs 三者一致
//
// 工作流:
//
//   - 修改某个命令 metadata: 直接改 cmd_gen_<class>.go,
//     运行 `make cmdgen-extract` 同步 yaml,提交两侧改动。
//
//   - 新增/删除一个命令: 编辑 specs/cmd_gen.yaml,运行 `make cmdgen-generate`
//     重生成全部 cmd_gen_<class>.go 与 cmd_gen_registry_test.go (双向自动同步)。
//
//   - CI 守门: `make cmdgen-check` (集成在 `make ci`) 三重比对:
//       1. specs/cmd_gen.yaml ↔ cmd_gen_<class>.go 元数据字段一致
//       2. specs/cmd_gen.yaml ↔ cmd_gen_registry_test.go 384 项注册表一致
//       3. cmd_gen_full_test.go 全量 384 命令的 9 个 metadata 字段运行时校验
