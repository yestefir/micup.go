package rest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/yestefir/micup.go/src/model"
)

func (c *Client) Invite(ctx context.Context, code string) (model.Invite, error) {
	if code == "" {
		return model.Invite{}, errors.New("invite code is required")
	}
	var invite model.Invite
	err := c.call(ctx, &request{method: http.MethodGet, path: path("invites", url.PathEscape(code)), dest: &invite})
	return invite, err
}

func (c *Client) CreateInvite(ctx context.Context, channelID string, create model.InviteCreate, options ...RequestOption) (model.Invite, error) {
	if channelID == "" {
		return model.Invite{}, errors.New("channel ID is required")
	}
	if create.MaxUses < 0 || create.MaxUses > model.MaxInviteUses || create.MaxAge < 0 || create.MaxAge > model.MaxInviteAge {
		return model.Invite{}, fmt.Errorf("an invite takes 0 to %d uses and an age of 0 to %d seconds", model.MaxInviteUses, model.MaxInviteAge)
	}
	body := map[string]any{"max_uses": create.MaxUses, "max_age": create.MaxAge}
	if create.Unique {
		body["unique"] = true
	}
	if create.Temporary {
		body["temporary"] = true
	}
	var invite model.Invite
	err := c.call(ctx, apply(&request{method: http.MethodPost, path: path("channels", url.PathEscape(channelID), "invites"), body: body, dest: &invite}, options))
	return invite, err
}

func (c *Client) ChannelInvites(ctx context.Context, channelID string) ([]model.Invite, error) {
	if channelID == "" {
		return nil, errors.New("channel ID is required")
	}
	var invites []model.Invite
	err := c.call(ctx, &request{method: http.MethodGet, path: path("channels", url.PathEscape(channelID), "invites"), dest: &invites})
	return invites, err
}

func (c *Client) GuildInvites(ctx context.Context, guildID string) ([]model.Invite, error) {
	if guildID == "" {
		return nil, errors.New("guild ID is required")
	}
	var invites []model.Invite
	err := c.call(ctx, &request{method: http.MethodGet, path: path("guilds", url.PathEscape(guildID), "invites"), dest: &invites})
	return invites, err
}

func (c *Client) DeleteInvite(ctx context.Context, code string, options ...RequestOption) error {
	if code == "" {
		return errors.New("invite code is required")
	}
	return c.call(ctx, apply(&request{method: http.MethodDelete, path: path("invites", url.PathEscape(code))}, options))
}
