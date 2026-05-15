package redisson

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// cmd_gen_full_test.go 用 specs/cmd_gen.yaml 作为 single source of truth，
// 通过反射对运行时所有 384 个 CommandXxx 全集做元数据契约校验。
//
// 与 cmd_gen_meta_test.go 的采样测试互补:
//   - cmd_gen_meta_test.go: 41 个手写代表用例，验证关键 Class / Forbid / Warning 语义
//   - cmd_gen_full_test.go (本文件): 384 个全量 yaml-driven 校验,确保 String/Class/
//     RequireVersion/Forbid/WarnVersion/Warning/WarningOnce/Instead/ETC 9 个字段
//     与 specs/cmd_gen.yaml 逐字段一致。
//
// 闭环:
//
//	specs/cmd_gen.yaml  ←→  cmd_gen_<class>.go (运行时 commandXxx)
//	                  本测试同时验证两端一致性
//
// 任一侧改动后另一侧未同步 → 测试失败。

// commandFromYAML 单条命令的 yaml 元数据(子集,只保留可对比的 metadata 字段)。
type commandFromYAML struct {
	Name           string
	CmdString      string
	Class          string
	RequireVersion string
	Forbid         bool
	WarnVersion    string
	Warning        string // 字面值或 <const:Name> 引用
	WarningOnce    bool
	Instead        string
	ETC            string
}

// loadYAMLForTest 解析 specs/cmd_gen.yaml,返回已展开常量的 384 个 commandFromYAML。
//
// 极简 YAML 解析,只支持 cmd/genmeta 输出的固定子集 (key: "value" 或 bool 直接量)。
// 与生产代码中的解析器独立 (不导入,避免测试依赖 main 包)。
func loadYAMLForTest(t *testing.T) []commandFromYAML {
	t.Helper()
	root := repoRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "specs/cmd_gen.yaml"))
	if err != nil {
		t.Fatalf("read specs/cmd_gen.yaml: %v", err)
	}

	consts := map[string]string{}
	var cmds []commandFromYAML

	var section string // "constants" / "commands"
	var cur *commandFromYAML
	commit := func() {
		if cur != nil {
			cmds = append(cmds, *cur)
			cur = nil
		}
	}

	for ln, line := range strings.Split(string(data), "\n") {
		s := strings.TrimRight(line, "\r")
		trimmed := strings.TrimSpace(s)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		// 顶层段标记
		if !strings.HasPrefix(s, " ") {
			commit()
			switch {
			case strings.HasPrefix(s, "constants:"):
				section = "constants"
			case strings.HasPrefix(s, "commands:"):
				section = "commands"
			}
			continue
		}
		switch section {
		case "constants":
			body := strings.TrimLeft(s, " ")
			idx := strings.Index(body, ": ")
			if idx < 0 {
				continue
			}
			name := body[:idx]
			v, err := strconv.Unquote(strings.TrimSpace(body[idx+2:]))
			if err != nil {
				t.Fatalf("line %d: unquote %q: %v", ln+1, body[idx+2:], err)
			}
			consts[name] = v
		case "commands":
			indent := 0
			for _, c := range s {
				if c == ' ' {
					indent++
				} else {
					break
				}
			}
			body := strings.TrimLeft(s, " ")
			if indent == 2 && strings.HasPrefix(body, "- name:") {
				commit()
				cur = &commandFromYAML{Name: strings.TrimSpace(strings.TrimPrefix(body, "- name:"))}
				continue
			}
			if cur == nil {
				continue
			}
			// 跳过 pipelineParams 列表项 (本测试不关心 pipeline,只校验 metadata 9 字段)
			if indent == 6 && strings.HasPrefix(body, "- ") {
				continue
			}
			idx := strings.Index(body, ": ")
			if idx < 0 {
				// "key:" 空 value (pipelineParams 等),忽略
				continue
			}
			key := body[:idx]
			vRaw := strings.TrimSpace(body[idx+2:])
			val := vRaw
			if strings.HasPrefix(vRaw, `"`) {
				uv, err := strconv.Unquote(vRaw)
				if err != nil {
					t.Fatalf("line %d: %v", ln+1, err)
				}
				val = uv
			}
			switch key {
			case "cmd":
				cur.CmdString = val
			case "class":
				cur.Class = val
			case "requireVersion":
				cur.RequireVersion = val
			case "forbid":
				cur.Forbid = vRaw == "true"
			case "warnVersion":
				cur.WarnVersion = val
			case "warning":
				cur.Warning = val
			case "warningOnce":
				cur.WarningOnce = vRaw == "true"
			case "instead":
				cur.Instead = val
			case "etc":
				cur.ETC = val
			}
		}
	}
	commit()

	// 展开 <const:Name> 引用为字面常量
	expand := func(v string) string {
		if !strings.HasPrefix(v, "<const:") || !strings.HasSuffix(v, ">") {
			return v
		}
		name := v[len("<const:") : len(v)-1]
		if cv, ok := consts[name]; ok {
			return cv
		}
		t.Fatalf("yaml references unknown const %q", name)
		return ""
	}
	for i := range cmds {
		cmds[i].Warning = expand(cmds[i].Warning)
		cmds[i].Instead = expand(cmds[i].Instead)
		cmds[i].ETC = expand(cmds[i].ETC)
	}
	return cmds
}

// repoRoot 取测试运行时的仓库根目录 (CGO/working dir 无关)。
func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("runtime.Caller failed")
	}
	return filepath.Dir(thisFile)
}

