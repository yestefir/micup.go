package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/yestefir/micup.go/src/model"
)

func TestResourceRoutes(t *testing.T) {
	c, rec := recordingClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete && !strings.Contains(r.URL.Path, "/tags/") {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		switch {
		case strings.HasSuffix(r.URL.Path, "/audit-logs"):
			fmt.Fprint(w, `{"audit_log_entries":[{"id":"9","action_type":22,"user_id":"1","target_id":"2","options":{"delete_message_seconds":60}}],"users":[{"id":"1"}]}`)
		case strings.HasSuffix(r.URL.Path, "/bans"), strings.HasSuffix(r.URL.Path, "/invites") && r.Method == http.MethodGet, strings.HasSuffix(r.URL.Path, "/emojis") && r.Method == http.MethodGet, strings.HasSuffix(r.URL.Path, "/stickers") && r.Method == http.MethodGet, strings.HasSuffix(r.URL.Path, "/thread-members"), r.URL.Path == "/api/v1/users/@me/channels" && r.Method == http.MethodGet, r.URL.Path == "/api/v1/users/@me/guilds":
			fmt.Fprint(w, `[]`)
		case r.URL.Path == "/api/v1/gateway/bot":
			fmt.Fprint(w, `{"url":"wss://gateway.example","shards":1,"session_start_limit":{"total":1000,"remaining":999,"reset_after":14400000,"max_concurrency":1}}`)
		case strings.HasSuffix(r.URL.Path, "/post-data"):
			fmt.Fprint(w, `{"threads":{"7":{"owner":null,"first_message":{"id":"7"}}}}`)
		case strings.HasSuffix(r.URL.Path, "/refresh-urls"):
			fmt.Fprint(w, `{"refreshed_urls":[{"original":"a","refreshed":"b"}]}`)
		case strings.HasSuffix(r.URL.Path, "/users") && strings.Contains(r.URL.Path, "/reactions/"):
			fmt.Fprint(w, `{"items":[{"id":"1"}],"has_more":false,"next_after":null}`)
		case r.Method == http.MethodPut || (r.Method == http.MethodPatch && (strings.HasSuffix(r.URL.Path, "/channels") || strings.HasSuffix(r.URL.Path, "/roles") || strings.HasSuffix(r.URL.Path, "/hoist-positions"))):
			if strings.Contains(r.URL.Path, "/tags/") {
				fmt.Fprint(w, `{"id":"1"}`)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			fmt.Fprint(w, `{"id":"1","code":"abc","threads":[],"members":[],"messages":[]}`)
		}
	})
	ctx := context.Background()
	everyone := model.Permissions(1024)
	archived := true
	cases := []struct {
		name   string
		call   func() error
		method string
		path   string
		query  string
		body   string
	}{
		{"create channel", func() error {
			_, err := c.CreateChannel(ctx, "5", model.ChannelCreate{Type: model.ChannelTypeGuildText, Name: "general", Options: model.ChannelEdit{Topic: model.Set("hi"), ParentID: model.Set("8")}})
			return err
		}, "POST", "/guilds/5/channels", "", `{"name":"general","parent_id":"8","topic":"hi","type":0}`},
		{"modify channel", func() error {
			_, err := c.ModifyChannel(ctx, "1", model.ChannelEdit{Topic: model.Clear[string](), RateLimitPerUser: model.Set(5), NSFW: model.Set(true)})
			return err
		}, "PATCH", "/channels/1", "", `{"nsfw_override":true,"rate_limit_per_user":5,"topic":null}`},
		{"delete channel", func() error { return c.DeleteChannel(ctx, "1") }, "DELETE", "/channels/1", "", ""},
		{"channel positions", func() error {
			return c.ModifyChannelPositions(ctx, "5", []model.ChannelPosition{{ID: "1", ParentID: model.Clear[string](), Position: model.Set(2), LockPermissions: true}})
		}, "PATCH", "/guilds/5/channels", "", `[{"id":"1","lock_permissions":true,"parent_id":null,"position":2}]`},
		{"set overwrite", func() error {
			return c.SetChannelPermission(ctx, "1", model.PermissionOverwrite{ID: "9", Type: model.OverwriteMember, Allow: model.PermissionViewChannel})
		}, "PUT", "/channels/1/permissions/9", "", `{"allow":"1024","deny":"0","type":1}`},
		{"delete overwrite", func() error { return c.DeleteChannelPermission(ctx, "1", "9") }, "DELETE", "/channels/1/permissions/9", "", ""},
		{"follow", func() error { _, err := c.FollowChannel(ctx, "1", "2"); return err }, "POST", "/channels/1/followers", "", `{"webhook_channel_id":"2"}`},
		{"create role", func() error {
			_, err := c.CreateRole(ctx, "5", model.RoleCreate{Name: "mod", Color: 255, Permissions: &everyone})
			return err
		}, "POST", "/guilds/5/roles", "", `{"color":255,"name":"mod","permissions":"1024"}`},
		{"modify role", func() error {
			_, err := c.ModifyRole(ctx, "5", "6", model.RoleEdit{Hoist: model.Set(true), HoistPosition: model.Clear[int](), UnicodeEmoji: model.Clear[string]()})
			return err
		}, "PATCH", "/guilds/5/roles/6", "", `{"hoist":true,"hoist_position":null,"unicode_emoji":null}`},
		{"delete role", func() error { return c.DeleteRole(ctx, "5", "6") }, "DELETE", "/guilds/5/roles/6", "", ""},
		{"role positions", func() error {
			return c.ModifyRolePositions(ctx, "5", []model.RolePosition{{ID: "6", Position: 3}})
		}, "PATCH", "/guilds/5/roles", "", `[{"id":"6","position":3}]`},
		{"hoist positions", func() error {
			return c.ModifyRoleHoistPositions(ctx, "5", []model.RoleHoistPosition{{ID: "6", HoistPosition: 1}})
		}, "PATCH", "/guilds/5/roles/hoist-positions", "", `[{"hoist_position":1,"id":"6"}]`},
		{"reset hoist", func() error { return c.ResetRoleHoistPositions(ctx, "5") }, "DELETE", "/guilds/5/roles/hoist-positions", "", ""},
		{"modify guild", func() error {
			_, err := c.ModifyGuild(ctx, "5", model.GuildEdit{Name: model.Set("new"), SystemChannelID: model.Clear[string]()})
			return err
		}, "PATCH", "/guilds/5", "", `{"name":"new","system_channel_id":null}`},
		{"guilds", func() error {
			_, err := c.CurrentUserGuilds(ctx, model.GuildQuery{Limit: 50, After: "3", WithCounts: true})
			return err
		}, "GET", "/users/@me/guilds", "after=3&limit=50&with_counts=true", ""},
		{"leave guild", func() error { return c.LeaveGuild(ctx, "5") }, "DELETE", "/users/@me/guilds/5", "", ""},
		{"audit log", func() error {
			log, err := c.AuditLog(ctx, "5", model.AuditLogQuery{Limit: 10, UserID: "1", ActionType: model.AuditMemberBanAdd})
			if err == nil && (len(log.Entries) != 1 || log.Entries[0].Options.DeleteMessageSeconds != "60" || len(log.Users) != 1) {
				err = fmt.Errorf("audit log decoded as %+v", log)
			}
			return err
		}, "GET", "/guilds/5/audit-logs", "action_type=22&limit=10&user_id=1", ""},
		{"bans", func() error { _, err := c.Bans(ctx, "5"); return err }, "GET", "/guilds/5/bans", "", ""},
		{"vanity", func() error { _, err := c.ModifyVanityURL(ctx, "5", ""); return err }, "PATCH", "/guilds/5/vanity-url", "", `{"code":null}`},
		{"gateway info", func() error {
			info, err := c.GatewayInfo(ctx)
			if err == nil && (info.Shards != 1 || info.SessionStartLimit.MaxConcurrency != 1) {
				err = fmt.Errorf("gateway info decoded as %+v", info)
			}
			return err
		}, "GET", "/gateway/bot", "", ""},
		{"current member", func() error { _, err := c.CurrentMember(ctx, "5"); return err }, "GET", "/guilds/5/members/@me", "", ""},
		{"nickname", func() error { _, err := c.SetNickname(ctx, "5", "DJ"); return err }, "PATCH", "/guilds/5/members/@me", "", `{"nick":"DJ"}`},
		{"clear timeout", func() error { _, err := c.ClearTimeout(ctx, "5", "2"); return err }, "PATCH", "/guilds/5/members/2", "", `{"communication_disabled_until":null}`},
		{"search members", func() error {
			_, err := c.SearchMembers(ctx, "5", model.MemberSearch{Query: "a", Limit: 10, RoleIDs: []string{"6"}})
			return err
		}, "POST", "/guilds/5/members-search", "", `{"limit":10,"query":"a","role_ids":["6"]}`},
		{"user", func() error { _, err := c.User(ctx, "2"); return err }, "GET", "/users/2", "", ""},
		{"profile", func() error {
			_, err := c.UserProfile(ctx, "2", model.UserProfileQuery{GuildID: "5", WithMutualGuilds: true})
			return err
		}, "GET", "/users/2/profile", "guild_id=5&with_mutual_guilds=true", ""},
		{"dm channels", func() error { _, err := c.DMChannels(ctx); return err }, "GET", "/users/@me/channels", "", ""},
		{"invite", func() error { _, err := c.Invite(ctx, "abc"); return err }, "GET", "/invites/abc", "", ""},
		{"create invite", func() error {
			invite, err := c.CreateInvite(ctx, "1", model.InviteCreate{MaxAge: 3600, Unique: true})
			if err == nil && invite.Code != "abc" {
				err = fmt.Errorf("invite decoded as %+v", invite)
			}
			return err
		}, "POST", "/channels/1/invites", "", `{"max_age":3600,"max_uses":0,"unique":true}`},
		{"channel invites", func() error { _, err := c.ChannelInvites(ctx, "1"); return err }, "GET", "/channels/1/invites", "", ""},
		{"guild invites", func() error { _, err := c.GuildInvites(ctx, "5"); return err }, "GET", "/guilds/5/invites", "", ""},
		{"delete invite", func() error { return c.DeleteInvite(ctx, "abc") }, "DELETE", "/invites/abc", "", ""},
		{"emojis", func() error { _, err := c.Emojis(ctx, "5"); return err }, "GET", "/guilds/5/emojis", "", ""},
		{"create emoji", func() error {
			_, err := c.CreateEmoji(ctx, "5", model.EmojiCreate{Name: "wave", Image: "data:image/png;base64,AA=="})
			return err
		}, "POST", "/guilds/5/emojis", "", `{"image":"data:image/png;base64,AA==","name":"wave"}`},
		{"clone emoji", func() error { _, err := c.CloneEmoji(ctx, "5", "7"); return err }, "POST", "/guilds/5/emojis/clone", "", `{"source_emoji_id":"7"}`},
		{"rename emoji", func() error { _, err := c.ModifyEmoji(ctx, "5", "7", "hello"); return err }, "PATCH", "/guilds/5/emojis/7", "", `{"name":"hello"}`},
		{"delete emoji", func() error { return c.DeleteEmoji(ctx, "5", "7", true) }, "DELETE", "/guilds/5/emojis/7", "purge=true", ""},
		{"emoji metadata", func() error { _, err := c.EmojiMetadata(ctx, "7"); return err }, "GET", "/emojis/7/metadata", "", ""},
		{"stickers", func() error { _, err := c.Stickers(ctx, "5"); return err }, "GET", "/guilds/5/stickers", "", ""},
		{"modify sticker", func() error {
			_, err := c.ModifySticker(ctx, "5", "7", model.StickerEdit{Name: "cat", Description: "a cat"})
			return err
		}, "PATCH", "/guilds/5/stickers/7", "", `{"description":"a cat","name":"cat","tags":[]}`},
		{"delete sticker", func() error { return c.DeleteSticker(ctx, "5", "7", false) }, "DELETE", "/guilds/5/stickers/7", "", ""},
		{"start thread", func() error {
			_, err := c.StartThread(ctx, "1", model.ThreadStart{Name: "talk", Type: model.ChannelTypePrivateThread, AutoArchiveDuration: model.AutoArchiveDay, Invitable: model.Set(false)})
			return err
		}, "POST", "/channels/1/threads", "", `{"auto_archive_duration":1440,"invitable":false,"name":"talk","type":12}`},
		{"thread from message", func() error {
			_, err := c.StartThreadFromMessage(ctx, "1", "3", model.ThreadStart{Name: "talk"})
			return err
		}, "POST", "/channels/1/messages/3/threads", "", `{"name":"talk"}`},
		{"forum post", func() error {
			_, err := c.CreateForumPost(ctx, "1", model.ForumPostStart{Name: "help", Message: model.MessageOptions{Content: "hi"}, AppliedTags: []string{"4"}})
			return err
		}, "POST", "/channels/1/threads", "", `{"applied_tags":["4"],"message":{"allowed_mentions":{"parse":[]},"content":"hi"},"name":"help"}`},
		{"archive thread", func() error { _, err := c.ArchiveThread(ctx, "1", true); return err }, "PATCH", "/channels/1", "", `{"archived":true}`},
		{"join thread", func() error { return c.JoinThread(ctx, "1") }, "PUT", "/channels/1/thread-members/@me", "", ""},
		{"leave thread", func() error { return c.LeaveThread(ctx, "1") }, "DELETE", "/channels/1/thread-members/@me", "", ""},
		{"add thread member", func() error { return c.AddThreadMember(ctx, "1", "2") }, "PUT", "/channels/1/thread-members/2", "", ""},
		{"thread members", func() error {
			_, err := c.ThreadMembers(ctx, "1", model.ThreadMemberQuery{WithMember: true, Limit: 50})
			return err
		}, "GET", "/channels/1/thread-members", "limit=50&with_member=true", ""},
		{"active threads", func() error { _, err := c.ActiveThreads(ctx, "5"); return err }, "GET", "/guilds/5/threads/active", "", ""},
		{"archived public", func() error {
			_, err := c.PublicArchivedThreads(ctx, "1", model.ArchivedThreadQuery{Limit: 10})
			return err
		}, "GET", "/channels/1/threads/archived/public", "limit=10", ""},
		{"archived joined", func() error {
			_, err := c.JoinedPrivateArchivedThreads(ctx, "1", model.ArchivedThreadQuery{})
			return err
		}, "GET", "/channels/1/users/@me/threads/archived/private", "", ""},
		{"search threads", func() error {
			_, err := c.SearchThreads(ctx, "1", model.ThreadSearch{Name: "bug", Tags: []string{"4", "5"}, Archived: &archived})
			return err
		}, "GET", "/channels/1/threads/search", "archived=true&name=bug&tag=4&tag=5", ""},
		{"create tag", func() error {
			_, err := c.CreateForumTag(ctx, "1", model.ForumTag{Name: "bug", Moderated: true})
			return err
		}, "POST", "/channels/1/tags", "", `{"emoji_id":null,"emoji_name":null,"moderated":true,"name":"bug"}`},
		{"modify tag", func() error {
			_, err := c.ModifyForumTag(ctx, "1", model.ForumTag{ID: "4", Name: "bug"})
			return err
		}, "PUT", "/channels/1/tags/4", "", `{"emoji_id":null,"emoji_name":null,"moderated":false,"name":"bug"}`},
		{"delete tag", func() error { _, err := c.DeleteForumTag(ctx, "1", "4"); return err }, "DELETE", "/channels/1/tags/4", "", ""},
		{"post data", func() error {
			data, err := c.ForumPostData(ctx, "1", []string{"7"})
			if err == nil && (data["7"].FirstMessage == nil || data["7"].FirstMessage.ID != "7") {
				err = fmt.Errorf("post data decoded as %+v", data)
			}
			return err
		}, "POST", "/channels/1/post-data", "", `{"thread_ids":["7"]}`},
		{"reaction users", func() error {
			page, err := c.ReactionUsers(ctx, "1", "3", "wave:7", model.ReactionUsersQuery{Limit: 10})
			if err == nil && len(page.Items) != 1 {
				err = fmt.Errorf("reaction users decoded as %+v", page)
			}
			return err
		}, "GET", "/channels/1/messages/3/reactions/wave:7/users", "limit=10", ""},
		{"crosspost", func() error { _, err := c.CrosspostMessage(ctx, "1", "3"); return err }, "POST", "/channels/1/messages/3/crosspost", "", ""},
		{"delete attachment", func() error { return c.DeleteAttachment(ctx, "1", "3", "4") }, "DELETE", "/channels/1/messages/3/attachments/4", "", ""},
		{"refresh urls", func() error {
			urls, err := c.RefreshAttachmentURLs(ctx, []string{"a"})
			if err == nil && (len(urls) != 1 || urls[0].Refreshed != "b") {
				err = fmt.Errorf("refreshed urls decoded as %+v", urls)
			}
			return err
		}, "POST", "/attachments/refresh-urls", "", `{"attachment_urls":["a"]}`},
		{"forward", func() error {
			_, err := c.ForwardMessage(ctx, "1", model.MessageForward{ChannelID: "2", MessageID: "3"})
			return err
		}, "POST", "/channels/1/messages", "", `{"message_reference":{"channel_id":"2","message_id":"3","type":1}}`},
		{"search", func() error {
			_, err := c.SearchMessages(ctx, model.MessageSearch{GuildID: "5", Content: "hello", Has: []string{model.SearchHasLink}, HitsPerPage: 10})
			return err
		}, "POST", "/search/messages", "", `{"content":"hello","context_guild_id":"5","has":["link"],"hits_per_page":10}`},
		{"webhook", func() error { _, err := c.Webhook(ctx, "8"); return err }, "GET", "/webhooks/8", "", ""},
		{"modify webhook", func() error {
			_, err := c.ModifyWebhook(ctx, "8", model.WebhookEdit{Name: model.Set("news"), Avatar: model.Clear[string]()})
			return err
		}, "PATCH", "/webhooks/8", "", `{"avatar":null,"name":"news"}`},
		{"edit webhook message", func() error {
			content := "fixed"
			_, err := c.EditWebhookMessage(ctx, "8", "tok", "3", "9", model.MessageEdit{Content: &content})
			return err
		}, "PATCH", "/webhooks/8/tok/messages/3", "thread_id=9", `{"content":"fixed"}`},
		{"delete webhook message", func() error { return c.DeleteWebhookMessage(ctx, "8", "tok", "3", "") }, "DELETE", "/webhooks/8/tok/messages/3", "", ""},
		{"delete response", func() error {
			return c.DeleteInteractionResponse(ctx, model.Interaction{ApplicationID: "8", Token: "tok"})
		}, "DELETE", "/webhooks/8/tok/messages/@original", "", ""},
		{"delete followup", func() error {
			return c.DeleteInteractionFollowup(ctx, model.Interaction{ApplicationID: "8", Token: "tok"}, "3")
		}, "DELETE", "/webhooks/8/tok/messages/3", "", ""},
	}
	for _, tc := range cases {
		if err := tc.call(); err != nil {
			t.Errorf("%s: %v", tc.name, err)
			continue
		}
		got := rec.last()
		if got.method != tc.method || got.path != tc.path || got.query != tc.query {
			t.Errorf("%s: sent %s %s?%s, want %s %s?%s", tc.name, got.method, got.path, got.query, tc.method, tc.path, tc.query)
		}
		if tc.body == "" {
			if len(got.raw) != 0 {
				t.Errorf("%s: sent body %s, want none", tc.name, got.raw)
			}
			continue
		}
		var sent, want any
		if json.Unmarshal(got.raw, &sent) != nil || json.Unmarshal([]byte(tc.body), &want) != nil {
			t.Errorf("%s: body %s is not JSON", tc.name, got.raw)
			continue
		}
		a, _ := json.Marshal(sent)
		b, _ := json.Marshal(want)
		if string(a) != string(b) {
			t.Errorf("%s: sent body %s, want %s", tc.name, a, b)
		}
	}
}

