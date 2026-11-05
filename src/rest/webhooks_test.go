package rest

import (
	"context"
	"fmt"
	"net/http"
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
