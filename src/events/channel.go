package events

import "github.com/yestefir/micup.go/src/model"

type ChannelUpdateBulk struct {
	GuildID  string          `json:"guild_id"`
	Channels []model.Channel `json:"channels"`
}

type ChannelPinsUpdate struct {
	GuildID          string  `json:"guild_id,omitempty"`
	ChannelID        string  `json:"channel_id"`
	LastPinTimestamp *string `json:"last_pin_timestamp,omitempty"`
}
