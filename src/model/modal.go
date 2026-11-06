package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"unicode/utf8"
)

const (
	ComponentTypeTextInput = 4

	TextInputStyleShort     = 1
	TextInputStyleParagraph = 2

	MaxModalInputs      = 5
	MaxModalTitle       = 45
	MaxTextInputLabel   = 45
	MaxTextInputLength  = 4000
	MaxTextInputPrompts = 100
)

type Modal struct {
	CustomID string
	Title    string
	Inputs   []TextInput
}

type TextInput struct {
	CustomID    string
	Label       string
	Style       int
	Placeholder string
	Value       string
	MinLength   *int
	MaxLength   *int
	Required    *bool
}

func ShortInput(customID, label string) TextInput {
	return TextInput{CustomID: customID, Label: label, Style: TextInputStyleShort}
}

func ParagraphInput(customID, label string) TextInput {
	return TextInput{CustomID: customID, Label: label, Style: TextInputStyleParagraph}
}

func (t TextInput) WithLength(min, max int) TextInput {
	t.MinLength, t.MaxLength = &min, &max
	return t
}

func (t TextInput) WithRequired(required bool) TextInput {
	t.Required = &required
	return t
}

func (t TextInput) WithValue(value string) TextInput {
	t.Value = value
	return t
}

func (t TextInput) WithPlaceholder(placeholder string) TextInput {
	t.Placeholder = placeholder
	return t
}

func (m Modal) Validate() error {
	if n := utf8.RuneCountInString(m.Title); n == 0 || n > MaxModalTitle {
		return fmt.Errorf("modal title must be 1-%d characters", MaxModalTitle)
	}
	if n := len(m.CustomID); n == 0 || n > MaxCustomIDLength {
		return fmt.Errorf("modal custom_id must be 1-%d characters", MaxCustomIDLength)
	}
	if len(m.Inputs) == 0 || len(m.Inputs) > MaxModalInputs {
		return fmt.Errorf("a modal holds 1-%d text inputs", MaxModalInputs)
	}
	seen := map[string]bool{}
	for _, input := range m.Inputs {
		if n := len(input.CustomID); n == 0 || n > MaxCustomIDLength {
			return fmt.Errorf("text input custom_id must be 1-%d characters", MaxCustomIDLength)
		}
		if seen[input.CustomID] {
			return fmt.Errorf("two text inputs use custom_id %q", input.CustomID)
		}
		seen[input.CustomID] = true
		if n := utf8.RuneCountInString(input.Label); n == 0 || n > MaxTextInputLabel {
			return fmt.Errorf("text input %q needs a label of 1-%d characters", input.CustomID, MaxTextInputLabel)
		}
		if input.Style != TextInputStyleShort && input.Style != TextInputStyleParagraph {
			return fmt.Errorf("text input %q has unknown style %d", input.CustomID, input.Style)
		}
		if utf8.RuneCountInString(input.Placeholder) > MaxTextInputPrompts {
			return fmt.Errorf("text input %q placeholder is over %d characters", input.CustomID, MaxTextInputPrompts)
		}
		low, high := 0, MaxTextInputLength
		if input.MinLength != nil {
			low = *input.MinLength
		}
		if input.MaxLength != nil {
			high = *input.MaxLength
		}
		if low < 0 || high < 1 || low > high || high > MaxTextInputLength {
			return fmt.Errorf("text input %q length must stay within 0-%d with min <= max", input.CustomID, MaxTextInputLength)
		}
		if n := utf8.RuneCountInString(input.Value); n > high {
			return fmt.Errorf("text input %q value is longer than its maximum", input.CustomID)
		}
	}
	return nil
}

func (m Modal) MarshalJSON() ([]byte, error) {
	rows := make([]map[string]any, len(m.Inputs))
	for i, input := range m.Inputs {
		rows[i] = map[string]any{"type": ComponentTypeActionRow, "components": []TextInput{input}}
	}
	return json.Marshal(map[string]any{"custom_id": m.CustomID, "title": m.Title, "components": rows})
}

func (t TextInput) MarshalJSON() ([]byte, error) {
	body := map[string]any{"type": ComponentTypeTextInput, "custom_id": t.CustomID, "label": t.Label, "style": t.Style}
	if t.Placeholder != "" {
		body["placeholder"] = t.Placeholder
	}
	if t.Value != "" {
		body["value"] = t.Value
	}
	if t.MinLength != nil {
		body["min_length"] = *t.MinLength
	}
	if t.MaxLength != nil {
		body["max_length"] = *t.MaxLength
	}
	if t.Required != nil {
		body["required"] = *t.Required
	}
	return json.Marshal(body)
}

type ModalData struct {
	CustomID string
	Values   map[string]string
}

func (d ModalData) Value(customID string) (string, bool) {
	value, ok := d.Values[customID]
	return value, ok
}

func (i Interaction) IsModalSubmit() bool { return i.Type == InteractionTypeModalSubmit }

func (i Interaction) ModalData() (ModalData, bool) {
	if i.Type != InteractionTypeModalSubmit {
		return ModalData{}, false
	}
	var raw struct {
		CustomID   string `json:"custom_id"`
		Components []struct {
			Components []struct {
				CustomID string `json:"custom_id"`
				Value    string `json:"value"`
			} `json:"components"`
			CustomID string `json:"custom_id"`
			Value    string `json:"value"`
		} `json:"components"`
	}
	if json.Unmarshal(i.Data, &raw) != nil || raw.CustomID == "" {
		return ModalData{}, false
	}
	data := ModalData{CustomID: raw.CustomID, Values: map[string]string{}}
	for _, row := range raw.Components {
		if row.CustomID != "" {
			data.Values[row.CustomID] = row.Value
		}
		for _, input := range row.Components {
			data.Values[input.CustomID] = input.Value
		}
	}
	return data, true
}

var errModalTarget = errors.New("only an application command or component interaction can open a modal")

func ValidateModalTarget(i Interaction) error {
	if i.Type != InteractionTypeApplicationCommand && i.Type != InteractionTypeMessageComponent {
		return errModalTarget
	}
	return nil
}
