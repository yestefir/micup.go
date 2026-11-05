package state

import (
	"encoding/json"
	"testing"

	"github.com/yestefir/micup.go/src/gateway"
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
