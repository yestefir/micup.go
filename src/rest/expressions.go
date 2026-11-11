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

const (
	maxStickerDescription = 500
	maxStickerTags        = 10
	maxStickerTag         = 30
)

func (c *Client) Emojis(ctx context.Context, guildID string) ([]model.Emoji, error) {
	if guildID == "" {
		return nil, errors.New("guild ID is required")
	}
	var emojis []model.Emoji
	err := c.call(ctx, &request{method: http.MethodGet, path: path("guilds", url.PathEscape(guildID), "emojis"), dest: &emojis})
	return emojis, err
}

func (c *Client) CreateEmoji(ctx context.Context, guildID string, create model.EmojiCreate, options ...RequestOption) (model.Emoji, error) {
	if guildID == "" {
		return model.Emoji{}, errors.New("guild ID is required")
	}
	if err := validateEmoji(create); err != nil {
		return model.Emoji{}, err
	}
	var emoji model.Emoji
	err := c.call(ctx, apply(&request{method: http.MethodPost, path: path("guilds", url.PathEscape(guildID), "emojis"), body: create, dest: &emoji}, options))
	return emoji, err
}

func (c *Client) CreateEmojis(ctx context.Context, guildID string, emojis []model.EmojiCreate, options ...RequestOption) (model.EmojiBulkResult, error) {
	if guildID == "" {
		return model.EmojiBulkResult{}, errors.New("guild ID is required")
	}
	if len(emojis) == 0 || len(emojis) > model.MaxBulkExpressions {
		return model.EmojiBulkResult{}, fmt.Errorf("a bulk create takes 1 to %d emojis", model.MaxBulkExpressions)
	}
	for i, emoji := range emojis {
		if err := validateEmoji(emoji); err != nil {
			return model.EmojiBulkResult{}, fmt.Errorf("emoji %d: %w", i, err)
		}
	}
	var result model.EmojiBulkResult
	err := c.call(ctx, apply(&request{method: http.MethodPost, path: path("guilds", url.PathEscape(guildID), "emojis", "bulk"), body: map[string]any{"emojis": emojis}, dest: &result}, options))
	return result, err
}

func (c *Client) CloneEmoji(ctx context.Context, guildID, sourceEmojiID string, options ...RequestOption) (model.Emoji, error) {
	if guildID == "" || sourceEmojiID == "" {
		return model.Emoji{}, errors.New("guild ID and the source emoji ID are required")
	}
	var emoji model.Emoji
	err := c.call(ctx, apply(&request{method: http.MethodPost, path: path("guilds", url.PathEscape(guildID), "emojis", "clone"), body: map[string]string{"source_emoji_id": sourceEmojiID}, dest: &emoji}, options))
	return emoji, err
}

func (c *Client) ModifyEmoji(ctx context.Context, guildID, emojiID, name string, options ...RequestOption) (model.Emoji, error) {
	if guildID == "" || emojiID == "" {
		return model.Emoji{}, errors.New("guild ID and emoji ID are required")
	}
	if !validEmojiName(name) {
		return model.Emoji{}, fmt.Errorf("emoji name must be 2 to %d ASCII letters, digits or underscores", model.MaxEmojiName)
	}
	var emoji model.Emoji
	err := c.call(ctx, apply(&request{method: http.MethodPatch, path: path("guilds", url.PathEscape(guildID), "emojis", url.PathEscape(emojiID)), body: map[string]string{"name": name}, dest: &emoji}, options))
	return emoji, err
}

func (c *Client) DeleteEmoji(ctx context.Context, guildID, emojiID string, purge bool, options ...RequestOption) error {
	if guildID == "" || emojiID == "" {
		return errors.New("guild ID and emoji ID are required")
	}
	return c.call(ctx, apply(&request{method: http.MethodDelete, path: path("guilds", url.PathEscape(guildID), "emojis", url.PathEscape(emojiID)), query: purgeQuery(purge)}, options))
}

func (c *Client) EmojiMetadata(ctx context.Context, emojiID string) (model.ExpressionMetadata, error) {
	return c.expressionMetadata(ctx, "emojis", emojiID)
}

func (c *Client) Stickers(ctx context.Context, guildID string) ([]model.Sticker, error) {
	if guildID == "" {
		return nil, errors.New("guild ID is required")
	}
	var stickers []model.Sticker
	err := c.call(ctx, &request{method: http.MethodGet, path: path("guilds", url.PathEscape(guildID), "stickers"), dest: &stickers})
	return stickers, err
}

func (c *Client) CreateSticker(ctx context.Context, guildID string, create model.StickerCreate, options ...RequestOption) (model.Sticker, error) {
	if guildID == "" {
		return model.Sticker{}, errors.New("guild ID is required")
	}
	if create.Image == "" {
		return model.Sticker{}, errors.New("sticker image is required")
	}
	if err := validateSticker(create.Name, create.Description, create.Tags); err != nil {
		return model.Sticker{}, err
	}
	var sticker model.Sticker
	err := c.call(ctx, apply(&request{method: http.MethodPost, path: path("guilds", url.PathEscape(guildID), "stickers"), body: create, dest: &sticker}, options))
	return sticker, err
}

