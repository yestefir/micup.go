package rest

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/yestefir/micup.go/src/model"
)

func (c *Client) Guild(ctx context.Context, guildID string) (model.Guild, error) {
	if guildID == "" {
		return model.Guild{}, errors.New("guild ID is required")
	}
	var guild model.Guild
	err := c.call(ctx, &request{method: http.MethodGet, path: path("guilds", url.PathEscape(guildID)), dest: &guild})
	return guild, err
}

func (c *Client) GuildChannels(ctx context.Context, guildID string) ([]model.Channel, error) {
	if guildID == "" {
		return nil, errors.New("guild ID is required")
	}
	var channels []model.Channel
	err := c.call(ctx, &request{method: http.MethodGet, path: path("guilds", url.PathEscape(guildID), "channels"), dest: &channels})
	return channels, err
}

func (c *Client) Roles(ctx context.Context, guildID string) ([]model.Role, error) {
	if guildID == "" {
		return nil, errors.New("guild ID is required")
	}
	var roles []model.Role
	err := c.call(ctx, &request{method: http.MethodGet, path: path("guilds", url.PathEscape(guildID), "roles"), dest: &roles})
	return roles, err
}
