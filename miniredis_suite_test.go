//go:build miniredis_test

package redisson

import "testing"

// miniredis 套覆盖与 miniredis 兼容的命令子集。
// 每个 Test 在独立 miniredis 进程中运行，无外部 Redis 依赖。
//
// 命令子集见 miniredis_units_test.go（不含依赖真实 TTL 的 unit）。

func TestMR_String(t *testing.T) { doMiniredisTestUnits(t, miniredisStringUnits) }
func TestMR_Hash(t *testing.T)   { doMiniredisTestUnits(t, miniredisHashUnits) }
func TestMR_Set(t *testing.T)    { doMiniredisTestUnits(t, miniredisSetUnits) }
