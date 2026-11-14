package rest

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/yestefir/micup.go/src/model"
)

func TestWebhookEndpoints(t *testing.T) {
	c, rec := recordingClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/channels/1/webhooks":
			fmt.Fprint(w, `{"id":"w","token":"secret","name":"Now playing"}`)
		case r.Method == http.MethodPost && r.URL.Query().Get("wait") == "true":
			fmt.Fprint(w, `{"id":"m"}`)
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	})
	ctx := context.Background()
	hook, err := c.CreateWebhook(ctx, "1", "Now playing", WithReason("setup"))
	if err != nil || hook.Token != "secret" || rec.last().reason != "setup" {
		t.Fatalf("create webhook failed: %v", err)
	}
	if _, err = c.ExecuteWebhook(ctx, hook.ID, hook.Token, model.WebhookMessage{MessageOptions: model.MessageOptions{Content: "hi"}, Username: "DJ"}); err != nil || rec.last().path != "/webhooks/w/secret" || rec.last().body["username"] != "DJ" || rec.last().auth == "" {
		t.Fatalf("execute without wait failed: %v %+v", err, rec.last())
	}
	created, err := c.ExecuteWebhook(ctx, hook.ID, hook.Token, model.WebhookMessage{MessageOptions: model.MessageOptions{Content: "hi"}, Wait: true})
	if err != nil || created == nil || created.ID != "m" {
		t.Fatalf("execute with wait failed: %v", err)
	}
	if err = c.DeleteWebhook(ctx, hook.ID); err != nil || rec.last().method != http.MethodDelete {
		t.Fatalf("delete webhook failed: %v", err)
	}
}

func TestWebhookCreationNameValidation(t *testing.T) {
	for _, item := range []struct {
		label   string
		channel string
		name    string
		valid   bool
	}{
		{"ascii boundary", "1", strings.Repeat("a", model.MaxWebhookName), true},
		{"unicode boundary", "1", strings.Repeat("é", model.MaxWebhookName), true},
		{"emoji boundary", "1", strings.Repeat("🎵", model.MaxWebhookName), true},
		{"ascii too long", "1", strings.Repeat("a", model.MaxWebhookName+1), false},
		{"unicode too long", "1", strings.Repeat("é", model.MaxWebhookName+1), false},
		{"empty name", "1", "", false},
		{"missing channel", "", "name", false},
	} {
		t.Run(item.label, func(t *testing.T) {
			for _, variant := range []string{"basic", "avatar"} {
				t.Run(variant, func(t *testing.T) {
					c, rec := recordingClient(t, func(w http.ResponseWriter, r *http.Request) {
						fmt.Fprint(w, `{"id":"8"}`)
					})
					var err error
					if variant == "basic" {
						_, err = c.CreateWebhook(context.Background(), item.channel, item.name, WithReason("setup"))
					} else {
						_, err = c.CreateWebhookWithAvatar(context.Background(), item.channel, item.name, "", WithReason("setup"))
					}
					if !item.valid {
						if err == nil || len(rec.requests) != 0 {
							t.Fatalf("invalid name/channel accepted: error %v, requests %d", err, len(rec.requests))
						}
						return
					}
					if err != nil || len(rec.requests) != 1 {
						t.Fatalf("valid name rejected: error %v, requests %d", err, len(rec.requests))
					}
					got := rec.last()
					if got.method != http.MethodPost || got.path != "/channels/1/webhooks" || got.body["name"] != item.name || got.reason != "setup" || got.auth == "" {
						t.Fatalf("webhook creation request changed: %+v", got)
					}
					if _, present := got.body["avatar"]; present {
						t.Fatal("empty avatar must be omitted")
					}
				})
			}
		})
	}
}
