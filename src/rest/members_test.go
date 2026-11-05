package rest

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/yestefir/micup.go/src/model"
)

func TestMemberAndModerationEndpoints(t *testing.T) {
	c, rec := recordingClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodPatch:
			if r.URL.Path == "/api/v1/guilds/9/members" {
				fmt.Fprint(w, `[{"user":{"id":"u"},"roles":["r"]}]`)
				return
			}
			fmt.Fprint(w, `{"user":{"id":"u"},"roles":["r"]}`)
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	})
	ctx := context.Background()
	if members, err := c.Members(ctx, "9", model.MemberQuery{Limit: 500, After: "3"}); err != nil || len(members) != 1 || rec.last().query != "after=3&limit=500" {
		t.Fatalf("list members failed: %v", err)
	}
	if _, err := c.ModifyMember(ctx, "9", "u", model.MemberEdit{Disconnect: true}, WithReason("afk")); err != nil || rec.last().reason != "afk" {
		t.Fatalf("disconnect failed: %v", err)
	} else if v, ok := rec.last().body["channel_id"]; !ok || v != nil {
		t.Fatalf("disconnect did not send channel_id null: %+v", rec.last().body)
	}
	channel := "voice"
	if _, err := c.ModifyMember(ctx, "9", "u", model.MemberEdit{ChannelID: &channel, Disconnect: true}); err == nil {
		t.Fatal("move and disconnect together accepted")
	}
	if _, err := c.ModifyMember(ctx, "9", "u", model.MemberEdit{}); err == nil {
		t.Fatal("empty member edit accepted")
	}
	if err := c.AddMemberRole(ctx, "9", "u", "r"); err != nil || rec.last().method != http.MethodPut || rec.last().path != "/guilds/9/members/u/roles/r" {
		t.Fatalf("add role failed: %v", err)
	}
	if err := c.KickMember(ctx, "9", "u", WithReason("rules")); err != nil || rec.last().method != http.MethodDelete || rec.last().path != "/guilds/9/members/u" {
		t.Fatalf("kick failed: %v", err)
	}
	if err := c.BanMember(ctx, "9", "u", model.Ban{DeleteMessageSeconds: 3600, Reason: "spam"}); err != nil || rec.last().path != "/guilds/9/bans/u" || rec.last().body["delete_message_seconds"] != float64(3600) {
		t.Fatalf("ban failed: %v %+v", err, rec.last().body)
	}
	if err := c.BanMember(ctx, "9", "u", model.Ban{DurationSeconds: 10}); err == nil {
		t.Fatal("invalid ban duration accepted")
	}
	if err := c.UnbanMember(ctx, "9", "u"); err != nil || rec.last().method != http.MethodDelete {
		t.Fatalf("unban failed: %v", err)
	}
}
