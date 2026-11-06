package model

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"unicode/utf8"
)

const (
	CommandTypeChatInput = 1
	CommandTypeUser      = 2
	CommandTypeMessage   = 3
)

const (
	OptionTypeSubcommand      = 1
	OptionTypeSubcommandGroup = 2
	OptionTypeString          = 3
	OptionTypeInteger         = 4
	OptionTypeBoolean         = 5
	OptionTypeUser            = 6
	OptionTypeChannel         = 7
	OptionTypeRole            = 8
	OptionTypeMentionable     = 9
	OptionTypeNumber          = 10
	OptionTypeAttachment      = 11
)

const (
	MaxCommandOptions     = 25
	MaxCommandChoices     = 25
	MaxCommandName        = 32
	MaxCommandDescription = 100
)

type ApplicationCommand struct {
	ID                       string            `json:"id,omitempty"`
	ApplicationID            string            `json:"application_id,omitempty"`
	GuildID                  string            `json:"guild_id,omitempty"`
	Version                  string            `json:"version,omitempty"`
	Type                     int               `json:"type,omitempty"`
	Name                     string            `json:"name"`
	NameLocalizations        map[string]string `json:"name_localizations,omitempty"`
	Description              string            `json:"description,omitempty"`
	DescriptionLocalizations map[string]string `json:"description_localizations,omitempty"`
	Options                  []CommandOption   `json:"options,omitempty"`
	DefaultMemberPermissions *Permissions      `json:"default_member_permissions,omitempty"`
	DMPermission             *bool             `json:"dm_permission,omitempty"`
}

type CommandOption struct {
	Type                     int               `json:"type"`
	Name                     string            `json:"name"`
	NameLocalizations        map[string]string `json:"name_localizations,omitempty"`
	Description              string            `json:"description"`
	DescriptionLocalizations map[string]string `json:"description_localizations,omitempty"`
	Required                 bool              `json:"required,omitempty"`
	Choices                  []CommandChoice   `json:"choices,omitempty"`
	Options                  []CommandOption   `json:"options,omitempty"`
	ChannelTypes             []int             `json:"channel_types,omitempty"`
	MinValue                 *float64          `json:"min_value,omitempty"`
	MaxValue                 *float64          `json:"max_value,omitempty"`
	MinLength                *int              `json:"min_length,omitempty"`
	MaxLength                *int              `json:"max_length,omitempty"`
	Autocomplete             bool              `json:"autocomplete,omitempty"`
}

type CommandChoice struct {
	Name              string            `json:"name"`
	NameLocalizations map[string]string `json:"name_localizations,omitempty"`
	Value             any               `json:"value"`
}

var chatInputName = regexp.MustCompile(`^[-_\p{Ll}\p{Lo}\p{N}]{1,32}$`)

func (c ApplicationCommand) Validate() error {
	kind := c.Type
	if kind == 0 {
		kind = CommandTypeChatInput
	}
	switch kind {
	case CommandTypeChatInput:
		if !chatInputName.MatchString(c.Name) {
			return fmt.Errorf("command name %q must be 1-32 lowercase letters, digits, - or _", c.Name)
		}
		if err := validDescription("command "+c.Name, c.Description); err != nil {
			return err
		}
		return validOptions(c.Name, c.Options, 0)
	case CommandTypeUser, CommandTypeMessage:
		if n := utf8.RuneCountInString(c.Name); n == 0 || n > MaxCommandName {
			return fmt.Errorf("context menu command name must be 1-%d characters", MaxCommandName)
		}
		if c.Description != "" || len(c.Options) != 0 {
			return fmt.Errorf("context menu command %q cannot have a description or options", c.Name)
		}
		return nil
	}
	return fmt.Errorf("unknown command type %d", c.Type)
}

func validDescription(what, description string) error {
	if n := utf8.RuneCountInString(description); n == 0 || n > MaxCommandDescription {
		return fmt.Errorf("%s needs a description of 1-%d characters", what, MaxCommandDescription)
	}
	return nil
}

