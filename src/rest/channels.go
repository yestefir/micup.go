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

func (c *Client) Channel(ctx context.Context, channelID string) (model.Channel, error) {
	if channelID == "" {
		return model.Channel{}, errors.New("channel ID is required")
	}
	var channel model.Channel
	err := c.Do(ctx, http.MethodGet, path("channels", url.PathEscape(channelID)), nil, &channel)
	return channel, err
}

func (c *Client) CreateChannel(ctx context.Context, guildID string, create model.ChannelCreate, options ...RequestOption) (model.Channel, error) {
	if guildID == "" {
		return model.Channel{}, errors.New("guild ID is required")
	}
	if n := utf8.RuneCountInString(create.Name); n == 0 || n > model.MaxChannelName {
		return model.Channel{}, fmt.Errorf("channel name must be 1 to %d characters", model.MaxChannelName)
	}
	if create.Options.Name.IsSet() || create.Options.Type.IsSet() {
		return model.Channel{}, errors.New("set the name and type of a new channel on ChannelCreate, not in Options")
	}
	if create.Options.RTCRegion.IsSet() || create.Options.RTCP2P.IsSet() {
		return model.Channel{}, errors.New("a voice region and peer-to-peer voice are set with ModifyChannel")
	}
	body, err := channelBody(create.Options)
	if err != nil {
		return model.Channel{}, err
	}
	body["type"] = create.Type
	body["name"] = create.Name
	var channel model.Channel
	err = c.call(ctx, apply(&request{method: http.MethodPost, path: path("guilds", url.PathEscape(guildID), "channels"), body: body, dest: &channel}, options))
	return channel, err
}

func (c *Client) ModifyChannel(ctx context.Context, channelID string, edit model.ChannelEdit, options ...RequestOption) (model.Channel, error) {
	if channelID == "" {
		return model.Channel{}, errors.New("channel ID is required")
	}
	body, err := channelBody(edit)
	if err != nil {
		return model.Channel{}, err
	}
	if len(body) == 0 {
		return model.Channel{}, errors.New("channel edit has no changes")
	}
	var channel model.Channel
	err = c.call(ctx, apply(&request{method: http.MethodPatch, path: path("channels", url.PathEscape(channelID)), body: body, dest: &channel}, options))
	return channel, err
}

func (c *Client) DeleteChannel(ctx context.Context, channelID string, options ...RequestOption) error {
	if channelID == "" {
		return errors.New("channel ID is required")
	}
	return c.call(ctx, apply(&request{method: http.MethodDelete, path: path("channels", url.PathEscape(channelID))}, options))
}

func (c *Client) ModifyChannelPositions(ctx context.Context, guildID string, positions []model.ChannelPosition) error {
	if guildID == "" {
		return errors.New("guild ID is required")
	}
	if len(positions) == 0 || len(positions) > model.MaxChannelPositions {
		return fmt.Errorf("channel positions take 1 to %d entries", model.MaxChannelPositions)
	}
	body := make([]map[string]any, len(positions))
	for i, position := range positions {
		if position.ID == "" {
			return fmt.Errorf("channel position %d has no channel ID", i)
		}
		if _, ok := position.Position.Value(); position.Position.IsSet() && !ok {
			return fmt.Errorf("channel position %d cannot clear the position", i)
		}
		entry := map[string]any{"id": position.ID}
		position.Position.Apply(entry, "position")
		position.ParentID.Apply(entry, "parent_id")
		position.PrecedingSiblingID.Apply(entry, "preceding_sibling_id")
		if position.LockPermissions {
			entry["lock_permissions"] = true
		}
		body[i] = entry
	}
	return c.call(ctx, &request{method: http.MethodPatch, path: path("guilds", url.PathEscape(guildID), "channels"), body: body})
}

func (c *Client) SetChannelPermission(ctx context.Context, channelID string, overwrite model.PermissionOverwrite, options ...RequestOption) error {
	if channelID == "" || overwrite.ID == "" {
		return errors.New("channel ID and the overwrite's role or member ID are required")
	}
	if overwrite.Type != model.OverwriteRole && overwrite.Type != model.OverwriteMember {
		return fmt.Errorf("unknown permission overwrite type %d", overwrite.Type)
	}
	body := map[string]any{"type": overwrite.Type, "allow": overwrite.Allow, "deny": overwrite.Deny}
	return c.call(ctx, apply(&request{method: http.MethodPut, path: path("channels", url.PathEscape(channelID), "permissions", url.PathEscape(overwrite.ID)), body: body}, options))
}

func (c *Client) DeleteChannelPermission(ctx context.Context, channelID, overwriteID string, options ...RequestOption) error {
	if channelID == "" || overwriteID == "" {
		return errors.New("channel ID and the overwrite's role or member ID are required")
	}
	return c.call(ctx, apply(&request{method: http.MethodDelete, path: path("channels", url.PathEscape(channelID), "permissions", url.PathEscape(overwriteID))}, options))
}

