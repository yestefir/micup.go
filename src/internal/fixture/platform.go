package fixture

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/yestefir/micup.go/src/gateway"
)

const (
	Token      = "fixture"
	BotID      = "bot"
	GuildID    = "4194304"
	VoiceID    = "voice"
	LockedID   = "locked"
	TextID     = "text"
	Connection = "conn-1"

	maxMemberInFlight = 2
)

type Identify struct {
	Shard int
	At    time.Time
}

type Platform struct {
	server            *httptest.Server
	mu                sync.Mutex
	conns             []*websocket.Conn
	latest            *peer
	identifyLog       []Identify
	lastIdentify      map[string]json.RawMessage
	presences         []json.RawMessage
	subscriptions     []json.RawMessage
	voiceUpdates      []json.RawMessage
	CloseCode         int
	CloseShard        int
	RefuseVoice       string
	RateLimitMembers  bool
	MemberDelay       time.Duration
	MemberInFlight    atomic.Int64
	MemberDrops       atomic.Int64
	Identifies        atomic.Int64
	Resumes           atomic.Int64
	Leaves            atomic.Int64
	Connects          atomic.Int64
	ChannelLookups    atomic.Int64
	lastResumeSession atomic.Value
}

type peer struct {
	ws    *websocket.Conn
	write sync.Mutex
	seq   int64
}

func New(t testing.TB) *Platform {
	t.Helper()
	p := &Platform{CloseShard: -1}
	upgrader := websocket.Upgrader{}
	p.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/fluxer":
			_ = json.NewEncoder(w).Encode(map[string]any{"endpoints": map[string]string{"api_public": p.APIURL(), "gateway": p.GatewayURL()}})
		case "/api/v1/users/@me":
			fmt.Fprintf(w, `{"id":%q,"bot":true}`, BotID)
		case "/api/v1/channels/" + TextID:
			p.ChannelLookups.Add(1)
			fmt.Fprintf(w, `{"id":%q,"guild_id":%q,"type":0}`, TextID, GuildID)
		case "/gateway":
			ws, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			p.serve(ws)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(func() {
		p.mu.Lock()
		for _, c := range p.conns {
			_ = c.Close()
		}
		p.mu.Unlock()
		p.server.Close()
	})
	return p
}

func (p *Platform) URL() string    { return p.server.URL }
func (p *Platform) APIURL() string { return p.server.URL + "/api" }
func (p *Platform) GatewayURL() string {
	return "ws" + strings.TrimPrefix(p.server.URL, "http") + "/gateway"
}

func (p *Platform) LastResumeSession() string {
	v, _ := p.lastResumeSession.Load().(string)
	return v
}

func (p *Platform) IdentifyLog() []Identify {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]Identify(nil), p.identifyLog...)
}

func (p *Platform) Send(packet any) {
	p.mu.Lock()
	c := p.latest
	p.mu.Unlock()
	if c != nil {
		c.send(packet)
	}
}

func (p *Platform) Event(name string, data any) {
	p.mu.Lock()
	c := p.latest
	p.mu.Unlock()
	if c != nil {
		c.event(name, data)
	}
}

func (c *peer) send(packet any) {
	c.write.Lock()
	defer c.write.Unlock()
	_ = c.ws.SetWriteDeadline(time.Now().Add(time.Second))
	_ = c.ws.WriteJSON(packet)
}

func (c *peer) event(name string, data any) {
	c.write.Lock()
	c.seq++
	seq := c.seq
	c.write.Unlock()
	c.send(map[string]any{"op": gateway.OpDispatch, "t": name, "s": seq, "d": data})
}

func (c *peer) close(code int) {
	c.write.Lock()
	defer c.write.Unlock()
	_ = c.ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(code, "rejected"))
}

