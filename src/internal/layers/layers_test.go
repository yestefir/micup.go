package layers

import (
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
	"model":            {},
	"internal/version": {},
	"internal/httpx":   {},
	"discovery":        {"internal/httpx"},
	"rest":             {"model", "internal/httpx", "internal/version"},
	"gateway":          {"internal/version"},
	"state":            {"gateway", "model"},
	"session":          {"gateway", "model", "state"},
	"shard":            {"gateway", "model", "session"},
	"voice":            {},
	"bot":              {"discovery", "gateway", "internal/httpx", "internal/version", "rest", "session", "shard"},
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
