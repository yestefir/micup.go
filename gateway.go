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

const (
	OpDispatch            = 0
	OpHeartbeat           = 1
	OpIdentify            = 2
	OpVoiceStateUpdate    = 4
	OpResume              = 6
	OpReconnect           = 7
	OpInvalidSession      = 9
	OpHello               = 10
	OpHeartbeatAck        = 11
	gatewayHelloTimeout   = 10 * time.Second
	gatewayWriteTimeout   = 5 * time.Second
	gatewayReadLimitBytes = 8 << 20
)

var (
	ErrReconnectRequested = errors.New("Micup Gateway requested reconnect")
	ErrSessionInvalidated = errors.New("Micup Gateway invalidated the session")
	ErrGatewayClosed      = errors.New("Micup Gateway is disconnected")
)

type CloseError struct {
	Code int
	Text string
}

func (e *CloseError) Error() string {
	return fmt.Sprintf("Micup Gateway closed the connection (code %d)", e.Code)
}

func (e *CloseError) Fatal() bool {
	switch e.Code {
	case 4001, 4002, 4003, 4004, 4005, 4010, 4011, 4012, 4013, 4014:
		return true
	}
	return false
}

func (e *CloseError) resetsSession() bool { return e.Code == 4007 || e.Code == 4009 }

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
	_ = s.conn.SetWriteDeadline(time.Now().Add(gatewayWriteTimeout))
	return s.conn.WriteJSON(map[string]any{"op": op, "d": data})
}

type resumeState struct {
	mu        sync.Mutex
	sessionID string
	sequence  *int64
}

func (r *resumeState) set(sequence *int64) {
	if sequence == nil {
		return
	}
	n := *sequence
	r.mu.Lock()
	r.sequence = &n
	r.mu.Unlock()
}

func (r *resumeState) last() *int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.sequence
}

func (r *resumeState) snapshot() (string, *int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.sessionID, r.sequence
}

func (r *resumeState) start(sessionID string) {
	r.mu.Lock()
	r.sessionID = sessionID
	r.mu.Unlock()
}

func (r *resumeState) reset() {
	r.mu.Lock()
	r.sessionID = ""
	r.sequence = nil
	r.mu.Unlock()
}

func (r *resumeState) resumable() bool {
	sessionID, sequence := r.snapshot()
	return sessionID != "" && sequence != nil
}

type connectionHooks struct {
	userAgent      string
	browser        string
	shardID        int
	shardCount     int
	attached       func(*gatewaySession)
	beforeIdentify func(context.Context) error
	resume         *resumeState
}

func runConnection(parent context.Context, client *Client, handler EventHandler, hooks connectionHooks) error {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	resume := hooks.resume
	if hooks.beforeIdentify != nil && !resume.resumable() {
		if err := hooks.beforeIdentify(ctx); err != nil {
			return err
		}
	}
	dialer := *websocket.DefaultDialer
	conn, _, err := dialer.DialContext(ctx, client.gatewayURL, http.Header{"User-Agent": []string{hooks.userAgent}})
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return errors.New("Micup Gateway connection failed")
	}
	conn.SetReadLimit(gatewayReadLimitBytes)
	_ = conn.SetReadDeadline(time.Now().Add(gatewayHelloTimeout))
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()
	defer conn.Close()
	session := &gatewaySession{conn: conn}
	if hooks.attached != nil {
		hooks.attached(session)
		defer hooks.attached(nil)
	}
	var acknowledged atomic.Bool
	acknowledged.Store(true)
	identified := false
	for {
		var packet gatewayPacket
		if err = conn.ReadJSON(&packet); err != nil {
			if parent.Err() != nil {
				return parent.Err()
			}
			var closed *websocket.CloseError
			if errors.As(err, &closed) {
				result := &CloseError{Code: closed.Code, Text: closed.Text}
				if result.resetsSession() {
					resume.reset()
				}
				return result
			}
			return fmt.Errorf("Micup Gateway read failed: %w", err)
		}
		resume.set(packet.Sequence)
		switch packet.Op {
		case OpHello:
			if identified {
				return errors.New("duplicate Micup Gateway hello")
			}
			var hello struct {
				Interval float64 `json:"heartbeat_interval"`
			}
			if json.Unmarshal(packet.Data, &hello) != nil || hello.Interval < 100 {
				return errors.New("invalid Micup heartbeat interval")
			}
			_ = conn.SetReadDeadline(time.Time{})
			if sessionID, sequence := resume.snapshot(); sessionID != "" && sequence != nil {
				err = session.send(OpResume, map[string]any{"token": client.config.Token, "session_id": sessionID, "seq": *sequence})
			} else {
				err = session.send(OpIdentify, map[string]any{"token": client.config.Token, "properties": map[string]string{"os": "go", "browser": hooks.browser, "device": hooks.browser}, "shard": []int{hooks.shardID, hooks.shardCount}})
			}
			if err != nil {
				return err
			}
			go heartbeat(ctx, cancel, session, resume, &acknowledged, time.Duration(hello.Interval)*time.Millisecond)
			identified = true
		case OpHeartbeatAck:
			acknowledged.Store(true)
		case OpHeartbeat:
			if err = session.send(OpHeartbeat, resume.last()); err != nil {
				return err
			}
		case OpReconnect:
			return ErrReconnectRequested
		case OpInvalidSession:
			resume.reset()
			return ErrSessionInvalidated
		case OpDispatch:
			if packet.Name == EventReady {
				var ready struct {
					SessionID string `json:"session_id"`
				}
				_ = json.Unmarshal(packet.Data, &ready)
				resume.start(ready.SessionID)
			}
			if err = handler(ctx, Event{Name: packet.Name, Sequence: resume.last(), ShardID: hooks.shardID, Data: append(json.RawMessage(nil), packet.Data...), session: session}); err != nil {
				return err
			}
		}
	}
}

func heartbeat(ctx context.Context, cancel context.CancelFunc, session *gatewaySession, resume *resumeState, acknowledged *atomic.Bool, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !acknowledged.Swap(false) {
				cancel()
				return
			}
			if session.send(OpHeartbeat, resume.last()) != nil {
				cancel()
				return
			}
		}
	}
}
