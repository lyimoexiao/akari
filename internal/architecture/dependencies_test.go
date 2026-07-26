package architecture_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

var forbiddenImports = map[string]struct{}{
	"github.com/casbin/casbin/v3":                {},
	"github.com/casbin/casbin/v3/model":          {},
	"github.com/lyimoexiao/akari/pkg/cache": {},
	"gorm.io/gorm": {},
}

func Test_Business_packages_do_not_import_concrete_infrastructure(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve architecture test path")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
	packages := []string{"permission", "role", "user", "requestlog", "router"}
	for _, packageName := range packages {
		files, err := filepath.Glob(filepath.Join(root, "internal", packageName, "*.go"))
		if err != nil {
			t.Fatalf("list %s files: %v", packageName, err)
		}
		for _, path := range files {
			if strings.HasSuffix(path, "_test.go") {
				continue
			}
			checkFile(t, path)
		}
	}
}

func checkFile(t *testing.T, path string) {
	t.Helper()
	fileSet := token.NewFileSet()
	parsed, err := parser.ParseFile(fileSet, path, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	for _, imported := range parsed.Imports {
		importPath, err := strconv.Unquote(imported.Path.Value)
		if err != nil {
			t.Fatalf("parse import in %s: %v", path, err)
		}
		if _, forbidden := forbiddenImports[importPath]; forbidden {
			t.Errorf("%s imports concrete infrastructure %s", path, importPath)
		}
	}
	ast.Inspect(parsed, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || len(call.Args) == 0 {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "Set" {
			return true
		}
		key, ok := call.Args[0].(*ast.BasicLit)
		if !ok || key.Kind != token.STRING {
			return true
		}
		value, err := strconv.Unquote(key.Value)
		if err == nil && (value == "db" || value == "cache") {
			t.Errorf("%s injects concrete dependency through context key %q", path, value)
		}
		return true
	})
}
