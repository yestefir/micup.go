package state

import (
	"encoding/json"
	"testing"

	"github.com/yestefir/micup.go/src/gateway"
	"github.com/yestefir/micup.go/src/model"
)

func raw(t *testing.T, v any) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestCacheTracksGuildsChannelsAndVoice(t *testing.T) {
	c := New("bot")
	c.Apply(gateway.EventReady, raw(t, map[string]any{"guilds": []any{map[string]any{"id": "g", "channels": []map[string]string{{"id": "voice"}}, "voice_states": []map[string]string{{"user_id": "u", "channel_id": "voice"}}}}}))
	if guild, ok := c.GuildForChannel("voice"); !ok || guild != "g" || c.VoiceChannel("g", "u") != "voice" {
		t.Fatal("READY was not applied")
	}
	c.Apply(gateway.EventChannelCreate, raw(t, map[string]string{"id": "text", "guild_id": "g"}))
	c.Apply(gateway.EventChannelDelete, raw(t, map[string]string{"id": "voice"}))
	if _, ok := c.GuildForChannel("voice"); ok || c.Channels() != 1 {
		t.Fatal("channel events were not applied")
	}
	move := c.Apply(gateway.EventVoiceStateUpdate, raw(t, map[string]string{"guild_id": "g", "user_id": "bot", "channel_id": "voice"}))
	if move == nil || move.Previous != "" || move.Current != "voice" || c.BotVoiceChannel("g") != "voice" {
		t.Fatalf("bot voice move not reported: %+v", move)
	}
	if c.Apply(gateway.EventVoiceStateUpdate, raw(t, map[string]string{"guild_id": "g", "user_id": "u", "channel_id": ""})) != nil || c.VoiceChannel("g", "u") != "" {
		t.Fatal("user voice leave was not applied")
	}
	c.Apply(gateway.EventGuildDelete, raw(t, map[string]any{"id": "g", "unavailable": true}))
	if c.Guilds() != 1 || c.Channels() != 1 {
		t.Fatal("temporary outage dropped cached guild data")
	}
	c.Apply(gateway.EventGuildDelete, raw(t, map[string]string{"id": "g"}))
	if c.Guilds() != 0 || c.Channels() != 0 {
		t.Fatal("guild delete left cached data")
	}
}

func TestCacheComputesBotPermissions(t *testing.T) {
	c := New("bot")
	c.Apply(gateway.EventGuildCreate, raw(t, map[string]any{
		"id":         "g",
		"properties": map[string]string{"owner_id": "owner"},
		"roles":      []map[string]any{{"id": "g", "permissions": "1049600"}, {"id": "dj", "permissions": "2097152"}},
		"channels":   []map[string]any{{"id": "open", "type": 2}, {"id": "locked", "type": 2, "permission_overwrites": []map[string]any{{"id": "g", "type": 0, "allow": "0", "deny": "1048576"}}}},
		"members":    []map[string]any{{"user": map[string]string{"id": "bot"}, "roles": []string{}}},
	}))
	if p, ok := c.BotPermissions("g", "open"); !ok || !p.Has(model.PermissionConnect) {
		t.Fatalf("bot cannot connect to an open channel: %v %b", ok, p)
	}
	if p, ok := c.BotPermissions("g", "locked"); !ok || p.Has(model.PermissionConnect) {
		t.Fatalf("channel overwrite was ignored: %b", p)
	}
	if p, ok := c.Permissions("g", "locked", "owner"); !ok || p != model.PermissionAll {
		t.Fatal("owner was not resolved from properties")
	}
	if _, ok := c.Permissions("g", "open", "stranger"); ok {
		t.Fatal("unknown member reported known permissions")
	}
	c.Apply(gateway.EventGuildMemberUpdate, raw(t, map[string]any{"guild_id": "g", "user": map[string]string{"id": "bot"}, "roles": []string{"dj"}}))
	if p, _ := c.BotPermissions("g", "open"); !p.Has(model.PermissionSpeak) {
		t.Fatal("member role update was not applied")
	}
	c.Apply(gateway.EventGuildRoleUpdate, raw(t, map[string]any{"guild_id": "g", "role": map[string]any{"id": "dj", "permissions": "0"}}))
	if p, _ := c.BotPermissions("g", "open"); p.Has(model.PermissionSpeak) {
		t.Fatal("role update was not applied")
	}
	c.Apply(gateway.EventChannelUpdateBulk, raw(t, map[string]any{"guild_id": "g", "channels": []map[string]any{{"id": "locked", "type": 2}}}))
	if p, _ := c.BotPermissions("g", "locked"); !p.Has(model.PermissionConnect) {
		t.Fatal("bulk channel update was not applied")
	}
	c.Apply(gateway.EventMessageCreate, raw(t, map[string]any{"guild_id": "g", "author": map[string]string{"id": "u"}, "member": map[string]any{"roles": []string{"dj"}}}))
	if roles, ok := c.MemberRoles("g", "u"); !ok || len(roles) != 1 {
		t.Fatal("message member was not cached")
	}
	c.Apply(gateway.EventGuildMembersChunk, raw(t, map[string]any{"guild_id": "g", "members": []map[string]any{{"user": map[string]string{"id": "v"}, "roles": []string{}}}}))
	if _, ok := c.MemberRoles("g", "v"); !ok {
		t.Fatal("member chunk was not cached")
	}
	c.Apply(gateway.EventGuildMemberRemove, raw(t, map[string]any{"guild_id": "g", "user": map[string]string{"id": "v"}}))
	if _, ok := c.MemberRoles("g", "v"); ok {
		t.Fatal("removed member stayed cached")
	}
}

