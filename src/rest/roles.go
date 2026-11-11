package rest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"unicode/utf8"

	"github.com/yestefir/micup.go/src/model"
)

func (c *Client) CreateRole(ctx context.Context, guildID string, create model.RoleCreate, options ...RequestOption) (model.Role, error) {
	if guildID == "" {
		return model.Role{}, errors.New("guild ID is required")
	}
	if err := validateRole(create.Name, create.Color); err != nil {
		return model.Role{}, err
	}
	if create.Icon != "" && create.UnicodeEmoji != "" {
		return model.Role{}, errors.New("a role has an image icon or a Unicode emoji, not both")
	}
	body := map[string]any{"name": create.Name}
	if create.Color != 0 {
		body["color"] = create.Color
	}
	if create.Permissions != nil {
		body["permissions"] = *create.Permissions
	}
	if create.Icon != "" {
		body["icon"] = create.Icon
	}
	if create.UnicodeEmoji != "" {
		body["unicode_emoji"] = create.UnicodeEmoji
	}
	if create.IconSourceRoleID != "" {
		body["icon_source_role_id"] = create.IconSourceRoleID
	}
	var role model.Role
	err := c.call(ctx, apply(&request{method: http.MethodPost, path: path("guilds", url.PathEscape(guildID), "roles"), body: body, dest: &role}, options))
	return role, err
}

func (c *Client) ModifyRole(ctx context.Context, guildID, roleID string, edit model.RoleEdit, options ...RequestOption) (model.Role, error) {
	if guildID == "" || roleID == "" {
		return model.Role{}, errors.New("guild ID and role ID are required")
	}
	name, named := edit.Name.Value()
	color, _ := edit.Color.Value()
	if edit.Name.IsNull() || edit.Color.IsNull() || edit.Permissions.IsNull() || edit.Hoist.IsNull() || edit.Mentionable.IsNull() {
		return model.Role{}, errors.New("only the hoist position, icon and emoji of a role can be cleared")
	}
	if named {
		if err := validateRole(name, color); err != nil {
			return model.Role{}, err
		}
	} else if color < 0 || color > model.MaxRoleColor {
		return model.Role{}, fmt.Errorf("role color must be 0 to %d", model.MaxRoleColor)
	}
	_, icon := edit.Icon.Value()
	_, emoji := edit.UnicodeEmoji.Value()
	if icon && emoji {
		return model.Role{}, errors.New("a role has an image icon or a Unicode emoji, not both")
	}
	body := map[string]any{}
	edit.Name.Apply(body, "name")
	edit.Color.Apply(body, "color")
	edit.Permissions.Apply(body, "permissions")
	edit.Hoist.Apply(body, "hoist")
	edit.HoistPosition.Apply(body, "hoist_position")
	edit.Mentionable.Apply(body, "mentionable")
	edit.Icon.Apply(body, "icon")
	edit.UnicodeEmoji.Apply(body, "unicode_emoji")
	if len(body) == 0 {
		return model.Role{}, errors.New("role edit has no changes")
	}
	var role model.Role
	err := c.call(ctx, apply(&request{method: http.MethodPatch, path: rolePath(guildID, roleID), body: body, dest: &role}, options))
	return role, err
}

func (c *Client) DeleteRole(ctx context.Context, guildID, roleID string, options ...RequestOption) error {
	if guildID == "" || roleID == "" {
		return errors.New("guild ID and role ID are required")
	}
	if guildID == roleID {
		return errors.New("the everyone role cannot be deleted")
	}
	return c.call(ctx, apply(&request{method: http.MethodDelete, path: rolePath(guildID, roleID)}, options))
}

func (c *Client) ModifyRolePositions(ctx context.Context, guildID string, positions []model.RolePosition, options ...RequestOption) error {
	if guildID == "" || len(positions) == 0 {
		return errors.New("guild ID and at least one role position are required")
	}
	for _, position := range positions {
		if position.ID == "" || position.ID == guildID {
			return errors.New("every role position needs a role ID other than the everyone role")
		}
	}
	return c.call(ctx, apply(&request{method: http.MethodPatch, path: path("guilds", url.PathEscape(guildID), "roles"), body: positions}, options))
}

func (c *Client) ModifyRoleHoistPositions(ctx context.Context, guildID string, positions []model.RoleHoistPosition, options ...RequestOption) error {
	if guildID == "" || len(positions) == 0 {
		return errors.New("guild ID and at least one hoist position are required")
	}
	for _, position := range positions {
		if position.ID == "" || position.ID == guildID {
			return errors.New("every hoist position needs a role ID other than the everyone role")
		}
	}
	return c.call(ctx, apply(&request{method: http.MethodPatch, path: path("guilds", url.PathEscape(guildID), "roles", "hoist-positions"), body: positions}, options))
}

func (c *Client) ResetRoleHoistPositions(ctx context.Context, guildID string, options ...RequestOption) error {
	if guildID == "" {
		return errors.New("guild ID is required")
	}
	return c.call(ctx, apply(&request{method: http.MethodDelete, path: path("guilds", url.PathEscape(guildID), "roles", "hoist-positions")}, options))
}

func validateRole(name string, color int) error {
	if n := utf8.RuneCountInString(name); n == 0 || n > model.MaxRoleName {
		return fmt.Errorf("role name must be 1 to %d characters", model.MaxRoleName)
	}
	if color < 0 || color > model.MaxRoleColor {
		return fmt.Errorf("role color must be 0 to %d", model.MaxRoleColor)
	}
	return nil
}

func rolePath(guildID, roleID string) string {
	return path("guilds", url.PathEscape(guildID), "roles", url.PathEscape(roleID))
}
