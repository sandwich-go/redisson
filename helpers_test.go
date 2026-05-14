//go:build integration || miniredis_test

package redisson

import (
	"context"
	"sort"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

// dbAlloc 全局 DB 分配器，用于让每个 TestClient_X 落在不同的 logical DB，
// 从而可以 t.Parallel() 同时运行而互不干扰。
//
// Redis 单实例支持 16 个 db (0-15)。我们从 1 开始分配，把 0 留给手动调试。
// 当超过 15 时回卷重用（极端情况，本仓库测试 < 16 个）。
var dbAlloc atomic.Int32

// nextDB 取下一个可用 DB 编号 (1..15)。
func nextDB() int {
	n := dbAlloc.Add(1)
	return int(((n - 1) % 15) + 1)
}

// eventually 轮询直到 cond() 返回 true 或超时；用于替代 time.Sleep。
//
// 默认每 10ms 检查一次。timeout 是总等待上限。
//
// 用法:
//
//	eventually(func() bool { return r.Val() == 1 }, 2*time.Second, "value should reach 1")
func eventually(cond func() bool, timeout time.Duration, msg string) {
	if cond() {
		return
	}
	deadline := time.Now().Add(timeout)
	tick := 10 * time.Millisecond
	for time.Now().Before(deadline) {
		time.Sleep(tick)
		if cond() {
			return
		}
		// 指数退避，最多到 100ms 一次，减少 Redis 压力
		if tick < 100*time.Millisecond {
			tick *= 2
		}
	}
	So(false, ShouldBeTrue) // 转化为 Convey 断言失败
	_ = msg                 // 显式标记未使用，便于读者识别失败原因
}

// eventuallyEq 是 eventually 的便利封装：等待 actual() == expected。
func eventuallyEq[T comparable](actual func() T, expected T, timeout time.Duration) {
	eventually(func() bool { return actual() == expected }, timeout,
		"expected value not reached")
}

// eventuallyExpired 等待 key 过期（c.Get 返回 redis.Nil 错误）。
//
// timeout 应略大于 ttl + 一些时钟漂移容忍度（如 ttl + 500ms）。
func eventuallyExpired(ctx context.Context, c Cmdable, key string, timeout time.Duration) {
	eventually(func() bool {
		return IsNil(c.Get(ctx, key).Err())
	}, timeout, "key did not expire")
}

// 本文件集中存放跨 _test.go 共享的测试 helper，避免散落在各 cmd_*_test.go 中。

// TestUnitName 命令名抽象，方便用 Command 常量直接当 unit 名。
type TestUnitName interface {
	String() string
}

// TestUnit 表驱动的测试单元。
type TestUnit struct {
	Name TestUnitName
	Func func(ctx context.Context, c Cmdable) []string
}

// doTestUnitClean 在每个 unit 结束后清理写入的 key；
// 然后 FlushDB（不再 FlushAll，避免影响并行运行的其他测试 DB）。
//
// 逐个 Del 避免 rueidis 多 key 跨槽预校验（同 unit 不同语义的多 key 不一定同 hashtag）。
func doTestUnitClean(ctx context.Context, c Cmdable, keys []string) {
	for _, k := range keys {
		So(c.Del(ctx, k).Err(), ShouldBeNil)
	}
	if !c.Options().GetDevelopment() {
		c.FlushDB(context.Background())
	}
}

func _doTestUnits(t *testing.T, c Cmdable, unitsFunc func() []TestUnit) {
	t.Cleanup(func() {
		_ = c.Close()
	})
	if !c.Options().GetDevelopment() {
		c.FlushDB(context.Background())
	}
	var ctx = context.Background()
	for _, v := range unitsFunc() {
		Convey(v.Name.String(), t, func() { doTestUnitClean(ctx, c, v.Func(ctx, c)) })
	}
}

// doTestUnits 用默认 standalone 配置（Development=false）跑测试。
// 自动分配独立 DB（避免与并行运行的其他测试相互污染）。
//
// 仅在 integration tag 下真实运行；纯 miniredis_test 构建会 t.Skip
// （miniredis 套通过 doMiniredisTestUnits 跑相同 testXxx helper）。
func doTestUnits(t *testing.T, unitsFunc func() []TestUnit) {
	if !realRedisAvailable {
		t.Skip("real Redis not available; this test runs under -tags integration")
	}
	t.Parallel()
	c := MustNewClient(NewConf(WithDevelopment(false), WithDB(nextDB())))
	_doTestUnits(t, c, unitsFunc)
}

// doClusterTestUnits 模拟 cluster 路径（开启 Development）。
// 不并行：cluster 模式下使用 db 0 才有效，且测试本身依赖跨 slot 行为。
func doClusterTestUnits(t *testing.T, unitsFunc func() []TestUnit) {
	if !realRedisAvailable {
		t.Skip("real Redis not available; this test runs under -tags integration")
	}
	c := MustNewClient(NewConf(WithDevelopment(true)))
	_doTestUnits(t, c, unitsFunc)
}

// stringSliceEqual 比较两个 []string；absolute=true 时严格按下标比较，
// 否则会对两端排序后再比较（用于 Redis 命令返回顺序不稳的场景）。
func stringSliceEqual(a, b []string, absolute bool) bool {
	if a == nil && b != nil {
		return false
	}
	if b == nil && a != nil {
		return false
	}
	if len(b) != len(a) {
		return false
	}
	if !absolute {
		sort.Strings(a)
		sort.Strings(b)
	}
	for k, v := range a {
		if v != b[k] {
			return false
		}
	}
	return true
}

// interfaceSliceEqual 严格比较两个 []any（按下标）。
func interfaceSliceEqual(a, b []any) bool {
	if a == nil && b != nil {
		return false
	}
	if b == nil && a != nil {
		return false
	}
	if len(b) != len(a) {
		return false
	}
	for k, v := range a {
		if v != b[k] {
			return false
		}
	}
	return true
}

// stringStructMapEqual 比较 map[string]struct{}（用于 set 类命令断言）。
func stringStructMapEqual(a, b map[string]struct{}) bool {
	if a == nil && b != nil {
		return false
	}
	if b == nil && a != nil {
		return false
	}
	if len(b) != len(a) {
		return false
	}
	for k := range a {
		if _, ok := b[k]; !ok {
			return false
		}
	}
	return true
}
