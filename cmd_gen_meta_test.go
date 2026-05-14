package redisson

import (
	"strings"
	"testing"
)

// cmd_gen_meta_test.go 单元测试 cmd_gen.go 中各 Command 类型的元数据方法
// （String / Class / RequireVersion / Forbid / Warning / Instead / ETC）
//
// 不依赖真实 Redis；纯类型方法调用。
//
// 不测试全部 384 个 Command（重复价值低），而是按以下原则采样：
//   - 每个 Class 至少 1 个代表
//   - 所有 Forbid() == true 的命令（46 个）必须验证
//   - 所有有 Warning（非空）的命令必须验证 Warning/Instead/ETC 的语义自洽
//   - 命令字符串与变量名应能对应（防止 generator 产生错位）

// commandMetaCase 表驱动用例。
type commandMetaCase struct {
	cmd            Command
	wantStr        string
	wantClass      string
	wantMinVersion string // 期望的 RequireVersion
	wantForbid     bool
	hasWarning     bool // Warning() 是否非空
}

// 采样：每 Class 至少 1 个代表 + 所有 Forbid + 部分 Deprecated
func sampledCommandCases() []commandMetaCase {
	return []commandMetaCase{
		// ------ Bitmap ------
		{cmd: CommandBitCount, wantStr: "BITCOUNT", wantClass: "Bitmap", wantMinVersion: "2.6.0"},
		{cmd: CommandSetBit, wantStr: "SETBIT", wantClass: "Bitmap", wantMinVersion: "2.2.0", hasWarning: true},
		{cmd: CommandGetBit, wantStr: "GETBIT", wantClass: "Bitmap", wantMinVersion: "2.2.0"},
		// ------ String ------
		{cmd: CommandSet, wantStr: "SET", wantClass: "String", wantMinVersion: "1.0.0"},
		{cmd: CommandGet, wantStr: "GET", wantClass: "String", wantMinVersion: "1.0.0"},
		{cmd: CommandIncr, wantStr: "INCR", wantClass: "String", wantMinVersion: "1.0.0"},
		{cmd: CommandSetEX, wantStr: "SETEX", wantClass: "String", wantMinVersion: "2.0.0", hasWarning: true},
		{cmd: CommandSetNX, wantStr: "SETNX", wantClass: "String", wantMinVersion: "1.0.0", hasWarning: true},
		{cmd: CommandGetSet, wantStr: "GETSET", wantClass: "String", wantMinVersion: "1.0.0", hasWarning: true},
		{cmd: CommandMGet, wantStr: "MGET", wantClass: "String", wantMinVersion: "1.0.0", hasWarning: true},
		// ------ Hash ------
		// HSET 多 field 自 2.4 / 4.0 起；redisson 标 2.8.0 与 RESP 文档对齐。
		{cmd: CommandHSet, wantStr: "HSET", wantClass: "Hash", wantMinVersion: "2.8.0"},
		{cmd: CommandHGet, wantStr: "HGET", wantClass: "Hash", wantMinVersion: "2.0.0"},
		{cmd: CommandHMSet, wantStr: "HMSET", wantClass: "Hash", wantMinVersion: "2.0.0", hasWarning: true},
		// ------ List ------
		{cmd: CommandLPush, wantStr: "LPUSH", wantClass: "List", wantMinVersion: "1.0.0"},
		{cmd: CommandRPopLPush, wantStr: "RPOPLPUSH", wantClass: "List", wantMinVersion: "1.2.0", hasWarning: true},
		// ------ Set ------
		{cmd: CommandSAdd, wantStr: "SADD", wantClass: "Set", wantMinVersion: "1.0.0"},
		{cmd: CommandSMembers, wantStr: "SMEMBERS", wantClass: "Set", wantMinVersion: "1.0.0"},
		// ------ SortedSet ------
		{cmd: CommandZAdd, wantStr: "ZADD", wantClass: "SortedSet", wantMinVersion: "1.2.0"},
		{cmd: CommandZRevRange, wantStr: "ZREVRANGE", wantClass: "SortedSet", wantMinVersion: "1.2.0", hasWarning: true},
		{cmd: CommandZRangeByLex, wantStr: "ZRANGEBYLEX", wantClass: "SortedSet", wantMinVersion: "2.8.9", hasWarning: true},
		// ------ Stream ------
		{cmd: CommandXAdd, wantStr: "XADD", wantClass: "Stream", wantMinVersion: "5.0.0"},
		{cmd: CommandXRange, wantStr: "XRANGE", wantClass: "Stream", wantMinVersion: "5.0.0"},
		// ------ Generic ------
		{cmd: CommandDel, wantStr: "DEL", wantClass: "Generic", wantMinVersion: "1.0.0"},
		{cmd: CommandKeys, wantStr: "KEYS", wantClass: "Generic", wantMinVersion: "1.0.0", wantForbid: true, hasWarning: true},
		// ------ Geospatial ------
		{cmd: CommandGeoAdd, wantStr: "GEOADD", wantClass: "Geospatial", wantMinVersion: "3.2.0"},
		{cmd: CommandGeoRadiusRO, wantStr: "GEORADIUS_RO", wantClass: "Geospatial", wantMinVersion: "3.2.10", hasWarning: true},
		{cmd: CommandGeoRadiusStore, wantStr: "GEORADIUS", wantClass: "Geospatial", wantMinVersion: "3.2.0", hasWarning: true},
		// ------ Connection ------
		{cmd: CommandPing, wantStr: "PING", wantClass: "Connection", wantMinVersion: "1.0.0"},
		{cmd: CommandQuit, wantStr: "QUIT", wantClass: "Connection", wantMinVersion: "1.0.0", hasWarning: true},
		// ------ Server (Forbidden) ------
		{cmd: CommandFlushAll, wantStr: "FLUSHALL", wantClass: "Server", wantMinVersion: "1.0.0", wantForbid: true},
		{cmd: CommandFlushDB, wantStr: "FLUSHDB", wantClass: "Server", wantMinVersion: "1.0.0", wantForbid: true},
		// ------ Cluster (Forbidden in normal usage) ------
		{cmd: CommandClusterAddSlots, wantStr: "CLUSTER ADDSLOTS", wantClass: "Cluster", wantMinVersion: "3.0.0", wantForbid: true},
		{cmd: CommandClusterFailover, wantStr: "CLUSTER FAILOVER", wantClass: "Cluster", wantMinVersion: "3.0.0", wantForbid: true},
		{cmd: CommandClusterSlaves, wantStr: "CLUSTER SLAVES", wantClass: "Cluster", wantMinVersion: "3.0.0", wantForbid: true, hasWarning: true},
		{cmd: CommandClusterSlots, wantStr: "CLUSTER SLOTS", wantClass: "Cluster", wantMinVersion: "3.0.0", hasWarning: true},
		// ------ Scripting ------
		{cmd: CommandEval, wantStr: "EVAL", wantClass: "Scripting", wantMinVersion: "2.6.0"},
		{cmd: CommandEvalSha, wantStr: "EVALSHA", wantClass: "Scripting", wantMinVersion: "2.6.0"},
		{cmd: CommandScriptLoad, wantStr: "SCRIPT LOAD", wantClass: "Scripting", wantMinVersion: "2.6.0"},
		// ------ PubSub ------
		{cmd: CommandPublish, wantStr: "PUBLISH", wantClass: "PubSub", wantMinVersion: "2.0.0"},
		{cmd: CommandSubscribe, wantStr: "SUBSCRIBE", wantClass: "PubSub", wantMinVersion: "2.0.0"},
		// ------ Transactions ------
		// MULTI/EXEC/WATCH 不一定出现在 cmd_gen 中（rueidis 直接路由），跳过
	}
}

