// genmeta 是 redisson 的 cmd_gen_*.go 一致性校验器、规约文件读写工具与代码生成器。
//
// 现状:
//
//	cmd_gen.go 历史上是单文件 7204 行,内含 384 个命令的元数据对象 (Class/RequireVersion/
//	Forbid/Warning/Instead/ETC) 与 Pipeliner P()/Cmd() 包装方法。文件过大影响 IDE
//	加载与 PR 审查,虽然是 generated 代码但缺乏可重复生成的工具链与说明书。
//
// 工作模式 (互斥):
//
//  1. -extract=specs/cmd_gen.yaml: 从现有 cmd_gen.go AST 提取 384 个命令元数据为
//     YAML 说明书 (人类可读、版本可控)。
//
//  2. -generate=specs/cmd_gen.yaml: 从 YAML 说明书重新生成 cmd_gen_<class>.go (15 个
//     文件,按 Class 拆分)。生成结果与原 cmd_gen.go 行为完全一致。
//
//  3. -check (默认): 同时跑 extract + 与 specs/cmd_gen.yaml 比对,确认代码与说明书
//     未漂移。 CI 用此模式守门。
//
// 集成方式:
//
//	使用 go generate 调用 (在被 generated 的源文件中加上指令),或者
//	直接 'make cmdgen-check' (与 builder-check 并列)。
//
// 三位一体闭环:
//
//	specs/cmd_gen.yaml  ←  extract  ←  cmd_gen_<class>.go (15 文件)
//	                  →  generate  →
//	                  →   check    →  CI 守门 + 元数据测试 (cmd_gen_meta_test.go) 双向校验
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// ============================================================================
// 数据结构
// ============================================================================

// Meta 单条命令的全部元数据。
//
// 字段命名遵循 redisson Command interface (Class/RequireVersion 等);
// Pipeline* 字段仅在 HasPipeline=true 时有意义。
type Meta struct {
	Name           string `yaml:"name"`           // BitCount (= Go 标识符,首字母大写)
	CmdString      string `yaml:"cmd"`            // BITCOUNT (Redis 协议名,全大写,可能含空格)
	Class          string `yaml:"class"`          // Bitmap / SortedSet / ...
	RequireVersion string `yaml:"requireVersion"` // 2.6.0
	Forbid         bool   `yaml:"forbid,omitempty"`
	WarnVersion    string `yaml:"warnVersion,omitempty"`
	Warning        string `yaml:"warning,omitempty"` // 字面字符串,常量已展开
	WarningOnce    bool   `yaml:"warningOnce,omitempty"`
	Instead        string `yaml:"instead,omitempty"`
	ETC            string `yaml:"etc,omitempty"`

	// HasPipeline 标记本命令是否生成 P()/Cmd() 流水线包装方法。
	// 部分阻塞命令 (BLPOP/BRPOP/BZPOPMIN 等) 不适合 pipeline,无 P()/Cmd()。
	HasPipeline bool `yaml:"hasPipeline,omitempty"`

	// PipelineRet PR(b BaseCmd) 返回的具体 Cmd 接口名,如 IntCmd / StringSliceCmd。
	PipelineRet string `yaml:"pipelineRet,omitempty"`

	// PipelineCmdImpl Cmd 方法 body 中实例化的具体类型,如 intCmd / stringSliceCmd。
	PipelineCmdImpl string `yaml:"pipelineCmdImpl,omitempty"`

	// PipelineBuilder Cmd 方法调用的 builder 函数名,如 BitCountCompleted。
	PipelineBuilder string `yaml:"pipelineBuilder,omitempty"`

	// PipelineParams Cmd 方法的形参列表,如 ["key string", "bc *BitCount"]。
	PipelineParams []string `yaml:"pipelineParams,omitempty"`

	// PipelineCallExpr 传给 builder 函数的实参表达式,如 "key, bc" 或 "key, args...".
	PipelineCallExpr string `yaml:"pipelineCallExpr,omitempty"`
}

// Spec 整个 cmd_gen.yaml 的顶层结构。
type Spec struct {
	// Constants 引用到的字符串常量 (Warning/Instead/ETC 内复用的多行长文本)。
	// extract 阶段从 cmd_gen.go const 块提取; generate 阶段写回到 cmd_gen.go 顶部。
	Constants map[string]string `yaml:"constants,omitempty"`
	Commands  []Meta            `yaml:"commands"`
}

