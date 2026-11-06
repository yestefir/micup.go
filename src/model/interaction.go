package model

import "encoding/json"

const (
	InteractionTypePing               = 1
	InteractionTypeApplicationCommand = 2
	InteractionTypeMessageComponent   = 3
	InteractionTypeAutocomplete       = 4
	InteractionTypeModalSubmit        = 5
)

const (
	InteractionContextGuild          = 0
	InteractionContextBotDM          = 1
	InteractionContextPrivateChannel = 2
)

type Interaction struct {
	ID                           string            `json:"id"`
	ApplicationID                string            `json:"application_id"`
	Type                         int               `json:"type"`
	Data                         json.RawMessage   `json:"data,omitempty"`
	GuildID                      string            `json:"guild_id,omitempty"`
	ChannelID                    string            `json:"channel_id,omitempty"`
	Token                        string            `json:"token"`
	Version                      int               `json:"version,omitempty"`
	Member                       *Member           `json:"member,omitempty"`
	User                         *User             `json:"user,omitempty"`
	Message                      *Message          `json:"message,omitempty"`
	AppPermissions               *Permissions      `json:"app_permissions,omitempty"`
	Locale                       string            `json:"locale,omitempty"`
	GuildLocale                  string            `json:"guild_locale,omitempty"`
	Channel                      *Channel          `json:"channel,omitempty"`
	Guild                        *InteractionGuild `json:"guild,omitempty"`
	Context                      *int              `json:"context,omitempty"`
	AttachmentSizeLimit          int64             `json:"attachment_size_limit,omitempty"`
	AuthorizingIntegrationOwners map[string]string `json:"authorizing_integration_owners,omitempty"`
}

type InteractionGuild struct {
	ID       string   `json:"id"`
	Locale   string   `json:"locale,omitempty"`
	Features []string `json:"features,omitempty"`
}

type ComponentData struct {
	CustomID      string `json:"custom_id"`
	ComponentType int    `json:"component_type"`
}

func (i Interaction) ComponentData() (ComponentData, bool) {
	var data ComponentData
	if i.Type != InteractionTypeMessageComponent || json.Unmarshal(i.Data, &data) != nil {
		return ComponentData{}, false
	}
	return data, true
}

func (i Interaction) IsButton() bool {
	data, ok := i.ComponentData()
	return ok && data.ComponentType == ComponentTypeButton
}

func (i Interaction) CustomID() string {
	data, _ := i.ComponentData()
	return data.CustomID
}

func (i Interaction) Invoker() *User {
	if i.Member != nil && i.Member.User != nil {
		return i.Member.User
	}
	return i.User
}