func validOptions(parent string, options []CommandOption, depth int) error {
	if len(options) > MaxCommandOptions {
		return fmt.Errorf("%s has %d options, Micup allows %d", parent, len(options), MaxCommandOptions)
	}
	seen := map[string]bool{}
	optional := false
	for _, option := range options {
		where := parent + " " + option.Name
		if !chatInputName.MatchString(option.Name) {
			return fmt.Errorf("option name %q must be 1-32 lowercase letters, digits, - or _", option.Name)
		}
		if seen[option.Name] {
			return fmt.Errorf("%s has two options named %q", parent, option.Name)
		}
		seen[option.Name] = true
		if err := validDescription("option "+where, option.Description); err != nil {
			return err
		}
		switch option.Type {
		case OptionTypeSubcommandGroup, OptionTypeSubcommand:
			if option.Type == OptionTypeSubcommandGroup && depth > 0 || option.Type == OptionTypeSubcommand && depth > 1 {
				return fmt.Errorf("%s is nested too deeply", where)
			}
			if err := validOptions(where, option.Options, depth+1); err != nil {
				return err
			}
			continue
		case OptionTypeString, OptionTypeInteger, OptionTypeBoolean, OptionTypeUser, OptionTypeChannel, OptionTypeRole, OptionTypeMentionable, OptionTypeNumber, OptionTypeAttachment:
		default:
			return fmt.Errorf("%s has unknown option type %d", where, option.Type)
		}
		if option.Required && optional {
			return fmt.Errorf("required option %s must come before optional ones", where)
		}
		optional = optional || !option.Required
		if len(option.Choices) > MaxCommandChoices {
			return fmt.Errorf("%s has %d choices, Micup allows %d", where, len(option.Choices), MaxCommandChoices)
		}
		if len(option.Choices) > 0 && option.Autocomplete {
			return fmt.Errorf("%s cannot use both choices and autocomplete", where)
		}
	}
	return nil
}

type CommandData struct {
	ID       string              `json:"id"`
	Name     string              `json:"name"`
	Type     int                 `json:"type"`
	GuildID  string              `json:"guild_id,omitempty"`
	TargetID string              `json:"target_id,omitempty"`
	Options  []CommandDataOption `json:"options,omitempty"`
	Resolved *Resolved           `json:"resolved,omitempty"`
}

type CommandDataOption struct {
	Name    string              `json:"name"`
	Type    int                 `json:"type"`
	Value   json.RawMessage     `json:"value,omitempty"`
	Options []CommandDataOption `json:"options,omitempty"`
	Focused bool                `json:"focused,omitempty"`
}

type Resolved struct {
	Users       map[string]User       `json:"users,omitempty"`
	Members     map[string]Member     `json:"members,omitempty"`
	Roles       map[string]Role       `json:"roles,omitempty"`
	Channels    map[string]Channel    `json:"channels,omitempty"`
	Messages    map[string]Message    `json:"messages,omitempty"`
	Attachments map[string]Attachment `json:"attachments,omitempty"`
}

func (i Interaction) CommandData() (CommandData, bool) {
	var data CommandData
	if i.Type != InteractionTypeApplicationCommand && i.Type != InteractionTypeAutocomplete || json.Unmarshal(i.Data, &data) != nil {
		return CommandData{}, false
	}
	return data, true
}

func (i Interaction) IsCommand() bool { return i.Type == InteractionTypeApplicationCommand }

func (i Interaction) IsAutocomplete() bool { return i.Type == InteractionTypeAutocomplete }

func (d CommandData) Path() []string {
	path := []string{d.Name}
	options := d.Options
	for len(options) == 1 && (options[0].Type == OptionTypeSubcommandGroup || options[0].Type == OptionTypeSubcommand) {
		path = append(path, options[0].Name)
		options = options[0].Options
	}
	return path
}

