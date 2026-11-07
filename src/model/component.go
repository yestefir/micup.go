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
	Type       int
	Components []Button
	Select     *SelectMenu
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
	if r.Type == 0 {
		r.Type = ComponentTypeActionRow
	}
	components := []any{}
	for _, button := range r.Components {
		components = append(components, button)
	}
	if r.Select != nil {
		components = append(components, *r.Select)
	}
	return json.Marshal(struct {
		Type       int   `json:"type"`
		Components []any `json:"components"`
	}{r.Type, components})
}

func (r *ActionRow) UnmarshalJSON(raw []byte) error {
	var row struct {
		Type       int               `json:"type"`
		Components []json.RawMessage `json:"components"`
	}
	if err := json.Unmarshal(raw, &row); err != nil {
		return err
	}
	*r = ActionRow{Type: row.Type}
	for _, component := range row.Components {
		var head struct {
			Type int `json:"type"`
		}
		if err := json.Unmarshal(component, &head); err != nil {
			return err
		}
		if IsSelectType(head.Type) {
			var menu SelectMenu
			if err := json.Unmarshal(component, &menu); err != nil {
				return err
			}
			r.Select = &menu
			continue
		}
		var button Button
		if err := json.Unmarshal(component, &button); err != nil {
			return err
		}
		r.Components = append(r.Components, button)
	}
	return nil
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
		out[i] = ActionRow{Type: row.Type, Components: make([]Button, len(row.Components)), Select: row.Select}
		for j, button := range row.Components {
			if button.Style != ButtonStyleLink {
				button.Disabled = true
			}
			out[i].Components[j] = button
		}
	}
	return out
}

func DisableComponents(rows []ActionRow) []ActionRow {
	out := DisableButtons(rows)
	for i := range out {
		if out[i].Select != nil {
			menu := *out[i].Select
			menu.Disabled = true
			out[i].Select = &menu
		}
	}
	return out
}
