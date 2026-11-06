package bot_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yestefir/micup.go/src/bot"
	"github.com/yestefir/micup.go/src/internal/fixture"
	"github.com/yestefir/micup.go/src/shard"
)

func platform(t *testing.T, user, application string) string {
	t.Helper()
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/fluxer":
			_ = json.NewEncoder(w).Encode(map[string]any{"endpoints": map[string]string{"api_public": server.URL + "/api", "gateway": "ws" + server.URL[4:] + "/gateway"}})
		case "/api/v1/users/@me":
			fmt.Fprint(w, user)
		case "/api/v1/applications/@me":
			if application == "" {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			fmt.Fprint(w, application)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	return server.URL
}

func TestNewDiscoversBotAndReadsIntent(t *testing.T) {
	origin := platform(t, `{"id":"bot-id","bot":true}`, `{"id":"app-id","flags":262144}`)
	b, err := bot.New(context.Background(), bot.Config{Token: "fixture", Origin: origin})
	if err != nil {
		t.Fatal(err)
	}
	if b.ID != "bot-id" || b.ApplicationID != "app-id" || b.MessageContentEnabled == nil || !*b.MessageContentEnabled {
		t.Fatal("bot identity or intent was not read")
	}
	if b.Endpoints().API != origin+"/api" || b.SessionConfig().BotID != "bot-id" {
		t.Fatal("bot wiring was not built from discovery")
	}
}

func TestNewToleratesMissingApplication(t *testing.T) {
	b, err := bot.New(context.Background(), bot.Config{Token: "fixture", Origin: platform(t, `{"id":"bot-id","bot":true}`, "")})
	if err != nil || b.MessageContentEnabled != nil || b.ApplicationID != "bot-id" {
		t.Fatalf("missing application was not tolerated: %v", err)
	}
}

func TestNewRejectsUserCredential(t *testing.T) {
	if _, err := bot.New(context.Background(), bot.Config{Token: "secret", Origin: platform(t, `{"id":"human","bot":false}`, "")}); err == nil {
		t.Fatal("user credential accepted")
	}
}

func TestNewRejectsInvalidConfig(t *testing.T) {
	for name, config := range map[string]bot.Config{
		"token":  {Token: "a\nb"},
		"shard":  {Token: "t", ShardID: 3, ShardCount: 2},
		"origin": {Token: "t", Origin: "ftp://micup.gg"},
	} {
		if _, err := bot.New(context.Background(), config); err == nil {
			t.Fatalf("%s accepted", name)
		}
	}
}

func TestBotRunsShardManager(t *testing.T) {
	p := fixture.New(t)
	b, err := bot.New(context.Background(), bot.Config{Token: fixture.Token, Origin: p.URL()})
	if err != nil {
		t.Fatal(err)
	}
	m, err := b.NewShardManager(shard.Options{ShardCount: 2, IdentifyInterval: time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- m.Run(ctx) }()
	fixture.Eventually(t, m.Ready)
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("manager did not stop cleanly: %v", err)
	}
}
