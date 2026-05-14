//go:build integration

package redisson

// realRedisAvailable 标记测试是否能连真实 Redis。
//
// 通过 build tag 拆分实现：
//   - integration tag        → true（CI 提供 Redis 服务）
//   - 仅 miniredis_test tag  → false（无 Redis，cmd_*_test 套需要 t.Skip）
//
// 这样 cmd_*_test.go 的顶级 TestClient_* 函数仍能被 miniredis_test 构建编译
// （miniredis_units_test.go 引用了同文件内的 testXxx helper），但运行时
// 在没有真实 Redis 的环境下被 doTestUnits / doClusterTestUnits 自动跳过。
const realRedisAvailable = true