func (c *Client) FollowChannel(ctx context.Context, announcementChannelID, targetChannelID string, options ...RequestOption) (model.FollowedChannel, error) {
	if announcementChannelID == "" || targetChannelID == "" {
		return model.FollowedChannel{}, errors.New("the announcement channel ID and the target channel ID are required")
	}
	var followed model.FollowedChannel
	err := c.call(ctx, apply(&request{method: http.MethodPost, path: path("channels", url.PathEscape(announcementChannelID), "followers"), body: map[string]string{"webhook_channel_id": targetChannelID}, dest: &followed}, options))
	return followed, err
}

func (c *Client) ChannelFollowerStats(ctx context.Context, channelID string) (model.ChannelFollowerStats, error) {
	if channelID == "" {
		return model.ChannelFollowerStats{}, errors.New("channel ID is required")
	}
	var stats model.ChannelFollowerStats
	err := c.call(ctx, &request{method: http.MethodGet, path: path("channels", url.PathEscape(channelID), "follower-stats"), dest: &stats})
	return stats, err
}

func channelBody(edit model.ChannelEdit) (map[string]any, error) {
	if name, ok := edit.Name.Value(); ok {
		if n := utf8.RuneCountInString(name); n == 0 || n > model.MaxChannelName {
			return nil, fmt.Errorf("channel name must be 1 to %d characters", model.MaxChannelName)
		}
	}
	if topic, ok := edit.Topic.Value(); ok && utf8.RuneCountInString(topic) > model.MaxForumTopic {
		return nil, fmt.Errorf("channel topic is longer than %d characters", model.MaxForumTopic)
	}
	for _, seconds := range []model.Optional[int]{edit.RateLimitPerUser, edit.DefaultThreadRateLimitPerUser} {
		if value, ok := seconds.Value(); ok && (value < 0 || value > model.MaxSlowmodeSeconds) {
			return nil, fmt.Errorf("slowmode must be 0 to %d seconds", model.MaxSlowmodeSeconds)
		}
	}
	if tags, ok := edit.AvailableTags.Value(); ok {
		if err := validateForumTags(tags); err != nil {
			return nil, err
		}
	}
	body := map[string]any{}
	edit.Name.Apply(body, "name")
	edit.Type.Apply(body, "type")
	edit.Topic.Apply(body, "topic")
	edit.URL.Apply(body, "url")
	edit.ParentID.Apply(body, "parent_id")
	edit.Bitrate.Apply(body, "bitrate")
	edit.UserLimit.Apply(body, "user_limit")
	edit.VoiceConnectionLimit.Apply(body, "voice_connection_limit")
	if overwrites, ok := edit.PermissionOverwrites.Value(); ok {
		if overwrites == nil {
			overwrites = []model.PermissionOverwrite{}
		}
		body["permission_overwrites"] = overwrites
	} else if edit.PermissionOverwrites.IsNull() {
		body["permission_overwrites"] = []model.PermissionOverwrite{}
	}
	edit.NSFW.Apply(body, "nsfw_override")
	edit.ContentWarningLevel.Apply(body, "content_warning_level")
	edit.ContentWarningText.Apply(body, "content_warning_text")
	edit.RateLimitPerUser.Apply(body, "rate_limit_per_user")
	edit.RTCRegion.Apply(body, "rtc_region")
	edit.RTCP2P.Apply(body, "rtc_p2p")
	edit.DefaultAutoArchiveDuration.Apply(body, "default_auto_archive_duration")
	edit.DefaultThreadRateLimitPerUser.Apply(body, "default_thread_rate_limit_per_user")
	if tags, ok := edit.AvailableTags.Value(); ok {
		if tags == nil {
			tags = []model.ForumTag{}
		}
		body["available_tags"] = tags
	} else if edit.AvailableTags.IsNull() {
		body["available_tags"] = []model.ForumTag{}
	}
	edit.DefaultReactionEmoji.Apply(body, "default_reaction_emoji")
	edit.DefaultSortOrder.Apply(body, "default_sort_order")
	edit.DefaultForumLayout.Apply(body, "default_forum_layout")
	edit.DefaultTagSetting.Apply(body, "default_tag_setting")
	edit.Flags.Apply(body, "flags")
	return body, nil
}

func validateForumTags(tags []model.ForumTag) error {
	if len(tags) > model.MaxForumTags {
		return fmt.Errorf("a forum can have at most %d tags", model.MaxForumTags)
	}
	names := map[string]bool{}
	for i, tag := range tags {
		if err := validateForumTag(tag); err != nil {
			return fmt.Errorf("forum tag %d: %w", i, err)
		}
		if names[tag.Name] {
			return fmt.Errorf("forum tag name %q is used twice", tag.Name)
		}
		names[tag.Name] = true
	}
	return nil
}

func validateForumTag(tag model.ForumTag) error {
	if n := utf8.RuneCountInString(tag.Name); n == 0 || n > model.MaxForumTagName {
		return fmt.Errorf("name must be 1 to %d characters", model.MaxForumTagName)
	}
	if tag.EmojiID != nil && tag.EmojiName != nil {
		return errors.New("set a custom emoji ID or a Unicode emoji, not both")
	}
	return nil
}
