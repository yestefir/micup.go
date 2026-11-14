package state

import (
	"encoding/json"
	"reflect"
	"sync"
	"testing"

	"github.com/yestefir/micup.go/src/gateway"
	"github.com/yestefir/micup.go/src/model"
)

const snapshotChannel = `{
	"id":"channel","guild_id":"guild","type":15,"name":"forum",
	"topic":"topic","url":"url","icon":"icon","rtc_region":"region","nsfw_override":true,
	"content_warning_text":"warning","last_pin_timestamp":"pin","permissions":"1024",
	"permission_overwrites":[{"id":"guild","type":0,"allow":"1024","deny":"0"}],
	"recipients":[{"id":"user","username":"name","global_name":"global","avatar":"avatar","avatar_color":1,"banner":"banner","banner_color":2,"accent_color":3,"bio":"bio","pronouns":"pronouns"}],
	"nicks":{"user":"nick"},"applied_tags":["tag"],"member_ids_preview":["user"],
	"thread_metadata":{"archived":false,"auto_archive_duration":60,"invitable":true},
	"member":{"user_id":"user","flags":1,"member":{"user":{"id":"user","username":"name","global_name":"global","avatar":"avatar","avatar_color":1,"banner":"banner","banner_color":2,"accent_color":3,"bio":"bio","pronouns":"pronouns"},"nick":"nick","avatar":"avatar","banner":"banner","accent_color":1,"roles":["role"],"communication_disabled_until":"until","permissions":"1024"}},
	"available_tags":[{"id":"tag","name":"tag","emoji_id":"emoji","emoji_name":"name"}],
	"default_reaction_emoji":{"emoji_id":"emoji","emoji_name":"name"},"default_sort_order":1
}`

const snapshotRole = `{"guild_id":"guild","role":{"id":"role","name":"role","permissions":"1024","colors":{"primary_color":1,"secondary_color":2,"tertiary_color":3},"hoist_position":1,"icon":"icon","unicode_emoji":"emoji","icon_emoji_id":"emoji-id"}}`

func mutateSnapshotUser(user *model.User) {
	user.Username = "changed"
	*user.GlobalName, *user.Avatar, *user.Banner, *user.Bio, *user.Pronouns = "changed", "changed", "changed", "changed", "changed"
	*user.AvatarColor, *user.BannerColor, *user.AccentColor = 9, 9, 9
}

func requireReferenceFixture(t *testing.T, value reflect.Value, path string) {
	t.Helper()
	switch value.Kind() {
	case reflect.Pointer:
		if value.IsNil() {
			t.Fatalf("populate %s in the ownership fixture and review its snapshot copy", path)
		}
		requireReferenceFixture(t, value.Elem(), path)
	case reflect.Struct:
		for i := range value.NumField() {
			requireReferenceFixture(t, value.Field(i), path+"."+value.Type().Field(i).Name)
		}
	case reflect.Slice:
		if value.Len() == 0 {
			t.Fatalf("populate %s in the ownership fixture and review its snapshot copy", path)
		}
		for i := range value.Len() {
			requireReferenceFixture(t, value.Index(i), path+"[]")
		}
	case reflect.Map:
		if value.Len() == 0 {
			t.Fatalf("populate %s in the ownership fixture and review its snapshot copy", path)
		}
		for _, key := range value.MapKeys() {
			requireReferenceFixture(t, value.MapIndex(key), path+"[]")
		}
	}
}

func TestChannelSnapshotsOwnMutableFields(t *testing.T) {
	mutations := map[string]func(*model.Channel){
		"scalar pointers": func(channel *model.Channel) {
			*channel.Topic, *channel.URL, *channel.Icon, *channel.RTCRegion = "changed", "changed", "changed", "changed"
			*channel.NSFWOverride = false
			*channel.ContentWarningText, *channel.LastPinTimestamp = "changed", "changed"
			*channel.Permissions = 0
		},
		"collections": func(channel *model.Channel) {
			channel.PermissionOverwrites[0].Allow = 0
			mutateSnapshotUser(&channel.Recipients[0])
			channel.Nicks["user"] = "changed"
			channel.AppliedTags[0], channel.MemberIDsPreview[0] = "changed", "changed"
		},
		"thread": func(channel *model.Channel) {
			channel.ThreadMetadata.Archived = true
			*channel.ThreadMetadata.Invitable = false
			channel.Member.Flags = 9
			member := channel.Member.Member
			mutateSnapshotUser(member.User)
			*member.Nick, *member.Avatar, *member.Banner, *member.CommunicationDisabledUntil = "changed", "changed", "changed", "changed"
			*member.AccentColor, *member.Permissions = 9, 0
			member.Roles[0] = "changed"
		},
		"forum": func(channel *model.Channel) {
			channel.AvailableTags[0].Name = "changed"
			*channel.AvailableTags[0].EmojiID, *channel.AvailableTags[0].EmojiName = "changed", "changed"
			*channel.DefaultReactionEmoji.EmojiID, *channel.DefaultReactionEmoji.EmojiName = "changed", "changed"
			*channel.DefaultSortOrder = 9
		},
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			c := New("bot")
			c.Apply(gateway.EventChannelCreate, json.RawMessage(snapshotChannel))
			var want model.Channel
			if err := json.Unmarshal([]byte(snapshotChannel), &want); err != nil {
				t.Fatal(err)
			}
			requireReferenceFixture(t, reflect.ValueOf(want), "channel")
			snapshot, ok := c.Channel("channel")
			if !ok || !reflect.DeepEqual(snapshot, want) {
				t.Fatal("snapshot did not preserve the cached channel")
			}
			mutate(&snapshot)
			if got, _ := c.Channel("channel"); !reflect.DeepEqual(got, want) {
				t.Fatal("editing returned channel changed cached state")
			}
		})
	}
}

