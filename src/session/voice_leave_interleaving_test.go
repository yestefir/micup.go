package session

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/yestefir/micup.go/src/gateway"
	"github.com/yestefir/micup.go/src/model"
)

type blockedLeaveConn struct {
	net.Conn
	block   atomic.Bool
	started chan struct{}
	release chan struct{}
	fail    bool
}

func (c *blockedLeaveConn) Write(p []byte) (int, error) {
	if c.block.CompareAndSwap(true, false) {
		close(c.started)
		<-c.release
		if c.fail {
			return 0, io.ErrClosedPipe
		}
	}
	return c.Conn.Write(p)
}

func TestVoiceLeaveCleanupDuringGrantChanges(t *testing.T) {
	// gateway.Run copies DefaultDialer when opening its connection. These tests
	// run serially and restore the dialer after all Gateway work has stopped.
	for _, mode := range []string{"unchanged", "rotation", "replacement", "reset"} {
		for _, fail := range []bool{false, true} {
			name := mode + "/sent"
			if fail {
				name = mode + "/failed"
			}
			t.Run(name, func(t *testing.T) {
				frames := make(chan map[string]any, 1)
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					ws, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
					if err != nil {
						return
					}
					defer ws.Close()
					var frame map[string]any
					if ws.ReadJSON(&frame) == nil {
						frames <- frame
					}
				}))
				defer server.Close()
				transports := make(chan *blockedLeaveConn, 1)
				originalDialer := websocket.DefaultDialer
				dialer := *originalDialer
				dialer.NetDialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
					conn, err := (&net.Dialer{}).DialContext(ctx, network, addr)
					if err != nil {
						return nil, err
					}
					transport := &blockedLeaveConn{Conn: conn, started: make(chan struct{}), release: make(chan struct{}), fail: fail}
					transports <- transport
					return transport, nil
				}
				websocket.DefaultDialer = &dialer
				defer func() { websocket.DefaultDialer = originalDialer }()
				grant := model.VoiceServerUpdate{GuildID: "guild", ChannelID: "voice", ConnectionID: "old", Endpoint: "wss://fixture.invalid", Token: "old-token"}
				s := &Session{grants: map[string]model.VoiceServerUpdate{"guild": grant}, options: Options{OnVoiceGrant: func(model.VoiceServerUpdate) {}}}
				ctx, cancel := context.WithCancel(context.Background())
				attached := make(chan struct{})
				runDone := make(chan error, 1)
				go func() {
					runDone <- gateway.Run(ctx, gateway.Options{
						URL: "ws" + strings.TrimPrefix(server.URL, "http"), Token: "fixture",
						OnConnect: func(conn *gateway.Conn) { s.attach(conn); close(attached) },
					}, func(context.Context, gateway.Event) error { return nil })
				}()
				defer func() {
					cancel()
					select {
					case <-runDone:
					case <-time.After(3 * time.Second):
						t.Error("Gateway did not stop")
					}
				}()
				select {
				case <-attached:
				case <-time.After(3 * time.Second):
					t.Fatal("Gateway did not attach")
				}
				transport := <-transports
				var release sync.Once
				unblock := func() { release.Do(func() { close(transport.release) }) }
				defer unblock()
				transport.block.Store(true)
				leaveDone := make(chan error, 1)
				go func() { leaveDone <- s.LeaveVoice("guild", "old") }()
				select {
				case <-transport.started:
				case <-time.After(3 * time.Second):
					t.Fatal("leave write did not reach the barrier")
				}
				want, wantPresent := grant, true
				switch mode {
				case "rotation", "replacement":
					want.Token = "new-token"
					if mode == "replacement" {
						want.ConnectionID = "new"
					}
					data, err := json.Marshal(want)
					if err != nil {
						t.Fatal(err)
					}
					s.handleVoiceServerUpdate(gateway.Event{Data: data})
				case "reset":
					s.dropAllGrants()
					wantPresent = false
				}
				unblock()
				select {
				case err := <-leaveDone:
					if (fail && !errors.Is(err, io.ErrClosedPipe)) || (!fail && err != nil) {
						t.Fatalf("leave result = %v, fail=%v", err, fail)
					}
				case <-time.After(3 * time.Second):
					t.Fatal("leave did not finish")
				}
				if mode == "unchanged" && !fail {
					wantPresent = false
				}
				if got, present := s.Voice("guild"); present != wantPresent || (present && got != want) {
					t.Errorf("grant after %s: present=%v, want present=%v and expected snapshot", name, present, wantPresent)
				}
				if !fail {
					select {
					case frame := <-frames:
						data, ok := frame["d"].(map[string]any)
						if !ok || frame["op"] != float64(gateway.OpVoiceStateUpdate) || data["connection_id"] != "old" || data["channel_id"] != nil {
							t.Errorf("unexpected leave frame: %v", frame)
						}
					case <-time.After(3 * time.Second):
						t.Fatal("server did not receive leave")
					}
				}
			})
		}
	}
}