func (p *Platform) serve(ws *websocket.Conn) {
	defer ws.Close()
	c := &peer{ws: ws}
	p.Connects.Add(1)
	p.mu.Lock()
	p.conns = append(p.conns, ws)
	p.latest = c
	p.mu.Unlock()
	c.send(map[string]any{"op": gateway.OpHello, "d": map[string]int{"heartbeat_interval": 200}})
	for {
		_, data, err := ws.ReadMessage()
		if err != nil {
			return
		}
		var packet struct {
			Op int             `json:"op"`
			D  json.RawMessage `json:"d"`
		}
		if json.Unmarshal(data, &packet) != nil {
			continue
		}
		var fields map[string]json.RawMessage
		_ = json.Unmarshal(packet.D, &fields)
		switch packet.Op {
		case gateway.OpHeartbeat:
			c.send(map[string]any{"op": gateway.OpHeartbeatAck})
		case gateway.OpRequestGuildMembers:
			var guild, nonce string
			_ = json.Unmarshal(fields["guild_id"], &guild)
			_ = json.Unmarshal(fields["nonce"], &nonce)
			p.mu.Lock()
			limited := p.RateLimitMembers
			p.mu.Unlock()
			if limited {
				c.event(gateway.EventRateLimited, map[string]any{"opcode": gateway.OpRequestGuildMembers, "retry_after": 12.5, "meta": map[string]string{"guild_id": guild, "nonce": nonce}})
				continue
			}
			if p.MemberInFlight.Add(1) > maxMemberInFlight {
				p.MemberInFlight.Add(-1)
				p.MemberDrops.Add(1)
				continue
			}
			p.mu.Lock()
			delay := p.MemberDelay
			p.mu.Unlock()
			go func() {
				defer p.MemberInFlight.Add(-1)
				time.Sleep(delay)
				for i, user := range []string{"listener", "dj"} {
					c.event(gateway.EventGuildMembersChunk, map[string]any{"guild_id": guild, "nonce": nonce, "chunk_index": i, "chunk_count": 2, "members": []map[string]any{{"user": map[string]string{"id": user}, "roles": []string{}}}})
				}
			}()
		case gateway.OpRequestGuildCounts:
			var ids []string
			var nonce string
			_ = json.Unmarshal(fields["guild_ids"], &ids)
			_ = json.Unmarshal(fields["nonce"], &nonce)
			counts := []map[string]any{}
			for _, id := range ids {
				counts = append(counts, map[string]any{"guild_id": id, "member_count": 10, "online_count": 3})
			}
			c.event(gateway.EventGuildCountsUpdate, map[string]any{"counts": counts, "nonce": nonce})
		case gateway.OpRequestChannelMemberCounts:
			var guild, nonce string
			var ids []string
			_ = json.Unmarshal(fields["guild_id"], &guild)
			_ = json.Unmarshal(fields["channel_ids"], &ids)
			_ = json.Unmarshal(fields["nonce"], &nonce)
			counts := []map[string]any{}
			for _, id := range ids {
				counts = append(counts, map[string]any{"guild_id": guild, "channel_id": id, "member_count": 4, "online_count": 2})
			}
			c.event(gateway.EventChannelMemberCountsUpdate, map[string]any{"counts": counts, "nonce": nonce})
		case gateway.OpLazyRequest:
			p.mu.Lock()
			p.subscriptions = append(p.subscriptions, append(json.RawMessage(nil), packet.D...))
			p.mu.Unlock()
		case gateway.OpPresenceUpdate:
			p.mu.Lock()
			p.presences = append(p.presences, append(json.RawMessage(nil), packet.D...))
			p.mu.Unlock()
		case gateway.OpIdentify:
			if !p.identify(c, fields) {
				return
			}
		case gateway.OpResume:
			p.Resumes.Add(1)
			var session string
			_ = json.Unmarshal(fields["session_id"], &session)
			p.lastResumeSession.Store(session)
			c.event(gateway.EventResumed, map[string]any{})
		case gateway.OpVoiceStateUpdate:
			p.mu.Lock()
			p.voiceUpdates = append(p.voiceUpdates, append(json.RawMessage(nil), packet.D...))
			p.mu.Unlock()
			var guild, channel string
			_ = json.Unmarshal(fields["guild_id"], &guild)
			_ = json.Unmarshal(fields["channel_id"], &channel)
			if channel == "" {
				p.Leaves.Add(1)
				continue
			}
			p.mu.Lock()
			refuse := p.RefuseVoice
			p.mu.Unlock()
			if refuse != "" {
				c.send(map[string]any{"op": gateway.OpGatewayError, "d": map[string]string{"code": refuse, "message": "refused"}})
				continue
			}
			c.event(gateway.EventVoiceServerUpdate, map[string]string{"guild_id": guild, "connection_id": Connection, "endpoint": "wss://voice.invalid", "token": "grant-token"})
			c.event(gateway.EventVoiceStateUpdate, map[string]string{"guild_id": guild, "user_id": BotID, "channel_id": channel})
		}
	}
}

func (p *Platform) identify(c *peer, fields map[string]json.RawMessage) bool {
	var pair []int
	_ = json.Unmarshal(fields["shard"], &pair)
	if len(pair) != 2 {
		pair = []int{0, 1}
	}
	p.Identifies.Add(1)
	p.mu.Lock()
	p.identifyLog = append(p.identifyLog, Identify{Shard: pair[0], At: time.Now()})
	p.lastIdentify = fields
	code := p.CloseCode
	if p.CloseShard >= 0 && p.CloseShard != pair[0] {
		code = 0
	}
	p.mu.Unlock()
	if code != 0 {
		c.close(code)
		return false
	}
	guilds := []any{}
	if owner := (4194304 >> 22) % pair[1]; owner == pair[0] {
		guilds = append(guilds, map[string]any{
			"id":           GuildID,
			"properties":   map[string]string{"owner_id": "owner"},
			"roles":        []map[string]string{{"id": GuildID, "permissions": "1049600"}},
			"channels":     []map[string]any{{"id": VoiceID, "type": 2}, {"id": LockedID, "type": 2, "permission_overwrites": []map[string]any{{"id": GuildID, "type": 0, "allow": "0", "deny": "1048576"}}}},
			"members":      []map[string]any{{"user": map[string]string{"id": BotID}, "roles": []string{}}},
			"voice_states": []map[string]string{{"user_id": "listener", "channel_id": VoiceID}},
		})
	}
	c.event(gateway.EventReady, map[string]any{"session_id": fmt.Sprintf("session-%d", pair[0]+1), "guilds": guilds})
	return true
}

func Eventually(t testing.TB, check func() bool) {
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

func (p *Platform) LastIdentify() map[string]json.RawMessage {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.lastIdentify
}

func (p *Platform) Presences() []json.RawMessage {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]json.RawMessage(nil), p.presences...)
}

func (p *Platform) Subscriptions() []json.RawMessage {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]json.RawMessage(nil), p.subscriptions...)
}

func (p *Platform) VoiceUpdates() []json.RawMessage {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]json.RawMessage(nil), p.voiceUpdates...)
}
