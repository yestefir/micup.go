package micup_test

import (
	"flag"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

var updateAPI = flag.Bool("update-api", false, "rewrite testdata/api.golden from the facades")

func TestPublicAPIIsStable(t *testing.T) {
	var names []string
	for _, pkg := range []struct{ prefix, dir string }{{"micup", "."}, {"voice", "voice"}} {
		files, err := filepath.Glob(filepath.Join(pkg.dir, "*.go"))
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range files {
			if strings.HasSuffix(path, "_test.go") {
				continue
			}
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			names = append(names, exportedNames(pkg.prefix, file)...)
		}
	}
	sort.Strings(names)
	got := strings.Join(names, "\n") + "\n"
	golden := filepath.Join("testdata", "api.golden")
	if *updateAPI {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("%v; run go test -run TestPublicAPIIsStable -update-api .", err)
	}
	have := map[string]bool{}
	for _, name := range names {
		have[name] = true
	}
	recorded := map[string]bool{}
	for _, name := range strings.Fields(string(want)) {
		recorded[name] = true
		if !have[name] {
			t.Errorf("public name %s was removed", name)
		}
	}
	for _, name := range names {
		if !recorded[name] {
			t.Errorf("new public name %s is not recorded; run go test -run TestPublicAPIIsStable -update-api .", name)
		}
	}
}

func exportedNames(prefix string, file *ast.File) []string {
	var names []string
	add := func(kind, name string) {
		if ast.IsExported(name) {
			names = append(names, prefix+"."+kind+"."+name)
		}
	}
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Recv == nil {
				add("func", d.Name.Name)
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					add("type", s.Name.Name)
				case *ast.ValueSpec:
					kind := "var"
					if d.Tok == token.CONST {
						kind = "const"
					}
					for _, n := range s.Names {
						add(kind, n.Name)
					}
				}
			}
		}
	}
	return names
}
