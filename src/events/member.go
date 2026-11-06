package events

import "github.com/yestefir/micup.go/src/model"

type GuildMember struct {
	model.Member
	GuildID string `json:"guild_id"`
}

type GuildMemberRemove struct {
	GuildID string     `json:"guild_id"`
	User    model.User `json:"user"`
}

type MembersChunk struct {
	GuildID    string           `json:"guild_id"`
	Members    []model.Member   `json:"members"`
	ChunkIndex int              `json:"chunk_index"`
	ChunkCount int              `json:"chunk_count"`
	NotFound   []string         `json:"not_found,omitempty"`
	Presences  []PresenceUpdate `json:"presences,omitempty"`
	Nonce      string           `json:"nonce,omitempty"`
}

type GuildBan struct {
	GuildID string     `json:"guild_id"`
	User    model.User `json:"user"`
}
