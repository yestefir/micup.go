package rest

import (
	"context"
	"errors"
	"net/http"

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
