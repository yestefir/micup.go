package gateway

const (
	OpDispatch         = 0
	OpHeartbeat        = 1
	OpIdentify         = 2
	OpVoiceStateUpdate = 4
	OpResume           = 6
	OpReconnect        = 7
	OpInvalidSession   = 9
	OpHello            = 10
	OpHeartbeatAck     = 11
)

const (
	ProtocolVersion = "1"
	Encoding        = "json"
)
