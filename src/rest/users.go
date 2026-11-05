package rest

import (
	"context"
	"net/http"

	"github.com/yestefir/micup.go/src/model"
)

const messageContentIntent = (1 << 18) | (1 << 19)

type Application struct {
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