func TestCacheTracksThreadsExpressionsAndGuildDetails(t *testing.T) {
	c := New("bot")
	c.Apply(gateway.EventGuildCreate, raw(t, map[string]any{
		"id":         "5",
		"properties": map[string]any{"owner_id": "owner", "name": "Den", "features": []string{"BANNER"}},
		"roles":      []map[string]any{{"id": "5", "permissions": "3072", "position": 0}, {"id": "mod", "permissions": "0", "position": 2}, {"id": "dj", "permissions": "0", "position": 1}},
		"channels":   []map[string]any{{"id": "text", "type": 0, "position": 1, "permission_overwrites": []map[string]any{{"id": "5", "type": 0, "allow": "0", "deny": "2048"}}}},
		"threads":    []map[string]any{{"id": "t1", "type": 11, "parent_id": "text"}},
		"emojis":     []map[string]any{{"id": "e", "name": "wave"}},
		"members":    []map[string]any{{"user": map[string]string{"id": "bot"}, "roles": []string{"mod"}}, {"user": map[string]string{"id": "u"}, "roles": []string{"dj"}}},
	}))
	guild, ok := c.Guild("5")
	if !ok || guild.Name != "Den" || !guild.HasFeature("BANNER") || len(guild.Roles) != 3 || guild.Roles[0].ID != "mod" || len(c.Emojis("5")) != 1 {
		t.Fatalf("guild details: %+v", guild)
	}
	if threads := c.Threads("5"); len(threads) != 1 || len(c.GuildChannels("5")) != 1 {
		t.Fatalf("threads %v channels %v", threads, c.GuildChannels("5"))
	}
	if p, ok := c.BotPermissions("5", "t1"); !ok || p.Has(model.PermissionSendMessages) {
		t.Fatalf("thread permissions must follow the parent's overwrites: %v %b", ok, p)
	}
	if !c.CanManage("5", "bot", "u") || c.CanManage("5", "u", "bot") || c.CanManage("5", "bot", "owner") || !c.CanManage("5", "owner", "bot") {
		t.Fatal("role hierarchy checks are wrong")
	}
	c.Apply(gateway.EventThreadCreate, raw(t, map[string]any{"id": "t2", "type": 12, "guild_id": "5", "parent_id": "text", "newly_created": true}))
	c.Apply(gateway.EventThreadDelete, raw(t, map[string]any{"id": "t1", "guild_id": "5", "parent_id": "text", "type": 11}))
	if threads := c.Threads("5"); len(threads) != 1 || threads[0].ID != "t2" {
		t.Fatalf("thread events: %+v", threads)
	}
	c.Apply(gateway.EventThreadListSync, raw(t, map[string]any{"guild_id": "5", "threads": []map[string]any{{"id": "t3", "type": 11, "parent_id": "text"}}}))
	if threads := c.Threads("5"); len(threads) != 1 || threads[0].ID != "t3" || threads[0].GuildID != "5" {
		t.Fatalf("thread list sync: %+v", threads)
	}
	c.Apply(gateway.EventGuildEmojisUpdate, raw(t, map[string]any{"guild_id": "5", "emojis": []any{}}))
	c.Apply(gateway.EventGuildStickersUpdate, raw(t, map[string]any{"guild_id": "5", "stickers": []map[string]string{{"id": "s", "name": "cat"}}}))
	if len(c.Emojis("5")) != 0 || len(c.Stickers("5")) != 1 {
		t.Fatal("expression updates were not applied")
	}
	c.Apply(gateway.EventGuildUpdate, raw(t, map[string]any{"guild_id": "5", "id": "5", "name": "Den 2", "owner_id": "owner"}))
	if guild, _ := c.Guild("5"); guild.Name != "Den 2" || len(guild.Roles) != 3 {
		t.Fatalf("guild update: %+v", guild)
	}
	c.Apply(gateway.EventChannelDelete, raw(t, map[string]string{"id": "text", "guild_id": "5"}))
	if len(c.Threads("5")) != 0 {
		t.Fatal("deleting a channel must drop its threads")
	}
	if ids := c.GuildIDs(); len(ids) != 1 || ids[0] != "5" {
		t.Fatalf("guild IDs: %v", ids)
	}
}
