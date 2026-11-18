package events

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/yestefir/micup.go/src/gateway"
)

func memberEvent(t *testing.T, name, guildID, userID string, tag any) gateway.Event {
	t.Helper()
	user := map[string]any{"id": userID, "username": "u", "tag_guild_id": tag}
	raw, err := json.Marshal(map[string]any{"guild_id": guildID, "user": user, "roles": []string{}})
	if err != nil {
		t.Fatal(err)
	}
	return gateway.Event{Name: name, Data: raw}
}

func TestGuildTagChangeReportsOnlyRealChanges(t *testing.T) {
	var got []GuildTagChange
	m := NewMux()
	m.OnGuildTagChange(func(_ context.Context, change GuildTagChange) error {
		got = append(got, change)
		return nil
	})
	send := func(event gateway.Event) {
		t.Helper()
		if err := m.Handle(context.Background(), event); err != nil {
			t.Fatal(err)
		}
	}
	send(memberEvent(t, gateway.EventGuildMemberUpdate, "g", "u1", nil))
	send(memberEvent(t, gateway.EventGuildMemberUpdate, "g", "u1", "t1"))
	send(memberEvent(t, gateway.EventGuildMemberUpdate, "g", "u1", "t1"))
	send(memberEvent(t, gateway.EventGuildMemberUpdate, "g", "u1", "t2"))
	send(memberEvent(t, gateway.EventGuildMemberUpdate, "g", "u1", nil))
	want := []GuildTagChange{
		{UserID: "u1", GuildID: "g", Previous: "", Current: "t1"},
		{UserID: "u1", GuildID: "g", Previous: "t1", Current: "t2"},
		{UserID: "u1", GuildID: "g", Previous: "t2", Current: ""},
	}
	if len(got) != len(want) {
		t.Fatalf("changes = %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("change %d = %+v, want %+v", i, got[i], want[i])
		}
	}
	if !got[2].Cleared() || got[0].Cleared() {
		t.Error("Cleared reported the wrong state")
	}
}

func TestGuildTagChangeUsesChunksAsBaselineAndTracksUserUpdates(t *testing.T) {
	var got []GuildTagChange
	m := NewMux()
	m.OnGuildTagChange(func(_ context.Context, change GuildTagChange) error {
		got = append(got, change)
		return nil
	})
	chunk, _ := json.Marshal(map[string]any{"guild_id": "g", "members": []map[string]any{{"user": map[string]any{"id": "u2", "username": "u", "tag_guild_id": "t1"}, "roles": []string{}}}})
	if err := m.Handle(context.Background(), gateway.Event{Name: gateway.EventGuildMembersChunk, Data: chunk}); err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("chunk produced changes: %+v", got)
	}
	if err := m.Handle(context.Background(), memberEvent(t, gateway.EventGuildMemberUpdate, "g", "u2", "t1")); err != nil || len(got) != 0 {
		t.Fatalf("unchanged tag after a chunk reported %+v, %v", got, err)
	}
	self, _ := json.Marshal(map[string]any{"id": "bot", "username": "b", "tag_guild_id": "t9"})
	if err := m.Handle(context.Background(), gateway.Event{Name: gateway.EventUserUpdate, Data: self}); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].UserID != "bot" || got[0].GuildID != "" || got[0].Current != "t9" {
		t.Fatalf("user update change = %+v", got)
	}
}

func TestGuildTagChangeTrackersAreIndependent(t *testing.T) {
	counts := [2]int{}
	m := NewMux()
	for i := range counts {
		m.OnGuildTagChange(func(context.Context, GuildTagChange) error { counts[i]++; return nil })
	}
	if err := m.Handle(context.Background(), memberEvent(t, gateway.EventGuildMemberUpdate, "g", "u", "t")); err != nil {
		t.Fatal(err)
	}
	if counts != [2]int{1, 1} {
		t.Fatalf("counts = %v, want both listeners notified", counts)
	}
}
