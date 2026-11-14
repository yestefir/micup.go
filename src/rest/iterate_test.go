package rest

import (
	"context"
	"fmt"
	"iter"
	"net/http"
	"slices"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/yestefir/micup.go/src/model"
)

func checkIteratorIDs[T any](t *testing.T, seq iter.Seq2[T, error], id func(T) string, want []string) {
	t.Helper()
	var got []string
	for item, err := range seq {
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, id(item))
	}
	if !slices.Equal(got, want) {
		t.Errorf("iterator returned %v, want %v", got, want)
	}
}

func TestMessageIteratorCanBeReused(t *testing.T) {
	for _, ascending := range []bool{false, true} {
		t.Run(fmt.Sprint(ascending), func(t *testing.T) {
			c := newClient(t, func(w http.ResponseWriter, r *http.Request) {
				q := r.URL.Query()
				switch {
				case ascending && q.Get("after") == "5", !ascending && q.Get("before") == "40":
					fmt.Fprint(w, `[{"id":"30"},{"id":"20"}]`)
				case ascending && q.Get("after") == "30":
					fmt.Fprint(w, `[{"id":"40"}]`)
				case !ascending && q.Get("before") == "20":
					fmt.Fprint(w, `[{"id":"10"}]`)
				default:
					t.Errorf("unexpected query: %s", r.URL.RawQuery)
					fmt.Fprint(w, `[]`)
				}
			})
			query := model.MessageQuery{Limit: 2, Before: "40"}
			want := []string{"30", "20", "10"}
			if ascending {
				query.Before, query.After = "", "5"
				want = []string{"20", "30", "40"}
			}
			seq := c.EachMessage(context.Background(), "1", query)
			id := func(m model.Message) string { return m.ID }
			for range 2 {
				checkIteratorIDs(t, seq, id, want)
			}
			concurrentSeq := c.EachMessage(context.Background(), "1", query)
			for i := range 8 {
				t.Run(fmt.Sprintf("concurrent-%d", i), func(t *testing.T) {
					t.Parallel()
					checkIteratorIDs(t, concurrentSeq, id, want)
				})
			}
		})
	}
}

func TestAuditLogIteratorCanBeReused(t *testing.T) {
	c := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("before") == "40" {
			fmt.Fprint(w, `{"audit_log_entries":[{"id":"30"},{"id":"20"}]}`)
		} else {
			fmt.Fprint(w, `{"audit_log_entries":[{"id":"10"}]}`)
		}
	})
	seq := c.EachAuditLogEntry(context.Background(), "1", model.AuditLogQuery{Limit: 2, Before: "40"})
	id := func(e model.AuditLogEntry) string { return e.ID }
	want := []string{"30", "20", "10"}
	for range 2 {
		checkIteratorIDs(t, seq, id, want)
	}
	concurrentSeq := c.EachAuditLogEntry(context.Background(), "1", model.AuditLogQuery{Limit: 2, Before: "40"})
	for i := range 8 {
		t.Run(fmt.Sprintf("concurrent-%d", i), func(t *testing.T) {
			t.Parallel()
			checkIteratorIDs(t, concurrentSeq, id, want)
		})
	}
}

func TestIteratorsStopOnRepeatedPage(t *testing.T) {
	for _, resource := range []string{"messages-before", "messages-after", "audit-log", "guilds"} {
		t.Run(resource, func(t *testing.T) {
			var calls atomic.Int64
			c := newClient(t, func(w http.ResponseWriter, r *http.Request) {
				if calls.Add(1) > 2 {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				switch resource {
				case "audit-log":
					fmt.Fprint(w, `{"audit_log_entries":[{"id":"9"}]}`)
				case "guilds":
					fmt.Fprint(w, "["+strings.Repeat(`{"id":"9"},`, maxGuildsPerPage-1)+`{"id":"9"}]`)
				default:
					fmt.Fprint(w, `[{"id":"9"}]`)
				}
			})
			ctx := context.Background()
			switch resource {
			case "audit-log":
				for _, err := range c.EachAuditLogEntry(ctx, "1", model.AuditLogQuery{Limit: 1}) {
					if err != nil {
						t.Errorf("iteration did not stop before server cutoff: %v", err)
					}
				}
			case "guilds":
				for _, err := range c.EachGuild(ctx) {
					if err != nil {
						t.Errorf("iteration did not stop before server cutoff: %v", err)
					}
				}
			default:
				query := model.MessageQuery{Limit: 1}
				if resource == "messages-after" {
					query.After = "5"
				}
				for _, err := range c.EachMessage(ctx, "1", query) {
					if err != nil {
						t.Errorf("iteration did not stop before server cutoff: %v", err)
					}
				}
			}
			if got := calls.Load(); got != 2 {
				t.Errorf("sent %d requests for a repeated page, want 2", got)
			}
		})
	}
}

func TestGuildIteratorFollowsAdvancingCursor(t *testing.T) {
	var calls atomic.Int64
	c := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		switch r.URL.Query().Get("after") {
		case "":
			guilds := make([]string, maxGuildsPerPage)
			for i := range guilds {
				guilds[i] = fmt.Sprintf(`{"id":"%d"}`, i+1)
			}
			fmt.Fprint(w, "["+strings.Join(guilds, ",")+"]")
		case fmt.Sprint(maxGuildsPerPage):
			fmt.Fprintf(w, `[{"id":"%d"}]`, maxGuildsPerPage+1)
		default:
			t.Errorf("unexpected cursor: %s", r.URL.RawQuery)
			fmt.Fprint(w, `[]`)
		}
	})
	want := make([]string, maxGuildsPerPage+1)
	for i := range want {
		want[i] = fmt.Sprint(i + 1)
	}
	checkIteratorIDs(t, c.EachGuild(context.Background()), func(g model.Guild) string { return g.ID }, want)
	if got := calls.Load(); got != 2 {
		t.Errorf("sent %d requests, want 2", got)
	}
}
