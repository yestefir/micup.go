package micup_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	micup "github.com/yestefir/micup.go"
	"github.com/yestefir/micup.go/src/session"
	"github.com/yestefir/micup.go/voice"
)

func TestFacadeBuildsBotFromOneImport(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/fluxer":
			_ = json.NewEncoder(w).Encode(map[string]any{"endpoints": map[string]string{"api_public": server.URL + "/api", "gateway": "ws" + server.URL[4:] + "/gateway"}})
		case "/api/v1/users/@me":
			fmt.Fprint(w, `{"id":"bot-id","bot":true}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	b, err := micup.New(context.Background(), micup.Config{Token: "fixture", Origin: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	var s *micup.Session
	if s, err = b.NewSession(micup.SessionOptions{Handler: func(context.Context, micup.Event) error { return nil }}); err != nil || s.ShardID() != 0 {
		t.Fatalf("session was not built: %v", err)
	}
	var m *micup.ShardManager
	if m, err = b.NewShardManager(micup.ShardOptions{ShardCount: 2}); err != nil || m.ShardCount() != 2 {
		t.Fatalf("shard manager was not built: %v", err)
	}
	if _, err = s.JoinVoice(context.Background(), "1", "2", micup.VoiceJoin{}); !errors.Is(err, micup.ErrNotReady) {
		t.Fatalf("facade error is not the src error: %v", err)
	}
}

func TestFacadeNamesMatchSource(t *testing.T) {
	if micup.ErrVoiceMoved != session.ErrVoiceMoved || micup.EventReady != "READY" || micup.OpVoiceStateUpdate != 4 {
		t.Fatal("facade re-export differs from src")
	}
	if id, err := micup.ShardForGuild("4194304", 2); err != nil || id != 1 {
		t.Fatalf("shard mapping differs: %d %v", id, err)
	}
	if voice.FrameSamples != 1920 {
		t.Fatal("voice facade differs from src")
	}
	var _ micup.MessageOptions = micup.MessageOptions{Embeds: []micup.Embed{{Title: "t"}}}
}