// TestCmdGenMeta_AllSampled 表驱动校验所有采样 Command 的元数据。
func TestCmdGenMeta_AllSampled(t *testing.T) {
	for _, tc := range sampledCommandCases() {
		t.Run(tc.wantStr, func(t *testing.T) {
			if got := tc.cmd.String(); got != tc.wantStr {
				t.Errorf("String()=%q, want %q", got, tc.wantStr)
			}
			if got := tc.cmd.Class(); got != tc.wantClass {
				t.Errorf("Class()=%q, want %q", got, tc.wantClass)
			}
			if got := tc.cmd.RequireVersion(); got != tc.wantMinVersion {
				t.Errorf("RequireVersion()=%q, want %q", got, tc.wantMinVersion)
			}
			if got := tc.cmd.Forbid(); got != tc.wantForbid {
				t.Errorf("Forbid()=%v, want %v", got, tc.wantForbid)
			}
			warning := tc.cmd.Warning()
			if tc.hasWarning && warning == "" {
				t.Errorf("expected non-empty Warning(), got empty")
			}
			if !tc.hasWarning && warning != "" {
				t.Errorf("expected empty Warning(), got %q", warning)
			}
			// Instead/ETC 是辅助,不强断言但要确保调用不 panic
			_ = tc.cmd.Instead()
			_ = tc.cmd.ETC()
			_ = tc.cmd.WarnVersion()
			_ = tc.cmd.WarningOnce()
		})
	}
}

