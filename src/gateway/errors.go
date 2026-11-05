package gateway

import (
	"errors"
	"fmt"
)

var (
	ErrReconnectRequested = errors.New("Micup Gateway requested reconnect")
	ErrSessionInvalidated = errors.New("Micup Gateway invalidated the session")
	ErrClosed             = errors.New("Micup Gateway is disconnected")
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

func (e *CloseError) ResetsSession() bool { return e.Code == 4007 || e.Code == 4009 }
