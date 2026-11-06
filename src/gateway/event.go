package gateway

import (
	"context"
	"encoding/json"
	"errors"
)

type Event struct {
	Name     string
	Sequence *int64
	ShardID  int
	Data     json.RawMessage
	conn     *Conn
}

type Handler func(context.Context, Event) error

func (e Event) Decode(dest any) error { return json.Unmarshal(e.Data, dest) }

func (e Event) Send(ctx context.Context, opcode int, data any) error {
	if e.conn == nil {
		return errors.New("event is not attached to a Gateway connection")
	}
	return e.conn.Send(ctx, opcode, data)
}
