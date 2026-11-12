package session_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yestefir/micup.go/src/gateway"
	"github.com/yestefir/micup.go/src/internal/fixture"
	"github.com/yestefir/micup.go/src/session"
)

func TestSessionCountRequestsAndSubscriptions(t *testing.T) {
	p := fixture.New(t)
	var readies, resumes atomic.Int64
	s := newSession(t, p, session.Options{OnReady: func(resumed bool) {
		if resumed {
			resumes.Add(1)
		} else {
			readies.Add(1)
		}
	}})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := s.RequestGuildCounts(ctx, []string{fixture.GuildID}); !errors.Is(err, session.ErrNotReady) {
		t.Fatalf("count before ready: %v", err)
	}
	run(t, s)
	fixture.Eventually(t, s.Ready)
	if readies.Load() != 1 || s.SessionID() != "session-1" {
		t.Fatalf("ready hook ran %d times, session %q", readies.Load(), s.SessionID())
	}
	counts, err := s.RequestGuildCounts(ctx, []string{fixture.GuildID, fixture.GuildID, ""})
	if err != nil || len(counts) != 1 || counts[0].MemberCount != 10 || counts[0].OnlineCount != 3 {
		t.Fatalf("guild counts: %+v %v", counts, err)
	}
	channels, err := s.RequestChannelMemberCounts(ctx, fixture.GuildID, []string{fixture.VoiceID, fixture.TextID})
	if err != nil || len(channels) != 2 || channels[0].MemberCount != 4 {
		t.Fatalf("channel counts: %+v %v", channels, err)
	}
	if _, err := s.RequestGuildCounts(ctx, nil); err == nil {
		t.Fatal("an empty count request was accepted")
	}
	active := true
	if err := s.Subscribe(ctx, map[string]gateway.Subscription{fixture.GuildID: {Active: &active, MemberListChannels: map[string][][2]int{fixture.TextID: {{0, 99}}}}}); err != nil {
		t.Fatal(err)
	}
	fixture.Eventually(t, func() bool { return len(p.Subscriptions()) == 1 })
	if sent := string(p.Subscriptions()[0]); !strings.Contains(sent, `"subscriptions":{"4194304":{"active":true,"member_list_channels":{"text":[[0,99]]}}}`) {
		t.Fatalf("subscription sent as %s", sent)
	}
	if err := s.Subscribe(ctx, map[string]gateway.Subscription{fixture.GuildID: {MemberListChannels: map[string][][2]int{fixture.TextID: {{0, 200}}}}}); err == nil {
		t.Fatal("an oversized member list range was accepted")
	}
}

func TestUpdateVoiceStateKeepsTheConnection(t *testing.T) {
	p := fixture.New(t)
	s := newSession(t, p, session.Options{})
	run(t, s)
	fixture.Eventually(t, s.Ready)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.UpdateVoiceState(ctx, fixture.GuildID, session.VoiceJoin{SelfDeaf: true}); !errors.Is(err, session.ErrNotInVoice) {
		t.Fatalf("update before join: %v", err)
	}
	if _, err := s.JoinVoice(ctx, fixture.GuildID, fixture.VoiceID, session.VoiceJoin{}); err != nil {
		t.Fatal(err)
	}
	if err := s.UpdateVoiceState(ctx, fixture.GuildID, session.VoiceJoin{SelfDeaf: true}); err != nil {
		t.Fatal(err)
	}
	fixture.Eventually(t, func() bool { return len(p.VoiceUpdates()) == 2 })
	var update map[string]any
	_ = json.Unmarshal(p.VoiceUpdates()[1], &update)
	if update["connection_id"] != fixture.Connection || update["channel_id"] != fixture.VoiceID || update["self_deaf"] != true {
		t.Fatalf("voice state update sent as %v", update)
	}
}

func TestReadyHookCanUseGatewayRequests(t *testing.T) {
	p := fixture.New(t)
	result := make(chan error, 1)
	var s *session.Session
	s = newSession(t, p, session.Options{OnReady: func(bool) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := s.RequestGuildCounts(ctx, []string{fixture.GuildID})
		result <- err
	}})
	run(t, s)
	select {
	case err := <-result:
		if err != nil {
			t.Fatalf("count request from OnReady: %v", err)
		}
	case <-time.After(8 * time.Second):
		t.Fatal("OnReady blocked the Gateway")
	}
}
