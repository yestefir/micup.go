package gateway

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
	"testing"
)

func TestKnownEventsListsEveryName(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "names.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	listed := map[string]bool{}
	for _, name := range KnownEvents() {
		if listed[name] {
			t.Errorf("%s is listed twice", name)
		}
		listed[name] = true
	}
	declared := 0
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			value := spec.(*ast.ValueSpec)
			for i, ident := range value.Names {
				if !strings.HasPrefix(ident.Name, "Event") {
					continue
				}
				declared++
				lit, _ := value.Values[i].(*ast.BasicLit)
				name, _ := strconv.Unquote(lit.Value)
				if !listed[name] {
					t.Errorf("%s (%s) is missing from knownEvents", ident.Name, name)
				}
			}
		}
	}
	if declared != len(listed) {
		t.Errorf("names.go declares %d events, knownEvents lists %d", declared, len(listed))
	}
}

func TestKnownEventsReturnsACopy(t *testing.T) {
	KnownEvents()[0] = "CHANGED"
	if KnownEvents()[0] != EventReady {
		t.Fatal("KnownEvents exposed its backing slice")
	}
}
