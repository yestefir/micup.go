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

type MemberListGroup struct {
	ID    string `json:"id"`
	Count int    `json:"count"`
}

type MemberListMember struct {
	model.Member
	Presence *PresenceUpdate `json:"presence,omitempty"`
}

type MemberListItem struct {
	Group  *MemberListGroup  `json:"group,omitempty"`
	Member *MemberListMember `json:"member,omitempty"`
}

type MemberListOperation struct {
	Op    string           `json:"op"`
	Range []int            `json:"range,omitempty"`
	Items []MemberListItem `json:"items,omitempty"`
}

type MemberListUpdate struct {
	GuildID     string                `json:"guild_id"`
	ID          string                `json:"id"`
	ChannelID   string                `json:"channel_id,omitempty"`
	MemberCount int                   `json:"member_count"`
	OnlineCount int                   `json:"online_count"`
	Groups      []MemberListGroup     `json:"groups,omitempty"`
	Ops         []MemberListOperation `json:"ops,omitempty"`
}
