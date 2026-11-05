package rest

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/yestefir/micup.go/src/model"
)

func (c *Client) Channel(ctx context.Context, channelID string) (model.Channel, error) {
	if channelID == "" {
		return model.Channel{}, errors.New("channel ID is required")
	}
	var channel model.Channel
	err := c.Do(ctx, http.MethodGet, path("channels", url.PathEscape(channelID)), nil, &channel)
	return channel, err
}
