package shard_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/yestefir/micup.go/src/gateway"
	"github.com/yestefir/micup.go/src/internal/fixture"
	"github.com/yestefir/micup.go/src/model"
	"github.com/yestefir/micup.go/src/rest"
	"github.com/yestefir/micup.go/src/session"
	"github.com/yestefir/micup.go/src/shard"
)

func base(t *testing.T, p *fixture.Platform) session.Config {
	t.Helper()
	client, err := rest.New(rest.Config{BaseURL: p.APIURL(), Token: fixture.Token})
	if err != nil {
		t.Fatal(err)
	}
	return session.Config{Endpoint: p.GatewayURL(), Token: fixture.Token, BotID: fixture.BotID, Channels: client}
}

func newManager(t *testing.T, p *fixture.Platform, options shard.Options) *shard.Manager {
	t.Helper()
	m, err := shard.NewManager(base(t, p), options)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func run(t *testing.T, m *shard.Manager) chan error {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- m.Run(ctx) }()
	t.Cleanup(func() { cancel(); <-done })
	return done
}

func TestForGuild(t *testing.T) {
	id, err := shard.ForGuild("4194304", 2)
	if err != nil || id != 1 {
		t.Fatalf("unexpected shard: %d %v", id, err)
	}
	if _, err = shard.ForGuild("guild", 2); err == nil {
		t.Fatal("non-numeric guild accepted")
	}
	if _, err = shard.ForGuild("1", 0); err == nil {
		t.Fatal("zero shard count accepted")
	}
}

func TestNewManagerRejectsInvalidOptions(t *testing.T) {
	config := base(t, fixture.New(t))
	for name, options := range map[string]shard.Options{
		"duplicate":    {ShardCount: 2, ShardIDs: []int{1, 1}},
		"out of range": {ShardCount: 2, ShardIDs: []int{2}},
		"negative":     {ShardCount: 2, ShardIDs: []int{-1}},
		"zero count":   {ShardCount: -1},
	} {
		if _, err := shard.NewManager(config, options); err == nil {
			t.Fatalf("%s options accepted", name)
		}
	}
}

func TestManagerIdentifiesEveryShardInTurn(t *testing.T) {
	p := fixture.New(t)
	var mu sync.Mutex
	seen := map[int]bool{}
	interval := 80 * time.Millisecond
	m := newManager(t, p, shard.Options{ShardCount: 3, IdentifyInterval: interval, Session: session.Options{Handler: func(_ context.Context, event gateway.Event) error {
		if event.Name == gateway.EventReady {
			mu.Lock()
			seen[event.ShardID] = true
			mu.Unlock()
		}
		return nil
	}}})
	run(t, m)
	fixture.Eventually(t, m.Ready)
	log := p.IdentifyLog()
	if len(log) != 3 {
		t.Fatalf("expected 3 identifies, got %d", len(log))
	}
	shards := map[int]bool{}
	for i, record := range log {
		shards[record.Shard] = true
		if i > 0 && record.At.Sub(log[i-1].At) < interval-10*time.Millisecond {
			t.Fatalf("identify %d came %v after the previous one", i, record.At.Sub(log[i-1].At))
		}
	}
	mu.Lock()
	defer mu.Unlock()
	if len(shards) != 3 || len(seen) != 3 {
		t.Fatalf("not every shard identified and reported READY: %v %v", shards, seen)
	}
}