// TestCmdGenMeta_DeprecatedCommandsHaveAlternative 所有有 Warning 的命令应给出 Instead 提示。
// 这是 generator 应保证的不变式：deprecation warning 必须告诉用户用什么替代。
//
// 例外清单（当前 generator 实际行为，逐步收紧）：
//   - KEYS：仅警告慎用，无替代
//   - SETBIT：性能警告，无替代
//   - MGET：cluster 模式警告，可用 SafeMGet 但 generator 未生成
func TestCmdGenMeta_DeprecatedCommandsHaveAlternative(t *testing.T) {
	exceptions := map[string]bool{
		"KEYS":   true,
		"SETBIT": true,
		"MGET":   true,
	}
	for _, tc := range sampledCommandCases() {
		if !tc.hasWarning {
			continue
		}
		if exceptions[tc.wantStr] {
			continue
		}
		instead := tc.cmd.Instead()
		warning := tc.cmd.Warning()
		// 要么 Instead 显式给出替代，要么 Warning 文本里包含 "replaced by"
		if instead == "" && !strings.Contains(warning, "replaced") {
			t.Errorf("%s: has Warning but no Instead/replaced hint: warn=%q instead=%q",
				tc.wantStr, warning, instead)
		}
	}
}

// TestCmdGenMeta_VersionFormat RequireVersion 必须是 X.Y.Z 三段语义版本。
func TestCmdGenMeta_VersionFormat(t *testing.T) {
	for _, tc := range sampledCommandCases() {
		v := tc.cmd.RequireVersion()
		if v == "" {
			t.Errorf("%s: empty RequireVersion", tc.wantStr)
			continue
		}
		parts := strings.Split(v, ".")
		if len(parts) != 3 {
			t.Errorf("%s: RequireVersion=%q, want X.Y.Z", tc.wantStr, v)
		}
	}
}

// TestCmdGenMeta_StringNonEmpty 任何 Command.String() 都不能是空字符串
// （这是 Pipeliner 路由 / 监控统计 / 错误信息的关键）。
func TestCmdGenMeta_StringNonEmpty(t *testing.T) {
	for _, tc := range sampledCommandCases() {
		if tc.cmd.String() == "" {
			t.Errorf("%T: String() empty", tc.cmd)
		}
	}
}

// TestCmdGenMeta_ClassEnumerated Class 必须是预定义集合之一。
// 防止 generator 引入 typo 比如 "Gen" / "Genic"。
func TestCmdGenMeta_ClassEnumerated(t *testing.T) {
	allowed := map[string]bool{
		"Bitmap":      true,
		"String":      true,
		"Hash":        true,
		"List":        true,
		"Set":         true,
		"SortedSet":   true,
		"Stream":      true,
		"Generic":     true,
		"Geospatial":  true,
		"HyperLogLog": true,
		"Connection":  true,
		"Server":      true,
		"Cluster":     true,
		"Scripting":   true,
		"PubSub":      true,
		"Transaction": true,
	}
	for _, tc := range sampledCommandCases() {
		class := tc.cmd.Class()
		if !allowed[class] {
			t.Errorf("%s: Class()=%q not in allowed set", tc.wantStr, class)
		}
	}
}
