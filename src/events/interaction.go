package events

type ApplicationCommandsUpdate struct {
	ApplicationID string `json:"application_id"`
	GuildID       string `json:"guild_id,omitempty"`
	ScopeID       string `json:"scope_id,omitempty"`
	Version       string `json:"version"`
}
