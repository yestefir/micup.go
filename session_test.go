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
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

type fakePlatform struct {
	t                                     *testing.T
	server                                *httptest.Server
	mu                                    sync.Mutex
	write                                 sync.Mutex
	conn                                  *websocket.Conn
	seq                                   atomic.Int64
	identifies, resumes, leaves, connects atomic.Int64
	channelLookups                        atomic.Int64
	closeCode                             int
	lastResumeSession                     atomic.Value
}

func newFakePlatform(t *testing.T) *fakePlatform {
	f := &fakePlatform{t: t}
	upgrader := websocket.Upgrader{}
	f.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/fluxer":
			_ = json.NewEncoder(w).Encode(map[string]any{"endpoints": map[string]string{"api_public": f.server.URL + "/api", "gateway": "ws" + strings.TrimPrefix(f.server.URL, "http") + "/gateway"}})
		case "/api/v1/users/@me":
			fmt.Fprint(w, `{"id":"bot","bot":true}`)
		case "/api/v1/channels/text":
			f.channelLookups.Add(1)
			fmt.Fprint(w, `{"id":"text","guild_id":"4194304","type":0}`)
		case "/gateway":
			ws, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			f.connects.Add(1)
			f.mu.Lock()
			f.conn = ws
			code := f.closeCode
			f.mu.Unlock()
			defer ws.Close()
			f.send(map[string]any{"op": 10, "d": map[string]int{"heartbeat_interval": 200}})
			for {
				var p struct {
					Op int                        `json:"op"`
					D  map[string]json.RawMessage `json:"d"`
				}
				_, data, e := ws.ReadMessage()
				if e != nil {
					return
				}
				var op struct {
					Op int `json:"op"`
				}
				_ = json.Unmarshal(data, &op)
				if op.Op == OpHeartbeat {
					f.send(map[string]any{"op": OpHeartbeatAck})
					continue
				}
				if json.Unmarshal(data, &p) != nil {
					continue
				}
				switch p.Op {
				case OpIdentify:
					f.identifies.Add(1)
					if code != 0 {
						f.write.Lock()
						_ = ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(code, "rejected"))
						f.write.Unlock()
						return
					}
					f.event("READY", map[string]any{"session_id": "session-1", "guilds": []any{map[string]any{"id": "4194304", "channels": []map[string]string{{"id": "voice"}}, "voice_states": []map[string]string{{"user_id": "listener", "channel_id": "voice"}}}}})
				case OpResume:
					f.resumes.Add(1)
					var session string
					_ = json.Unmarshal(p.D["session_id"], &session)
					f.lastResumeSession.Store(session)
					f.event("RESUMED", map[string]any{})
				case OpVoiceStateUpdate:
					var guild, channel string
					_ = json.Unmarshal(p.D["guild_id"], &guild)
					_ = json.Unmarshal(p.D["channel_id"], &channel)
					if channel == "" {
						f.leaves.Add(1)
						continue
					}
					f.event("VOICE_SERVER_UPDATE", map[string]string{"guild_id": guild, "connection_id": "conn-1", "endpoint": "wss://voice.invalid", "token": "grant-token"})
					f.event("VOICE_STATE_UPDATE", map[string]string{"guild_id": guild, "user_id": "bot", "channel_id": channel})
				}
			}
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(func() {
		f.mu.Lock()
		if f.conn != nil {
			_ = f.conn.Close()
		}
		f.mu.Unlock()
		f.server.Close()
	})
	return f
}

func (f *fakePlatform) send(p any) {
	f.write.Lock()
	defer f.write.Unlock()
	f.mu.Lock()
	conn := f.conn
	f.mu.Unlock()
	if conn != nil {
		_ = conn.SetWriteDeadline(time.Now().Add(time.Second))
		_ = conn.WriteJSON(p)
	}
}

func (f *fakePlatform) event(name string, data any) {
	f.send(map[string]any{"op": OpDispatch, "t": name, "s": f.seq.Add(1), "d": data})
}

