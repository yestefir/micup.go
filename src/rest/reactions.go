package rest

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/yestefir/micup.go/src/model"
)

func EmojiString(emoji model.Emoji) string {
	if emoji.ID != "" {
		return emoji.Name + ":" + emoji.ID
	}
	return emoji.Name
}

func (c *Client) AddReaction(ctx context.Context, channelID, messageID, emoji string) error {
	return c.reaction(ctx, http.MethodPut, channelID, messageID, emoji, "@me")
}

func (c *Client) RemoveReaction(ctx context.Context, channelID, messageID, emoji string) error {
	return c.reaction(ctx, http.MethodDelete, channelID, messageID, emoji, "@me")
}

func (c *Client) RemoveUserReaction(ctx context.Context, channelID, messageID, emoji, userID string) error {
	if userID == "" {
		return errors.New("user ID is required")
	}
	return c.reaction(ctx, http.MethodDelete, channelID, messageID, emoji, url.PathEscape(userID))
}

func (c *Client) RemoveEmojiReactions(ctx context.Context, channelID, messageID, emoji string) error {
	return c.reaction(ctx, http.MethodDelete, channelID, messageID, emoji, "")
}

func (c *Client) RemoveAllReactions(ctx context.Context, channelID, messageID string) error {
	if channelID == "" || messageID == "" {
		return errors.New("channel ID and message ID are required")
	}
	return c.call(ctx, &request{method: http.MethodDelete, path: messagePath(channelID, messageID) + "/reactions"})
}

func (c *Client) reaction(ctx context.Context, method, channelID, messageID, emoji, target string) error {
	if channelID == "" || messageID == "" {
		return errors.New("channel ID and message ID are required")
	}
	if emoji == "" || len(emoji) > 64 {
		return errors.New("emoji must be 1 to 64 characters")
	}
	p := messagePath(channelID, messageID) + "/reactions/" + url.PathEscape(emoji)
	if target != "" {
		p += "/" + target
	}
	return c.call(ctx, &request{method: method, path: p})
}
