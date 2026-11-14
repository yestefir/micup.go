package session

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/yestefir/micup.go/src/gateway"
	"github.com/yestefir/micup.go/src/model"
)

func deliverVoiceGrant(t *testing.T, s *Session, grant model.VoiceServerUpdate) {
	t.Helper()
	data, err := json.Marshal(grant)
	if err != nil {
		t.Fatal(err)
	}
	s.handleVoiceServerUpdate(gateway.Event{Data: data})
}

func TestPendingVoiceJoinUsesLatestGrant(t *testing.T) {
	for _, consumed := range []bool{false, true} {
		for _, latest := range []string{"valid", "missing_token", "moved"} {
			name := "buffered/" + latest
			if consumed {
				name = "consumed/" + latest
			}
			t.Run(name, func(t *testing.T) {
				s, err := New(Config{Endpoint: "wss://localhost", Token: "fixture"}, Options{})
				if err != nil {
					t.Fatal(err)
				}
				pending, err := s.beginJoin("g")
				if err != nil {
					t.Fatal(err)
				}
				old := model.VoiceServerUpdate{GuildID: "g", ChannelID: "v", ConnectionID: "old", Endpoint: "wss://old.invalid", Token: "old-fixture"}
				deliverVoiceGrant(t, s, old)
				var result joinResult
				if consumed {
					result = <-pending
				}
				newer := model.VoiceServerUpdate{GuildID: "g", ChannelID: "v", ConnectionID: "new", Endpoint: "wss://new.invalid", Token: "new-fixture"}
				var wantErr error
				switch latest {
				case "missing_token":
					newer.Token, wantErr = "", ErrVoiceGrant
				case "moved":
					newer.ChannelID, wantErr = "other", ErrVoiceMoved
				}
				deliverVoiceGrant(t, s, newer)
				if !consumed {
					result = <-pending
				}
				grant, err := s.finishJoin("g", "v", pending, result)
				if !errors.Is(err, wantErr) {
					t.Errorf("latest=%s returned %v, want %v", latest, err, wantErr)
				}
				current, stored := s.Voice("g")
				if wantErr == nil {
					if grant != newer || !stored || current != newer {
						t.Error("latest grant was lost or overwritten by the earlier consumed result")
					}
				} else if stored {
					t.Error("older grant was stored instead of rejecting the latest update")
				}
				s.endJoin("g", pending)
			})
		}
	}
}

func TestCommittedVoiceJoinRoutesFurtherGrants(t *testing.T) {
	for _, rotate := range []bool{false, true} {
		name := "voice_loss"
		if rotate {
			name = "rotation_callback"
		}
		t.Run(name, func(t *testing.T) {
			callbacks := 0
			options := Options{OnVoiceLost: func(string, string) { callbacks++ }}
			if rotate {
				options.OnVoiceGrant = func(model.VoiceServerUpdate) { callbacks++ }
			}
			s, err := New(Config{Endpoint: "wss://localhost", Token: "fixture"}, options)
			if err != nil {
				t.Fatal(err)
			}
			pending, err := s.beginJoin("g")
			if err != nil {
				t.Fatal(err)
			}
			grant := model.VoiceServerUpdate{GuildID: "g", ChannelID: "v", ConnectionID: "old", Endpoint: "wss://voice.invalid", Token: "fixture"}
			if _, err := s.finishJoin("g", "v", pending, joinResult{grant: grant}); err != nil {
				t.Fatal(err)
			}
			// The old caller has committed but has not yet executed its defer.
			grant.ConnectionID = "new"
			deliverVoiceGrant(t, s, grant)
			if callbacks != 1 {
				t.Errorf("post-commit grant produced %d callbacks, want 1", callbacks)
			}
			current, stored := s.Voice("g")
			if rotate && (!stored || current != grant) {
				t.Error("post-commit rotation was not stored")
			}
			if !rotate && stored {
				t.Error("unhandled post-commit grant did not retire the earlier grant")
			}
			fresh, err := s.beginJoin("g")
			if err != nil {
				t.Fatalf("committed join kept its pending reservation: %v", err)
			}
			s.endJoin("g", pending)
			if s.pending["g"].results != fresh {
				t.Error("old deferred cleanup removed the fresh reservation")
			}
			s.endJoin("g", fresh)
		})
	}
}

func TestBufferedVoiceRefusalSurvivesLaterGrant(t *testing.T) {
	s, err := New(Config{Endpoint: "wss://localhost", Token: "fixture"}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	pending, err := s.beginJoin("g")
	if err != nil {
		t.Fatal(err)
	}
	s.handleGatewayError(gateway.GatewayError{Code: gateway.VoiceChannelFull})
	deliverVoiceGrant(t, s, model.VoiceServerUpdate{GuildID: "g", ChannelID: "v", Endpoint: "wss://voice.invalid", Token: "fixture"})
	if _, err := s.finishJoin("g", "v", pending, <-pending); !errors.Is(err, ErrVoiceRefused) {
		t.Fatalf("later grant hid a queued refusal: %v", err)
	}
	s.endJoin("g", pending)
}
