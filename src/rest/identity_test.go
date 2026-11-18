package rest

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/yestefir/micup.go/src/model"
)

func TestTokenIDReadsTheLeadingSnowflake(t *testing.T) {
	head := base64.RawStdEncoding.EncodeToString([]byte("1568266374592593922"))
	for token, want := range map[string]string{head + ".time.secret": "1568266374592593922", head + "=.x.y": "1568266374592593922"} {
		if got, ok := TokenID(token); !ok || got != want {
			t.Errorf("TokenID(%q) = %q, %v", token, got, ok)
		}
	}
	for _, token := range []string{"", "nodots", ".x.y", "%%%.x.y", base64.RawStdEncoding.EncodeToString([]byte("abc")) + ".x.y"} {
		if got, ok := TokenID(token); ok {
			t.Errorf("TokenID(%q) = %q, want failure", token, got)
		}
	}
}

func identityServer(t *testing.T, appJSON string) (*Client, *recorder) {
	t.Helper()
	return recordingClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch strings.TrimPrefix(r.URL.Path, "/api/v1") {
		case "/users/@me":
			fmt.Fprint(w, `{"id":"bot1","username":"b","bot":true}`)
		case "/applications/@me":
			if appJSON == "" {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			fmt.Fprint(w, appJSON)
		case "/oauth2/applications/app1/public":
			fmt.Fprint(w, `{"id":"app1","name":"n","scopes":["bot"],"bot_public":true,"bot":{"id":"bot1","username":"b"}}`)
		default:
			fmt.Fprint(w, `[]`)
		}
	})
}

func TestIdentityIsFetchedOnceAndCached(t *testing.T) {
	c, rec := identityServer(t, `{"id":"app1","name":"n","bot_public":true,"owner":{"id":"owner1","username":"o"}}`)
	ctx := context.Background()
	for range 3 {
		if id, err := c.SelfID(ctx); err != nil || id != "bot1" {
			t.Fatalf("SelfID = %q, %v", id, err)
		}
		app, err := c.Application(ctx)
		if err != nil || app.ID != "app1" || app.Owner == nil || app.Owner.ID != "owner1" || !app.BotPublic {
			t.Fatalf("Application = %+v, %v", app, err)
		}
	}
	if len(rec.requests) != 2 {
		t.Fatalf("requests = %d, want one per lookup", len(rec.requests))
	}
}

func TestCommandsUseTheBotsOwnApplicationWhenIDIsEmpty(t *testing.T) {
	c, rec := identityServer(t, `{"id":"app1","name":"n"}`)
	ctx := context.Background()
	if _, err := c.GuildCommands(ctx, "", "g"); err != nil {
		t.Fatal(err)
	}
	if got := rec.last().path; got != "/applications/app1/guilds/g/commands" {
		t.Fatalf("path = %s", got)
	}
	if _, err := c.SetGlobalCommands(ctx, "explicit", []model.ApplicationCommand{}); err != nil || rec.last().path != "/applications/explicit/commands" {
		t.Fatalf("explicit ID was replaced: %v %s", err, rec.last().path)
	}
}

func TestApplicationIDFallsBackToTheUserID(t *testing.T) {
	c, _ := identityServer(t, "")
	id, err := c.ApplicationID(context.Background())
	if err != nil || id != "bot1" {
		t.Fatalf("ApplicationID = %q, %v", id, err)
	}
}

func TestPublicApplicationDefaultsToTheBotsApplication(t *testing.T) {
	c, rec := identityServer(t, `{"id":"app1","name":"n"}`)
	app, err := c.PublicApplication(context.Background(), "")
	if err != nil || app.ID != "app1" || len(app.Scopes) != 1 || app.Bot == nil || app.Bot.ID != "bot1" {
		t.Fatalf("PublicApplication = %+v, %v", app, err)
	}
	if rec.last().path != "/oauth2/applications/app1/public" {
		t.Fatalf("path = %s", rec.last().path)
	}
}
