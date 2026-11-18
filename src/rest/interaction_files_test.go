package rest

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/yestefir/micup.go/src/model"
)

func TestNewInteractionMessagesDoNotSilentlyDiscardFiles(t *testing.T) {
	for _, followup := range []bool{false, true} {
		for _, uploaded := range []bool{false, true} {
			t.Run(fmt.Sprintf("followup=%v/uploaded=%v", followup, uploaded), func(t *testing.T) {
				c, rec := recordingClient(t, func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprint(w, `{"id":"message"}`)
				})
				options := model.MessageOptions{Content: "see attachment"}
				if uploaded {
					options.Attachments = []model.UploadedAttachment{{Filename: "file.txt", UploadFilename: "fixture-key", Size: 5}}
				} else {
					options.Files = []model.File{{Name: "file.txt", Reader: strings.NewReader("hello")}}
				}
				send := func(options model.MessageOptions) error {
					if followup {
						_, err := c.SendInteractionFollowup(context.Background(), click(), options, false)
						return err
					}
					return c.RespondToInteraction(context.Background(), click(), options, false)
				}
				err := send(options)
				if err == nil {
					t.Error("unsupported file option was silently discarded")
				}
				if len(rec.requests) != 0 {
					t.Errorf("unsupported file option sent %d requests", len(rec.requests))
				}
				if err := send(model.MessageOptions{Content: "text only"}); err != nil || len(rec.requests) != 1 {
					t.Errorf("text-only control returned %v with %d requests", err, len(rec.requests))
				}
			})
		}
	}
}

func TestInteractionUpdateRejectsKeptAttachmentOptions(t *testing.T) {
	for _, tc := range []struct {
		name string
		kept []model.KeptAttachment
	}{
		{"unchanged", nil},
		{"remove all", []model.KeptAttachment{}},
		{"keep selected", []model.KeptAttachment{{ID: "attachment"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c, rec := recordingClient(t, nil)
			content := "updated text"
			err := c.UpdateInteractionMessage(context.Background(), click(), model.MessageEdit{Content: &content, KeepAttachments: tc.kept})
			if tc.kept == nil {
				if err != nil || len(rec.requests) != 1 {
					t.Fatalf("text update returned %v with %d requests", err, len(rec.requests))
				}
			} else if err == nil || len(rec.requests) != 0 {
				t.Fatalf("attachment edit silently discarded: error %v, requests %d", err, len(rec.requests))
			}
		})
	}
}
