package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const writeTimeout = 5 * time.Second

type Conn struct {
	ws        *websocket.Conn
	mu        sync.Mutex
	commands  *windowLimiter
	presences *windowLimiter
	zstd      *zstdWriter
	// heartbeatAt is when the last heartbeat was written, in Unix nanoseconds;
	// latency is the round trip to its acknowledgement.
	heartbeatAt atomic.Int64
	latency     atomic.Int64
}

// Latency is the round trip of the last acknowledged heartbeat, or 0 before
// the first acknowledgement on this connection.
func (c *Conn) Latency() time.Duration { return time.Duration(c.latency.Load()) }

func (c *Conn) acknowledged() {
	if sent := c.heartbeatAt.Swap(0); sent != 0 {
		c.latency.Store(max(time.Now().UnixNano()-sent, 1))
	}
}

func newConn(ws *websocket.Conn) *Conn {
	return &Conn{ws: ws, commands: newWindowLimiter(commandBudget, commandWindow), presences: newWindowLimiter(presenceBudget, presenceWindow)}
}

func (c *Conn) Send(ctx context.Context, opcode int, data any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	encoded, err := json.Marshal(map[string]any{"op": opcode, "d": data})
	if err != nil {
		return err
	}
	if len(encoded) > maxPayloadSize {
		return fmt.Errorf("Micup Gateway payload is %d bytes, over the %d byte limit", len(encoded), maxPayloadSize)
	}
	if opcode != OpHeartbeat && opcode != OpIdentify && opcode != OpResume {
		if err = c.commands.wait(ctx); err != nil {
			return err
		}
	}
	if opcode == OpPresenceUpdate {
		if err = c.presences.wait(ctx); err != nil {
			return err
		}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	kind := websocket.TextMessage
	if c.zstd != nil {
		if encoded, err = c.zstd.frame(encoded); err != nil {
			return err
		}
		kind = websocket.BinaryMessage
	}
	_ = c.ws.SetWriteDeadline(time.Now().Add(writeTimeout))
	if opcode == OpHeartbeat {
		c.heartbeatAt.Store(time.Now().UnixNano())
	}
	return c.ws.WriteMessage(kind, encoded)
}
