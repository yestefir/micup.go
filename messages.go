package micup

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

const (
	maxEmbeds            = 10
	messageFlagEphemeral = 64
)

type Message struct {
	ID        string  `json:"id"`
	ChannelID string  `json:"channel_id"`
	GuildID   string  `json:"guild_id,omitempty"`
	Content   string  `json:"content"`
	Embeds    []Embed `json:"embeds,omitempty"`
	Flags     int     `json:"flags,omitempty"`
}

func (c *Client) Channel(ctx context.Context, channelID string) (Channel, error) {
	if channelID == "" {
		return Channel{}, errors.New("channel ID is required")
	}
	var channel Channel
	err := c.REST(ctx, http.MethodGet, "/channels/"+url.PathEscape(channelID), nil, &channel)
	return channel, err
}

func (c *Client) SendMessage(ctx context.Context, channelID, content, replyTo string) (Message, error) {
	if content == "" {
		return Message{}, errors.New("message content is required")
	}
	return c.SendMessageWithOptions(ctx, channelID, MessageOptions{Content: content, ReplyTo: replyTo})
}

func (c *Client) SendMessageWithOptions(ctx context.Context, channelID string, options MessageOptions) (Message, error) {
	if channelID == "" {
		return Message{}, errors.New("channel ID is required")
	}
	body, err := messageBody(options, "message")
	if err != nil {
		return Message{}, err
	}
	if options.ReplyTo != "" {
		body["message_reference"] = map[string]string{"message_id": options.ReplyTo, "channel_id": channelID}
	}
	if options.Nonce != "" {
		body["nonce"] = options.Nonce
	}
	if options.EnforceNonce {
		body["enforce_nonce"] = true
	}
	var message Message
	err = c.REST(ctx, http.MethodPost, "/channels/"+url.PathEscape(channelID)+"/messages", body, &message)
	return message, err
}

func messageBody(options MessageOptions, kind string) (map[string]any, error) {
	if options.Content == "" && len(options.Embeds) == 0 {
		return nil, fmt.Errorf("%s content or at least one embed is required", kind)
	}
	if len(options.Embeds) > maxEmbeds {
		return nil, fmt.Errorf("%s can contain at most %d embeds", kind, maxEmbeds)
	}
	body := map[string]any{"allowed_mentions": map[string]any{"parse": []string{}}}
	if options.Content != "" {
		body["content"] = options.Content
	}
	if len(options.Embeds) != 0 {
		body["embeds"] = options.Embeds
	}
	return body, nil
}