// ============================================================================
// 入口
// ============================================================================

func main() {
	var (
		extract  string
		generate string
		check    bool
		root     string
	)
	flag.StringVar(&extract, "extract", "", "extract metadata from cmd_gen*.go to YAML path")
	flag.StringVar(&generate, "generate", "", "generate cmd_gen_<class>.go from YAML path")
	flag.BoolVar(&check, "check", false, "verify cmd_gen*.go matches specs/cmd_gen.yaml")
	flag.StringVar(&root, "root", ".", "redisson source root")
	flag.Parse()

	switch {
	case extract != "":
		if err := doExtract(root, extract); err != nil {
			fatal(err)
		}
	case generate != "":
		if err := doGenerate(root, generate); err != nil {
			fatal(err)
		}
	case check:
		if err := doCheck(root, "specs/cmd_gen.yaml"); err != nil {
			fatal(err)
		}
	default:
		flag.Usage()
		os.Exit(2)
	}
}

func fatal(err error) {
	_, _ = fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

// ============================================================================
// extract: cmd_gen*.go AST → Spec
// ============================================================================

func doExtract(root, outPath string) error {
	spec, err := extractSpec(root)
	if err != nil {
		return err
	}
	data, err := marshalSpec(spec)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(outPath, data, 0o644); err != nil {
		return err
	}
	fmt.Printf("extracted %d commands to %s\n", len(spec.Commands), outPath)
	return nil
}

// extractSpec 扫描 root 下所有 cmd_gen*.go,产出完整 Spec。
//
// 解析规则 (与 cmd_gen.go 当前形态严格对应):
//
//   - 形如 `var CommandFoo commandFoo` 的全局声明 → 注册一个 Meta{Name: "Foo"}
//   - 方法接收者 `commandFoo` (无 P 后缀) 上的 String/Class/RequireVersion 等
//     字面 string return → 直接写入 meta 字段
//   - 接收者 `commandFooP` 上的 Cmd(...) 方法 → 提取签名与 body
//   - const 块的 commandXxxWarning 等字符串常量 → 写入 spec.Constants 并把
//     meta 的对应字段改为 "<const:commandXxxWarning>" 引用形式 (generate 时
//     再展开)
func extractSpec(root string) (*Spec, error) {
	spec := &Spec{Constants: map[string]string{}}
	metas := map[string]*Meta{}

	files, err := listCmdGenFiles(root)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no cmd_gen*.go files found under %s", root)
	}

	fset := token.NewFileSet()
	for _, path := range files {
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		for _, decl := range f.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				extractGenDecl(d, spec, metas)
			case *ast.FuncDecl:
				extractFuncDecl(d, metas)
			}
		}
	}

	// 把 metas map 排序为 slice
	names := make([]string, 0, len(metas))
	for n := range metas {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		spec.Commands = append(spec.Commands, *metas[n])
	}
	return spec, nil
}

func listCmdGenFiles(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		// 严格匹配 cmd_gen.go (历史单文件) 与 cmd_gen_<class>.go (新拆分),
		// 排除:
		//   - cmd_generic.go (前缀冲突,实际是手写的 generic 命令实现)
		//   - cmd_gen_doc.go (纯 documentation 文件,不含 commandXxx 声明)
		//   - 任意 _test.go
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		if name == "cmd_generic.go" || name == "cmd_gen_doc.go" {
			continue
		}
		isOldSingle := name == "cmd_gen.go"
		isNewSplit := strings.HasPrefix(name, "cmd_gen_")
		if isOldSingle || isNewSplit {
			out = append(out, filepath.Join(root, name))
		}
	}
	sort.Strings(out)
	return out, nil
}