func TestManagerRoutesGuildsToOwningShard(t *testing.T) {
	p := fixture.New(t)
	m := newManager(t, p, shard.Options{ShardCount: 2, IdentifyInterval: time.Millisecond})
	run(t, m)
	fixture.Eventually(t, m.Ready)
	s, err := m.ForGuild(fixture.GuildID)
	if err != nil || s.ShardID() != 1 {
		t.Fatalf("guild was not routed to shard 1: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	grant, err := m.JoinVoice(ctx, fixture.GuildID, fixture.VoiceID, session.VoiceJoin{})
	if err != nil || grant.ConnectionID != fixture.Connection {
		t.Fatalf("voice join failed: %+v %v", grant, err)
	}
	if _, ok := m.Voice(fixture.GuildID); !ok {
		t.Fatal("grant was not stored on the owning shard")
	}
	if guild, err := m.ResolveGuild(ctx, fixture.VoiceID); err != nil || guild != fixture.GuildID {
		t.Fatalf("cached channel was not resolved: %q %v", guild, err)
	}
	if guild, err := m.ResolveGuild(ctx, fixture.TextID); err != nil || guild != fixture.GuildID {
		t.Fatalf("REST channel was not resolved: %q %v", guild, err)
	}
}

func TestManagerRejectsGuildsOnOtherProcesses(t *testing.T) {
	m := newManager(t, fixture.New(t), shard.Options{ShardCount: 2, ShardIDs: []int{0}})
	if _, err := m.ForGuild(fixture.GuildID); !errors.Is(err, shard.ErrNotOwned) {
		t.Fatalf("guild on a foreign shard was accepted: %v", err)
	}
	if len(m.Shards()) != 1 || m.ShardCount() != 2 {
		t.Fatal("manager did not keep its shard subset")
	}
}

func TestManagerStopsAllShardsOnFatalClose(t *testing.T) {
	p := fixture.New(t)
	p.CloseCode = 4004
	p.CloseShard = 1
	m := newManager(t, p, shard.Options{ShardCount: 2, IdentifyInterval: time.Millisecond, Session: session.Options{MinBackoff: 10 * time.Millisecond}})
	done := run(t, m)
	select {
	case err := <-done:
		var closed *gateway.CloseError
		if !errors.As(err, &closed) || closed.Code != 4004 {
			t.Fatalf("fatal close was not returned: %v", err)
		}
		if !strings.Contains(err.Error(), "shard 1") {
			t.Fatalf("error does not name the shard: %v", err)
		}
		done <- err
	case <-time.After(5 * time.Second):
		t.Fatal("manager kept running after a fatal close")
	}
}

func TestManagerPresenceReachesEveryShard(t *testing.T) {
	p := fixture.New(t)
	m := newManager(t, p, shard.Options{ShardCount: 2, IdentifyInterval: time.Millisecond})
	run(t, m)
	fixture.Eventually(t, m.Ready)
	ctx := context.Background()
	if err := m.SetCustomStatus(ctx, model.CustomStatus{Text: "!yardım"}); err != nil {
		t.Fatal(err)
	}
	if err := m.SetStatus(ctx, model.StatusIdle); err != nil {
		t.Fatal(err)
	}
	// Shards send on separate sockets, so arrival order is not fixed.
	fixture.Eventually(t, func() bool { return len(p.Presences()) == 4 })
	final := 0
	for _, raw := range p.Presences() {
		if strings.Contains(string(raw), `"idle"`) && strings.Contains(string(raw), "!yardım") {
			final++
		}
	}
	if final != 2 {
		t.Fatalf("each shard must send the status with the kept custom status: %s", p.Presences())
	}
	if presence := m.Presence(); presence.Status != model.StatusIdle || presence.CustomStatus == nil {
		t.Fatalf("manager presence is wrong: %+v", presence)
	}
	if err := m.SetCustomStatus(ctx, model.CustomStatus{}); err == nil || len(p.Presences()) != 4 {
		t.Fatal("an invalid custom status reached a shard")
	}
}

func TestManagerRunsOnce(t *testing.T) {
	m := newManager(t, fixture.New(t), shard.Options{IdentifyInterval: time.Millisecond})
	run(t, m)
	fixture.Eventually(t, m.Ready)
	if err := m.Run(context.Background()); !errors.Is(err, shard.ErrRunning) {
		t.Fatalf("second Run was accepted: %v", err)
	}
}
