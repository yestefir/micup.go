package events

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/yestefir/micup.go/src/gateway"
	"github.com/yestefir/micup.go/src/model"
)

var uncaptured = map[string]bool{
	gateway.EventRateLimited:       true,
	gateway.EventGuildSync:         true,
	gateway.EventGuildDelete:       true,
	gateway.EventGuildMemberAdd:    true,
	gateway.EventGuildMemberRemove: true,
	gateway.EventGuildBanAdd:       true,
	gateway.EventGuildBanRemove:    true,
	gateway.EventPresenceUpdate:    true,

	gateway.EventGuildMemberListUpdate:     true,
	gateway.EventGuildCountsUpdate:         true,
	gateway.EventChannelRecipientAdd:       true,
	gateway.EventChannelRecipientRemove:    true,
	gateway.EventChannelMemberCountsUpdate: true,
	gateway.EventPresenceUpdateBulk:        true,
	gateway.EventPassiveUpdates:            true,
	gateway.EventReactionAddMany:           true,
	gateway.EventThreadMembersUpdate:       true,
	gateway.EventThreadMemberListUpdate:    true,
	gateway.EventCallCreate:                true,
	gateway.EventCallUpdate:                true,
	gateway.EventCallDelete:                true,
	gateway.EventEntranceSoundPlay:         true,
	gateway.EventRelationshipAdd:           true,
	gateway.EventRelationshipUpdate:        true,
	gateway.EventRelationshipRemove:        true,
}

func fixtures(t *testing.T) map[string][]string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join("testdata", "*.json"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no fixtures: %v", err)
	}
	byName := map[string][]string{}
	for _, file := range files {
		name, _, _ := strings.Cut(filepath.Base(file), ".")
		byName[name] = append(byName[name], file)
	}
	return byName
}

func load(t *testing.T, file string) gateway.Event {
	t.Helper()
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	name, _, _ := strings.Cut(filepath.Base(file), ".")
	return gateway.Event{Name: name, Data: data}
}

func TestRegistryMatchesKnownEvents(t *testing.T) {
	known := map[string]bool{}
	for _, name := range gateway.KnownEvents() {
		known[name] = true
		if _, ok := registry[name]; !ok {
			t.Errorf("%s has no payload type", name)
		}
	}
	for name := range registry {
		if !known[name] {
			t.Errorf("%s is registered but not in gateway.KnownEvents", name)
		}
	}
}

func TestEveryEventHasCapturedPayloadOrIsListedUncaptured(t *testing.T) {
	captured := fixtures(t)
	for _, name := range gateway.KnownEvents() {
		has := len(captured[name]) > 0
		if !has && !uncaptured[name] {
			t.Errorf("%s has no fixture in testdata and is not listed as uncaptured", name)
		}
		if has && uncaptured[name] {
			t.Errorf("%s now has a fixture; remove it from uncaptured", name)
		}
	}
}

func TestCapturedPayloadsDecode(t *testing.T) {
	for name, files := range fixtures(t) {
		if _, ok := registry[name]; !ok {
			t.Errorf("fixture for unregistered event %s", name)
			continue
		}
		for _, file := range files {
			value, err := Decode(load(t, file))
			if err != nil {
				t.Errorf("%s: %v", file, err)
				continue
			}
			if v := reflect.ValueOf(value); v.IsZero() && !(v.Kind() == reflect.Struct && v.NumField() == 0) {
				t.Errorf("%s decoded to a zero value", file)
			}
		}
	}
}