// extractGenDecl 处理:
//   - var CommandFoo commandFoo  → 注册命令名
//   - type commandFoo string / type commandFooP struct{p Pipeliner} → 忽略
//   - const block 内的字符串常量 → 写入 Constants
func extractGenDecl(d *ast.GenDecl, spec *Spec, metas map[string]*Meta) {
	switch d.Tok {
	case token.VAR:
		for _, s := range d.Specs {
			vs, ok := s.(*ast.ValueSpec)
			if !ok || len(vs.Names) != 1 {
				continue
			}
			name := vs.Names[0].Name
			if !strings.HasPrefix(name, "Command") {
				continue
			}
			cmdName := strings.TrimPrefix(name, "Command")
			if _, exists := metas[cmdName]; !exists {
				metas[cmdName] = &Meta{Name: cmdName}
			}
		}
	case token.CONST:
		for _, s := range d.Specs {
			vs, ok := s.(*ast.ValueSpec)
			if !ok || len(vs.Names) == 0 || len(vs.Values) == 0 {
				continue
			}
			name := vs.Names[0].Name
			if !strings.HasPrefix(name, "command") {
				continue
			}
			lit, ok := vs.Values[0].(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				continue
			}
			v, err := strconv.Unquote(lit.Value)
			if err != nil {
				continue
			}
			spec.Constants[name] = v
		}
	}
}

// extractFuncDecl 解析 func (commandFoo) String() / func (b commandFooP) Cmd(...)。
func extractFuncDecl(d *ast.FuncDecl, metas map[string]*Meta) {
	if d.Recv == nil || len(d.Recv.List) == 0 {
		return
	}
	recvType := types.ExprString(d.Recv.List[0].Type)
	if !strings.HasPrefix(recvType, "command") {
		return
	}
	isP := strings.HasSuffix(recvType, "P")
	cmdName := strings.TrimPrefix(recvType, "command")
	if isP {
		cmdName = strings.TrimSuffix(cmdName, "P")
	}
	m, ok := metas[cmdName]
	if !ok {
		// 没注册过 var 但有方法,跳过 (理论不应发生)
		return
	}

	body := d.Body
	if body == nil {
		return
	}
	method := d.Name.Name

	if !isP {
		extractMetaMethod(m, method, body)
	} else if method == "Cmd" {
		extractCmdMethod(m, d)
	}
}

// extractMetaMethod 解析非 P 接收者的 metadata 方法。
func extractMetaMethod(m *Meta, method string, body *ast.BlockStmt) {
	switch method {
	case "P":
		m.HasPipeline = true
	case "PR":
		if len(body.List) != 1 {
			return
		}
		ret, ok := body.List[0].(*ast.ReturnStmt)
		if !ok || len(ret.Results) != 1 {
			return
		}
		ta, ok := ret.Results[0].(*ast.TypeAssertExpr)
		if !ok {
			return
		}
		m.PipelineRet = types.ExprString(ta.Type)
	default:
		// String/Class/RequireVersion/WarnVersion/Warning/Instead/ETC: return "literal"
		// Forbid/WarningOnce: return true/false
		if len(body.List) != 1 {
			return
		}
		ret, ok := body.List[0].(*ast.ReturnStmt)
		if !ok || len(ret.Results) != 1 {
			return
		}
		setStringField(m, method, ret.Results[0])
		setBoolField(m, method, ret.Results[0])
	}
}

// setStringField 处理 String/Class/RequireVersion 等字符串字段,
// 既支持字面量也支持常量引用 (如 commandKeysWarning),后者保留 <const:Name> 形式。
func setStringField(m *Meta, method string, expr ast.Expr) {
	var v string
	switch e := expr.(type) {
	case *ast.BasicLit:
		if e.Kind != token.STRING {
			return
		}
		s, err := strconv.Unquote(e.Value)
		if err != nil {
			return
		}
		v = s
	case *ast.Ident:
		// 常量引用,留待 generate 时再展开
		v = "<const:" + e.Name + ">"
	default:
		return
	}
	switch method {
	case "String":
		m.CmdString = v
	case "Class":
		m.Class = v
	case "RequireVersion":
		m.RequireVersion = v
	case "WarnVersion":
		m.WarnVersion = v
	case "Warning":
		m.Warning = v
	case "Instead":
		m.Instead = v
	case "ETC":
		m.ETC = v
	}
}

func setBoolField(m *Meta, method string, expr ast.Expr) {
	id, ok := expr.(*ast.Ident)
	if !ok {
		return
	}
	b := id.Name == "true"
	switch method {
	case "Forbid":
		m.Forbid = b
	case "WarningOnce":
		m.WarningOnce = b
	}
}

