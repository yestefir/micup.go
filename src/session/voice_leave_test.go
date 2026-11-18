package session

import (
	"errors"
	"testing"

	"github.com/yestefir/micup.go/src/gateway"
	"github.com/yestefir/micup.go/src/model"
)

func TestFailedVoiceLeaveCanBeRetried(t *testing.T) {
	grant := model.VoiceServerUpdate{GuildID: "guild", ChannelID: "voice", ConnectionID: "connection", Endpoint: "wss://fixture.invalid", Token: "fixture"}
	s := &Session{grants: map[string]model.VoiceServerUpdate{"guild": grant}}
	for attempt := range 2 {
		if err := s.LeaveVoice("guild", "connection"); !errors.Is(err, gateway.ErrClosed) {
			t.Errorf("attempt %d returned %v, want closed error", attempt+1, err)
		}
		if got, ok := s.Voice("guild"); !ok || got != grant {
			t.Errorf("attempt %d discarded the grant before sending the leave", attempt+1)
		}
	}
}

func TestStaleVoiceLeavePreservesCurrentGrant(t *testing.T) {
	grant := model.VoiceServerUpdate{GuildID: "guild", ConnectionID: "current"}
	s := &Session{grants: map[string]model.VoiceServerUpdate{"guild": grant}}
	for _, connection := range []string{"stale", ""} {
		if err := s.LeaveVoice("guild", connection); err != nil {
			t.Fatalf("stale leave returned %v", err)
		}
		if got, ok := s.Voice("guild"); !ok || got != grant {
			t.Fatal("stale leave discarded the current grant")
		}
	}
}
