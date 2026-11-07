package model

const (
	ComponentTypeStringSelect      = 3
	ComponentTypeUserSelect        = 5
	ComponentTypeRoleSelect        = 6
	ComponentTypeMentionableSelect = 7
	ComponentTypeChannelSelect     = 8
)

const (
	SelectDefaultUser    = "user"
	SelectDefaultRole    = "role"
	SelectDefaultChannel = "channel"
)

const (
	MaxSelectOptions           = 25
	MaxSelectPlaceholder       = 150
	MaxSelectOptionLabel       = 100
	MaxSelectOptionValue       = 100
	MaxSelectOptionDescription = 100
)

type SelectMenu struct {
	Type          int             `json:"type"`
	CustomID      string          `json:"custom_id"`
	Options       []SelectOption  `json:"options,omitempty"`
	ChannelTypes  []int           `json:"channel_types,omitempty"`
	Placeholder   string          `json:"placeholder,omitempty"`
	DefaultValues []SelectDefault `json:"default_values,omitempty"`
	MinValues     *int            `json:"min_values,omitempty"`
	MaxValues     *int            `json:"max_values,omitempty"`
	Disabled      bool            `json:"disabled,omitempty"`
}

type SelectOption struct {
	Label       string `json:"label"`
	Value       string `json:"value"`
	Description string `json:"description,omitempty"`
	Emoji       *Emoji `json:"emoji,omitempty"`
	Default     bool   `json:"default,omitempty"`
}

type SelectDefault struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

func IsSelectType(componentType int) bool {
	switch componentType {
	case ComponentTypeStringSelect, ComponentTypeUserSelect, ComponentTypeRoleSelect, ComponentTypeMentionableSelect, ComponentTypeChannelSelect:
		return true
	}
	return false
}

func NewSelectRow(menu SelectMenu) ActionRow {
	return ActionRow{Type: ComponentTypeActionRow, Select: &menu}
}

func StringSelect(customID string, options ...SelectOption) SelectMenu {
	return SelectMenu{Type: ComponentTypeStringSelect, CustomID: customID, Options: options}
}

func UserSelect(customID string) SelectMenu {
	return SelectMenu{Type: ComponentTypeUserSelect, CustomID: customID}
}

func RoleSelect(customID string) SelectMenu {
	return SelectMenu{Type: ComponentTypeRoleSelect, CustomID: customID}
}

func MentionableSelect(customID string) SelectMenu {
	return SelectMenu{Type: ComponentTypeMentionableSelect, CustomID: customID}
}

func ChannelSelect(customID string, channelTypes ...int) SelectMenu {
	return SelectMenu{Type: ComponentTypeChannelSelect, CustomID: customID, ChannelTypes: channelTypes}
}

func NewSelectOption(label, value string) SelectOption {
	return SelectOption{Label: label, Value: value}
}

func (o SelectOption) WithDescription(description string) SelectOption {
	o.Description = description
	return o
}

func (o SelectOption) WithEmoji(emoji Emoji) SelectOption {
	o.Emoji = &emoji
	return o
}

func (o SelectOption) WithDefault(selected bool) SelectOption {
	o.Default = selected
	return o
}

func (m SelectMenu) WithPlaceholder(placeholder string) SelectMenu {
	m.Placeholder = placeholder
	return m
}

func (m SelectMenu) WithValues(min, max int) SelectMenu {
	m.MinValues, m.MaxValues = &min, &max
	return m
}

func (m SelectMenu) WithDisabled(disabled bool) SelectMenu {
	m.Disabled = disabled
	return m
}

func (m SelectMenu) WithDefaults(defaults ...SelectDefault) SelectMenu {
	m.DefaultValues = defaults
	return m
}

func DefaultUser(id string) SelectDefault { return SelectDefault{ID: id, Type: SelectDefaultUser} }

func DefaultRole(id string) SelectDefault { return SelectDefault{ID: id, Type: SelectDefaultRole} }

func DefaultChannel(id string) SelectDefault {
	return SelectDefault{ID: id, Type: SelectDefaultChannel}
}

func (i Interaction) IsSelectMenu() bool {
	data, ok := i.ComponentData()
	return ok && IsSelectType(data.ComponentType)
}

func (i Interaction) SelectedValues() []string {
	data, _ := i.ComponentData()
	return data.Values
}

func (d ComponentData) Users() []User {
	if d.Resolved == nil {
		return nil
	}
	var users []User
	for _, id := range d.Values {
		if user, ok := d.Resolved.Users[id]; ok {
			users = append(users, user)
		}
	}
	return users
}

func (d ComponentData) Members() []Member {
	if d.Resolved == nil {
		return nil
	}
	var members []Member
	for _, id := range d.Values {
		member, ok := d.Resolved.Members[id]
		if !ok {
			continue
		}
		if member.User == nil {
			if user, found := d.Resolved.Users[id]; found {
				member.User = &user
			}
		}
		members = append(members, member)
	}
	return members
}

func (d ComponentData) Roles() []Role {
	if d.Resolved == nil {
		return nil
	}
	var roles []Role
	for _, id := range d.Values {
		if role, ok := d.Resolved.Roles[id]; ok {
			roles = append(roles, role)
		}
	}
	return roles
}

func (d ComponentData) Channels() []Channel {
	if d.Resolved == nil {
		return nil
	}
	var channels []Channel
	for _, id := range d.Values {
		if channel, ok := d.Resolved.Channels[id]; ok {
			channels = append(channels, channel)
		}
	}
	return channels
}
