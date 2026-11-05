package rest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/yestefir/micup.go/src/model"
)

func (c *Client) SendMessage(ctx context.Context, channelID, content, replyTo string) (model.Message, error) {
	if content == "" {
		return model.Message{}, errors.New("message content is required")
	}
	return c.SendMessageWithOptions(ctx, channelID, model.MessageOptions{Content: content, ReplyTo: replyTo})
}

func (c *Client) SendMessageWithOptions(ctx context.Context, channelID string, options model.MessageOptions) (model.Message, error) {
	if channelID == "" {
		return model.Message{}, errors.New("channel ID is required")
	}
	body, err := messageBody(options, "message")
	if err != nil {
		return model.Message{}, err
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
	var message model.Message
	err = c.Do(ctx, http.MethodPost, path("channels", url.PathEscape(channelID), "messages"), body, &message)
	return message, err
}

func messageBody(options model.MessageOptions, kind string) (map[string]any, error) {
	if options.Content == "" && len(options.Embeds) == 0 {
		return nil, fmt.Errorf("%s content or at least one embed is required", kind)
	}
	if len(options.Embeds) > model.MaxEmbeds {
		return nil, fmt.Errorf("%s can contain at most %d embeds", kind, model.MaxEmbeds)
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