func decodeFirst[T any](t *testing.T, name string) T {
	t.Helper()
	value, err := As[T](load(t, filepath.Join("testdata", name+".0.json")))
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func TestCapturedThreadAndEmojiPayloads(t *testing.T) {
	created := decodeFirst[ThreadCreate](t, gateway.EventThreadCreate)
	if !created.NewlyCreated || !created.IsThread() || created.ParentID == "" || created.ThreadMetadata == nil || created.Member == nil || created.Member.UserID == "" {
		t.Errorf("thread create: %+v", created)
	}
	updated := decodeFirst[model.Channel](t, gateway.EventThreadUpdate)
	if updated.ID == "" || updated.ThreadMetadata == nil {
		t.Errorf("thread update: %+v", updated)
	}
	deleted := decodeFirst[ThreadDelete](t, gateway.EventThreadDelete)
	if deleted.ID == "" || deleted.GuildID == "" || deleted.ParentID == "" || deleted.Type != model.ChannelTypePublicThread {
		t.Errorf("thread delete: %+v", deleted)
	}
	member := decodeFirst[model.ThreadMember](t, gateway.EventThreadMemberUpdate)
	if member.ID == "" || member.UserID == "" || member.GuildID == "" || member.Flags != model.ThreadMemberHasInteracted {
		t.Errorf("thread member update: %+v", member)
	}
	sync := decodeFirst[ThreadListSync](t, gateway.EventThreadListSync)
	if sync.GuildID == "" || len(sync.Threads) != 2 || len(sync.Members) != 1 || sync.Threads[0].Flags != model.ChannelFlagPinned || !sync.Threads[0].ThreadMetadata.Locked {
		t.Errorf("thread list sync: %+v", sync)
	}
	stickers, err := As[GuildStickersUpdate](load(t, filepath.Join("testdata", gateway.EventGuildStickersUpdate+".1.json")))
	if err != nil || len(stickers.Stickers) != 1 || len(stickers.Stickers[0].Tags) != 2 || stickers.Stickers[0].Description != "renamed" {
		t.Errorf("stickers update: %+v %v", stickers, err)
	}
	emojis, err := As[GuildEmojisUpdate](load(t, filepath.Join("testdata", gateway.EventGuildEmojisUpdate+".1.json")))
	if err != nil || emojis.GuildID == "" || len(emojis.Emojis) != 1 || emojis.Emojis[0].ID == "" {
		t.Errorf("emojis update: %+v %v", emojis, err)
	}
}

func TestCapturedPayloadFields(t *testing.T) {
	ready := decodeFirst[Ready](t, gateway.EventReady)
	if ready.User.ID == "" || ready.SessionID == "" || len(ready.Guilds) == 0 || ready.Application.ID == "" || len(ready.Shard) != 2 {
		t.Errorf("ready: %+v", ready)
	}

	if ready.GatewayRollout == nil || ready.GatewayRollout.GameActivityEpoch == 0 {
		t.Errorf("ready rollout: %+v", ready.GatewayRollout)
	}
	if rollout := decodeFirst[GatewayRollout](t, gateway.EventGatewayRolloutUpdate); rollout.GameActivityEpoch == 0 {
		t.Errorf("rollout update: %+v", rollout)
	}

	guild := decodeFirst[GuildCreate](t, gateway.EventGuildCreate)
	if guild.ID == "" || guild.OwnerID == "" || len(guild.Roles) == 0 || len(guild.Channels) == 0 || len(guild.Members) == 0 || guild.JoinedAt == "" {
		t.Errorf("guild create: %+v", guild)
	}
	update := decodeFirst[GuildUpdate](t, gateway.EventGuildUpdate)
	if update.ID == "" || update.OwnerID == "" || len(update.Roles) == 0 {
		t.Errorf("guild update: %+v", update)
	}

	chunk := decodeFirst[MembersChunk](t, gateway.EventGuildMembersChunk)
	if chunk.GuildID == "" || len(chunk.Members) == 0 || len(chunk.Presences) == 0 || chunk.ChunkCount != 1 {
		t.Errorf("members chunk: %+v", chunk)
	}
	for _, presence := range chunk.Presences {
		if presence.User.ID == "" || presence.Status == "" {
			t.Errorf("chunk presence: %+v", presence)
		}
	}

	audit := decodeFirst[AuditLogEntryCreate](t, gateway.EventGuildAuditLogEntryCreate)
	if audit.ID == "" || audit.ActionType == 0 || audit.Options == nil || audit.Options.ChannelID == "" {
		t.Errorf("audit log entry: %+v", audit)
	}

	role := decodeFirst[RoleEvent](t, gateway.EventGuildRoleCreate)
	if role.GuildID == "" || role.Role.ID == "" || role.Role.Colors == nil {
		t.Errorf("role: %+v", role)
	}
	roles := decodeFirst[RoleUpdateBulk](t, gateway.EventGuildRoleUpdateBulk)
	if roles.GuildID == "" || len(roles.Roles) == 0 {
		t.Errorf("roles: %+v", roles)
	}

	member := decodeFirst[GuildMember](t, gateway.EventGuildMemberUpdate)
	if member.GuildID == "" || member.User == nil || member.User.ID == "" {
		t.Errorf("member: %+v", member)
	}

	channels := decodeFirst[ChannelUpdateBulk](t, gateway.EventChannelUpdateBulk)
	if channels.GuildID == "" || len(channels.Channels) == 0 {
		t.Errorf("channels: %+v", channels)
	}
	pins := decodeFirst[ChannelPinsUpdate](t, gateway.EventChannelPinsUpdate)
	if pins.ChannelID == "" || pins.LastPinTimestamp == nil {
		t.Errorf("pins: %+v", pins)
	}

	message := decodeFirst[MessageCreate](t, gateway.EventMessageCreate)
	if message.ID == "" || message.Author.ID == "" || message.Member == nil || message.Content == "" {
		t.Errorf("message: %+v", message)
	}
	deleted := decodeFirst[MessageDelete](t, gateway.EventMessageDelete)
	if deleted.ID == "" || deleted.AuthorID == "" || deleted.Content == nil {
		t.Errorf("message delete: %+v", deleted)
	}
	bulk := decodeFirst[MessageDeleteBulk](t, gateway.EventMessageDeleteBulk)
	if len(bulk.IDs) == 0 {
		t.Errorf("bulk delete: %+v", bulk)
	}

	reaction := decodeFirst[ReactionEvent](t, gateway.EventReactionAdd)
	if reaction.UserID == "" || reaction.Emoji.Name == "" || reaction.Member == nil {
		t.Errorf("reaction: %+v", reaction)
	}
	emoji := decodeFirst[ReactionRemoveEmoji](t, gateway.EventReactionRemoveEmoji)
	if emoji.MessageID == "" || emoji.Emoji.Name == "" {
		t.Errorf("reaction remove emoji: %+v", emoji)
	}

	typing := decodeFirst[TypingStart](t, gateway.EventTypingStart)
	if typing.UserID == "" || typing.Timestamp == 0 {
		t.Errorf("typing: %+v", typing)
	}

	invite := decodeFirst[model.Invite](t, gateway.EventInviteCreate)
	if invite.Code == "" || invite.Inviter == nil || invite.MaxAge == 0 {
		t.Errorf("invite: %+v", invite)
	}

	sessions := decodeFirst[SessionsReplace](t, gateway.EventSessionsReplace)
	if len(sessions) == 0 || sessions[0].Status == "" {
		t.Errorf("sessions: %+v", sessions)
	}

	self := decodeFirst[model.User](t, gateway.EventUserUpdate)
	if self.ID == "" || !self.Bot || self.Discriminator == "" {
		t.Errorf("user update: %+v", self)
	}

	commands := decodeFirst[ApplicationCommandsUpdate](t, gateway.EventApplicationCommandsUpdate)
	if commands.ApplicationID == "" || commands.Version == "" {
		t.Errorf("commands update: %+v", commands)
	}

	click := decodeFirst[model.Interaction](t, gateway.EventInteractionCreate)
	if !click.IsButton() || click.CustomID() == "" || click.Message == nil || click.Channel == nil || click.Guild == nil || click.Invoker() == nil || click.Member.Permissions == nil {
		t.Errorf("button interaction: %+v", click)
	}
	command, err := As[model.Interaction](load(t, filepath.Join("testdata", gateway.EventInteractionCreate+".2.json")))
	if err != nil {
		t.Fatal(err)
	}
	data, ok := command.CommandData()
	if !ok || !command.IsCommand() || command.Context == nil || command.AttachmentSizeLimit == 0 {
		t.Fatalf("command interaction: %+v", command)
	}
	if role, ok := data.Role("role"); !ok || role.Name == "" || strings.Join(data.Path(), " ") != "admin role add" {
		t.Errorf("command data: %+v", data)
	}
	command, _ = As[model.Interaction](load(t, filepath.Join("testdata", gateway.EventInteractionCreate+".3.json")))
	data, _ = command.CommandData()
	if channel, ok := data.Channel("where"); !ok || channel.Permissions == nil {
		t.Errorf("channel option: %+v", data)
	}
	command, _ = As[model.Interaction](load(t, filepath.Join("testdata", gateway.EventInteractionCreate+".4.json")))
	data, _ = command.CommandData()
	if text, _ := data.String("text"); text == "" {
		t.Errorf("string option: %+v", data)
	}
	if times, ok := data.Int("times"); !ok || times != 1 {
		t.Errorf("integer option: %+v", data)
	}

	voice := decodeFirst[model.VoiceState](t, gateway.EventVoiceStateUpdate)
	if voice.GuildID == "" || voice.UserID == "" || voice.Member == nil {
		t.Errorf("voice state: %+v", voice)
	}
	server := decodeFirst[model.VoiceServerUpdate](t, gateway.EventVoiceServerUpdate)
	if server.GuildID == "" || server.Endpoint == "" || server.Token == "" {
		t.Errorf("voice server: %+v", server)
	}
}

func TestGuildPropertiesFillOwner(t *testing.T) {
	var guild GuildCreate
	if err := json.Unmarshal([]byte(`{"id":"g","properties":{"owner_id":"o","name":"n"}}`), &guild); err != nil {
		t.Fatal(err)
	}
	if guild.OwnerID != "o" || guild.Name != "n" {
		t.Fatalf("properties were not applied: %+v", guild.Guild)
	}
	var update GuildUpdate
	if err := json.Unmarshal([]byte(`{"guild_id":"g","owner_id":"o"}`), &update); err != nil {
		t.Fatal(err)
	}
	if update.ID != "g" || update.OwnerID != "o" {
		t.Fatalf("guild update: %+v", update.Guild)
	}
}

func TestDecodeKeepsUnknownEventsRaw(t *testing.T) {
	value, err := Decode(gateway.Event{Name: "SOMETHING_NEW", Data: json.RawMessage(`{"a":1}`)})
	raw, ok := value.(json.RawMessage)
	if err != nil || !ok || string(raw) != `{"a":1}` {
		t.Fatalf("unknown event: %v %v", value, err)
	}
}

func TestAsRejectsWrongType(t *testing.T) {
	if _, err := As[MessageDelete](gateway.Event{Name: gateway.EventMessageCreate, Data: json.RawMessage(`{}`)}); err == nil {
		t.Fatal("wrong payload type accepted")
	}
	if _, err := As[MessageCreate](gateway.Event{Name: gateway.EventMessageCreate, Data: json.RawMessage(`{"id":`)}); err == nil {
		t.Fatal("malformed payload accepted")
	}
}

func TestParseKeepsFieldsAroundTypeMismatch(t *testing.T) {
	guild, ok := Parse[GuildCreate](json.RawMessage(`{"id":"g","member_count":"many","properties":{"owner_id":"o"}}`))
	if !ok || guild.ID != "g" || guild.OwnerID != "o" {
		t.Fatalf("type mismatch dropped the payload: %v %+v", ok, guild.Guild)
	}
	if _, ok := Parse[GuildCreate](json.RawMessage(`{"id":`)); ok {
		t.Fatal("malformed JSON accepted")
	}
}
