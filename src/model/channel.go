package model

type Channel struct {
	ID      string `json:"id"`
	GuildID string `json:"guild_id"`
	Name    string `json:"name"`
	Type    int    `json:"type"`
}