// extractCmdMethod 解析 P 接收者的 Cmd 方法体:
//
//	func (b commandFooP) Cmd(key string, bc *BitCount) {
//	    b.p.cmd(b.p.builder().FooCompleted(key, bc), &intCmd{})
//	}
func extractCmdMethod(m *Meta, d *ast.FuncDecl) {
	m.HasPipeline = true

	// params
	for _, field := range d.Type.Params.List {
		t := types.ExprString(field.Type)
		for _, name := range field.Names {
			m.PipelineParams = append(m.PipelineParams, name.Name+" "+t)
		}
	}

	body := d.Body
	if len(body.List) != 1 {
		return
	}
	exprStmt, ok := body.List[0].(*ast.ExprStmt)
	if !ok {
		return
	}
	call, ok := exprStmt.X.(*ast.CallExpr)
	if !ok || len(call.Args) != 2 {
		return
	}
	// arg0: b.p.builder().FooCompleted(args...)
	if c1, ok := call.Args[0].(*ast.CallExpr); ok {
		if sel, ok := c1.Fun.(*ast.SelectorExpr); ok {
			m.PipelineBuilder = sel.Sel.Name
			args := make([]string, 0, len(c1.Args))
			for _, a := range c1.Args {
				args = append(args, types.ExprString(a))
			}
			expr := strings.Join(args, ", ")
			if c1.Ellipsis != token.NoPos && len(args) > 0 {
				// 最后一个参数是 spread (foo...)
				expr = strings.Join(args[:len(args)-1], ", ")
				if len(args) > 0 {
					if expr != "" {
						expr += ", "
					}
					expr += args[len(args)-1] + "..."
				}
			}
			m.PipelineCallExpr = expr
		}
	}
	// arg1: &xxxCmd{}
	if u, ok := call.Args[1].(*ast.UnaryExpr); ok {
		if c2, ok := u.X.(*ast.CompositeLit); ok {
			m.PipelineCmdImpl = types.ExprString(c2.Type)
		}
	}
}

// ============================================================================
// generate: Spec → cmd_gen_<class>.go
// ============================================================================

func doGenerate(root, specPath string) error {
	spec, err := readSpec(specPath)
	if err != nil {
		return err
	}
	files, err := generateFiles(spec)
	if err != nil {
		return err
	}
	// 删除旧的 cmd_gen.go (单文件) 与 cmd_gen_*.go (本工具上次产出),
	// 但保留 _test.go。
	if err := cleanOldGenFiles(root); err != nil {
		return err
	}
	for path, content := range files {
		full := filepath.Join(root, path)
		if err := os.WriteFile(full, content, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", full, err)
		}
	}
	fmt.Printf("generated %d files (%d commands)\n", len(files), len(spec.Commands))
	return nil
}

func cleanOldGenFiles(root string) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		name := e.Name()
		if name == "cmd_gen.go" || strings.HasPrefix(name, "cmd_gen_") {
			full := filepath.Join(root, name)
			if err := os.Remove(full); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("remove %s: %w", full, err)
			}
		}
	}
	return nil
}

// generateFiles 按 Class 分组产出 cmd_gen_<class>.go;同时产出 cmd_gen_const.go
// 集中存放共享常量 (Warning/Instead/ETC 长文本)。
func generateFiles(spec *Spec) (map[string][]byte, error) {
	byClass := map[string][]Meta{}
	for _, m := range spec.Commands {
		byClass[m.Class] = append(byClass[m.Class], m)
	}
	for cls := range byClass {
		sort.Slice(byClass[cls], func(i, j int) bool {
			return byClass[cls][i].Name < byClass[cls][j].Name
		})
	}

	out := map[string][]byte{}

	// 共享常量 + Command interface 定义 (无 import 需求)
	{
		buf := &bytes.Buffer{}
		writeFileHeader(buf)
		writeCommandInterface(buf)
		writeConstants(buf, spec.Constants)
		formatted, err := format.Source(buf.Bytes())
		if err != nil {
			return nil, fmt.Errorf("format cmd_gen_const.go: %w\n%s", err, buf.String())
		}
		out["cmd_gen_const.go"] = formatted
	}

	// 每个 class 一个文件
	classes := make([]string, 0, len(byClass))
	for cls := range byClass {
		classes = append(classes, cls)
	}
	sort.Strings(classes)
	for _, cls := range classes {
		buf := &bytes.Buffer{}
		writeFileHeader(buf)
		// time 仅当某些命令引用时才需要 import,这里固定加并用 _ 避免空 import 错
		needTime := false
		for _, m := range byClass[cls] {
			for _, p := range m.PipelineParams {
				if strings.Contains(p, "time.Duration") || strings.Contains(p, "time.Time") {
					needTime = true
					break
				}
			}
			if needTime {
				break
			}
		}
		if needTime {
			writeImports(buf, "time")
		} else {
			fmt.Fprintln(buf, "")
		}
		for _, m := range byClass[cls] {
			writeCommand(buf, &m, spec.Constants)
		}
		formatted, err := format.Source(buf.Bytes())
		if err != nil {
			return nil, fmt.Errorf("format cmd_gen_%s.go: %w\n%s", lowerClass(cls), err, buf.String())
		}
		out["cmd_gen_"+lowerClass(cls)+".go"] = formatted
	}
	return out, nil
}

