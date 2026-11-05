package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const (
	helloTimeout       = 10 * time.Second
	readLimitBytes     = 8 << 20
	minHeartbeatMillis = 100
)

type packet struct {
	Op       int             `json:"op"`
	Data     json.RawMessage `json:"d"`
	Sequence *int64          `json:"s"`
	Name     string          `json:"t"`
}

func Run(parent context.Context, options Options, handler Handler) error {
	if handler == nil {
		return errors.New("Gateway event handler is required")
	}
	options, err := options.normalize()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	resume := options.Resume
	if options.BeforeIdentify != nil && !resume.Resumable() {
		if err := options.BeforeIdentify(ctx); err != nil {
			return err
		}
	}
	ws, err := dial(ctx, options)
	if err != nil {
		return err
	}
	stop := context.AfterFunc(ctx, func() { _ = ws.Close() })
	defer stop()
	defer ws.Close()
	conn := newConn(ws)
	var reader packetReader = plainReader{ws: ws}
	if options.Compress {
		zr, err := newZstdReader(ws)
		if err != nil {
			return err
		}
		defer zr.close()
		reader = zr
		if conn.zstd, err = newZstdWriter(); err != nil {
			return err
		}
	}
	if options.OnConnect != nil {
		options.OnConnect(conn)
	}
	if options.OnDisconnect != nil {
		defer options.OnDisconnect()
	}
	var acknowledged atomic.Bool
	acknowledged.Store(true)
	identified := false
	for {
		var p packet
		if err = reader.next(&p); err != nil {
			return readError(parent, resume, err)
		}
		switch p.Op {
		case OpHello:
			if identified {
				return errors.New("duplicate Micup Gateway hello")
			}
			interval, err := helloInterval(p.Data)
			if err != nil {
				return err
			}
			_ = ws.SetReadDeadline(time.Time{})
			if err = identify(conn, options, resume); err != nil {
				return err
			}
			go heartbeat(ctx, cancel, conn, resume, &acknowledged, interval)
			identified = true
		case OpHeartbeatAck:
			acknowledged.Store(true)
			conn.acknowledged()
		case OpHeartbeat:
			if err = conn.Send(ctx, OpHeartbeat, resume.last()); err != nil {
				return err
			}
		case OpReconnect:
			return ErrReconnectRequested
		case OpInvalidSession:
			if err = reidentify(ctx, conn, options, resume); err != nil {
				return err
			}
		case OpGatewayError:
			if options.OnGatewayError != nil {
				var refusal GatewayError
				_ = json.Unmarshal(p.Data, &refusal)
				options.OnGatewayError(refusal)
			}
		case OpDispatch:
			if p.Name == EventReady {
				var ready struct {
					SessionID string `json:"session_id"`
				}
				_ = json.Unmarshal(p.Data, &ready)
				resume.start(ready.SessionID)
			}
			event := Event{Name: p.Name, Sequence: copySequence(p.Sequence), ShardID: options.ShardID, Data: append(json.RawMessage(nil), p.Data...), conn: conn}
			if err = handler(ctx, event); err != nil {
				return err
			}
			resume.observe(p.Sequence)
		}
	}
}

func dial(ctx context.Context, options Options) (*websocket.Conn, error) {
	dialer := *websocket.DefaultDialer
	ws, _, err := dialer.DialContext(ctx, options.URL, http.Header{"User-Agent": []string{options.UserAgent}})
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, errors.New("Micup Gateway connection failed")
	}
	ws.SetReadLimit(readLimitBytes)
	_ = ws.SetReadDeadline(time.Now().Add(helloTimeout))
	return ws, nil
}

func readError(parent context.Context, resume *Resume, err error) error {
	if parent.Err() != nil {
		return parent.Err()
	}
	var closed *websocket.CloseError
	if errors.As(err, &closed) {
		result := &CloseError{Code: closed.Code, Text: closed.Text}
		if result.ResetsSession() {
			resume.Reset()
		}
		return result
	}
	return fmt.Errorf("Micup Gateway read failed: %w", err)
}

func helloInterval(data json.RawMessage) (time.Duration, error) {
	var hello struct {
		Interval float64 `json:"heartbeat_interval"`
	}
	if json.Unmarshal(data, &hello) != nil || hello.Interval < minHeartbeatMillis {
		return 0, errors.New("invalid Micup heartbeat interval")
	}
	return time.Duration(hello.Interval) * time.Millisecond, nil
}

func reidentify(ctx context.Context, conn *Conn, options Options, resume *Resume) error {
	resume.Reset()
	if options.OnSessionInvalidated != nil {
		options.OnSessionInvalidated()
	}
	if options.BeforeIdentify != nil {
		if err := options.BeforeIdentify(ctx); err != nil {
			return err
		}
	}
	return identify(conn, options, resume)
}

func copySequence(sequence *int64) *int64 {
	if sequence == nil {
		return nil
	}
	n := *sequence
	return &n
}

func identify(conn *Conn, options Options, resume *Resume) error {
	if sessionID, sequence := resume.snapshot(); sessionID != "" && sequence != nil {
		return conn.Send(context.Background(), OpResume, map[string]any{"token": options.Token, "session_id": sessionID, "seq": *sequence})
	}
	payload := map[string]any{
		"token":      options.Token,
		"properties": map[string]string{"os": "go", "browser": options.Browser, "device": options.Browser},
		"shard":      []int{options.ShardID, options.ShardCount},
	}
	if len(options.IgnoredEvents) != 0 {
		payload["ignored_events"] = options.IgnoredEvents
	}
	if presence := options.presence(); presence != nil {
		payload["presence"] = presence
	}
	if options.Flags != 0 {
		payload["flags"] = options.Flags
	}
	return conn.Send(context.Background(), OpIdentify, payload)
}

func heartbeat(ctx context.Context, cancel context.CancelFunc, conn *Conn, resume *Resume, acknowledged *atomic.Bool, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !acknowledged.Swap(false) || conn.Send(ctx, OpHeartbeat, resume.last()) != nil {
				cancel()
				return
			}
		}
	}
}
