package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// Stall one application write after the WebSocket handshake, without relying
// on operating-system socket buffer sizes or changing production timeouts.
type stalledWriteConn struct {
	net.Conn
	stall   atomic.Bool
	started chan struct{}
	release chan struct{}
}

func (c *stalledWriteConn) Write(p []byte) (int, error) {
	if c.stall.CompareAndSwap(true, false) {
		close(c.started)
		<-c.release
	}
	return c.Conn.Write(p)
}

func TestConnQueuedSendEndsWithContextWhileWriterIsBlocked(t *testing.T) {
	for _, deadline := range []bool{false, true} {
		for _, opcode := range []int{OpVoiceStateUpdate, OpHeartbeat} {
			name := "cancel/command"
			if deadline {
				name = "deadline/command"
			}
			if opcode == OpHeartbeat {
				name = strings.TrimSuffix(name, "command") + "heartbeat"
			}
			t.Run(name, func(t *testing.T) {
				frames := make(chan int, 3)
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					ws, err := (&websocket.Upgrader{}).Upgrade(w, r, nil)
					if err != nil {
						return
					}
					defer ws.Close()
					_ = ws.SetReadDeadline(time.Now().Add(5 * time.Second))
					for {
						var p packet
						if ws.ReadJSON(&p) != nil {
							return
						}
						frames <- p.Op
					}
				}))
				defer server.Close()
				var transport *stalledWriteConn
				dialer := websocket.Dialer{NetDialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					conn, err := (&net.Dialer{}).DialContext(ctx, network, addr)
					if err != nil {
						return nil, err
					}
					transport = &stalledWriteConn{Conn: conn, started: make(chan struct{}), release: make(chan struct{})}
					return transport, nil
				}}
				ws, _, err := dialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), nil)
				if err != nil {
					t.Fatal(err)
				}
				defer ws.Close()
				c := newConn(ws)
				transport.stall.Store(true)
				first := make(chan error, 1)
				func() {
					defer close(transport.release)
					go func() { first <- c.Send(context.Background(), OpIdentify, nil) }()
					select {
					case <-transport.started:
					case <-time.After(3 * time.Second):
						t.Fatal("first write did not stall")
					}
					ctx, cancel := context.WithCancel(context.Background())
					want := context.Canceled
					if deadline {
						cancel()
						ctx, cancel = context.WithTimeout(context.Background(), 100*time.Millisecond)
						want = context.DeadlineExceeded
					}
					defer cancel()
					encoded := make(chan struct{})
					done := make(chan error, 1)
					go func() { done <- c.Send(ctx, opcode, queuedPayload{encoded: encoded, data: json.RawMessage(`null`)}) }()
					select {
					case <-encoded:
					case <-time.After(3 * time.Second):
						t.Fatal("queued send did not serialize")
					}
					if !deadline {
						cancel()
					}
					<-ctx.Done()
					select {
					case err := <-done:
						if !errors.Is(err, want) {
							t.Errorf("queued send returned %v, want %v", err, want)
						}
					case <-time.After(500 * time.Millisecond):
						t.Error("queued send still waits for the blocked writer after its context ended")
					}
				}()
				if err := <-first; err != nil {
					t.Fatalf("cancelling a queued send interrupted the active writer: %v", err)
				}
				if c.heartbeatAt.Load() != 0 {
					t.Error("cancelled send changed heartbeat state")
				}
				if err := c.Send(context.Background(), OpResume, nil); err != nil {
					t.Fatal(err)
				}
				for _, want := range []int{OpIdentify, OpResume} {
					select {
					case got := <-frames:
						if got != want {
							t.Errorf("received opcode %d, want %d; cancelled send reached the wire", got, want)
						}
					case <-time.After(3 * time.Second):
						t.Fatal("active frame was not received")
					}
				}
			})
		}
	}
}

var benchmarkConn *Conn

func BenchmarkConnConstruction(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		benchmarkConn = newConn(nil)
	}
}