func lowerClass(cls string) string {
	// SortedSet → sortedset, HyperLog → hyperlog, PubSub → pubsub
	return strings.ToLower(cls)
}

func writeFileHeader(buf *bytes.Buffer) {
	fmt.Fprintln(buf, "// Code generated by cmd/genmeta. DO NOT EDIT.")
	fmt.Fprintln(buf, "package redisson")
	fmt.Fprintln(buf, "")
}

func writeImports(buf *bytes.Buffer, pkgs ...string) {
	if len(pkgs) == 0 {
		return
	}
	fmt.Fprintln(buf, "import (")
	for _, p := range pkgs {
		fmt.Fprintf(buf, "\t%q\n", p)
	}
	fmt.Fprintln(buf, ")")
	fmt.Fprintln(buf, "")
}

func writeCommandInterface(buf *bytes.Buffer) {
	fmt.Fprintln(buf, `// Command 是所有 redisson 命令的元数据接口,由 cmd_gen_<class>.go 中的
// commandXxx 类型实现。所有 commandXxx 的具体值为单一全局零值 (var CommandXxx commandXxx),
// 通过类型方法暴露 metadata,无运行时分配。`)
	fmt.Fprintln(buf, "type Command interface {")
	fmt.Fprintln(buf, "\tString() string")
	fmt.Fprintln(buf, "\tClass() string")
	fmt.Fprintln(buf, "\tRequireVersion() string")
	fmt.Fprintln(buf, "\tForbid() bool")
	fmt.Fprintln(buf, "\tWarnVersion() string")
	fmt.Fprintln(buf, "\tWarning() string")
	fmt.Fprintln(buf, "\tWarningOnce() bool")
	fmt.Fprintln(buf, "\tInstead() string")
	fmt.Fprintln(buf, "\tETC() string")
	fmt.Fprintln(buf, "}")
	fmt.Fprintln(buf, "")
}

func writeConstants(buf *bytes.Buffer, consts map[string]string) {
	if len(consts) == 0 {
		return
	}
	fmt.Fprintln(buf, "const (")
	names := make([]string, 0, len(consts))
	for n := range consts {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		fmt.Fprintf(buf, "\t%s = %s\n", n, strconv.Quote(consts[n]))
	}
	fmt.Fprintln(buf, ")")
	fmt.Fprintln(buf, "")
}

