package model

type VoiceState struct {
	GuildID          string   `json:"guild_id"`
	UserID           string   `json:"user_id"`
	ChannelID        string   `json:"channel_id"`
	SessionID        string   `json:"session_id,omitempty"`
	ConnectionID     string   `json:"connection_id,omitempty"`
	Member           *Member  `json:"member,omitempty"`
	Deaf             bool     `json:"deaf,omitempty"`
	Mute             bool     `json:"mute,omitempty"`
	SelfDeaf         bool     `json:"self_deaf,omitempty"`
	SelfMute         bool     `json:"self_mute,omitempty"`
	SelfStream       bool     `json:"self_stream,omitempty"`
	SelfVideo        bool     `json:"self_video,omitempty"`
	Suppress         bool     `json:"suppress,omitempty"`
	IsMobile         bool     `json:"is_mobile,omitempty"`
	ViewerStreamKeys []string `json:"viewer_stream_keys,omitempty"`
	E2EECapable      bool     `json:"e2ee_capable,omitempty"`
	P2P              bool     `json:"p2p,omitempty"`
	Version          int64    `json:"version,omitempty"`
}

type VoiceServerUpdate struct {
	GuildID      string `json:"guild_id"`
	ChannelID    string `json:"channel_id,omitempty"`
	Endpoint     string `json:"endpoint"`
	Token        string `json:"token"`
	ConnectionID string `json:"connection_id"`
	P2P          bool   `json:"p2p,omitempty"`
	E2EEKey      string `json:"e2ee_key,omitempty"`
}
