package rest

import (
	"fmt"
	"slices"
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
		if row.Select != nil {
			if len(row.Components) != 0 {
				return fmt.Errorf("%s action row %d cannot hold buttons and a select menu", kind, r)
			}
			if err := validateSelect(*row.Select, seen); err != nil {
				return fmt.Errorf("%s action row %d select menu: %w", kind, r, err)
			}
			continue
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

var selectDefaultTypes = map[int][]string{
	model.ComponentTypeUserSelect:        {model.SelectDefaultUser},
	model.ComponentTypeRoleSelect:        {model.SelectDefaultRole},
	model.ComponentTypeMentionableSelect: {model.SelectDefaultUser, model.SelectDefaultRole},
	model.ComponentTypeChannelSelect:     {model.SelectDefaultChannel},
}

func validateSelect(menu model.SelectMenu, seen map[string]bool) error {
	if !model.IsSelectType(menu.Type) {
		return fmt.Errorf("component type %d is not a select menu", menu.Type)
	}
	if menu.CustomID == "" || utf8.RuneCountInString(menu.CustomID) > model.MaxCustomIDLength {
		return fmt.Errorf("custom ID must be 1 to %d characters", model.MaxCustomIDLength)
	}
	if seen[menu.CustomID] {
		return fmt.Errorf("custom ID %q is used twice", menu.CustomID)
	}
	seen[menu.CustomID] = true
	if utf8.RuneCountInString(menu.Placeholder) > model.MaxSelectPlaceholder {
		return fmt.Errorf("placeholder is longer than %d characters", model.MaxSelectPlaceholder)
	}
	low, high := 1, 1
	if menu.MinValues != nil {
		low = *menu.MinValues
	}
	if menu.MaxValues != nil {
		high = *menu.MaxValues
	}
	if low < 0 || high < 1 || low > high || high > model.MaxSelectOptions {
		return fmt.Errorf("min and max values must stay within 0-%d with min <= max and max >= 1", model.MaxSelectOptions)
	}
	if menu.Type == model.ComponentTypeStringSelect {
		return validateSelectOptions(menu, high)
	}
	if len(menu.Options) != 0 {
		return fmt.Errorf("only a string select menu has options")
	}
	if len(menu.ChannelTypes) != 0 && menu.Type != model.ComponentTypeChannelSelect {
		return fmt.Errorf("only a channel select menu has channel types")
	}
	if len(menu.DefaultValues) > high {
		return fmt.Errorf("%d default values exceed max values %d", len(menu.DefaultValues), high)
	}
	for _, value := range menu.DefaultValues {
		if value.ID == "" || !slices.Contains(selectDefaultTypes[menu.Type], value.Type) {
			return fmt.Errorf("default value %q of type %q does not fit this select menu", value.ID, value.Type)
		}
	}
	return nil
}

func validateSelectOptions(menu model.SelectMenu, high int) error {
	if len(menu.ChannelTypes) != 0 || len(menu.DefaultValues) != 0 {
		return fmt.Errorf("a string select menu marks defaults on its options and has no channel types")
	}
	if len(menu.Options) == 0 || len(menu.Options) > model.MaxSelectOptions {
		return fmt.Errorf("a string select menu needs 1 to %d options", model.MaxSelectOptions)
	}
	if high > len(menu.Options) {
		return fmt.Errorf("max values %d is more than the %d options", high, len(menu.Options))
	}
	values := map[string]bool{}
	defaults := 0
	for o, option := range menu.Options {
		if n := utf8.RuneCountInString(option.Label); n == 0 || n > model.MaxSelectOptionLabel {
			return fmt.Errorf("option %d label must be 1 to %d characters", o, model.MaxSelectOptionLabel)
		}
		if n := utf8.RuneCountInString(option.Value); n == 0 || n > model.MaxSelectOptionValue {
			return fmt.Errorf("option %d value must be 1 to %d characters", o, model.MaxSelectOptionValue)
		}
		if utf8.RuneCountInString(option.Description) > model.MaxSelectOptionDescription {
			return fmt.Errorf("option %d description is longer than %d characters", o, model.MaxSelectOptionDescription)
		}
		if values[option.Value] {
			return fmt.Errorf("option value %q is used twice", option.Value)
		}
		values[option.Value] = true
		if option.Default {
			defaults++
		}
	}
	if defaults > high {
		return fmt.Errorf("%d default options exceed max values %d", defaults, high)
	}
	return nil
}

func componentList(rows []model.ActionRow) []model.ActionRow {
	if rows == nil {
		return []model.ActionRow{}
	}
	return rows
}
