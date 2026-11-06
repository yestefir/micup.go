package micup_test

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yestefir/micup.go/src/events"
	"github.com/yestefir/micup.go/src/gateway"
)

var updateDocs = flag.Bool("update-docs", false, "rewrite the generated tables in docs/")

const (
	eventsDoc   = "docs/events.md"
	eventsStart = "<!-- events:start -->\n"
	eventsEnd   = "<!-- events:end -->"
)

func eventsTable(t *testing.T) string {
	t.Helper()
	facade, err := os.ReadFile("events.go")
	if err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	b.WriteString("| Event | Constant | Payload | Captured |\n|---|---|---|---|\n")
	for _, name := range gateway.KnownEvents() {
		typ, ok := events.PayloadType(name)
		if !ok {
			t.Fatalf("%s has no payload type", name)
		}
		captured, _ := filepath.Glob(filepath.Join("src", "events", "testdata", name+".*.json"))
		mark := "no"
		if len(captured) > 0 {
			mark = "yes"
		}
		if !strings.Contains(string(facade), "\t"+constantName(name)+" ") {
			t.Fatalf("facade has no constant %s for %s", constantName(name), name)
		}
		fmt.Fprintf(&b, "| `%s` | `micup.%s` | `micup.%s` | %s |\n", name, constantName(name), typ.Name(), mark)
	}
	return b.String()
}

func constantName(name string) string {
	switch name {
	case gateway.EventReactionAdd, gateway.EventReactionRemove, gateway.EventReactionRemoveAll, gateway.EventReactionRemoveEmoji:
		name = strings.TrimPrefix(name, "MESSAGE_")
	}
	var b strings.Builder
	b.WriteString("Event")
	for _, part := range strings.Split(name, "_") {
		b.WriteString(part[:1] + strings.ToLower(part[1:]))
	}
	return b.String()
}

func TestEventDocsAreCurrent(t *testing.T) {
	data, err := os.ReadFile(eventsDoc)
	if err != nil {
		t.Fatal(err)
	}
	doc := strings.ReplaceAll(string(data), "\r\n", "\n")
	start := strings.Index(doc, eventsStart)
	end := strings.Index(doc, eventsEnd)
	if start < 0 || end < start {
		t.Fatalf("%s lacks the events markers", eventsDoc)
	}
	want := doc[:start+len(eventsStart)] + eventsTable(t) + doc[end:]
	if *updateDocs {
		if err := os.WriteFile(eventsDoc, []byte(want), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	if doc != want {
		t.Fatalf("%s is out of date; run go test -run TestEventDocsAreCurrent -update-docs .", eventsDoc)
	}
}
