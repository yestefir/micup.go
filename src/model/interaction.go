package model

import "encoding/json"

type Interaction struct {
	ID            string          `json:"id"`
	ApplicationID string          `json:"application_id"`
	Type          int             `json:"type"`
	Data          json.RawMessage `json:"data,omitempty"`
	GuildID       string          `json:"guild_id,omitempty"`
	ChannelID     string          `json:"channel_id,omitempty"`
	Token         string          `json:"token"`
	Version       int             `json:"version,omitempty"`
	Member        json.RawMessage `json:"member,omitempty"`
	User          json.RawMessage `json:"user,omitempty"`
}
