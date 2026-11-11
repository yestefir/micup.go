package rest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"unicode/utf8"

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
	if len(options.StickerIDs) != 0 {
		body["sticker_ids"] = options.StickerIDs
	}
	if message.ThreadName != "" {
		if message.ThreadID != "" {
			return nil, errors.New("set either ThreadID to post in a thread or ThreadName to create a forum post")
		}
		if utf8.RuneCountInString(message.ThreadName) > model.MaxChannelName {
			return nil, fmt.Errorf("thread name is longer than %d characters", model.MaxChannelName)
		}
		body["thread_name"] = message.ThreadName
	}
	if len(message.AppliedTags) != 0 {
		if message.ThreadName == "" || len(message.AppliedTags) > model.MaxAppliedTags {
			return nil, fmt.Errorf("applied tags need ThreadName and take at most %d tags", model.MaxAppliedTags)
		}
		body["applied_tags"] = message.AppliedTags
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

func (c *Client) Webhook(ctx context.Context, webhookID string) (model.Webhook, error) {
	if webhookID == "" {
		return model.Webhook{}, errors.New("webhook ID is required")
	}
	var webhook model.Webhook
	err := c.call(ctx, &request{method: http.MethodGet, path: path("webhooks", url.PathEscape(webhookID)), dest: &webhook})
	return webhook, err
}

func (c *Client) CreateWebhookWithAvatar(ctx context.Context, channelID, name, avatar string, options ...RequestOption) (model.Webhook, error) {
	if channelID == "" || name == "" || utf8.RuneCountInString(name) > model.MaxWebhookName {
		return model.Webhook{}, fmt.Errorf("channel ID and a webhook name of 1 to %d characters are required", model.MaxWebhookName)
	}
	body := map[string]any{"name": name}
	if avatar != "" {
		body["avatar"] = avatar
	}
	var webhook model.Webhook
	err := c.call(ctx, apply(&request{method: http.MethodPost, path: path("channels", url.PathEscape(channelID), "webhooks"), body: body, dest: &webhook}, options))
	return webhook, err
}

func (c *Client) ModifyWebhook(ctx context.Context, webhookID string, edit model.WebhookEdit, options ...RequestOption) (model.Webhook, error) {
	if webhookID == "" {
		return model.Webhook{}, errors.New("webhook ID is required")
	}
	body, err := webhookBody(edit)
	if err != nil {
		return model.Webhook{}, err
	}
	var webhook model.Webhook
	err = c.call(ctx, apply(&request{method: http.MethodPatch, path: path("webhooks", url.PathEscape(webhookID)), body: body, dest: &webhook}, options))
	return webhook, err
}

func (c *Client) WebhookWithToken(ctx context.Context, webhookID, token string) (model.Webhook, error) {
	p, err := webhookTokenPath(webhookID, token)
	if err != nil {
		return model.Webhook{}, err
	}
	var webhook model.Webhook
	err = c.call(ctx, &request{method: http.MethodGet, path: p, dest: &webhook, anonymous: true})
	return webhook, err
}

func (c *Client) ModifyWebhookWithToken(ctx context.Context, webhookID, token string, edit model.WebhookEdit) (model.Webhook, error) {
	p, err := webhookTokenPath(webhookID, token)
	if err != nil {
		return model.Webhook{}, err
	}
	if edit.ChannelID.IsSet() {
		return model.Webhook{}, errors.New("a webhook token cannot move the webhook; use ModifyWebhook")
	}
	body, err := webhookBody(edit)
	if err != nil {
		return model.Webhook{}, err
	}
	var webhook model.Webhook
	err = c.call(ctx, &request{method: http.MethodPatch, path: p, body: body, dest: &webhook, anonymous: true})
	return webhook, err
}

func (c *Client) DeleteWebhookWithToken(ctx context.Context, webhookID, token string) error {
	p, err := webhookTokenPath(webhookID, token)
	if err != nil {
		return err
	}
	return c.call(ctx, &request{method: http.MethodDelete, path: p, anonymous: true})
}

func (c *Client) WebhookMessage(ctx context.Context, webhookID, token, messageID, threadID string) (model.Message, error) {
	r, err := webhookMessageRequest(http.MethodGet, webhookID, token, messageID, threadID)
	if err != nil {
		return model.Message{}, err
	}
	var message model.Message
	r.dest = &message
	err = c.call(ctx, r)
	return message, err
}

func (c *Client) EditWebhookMessage(ctx context.Context, webhookID, token, messageID, threadID string, edit model.MessageEdit) (model.Message, error) {
	r, err := webhookMessageRequest(http.MethodPatch, webhookID, token, messageID, threadID)
	if err != nil {
		return model.Message{}, err
	}
	if edit.Components != nil || len(edit.Files) != 0 || len(edit.Attachments) != 0 || edit.KeepAttachments != nil {
		return model.Message{}, errors.New("a webhook message edit changes only content, embeds and flags")
	}
	body, err := editBody(edit, "webhook message")
	if err != nil {
		return model.Message{}, err
	}
	delete(body, "allowed_mentions")
	if len(body) == 0 {
		return model.Message{}, errors.New("webhook message edit needs content, embeds or flags")
	}
	var message model.Message
	r.body, r.dest = body, &message
	err = c.call(ctx, r)
	return message, err
}

func (c *Client) DeleteWebhookMessage(ctx context.Context, webhookID, token, messageID, threadID string) error {
	r, err := webhookMessageRequest(http.MethodDelete, webhookID, token, messageID, threadID)
	if err != nil {
		return err
	}
	return c.call(ctx, r)
}

func webhookBody(edit model.WebhookEdit) (map[string]any, error) {
	if name, ok := edit.Name.Value(); edit.Name.IsSet() && (!ok || name == "" || utf8.RuneCountInString(name) > model.MaxWebhookName) {
		return nil, fmt.Errorf("webhook name must be 1 to %d characters", model.MaxWebhookName)
	}
	if edit.ChannelID.IsNull() {
		return nil, errors.New("a webhook always belongs to a channel")
	}
	body := map[string]any{}
	edit.Name.Apply(body, "name")
	edit.Avatar.Apply(body, "avatar")
	edit.ChannelID.Apply(body, "channel_id")
	if len(body) == 0 {
		return nil, errors.New("webhook edit has no changes")
	}
	return body, nil
}

func webhookTokenPath(webhookID, token string) (string, error) {
	if webhookID == "" || token == "" {
		return "", errors.New("webhook ID and token are required")
	}
	return path("webhooks", url.PathEscape(webhookID), url.PathEscape(token)), nil
}

func webhookMessageRequest(method, webhookID, token, messageID, threadID string) (*request, error) {
	p, err := webhookTokenPath(webhookID, token)
	if err != nil {
		return nil, err
	}
	if messageID == "" {
		return nil, errors.New("message ID is required")
	}
	values := url.Values{}
	setQuery(values, "thread_id", threadID)
	return &request{method: method, path: p + "/messages/" + url.PathEscape(messageID), query: values, anonymous: true}, nil
}