func TestWebhookTokenRoutesSendNoBotToken(t *testing.T) {
	c, rec := recordingClient(t, func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, `{"id":"8"}`) })
	ctx := context.Background()
	if _, err := c.WebhookWithToken(ctx, "8", "tok"); err != nil || rec.last().auth != "" {
		t.Fatalf("token lookup sent %q: %v", rec.last().auth, err)
	}
	if _, err := c.ModifyWebhookWithToken(ctx, "8", "tok", model.WebhookEdit{ChannelID: model.Set("1")}); err == nil {
		t.Fatal("a token edit must not move the webhook")
	}
	if _, err := c.Webhook(ctx, "8"); err != nil || rec.last().auth == "" {
		t.Fatalf("bot lookup sent no token: %v", err)
	}
}

func TestResourceValidation(t *testing.T) {
	c, rec := recordingClient(t, nil)
	ctx := context.Background()
	long := strings.Repeat("x", 300)
	checks := map[string]error{
		"channel name":      first(c.CreateChannel(ctx, "5", model.ChannelCreate{Name: ""})),
		"channel edit":      first(c.ModifyChannel(ctx, "1", model.ChannelEdit{})),
		"slowmode":          first(c.ModifyChannel(ctx, "1", model.ChannelEdit{RateLimitPerUser: model.Set(99999)})),
		"overwrite type":    c.SetChannelPermission(ctx, "1", model.PermissionOverwrite{ID: "2", Type: 7}),
		"role name":         first(c.CreateRole(ctx, "5", model.RoleCreate{Name: long})),
		"role icon":         first(c.CreateRole(ctx, "5", model.RoleCreate{Name: "a", Icon: "x", UnicodeEmoji: "y"})),
		"everyone role":     c.DeleteRole(ctx, "5", "5"),
		"invite age":        first(c.CreateInvite(ctx, "1", model.InviteCreate{MaxAge: 9999999})),
		"emoji name":        first(c.CreateEmoji(ctx, "5", model.EmojiCreate{Name: "a b", Image: "x"})),
		"thread type":       first(c.StartThread(ctx, "1", model.ThreadStart{Name: "a", Type: model.ChannelTypeGuildVoice})),
		"archive duration":  first(c.StartThread(ctx, "1", model.ThreadStart{Name: "a", AutoArchiveDuration: 7})),
		"forum post":        first(c.CreateForumPost(ctx, "1", model.ForumPostStart{Name: "a"})),
		"audit cursors":     first(c.AuditLog(ctx, "5", model.AuditLogQuery{Before: "1", After: "2"})),
		"search context":    first(c.SearchMessages(ctx, model.MessageSearch{Content: "a"})),
		"timeout":           first(c.TimeoutMember(ctx, "5", "2", 400*24*time.Hour, "")),
		"forward with text": first(c.SendMessageWithOptions(ctx, "1", model.MessageOptions{Content: "a", Forward: &model.MessageForward{ChannelID: "2", MessageID: "3"}})),
		"embed title":       first(c.SendMessageWithOptions(ctx, "1", model.MessageOptions{Embeds: []model.Embed{{Title: long}}})),
		"guild name":        first(c.ModifyGuild(ctx, "5", model.GuildEdit{Name: model.Clear[string]()})),
		"tag emoji":         first(c.CreateForumTag(ctx, "1", model.ForumTag{Name: "a", EmojiID: &long, EmojiName: &long})),
	}
	for name, err := range checks {
		if err == nil {
			t.Errorf("%s: invalid input was accepted", name)
		}
	}
	if len(rec.requests) != 0 {
		t.Errorf("invalid input sent %d requests", len(rec.requests))
	}
}

