package state

import (
	"encoding/json"
	"reflect"
	"sync"
	"testing"

	"github.com/yestefir/micup.go/src/gateway"
	"github.com/yestefir/micup.go/src/model"
)

func extendedSnapshotCache(t testing.TB) *Cache {
	t.Helper()
	var guild model.Guild
	if err := json.Unmarshal([]byte(`{"id":"guild","name":"name","icon":"icon","banner":"banner","splash":"splash","features":["feature"],"afk_channel_id":"afk","system_channel_id":"system","rules_channel_id":"rules","vanity_url_code":"vanity","embed_splash":"embed","badge_ids":["badge"],"content_warning_text":"warning","message_history_cutoff":"cutoff","permissions":"1024"}`), &guild); err != nil {
		t.Fatal(err)
	}
	var role struct{ Role model.Role }
	if err := json.Unmarshal([]byte(snapshotRole), &role); err != nil {
		t.Fatal(err)
	}
	var channel model.Channel
	if err := json.Unmarshal([]byte(snapshotChannel), &channel); err != nil {
		t.Fatal(err)
	}
	guild.Roles = []model.Role{role.Role}
	guild.Emojis = []model.Emoji{{ID: "emoji", Name: "emoji", User: &channel.Recipients[0]}}
	guild.Stickers = []model.Sticker{{ID: "sticker", Name: "sticker", Tags: []string{"tag"}, User: &channel.Recipients[0]}}
	c := New("bot")
	apply := func(name string, value any) {
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		c.Apply(name, data)
	}
	apply(gateway.EventGuildCreate, guild)
	c.Apply(gateway.EventGuildMemberAdd, json.RawMessage(`{"guild_id":"guild","user":{"id":"user"},"roles":["role"]}`))
	c.Apply(gateway.EventChannelCreate, json.RawMessage(snapshotChannel))
	channel.ID, channel.Type = "thread", model.ChannelTypePublicThread
	apply(gateway.EventThreadCreate, channel)
	return c
}

func TestExtendedCacheSnapshotsOwnMutableFields(t *testing.T) {
	for _, getter := range []struct {
		name string
		get  func(*Cache) any
		edit func(any)
	}{
		{"guild", func(c *Cache) any { v, _ := c.Guild("guild"); return v }, func(v any) {
			g := v.(model.Guild)
			for _, p := range []*string{g.Icon, g.Banner, g.Splash, g.AFKChannelID, g.SystemChannelID, g.RulesChannelID, g.VanityURLCode, g.EmbedSplash, g.ContentWarningText, g.MessageHistoryCutoff} {
				*p = "changed"
			}
			g.Features[0], g.BadgeIDs[0], *g.Permissions = "changed", "changed", 0
			g.Roles[0].Colors.PrimaryColor = 9
			*g.Emojis[0].User.Avatar = "changed"
			g.Stickers[0].Tags[0], *g.Stickers[0].User.Bio = "changed", "changed"
		}},
		{"guild_channels", func(c *Cache) any { return c.GuildChannels("guild") }, func(v any) {
			ch := v.([]model.Channel)[0]
			ch.Nicks["user"], *ch.Topic, *ch.Recipients[0].Avatar = "changed", "changed", "changed"
			ch.ThreadMetadata.Locked = true
		}},
		{"threads", func(c *Cache) any { return c.Threads("guild") }, func(v any) {
			ch := v.([]model.Channel)[0]
			ch.Nicks["user"], *ch.Topic, *ch.Recipients[0].Avatar = "changed", "changed", "changed"
			ch.ThreadMetadata.Locked = true
		}},
		{"roles", func(c *Cache) any { return c.Roles("guild") }, func(v any) { v.([]model.Role)[0].Colors.PrimaryColor = 9 }},
		{"highest_role", func(c *Cache) any { v, _ := c.HighestRole("guild", "user"); return v }, func(v any) { v.(model.Role).Colors.PrimaryColor = 9 }},
		{"emojis", func(c *Cache) any { return c.Emojis("guild") }, func(v any) { *v.([]model.Emoji)[0].User.Avatar = "changed" }},
		{"stickers", func(c *Cache) any { return c.Stickers("guild") }, func(v any) {
			s := v.([]model.Sticker)[0]
			s.Tags[0], *s.User.Avatar = "changed", "changed"
		}},
	} {
		t.Run(getter.name, func(t *testing.T) {
			c := extendedSnapshotCache(t)
			first := getter.get(c)
			requireReferenceFixture(t, reflect.ValueOf(first), getter.name)
			want, err := json.Marshal(first)
			if err != nil {
				t.Fatal(err)
			}
			getter.edit(getter.get(c))
			for _, value := range []any{first, getter.get(c)} {
				got, err := json.Marshal(value)
				if err != nil || string(got) != string(want) {
					t.Error("editing another snapshot changed cached state or an earlier snapshot")
				}
			}
		})
	}
}

func TestExtendedCacheSnapshotsCanBeEditedConcurrently(t *testing.T) {
	c := extendedSnapshotCache(t)
	var edits sync.WaitGroup
	for range 8 {
		edits.Add(1)
		go func() {
			defer edits.Done()
			for range 20 {
				c.Threads("guild")[0].ThreadMetadata.Locked = true
				c.Roles("guild")[0].Colors.PrimaryColor = 9
			}
		}()
	}
	edits.Wait()
}

func TestExtendedCacheSnapshotsPreserveAbsentFields(t *testing.T) {
	c := New("bot")
	c.Apply(gateway.EventGuildCreate, json.RawMessage(`{"id":"guild"}`))
	guild, ok := c.Guild("guild")
	if !ok || !reflect.DeepEqual(guild, model.Guild{ID: "guild", Roles: []model.Role{}}) {
		t.Fatal("minimal guild changed absent metadata or collections")
	}
	if c.Emojis("guild") != nil || c.Stickers("guild") != nil || c.GuildChannels("guild") != nil || c.Threads("guild") != nil {
		t.Fatal("absent collections became populated")
	}
	if _, ok := c.HighestRole("missing", "user"); ok {
		t.Fatal("missing guild gained a role")
	}
	c.Apply(gateway.EventGuildRoleCreate, json.RawMessage(`{"guild_id":"guild","role":{"id":"role"}}`))
	c.Apply(gateway.EventGuildMemberAdd, json.RawMessage(`{"guild_id":"guild","user":{"id":"user"},"roles":["role"]}`))
	role, ok := c.HighestRole("guild", "user")
	if !ok || !reflect.DeepEqual(role, model.Role{ID: "role"}) {
		t.Fatal("minimal highest role changed absent metadata")
	}
}

var benchmarkGuild model.Guild
var benchmarkChannels []model.Channel
var benchmarkRoles []model.Role
var benchmarkEmojis []model.Emoji
var benchmarkStickers []model.Sticker

func BenchmarkExtendedCacheSnapshot(b *testing.B) {
	c := extendedSnapshotCache(b)
	b.Run("guild", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchmarkGuild, _ = c.Guild("guild")
		}
	})
	b.Run("guild_channels", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchmarkChannels = c.GuildChannels("guild")
		}
	})
	b.Run("threads", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchmarkChannels = c.Threads("guild")
		}
	})
	b.Run("roles", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchmarkRoles = c.Roles("guild")
		}
	})
	b.Run("highest_role", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchmarkRole, _ = c.HighestRole("guild", "user")
		}
	})
	b.Run("emojis", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchmarkEmojis = c.Emojis("guild")
		}
	})
	b.Run("stickers", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			benchmarkStickers = c.Stickers("guild")
		}
	})
}
