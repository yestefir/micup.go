package rest

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

type uploadPlanStats struct {
	puts      atomic.Int64
	completed atomic.Int64
	mu        sync.Mutex
	bodies    map[string]string
}

type countedUploadReader struct {
	reader io.Reader
	read   atomic.Int64
}

func (r *countedUploadReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	r.read.Add(int64(n))
	return n, err
}

func multipartPlanClient(t *testing.T, partSize int64, numbers []int) (*Client, *uploadPlanStats) {
	t.Helper()
	stats := &uploadPlanStats{bodies: map[string]string{}}
	var c *Client
	c = newClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/attachments"):
			parts := []map[string]any{}
			for _, number := range numbers {
				parts = append(parts, map[string]any{"part_number": number, "upload_url": strings.TrimSuffix(c.baseURL, "/api/v1") + "/put/" + strconv.Itoa(number)})
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"attachments": []map[string]any{{"upload_filename": "key", "upload_mode": "multipart", "upload_id": "upload", "part_size": partSize, "parts": parts}}})
		case strings.HasPrefix(r.URL.Path, "/put/"):
			stats.puts.Add(1)
			if r.Header.Get("Authorization") != "" {
				t.Error("upload received bot credentials")
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Errorf("reading uploaded bytes: %v", err)
			}
			stats.mu.Lock()
			stats.bodies[strings.TrimPrefix(r.URL.Path, "/put/")] = string(body)
			stats.mu.Unlock()
			w.WriteHeader(http.StatusNoContent)
		case strings.HasSuffix(r.URL.Path, "/complete"):
			stats.completed.Add(1)
			w.WriteHeader(http.StatusNoContent)
		default:
			http.NotFound(w, r)
		}
	})
	return c, stats
}

func TestMultipartUploadUsesPartNumbers(t *testing.T) {
	for name, numbers := range map[string][]int{"ordered": {1, 2}, "reversed": {2, 1}} {
		t.Run(name, func(t *testing.T) {
			c, stats := multipartPlanClient(t, 4, numbers)
			attachment, err := c.UploadAttachment(context.Background(), "channel", Upload{Name: "file.bin", Size: 6, Reader: strings.NewReader("abcdef")})
			if err != nil || attachment.Size != 6 || stats.completed.Load() != 1 || stats.puts.Load() != 2 {
				t.Fatalf("valid numbered upload failed: %v; puts %d; completions %d", err, stats.puts.Load(), stats.completed.Load())
			}
			stats.mu.Lock()
			defer stats.mu.Unlock()
			if stats.bodies["1"] != "abcd" || stats.bodies["2"] != "ef" {
				t.Errorf("file bytes assigned to wrong numbered parts: %v", stats.bodies)
			}
		})
	}
}

func TestInvalidMultipartPlansDoNotConsumeOrUpload(t *testing.T) {
	for _, item := range []struct {
		name     string
		partSize int64
		numbers  []int
	}{
		{"zero part size", 0, []int{1, 2}},
		{"no parts", 4, nil},
		{"too few parts", 4, []int{1}},
		{"too many parts", 4, []int{1, 2, 3}},
		{"duplicate number", 4, []int{1, 1}},
		{"zero number", 4, []int{0, 1}},
		{"negative number", 4, []int{-1, 1}},
		{"missing number", 4, []int{1, 3}},
	} {
		t.Run(item.name, func(t *testing.T) {
			c, stats := multipartPlanClient(t, item.partSize, item.numbers)
			reader := &countedUploadReader{reader: strings.NewReader("abcdef")}
			_, err := c.UploadAttachment(context.Background(), "channel", Upload{Name: "file.bin", Size: 6, Reader: reader})
			if err == nil || reader.read.Load() != 0 || stats.puts.Load() != 0 || stats.completed.Load() != 0 {
				t.Errorf("invalid plan had side effects: error %v; read %d; puts %d; completions %d", err, reader.read.Load(), stats.puts.Load(), stats.completed.Load())
			}
		})
	}
}

func TestMultipartUploadShortReaderDoesNotComplete(t *testing.T) {
	c, stats := multipartPlanClient(t, 4, []int{1, 2})
	reader := &countedUploadReader{reader: strings.NewReader("abc")}
	_, err := c.UploadAttachment(context.Background(), "channel", Upload{Name: "file.bin", Size: 6, Reader: reader})
	if !errors.Is(err, io.ErrUnexpectedEOF) || stats.puts.Load() != 0 || stats.completed.Load() != 0 || reader.read.Load() != 3 {
		t.Errorf("short read was not preserved: error %v; read %d; puts %d; completions %d", err, reader.read.Load(), stats.puts.Load(), stats.completed.Load())
	}
}