func TestRoleSnapshotsOwnMutableFields(t *testing.T) {
	c := New("bot")
	c.Apply(gateway.EventGuildRoleCreate, json.RawMessage(snapshotRole))
	want, ok := c.Role("guild", "role")
	if !ok {
		t.Fatal("role was not cached")
	}
	snapshot, _ := c.Role("guild", "role")
	snapshot.Colors.PrimaryColor = 9
	*snapshot.Colors.SecondaryColor, *snapshot.Colors.TertiaryColor, *snapshot.HoistPosition = 9, 9, 9
	*snapshot.Icon, *snapshot.UnicodeEmoji, *snapshot.IconEmojiID = "changed", "changed", "changed"
	var expected struct {
		Role model.Role `json:"role"`
	}
	if err := json.Unmarshal([]byte(snapshotRole), &expected); err != nil {
		t.Fatal(err)
	}
	requireReferenceFixture(t, reflect.ValueOf(expected.Role), "role")
	if !reflect.DeepEqual(want, expected.Role) {
		t.Error("editing another snapshot changed a previously returned role")
	}
	if got, _ := c.Role("guild", "role"); !reflect.DeepEqual(got, expected.Role) {
		t.Error("editing returned role changed cached state")
	}
}

func TestCacheSnapshotsCanBeEditedConcurrently(t *testing.T) {
	c := New("bot")
	c.Apply(gateway.EventChannelCreate, json.RawMessage(snapshotChannel))
	c.Apply(gateway.EventGuildRoleCreate, json.RawMessage(snapshotRole))
	var edits sync.WaitGroup
	for range 8 {
		edits.Add(1)
		go func() {
			defer edits.Done()
			for range 100 {
				channel, _ := c.Channel("channel")
				channel.ThreadMetadata.Locked = true
				role, _ := c.Role("guild", "role")
				role.Colors.PrimaryColor = 9
			}
		}()
	}
	edits.Wait()
}

func TestCacheSnapshotsPreserveAbsentFields(t *testing.T) {
	c := New("bot")
	c.Remember("channel", "guild")
	if got, ok := c.Channel("channel"); !ok || !reflect.DeepEqual(got, model.Channel{ID: "channel", GuildID: "guild"}) {
		t.Fatal("channel snapshot changed absent fields")
	}
	c.Apply(gateway.EventGuildRoleCreate, json.RawMessage(`{"guild_id":"guild","role":{"id":"role"}}`))
	if got, ok := c.Role("guild", "role"); !ok || !reflect.DeepEqual(got, model.Role{ID: "role"}) {
		t.Fatal("role snapshot changed absent fields")
	}
}

var benchmarkChannel model.Channel
var benchmarkRole model.Role

func BenchmarkCacheSnapshot(b *testing.B) {
	for _, populated := range []bool{false, true} {
		name := "minimal"
		c := New("bot")
		if populated {
			name = "populated"
			c.Apply(gateway.EventChannelCreate, json.RawMessage(snapshotChannel))
			c.Apply(gateway.EventGuildRoleCreate, json.RawMessage(snapshotRole))
		} else {
			c.Remember("channel", "guild")
			c.Apply(gateway.EventGuildRoleCreate, json.RawMessage(`{"guild_id":"guild","role":{"id":"role"}}`))
		}
		b.Run("channel/"+name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchmarkChannel, _ = c.Channel("channel")
			}
		})
		b.Run("role/"+name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				benchmarkRole, _ = c.Role("guild", "role")
			}
		})
	}
}
