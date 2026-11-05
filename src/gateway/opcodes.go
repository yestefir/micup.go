package gateway

const (
	OpDispatch                   = 0
	OpHeartbeat                  = 1
	OpIdentify                   = 2
	OpPresenceUpdate             = 3
	OpVoiceStateUpdate           = 4
	OpResume                     = 6
	OpReconnect                  = 7
	OpRequestGuildMembers        = 8
	OpInvalidSession             = 9
	OpHello                      = 10
	OpHeartbeatAck               = 11
	OpGatewayError               = 12
	OpLazyRequest                = 14
	OpRequestGuildCounts         = 15
	OpRequestChannelMemberCounts = 16
)

const (
	ProtocolVersion = "1"
	Encoding        = "json"
)
