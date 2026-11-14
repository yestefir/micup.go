package rest

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/yestefir/micup.go/src/model"
)

type cancelOnRead struct {
	cancel context.CancelFunc
	reads  int
}

func (r *cancelOnRead) Read(p []byte) (int, error) {
	r.reads++
	if r.reads == 1 {
		r.cancel()
	}
	if r.reads == 8 {
		return 0, io.EOF
	}
	p[0] = 'x'
	return 1, nil
}

func TestSendMessageStopsReadingCanceledMultipart(t *testing.T) {
	for _, tc := range []struct {
		name        string
		preCanceled bool
		wantReads   int
	}{
		{name: "before encoding", preCanceled: true},
		{name: "during encoding", wantReads: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newClient(t, func(w http.ResponseWriter, r *http.Request) {
				t.Error("canceled multipart request reached the server")
				w.WriteHeader(http.StatusNoContent)
			})
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if tc.preCanceled {
				cancel()
			}
			first := &cancelOnRead{cancel: cancel}
			second := &cancelOnRead{cancel: cancel}
			_, err := c.SendMessageWithOptions(ctx, "1", model.MessageOptions{Files: []model.File{
				{Name: "first.txt", Reader: first},
				{Name: "second.txt", Reader: second},
			}})
			if !errors.Is(err, context.Canceled) {
				t.Errorf("send error = %v, want context.Canceled", err)
			}
			if first.reads != tc.wantReads || second.reads != 0 {
				t.Errorf("reader calls = (%d, %d), want (%d, 0)", first.reads, second.reads, tc.wantReads)
			}
		})
	}
}

func TestMultipartRetryReusesEncodedBody(t *testing.T) {
	var attempts [][]byte
	c, err := New(Config{BaseURL: "https://fixture.invalid/api", Token: "fixture", HTTP: &http.Client{Transport: retryTransport(func(r *http.Request) (*http.Response, error) {
		defer r.Body.Close()
		data, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		attempts = append(attempts, data)
		if len(attempts) == 1 {
			return &http.Response{StatusCode: http.StatusTooManyRequests, Body: io.NopCloser(strings.NewReader(`{"retry_after":0.001}`))}, nil
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"id":"message"}`))}, nil
	})}})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	message, err := c.SendMessageWithOptions(ctx, "1", model.MessageOptions{Files: []model.File{
		{Name: "file.txt", Reader: strings.NewReader("fixture attachment")},
	}})
	if err != nil || message.ID != "message" {
		t.Fatalf("multipart send = (%+v, %v), want successful retry", message, err)
	}
	if len(attempts) != 2 || !bytes.Equal(attempts[0], attempts[1]) || !bytes.Contains(attempts[0], []byte("fixture attachment")) {
		t.Fatalf("expected two identical multipart bodies containing the file, got %d attempts", len(attempts))
	}
}
