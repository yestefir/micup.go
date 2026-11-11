package rest

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/yestefir/micup.go/src/model"
)

const messageContentIntent = (1 << 18) | (1 << 19)

type Application struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Flags *uint64 `json:"flags"`
}

func (a Application) MessageContentEnabled() *bool {
	if a.Flags == nil {
		return nil
	}
	enabled := *a.Flags&messageContentIntent != 0
	return &enabled
}

func (c *Client) CurrentUser(ctx context.Context) (model.User, error) {
	var user model.User
	err := c.Do(ctx, http.MethodGet, "/users/@me", nil, &user)
	return user, err
}

func (c *Client) CurrentApplication(ctx context.Context) (Application, error) {
	var app Application
	err := c.Do(ctx, http.MethodGet, "/applications/@me", nil, &app)
	return app, err
}

func (c *Client) CreateDM(ctx context.Context, userID string) (model.Channel, error) {
	if userID == "" {
		return model.Channel{}, errors.New("user ID is required")
	}
	var channel model.Channel
	err := c.call(ctx, &request{method: http.MethodPost, path: "/users/@me/channels", body: map[string]string{"recipient_id": userID}, dest: &channel})
	return channel, err
}

func (c *Client) User(ctx context.Context, userID string) (model.User, error) {
	if userID == "" {
		return model.User{}, errors.New("user ID is required")
	}
	var user model.User
	err := c.call(ctx, &request{method: http.MethodGet, path: path("users", url.PathEscape(userID)), dest: &user})
	return user, err
}

func (c *Client) UserProfile(ctx context.Context, userID string, query model.UserProfileQuery) (model.UserProfile, error) {
	if userID == "" {
		return model.UserProfile{}, errors.New("user ID is required")
	}
	values := url.Values{}
	setQuery(values, "guild_id", query.GuildID)
	if query.WithMutualFriends {
		values.Set("with_mutual_friends", "true")
	}
	if query.WithMutualGuilds {
		values.Set("with_mutual_guilds", "true")
	}
	var profile model.UserProfile
	err := c.call(ctx, &request{method: http.MethodGet, path: path("users", url.PathEscape(userID), "profile"), query: values, dest: &profile})
	return profile, err
}

func (c *Client) DMChannels(ctx context.Context) ([]model.Channel, error) {
	var channels []model.Channel
	err := c.call(ctx, &request{method: http.MethodGet, path: "/users/@me/channels", dest: &channels})
	return channels, err
}

func (c *Client) SendDM(ctx context.Context, userID string, options model.MessageOptions) (model.Message, error) {
	channel, err := c.CreateDM(ctx, userID)
	if err != nil {
		return model.Message{}, err
	}
	return c.SendMessageWithOptions(ctx, channel.ID, options)
}
