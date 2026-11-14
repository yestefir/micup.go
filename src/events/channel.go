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

type ChannelRecipient struct {
	ChannelID string     `json:"channel_id"`
	User      model.User `json:"user"`
}

type ThreadCreate struct {
	model.Channel
	NewlyCreated bool `json:"newly_created,omitempty"`
}

type ThreadDelete struct {
	ID       string `json:"id"`
	GuildID  string `json:"guild_id"`
	ParentID string `json:"parent_id,omitempty"`
	Type     int    `json:"type"`
}

type ThreadListSync struct {
	GuildID    string               `json:"guild_id"`
	ChannelIDs []string             `json:"channel_ids,omitempty"`
	Threads    []model.Channel      `json:"threads"`
	Members    []model.ThreadMember `json:"members,omitempty"`
}

type ThreadMemberPresence struct {
	model.ThreadMember
	Presence *PresenceUpdate `json:"presence,omitempty"`
}

type ThreadMembersUpdate struct {
	ID               string                 `json:"id"`
	GuildID          string                 `json:"guild_id"`
	MemberCount      int                    `json:"member_count"`
	AddedMembers     []ThreadMemberPresence `json:"added_members,omitempty"`
	RemovedMemberIDs []string               `json:"removed_member_ids,omitempty"`
}

type ThreadMemberListUpdate struct {
	GuildID  string                 `json:"guild_id"`
	ThreadID string                 `json:"thread_id"`
	Members  []ThreadMemberPresence `json:"members"`
}
