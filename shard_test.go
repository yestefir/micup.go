package micup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

type identifyRecord struct {
	shard int
	at    time.Time
}

type fakeShardPlatform struct {
	server      *httptest.Server
	mu          sync.Mutex
	identifies  []identifyRecord
	conns       []*websocket.Conn
	rejectShard int
}

func newFakeShardPlatform(t *testing.T) *fakeShardPlatform {
	f := &fakeShardPlatform{rejectShard: -1}
	upgrader := websocket.Upgrader{}
	f.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/fluxer":
			_ = json.NewEncoder(w).Encode(map[string]any{"endpoints": map[string]string{"api_public": f.server.URL + "/api", "gateway": "ws" + strings.TrimPrefix(f.server.URL, "http") + "/gateway"}})
		case "/api/v1/users/@me":
			fmt.Fprint(w, `{"id":"bot","bot":true}`)
		case "/api/v1/channels/text":
			fmt.Fprint(w, `{"id":"text","guild_id":"4194304","type":0}`)
		case "/gateway":
			ws, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			f.mu.Lock()
			f.conns = append(f.conns, ws)
			f.mu.Unlock()
			f.serve(ws)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(func() {
		f.mu.Lock()
		for _, c := range f.conns {
			_ = c.Close()
		}
		f.mu.Unlock()
		f.server.Close()
	})
	return f
}

func (f *fakeShardPlatform) serve(ws *websocket.Conn) {
	defer ws.Close()
	var write sync.Mutex
	send := func(p any) {
		write.Lock()
		defer write.Unlock()
		_ = ws.SetWriteDeadline(time.Now().Add(time.Second))
		_ = ws.WriteJSON(p)
	}
	seq := int64(0)
	event := func(name string, data any) {
		seq++
		send(map[string]any{"op": OpDispatch, "t": name, "s": seq, "d": data})
	}
	send(map[string]any{"op": OpHello, "d": map[string]int{"heartbeat_interval": 200}})
	for {
		var p struct {
			Op int                        `json:"op"`
			D  map[string]json.RawMessage `json:"d"`
		}
		_, data, err := ws.ReadMessage()
		if err != nil {
			return
		}
		if json.Unmarshal(data, &p) != nil {
			send(map[string]any{"op": OpHeartbeatAck})
			continue
		}
		switch p.Op {
		case OpHeartbeat:
			send(map[string]any{"op": OpHeartbeatAck})
		case OpIdentify:
			var pair []int
			_ = json.Unmarshal(p.D["shard"], &pair)
			f.mu.Lock()
			f.identifies = append(f.identifies, identifyRecord{shard: pair[0], at: time.Now()})
			reject := f.rejectShard == pair[0]
			f.mu.Unlock()
			if reject {
				write.Lock()
				_ = ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(4004, "rejected"))
				write.Unlock()
				return
			}
			guilds := []any{}
			if pair[0] == 1 {
				guilds = append(guilds, map[string]any{"id": "4194304", "channels": []map[string]string{{"id": "voice"}}})
			}
			event(EventReady, map[string]any{"session_id": fmt.Sprintf("session-%d", pair[0]), "guilds": guilds})
		case OpVoiceStateUpdate:
			var guild, channel string
			_ = json.Unmarshal(p.D["guild_id"], &guild)
			_ = json.Unmarshal(p.D["channel_id"], &channel)
			if channel != "" {
				event(EventVoiceServerUpdate, map[string]string{"guild_id": guild, "connection_id": "conn-1", "endpoint": "wss://voice.invalid", "token": "grant-token"})
			}
		}
	}
}

func (f *fakeShardPlatform) records() []identifyRecord {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]identifyRecord(nil), f.identifies...)
}

func (f *fakeShardPlatform) client(t *testing.T) *Client {
	t.Helper()
	c, err := New(context.Background(), Config{Token: "fixture", Origin: f.server.URL})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func runManager(t *testing.T, m *ShardManager) chan error {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- m.Run(ctx) }()
	t.Cleanup(func() { cancel(); <-done })
	return done
}

func TestShardForGuild(t *testing.T) {
	shard, err := ShardForGuild("4194304", 2)
	if err != nil || shard != 1 {
		t.Fatalf("unexpected shard: %d %v", shard, err)
	}
	if _, err = ShardForGuild("guild", 2); err == nil {
		t.Fatal("non-numeric guild accepted")
	}
	if _, err = ShardForGuild("1", 0); err == nil {
		t.Fatal("zero shard count accepted")
	}
}

