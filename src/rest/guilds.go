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

const maxGuildName = 100

func (c *Client) Guild(ctx context.Context, guildID string) (model.Guild, error) {
	if guildID == "" {
		return model.Guild{}, errors.New("guild ID is required")
	}
	var guild model.Guild
	err := c.call(ctx, &request{method: http.MethodGet, path: path("guilds", url.PathEscape(guildID)), dest: &guild})
	return guild, err
}

func (c *Client) GuildChannels(ctx context.Context, guildID string) ([]model.Channel, error) {
	if guildID == "" {
		return nil, errors.New("guild ID is required")
	}
	var channels []model.Channel
	err := c.call(ctx, &request{method: http.MethodGet, path: path("guilds", url.PathEscape(guildID), "channels"), dest: &channels})
	return channels, err
}

func (c *Client) Roles(ctx context.Context, guildID string) ([]model.Role, error) {
	if guildID == "" {
		return nil, errors.New("guild ID is required")
	}
	var roles []model.Role
	err := c.call(ctx, &request{method: http.MethodGet, path: path("guilds", url.PathEscape(guildID), "roles"), dest: &roles})
	return roles, err
}

const maxGuildsPerPage = 200

func (c *Client) ModifyGuild(ctx context.Context, guildID string, edit model.GuildEdit, options ...RequestOption) (model.Guild, error) {
	if guildID == "" {
		return model.Guild{}, errors.New("guild ID is required")
	}
	if name, ok := edit.Name.Value(); edit.Name.IsSet() && (!ok || name == "" || utf8.RuneCountInString(name) > maxGuildName) {
		return model.Guild{}, fmt.Errorf("guild name must be 1 to %d characters", maxGuildName)
	}
	if timeout, ok := edit.AFKTimeout.Value(); ok && (timeout < 60 || timeout > 3600) {
		return model.Guild{}, errors.New("AFK timeout must be 60 to 3600 seconds")
	}
	body := map[string]any{}
	edit.Name.Apply(body, "name")
	edit.Icon.Apply(body, "icon")
	edit.Banner.Apply(body, "banner")
	edit.Splash.Apply(body, "splash")
	edit.EmbedSplash.Apply(body, "embed_splash")
	edit.SplashCardAlignment.Apply(body, "splash_card_alignment")
	edit.SystemChannelID.Apply(body, "system_channel_id")
	edit.SystemChannelFlags.Apply(body, "system_channel_flags")
	edit.AFKChannelID.Apply(body, "afk_channel_id")
	edit.AFKTimeout.Apply(body, "afk_timeout")
	edit.DefaultMessageNotifications.Apply(body, "default_message_notifications")
	edit.VerificationLevel.Apply(body, "verification_level")
	edit.NSFW.Apply(body, "nsfw")
	edit.ContentWarningLevel.Apply(body, "content_warning_level")
	edit.ContentWarningText.Apply(body, "content_warning_text")
	edit.ExplicitContentFilter.Apply(body, "explicit_content_filter")
	edit.Features.Apply(body, "features")
	edit.MessageHistoryCutoff.Apply(body, "message_history_cutoff")
	if len(body) == 0 {
		return model.Guild{}, errors.New("guild edit has no changes")
	}
	var guild model.Guild
	err := c.call(ctx, apply(&request{method: http.MethodPatch, path: path("guilds", url.PathEscape(guildID)), body: body, dest: &guild}, options))
	return guild, err
}

func (c *Client) CurrentUserGuilds(ctx context.Context, query model.GuildQuery) ([]model.Guild, error) {
	if query.Limit < 0 || query.Limit > maxGuildsPerPage {
		return nil, fmt.Errorf("guild limit must be between 1 and %d", maxGuildsPerPage)
	}
	values := url.Values{}
	setQuery(values, "limit", limitString(query.Limit))
	setQuery(values, "before", query.Before)
	setQuery(values, "after", query.After)
	if query.WithCounts {
		values.Set("with_counts", "true")
	}
	var guilds []model.Guild
	err := c.call(ctx, &request{method: http.MethodGet, path: "/users/@me/guilds", query: values, dest: &guilds})
	return guilds, err
}

func (c *Client) LeaveGuild(ctx context.Context, guildID string) error {
	if guildID == "" {
		return errors.New("guild ID is required")
	}
	return c.call(ctx, &request{method: http.MethodDelete, path: path("users", "@me", "guilds", url.PathEscape(guildID))})
}

func (c *Client) AuditLog(ctx context.Context, guildID string, query model.AuditLogQuery) (model.AuditLog, error) {
	if guildID == "" {
		return model.AuditLog{}, errors.New("guild ID is required")
	}
	if query.Limit < 0 || query.Limit > model.MaxAuditLogEntries {
		return model.AuditLog{}, fmt.Errorf("audit log limit must be between 1 and %d", model.MaxAuditLogEntries)
	}
	if query.Before != "" && query.After != "" {
		return model.AuditLog{}, errors.New("set either Before or After, not both")
	}
	values := url.Values{}
	setQuery(values, "limit", limitString(query.Limit))
	setQuery(values, "before", query.Before)
	setQuery(values, "after", query.After)
	setQuery(values, "user_id", query.UserID)
	setQuery(values, "action_type", limitString(query.ActionType))
	var log model.AuditLog
	err := c.call(ctx, &request{method: http.MethodGet, path: path("guilds", url.PathEscape(guildID), "audit-logs"), query: values, dest: &log})
	return log, err
}

func (c *Client) Bans(ctx context.Context, guildID string) ([]model.BanEntry, error) {
	if guildID == "" {
		return nil, errors.New("guild ID is required")
	}
	var bans []model.BanEntry
	err := c.call(ctx, &request{method: http.MethodGet, path: path("guilds", url.PathEscape(guildID), "bans"), dest: &bans})
	return bans, err
}

func (c *Client) VanityURL(ctx context.Context, guildID string) (model.VanityURL, error) {
	if guildID == "" {
		return model.VanityURL{}, errors.New("guild ID is required")
	}
	var vanity model.VanityURL
	err := c.call(ctx, &request{method: http.MethodGet, path: path("guilds", url.PathEscape(guildID), "vanity-url"), dest: &vanity})
	return vanity, err
}

func (c *Client) ModifyVanityURL(ctx context.Context, guildID, code string, options ...RequestOption) (model.VanityURL, error) {
	if guildID == "" {
		return model.VanityURL{}, errors.New("guild ID is required")
	}
	body := map[string]any{"code": nil}
	if code != "" {
		body["code"] = code
	}
	var vanity model.VanityURL
	err := c.call(ctx, apply(&request{method: http.MethodPatch, path: path("guilds", url.PathEscape(guildID), "vanity-url"), body: body, dest: &vanity}, options))
	return vanity, err
}

func (c *Client) GatewayInfo(ctx context.Context) (model.GatewayInfo, error) {
	var info model.GatewayInfo
	err := c.call(ctx, &request{method: http.MethodGet, path: "/gateway/bot", dest: &info})
	return info, err
}
