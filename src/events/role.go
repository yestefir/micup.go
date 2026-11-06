package events

import "github.com/yestefir/micup.go/src/model"

type RoleEvent struct {
	GuildID string     `json:"guild_id"`
	Role    model.Role `json:"role"`
}

type RoleUpdateBulk struct {
	GuildID string       `json:"guild_id"`
	Roles   []model.Role `json:"roles"`
}

type RoleDelete struct {
	GuildID string `json:"guild_id"`
	RoleID  string `json:"role_id"`
}
