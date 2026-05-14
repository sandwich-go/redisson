package redisson

import (
	"context"
	"sort"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

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
// 当非 Development 模式时额外做一次 FlushAll，确保 unit 之间无残留。
func doTestUnitClean(ctx context.Context, c Cmdable, keys []string) {
	if len(keys) > 0 {
		So(c.Del(ctx, keys...).Err(), ShouldBeNil)
	}
	if !c.Options().GetDevelopment() {
		c.FlushAll(context.Background())
	}
}

func _doTestUnits(t *testing.T, c Cmdable, unitsFunc func() []TestUnit) {
	t.Cleanup(func() {
		_ = c.Close()
	})
	if !c.Options().GetDevelopment() {
		c.FlushAll(context.Background())
	}
	var ctx = context.Background()
	for _, v := range unitsFunc() {
		Convey(v.Name.String(), t, func() { doTestUnitClean(ctx, c, v.Func(ctx, c)) })
	}
}

// doTestUnits 用默认 standalone 配置（Development=false）跑测试。
func doTestUnits(t *testing.T, unitsFunc func() []TestUnit) {
	c := MustNewClient(NewConf(WithDevelopment(false)))
	_doTestUnits(t, c, unitsFunc)
}

// doClusterTestUnits 模拟 cluster 路径（开启 Development）。
func doClusterTestUnits(t *testing.T, unitsFunc func() []TestUnit) {
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
