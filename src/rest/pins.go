package rest

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/yestefir/micup.go/src/model"
)

func (c *Client) PinMessage(ctx context.Context, channelID, messageID string, options ...RequestOption) error {
	return c.pin(ctx, http.MethodPut, channelID, messageID, options)
}

func (c *Client) UnpinMessage(ctx context.Context, channelID, messageID string, options ...RequestOption) error {
	return c.pin(ctx, http.MethodDelete, channelID, messageID, options)
}

func (c *Client) PinnedMessages(ctx context.Context, channelID string, query model.PinQuery) (model.PinsPage, error) {
	if channelID == "" {
		return model.PinsPage{}, errors.New("channel ID is required")
	}
	values := url.Values{}
	setQuery(values, "limit", limitString(query.Limit))
	setQuery(values, "before", query.Before)
	var page model.PinsPage
	err := c.call(ctx, &request{method: http.MethodGet, path: path("channels", url.PathEscape(channelID), "messages", "pins"), query: values, dest: &page})
	return page, err
}

func (c *Client) pin(ctx context.Context, method, channelID, messageID string, options []RequestOption) error {
	if channelID == "" || messageID == "" {
		return errors.New("channel ID and message ID are required")
	}
	return c.call(ctx, apply(&request{method: method, path: path("channels", url.PathEscape(channelID), "pins", url.PathEscape(messageID))}, options))
}
