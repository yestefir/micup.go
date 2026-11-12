package layers

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const module = "github.com/yestefir/micup.go/src/"

var allowed = map[string][]string{
	"model":              {},
	"internal/version":   {},
	"internal/httpx":     {},
	"internal/anonymize": {},
	"discovery":          {"internal/httpx"},
	"rest":               {"model", "internal/httpx", "internal/version"},
	"gateway":            {"internal/version", "model"},
	"events":             {"gateway", "model"},
	"state":              {"events", "gateway", "model"},
	"session":            {"events", "gateway", "model", "state"},
	"shard":              {"events", "gateway", "model", "session"},
	"voice":              {},
	"bot":                {"discovery", "gateway", "internal/httpx", "internal/version", "model", "rest", "session", "shard"},
}

func TestPackageLayers(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		rel, _ := filepath.Rel(root, filepath.Dir(path))
		pkg := filepath.ToSlash(rel)
		if pkg == "internal/fixture" || pkg == "internal/layers" {
			return nil
		}
		rules, known := allowed[pkg]
		if !known {
			t.Errorf("package %s has no layer rule; add it to allowed", pkg)
			return nil
		}
		seen[pkg] = true
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, spec := range file.Imports {
			imported, _ := strconv.Unquote(spec.Path.Value)
			if !strings.HasPrefix(imported, module) {
				continue
			}
			target := strings.TrimPrefix(imported, module)
			ok := false
			for _, rule := range rules {
				ok = ok || rule == target
			}
			if !ok {
				t.Errorf("%s must not import %s (%s)", pkg, target, filepath.Base(path))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for pkg := range allowed {
		if !seen[pkg] {
			t.Errorf("layer rule for %s has no package", pkg)
		}
	}
}

func TestFacadesOnlyForward(t *testing.T) {
	for _, dir := range []string{"../../..", "../../../voice"} {
		matches, err := filepath.Glob(filepath.Join(dir, "*.go"))
		if err != nil {
			t.Fatal(err)
		}
		if len(matches) == 0 {
			t.Fatalf("facade %s has no Go files", dir)
		}
		for _, path := range matches {
			if strings.HasSuffix(path, "_test.go") {
				continue
			}
			checkFacade(t, path)
		}
	}
}

func checkFacade(t *testing.T, path string) {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, spec := range file.Imports {
		imported, _ := strconv.Unquote(spec.Path.Value)
		if strings.HasPrefix(imported, module+"internal/") {
			t.Errorf("%s must not import %s", path, imported)
		}
	}
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok && !ts.Assign.IsValid() {
					t.Errorf("%s: type %s must be an alias (=) of a src type", path, ts.Name.Name)
				}
			}
		case *ast.FuncDecl:
			if d.Recv != nil || d.Body == nil || len(d.Body.List) != 1 {
				t.Errorf("%s: func %s must be a single forwarding statement", path, d.Name.Name)
				continue
			}
			ret, ok := d.Body.List[0].(*ast.ReturnStmt)
			if !ok || len(ret.Results) != 1 {
				t.Errorf("%s: func %s must only return a src call", path, d.Name.Name)
				continue
			}
			if _, ok := ret.Results[0].(*ast.CallExpr); !ok {
				t.Errorf("%s: func %s must only return a src call", path, d.Name.Name)
			}
		}
	}
}
