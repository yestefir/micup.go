package session_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yestefir/micup.go/src/gateway"
	"github.com/yestefir/micup.go/src/internal/fixture"
	"github.com/yestefir/micup.go/src/model"
	"github.com/yestefir/micup.go/src/rest"
	"github.com/yestefir/micup.go/src/session"
)

func newSession(t *testing.T, p *fixture.Platform, options session.Options) *session.Session {
	t.Helper()
	client, err := rest.New(rest.Config{BaseURL: p.APIURL(), Token: fixture.Token})
	if err != nil {
		t.Fatal(err)
	}
	s, err := session.New(session.Config{Endpoint: p.GatewayURL(), Token: fixture.Token, BotID: fixture.BotID, Channels: client}, options)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func run(t *testing.T, s *session.Session) chan error {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- s.Run(ctx) }()
	t.Cleanup(func() { cancel(); <-done })
	return done
}

type lostVoice struct {
	mu     sync.Mutex
	events []string
}

func (l *lostVoice) add(guild, connection string) {
	l.mu.Lock()
	l.events = append(l.events, guild+"/"+connection)
	l.mu.Unlock()
}

func (l *lostVoice) count() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.events)
}

func (l *lostVoice) last() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.events[len(l.events)-1]
}

func TestSessionTracksStateResumesAndJoinsVoice(t *testing.T) {
	p := fixture.New(t)
	var events atomic.Int64
	lost := &lostVoice{}
	s := newSession(t, p, session.Options{MinBackoff: 10 * time.Millisecond, OnVoiceLost: lost.add, Handler: func(context.Context, gateway.Event) error { events.Add(1); return nil }})
	run(t, s)
	fixture.Eventually(t, s.Ready)
	if s.State().VoiceChannel(fixture.GuildID, "listener") != fixture.VoiceID || s.State().Guilds() != 1 {
		t.Fatal("READY guilds were not applied to state")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	grant, err := s.JoinVoice(ctx, fixture.GuildID, fixture.VoiceID, session.VoiceJoin{SelfDeaf: true})
	if err != nil {
		t.Fatal(err)
	}
	if grant.ConnectionID != fixture.Connection || grant.ChannelID != fixture.VoiceID || grant.Token != "grant-token" {
		t.Fatalf("unexpected voice grant: %+v", grant)
	}
	fixture.Eventually(t, func() bool { return s.State().BotVoiceChannel(fixture.GuildID) == fixture.VoiceID })
	if current, ok := s.Voice(fixture.GuildID); !ok || current.ConnectionID != fixture.Connection {
		t.Fatal("voice grant was not stored")
	}
	p.Send(map[string]any{"op": gateway.OpReconnect})
	fixture.Eventually(t, func() bool { return p.Resumes.Load() == 1 && s.Ready() })
	if p.Identifies.Load() != 1 || p.LastResumeSession() != "session-1" {
		t.Fatal("reconnect did not resume the existing session")
	}
	p.Event(gateway.EventVoiceStateUpdate, map[string]string{"guild_id": fixture.GuildID, "user_id": fixture.BotID, "channel_id": ""})
	fixture.Eventually(t, func() bool { return lost.count() == 1 })
	if lost.last() != fixture.GuildID+"/"+fixture.Connection {
		t.Fatalf("voice loss reported the wrong connection: %s", lost.last())
	}
	if _, ok := s.Voice(fixture.GuildID); ok {
		t.Fatal("lost voice grant was kept")
	}
	if _, err = s.JoinVoice(ctx, fixture.GuildID, fixture.VoiceID, session.VoiceJoin{}); err != nil {
		t.Fatal(err)
	}
	if err = s.LeaveVoice(fixture.GuildID, "other"); err != nil || p.Leaves.Load() != 0 {
		t.Fatal("leave with a stale connection ID was sent")
	}
	if err = s.LeaveVoice(fixture.GuildID, fixture.Connection); err != nil {
		t.Fatal(err)
	}
	fixture.Eventually(t, func() bool { return p.Leaves.Load() == 1 })
	if _, err = s.JoinVoice(ctx, fixture.GuildID, fixture.VoiceID, session.VoiceJoin{}); err != nil {
		t.Fatal(err)
	}
	p.Event(gateway.EventVoiceServerUpdate, map[string]string{"guild_id": fixture.GuildID, "connection_id": "conn-2", "endpoint": "wss://voice.invalid", "token": "other"})
	fixture.Eventually(t, func() bool { return lost.count() == 2 })
	if events.Load() == 0 {
		t.Fatal("application handler did not receive events")
	}
}

func TestSessionResolvesGuildOnceThroughREST(t *testing.T) {
	p := fixture.New(t)
	s := newSession(t, p, session.Options{})
	for range 2 {
		guild, err := s.ResolveGuild(context.Background(), fixture.TextID)
		if err != nil || guild != fixture.GuildID {
			t.Fatalf("guild was not resolved: %q %v", guild, err)
		}
	}
	if p.ChannelLookups.Load() != 1 {
		t.Fatal("resolved channel was not cached")
	}
}

func TestSessionStopsOnFatalClose(t *testing.T) {
	p := fixture.New(t)
	p.CloseCode = 4004
	s := newSession(t, p, session.Options{MinBackoff: 10 * time.Millisecond})
	done := run(t, s)
	select {
	case err := <-done:
		var closed *gateway.CloseError
		if !errors.As(err, &closed) || closed.Code != 4004 || !closed.Fatal() {
			t.Fatalf("fatal close was not returned: %v", err)
		}
		done <- err
	case <-time.After(5 * time.Second):
		t.Fatal("session kept reconnecting after a fatal close")
	}
	if p.Connects.Load() != 1 {
		t.Fatal("session reconnected after a fatal close")
	}
}

func TestSessionIdentifiesAgainAfterInvalidSession(t *testing.T) {
	p := fixture.New(t)
	s := newSession(t, p, session.Options{MinBackoff: 10 * time.Millisecond})
	run(t, s)
	fixture.Eventually(t, s.Ready)
	p.Send(map[string]any{"op": gateway.OpInvalidSession, "d": false})
	fixture.Eventually(t, func() bool { return p.Identifies.Load() == 2 && s.Ready() })
	if p.Resumes.Load() != 0 {
		t.Fatal("invalidated session was resumed")
	}
	if p.Connects.Load() != 1 {
		t.Fatalf("invalid session opened a new socket instead of identifying on the same one: %d", p.Connects.Load())
	}
}

func TestJoinVoiceRequiresReadySession(t *testing.T) {
	s := newSession(t, fixture.New(t), session.Options{})
	if _, err := s.JoinVoice(context.Background(), fixture.GuildID, fixture.VoiceID, session.VoiceJoin{}); !errors.Is(err, session.ErrNotReady) {
		t.Fatalf("join before READY was accepted: %v", err)
	}
}

func TestNewRejectsInvalidConfig(t *testing.T) {
	for name, config := range map[string]session.Config{
		"no endpoint": {Token: "t"},
		"bad shard":   {Endpoint: "wss://micup.gg", Token: "t", ShardID: 1, ShardCount: 1},
	} {
		if _, err := session.New(config, session.Options{}); err == nil {
			t.Fatalf("%s accepted", name)
		}
	}
}

func TestJoinVoiceFailsFastOnGatewayRefusal(t *testing.T) {
	p := fixture.New(t)
	p.RefuseVoice = gateway.VoiceChannelFull
	s := newSession(t, p, session.Options{})
	run(t, s)
	fixture.Eventually(t, s.Ready)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	started := time.Now()
	_, err := s.JoinVoice(ctx, fixture.GuildID, fixture.VoiceID, session.VoiceJoin{})
	var refused *session.VoiceRefusedError
	if !errors.As(err, &refused) || refused.Code != gateway.VoiceChannelFull || !errors.Is(err, session.ErrVoiceRefused) {
		t.Fatalf("refusal was not returned: %v", err)
	}
	if time.Since(started) > time.Second {
		t.Fatal("refused join waited for the context instead of failing fast")
	}
}

func TestUnrequestedGrantRotatesWhenHandled(t *testing.T) {
	p := fixture.New(t)
	lost := &lostVoice{}
	rotated := make(chan string, 1)
	s := newSession(t, p, session.Options{OnVoiceLost: lost.add, OnVoiceGrant: func(grant model.VoiceServerUpdate) { rotated <- grant.ConnectionID }})
	run(t, s)
	fixture.Eventually(t, s.Ready)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, err := s.JoinVoice(ctx, fixture.GuildID, fixture.VoiceID, session.VoiceJoin{}); err != nil {
		t.Fatal(err)
	}
	p.Event(gateway.EventVoiceServerUpdate, map[string]string{"guild_id": fixture.GuildID, "channel_id": fixture.VoiceID, "connection_id": fixture.Connection, "endpoint": "wss://voice-2.invalid", "token": "new-token"})
	select {
	case id := <-rotated:
		if id != fixture.Connection {
			t.Fatalf("rotated grant has the wrong connection: %s", id)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("region change grant was not delivered")
	}
	if grant, ok := s.Voice(fixture.GuildID); !ok || grant.Endpoint != "wss://voice-2.invalid" {
		t.Fatalf("rotated grant was not stored: %+v", grant)
	}
	if lost.count() != 0 {
		t.Fatal("region change was reported as voice loss")
	}
}

func TestSessionSendsIdentifyOptionsAndPresence(t *testing.T) {
	p := fixture.New(t)
	s := newSession(t, p, session.Options{MinBackoff: 10 * time.Millisecond, IgnoredEvents: []string{"TYPING_START"}, Presence: &model.Presence{Status: model.StatusIdle}})
	run(t, s)
	fixture.Eventually(t, s.Ready)
	identify := p.LastIdentify()
	if string(identify["ignored_events"]) != `["TYPING_START"]` || !strings.Contains(string(identify["presence"]), `"idle"`) {
		t.Fatalf("identify options were not sent: %s %s", identify["ignored_events"], identify["presence"])
	}
	if err := s.UpdatePresence(context.Background(), model.Presence{Status: model.StatusDND, CustomStatus: &model.CustomStatus{Text: "playing"}}); err != nil {
		t.Fatal(err)
	}
	fixture.Eventually(t, func() bool { return len(p.Presences()) == 1 })
	if !strings.Contains(string(p.Presences()[0]), `"dnd"`) {
		t.Fatalf("presence update was not sent: %s", p.Presences()[0])
	}
	p.Send(map[string]any{"op": gateway.OpInvalidSession, "d": false})
	fixture.Eventually(t, func() bool { return p.Identifies.Load() == 2 })
	if !strings.Contains(string(p.LastIdentify()["presence"]), `"dnd"`) {
		t.Fatal("re-identify did not keep the updated presence")
	}
	if err := s.UpdatePresence(context.Background(), model.Presence{Status: "busy"}); err == nil {
		t.Fatal("invalid status accepted")
	}
	if _, err := session.New(session.Config{Endpoint: "wss://micup.gg", Token: "t"}, session.Options{IgnoredEvents: []string{"typing start"}}); err == nil {
		t.Fatal("invalid ignored event accepted")
	}
}

func TestJoinVoiceFailsFastWithoutConnectPermission(t *testing.T) {
	p := fixture.New(t)
	s := newSession(t, p, session.Options{})
	run(t, s)
	fixture.Eventually(t, s.Ready)
	if _, err := s.JoinVoice(context.Background(), fixture.GuildID, fixture.LockedID, session.VoiceJoin{}); !errors.Is(err, session.ErrVoicePermission) {
		t.Fatalf("join into a locked channel was sent: %v", err)
	}
	if perms, ok := s.State().BotPermissions(fixture.GuildID, fixture.VoiceID); !ok || !perms.Has(model.PermissionConnect) {
		t.Fatal("bot permissions were not computed from READY")
	}
}

func TestRequestGuildMembersCollectsChunks(t *testing.T) {
	p := fixture.New(t)
	s := newSession(t, p, session.Options{})
	run(t, s)
	fixture.Eventually(t, s.Ready)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	members, err := s.RequestGuildMembers(ctx, model.MembersRequest{GuildID: fixture.GuildID})
	if err != nil || len(members) != 2 {
		t.Fatalf("chunks were not collected: %v %d", err, len(members))
	}
	if _, ok := s.State().MemberRoles(fixture.GuildID, "dj"); !ok {
		t.Fatal("chunk members were not cached")
	}
	p.RateLimitMembers = true
	var limited *session.MembersRateLimitedError
	if _, err = s.RequestGuildMembers(ctx, model.MembersRequest{GuildID: fixture.GuildID}); !errors.As(err, &limited) || limited.RetryAfter != 12500*time.Millisecond || !errors.Is(err, session.ErrMembersRateLimited) {
		t.Fatalf("rate limit was not returned: %v", err)
	}
}
