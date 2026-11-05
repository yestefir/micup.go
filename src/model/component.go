package model

import "encoding/json"

const (
	ComponentTypeActionRow = 1
	ComponentTypeButton    = 2
)

const (
	ButtonStylePrimary   = 1
	ButtonStyleSecondary = 2
	ButtonStyleSuccess   = 3
	ButtonStyleDanger    = 4
	ButtonStyleLink      = 5
)

const (
	MaxActionRows        = 5
	MaxButtonsPerRow     = 5
	MaxCustomIDLength    = 100
	MaxButtonLabelLength = 80
)

type ActionRow struct {
	Type       int      `json:"type"`
	Components []Button `json:"components"`
}

type Button struct {
	Type     int    `json:"type"`
	Style    int    `json:"style"`
	Label    string `json:"label,omitempty"`
	Emoji    *Emoji `json:"emoji,omitempty"`
	CustomID string `json:"custom_id,omitempty"`
	URL      string `json:"url,omitempty"`
	Disabled bool   `json:"disabled,omitempty"`
}

func (r ActionRow) MarshalJSON() ([]byte, error) {
	type plain ActionRow
	if r.Type == 0 {
		r.Type = ComponentTypeActionRow
	}
	if r.Components == nil {
		r.Components = []Button{}
	}
	return json.Marshal(plain(r))
}

func (b Button) MarshalJSON() ([]byte, error) {
	type plain Button
	if b.Type == 0 {
		b.Type = ComponentTypeButton
	}
	return json.Marshal(plain(b))
}

func NewActionRow(buttons ...Button) ActionRow {
	return ActionRow{Type: ComponentTypeActionRow, Components: buttons}
}

func PrimaryButton(customID, label string) Button {
	return Button{Type: ComponentTypeButton, Style: ButtonStylePrimary, CustomID: customID, Label: label}
}

func SecondaryButton(customID, label string) Button {
	return Button{Type: ComponentTypeButton, Style: ButtonStyleSecondary, CustomID: customID, Label: label}
}

func SuccessButton(customID, label string) Button {
	return Button{Type: ComponentTypeButton, Style: ButtonStyleSuccess, CustomID: customID, Label: label}
}

func DangerButton(customID, label string) Button {
	return Button{Type: ComponentTypeButton, Style: ButtonStyleDanger, CustomID: customID, Label: label}
}

func LinkButton(url, label string) Button {
	return Button{Type: ComponentTypeButton, Style: ButtonStyleLink, URL: url, Label: label}
}

func (b Button) WithEmoji(emoji Emoji) Button {
	b.Emoji = &emoji
	return b
}

func (b Button) WithDisabled(disabled bool) Button {
	b.Disabled = disabled
	return b
}

func DisableButtons(rows []ActionRow) []ActionRow {
	out := make([]ActionRow, len(rows))
	for i, row := range rows {
		out[i] = ActionRow{Type: row.Type, Components: make([]Button, len(row.Components))}
		for j, button := range row.Components {
			if button.Style != ButtonStyleLink {
				button.Disabled = true
			}
			out[i].Components[j] = button
		}
	}
	return out
}
