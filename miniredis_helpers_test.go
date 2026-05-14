//go:build miniredis_test

package redisson

// miniredis_helpers_test.go 提供 miniredis 隔离套测试 helper.
//
// 与 helpers_test.go 的真实 Redis 路径并存，通过 build tag `miniredis_test` 区分。
// 必须同时启用 `redisson_miniredis` tag 才能让 setupMiniredisIfEnabled 工作:
//
//   go test -tags 'miniredis_test redisson_miniredis' ./...
//
// 每个测试一个独立的 miniredis 实例，无并行冲突，无 flaky。

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// doMiniredisTestUnits 用 miniredis 运行测试套，每个测试独立实例。
//
// 与 doTestUnits 不同：
//   - 不分配 logical DB（每个 test 自己一个 miniredis 实例，天然隔离）；
//   - 仍调用 t.Parallel()，允许并行；
//   - WithT(t) 让 connect.go 自动启动 miniredis（依赖 redisson_miniredis tag）；
//   - 显式 WithEnableCache(false)：miniredis 不支持 client side cache。
func doMiniredisTestUnits(t *testing.T, unitsFunc func() []TestUnit) {
	t.Parallel()
	c := MustNewClient(NewConf(
		WithDevelopment(false),
		WithT(t),
		WithEnableCache(false),
	))
	t.Cleanup(func() { _ = c.Close() })

	ctx := context.Background()
	for _, v := range unitsFunc() {
		v := v
		Convey(v.Name.String(), t, func() {
			doTestUnitClean(ctx, c, v.Func(ctx, c))
		})
	}
}