func TestShardManagerRejectsInvalidOptions(t *testing.T) {
	c := newFakeShardPlatform(t).client(t)
	for name, options := range map[string]ShardManagerOptions{
		"duplicate":    {ShardCount: 2, ShardIDs: []int{1, 1}},
		"out of range": {ShardCount: 2, ShardIDs: []int{2}},
		"negative":     {ShardCount: 2, ShardIDs: []int{-1}},
		"zero count":   {ShardCount: -1},
	} {
		if _, err := c.NewShardManager(options); err == nil {
			t.Fatalf("%s options accepted", name)
		}
	}
}

func TestShardManagerIdentifiesEveryShardInTurn(t *testing.T) {
	f := newFakeShardPlatform(t)
	var mu sync.Mutex
	seen := map[int]bool{}
	interval := 80 * time.Millisecond
	m, err := f.client(t).NewShardManager(ShardManagerOptions{ShardCount: 3, IdentifyInterval: interval, Session: SessionOptions{Handler: func(_ context.Context, event Event) error {
		if event.Name == EventReady {
			mu.Lock()
			seen[event.ShardID] = true
			mu.Unlock()
		}
		return nil
	}}})
	if err != nil {
		t.Fatal(err)
	}
	runManager(t, m)
	eventually(t, m.Ready)
	records := f.records()
	if len(records) != 3 {
		t.Fatalf("expected 3 identifies, got %d", len(records))
	}
	shards := map[int]bool{}
	for i, record := range records {
		shards[record.shard] = true
		if i > 0 && record.at.Sub(records[i-1].at) < interval-10*time.Millisecond {
			t.Fatalf("identify %d came %v after the previous one", i, record.at.Sub(records[i-1].at))
		}
	}
	mu.Lock()
	defer mu.Unlock()
	if len(shards) != 3 || len(seen) != 3 {
		t.Fatalf("not every shard identified and reported READY: %v %v", shards, seen)
	}
}

func TestShardManagerRoutesGuildsToOwningShard(t *testing.T) {
	f := newFakeShardPlatform(t)
	m, err := f.client(t).NewShardManager(ShardManagerOptions{ShardCount: 2, IdentifyInterval: time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	runManager(t, m)
	eventually(t, m.Ready)
	session, err := m.ForGuild("4194304")
	if err != nil || session.ShardID() != 1 {
		t.Fatalf("guild was not routed to shard 1: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	grant, err := m.JoinVoice(ctx, "4194304", "voice", VoiceJoin{})
	if err != nil || grant.ConnectionID != "conn-1" {
		t.Fatalf("voice join failed: %+v %v", grant, err)
	}
	if _, ok := m.Voice("4194304"); !ok {
		t.Fatal("grant was not stored on the owning shard")
	}
	if guild, err := m.ResolveGuild(ctx, "voice"); err != nil || guild != "4194304" {
		t.Fatalf("cached channel was not resolved: %q %v", guild, err)
	}
	if guild, err := m.ResolveGuild(ctx, "text"); err != nil || guild != "4194304" {
		t.Fatalf("REST channel was not resolved: %q %v", guild, err)
	}
}

func TestShardManagerRejectsGuildsOnOtherProcesses(t *testing.T) {
	f := newFakeShardPlatform(t)
	m, err := f.client(t).NewShardManager(ShardManagerOptions{ShardCount: 2, ShardIDs: []int{0}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = m.ForGuild("4194304"); !errors.Is(err, ErrShardNotOwned) {
		t.Fatalf("guild on a foreign shard was accepted: %v", err)
	}
	if len(m.Shards()) != 1 || m.ShardCount() != 2 {
		t.Fatal("manager did not keep its shard subset")
	}
}

func TestShardManagerStopsAllShardsOnFatalClose(t *testing.T) {
	f := newFakeShardPlatform(t)
	f.rejectShard = 1
	m, err := f.client(t).NewShardManager(ShardManagerOptions{ShardCount: 2, IdentifyInterval: time.Millisecond, Session: SessionOptions{MinBackoff: 10 * time.Millisecond}})
	if err != nil {
		t.Fatal(err)
	}
	done := runManager(t, m)
	select {
	case err := <-done:
		var closed *CloseError
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

func TestShardManagerRunsOnce(t *testing.T) {
	f := newFakeShardPlatform(t)
	m, err := f.client(t).NewShardManager(ShardManagerOptions{IdentifyInterval: time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	runManager(t, m)
	eventually(t, m.Ready)
	if err := m.Run(context.Background()); !errors.Is(err, ErrShardManagerRunning) {
		t.Fatalf("second Run was accepted: %v", err)
	}
}
