package session

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/yestefir/micup.go/src/model"
)

func statMember(id string, bot bool) model.Member {
	return model.Member{User: &model.User{ID: id, Bot: bot}}
}

func TestMemberInventoryEnumerationReconcilesEventsAndPagination(t *testing.T) {
	s := &memberInventory{}
	observe := func(name, body string) { s.observe(name, json.RawMessage(body)) }
	observe("READY", `{"guilds":[{"id":"g"}]}`)
	pages := 0
	page := func(_ context.Context, guild, after string) ([]model.Member, error) {
		pages++
		if guild != "g" {
			t.Fatal(guild)
		}
		if after == "" {
			observe("GUILD_MEMBER_ADD", `{"guild_id":"g","user":{"id":"new","bot":false}}`)
			observe("GUILD_MEMBER_REMOVE", `{"guild_id":"g","user":{"id":"0"}}`)
			members := make([]model.Member, 1000)
			for i := range members {
				members[i] = statMember(fmt.Sprint(i), false)
			}
			return members, nil
		}
		if after != "999" {
			t.Fatalf("wrong cursor %s", after)
		}
		return []model.Member{statMember("other-bot", true)}, nil
	}
	if !s.syncOne(context.Background(), page) {
		t.Fatal("no enumeration")
	}
	g := s.guilds["g"]
	if !g.complete || len(g.members) != 1001 || pages != 2 {
		t.Fatalf("inventory: %+v pages=%d", g, pages)
	}
	if _, ok := g.members["0"]; ok {
		t.Fatal("removed member resurrected")
	}
	if g.members["other-bot"] != true {
		t.Fatal("bot identity lost")
	}
	if _, ok := g.members["new"]; !ok {
		t.Fatal("concurrent addition lost")
	}
	observe("GUILD_MEMBER_ADD", `{"guild_id":"g","user":{"id":"new","bot":false}}`)
	if len(g.members) != 1001 {
		t.Fatal("duplicate event double-counted")
	}
}

func TestMemberInventoryResetAndUnavailableDoNotPublishOldEnumeration(t *testing.T) {
	for _, reset := range []string{`{"guilds":[{"id":"g"}]}`, `{"guilds":[]}`} {
		s := &memberInventory{}
		s.observe("READY", json.RawMessage(`{"guilds":[{"id":"g"}]}`))
		s.syncOne(context.Background(), func(context.Context, string, string) ([]model.Member, error) {
			s.observe("READY", json.RawMessage(reset))
			return []model.Member{statMember("old", false)}, nil
		})
		if g := s.guilds["g"]; g != nil && (g.complete || len(g.members) != 0) {
			t.Fatal("old enumeration crossed READY epoch")
		}
	}
	s := &memberInventory{}
	s.observe("GUILD_CREATE", json.RawMessage(`{"id":"g"}`))
	s.observe("GUILD_DELETE", json.RawMessage(`{"id":"g","unavailable":true}`))
	if s.guilds["g"].available {
		t.Fatal("unavailable guild removed or counted available")
	}
	s.observe("GUILD_DELETE", json.RawMessage(`{"id":"g"}`))
	if len(s.guilds) != 0 {
		t.Fatal("left guild retained")
	}
}

func TestMemberInventoryFailedEnumerationRemainsIncomplete(t *testing.T) {
	s := &memberInventory{}
	s.observe("READY", json.RawMessage(`{"guilds":[{"id":"g"}]}`))
	s.syncOne(context.Background(), func(context.Context, string, string) ([]model.Member, error) {
		return nil, context.DeadlineExceeded
	})
	if s.guilds["g"].complete || s.guilds["g"].syncing {
		t.Fatal("failed enumeration marked complete")
	}
	if s.syncOne(context.Background(), nil) {
		t.Fatal("retry backoff ignored")
	}
}

func TestMemberInventorySnapshotIsIndependent(t *testing.T) {
	s := &Session{}
	s.inventory.observe("READY", json.RawMessage(`{"guilds":[{"id":"g","members":[{"user":{"id":"u","bot":true}}]}]}`))
	first := s.MemberInventories()
	if len(first) != 1 || first[0].GuildID != "g" || !first[0].Available || first[0].Complete || !first[0].Members["u"] {
		t.Fatalf("unexpected snapshot: %+v", first)
	}
	delete(first[0].Members, "u")
	if !s.MemberInventories()[0].Members["u"] {
		t.Fatal("caller changed inventory")
	}
}

func TestMemberInventoryRejectsInvalidPage(t *testing.T) {
	s := &memberInventory{}
	s.observe("READY", json.RawMessage(`{"guilds":[{"id":"g"}]}`))
	s.syncOne(context.Background(), func(context.Context, string, string) ([]model.Member, error) {
		return []model.Member{{}}, nil
	})
	if s.guilds["g"].complete {
		t.Fatal("malformed member accepted as a complete inventory")
	}
}
