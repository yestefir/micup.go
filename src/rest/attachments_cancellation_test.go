package rest

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestMultipartUploadStopsReadingOnCancellation(t *testing.T) {
	const plan = `{"attachments":[{"upload_filename":"key","upload_mode":"multipart","upload_id":"upload","part_size":4,"parts":[{"part_number":1,"upload_url":"https://fixture.invalid/put/1"},{"part_number":2,"upload_url":"https://fixture.invalid/put/2"}]}]}`
	for _, tc := range []struct {
		name       string
		cancelPlan bool
		cancelRead bool
		wantReads  int
		wantPUTs   int
	}{
		{name: "after plan", cancelPlan: true},
		{name: "during first part", cancelRead: true, wantReads: 1},
		{name: "between parts", wantReads: 4, wantPUTs: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			reader := &cancelOnRead{cancel: func() {}}
			if tc.cancelRead {
				reader.cancel = cancel
			}
			puts, completions := 0, 0
			transport := retryTransport(func(r *http.Request) (*http.Response, error) {
				defer r.Body.Close()
				if r.Context().Err() != nil {
					t.Errorf("canceled request reached the transport: %s", r.URL.Path)
					return nil, r.Context().Err()
				}
				switch r.URL.Path {
				case "/api/v1/channels/channel/attachments":
					var body io.ReadCloser = io.NopCloser(strings.NewReader(plan))
					if tc.cancelPlan {
						body = retryCloseBody{ReadCloser: body, onClose: cancel}
					}
					return &http.Response{StatusCode: http.StatusOK, Body: body}, nil
				case "/put/1", "/put/2":
					puts++
					if n, err := io.Copy(io.Discard, r.Body); err != nil || n != 4 {
						t.Errorf("first PUT copied %d bytes, want 4: %v", n, err)
					}
					body := retryCloseBody{ReadCloser: http.NoBody, onClose: cancel}
					return &http.Response{StatusCode: http.StatusNoContent, Body: body}, nil
				case "/api/v1/channels/channel/attachments/complete":
					completions++
					return &http.Response{StatusCode: http.StatusNoContent, Body: http.NoBody}, nil
				default:
					t.Fatalf("unexpected upload path %q", r.URL.Path)
					return nil, errors.New("unexpected upload path")
				}
			})
			c, err := New(Config{BaseURL: "https://fixture.invalid/api", Token: "fixture", HTTP: &http.Client{Transport: transport}})
			if err != nil {
				t.Fatal(err)
			}
			_, err = c.UploadAttachment(ctx, "channel", Upload{Name: "file.bin", Size: 6, Reader: reader})
			if !errors.Is(err, context.Canceled) {
				t.Errorf("upload error = %v, want context.Canceled", err)
			}
			if reader.reads != tc.wantReads || puts != tc.wantPUTs || completions != 0 {
				t.Errorf("reads, PUTs, completions = (%d, %d, %d), want (%d, %d, 0)", reader.reads, puts, completions, tc.wantReads, tc.wantPUTs)
			}
		})
	}
}