// writeCommand 输出单个命令的全部声明 + 方法。
func writeCommand(buf *bytes.Buffer, m *Meta, _ map[string]string) {
	fmt.Fprintf(buf, "var Command%s command%s\n", m.Name, m.Name)
	fmt.Fprintf(buf, "type command%s string\n", m.Name)
	if m.HasPipeline {
		fmt.Fprintf(buf, "type command%sP struct{ p Pipeliner }\n", m.Name)
	}

	emitStrMethod(buf, m.Name, "String", m.CmdString)
	emitStrMethod(buf, m.Name, "Class", m.Class)
	emitStrMethod(buf, m.Name, "RequireVersion", m.RequireVersion)
	emitBoolMethod(buf, m.Name, "Forbid", m.Forbid)
	emitBoolMethod(buf, m.Name, "WarningOnce", m.WarningOnce)
	emitStrMethod(buf, m.Name, "WarnVersion", m.WarnVersion)
	emitStrMethod(buf, m.Name, "Warning", m.Warning)
	emitStrMethod(buf, m.Name, "Instead", m.Instead)
	emitStrMethod(buf, m.Name, "ETC", m.ETC)

	if m.HasPipeline {
		fmt.Fprintf(buf, "func (command%s) PR(b BaseCmd) %s { return b.(%s) }\n",
			m.Name, m.PipelineRet, m.PipelineRet)
		fmt.Fprintf(buf, "func (command%s) P(p Pipeliner) command%sP { return command%sP{p} }\n",
			m.Name, m.Name, m.Name)
		fmt.Fprintf(buf, "func (b command%sP) Cmd(%s) {\n", m.Name, strings.Join(m.PipelineParams, ", "))
		fmt.Fprintf(buf, "\tb.p.cmd(b.p.builder().%s(%s), &%s{})\n",
			m.PipelineBuilder, m.PipelineCallExpr, m.PipelineCmdImpl)
		fmt.Fprintln(buf, "}")
	}
	fmt.Fprintln(buf, "")
}

// emitStrMethod 输出 func (commandXxx) Foo() string { return "value" }
//
// 字面值 v 为 "<const:Name>" 时改输出 return Name (常量引用)。
func emitStrMethod(buf *bytes.Buffer, cmdName, method, v string) {
	if strings.HasPrefix(v, "<const:") && strings.HasSuffix(v, ">") {
		constName := v[len("<const:") : len(v)-1]
		fmt.Fprintf(buf, "func (command%s) %s() string { return %s }\n", cmdName, method, constName)
	} else {
		fmt.Fprintf(buf, "func (command%s) %s() string { return %s }\n", cmdName, method, strconv.Quote(v))
	}
}

func emitBoolMethod(buf *bytes.Buffer, cmdName, method string, v bool) {
	fmt.Fprintf(buf, "func (command%s) %s() bool { return %t }\n", cmdName, method, v)
}

// ============================================================================
// check: 比对当前代码与 spec
// ============================================================================

func doCheck(root, specPath string) error {
	specCurrent, err := extractSpec(root)
	if err != nil {
		return fmt.Errorf("extract: %w", err)
	}
	specStored, err := readSpec(specPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", specPath, err)
	}

	// 比对常量
	if !reflect.DeepEqual(specCurrent.Constants, specStored.Constants) {
		// 找差异常量
		var diffs []string
		for k, v := range specCurrent.Constants {
			if specStored.Constants[k] != v {
				diffs = append(diffs, fmt.Sprintf("  - %s (current vs spec mismatch)", k))
			}
		}
		for k := range specStored.Constants {
			if _, ok := specCurrent.Constants[k]; !ok {
				diffs = append(diffs, fmt.Sprintf("  - %s (in spec but missing from code)", k))
			}
		}
		return fmt.Errorf("constants drift:\n%s", strings.Join(diffs, "\n"))
	}

	// 比对命令
	currentMap := indexMetas(specCurrent.Commands)
	storedMap := indexMetas(specStored.Commands)
	var diffs []string
	for n, c := range currentMap {
		s, ok := storedMap[n]
		if !ok {
			diffs = append(diffs, fmt.Sprintf("  - %s: in code but missing from spec", n))
			continue
		}
		if !reflect.DeepEqual(c, s) {
			diffs = append(diffs, fmt.Sprintf("  - %s: drift", n))
		}
	}
	for n := range storedMap {
		if _, ok := currentMap[n]; !ok {
			diffs = append(diffs, fmt.Sprintf("  - %s: in spec but missing from code", n))
		}
	}
	if len(diffs) > 0 {
		return fmt.Errorf("cmd_gen drift between code and %s:\n%s\n\n  fix: re-run 'go run ./cmd/genmeta -extract=%s' or '-generate=%s'",
			specPath, strings.Join(diffs, "\n"), specPath, specPath)
	}
	fmt.Printf("ok: %d commands across %d files match %s\n",
		len(specCurrent.Commands), len(must(listCmdGenFiles(root))), specPath)
	return nil
}

func indexMetas(metas []Meta) map[string]Meta {
	out := make(map[string]Meta, len(metas))
	for _, m := range metas {
		out[m.Name] = m
	}
	return out
}

func must[T any](v T, err error) T {
	if err != nil {
		fatal(err)
	}
	return v
}

