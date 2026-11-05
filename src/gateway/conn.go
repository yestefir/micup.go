package gateway

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const writeTimeout = 5 * time.Second

type Conn struct {
	ws *websocket.Conn
	mu sync.Mutex
}

func (c *Conn) Send(opcode int, data any) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_ = c.ws.SetWriteDeadline(time.Now().Add(writeTimeout))
	return c.ws.WriteJSON(map[string]any{"op": opcode, "d": data})
}
