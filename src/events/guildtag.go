package events

import (
	"context"
	"sync"

	"github.com/yestefir/micup.go/src/gateway"
	"github.com/yestefir/micup.go/src/model"
)

type GuildTagChange struct {
	UserID   string
	GuildID  string
	Previous string
	Current  string
}

func (c GuildTagChange) Cleared() bool { return c.Current == "" }

type guildTagTracker struct {
	mu   sync.Mutex
	tags map[string]string
}

func (t *guildTagTracker) record(user *model.User) (previous, current string, known bool) {
	if user == nil || user.ID == "" {
		return "", "", false
	}
	if user.TagGuildID != nil {
		current = *user.TagGuildID
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	previous = t.tags[user.ID]
	if current == "" {
		delete(t.tags, user.ID)
	} else {
		t.tags[user.ID] = current
	}
	return previous, current, true
}

func (m *Mux) OnGuildTagChange(fn func(context.Context, GuildTagChange) error) {
	tracker := &guildTagTracker{tags: map[string]string{}}
	notify := func(ctx context.Context, guildID string, user *model.User) error {
		previous, current, ok := tracker.record(user)
		if !ok || previous == current {
			return nil
		}
		return fn(ctx, GuildTagChange{UserID: user.ID, GuildID: guildID, Previous: previous, Current: current})
	}
	member := func(ctx context.Context, event gateway.Event) error {
		value, err := As[GuildMember](event)
		if err != nil {
			return err
		}
		return notify(ctx, value.GuildID, value.User)
	}
	m.HandleRaw(gateway.EventGuildMemberAdd, member)
	m.HandleRaw(gateway.EventGuildMemberUpdate, member)
	m.HandleRaw(gateway.EventUserUpdate, func(ctx context.Context, event gateway.Event) error {
		user, err := As[model.User](event)
		if err != nil {
			return err
		}
		return notify(ctx, "", &user)
	})
	m.HandleRaw(gateway.EventGuildMembersChunk, func(_ context.Context, event gateway.Event) error {
		chunk, err := As[MembersChunk](event)
		if err != nil {
			return err
		}
		for _, value := range chunk.Members {
			tracker.record(value.User)
		}
		return nil
	})
}