// ============================================================================
// 极简 YAML 序列化 (避免依赖第三方库)
//
// 因为 Spec 结构稳定 (string/bool/[]string/map[string]string 全枚举),手写序列化
// 比引入 yaml.v3 更轻量,且不影响 go.mod 依赖图。
// ============================================================================

func marshalSpec(s *Spec) ([]byte, error) {
	buf := &bytes.Buffer{}
	fmt.Fprintln(buf, "# Code generated by cmd/genmeta -extract. Edit the source then re-run extract.")
	fmt.Fprintln(buf, "# Spec for cmd_gen_<class>.go: 384 commands across 15 classes.")
	fmt.Fprintln(buf, "")

	if len(s.Constants) > 0 {
		fmt.Fprintln(buf, "constants:")
		names := make([]string, 0, len(s.Constants))
		for n := range s.Constants {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			fmt.Fprintf(buf, "  %s: %s\n", n, yamlString(s.Constants[n]))
		}
		fmt.Fprintln(buf, "")
	}

	fmt.Fprintln(buf, "commands:")
	for _, m := range s.Commands {
		writeMeta(buf, &m)
	}
	return buf.Bytes(), nil
}

func writeMeta(buf *bytes.Buffer, m *Meta) {
	fmt.Fprintf(buf, "  - name: %s\n", m.Name)
	fmt.Fprintf(buf, "    cmd: %s\n", yamlString(m.CmdString))
	fmt.Fprintf(buf, "    class: %s\n", m.Class)
	fmt.Fprintf(buf, "    requireVersion: %s\n", yamlString(m.RequireVersion))
	if m.Forbid {
		fmt.Fprintln(buf, "    forbid: true")
	}
	if m.WarnVersion != "" {
		fmt.Fprintf(buf, "    warnVersion: %s\n", yamlString(m.WarnVersion))
	}
	if m.Warning != "" {
		fmt.Fprintf(buf, "    warning: %s\n", yamlString(m.Warning))
	}
	if m.WarningOnce {
		fmt.Fprintln(buf, "    warningOnce: true")
	}
	if m.Instead != "" {
		fmt.Fprintf(buf, "    instead: %s\n", yamlString(m.Instead))
	}
	if m.ETC != "" {
		fmt.Fprintf(buf, "    etc: %s\n", yamlString(m.ETC))
	}
	if m.HasPipeline {
		fmt.Fprintln(buf, "    hasPipeline: true")
		fmt.Fprintf(buf, "    pipelineRet: %s\n", m.PipelineRet)
		fmt.Fprintf(buf, "    pipelineCmdImpl: %s\n", m.PipelineCmdImpl)
		fmt.Fprintf(buf, "    pipelineBuilder: %s\n", m.PipelineBuilder)
		if len(m.PipelineParams) > 0 {
			fmt.Fprintln(buf, "    pipelineParams:")
			for _, p := range m.PipelineParams {
				fmt.Fprintf(buf, "      - %s\n", yamlString(p))
			}
		}
		if m.PipelineCallExpr != "" {
			fmt.Fprintf(buf, "    pipelineCallExpr: %s\n", yamlString(m.PipelineCallExpr))
		}
	}
}

// yamlString 用 strconv.Quote 总是产出双引号字符串,简单且对长文本/特殊字符安全。
// 避免实现 YAML block scalar 等复杂语法,但确保解析侧只需要支持引号字符串。
func yamlString(s string) string {
	return strconv.Quote(s)
}

// ============================================================================
// 极简 YAML 反序列化 (与 marshalSpec 配对)
//
// 不支持完整 YAML 语法,只支持 marshalSpec 输出的子集:
//   - 顶层 key: value 或 key:
//   - 二级:四空格缩进 list 项 "- name: ..."
//   - 字符串值都是双引号包裹
//   - 注释行 (#) 与空行忽略
// ============================================================================

func readSpec(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseSpec(string(data))
}

