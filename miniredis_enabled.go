//go:build redisson_miniredis

package redisson

import (
	"github.com/alicebob/miniredis/v2"
)

// setupMiniredisIfEnabled 在配置了 Tester 时启动一个 miniredis 实例并把
// 客户端的 Addrs 指向它。
//
// 仅在 build tag `redisson_miniredis` 下编译；正常构建（如线上服务）
// 不会链入 miniredis，避免增加二进制体积与攻击面。
func setupMiniredisIfEnabled(v ConfInterface, t Tester) error {
	mr := miniredis.RunT(t)
	_ = v.ApplyOption(WithAddrs(mr.Addr()))
	return nil
}
