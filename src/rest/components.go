package rest

import (
	"fmt"
	"unicode/utf8"

	"github.com/yestefir/micup.go/src/model"
)

func validateComponents(rows []model.ActionRow, kind string) error {
	if len(rows) > model.MaxActionRows {
		return fmt.Errorf("%s can contain at most %d action rows", kind, model.MaxActionRows)
	}
	seen := map[string]bool{}
	for r, row := range rows {
		if row.Type != 0 && row.Type != model.ComponentTypeActionRow {
			return fmt.Errorf("%s component %d is not an action row", kind, r)
		}
		if len(row.Components) == 0 || len(row.Components) > model.MaxButtonsPerRow {
			return fmt.Errorf("%s action row %d needs 1 to %d buttons", kind, r, model.MaxButtonsPerRow)
		}
		for b, button := range row.Components {
			if err := validateButton(button, seen); err != nil {
				return fmt.Errorf("%s action row %d button %d: %w", kind, r, b, err)
			}
		}
	}
	return nil
}

func validateButton(button model.Button, seen map[string]bool) error {
	if button.Type != 0 && button.Type != model.ComponentTypeButton {
		return fmt.Errorf("component type %d is not a button", button.Type)
	}
	if button.Style < model.ButtonStylePrimary || button.Style > model.ButtonStyleLink {
		return fmt.Errorf("unknown button style %d", button.Style)
	}
	if button.Label == "" && button.Emoji == nil {
		return fmt.Errorf("a label or an emoji is required")
	}
	if utf8.RuneCountInString(button.Label) > model.MaxButtonLabelLength {
		return fmt.Errorf("label is longer than %d characters", model.MaxButtonLabelLength)
	}
	if button.Style == model.ButtonStyleLink {
		if button.URL == "" || button.CustomID != "" {
			return fmt.Errorf("a link button needs a URL and no custom ID")
		}
		return nil
	}
	if button.URL != "" {
		return fmt.Errorf("only link buttons can have a URL")
	}
	if button.CustomID == "" || utf8.RuneCountInString(button.CustomID) > model.MaxCustomIDLength {
		return fmt.Errorf("custom ID must be 1 to %d characters", model.MaxCustomIDLength)
	}
	if seen[button.CustomID] {
		return fmt.Errorf("custom ID %q is used twice", button.CustomID)
	}
	seen[button.CustomID] = true
	return nil
}

func componentList(rows []model.ActionRow) []model.ActionRow {
	if rows == nil {
		return []model.ActionRow{}
	}
	return rows
}
