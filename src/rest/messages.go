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
	if options.Forward != nil {
		return c.forward(ctx, channelID, options)
	}
	if options.Content == "" && len(options.Embeds) == 0 && len(options.Components) == 0 && len(options.Files) == 0 && len(options.Attachments) == 0 && len(options.StickerIDs) == 0 {
		return model.Message{}, errors.New("message content, an embed, a button, a file or a sticker is required")
	}
	if len(options.StickerIDs) > model.MaxStickersPerMessage {
		return model.Message{}, fmt.Errorf("a message can carry at most %d stickers", model.MaxStickersPerMessage)
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

func (c *Client) forward(ctx context.Context, channelID string, options model.MessageOptions) (model.Message, error) {
	forward := options.Forward
	if forward.ChannelID == "" || forward.MessageID == "" {
		return model.Message{}, errors.New("a forward needs the source channel ID and message ID")
	}
	if options.Content != "" || len(options.Embeds) != 0 || len(options.Components) != 0 || len(options.Files) != 0 || len(options.Attachments) != 0 || len(options.StickerIDs) != 0 || options.ReplyTo != "" {
		return model.Message{}, errors.New("a forward carries only the forwarded message")
	}
	if len(forward.AttachmentIDs) > model.MaxAttachments || len(forward.EmbedIndices) > model.MaxEmbeds {
		return model.Message{}, fmt.Errorf("a forward picks at most %d attachments and %d embeds", model.MaxAttachments, model.MaxEmbeds)
	}
	reference := map[string]any{"type": model.MessageReferenceForward, "channel_id": forward.ChannelID, "message_id": forward.MessageID}
	if forward.GuildID != "" {
		reference["guild_id"] = forward.GuildID
	}
	if len(forward.AttachmentIDs) != 0 {
		reference["attachment_ids"] = forward.AttachmentIDs
	}
	if len(forward.EmbedIndices) != 0 {
		reference["embed_indices"] = forward.EmbedIndices
	}
	body := map[string]any{"message_reference": reference}
	if options.Nonce != "" {
		body["nonce"] = options.Nonce
	}
	if options.Flags != 0 {
		body["flags"] = options.Flags
	}
	var message model.Message
	err := c.call(ctx, &request{method: http.MethodPost, path: path("channels", url.PathEscape(channelID), "messages"), body: body, dest: &message, idempotent: options.Nonce != ""})
	return message, err
}

func (c *Client) ForwardMessage(ctx context.Context, channelID string, forward model.MessageForward) (model.Message, error) {
	return c.SendMessageWithOptions(ctx, channelID, model.MessageOptions{Forward: &forward})
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
	attachments, changed, err := editAttachments(edit)
	if err != nil {
		return model.Message{}, err
	}
	if changed {
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
		if err := validateEmbeds(*edit.Embeds, kind); err != nil {
			return nil, err
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
	if err := validateEmbeds(embeds, kind); err != nil {
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

func validateEmbeds(embeds []model.Embed, kind string) error {
	for i, embed := range embeds {
		if err := embed.Validate(); err != nil {
			return fmt.Errorf("%s embed %d: %w", kind, i, err)
		}
	}
	return nil
}

func attachmentMetadata(files []model.File, uploads []model.UploadedAttachment) []map[string]any {
	var items []map[string]any
	for i, file := range files {
		item := map[string]any{"id": i, "filename": file.Name}
		describeAttachment(item, file.Description, file.Title, file.Spoiler)
		items = append(items, item)
	}
	for i, upload := range uploads {
		item := map[string]any{"id": len(files) + i, "filename": upload.Filename, "upload_filename": upload.UploadFilename, "file_size": upload.Size, "content_type": contentTypeOr(upload.ContentType)}
		describeAttachment(item, upload.Description, upload.Title, upload.Spoiler)
		if upload.Duration != 0 {
			item["duration"] = upload.Duration
		}
		if upload.Waveform != "" {
			item["waveform"] = upload.Waveform
		}
		items = append(items, item)
	}
	return items
}

func describeAttachment(item map[string]any, description, title string, spoiler bool) {
	if description != "" {
		item["description"] = description
	}
	if title != "" {
		item["title"] = title
	}
	if spoiler {
		item["flags"] = model.AttachmentFlagSpoiler
	}
}

func keptAttachments(kept []model.KeptAttachment) ([]map[string]any, error) {
	items := make([]map[string]any, 0, len(kept))
	for _, attachment := range kept {
		if attachment.ID == "" {
			return nil, errors.New("a kept attachment needs its ID")
		}
		item := map[string]any{"id": attachment.ID}
		attachment.Title.Apply(item, "title")
		attachment.Description.Apply(item, "description")
		items = append(items, item)
	}
	return items, nil
}

func editAttachments(edit model.MessageEdit) ([]map[string]any, bool, error) {
	if edit.KeepAttachments == nil && len(edit.Files) == 0 && len(edit.Attachments) == 0 {
		return nil, false, nil
	}
	kept, err := keptAttachments(edit.KeepAttachments)
	if err != nil {
		return nil, false, err
	}
	if len(kept)+len(edit.Files)+len(edit.Attachments) > model.MaxAttachments {
		return nil, false, fmt.Errorf("a message can contain at most %d attachments", model.MaxAttachments)
	}
	return append(kept, attachmentMetadata(edit.Files, edit.Attachments)...), true, nil
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

func (c *Client) ReactionUsers(ctx context.Context, channelID, messageID, emoji string, query model.ReactionUsersQuery) (model.ReactionUsersPage, error) {
	if channelID == "" || messageID == "" {
		return model.ReactionUsersPage{}, errors.New("channel ID and message ID are required")
	}
	if emoji == "" || len(emoji) > 64 {
		return model.ReactionUsersPage{}, errors.New("emoji must be 1 to 64 characters")
	}
	if query.Limit < 0 || query.Limit > model.MaxReactionUsersPerPage {
		return model.ReactionUsersPage{}, fmt.Errorf("reaction user limit must be between 1 and %d", model.MaxReactionUsersPerPage)
	}
	values := url.Values{}
	setQuery(values, "limit", limitString(query.Limit))
	setQuery(values, "after", query.After)
	var page model.ReactionUsersPage
	err := c.call(ctx, &request{method: http.MethodGet, path: messagePath(channelID, messageID) + "/reactions/" + url.PathEscape(emoji) + "/users", query: values, dest: &page})
	return page, err
}

func (c *Client) CrosspostMessage(ctx context.Context, channelID, messageID string) (model.Message, error) {
	if channelID == "" || messageID == "" {
		return model.Message{}, errors.New("channel ID and message ID are required")
	}
	var message model.Message
	err := c.call(ctx, &request{method: http.MethodPost, path: messagePath(channelID, messageID) + "/crosspost", dest: &message})
	return message, err
}

func (c *Client) CrosspostSource(ctx context.Context, channelID, messageID string) (model.CrosspostSource, error) {
	if channelID == "" || messageID == "" {
		return model.CrosspostSource{}, errors.New("channel ID and message ID are required")
	}
	var source model.CrosspostSource
	err := c.call(ctx, &request{method: http.MethodGet, path: messagePath(channelID, messageID) + "/crosspost-source", dest: &source})
	return source, err
}

func (c *Client) DeleteAttachment(ctx context.Context, channelID, messageID, attachmentID string) error {
	if channelID == "" || messageID == "" || attachmentID == "" {
		return errors.New("channel ID, message ID and attachment ID are required")
	}
	return c.call(ctx, &request{method: http.MethodDelete, path: messagePath(channelID, messageID) + "/attachments/" + url.PathEscape(attachmentID)})
}

func (c *Client) RefreshAttachmentURLs(ctx context.Context, urls []string) ([]model.RefreshedURL, error) {
	if len(urls) == 0 || len(urls) > model.MaxRefreshURLs {
		return nil, fmt.Errorf("a refresh takes 1 to %d attachment URLs", model.MaxRefreshURLs)
	}
	var result struct {
		URLs []model.RefreshedURL `json:"refreshed_urls"`
	}
	err := c.call(ctx, &request{method: http.MethodPost, path: "/attachments/refresh-urls", body: map[string]any{"attachment_urls": urls}, dest: &result, idempotent: true})
	return result.URLs, err
}

func (c *Client) SuppressEmbeds(ctx context.Context, channelID, messageID string, suppress bool) (model.Message, error) {
	current, err := c.Message(ctx, channelID, messageID)
	if err != nil {
		return model.Message{}, err
	}
	flags := current.Flags & settableMessageFlags &^ model.MessageFlagSuppressEmbeds
	if suppress {
		flags |= model.MessageFlagSuppressEmbeds
	}
	return c.EditMessage(ctx, channelID, messageID, model.MessageEdit{Flags: &flags})
}

const settableMessageFlags = model.MessageFlagSuppressEmbeds | model.MessageFlagSuppressNotifications | model.MessageFlagVoiceMessage

func (c *Client) SearchMessages(ctx context.Context, search model.MessageSearch) (model.MessageSearchResult, error) {
	if search.GuildID == "" && search.ChannelID == "" {
		return model.MessageSearchResult{}, errors.New("a message search needs a guild ID or a channel ID")
	}
	if search.HitsPerPage < 0 || search.HitsPerPage > model.MaxSearchHitsPerPage || search.Page < 0 || search.Page > model.MaxSearchPage {
		return model.MessageSearchResult{}, fmt.Errorf("a message search takes 1 to %d hits per page and a page of 1 to %d", model.MaxSearchHitsPerPage, model.MaxSearchPage)
	}
	body := map[string]any{}
	for key, value := range map[string]string{"context_guild_id": search.GuildID, "context_channel_id": search.ChannelID, "content": search.Content, "min_id": search.MinID, "max_id": search.MaxID, "sort_by": search.SortBy, "sort_order": search.SortOrder} {
		if value != "" {
			body[key] = value
		}
	}
	for key, values := range map[string][]string{
		"channel_ids": search.ChannelIDs, "exclude_channel_id": search.ExcludeChannelIDs, "exact_phrases": search.ExactPhrases,
		"author_id": search.AuthorIDs, "exclude_author_id": search.ExcludeAuthorIDs, "author_type": search.AuthorTypes,
		"mentions": search.Mentions, "has": search.Has, "exclude_has": search.ExcludeHas, "link_hostname": search.LinkHostnames,
		"attachment_filename": search.AttachmentFilename, "attachment_extension": search.AttachmentExt,
	} {
		if len(values) != 0 {
			body[key] = values
		}
	}
	if search.MentionEveryone != nil {
		body["mention_everyone"] = *search.MentionEveryone
	}
	if search.Pinned != nil {
		body["pinned"] = *search.Pinned
	}
	if search.HitsPerPage != 0 {
		body["hits_per_page"] = search.HitsPerPage
	}
	if search.Page != 0 {
		body["page"] = search.Page
	}
	if search.IncludeNSFW {
		body["include_nsfw"] = true
	}
	var result model.MessageSearchResult
	err := c.call(ctx, &request{method: http.MethodPost, path: "/search/messages", body: body, dest: &result, idempotent: true})
	return result, err
}
