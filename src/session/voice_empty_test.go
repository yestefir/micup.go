package session

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/yestefir/micup.go/src/gateway"
)

func TestVoiceEmptyFiresWhenLastHumanLeavesBotChannel(t *testing.T) {
	type call struct{ guild, channel string }
	var calls []call
	s, err := New(Config{Endpoint: "wss://localhost", Token: "fixture", BotID: "bot"}, Options{OnVoiceEmpty: func(guildID, channelID string) { calls = append(calls, call{guildID, channelID}) }})
	if err != nil {
		t.Fatal(err)
	}
	update := func(user, channel string) {
		t.Helper()
		data, _ := json.Marshal(map[string]string{"guild_id": "g", "user_id": user, "channel_id": channel})
		if err := s.dispatch(context.Background(), gateway.Event{Name: gateway.EventVoiceStateUpdate, Data: data}); err != nil {
			t.Fatal(err)
		}
	}
	update("bot", "v")
	update("a", "v")
	update("b", "v")
	update("a", "")
	if len(calls) != 0 {
		t.Fatalf("callback fired while a listener remained: %v", calls)
	}
	update("c", "other")
	update("c", "")
	if len(calls) != 0 {
		t.Fatalf("callback fired for a different channel: %v", calls)
	}
	update("b", "")
	if len(calls) != 1 || calls[0] != (call{"g", "v"}) {
		t.Fatalf("callback calls = %v, want one for g/v", calls)
	}
	update("b", "v")
	update("b", "other")
	if len(calls) != 2 {
		t.Fatalf("callback did not fire after the listener moved away: %v", calls)
	}
}
