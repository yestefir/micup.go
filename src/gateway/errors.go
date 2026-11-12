package gateway

import (
	"errors"
	"fmt"
)

var (
	ErrReconnectRequested = errors.New("Micup Gateway requested reconnect")
	ErrClosed             = errors.New("Micup Gateway is disconnected")
)

const (
	CloseUnknownError         = 4000
	CloseUnknownOpcode        = 4001
	CloseDecodeError          = 4002
	CloseNotAuthenticated     = 4003
	CloseAuthenticationFailed = 4004
	CloseAlreadyAuthenticated = 4005
	CloseInvalidSequence      = 4007
	CloseRateLimited          = 4008
	CloseSessionTimeout       = 4009
	CloseInvalidShard         = 4010
	CloseShardingRequired     = 4011
	CloseInvalidAPIVersion    = 4012
)

type CloseError struct {
	Code int
	Text string
}

func (e *CloseError) Error() string {
	if e.Text != "" {
		return fmt.Sprintf("Micup Gateway closed the connection (code %d: %s)", e.Code, e.Text)
	}
	return fmt.Sprintf("Micup Gateway closed the connection (code %d)", e.Code)
}

func (e *CloseError) Fatal() bool {
	switch e.Code {
	case CloseUnknownOpcode, CloseDecodeError, CloseNotAuthenticated, CloseAuthenticationFailed, CloseAlreadyAuthenticated, CloseInvalidShard, CloseShardingRequired, CloseInvalidAPIVersion:
		return true
	}
	return false
}

func (e *CloseError) ResetsSession() bool {
	return e.Code == CloseInvalidSequence || e.Code == CloseSessionTimeout
}

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
