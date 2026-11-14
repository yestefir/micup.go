package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

type queuedPayload struct {
	encoded chan struct{}
	data    json.RawMessage
}

func (p queuedPayload) MarshalJSON() ([]byte, error) {
	close(p.encoded)
	return p.data, nil
}

func TestConnDoesNotSendCancelledQueuedCommands(t *testing.T) {
	for _, opcode := range []int{OpVoiceStateUpdate, OpHeartbeat} {
		name := "command"
		data := json.RawMessage(`{"guild_id":"guild","channel_id":"cancelled"}`)
		if opcode == OpHeartbeat {
			name, data = "heartbeat", json.RawMessage(`null`)
		}
		t.Run(name, func(t *testing.T) {
			frames := make(chan []byte, 2)
			upgrader := websocket.Upgrader{}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ws, err := upgrader.Upgrade(w, r, nil)
				if err != nil {
					return
				}
				defer ws.Close()
				_ = ws.SetReadDeadline(time.Now().Add(3 * time.Second))
				for {
					_, frame, err := ws.ReadMessage()
					if err != nil {
						return
					}
					frames <- frame
				}
			}))
			defer server.Close()
			ws, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), nil)
			if err != nil {
				t.Fatal(err)
			}
			defer ws.Close()
			c := newConn(ws)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			encoded := make(chan struct{})
			done := make(chan error, 1)
			func() {
				c.writeGate <- struct{}{}
				defer func() { <-c.writeGate }()
				go func() { done <- c.Send(ctx, opcode, queuedPayload{encoded: encoded, data: data}) }()
				select {
				case <-encoded:
				case <-time.After(3 * time.Second):
					t.Fatal("command did not reach serialization")
				}
				cancel()
			}()
			if err := <-done; !errors.Is(err, context.Canceled) {
				t.Errorf("queued cancellation returned %v, want context canceled", err)
			}
			if opcode == OpHeartbeat && c.heartbeatAt.Load() != 0 {
				t.Error("cancelled heartbeat changed the acknowledgement timing state")
			}
			if err := c.Send(context.Background(), OpVoiceStateUpdate, map[string]string{"guild_id": "guild", "channel_id": "active"}); err != nil {
				t.Fatal(err)
			}
			select {
			case frame := <-frames:
				var packet struct {
					Op int `json:"op"`
					D  struct {
						ChannelID string `json:"channel_id"`
					} `json:"d"`
				}
				if err := json.Unmarshal(frame, &packet); err != nil || packet.Op != OpVoiceStateUpdate || packet.D.ChannelID != "active" {
					t.Errorf("cancelled frame reached the Gateway: %s (decode error %v)", frame, err)
				}
			case <-time.After(3 * time.Second):
				t.Fatal("active command was not delivered")
			}
		})
	}
}
