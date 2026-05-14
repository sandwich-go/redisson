//go:build !integration && miniredis_test

package redisson

// realRedisAvailable=false：仅 miniredis_test 构建。doTestUnits 等
// 依赖真实 Redis 的 helper 会在每个 test 入口 t.Skip。
//
// 仍编译进 cmd_*_test.go 是为了让 miniredis_units_test.go 能引用其中的
// testXxx 单元测试函数（共享 fixture）。详见 realredis_enabled_test.go 注释。
const realRedisAvailable = false
