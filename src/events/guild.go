package events

import (
	"encoding/json"

	"github.com/yestefir/micup.go/src/model"
)

type GuildCreate struct {
	model.Guild
	Unavailable bool               `json:"unavailable,omitempty"`
	JoinedAt    string             `json:"joined_at,omitempty"`
	Channels    []model.Channel    `json:"channels,omitempty"`
	Threads     []model.Channel    `json:"threads,omitempty"`
	Members     []model.Member     `json:"members,omitempty"`
	VoiceStates []model.VoiceState `json:"voice_states,omitempty"`
	Presences   []PresenceUpdate   `json:"presences,omitempty"`
	Properties  *model.Guild       `json:"properties,omitempty"`
}

func (g *GuildCreate) UnmarshalJSON(data []byte) error {
	type plain GuildCreate
	err := json.Unmarshal(data, (*plain)(g))
	g.Guild = withProperties(g.Guild, g.Properties)
	return err
}

type GuildUpdate struct {
	model.Guild
	GuildID    string       `json:"guild_id,omitempty"`
	Properties *model.Guild `json:"properties,omitempty"`
}

func (g *GuildUpdate) UnmarshalJSON(data []byte) error {
	type plain GuildUpdate
	err := json.Unmarshal(data, (*plain)(g))
	if g.ID == "" {
		g.ID = g.GuildID
	}
	g.Guild = withProperties(g.Guild, g.Properties)
	return err
}

type GuildDelete struct {
	ID          string `json:"id"`
	Unavailable bool   `json:"unavailable,omitempty"`
}

func withProperties(guild model.Guild, properties *model.Guild) model.Guild {
	if properties == nil {
		return guild
	}
	if properties.OwnerID != "" {
		guild.OwnerID = properties.OwnerID
	}
	if guild.Name == "" {
		guild.Name = properties.Name
	}
	return guild
}