func first[T any](_ T, err error) error { return err }

func TestKeptAttachmentsOnEdit(t *testing.T) {
	c, rec := recordingClient(t, func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, `{"id":"3"}`) })
	content := "x"
	if _, err := c.EditMessage(context.Background(), "1", "3", model.MessageEdit{Content: &content}); err != nil {
		t.Fatal(err)
	}
	if _, sent := rec.last().body["attachments"]; sent {
		t.Fatal("an edit without attachment changes must not send attachments")
	}
	if _, err := c.EditMessage(context.Background(), "1", "3", model.MessageEdit{KeepAttachments: []model.KeptAttachment{{ID: "4", Description: model.Set("alt")}}}); err != nil {
		t.Fatal(err)
	}
	kept, _ := json.Marshal(rec.last().body["attachments"])
	if string(kept) != `[{"description":"alt","id":"4"}]` {
		t.Fatalf("kept attachments sent as %s", kept)
	}
	if _, err := c.EditMessage(context.Background(), "1", "3", model.MessageEdit{KeepAttachments: []model.KeptAttachment{}}); err != nil {
		t.Fatal(err)
	}
	if string(rec.last().raw) != `{"attachments":[]}` {
		t.Fatalf("removing every attachment sent %s", rec.last().raw)
	}
}