func (d CommandData) Leaf() []CommandDataOption {
	options := d.Options
	for len(options) == 1 && (options[0].Type == OptionTypeSubcommandGroup || options[0].Type == OptionTypeSubcommand) {
		options = options[0].Options
	}
	return options
}

func (d CommandData) Option(name string) (CommandDataOption, bool) {
	for _, option := range d.Leaf() {
		if option.Name == name {
			return option, true
		}
	}
	return CommandDataOption{}, false
}

func (d CommandData) Focused() (CommandDataOption, bool) {
	for _, option := range d.Leaf() {
		if option.Focused {
			return option, true
		}
	}
	return CommandDataOption{}, false
}

func (d CommandData) String(name string) (string, bool) {
	option, ok := d.Option(name)
	if !ok {
		return "", false
	}
	return option.String()
}

func (d CommandData) Int(name string) (int64, bool) {
	option, ok := d.Option(name)
	if !ok {
		return 0, false
	}
	return option.Int()
}

func (d CommandData) Float(name string) (float64, bool) {
	option, ok := d.Option(name)
	if !ok {
		return 0, false
	}
	return option.Float()
}

func (d CommandData) Bool(name string) (bool, bool) {
	option, ok := d.Option(name)
	if !ok {
		return false, false
	}
	var value bool
	return value, json.Unmarshal(option.Value, &value) == nil
}

func (d CommandData) User(name string) (User, bool) {
	id, ok := d.String(name)
	if !ok || d.Resolved == nil {
		return User{}, false
	}
	user, ok := d.Resolved.Users[id]
	return user, ok
}

func (d CommandData) Member(name string) (Member, bool) {
	id, ok := d.String(name)
	if !ok || d.Resolved == nil {
		return Member{}, false
	}
	member, ok := d.Resolved.Members[id]
	if ok && member.User == nil {
		if user, found := d.Resolved.Users[id]; found {
			member.User = &user
		}
	}
	return member, ok
}

func (d CommandData) Role(name string) (Role, bool) {
	id, ok := d.String(name)
	if !ok || d.Resolved == nil {
		return Role{}, false
	}
	role, ok := d.Resolved.Roles[id]
	return role, ok
}

func (d CommandData) Channel(name string) (Channel, bool) {
	id, ok := d.String(name)
	if !ok || d.Resolved == nil {
		return Channel{}, false
	}
	channel, ok := d.Resolved.Channels[id]
	return channel, ok
}

func (d CommandData) Attachment(name string) (Attachment, bool) {
	id, ok := d.String(name)
	if !ok || d.Resolved == nil {
		return Attachment{}, false
	}
	attachment, ok := d.Resolved.Attachments[id]
	return attachment, ok
}

func (d CommandData) TargetUser() (User, bool) {
	if d.Type != CommandTypeUser || d.Resolved == nil {
		return User{}, false
	}
	user, ok := d.Resolved.Users[d.TargetID]
	return user, ok
}

func (d CommandData) TargetMessage() (Message, bool) {
	if d.Type != CommandTypeMessage || d.Resolved == nil {
		return Message{}, false
	}
	message, ok := d.Resolved.Messages[d.TargetID]
	return message, ok
}

func (o CommandDataOption) String() (string, bool) {
	var text string
	if json.Unmarshal(o.Value, &text) == nil {
		return text, true
	}
	var number json.Number
	if json.Unmarshal(o.Value, &number) == nil {
		return number.String(), true
	}
	return "", false
}

func (o CommandDataOption) Int() (int64, bool) {
	text, ok := o.String()
	if !ok {
		return 0, false
	}
	value, err := strconv.ParseInt(text, 10, 64)
	return value, err == nil
}

func (o CommandDataOption) Float() (float64, bool) {
	text, ok := o.String()
	if !ok {
		return 0, false
	}
	value, err := strconv.ParseFloat(text, 64)
	return value, err == nil
}
