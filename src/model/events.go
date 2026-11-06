package model

type MessageDelete struct {
	ID        string  `json:"id"`
	ChannelID string  `json:"channel_id"`
	GuildID   string  `json:"guild_id,omitempty"`
	Content   *string `json:"content,omitempty"`
	AuthorID  string  `json:"author_id,omitempty"`
	Member    *Member `json:"member,omitempty"`
}

type MessageDeleteBulk struct {
	IDs       []string `json:"ids"`
	ChannelID string   `json:"channel_id"`
	GuildID   string   `json:"guild_id,omitempty"`
}

type ReactionEvent struct {
	UserID    string  `json:"user_id"`
	ChannelID string  `json:"channel_id"`
	MessageID string  `json:"message_id"`
	GuildID   string  `json:"guild_id,omitempty"`
	Emoji     Emoji   `json:"emoji"`
	Member    *Member `json:"member,omitempty"`
}

type TypingStart struct {
	ChannelID string  `json:"channel_id"`
	UserID    string  `json:"user_id"`
	Timestamp int64   `json:"timestamp"`
	GuildID   string  `json:"guild_id,omitempty"`
	Member    *Member `json:"member,omitempty"`
}

type GuildMemberRemove struct {
	GuildID string `json:"guild_id"`
	User    User   `json:"user"`
}

type RoleEvent struct {
	GuildID string `json:"guild_id"`
	Role    Role   `json:"role"`
}

type RoleDelete struct {
	GuildID string `json:"guild_id"`
	RoleID  string `json:"role_id"`
}
