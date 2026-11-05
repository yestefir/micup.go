package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/yestefir/micup.go/src/model"
)

type recorded struct {
	method, path, query, reason, contentType string
	body                                     map[string]any
	auth                                     string
	raw                                      []byte
}

type recorder struct {
	mu       sync.Mutex
	requests []recorded
}

func (r *recorder) last() recorded {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.requests[len(r.requests)-1]
}

func recordingClient(t *testing.T, respond func(w http.ResponseWriter, r *http.Request)) (*Client, *recorder) {
	t.Helper()
	rec := &recorder{}
	c := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewReader(raw))
		item := recorded{method: r.Method, path: strings.TrimPrefix(r.URL.EscapedPath(), "/api/v1"), query: r.URL.RawQuery, reason: r.Header.Get("X-Audit-Log-Reason"), contentType: r.Header.Get("Content-Type"), auth: r.Header.Get("Authorization"), raw: raw}
		_ = json.Unmarshal(raw, &item.body)
		rec.mu.Lock()
		rec.requests = append(rec.requests, item)
		rec.mu.Unlock()
		if respond != nil {
			respond(w, r)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	return c, rec
}

func TestMessageEndpoints(t *testing.T) {
	c, rec := recordingClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/messages"):
			fmt.Fprint(w, `[{"id":"5","channel_id":"1","content":"hi","author":{"id":"u"}}]`)
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/pins"):
			fmt.Fprint(w, `{"items":[{"message":{"id":"5"},"pinned_at":"2026-01-01T00:00:00Z"}],"has_more":false}`)
		case r.Method == http.MethodPatch || r.Method == http.MethodGet:
			fmt.Fprint(w, `{"id":"5","channel_id":"1","content":"edited"}`)
		case r.URL.Path == "/api/v1/users/@me/channels":
			fmt.Fprint(w, `{"id":"dm","type":1}`)
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	})
	ctx := context.Background()
	messages, err := c.Messages(ctx, "1", model.MessageQuery{Limit: 10, Before: "9"})
	if err != nil || len(messages) != 1 || messages[0].Author.ID != "u" || rec.last().query != "before=9&limit=10" {
		t.Fatalf("list failed: %v %+v %s", err, messages, rec.last().query)
	}
	content := "edited"
	if m, err := c.EditMessage(ctx, "1", "5", model.MessageEdit{Content: &content}); err != nil || m.Content != "edited" || rec.last().method != http.MethodPatch || rec.last().body["content"] != "edited" {
		t.Fatalf("edit failed: %v", err)
	}
	if err = c.DeleteMessage(ctx, "1", "5", WithReason("cleanup")); err != nil || rec.last().path != "/channels/1/messages/5" || rec.last().reason != "cleanup" {
		t.Fatalf("delete failed: %v %+v", err, rec.last())
	}
	if err = c.BulkDeleteMessages(ctx, "1", []string{"5", "6"}); err != nil || rec.last().path != "/channels/1/messages/bulk-delete" || len(rec.last().body["message_ids"].([]any)) != 2 {
		t.Fatalf("bulk delete failed: %v", err)
	}
	if err = c.BulkDeleteMessages(ctx, "1", nil); err == nil {
		t.Fatal("empty bulk delete accepted")
	}
	if err = c.AddReaction(ctx, "1", "5", "👍"); err != nil || rec.last().method != http.MethodPut || rec.last().path != "/channels/1/messages/5/reactions/%F0%9F%91%8D/@me" {
		t.Fatalf("reaction failed: %v %s", err, rec.last().path)
	}
	if err = c.RemoveUserReaction(ctx, "1", "5", EmojiString(model.Emoji{Name: "party", ID: "77"}), "u"); err != nil || rec.last().path != "/channels/1/messages/5/reactions/party:77/u" {
		t.Fatalf("user reaction removal failed: %v %s", err, rec.last().path)
	}
	if err = c.PinMessage(ctx, "1", "5"); err != nil || rec.last().method != http.MethodPut || rec.last().path != "/channels/1/pins/5" {
		t.Fatalf("pin failed: %v", err)
	}
	if page, err := c.PinnedMessages(ctx, "1", model.PinQuery{Limit: 5}); err != nil || len(page.Items) != 1 || page.Items[0].Message.ID != "5" {
		t.Fatalf("pins failed: %v %+v", err, page)
	}
	if err = c.TriggerTyping(ctx, "1"); err != nil || rec.last().path != "/channels/1/typing" {
		t.Fatalf("typing failed: %v", err)
	}
	if dm, err := c.CreateDM(ctx, "u"); err != nil || dm.ID != "dm" || rec.last().body["recipient_id"] != "u" {
		t.Fatalf("DM failed: %v", err)
	}
}

