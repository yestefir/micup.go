package session

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/yestefir/micup.go/src/model"
)

func inventoryFullPage(page int) []model.Member {
	members := make([]model.Member, 1000)
	for i := range members {
		members[i] = statMember(fmt.Sprint(page*1000+i), false)
	}
	return members
}

func TestMemberInventoryStopsObsoleteEnumeration(t *testing.T) {
	for _, change := range []struct{ name, event, body string }{
		{"ready", "READY", `{"guilds":[{"id":"g"}]}`},
		{"guild_sync", "GUILD_SYNC", `{"id":"g"}`},
		{"deleted", "GUILD_DELETE", `{"id":"g"}`},
		{"unavailable", "GUILD_DELETE", `{"id":"g","unavailable":true}`},
	} {
		t.Run(change.name, func(t *testing.T) {
			inventory := &memberInventory{}
			inventory.observe("READY", json.RawMessage(`{"guilds":[{"id":"g"}]}`))
			calls := 0
			inventory.syncOne(context.Background(), func(context.Context, string, string) ([]model.Member, error) {
				calls++
				if calls == 1 {
					// Change lifecycle while the first lookup is in progress.
					inventory.observe(change.event, json.RawMessage(change.body))
				}
				if calls == 3 {
					return nil, nil // Bound the baseline; do not emulate an endless service.
				}
				return inventoryFullPage(calls), nil
			})
			if calls != 1 {
				t.Errorf("obsolete enumeration made %d lookups, want only the in-flight lookup", calls)
			}
			if current := inventory.guilds["g"]; current != nil && current.complete {
				t.Error("obsolete enumeration published a complete inventory")
			}
		})
	}
}

func TestMemberInventoryStopsLookupsWithContext(t *testing.T) {
	for _, before := range []bool{true, false} {
		name, wantCalls := "during_page", 1
		if before {
			name, wantCalls = "before_start", 0
		}
		t.Run(name, func(t *testing.T) {
			inventory := &memberInventory{}
			inventory.observe("READY", json.RawMessage(`{"guilds":[{"id":"g"}]}`))
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if before {
				cancel()
			}
			calls := 0
			inventory.syncOne(ctx, func(context.Context, string, string) ([]model.Member, error) {
				calls++
				cancel()
				if calls == 1 {
					// A custom lookup may finish successfully as its context ends.
					return inventoryFullPage(1), nil
				}
				return nil, nil
			})
			if calls != wantCalls {
				t.Errorf("made %d lookups, want %d", calls, wantCalls)
			}
			if current := inventory.guilds["g"]; current.complete || current.syncing {
				t.Error("cancelled enumeration was completed or left syncing")
			}
		})
	}
}

func TestMemberInventoryDoesNotPublishObsoleteLastPage(t *testing.T) {
	for _, cancelled := range []bool{false, true} {
		name := "unavailable"
		if cancelled {
			name = "cancelled"
		}
		t.Run(name, func(t *testing.T) {
			inventory := &memberInventory{}
			inventory.observe("READY", json.RawMessage(`{"guilds":[{"id":"g"}]}`))
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			inventory.syncOne(ctx, func(context.Context, string, string) ([]model.Member, error) {
				if cancelled {
					cancel()
				} else {
					inventory.observe("GUILD_DELETE", json.RawMessage(`{"id":"g","unavailable":true}`))
				}
				return []model.Member{statMember("obsolete", false)}, nil
			})
			if current := inventory.guilds["g"]; current.complete || current.syncing {
				t.Error("obsolete last page was completed or left syncing")
			}
		})
	}
}
