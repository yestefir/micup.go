package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/klauspost/compress/zstd"
)

func TestRunSpeaksZstdStream(t *testing.T) {
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("compress") != CompressionZstdStream || r.URL.Query().Get("stream") != "1" {
			t.Error("compression parameters missing")
		}
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer ws.Close()
		var out bytes.Buffer
		encoder, _ := zstd.NewWriter(&out, zstd.WithEncoderConcurrency(1))
		send := func(v any) {
			out.Reset()
			data, _ := json.Marshal(v)
			_, _ = encoder.Write(data)
			_ = encoder.Flush()
			_ = ws.WriteMessage(websocket.BinaryMessage, out.Bytes())
		}
		pr, pw := io.Pipe()
		go func() {
			for {
				kind, data, err := ws.ReadMessage()
				if err != nil {
					_ = pw.CloseWithError(err)
					return
				}
				if kind != websocket.BinaryMessage {
					t.Error("client sent an uncompressed frame")
				}
				_, _ = pw.Write(data)
			}
		}()
		decoder, _ := zstd.NewReader(pr, zstd.WithDecoderConcurrency(1))
		defer decoder.Close()
		incoming := json.NewDecoder(decoder)
		send(map[string]any{"op": OpHello, "d": map[string]int{"heartbeat_interval": 5000}})
		var identify struct {
			Op int `json:"op"`
			D  struct {
				Token string `json:"token"`
			} `json:"d"`
		}
		if incoming.Decode(&identify) != nil || identify.Op != OpIdentify || identify.D.Token != "fixture" {
			t.Error("compressed identify was not readable")
			return
		}
		send(map[string]any{"op": OpDispatch, "s": 1, "t": EventReady, "d": map[string]string{"session_id": "s", "padding": strings.Repeat("micup ", 2000)}})
		var voice struct {
			Op int `json:"op"`
		}
		if incoming.Decode(&voice) != nil || voice.Op != OpVoiceStateUpdate {
			t.Error("second compressed client message was not readable")
		}
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	done := errors.New("done")
	err := Run(ctx, Options{URL: "ws" + strings.TrimPrefix(server.URL, "http"), Token: "fixture", Compress: true}, func(ctx context.Context, event Event) error {
		if event.Name != EventReady || !strings.Contains(string(event.Data), "micup micup") {
			t.Fatalf("compressed READY was not decoded: %s", event.Name)
		}
		if err := event.Send(ctx, OpVoiceStateUpdate, map[string]any{"guild_id": "g", "channel_id": nil}); err != nil {
			t.Fatal(err)
		}
		return done
	})
	if !errors.Is(err, done) {
		t.Fatalf("compressed session failed: %v", err)
	}
}
