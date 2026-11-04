package micup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestNewDiscoversBotAndReadsIntent(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/fluxer":
			_ = json.NewEncoder(w).Encode(map[string]any{"endpoints": map[string]string{"api_public": server.URL + "/api", "gateway": "ws" + server.URL[4:] + "/gateway"}})
		case "/api/v1/users/@me":
			if r.Header.Get("Authorization") != "Bot fixture" {
				t.Error("bot token not sent")
			}
			fmt.Fprint(w, `{"id":"bot-id","bot":true}`)
		case "/api/v1/applications/@me":
			fmt.Fprint(w, `{"flags":262144}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	c, err := New(context.Background(), Config{Token: "fixture", Origin: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	if c.BotID != "bot-id" || c.MessageContentEnabled == nil || !*c.MessageContentEnabled {
		t.Fatal("bot identity or intent was not read")
	}
}

func TestNewRejectsBotOnAnotherOrigin(t *testing.T) {
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Error("foreign origin received a request") }))
	defer other.Close()
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/fluxer" {
			_ = json.NewEncoder(w).Encode(map[string]any{"endpoints": map[string]string{"api_public": other.URL, "gateway": "ws" + server.URL[4:] + "/gateway"}})
			return
		}
		if r.URL.Path == "/api/v1/users/@me" {
			fmt.Fprint(w, `{"id":"bot-id","bot":true}`)
		}
	}))
	defer server.Close()
	if _, err := New(context.Background(), Config{Token: "fixture", Origin: server.URL}); err == nil {
		t.Fatal("foreign API origin accepted")
	}
}

func TestRESTRejectsNonRelativePath(t *testing.T) {
	c := &Client{apiURL: "https://micup.example/api/v1"}
	for _, path := range []string{"https://attacker.test/", "//attacker.test/"} {
		if err := c.REST(context.Background(), http.MethodGet, path, nil, nil); err == nil {
			t.Fatalf("invalid path accepted: %s", path)
		}
	}
}

func TestGatewayIdentifiesAndDeliversEvents(t *testing.T) {
	var server *httptest.Server
	upgrader := websocket.Upgrader{}
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/fluxer":
			_ = json.NewEncoder(w).Encode(map[string]any{"endpoints": map[string]string{"api_public": server.URL + "/api", "gateway": "ws" + strings.TrimPrefix(server.URL, "http") + "/gateway"}})
		case "/api/v1/users/@me":
			fmt.Fprint(w, `{"id":"bot-id","bot":true}`)
		case "/gateway":
			ws, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			defer ws.Close()
			_ = ws.WriteJSON(map[string]any{"op": 10, "d": map[string]int{"heartbeat_interval": 5000}})
			var identify struct {
				Op   int `json:"op"`
				Data struct {
					Token string `json:"token"`
					Shard []int  `json:"shard"`
				} `json:"d"`
			}
			if ws.ReadJSON(&identify) != nil || identify.Op != 2 || identify.Data.Token != "fixture" || len(identify.Data.Shard) != 2 || identify.Data.Shard[1] != 2 {
				t.Error("invalid Identify packet")
				return
			}
			_ = ws.WriteJSON(map[string]any{"op": 0, "s": 17, "t": "READY", "d": map[string]string{"session_id": "session"}})
			var sent struct {
				Op   int               `json:"op"`
				Data map[string]string `json:"d"`
			}
			if ws.ReadJSON(&sent) != nil || sent.Op != 4 || sent.Data["guild_id"] != "guild" {
				t.Error("Gateway event could not send an opcode")
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	client, err := New(context.Background(), Config{Token: "fixture", Origin: server.URL, ShardID: 1, ShardCount: 2})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	want := errors.New("handler complete")
	err = client.Gateway(ctx, func(ctx context.Context, event Event) error {
		if event.Name != "READY" || event.Sequence == nil || *event.Sequence != 17 {
			t.Fatal("Gateway event fields were lost")
		}
		var ready struct {
			SessionID string `json:"session_id"`
		}
		if json.Unmarshal(event.Data, &ready) != nil || ready.SessionID != "session" {
			t.Fatal("Gateway event data was changed")
		}
		if err := event.Send(ctx, 4, map[string]string{"guild_id": "guild", "channel_id": "voice"}); err != nil {
			t.Fatal(err)
		}
		return want
	})
	if !errors.Is(err, want) {
		t.Fatalf("Gateway returned the wrong handler result: %v", err)
	}
}