func TestIterators(t *testing.T) {
	c, rec := recordingClient(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		switch {
		case strings.HasSuffix(r.URL.Path, "/messages"):
			switch q.Get("before") {
			case "":
				fmt.Fprint(w, `[{"id":"30"},{"id":"20"}]`)
			case "20":
				fmt.Fprint(w, `[{"id":"10"}]`)
			}
		case strings.HasSuffix(r.URL.Path, "/users"):
			if q.Get("after") == "" {
				fmt.Fprint(w, `{"items":[{"id":"1"}],"has_more":true,"next_after":"1"}`)
			} else {
				fmt.Fprint(w, `{"items":[{"id":"2"}],"has_more":false,"next_after":null}`)
			}
		}
	})
	ctx := context.Background()
	var ids []string
	for message, err := range c.EachMessage(ctx, "1", model.MessageQuery{Limit: 2}) {
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, message.ID)
	}
	if strings.Join(ids, ",") != "30,20,10" || rec.last().query != "before=20&limit=2" {
		t.Fatalf("messages iterated as %v, last query %s", ids, rec.last().query)
	}
	var users []string
	for user, err := range c.EachReactionUser(ctx, "1", "3", "👍") {
		if err != nil {
			t.Fatal(err)
		}
		users = append(users, user.ID)
	}
	if strings.Join(users, ",") != "1,2" {
		t.Fatalf("reaction users iterated as %v", users)
	}
	for range c.EachMessage(ctx, "1", model.MessageQuery{}) {
		break
	}
	for _, err := range c.EachMessage(ctx, "1", model.MessageQuery{Around: "5"}) {
		if err == nil {
			t.Fatal("iteration around a message must fail")
		}
	}
}

func TestErrorHelpers(t *testing.T) {
	err := fmt.Errorf("wrapped: %w", &APIError{StatusCode: 404, Code: CodeUnknownChannel, Errors: []FieldError{{Path: "name", Code: "TOO_LONG"}}})
	if ErrorCode(err) != CodeUnknownChannel || !IsErrorCode(err, CodeUnknownGuild, CodeUnknownChannel) || !IsNotFound(err) || IsForbidden(err) {
		t.Fatal("API error helpers misread a wrapped error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) || !apiErr.HasFieldCode("TOO_LONG") || apiErr.HasFieldCode("OTHER") {
		t.Fatal("field code lookup failed")
	}
	if ErrorCode(context.Canceled) != "" || IsErrorCode(context.Canceled, "") || IsNotFound(nil) {
		t.Fatal("non-API errors must not match")
	}
}