func (f *fakePlatform) client(t *testing.T) *Client {
	t.Helper()
	c, err := New(context.Background(), Config{Token: "fixture", Origin: f.server.URL})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func eventually(t *testing.T, check func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if check() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition was not met")
}

func runSession(t *testing.T, s *Session) (context.CancelFunc, chan error) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- s.Run(ctx) }()
	t.Cleanup(func() { cancel(); <-done })
	return cancel, done
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
	f := newFakePlatform(t)
	var events atomic.Int64
	lost := &lostVoice{}
	s := f.client(t).NewSession(SessionOptions{MinBackoff: 10 * time.Millisecond, OnVoiceLost: lost.add, Handler: func(context.Context, Event) error { events.Add(1); return nil }})
	runSession(t, s)
	eventually(t, s.Ready)
	if s.State().VoiceChannel("4194304", "listener") != "voice" || s.State().Guilds() != 1 {
		t.Fatal("READY guilds were not applied to state")
	}
	if guild, ok := s.State().GuildForChannel("voice"); !ok || guild != "4194304" {
		t.Fatal("guild channel was not cached")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	grant, err := s.JoinVoice(ctx, "4194304", "voice", VoiceJoin{SelfDeaf: true})
	if err != nil {
		t.Fatal(err)
	}
	if grant.ConnectionID != "conn-1" || grant.ChannelID != "voice" || grant.Token != "grant-token" {
		t.Fatalf("unexpected voice grant: %+v", grant)
	}
	eventually(t, func() bool { return s.State().BotVoiceChannel("4194304") == "voice" })
	if current, ok := s.Voice("4194304"); !ok || current.ConnectionID != "conn-1" {
		t.Fatal("voice grant was not stored")
	}
	f.send(map[string]any{"op": OpReconnect})
	eventually(t, func() bool { return f.resumes.Load() == 1 && s.Ready() })
	if f.identifies.Load() != 1 || f.lastResumeSession.Load() != "session-1" {
		t.Fatal("reconnect did not resume the existing session")
	}
	f.event("VOICE_STATE_UPDATE", map[string]string{"guild_id": "4194304", "user_id": "bot", "channel_id": ""})
	eventually(t, func() bool { return lost.count() == 1 })
	if lost.last() != "4194304/conn-1" {
		t.Fatalf("voice loss reported the wrong connection: %s", lost.last())
	}
	if _, ok := s.Voice("4194304"); ok {
		t.Fatal("lost voice grant was kept")
	}
	if _, err = s.JoinVoice(ctx, "4194304", "voice", VoiceJoin{}); err != nil {
		t.Fatal(err)
	}
	if err = s.LeaveVoice("4194304", "other"); err != nil || f.leaves.Load() != 0 {
		t.Fatal("leave with a stale connection ID was sent")
	}
	if err = s.LeaveVoice("4194304", "conn-1"); err != nil {
		t.Fatal(err)
	}
	eventually(t, func() bool { return f.leaves.Load() == 1 })
	if _, err = s.JoinVoice(ctx, "4194304", "voice", VoiceJoin{}); err != nil {
		t.Fatal(err)
	}
	f.event("VOICE_SERVER_UPDATE", map[string]string{"guild_id": "4194304", "connection_id": "conn-2", "endpoint": "wss://voice.invalid", "token": "other"})
	eventually(t, func() bool { return lost.count() == 2 })
	if events.Load() == 0 {
		t.Fatal("application handler did not receive events")
	}
}

func TestSessionResolvesGuildOnceThroughREST(t *testing.T) {
	f := newFakePlatform(t)
	s := f.client(t).NewSession(SessionOptions{})
	for range 2 {
		guild, err := s.ResolveGuild(context.Background(), "text")
		if err != nil || guild != "4194304" {
			t.Fatalf("guild was not resolved: %q %v", guild, err)
		}
	}
	if f.channelLookups.Load() != 1 {
		t.Fatal("resolved channel was not cached")
	}
}

func TestSessionStopsOnFatalClose(t *testing.T) {
	f := newFakePlatform(t)
	f.closeCode = 4004
	s := f.client(t).NewSession(SessionOptions{MinBackoff: 10 * time.Millisecond})
	_, done := runSession(t, s)
	select {
	case err := <-done:
		var closed *CloseError
		if !errors.As(err, &closed) || closed.Code != 4004 || !closed.Fatal() {
			t.Fatalf("fatal close was not returned: %v", err)
		}
		done <- err
	case <-time.After(5 * time.Second):
		t.Fatal("session kept reconnecting after a fatal close")
	}
	if f.connects.Load() != 1 {
		t.Fatal("session reconnected after a fatal close")
	}
}

func TestSessionIdentifiesAgainAfterInvalidSession(t *testing.T) {
	f := newFakePlatform(t)
	s := f.client(t).NewSession(SessionOptions{MinBackoff: 10 * time.Millisecond})
	runSession(t, s)
	eventually(t, s.Ready)
	f.send(map[string]any{"op": OpInvalidSession, "d": false})
	eventually(t, func() bool { return f.identifies.Load() == 2 && s.Ready() })
	if f.resumes.Load() != 0 {
		t.Fatal("invalidated session was resumed")
	}
}

func TestJoinVoiceRequiresReadySession(t *testing.T) {
	f := newFakePlatform(t)
	s := f.client(t).NewSession(SessionOptions{})
	if _, err := s.JoinVoice(context.Background(), "4194304", "voice", VoiceJoin{}); !errors.Is(err, ErrNotReady) {
		t.Fatalf("join before READY was accepted: %v", err)
	}
}