func parseSpec(text string) (*Spec, error) {
	spec := &Spec{Constants: map[string]string{}}
	lines := strings.Split(text, "\n")

	var section string // "constants" / "commands"
	var current *Meta  // 正在解析的命令
	var inParams bool

	commit := func() {
		if current != nil {
			spec.Commands = append(spec.Commands, *current)
			current = nil
			inParams = false
		}
	}

	for i, raw := range lines {
		line := strings.TrimRight(raw, "\r\n")
		if strings.HasPrefix(strings.TrimSpace(line), "#") || strings.TrimSpace(line) == "" {
			continue
		}
		// 顶层 section
		if !strings.HasPrefix(line, " ") {
			commit()
			switch {
			case strings.HasPrefix(line, "constants:"):
				section = "constants"
			case strings.HasPrefix(line, "commands:"):
				section = "commands"
			default:
				return nil, fmt.Errorf("line %d: unknown top-level: %q", i+1, line)
			}
			continue
		}

		switch section {
		case "constants":
			// "  Name: \"value\""
			s := strings.TrimLeft(line, " ")
			idx := strings.Index(s, ": ")
			if idx < 0 {
				return nil, fmt.Errorf("line %d: bad constant line: %q", i+1, line)
			}
			name := s[:idx]
			vRaw := s[idx+2:]
			v, err := unquoteYAML(vRaw)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", i+1, err)
			}
			spec.Constants[name] = v
		case "commands":
			indent := countLeadingSpaces(line)
			s := strings.TrimLeft(line, " ")
			if indent == 2 && strings.HasPrefix(s, "- name:") {
				commit()
				current = &Meta{Name: strings.TrimSpace(strings.TrimPrefix(s, "- name:"))}
				continue
			}
			if current == nil {
				return nil, fmt.Errorf("line %d: stray command field outside item: %q", i+1, line)
			}
			if indent == 6 && strings.HasPrefix(s, "- ") {
				if !inParams {
					return nil, fmt.Errorf("line %d: list item without pipelineParams header", i+1)
				}
				v, err := unquoteYAML(strings.TrimPrefix(s, "- "))
				if err != nil {
					return nil, fmt.Errorf("line %d: %w", i+1, err)
				}
				current.PipelineParams = append(current.PipelineParams, v)
				continue
			}
			inParams = false
			// "    key: value" (indent 4)
			idx := strings.Index(s, ": ")
			if idx < 0 {
				// "    key:" (空 value,后续可能跟 list)
				key := strings.TrimSuffix(strings.TrimSpace(s), ":")
				if key == "pipelineParams" {
					inParams = true
				}
				continue
			}
			key := s[:idx]
			vRaw := s[idx+2:]
			if err := setMetaField(current, key, vRaw); err != nil {
				return nil, fmt.Errorf("line %d: %w", i+1, err)
			}
		}
	}
	commit()
	return spec, nil
}

func setMetaField(m *Meta, key, vRaw string) error {
	// bool
	switch key {
	case "forbid":
		m.Forbid = vRaw == "true"
		return nil
	case "warningOnce":
		m.WarningOnce = vRaw == "true"
		return nil
	case "hasPipeline":
		m.HasPipeline = vRaw == "true"
		return nil
	}
	// string (quoted 或 bare)
	v := vRaw
	if strings.HasPrefix(vRaw, `"`) {
		uv, err := unquoteYAML(vRaw)
		if err != nil {
			return err
		}
		v = uv
	}
	switch key {
	case "cmd":
		m.CmdString = v
	case "class":
		m.Class = v
	case "requireVersion":
		m.RequireVersion = v
	case "warnVersion":
		m.WarnVersion = v
	case "warning":
		m.Warning = v
	case "instead":
		m.Instead = v
	case "etc":
		m.ETC = v
	case "pipelineRet":
		m.PipelineRet = v
	case "pipelineCmdImpl":
		m.PipelineCmdImpl = v
	case "pipelineBuilder":
		m.PipelineBuilder = v
	case "pipelineCallExpr":
		m.PipelineCallExpr = v
	default:
		return fmt.Errorf("unknown field %q", key)
	}
	return nil
}

func unquoteYAML(s string) (string, error) {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, `"`) {
		return s, nil
	}
	return strconv.Unquote(s)
}

func countLeadingSpaces(s string) int {
	n := 0
	for _, c := range s {
		if c == ' ' {
			n++
		} else {
			break
		}
	}
	return n
}

// 仅用于消除 unused 警告 (filepath/fs 实际通过 listCmdGenFiles 间接使用)
var _ = filepath.Walk
var _ fs.FS = nil
