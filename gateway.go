package micup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

type Event struct {
	Name     string
	Sequence *int64
	Data     json.RawMessage
	session  *gatewaySession
}

type EventHandler func(context.Context, Event) error

func (e Event) Send(ctx context.Context, opcode int, data any) error {
	if e.session == nil {
		return errors.New("event is not attached to a Gateway session")
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return e.session.send(opcode, data)
}

type gatewayPacket struct {
	Op       int             `json:"op"`
	Data     json.RawMessage `json:"d"`
	Sequence *int64          `json:"s"`
	Name     string          `json:"t"`
}

type gatewaySession struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (s *gatewaySession) send(op int, data any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = s.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	return s.conn.WriteJSON(map[string]any{"op": op, "d": data})
}

func runGateway(parent context.Context, client *Client, handler EventHandler) error {
	if handler == nil {
		return errors.New("Gateway event handler is required")
	}
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	dialer := *websocket.DefaultDialer
	conn, _, err := dialer.DialContext(ctx, client.gatewayURL, http.Header{"User-Agent": []string{"micup.go/0.1"}})
	if err != nil {
		return errors.New("Micup Gateway connection failed")
	}
	conn.SetReadLimit(8 << 20)
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()
	defer conn.Close()
	session := &gatewaySession{conn: conn}
	var sequence *int64
	var sequenceMu sync.RWMutex
	var acknowledged atomic.Bool
	acknowledged.Store(true)
	var heartbeatCancel context.CancelFunc
	defer func() {
		if heartbeatCancel != nil {
			heartbeatCancel()
		}
	}()
	identified := false
	for {
		var packet gatewayPacket
		if err = conn.ReadJSON(&packet); err != nil {
			return fmt.Errorf("Micup Gateway read failed: %w", err)
		}
		if packet.Sequence != nil {
			n := *packet.Sequence
			sequenceMu.Lock()
			sequence = &n
			sequenceMu.Unlock()
		}
		switch packet.Op {
		case 10:
			if identified {
				return errors.New("duplicate Micup Gateway hello")
			}
			var hello struct {
				Interval float64 `json:"heartbeat_interval"`
			}
			if json.Unmarshal(packet.Data, &hello) != nil || hello.Interval < 100 {
				return errors.New("invalid Micup heartbeat interval")
			}
			if err = session.send(2, map[string]any{"token": client.config.Token, "properties": map[string]string{"os": "go", "browser": "micup.go", "device": "micup.go"}, "shard": []int{client.config.ShardID, client.config.ShardCount}}); err != nil {
				return err
			}
			heartbeatCtx, stopHeartbeat := context.WithCancel(ctx)
			heartbeatCancel = stopHeartbeat
			go func() {
				timer := time.NewTicker(time.Duration(hello.Interval) * time.Millisecond)
				defer timer.Stop()
				for {
					select {
					case <-heartbeatCtx.Done():
						return
					case <-timer.C:
						if !acknowledged.Swap(false) {
							cancel()
							return
						}
						sequenceMu.RLock()
						seq := sequence
						sequenceMu.RUnlock()
						if session.send(1, seq) != nil {
							cancel()
							return
						}
					}
				}
			}()
			identified = true
		case 11:
			acknowledged.Store(true)
		case 1:
			sequenceMu.RLock()
			seq := sequence
			sequenceMu.RUnlock()
			if err = session.send(1, seq); err != nil {
				return err
			}
		case 7:
			return errors.New("Micup Gateway requested reconnect")
		case 9:
			return errors.New("Micup Gateway invalidated the session")
		case 0:
			if err = handler(ctx, Event{Name: packet.Name, Sequence: sequence, Data: append(json.RawMessage(nil), packet.Data...), session: session}); err != nil {
				return err
			}
		}
	}
}
