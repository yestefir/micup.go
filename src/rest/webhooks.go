package rest

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/yestefir/micup.go/src/model"
)

func (c *Client) CreateWebhook(ctx context.Context, channelID, name string, options ...RequestOption) (model.Webhook, error) {
	if channelID == "" || name == "" || len(name) > 80 {
		return model.Webhook{}, errors.New("channel ID and a webhook name of 1 to 80 characters are required")
	}
	var webhook model.Webhook
	err := c.call(ctx, apply(&request{method: http.MethodPost, path: path("channels", url.PathEscape(channelID), "webhooks"), body: map[string]string{"name": name}, dest: &webhook}, options))
	return webhook, err
}

func (c *Client) ChannelWebhooks(ctx context.Context, channelID string) ([]model.Webhook, error) {
	if channelID == "" {
		return nil, errors.New("channel ID is required")
	}
	var webhooks []model.Webhook
	err := c.call(ctx, &request{method: http.MethodGet, path: path("channels", url.PathEscape(channelID), "webhooks"), dest: &webhooks})
	return webhooks, err
}

func (c *Client) GuildWebhooks(ctx context.Context, guildID string) ([]model.Webhook, error) {
	if guildID == "" {
		return nil, errors.New("guild ID is required")
	}
	var webhooks []model.Webhook
	err := c.call(ctx, &request{method: http.MethodGet, path: path("guilds", url.PathEscape(guildID), "webhooks"), dest: &webhooks})
	return webhooks, err
}

func (c *Client) DeleteWebhook(ctx context.Context, webhookID string, options ...RequestOption) error {
	if webhookID == "" {
		return errors.New("webhook ID is required")
	}
	return c.call(ctx, apply(&request{method: http.MethodDelete, path: path("webhooks", url.PathEscape(webhookID))}, options))
}

func (c *Client) ExecuteWebhook(ctx context.Context, webhookID, token string, message model.WebhookMessage) (*model.Message, error) {
	if webhookID == "" || token == "" {
		return nil, errors.New("webhook ID and token are required")
	}
	options := message.MessageOptions
	if options.Content == "" && len(options.Embeds) == 0 && len(options.Files) == 0 && len(options.Attachments) == 0 && len(options.StickerIDs) == 0 {
		return nil, errors.New("webhook message content, an embed, a file or a sticker is required")
	}
	if len(options.Components) != 0 {
		return nil, errors.New("webhook messages cannot carry buttons")
	}
	body, err := messageBody(options.Content, options.Embeds, nil, options.AllowedMentions, options.Files, options.Attachments, "webhook message")
	if err != nil {
		return nil, err
	}
	if message.Username != "" {
		body["username"] = message.Username
	}
	if message.AvatarURL != "" {
		body["avatar_url"] = message.AvatarURL
	}
	if options.Nonce != "" {
		body["nonce"] = options.Nonce
	}
	if options.Flags != 0 {
		body["flags"] = options.Flags
	}
	if options.ReplyTo != "" {
		body["message_reference"] = map[string]string{"message_id": options.ReplyTo}
	}
	values := url.Values{}
	if message.Wait {
		values.Set("wait", "true")
	}
	setQuery(values, "thread_id", message.ThreadID)
	r := &request{method: http.MethodPost, path: path("webhooks", url.PathEscape(webhookID), url.PathEscape(token)), query: values, body: body, files: options.Files, idempotent: options.Nonce != ""}
	if !message.Wait {
		return nil, c.call(ctx, r)
	}
	var created model.Message
	r.dest = &created
	if err = c.call(ctx, r); err != nil {
		return nil, err
	}
	return &created, nil
}
