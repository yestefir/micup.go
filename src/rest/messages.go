package rest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

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
	if options.Content == "" && len(options.Embeds) == 0 && len(options.Components) == 0 && len(options.Files) == 0 && len(options.Attachments) == 0 && len(options.StickerIDs) == 0 {
		return model.Message{}, errors.New("message content, an embed, a button, a file or a sticker is required")
	}
	body, err := messageBody(options.Content, options.Embeds, options.Components, options.AllowedMentions, options.Files, options.Attachments, "message")
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
	if options.Flags != 0 {
		body["flags"] = options.Flags
	}
	if options.TTS {
		body["tts"] = true
	}
	if len(options.StickerIDs) != 0 {
		body["sticker_ids"] = options.StickerIDs
	}
	var message model.Message
	err = c.call(ctx, &request{method: http.MethodPost, path: path("channels", url.PathEscape(channelID), "messages"), body: body, files: options.Files, dest: &message, idempotent: options.Nonce != ""})
	return message, err
}

func (c *Client) Message(ctx context.Context, channelID, messageID string) (model.Message, error) {
	if channelID == "" || messageID == "" {
		return model.Message{}, errors.New("channel ID and message ID are required")
	}
	var message model.Message
	err := c.call(ctx, &request{method: http.MethodGet, path: messagePath(channelID, messageID), dest: &message})
	return message, err
}

func (c *Client) Messages(ctx context.Context, channelID string, query model.MessageQuery) ([]model.Message, error) {
	if channelID == "" {
		return nil, errors.New("channel ID is required")
	}
	if query.Limit < 0 || query.Limit > model.MaxMessagesPerPage {
		return nil, fmt.Errorf("message limit must be between 1 and %d", model.MaxMessagesPerPage)
	}
	values := url.Values{}
	setQuery(values, "limit", limitString(query.Limit))
	setQuery(values, "before", query.Before)
	setQuery(values, "after", query.After)
	setQuery(values, "around", query.Around)
	var messages []model.Message
	err := c.call(ctx, &request{method: http.MethodGet, path: path("channels", url.PathEscape(channelID), "messages"), query: values, dest: &messages})
	return messages, err
}

func (c *Client) EditMessage(ctx context.Context, channelID, messageID string, edit model.MessageEdit) (model.Message, error) {
	if channelID == "" || messageID == "" {
		return model.Message{}, errors.New("channel ID and message ID are required")
	}
	body, err := editBody(edit, "message")
	if err != nil {
		return model.Message{}, err
	}
	if attachments := attachmentMetadata(edit.Files, edit.Attachments); len(attachments) != 0 {
		body["attachments"] = attachments
	}
	var message model.Message
	err = c.call(ctx, &request{method: http.MethodPatch, path: messagePath(channelID, messageID), body: body, files: edit.Files, dest: &message})
	return message, err
}

func (c *Client) DeleteMessage(ctx context.Context, channelID, messageID string, options ...RequestOption) error {
	if channelID == "" || messageID == "" {
		return errors.New("channel ID and message ID are required")
	}
	return c.call(ctx, apply(&request{method: http.MethodDelete, path: messagePath(channelID, messageID)}, options))
}

func (c *Client) BulkDeleteMessages(ctx context.Context, channelID string, messageIDs []string, options ...RequestOption) error {
	if channelID == "" {
		return errors.New("channel ID is required")
	}
	if len(messageIDs) == 0 || len(messageIDs) > model.MaxBulkDelete {
		return fmt.Errorf("bulk delete takes 1 to %d message IDs", model.MaxBulkDelete)
	}
	return c.call(ctx, apply(&request{method: http.MethodPost, path: path("channels", url.PathEscape(channelID), "messages", "bulk-delete"), body: map[string]any{"message_ids": messageIDs}}, options))
}

func (c *Client) TriggerTyping(ctx context.Context, channelID string) error {
	if channelID == "" {
		return errors.New("channel ID is required")
	}
	return c.call(ctx, &request{method: http.MethodPost, path: path("channels", url.PathEscape(channelID), "typing")})
}

func editBody(edit model.MessageEdit, kind string) (map[string]any, error) {
	body := map[string]any{}
	if edit.Content != nil {
		body["content"] = *edit.Content
	}
	if edit.Embeds != nil {
		if len(*edit.Embeds) > model.MaxEmbeds {
			return nil, fmt.Errorf("%s can contain at most %d embeds", kind, model.MaxEmbeds)
		}
		body["embeds"] = *edit.Embeds
	}
	if edit.Components != nil {
		if err := validateComponents(*edit.Components, kind); err != nil {
			return nil, err
		}
		body["components"] = componentList(*edit.Components)
	}
	if edit.Flags != nil {
		body["flags"] = *edit.Flags
	}
	if edit.AllowedMentions != nil {
		body["allowed_mentions"] = edit.AllowedMentions
	}
	return body, nil
}

func messagePath(channelID, messageID string) string {
	return path("channels", url.PathEscape(channelID), "messages", url.PathEscape(messageID))
}

func messageBody(content string, embeds []model.Embed, components []model.ActionRow, mentions *model.AllowedMentions, files []model.File, uploads []model.UploadedAttachment, kind string) (map[string]any, error) {
	if len(embeds) > model.MaxEmbeds {
		return nil, fmt.Errorf("%s can contain at most %d embeds", kind, model.MaxEmbeds)
	}
	if len(files)+len(uploads) > model.MaxAttachments {
		return nil, fmt.Errorf("%s can contain at most %d attachments", kind, model.MaxAttachments)
	}
	if err := validateComponents(components, kind); err != nil {
		return nil, err
	}
	body := map[string]any{}
	if mentions != nil {
		body["allowed_mentions"] = mentions
	} else {
		body["allowed_mentions"] = map[string]any{"parse": []string{}}
	}
	if content != "" {
		body["content"] = content
	}
	if len(embeds) != 0 {
		body["embeds"] = embeds
	}
	if len(components) != 0 {
		body["components"] = components
	}
	if attachments := attachmentMetadata(files, uploads); len(attachments) != 0 {
		body["attachments"] = attachments
	}
	return body, nil
}

func attachmentMetadata(files []model.File, uploads []model.UploadedAttachment) []map[string]any {
	var items []map[string]any
	for i, file := range files {
		item := map[string]any{"id": i, "filename": file.Name}
		if file.Description != "" {
			item["description"] = file.Description
		}
		items = append(items, item)
	}
	for i, upload := range uploads {
		item := map[string]any{"id": len(files) + i, "filename": upload.Filename, "upload_filename": upload.UploadFilename, "file_size": upload.Size, "content_type": contentTypeOr(upload.ContentType)}
		if upload.Description != "" {
			item["description"] = upload.Description
		}
		items = append(items, item)
	}
	return items
}

func apply(r *request, options []RequestOption) *request {
	for _, option := range options {
		option(r)
	}
	return r
}

func setQuery(values url.Values, key, value string) {
	if value != "" {
		values.Set(key, value)
	}
}

func limitString(limit int) string {
	if limit <= 0 {
		return ""
	}
	return strconv.Itoa(limit)
}
