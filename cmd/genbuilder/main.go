// genbuilder 是 redisson 的 builder_*.go 一致性验证器。
//
// 该工具扫描所有 builder_*.go 文件，验证 builder 类型方法签名遵循以下约定：
//
//  1. 函数名以 "Completed" 结尾（除非以小写开头作为内部 helper）；
//  2. 返回值仅有一个，且类型为 Completed；
//  3. receiver 为 (b builder)。
//
// 该工具作为 //go:generate 钩子调用，CI 中也会执行：
//
//	//go:generate go run ./cmd/genbuilder -check
//
// 当前不生成代码——所有 builder_*.go 仍为手写。
// 未来若引入 spec → 生成的工作流，将复用本工具的 AST 解析骨架。
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const builderFilePrefix = "builder_"

type funcInfo struct {
	File     string
	Line     int
	Name     string
	Recv     string // receiver type, e.g. "builder"
	NumParam int
	NumRet   int
	RetType  string
}

func main() {
	var (
		root     string
		checkOnly bool
		verbose  bool
	)
	flag.StringVar(&root, "root", ".", "package root containing builder_*.go")
	flag.BoolVar(&checkOnly, "check", true, "fail (exit 1) if invariants violated; cannot be disabled in current version")
	flag.BoolVar(&verbose, "v", false, "print all discovered functions")
	flag.Parse()
	_ = checkOnly // placeholder for future generate mode

	files, err := discover(root)
	if err != nil {
		fail("discover: %v", err)
	}
	if len(files) == 0 {
		fail("no builder_*.go files found under %s", root)
	}

	fset := token.NewFileSet()
	var funcs []funcInfo
	for _, f := range files {
		fi, err := parseFile(fset, f)
		if err != nil {
			fail("parse %s: %v", f, err)
		}
		funcs = append(funcs, fi...)
	}
	sort.Slice(funcs, func(i, j int) bool { return funcs[i].Name < funcs[j].Name })

	var problems []string
	for _, fi := range funcs {
		if fi.Recv != "builder" {
			problems = append(problems, fmt.Sprintf("%s:%d %s: receiver %q (want builder)", fi.File, fi.Line, fi.Name, fi.Recv))
			continue
		}
		// 公开函数（首字母大写）必须以 Completed 结尾且返回 Completed。
		if isExported(fi.Name) {
			if !strings.HasSuffix(fi.Name, "Completed") {
				problems = append(problems, fmt.Sprintf("%s:%d %s: exported builder funcs must end with 'Completed'", fi.File, fi.Line, fi.Name))
			}
			if fi.NumRet != 1 || fi.RetType != "Completed" {
				problems = append(problems, fmt.Sprintf("%s:%d %s: must return single Completed (got %d returns / %q)", fi.File, fi.Line, fi.Name, fi.NumRet, fi.RetType))
			}
		}
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "scanned %d builder funcs across %d files\n", len(funcs), len(files))
	}

	if len(problems) > 0 {
		for _, p := range problems {
			fmt.Fprintln(os.Stderr, "  - "+p)
		}
		fail("found %d violation(s)", len(problems))
	}
	fmt.Fprintf(os.Stderr, "ok: %d builder funcs across %d files passed invariants\n", len(funcs), len(files))
}

func discover(root string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// 仅扫描根目录，不进入子目录（cmd/、internal/ 等）。
			if path == root {
				return nil
			}
			return fs.SkipDir
		}
		name := filepath.Base(path)
		if !strings.HasPrefix(name, builderFilePrefix) || !strings.HasSuffix(name, ".go") {
			return nil
		}
		if strings.HasSuffix(name, "_test.go") {
			return nil
		}
		out = append(out, path)
		return nil
	})
	sort.Strings(out)
	return out, err
}

func parseFile(fset *token.FileSet, path string) ([]funcInfo, error) {
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	var out []funcInfo
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if fn.Recv == nil || len(fn.Recv.List) == 0 {
			continue
		}
		recvType := exprName(fn.Recv.List[0].Type)
		fi := funcInfo{
			File:     path,
			Line:     fset.Position(fn.Pos()).Line,
			Name:     fn.Name.Name,
			Recv:     recvType,
			NumParam: countFields(fn.Type.Params),
		}
		if fn.Type.Results != nil {
			fi.NumRet = countFields(fn.Type.Results)
			if fi.NumRet == 1 {
				fi.RetType = exprName(fn.Type.Results.List[0].Type)
			}
		}
		out = append(out, fi)
	}
	return out, nil
}

func exprName(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + exprName(t.X)
	case *ast.SelectorExpr:
		return exprName(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + exprName(t.Elt)
	case *ast.Ellipsis:
		return "..." + exprName(t.Elt)
	}
	return fmt.Sprintf("%T", e)
}

func countFields(fl *ast.FieldList) int {
	if fl == nil {
		return 0
	}
	n := 0
	for _, f := range fl.List {
		if len(f.Names) == 0 {
			n++
			continue
		}
		n += len(f.Names)
	}
	return n
}

func isExported(s string) bool {
	if s == "" {
		return false
	}
	c := s[0]
	return c >= 'A' && c <= 'Z'
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "genbuilder: "+format+"\n", args...)
	os.Exit(1)
}
