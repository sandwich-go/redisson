//go:build miniredis_test

package redisson

import (
	"context"
	"testing"

	"github.com/redis/rueidis/rueidiscompat"
)

// cmd_script_retryable_unit_test.go 用 miniredis 起 client，验证 CreateScript-
// Retryable / CreateScriptWithNameRetryable 的 script 实例:
//
//  1. Hash 与非 retryable 实例一致（EVALSHA 依赖此不变式）；
//  2. Eval / EvalRO / EvalSha / EvalShaRO / Run / RunRO 返回的 Cmd **必须是**
//     redisson 本包的 *anyCmd（走 rueidis 原生 + ToRetryable 路径的产物），
//     而不是 *rueidiscompat.Cmd（adapter 路径的产物）。
//
// 这层类型断言是 "retryable 开关确实接通" 的唯一直接证据:e2e 测试无法区分
// retryable 与非 retryable（都能返回正确值）。将来若不小心把 s.retryable 传
// 丢或分支写反，本测试会立即失败。
//
// 本文件不带 //go:build integration:走 miniredis，local go test 直接可跑
// （需 -tags 'miniredis_test redisson_miniredis'）。

// freshMiniredisClient 起一个独立 miniredis client，返回 *client。t.Cleanup
// 关闭。风格对齐 builder_unit_test.go 的 freshBuilderClient。
func freshMiniredisClient(t *testing.T) *client {
	t.Helper()
	c := MustNewClient(NewConf(
		WithDevelopment(false),
		WithT(t),
		WithEnableCache(false),
	))
	t.Cleanup(func() { _ = c.Close() })
	return c.(*client)
}

// assertAnyCmd 断言 cmd 是 *anyCmd 而不是 *rueidiscompat.Cmd。
func assertAnyCmd(t *testing.T, tag string, cmd Cmd) {
	t.Helper()
	if _, ok := cmd.(*anyCmd); !ok {
		if _, isCompat := cmd.(*rueidiscompat.Cmd); isCompat {
			t.Fatalf("%s: retryable 路径应产出 *anyCmd,实际拿到 *rueidiscompat.Cmd（说明 retryable 开关未接通）", tag)
		}
		t.Fatalf("%s: 期望 *anyCmd,实际类型 %T", tag, cmd)
	}
}

// assertCompatCmd 断言 cmd 是 *rueidiscompat.Cmd(非 retryable 路径的产物)。
// 用于对照:证明 CreateScript(非 retryable) 走的仍是 adapter 路径。
func assertCompatCmd(t *testing.T, tag string, cmd Cmd) {
	t.Helper()
	if _, ok := cmd.(*rueidiscompat.Cmd); !ok {
		t.Fatalf("%s: 非 retryable 路径应产出 *rueidiscompat.Cmd,实际类型 %T", tag, cmd)
	}
}

func TestCreateScriptRetryable_HashEqualsNonRetryable(t *testing.T) {
	c := freshMiniredisClient(t)

	sNormal := c.CreateScript(testRetryableScript)
	sRetry := c.CreateScriptRetryable(testRetryableScript)

	if sRetry.Hash() != sNormal.Hash() {
		t.Fatalf("Hash 不一致：retryable=%q normal=%q(EVALSHA 会因 SHA 不匹配失效)",
			sRetry.Hash(), sNormal.Hash())
	}
}

func TestCreateScriptWithNameRetryable_HashEqualsNonRetryable(t *testing.T) {
	c := freshMiniredisClient(t)

	sNormal := c.CreateScriptWithName("named", testRetryableScript)
	sRetry := c.CreateScriptWithNameRetryable("named", testRetryableScript)

	if sRetry.Hash() != sNormal.Hash() {
		t.Fatalf("Hash 不一致: retryable=%q normal=%q", sRetry.Hash(), sNormal.Hash())
	}
}