func (c *Client) CreateStickers(ctx context.Context, guildID string, stickers []model.StickerCreate, options ...RequestOption) (model.StickerBulkResult, error) {
	if guildID == "" {
		return model.StickerBulkResult{}, errors.New("guild ID is required")
	}
	if len(stickers) == 0 || len(stickers) > model.MaxBulkExpressions {
		return model.StickerBulkResult{}, fmt.Errorf("a bulk create takes 1 to %d stickers", model.MaxBulkExpressions)
	}
	for i, sticker := range stickers {
		if sticker.Image == "" {
			return model.StickerBulkResult{}, fmt.Errorf("sticker %d: image is required", i)
		}
		if err := validateSticker(sticker.Name, sticker.Description, sticker.Tags); err != nil {
			return model.StickerBulkResult{}, fmt.Errorf("sticker %d: %w", i, err)
		}
	}
	var result model.StickerBulkResult
	err := c.call(ctx, apply(&request{method: http.MethodPost, path: path("guilds", url.PathEscape(guildID), "stickers", "bulk"), body: map[string]any{"stickers": stickers}, dest: &result}, options))
	return result, err
}

func (c *Client) CloneSticker(ctx context.Context, guildID, sourceStickerID string, options ...RequestOption) (model.Sticker, error) {
	if guildID == "" || sourceStickerID == "" {
		return model.Sticker{}, errors.New("guild ID and the source sticker ID are required")
	}
	var sticker model.Sticker
	err := c.call(ctx, apply(&request{method: http.MethodPost, path: path("guilds", url.PathEscape(guildID), "stickers", "clone"), body: map[string]string{"source_sticker_id": sourceStickerID}, dest: &sticker}, options))
	return sticker, err
}

func (c *Client) ModifySticker(ctx context.Context, guildID, stickerID string, edit model.StickerEdit, options ...RequestOption) (model.Sticker, error) {
	if guildID == "" || stickerID == "" {
		return model.Sticker{}, errors.New("guild ID and sticker ID are required")
	}
	if err := validateSticker(edit.Name, edit.Description, edit.Tags); err != nil {
		return model.Sticker{}, err
	}
	body := map[string]any{"name": edit.Name, "tags": edit.Tags}
	if edit.Tags == nil {
		body["tags"] = []string{}
	}
	if edit.Description != "" {
		body["description"] = edit.Description
	}
	var sticker model.Sticker
	err := c.call(ctx, apply(&request{method: http.MethodPatch, path: path("guilds", url.PathEscape(guildID), "stickers", url.PathEscape(stickerID)), body: body, dest: &sticker}, options))
	return sticker, err
}

func (c *Client) DeleteSticker(ctx context.Context, guildID, stickerID string, purge bool, options ...RequestOption) error {
	if guildID == "" || stickerID == "" {
		return errors.New("guild ID and sticker ID are required")
	}
	return c.call(ctx, apply(&request{method: http.MethodDelete, path: path("guilds", url.PathEscape(guildID), "stickers", url.PathEscape(stickerID)), query: purgeQuery(purge)}, options))
}

func (c *Client) StickerMetadata(ctx context.Context, stickerID string) (model.ExpressionMetadata, error) {
	return c.expressionMetadata(ctx, "stickers", stickerID)
}

func (c *Client) expressionMetadata(ctx context.Context, kind, id string) (model.ExpressionMetadata, error) {
	if id == "" {
		return model.ExpressionMetadata{}, errors.New("expression ID is required")
	}
	var metadata model.ExpressionMetadata
	err := c.call(ctx, &request{method: http.MethodGet, path: path(kind, url.PathEscape(id), "metadata"), dest: &metadata})
	return metadata, err
}

func purgeQuery(purge bool) url.Values {
	if !purge {
		return nil
	}
	return url.Values{"purge": {"true"}}
}

func validateEmoji(create model.EmojiCreate) error {
	if !validEmojiName(create.Name) {
		return fmt.Errorf("emoji name must be 2 to %d ASCII letters, digits or underscores", model.MaxEmojiName)
	}
	if create.Image == "" {
		return errors.New("emoji image is required")
	}
	return nil
}

func validEmojiName(name string) bool {
	if len(name) < 2 || len(name) > model.MaxEmojiName {
		return false
	}
	for _, r := range name {
		if !(r == '_' || (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')) {
			return false
		}
	}
	return true
}

func validateSticker(name, description string, tags []string) error {
	if n := utf8.RuneCountInString(name); n < 2 || n > model.MaxStickerName {
		return fmt.Errorf("sticker name must be 2 to %d characters", model.MaxStickerName)
	}
	if utf8.RuneCountInString(description) > maxStickerDescription {
		return fmt.Errorf("sticker description is longer than %d characters", maxStickerDescription)
	}
	if len(tags) > maxStickerTags {
		return fmt.Errorf("a sticker has at most %d tags", maxStickerTags)
	}
	for _, tag := range tags {
		if n := utf8.RuneCountInString(tag); n == 0 || n > maxStickerTag {
			return fmt.Errorf("sticker tags must be 1 to %d characters", maxStickerTag)
		}
	}
	return nil
}
