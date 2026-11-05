package gateway

import (
	"errors"
	"fmt"
)

var (
	ErrReconnectRequested = errors.New("Micup Gateway requested reconnect")
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
	case 4001, 4002, 4003, 4004, 4005, 4010, 4011, 4012:
		return true
	}
	return false
}

func (e *CloseError) ResetsSession() bool { return e.Code == 4007 || e.Code == 4009 }

const (
	VoiceChannelFull        = "VOICE_CHANNEL_FULL"
	VoiceP2PConsentRequired = "VOICE_P2P_CONSENT_REQUIRED"
	VoiceP2PUnavailable     = "VOICE_P2P_UNAVAILABLE"
)

type GatewayError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e GatewayError) Error() string {
	return fmt.Sprintf("Micup Gateway refused the request (%s)", e.Code)
}
