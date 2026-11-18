package rest

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/yestefir/micup.go/src/model"
)

type identityCache struct {
	mu   sync.Mutex
	self *model.User
	app  *Application
}

type PublicApplication struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	Icon         *string     `json:"icon"`
	Description  *string     `json:"description"`
	RedirectURIs []string    `json:"redirect_uris"`
	Scopes       []string    `json:"scopes"`
	BotPublic    bool        `json:"bot_public"`
	Bot          *model.User `json:"bot"`
}

func TokenID(token string) (string, bool) {
	head, _, found := strings.Cut(token, ".")
	if !found || head == "" {
		return "", false
	}
	decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(head, "="))
	if err != nil || len(decoded) == 0 {
		return "", false
	}
	for _, digit := range decoded {
		if digit < '0' || digit > '9' {
			return "", false
		}
	}
	return string(decoded), true
}

func (c *Client) TokenID() (string, bool) { return TokenID(c.token) }

func (c *Client) Self(ctx context.Context) (model.User, error) {
	c.identity.mu.Lock()
	cached := c.identity.self
	c.identity.mu.Unlock()
	if cached != nil {
		return *cached, nil
	}
	user, err := c.CurrentUser(ctx)
	if err != nil {
		return model.User{}, err
	}
	if user.ID == "" {
		return model.User{}, errors.New("Micup returned no ID for the current user")
	}
	c.identity.mu.Lock()
	c.identity.self = &user
	c.identity.mu.Unlock()
	return user, nil
}

func (c *Client) SelfID(ctx context.Context) (string, error) {
	user, err := c.Self(ctx)
	return user.ID, err
}

func (c *Client) Application(ctx context.Context) (Application, error) {
	c.identity.mu.Lock()
	cached := c.identity.app
	c.identity.mu.Unlock()
	if cached != nil {
		return *cached, nil
	}
	app, err := c.CurrentApplication(ctx)
	if err != nil {
		return Application{}, err
	}
	if app.ID == "" {
		return Application{}, errors.New("Micup returned no ID for the current application")
	}
	c.identity.mu.Lock()
	c.identity.app = &app
	c.identity.mu.Unlock()
	return app, nil
}

func (c *Client) ApplicationID(ctx context.Context) (string, error) {
	if app, err := c.Application(ctx); err == nil {
		return app.ID, nil
	}
	return c.SelfID(ctx)
}

func (c *Client) resolveApplicationID(ctx context.Context, applicationID string) (string, error) {
	if applicationID != "" {
		return applicationID, nil
	}
	return c.ApplicationID(ctx)
}

func (c *Client) PublicApplication(ctx context.Context, applicationID string) (PublicApplication, error) {
	applicationID, err := c.resolveApplicationID(ctx, applicationID)
	if err != nil {
		return PublicApplication{}, err
	}
	var app PublicApplication
	err = c.call(ctx, &request{method: http.MethodGet, path: path("oauth2", "applications", url.PathEscape(applicationID), "public"), dest: &app})
	return app, err
}