// TestCreateScriptRetryable_TypeAssertion 逐条验证 retryable 实例的 6 条执行
// 路径都产出 *anyCmd。若将来某条路径漏传 s.retryable(比如把 Run 里的 fallback
// 分支写死成 false),本测试对应断言会失败。
func TestCreateScriptRetryable_TypeAssertion(t *testing.T) {
	c := freshMiniredisClient(t)
	ctx := context.Background()
	s := c.CreateScriptRetryable(testRetryableScript)

	// Eval:首发就走 EVAL,直接 assert anyCmd
	assertAnyCmd(t, "Eval", s.Eval(ctx, nil, "eval-v"))

	// EvalRO
	assertAnyCmd(t, "EvalRO", s.EvalRO(ctx, nil, "evalro-v"))

	// EvalSha:上一步 Eval 已经把脚本 cache 到 server，EVALSHA 应命中
	assertAnyCmd(t, "EvalSha", s.EvalSha(ctx, nil, "evalsha-v"))

	// EvalShaRO
	assertAnyCmd(t, "EvalShaRO", s.EvalShaRO(ctx, nil, "evalsharo-v"))

	// Run: 已缓存,evalSha 直接命中 → 走的仍是 retryable 的 evalSha 分支
	assertAnyCmd(t, "Run(cached)", s.Run(ctx, nil, "run-v"))

	// Run(fallback):flush 让 EVALSHA 报 NOSCRIPT → fallback 到 eval,fallback
	// 分支也必须走 retryable。
	// 注意:不能用 c.ScriptFlush(ctx) — rueidiscompat 里它走 doStringCmdPrimaries,
	// 内部先 ROLE 定位 primary,miniredis 不支持 ROLE 会导致 flush 无声失败。
	// 直接用 c.cmd.Do 发原生 SCRIPT FLUSH 绕过。
	flushMiniredisScripts(t, c)
	assertAnyCmd(t, "Run(fallback)", s.Run(ctx, nil, "run-fallback-v"))

	// RunRO: fallback 分支的 Run 已经把脚本 cache 回去,直接跑 RunRO cached
	assertAnyCmd(t, "RunRO(cached)", s.RunRO(ctx, nil, "runro-v"))

	// RunRO(fallback):同样用 flushMiniredisScripts 绕开 ROLE 限制。
	flushMiniredisScripts(t, c)
	assertAnyCmd(t, "RunRO(fallback)", s.RunRO(ctx, nil, "runro-fallback-v"))
}

// TestCreateScript_TypeAssertion 对照测试:非 retryable 实例的所有路径产出
// *rueidiscompat.Cmd（adapter 路径),证明本包"retryable / 非 retryable 分道"
// 的双路径设计是真实的,不是恒返回一种类型的误报。
func TestCreateScript_TypeAssertion(t *testing.T) {
	c := freshMiniredisClient(t)
	ctx := context.Background()
	s := c.CreateScript(testRetryableScript)

	assertCompatCmd(t, "Eval", s.Eval(ctx, nil, "eval-v"))
	assertCompatCmd(t, "EvalRO", s.EvalRO(ctx, nil, "evalro-v"))
	assertCompatCmd(t, "EvalSha", s.EvalSha(ctx, nil, "evalsha-v"))
	assertCompatCmd(t, "EvalShaRO", s.EvalShaRO(ctx, nil, "evalsharo-v"))
	assertCompatCmd(t, "Run(cached)", s.Run(ctx, nil, "run-v"))

	flushMiniredisScripts(t, c)
	assertCompatCmd(t, "Run(fallback)", s.Run(ctx, nil, "run-fallback-v"))

	assertCompatCmd(t, "RunRO(cached)", s.RunRO(ctx, nil, "runro-v"))

	flushMiniredisScripts(t, c)
	assertCompatCmd(t, "RunRO(fallback)", s.RunRO(ctx, nil, "runro-fallback-v"))
}

// testRetryableScript 本地脚本常量。testScript 定义在 cmd_script_test.go 里
// 但那份带 integration tag,与本文件的 miniredis_test tag 互斥,重定义一份即可。
const testRetryableScript = `return ARGV[1]`

// flushMiniredisScripts 用 rueidis 原生 c.cmd.Do 发 SCRIPT FLUSH,绕开
// rueidiscompat.ScriptFlush 的 ROLE 前置逻辑(miniredis 不支持 ROLE)。
// 用于测试 Run/RunRO 的 EVALSHA→NOSCRIPT fallback 分支。
//
// flush 后用 SCRIPT EXISTS 校验生效:若 rueidis 或 miniredis 未来行为变化
// 让 flush 又静默失败,这里能立即抓到,不至于让 fallback 测试假通过。
func flushMiniredisScripts(t *testing.T, c *client) {
	t.Helper()
	ctx := context.Background()
	res := c.cmd.Do(ctx, c.cmd.B().ScriptFlush().Build())
	if err := res.Error(); err != nil {
		t.Fatalf("SCRIPT FLUSH failed: %v", err)
	}
	// 用一个新建 script 拿到 SHA,验证 flush 后 cache 里应查不到
	probe := c.CreateScript(testRetryableScript)
	existsRes := c.cmd.Do(ctx, c.cmd.B().ScriptExists().Sha1(probe.Hash()).Build())
	arr, err := existsRes.ToArray()
	if err != nil {
		t.Fatalf("SCRIPT EXISTS after flush: %v", err)
	}
	if len(arr) == 1 {
		if v, _ := arr[0].ToInt64(); v != 0 {
			t.Fatalf("SCRIPT EXISTS after flush = 1, 脚本仍在 cache;flush 未生效")
		}
	}
}