func TestSendMessageWithFilesUsesMultipart(t *testing.T) {
	c, rec := recordingClient(t, func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Errorf("not multipart: %v", err)
		}
		var payload map[string]any
		_ = json.Unmarshal([]byte(r.FormValue("payload_json")), &payload)
		attachments, _ := payload["attachments"].([]any)
		file, header, err := r.FormFile("files[0]")
		if err != nil || len(attachments) != 1 || header.Filename != "a.txt" || payload["content"] != "see file" {
			t.Errorf("multipart body is wrong: %v %v", err, payload)
		} else {
			data, _ := io.ReadAll(file)
			if string(data) != "hello" {
				t.Errorf("file data lost: %q", data)
			}
		}
		fmt.Fprint(w, `{"id":"5","attachments":[{"id":"a","filename":"a.txt","size":5}]}`)
	})
	m, err := c.SendMessageWithOptions(context.Background(), "1", model.MessageOptions{Content: "see file", Files: []model.File{{Name: "a.txt", ContentType: "text/plain", Reader: strings.NewReader("hello")}}})
	if err != nil || len(m.Attachments) != 1 || !strings.HasPrefix(rec.last().contentType, "multipart/form-data") {
		t.Fatalf("send with file failed: %v %+v", err, m)
	}
}

func TestUploadAttachmentSinglepartAndMultipart(t *testing.T) {
	var server *httptest.Server
	var mu sync.Mutex
	puts := map[string]string{}
	completed := false
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut:
			data, _ := io.ReadAll(r.Body)
			mu.Lock()
			puts[r.URL.Path] = string(data)
			mu.Unlock()
			if r.Header.Get("Authorization") != "" {
				t.Error("upload URL received the bot token")
			}
		case r.URL.Path == "/api/v1/channels/1/attachments":
			var body struct {
				Attachments []struct {
					FileSize int64 `json:"file_size"`
				} `json:"attachments"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body.Attachments[0].FileSize <= 4 {
				fmt.Fprintf(w, `{"attachments":[{"upload_filename":"key-s","upload_mode":"singlepart","upload_url":"%s/put/single","content_type":"text/plain"}]}`, server.URL)
				return
			}
			fmt.Fprintf(w, `{"attachments":[{"upload_filename":"key-m","upload_mode":"multipart","upload_id":"up","part_size":4,"parts":[{"part_number":1,"upload_url":"%[1]s/put/1"},{"part_number":2,"upload_url":"%[1]s/put/2"}]}]}`, server.URL)
		case r.URL.Path == "/api/v1/channels/1/attachments/complete":
			mu.Lock()
			completed = true
			mu.Unlock()
			fmt.Fprint(w, `{"uploads":[{"upload_filename":"key-m"}]}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	c, err := New(Config{BaseURL: server.URL + "/api", Token: "fixture"})
	if err != nil {
		t.Fatal(err)
	}
	single, err := c.UploadAttachment(context.Background(), "1", Upload{Name: "s.txt", Size: 3, Reader: strings.NewReader("abc")})
	if err != nil || single.UploadFilename != "key-s" || puts["/put/single"] != "abc" {
		t.Fatalf("singlepart upload failed: %v %+v", err, puts)
	}
	multi, err := c.UploadAttachment(context.Background(), "1", Upload{Name: "m.bin", Size: 6, Reader: strings.NewReader("abcdef")})
	if err != nil || multi.UploadFilename != "key-m" || puts["/put/1"] != "abcd" || puts["/put/2"] != "ef" || !completed {
		t.Fatalf("multipart upload failed: %v %+v %v", err, puts, completed)
	}
}

func TestSendMessageAttachesUploads(t *testing.T) {
	c, rec := recordingClient(t, func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, `{"id":"5"}`) })
	_, err := c.SendMessageWithOptions(context.Background(), "1", model.MessageOptions{Attachments: []model.UploadedAttachment{{Filename: "m.bin", UploadFilename: "key-m", Size: 6}}, Flags: model.MessageFlagSuppressNotifications})
	attachments, _ := rec.last().body["attachments"].([]any)
	if err != nil || len(attachments) != 1 || attachments[0].(map[string]any)["upload_filename"] != "key-m" || rec.last().body["flags"] != float64(model.MessageFlagSuppressNotifications) {
		t.Fatalf("uploaded attachment was not sent: %v %+v", err, rec.last().body)
	}
}
