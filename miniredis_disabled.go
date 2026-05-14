//go:build !redisson_miniredis

package redisson

import "fmt"

// setupMiniredisIfEnabled 在未启用 build tag `redisson_miniredis` 时，
// 配置了 WithT(t) 但未启用对应 build tag 会返回错误。
//
// 用法：编译/测试时加上 `-tags redisson_miniredis`，
// 例如 `go test -tags redisson_miniredis ./...`，即可使用内置 miniredis mock。
func setupMiniredisIfEnabled(_ ConfInterface, _ Tester) error {
	return fmt.Errorf("redisson: WithT(t) requires building with `-tags redisson_miniredis`; " +
		"otherwise miniredis is excluded from the binary")
}