// TestCmdGenFull_MetadataMatchesYAML 全量校验:
// specs/cmd_gen.yaml 中所有 384 个命令的 8 个 metadata 字段,
// 必须与运行时 CommandXxx.Class/RequireVersion/... 完全一致。
//
// 注意:用 Name (Go 标识符,如 "ZAdd") 索引,而非 String() (Redis 协议名,
// 如 "ZADD")——后者在 ZAdd/ZAddArgs 等变体间共享会导致歧义。
//
// 任一侧改动 (代码或 yaml) 未同步另一侧,测试失败提示运行
// `go run ./cmd/genmeta -extract=specs/cmd_gen.yaml` 或 `-generate=...`。
func TestCmdGenFull_MetadataMatchesYAML(t *testing.T) {
	yamlCmds := loadYAMLForTest(t)
	registry := allRegisteredCommands()

	if len(yamlCmds) != len(registry) {
		t.Fatalf("yaml has %d commands, registry has %d (mismatch)", len(yamlCmds), len(registry))
	}

	rt := make(map[string]Command, len(registry))
	for _, e := range registry {
		rt[e.Name] = e.Command
	}

	for _, y := range yamlCmds {
		c, ok := rt[y.Name]
		if !ok {
			t.Errorf("%s: in yaml but missing from registry", y.Name)
			continue
		}
		check := func(field, want, got string) {
			if want != got {
				t.Errorf("%s.%s: yaml=%q runtime=%q", y.Name, field, want, got)
			}
		}
		check("String", y.CmdString, c.String())
		check("Class", y.Class, c.Class())
		check("RequireVersion", y.RequireVersion, c.RequireVersion())
		check("WarnVersion", y.WarnVersion, c.WarnVersion())
		check("Warning", y.Warning, c.Warning())
		check("Instead", y.Instead, c.Instead())
		check("ETC", y.ETC, c.ETC())
		if y.Forbid != c.Forbid() {
			t.Errorf("%s.Forbid: yaml=%v runtime=%v", y.Name, y.Forbid, c.Forbid())
		}
		if y.WarningOnce != c.WarningOnce() {
			t.Errorf("%s.WarningOnce: yaml=%v runtime=%v", y.Name, y.WarningOnce, c.WarningOnce())
		}
	}

	// 反向:registry 有但 yaml 缺
	yamlNames := make(map[string]bool, len(yamlCmds))
	for _, y := range yamlCmds {
		yamlNames[y.Name] = true
	}
	for _, e := range registry {
		if !yamlNames[e.Name] {
			t.Errorf("registry has %s but yaml does not", e.Name)
		}
	}
}

// TestCmdGenFull_Invariants 各全集不变式:
//   - String() 不为空
//   - Class() 在已知 15 个集合中
//   - RequireVersion() 形如 X.Y.Z 或为空
//   - WarnVersion 非空时, Warning 必须非空
func TestCmdGenFull_Invariants(t *testing.T) {
	knownClasses := map[string]bool{
		"Bitmap": true, "Cluster": true, "Connection": true, "Generic": true,
		"Geospatial": true, "Hash": true, "HyperLog": true, "List": true,
		"PubSub": true, "Scripting": true, "Server": true, "Set": true,
		"SortedSet": true, "Stream": true, "String": true,
	}

	for _, e := range allRegisteredCommands() {
		c := e.Command
		if c.String() == "" {
			t.Errorf("%s: String() empty", e.Name)
		}
		if !knownClasses[c.Class()] {
			t.Errorf("%s: unknown class %q", e.Name, c.Class())
		}
		if v := c.RequireVersion(); v != "" {
			parts := strings.Split(v, ".")
			if len(parts) != 3 {
				t.Errorf("%s: RequireVersion %q not in X.Y.Z form", e.Name, v)
			}
		}
		if c.WarnVersion() != "" {
			if c.Warning() == "" {
				t.Errorf("%s: WarnVersion=%q but Warning empty", e.Name, c.WarnVersion())
			}
		}
	}
}

// TestCmdGenFull_NoUnreachableConstants 校验 specs/cmd_gen.yaml 中
// constants 的每个常量都至少被一个命令的 Warning/Instead/ETC 引用。
func TestCmdGenFull_NoUnreachableConstants(t *testing.T) {
	root := repoRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "specs/cmd_gen.yaml"))
	if err != nil {
		t.Fatalf("read yaml: %v", err)
	}
	text := string(data)
	// 抽取 constants 块下所有常量名
	var constNames []string
	var inConst bool
	for _, line := range strings.Split(text, "\n") {
		s := strings.TrimRight(line, "\r")
		if strings.HasPrefix(s, "constants:") {
			inConst = true
			continue
		}
		if inConst {
			if !strings.HasPrefix(s, "  ") || strings.TrimSpace(s) == "" {
				if strings.TrimSpace(s) == "" {
					continue
				}
				inConst = false
				continue
			}
			body := strings.TrimLeft(s, " ")
			if idx := strings.Index(body, ": "); idx > 0 {
				constNames = append(constNames, body[:idx])
			}
		}
	}
	if len(constNames) == 0 {
		t.Fatalf("no constants extracted from yaml")
	}

	// 每个常量必须在文本中至少被引用一次 (除自定义之外)
	for _, name := range constNames {
		// 在 yaml 中搜索 "<const:NAME>" 引用
		needle := "<const:" + name + ">"
		if !strings.Contains(text, needle) {
			t.Errorf("constant %s declared but no command references %s", name, needle)
		}
	}
}
