package rest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
	"unicode/utf8"

	"github.com/yestefir/micup.go/src/model"
)

const (
	maxMembersPerPage = 1000
	maxMemberSearch   = 100
	maxNickname       = 32
	maxTimeout        = 365 * 24 * time.Hour
)

func (c *Client) Member(ctx context.Context, guildID, userID string) (model.Member, error) {
	if guildID == "" || userID == "" {
		return model.Member{}, errors.New("guild ID and user ID are required")
	}
	var member model.Member
	err := c.call(ctx, &request{method: http.MethodGet, path: memberPath(guildID, userID), dest: &member})
	return member, err
}

func (c *Client) Members(ctx context.Context, guildID string, query model.MemberQuery) ([]model.Member, error) {
	if guildID == "" {
		return nil, errors.New("guild ID is required")
	}
	if query.Limit < 0 || query.Limit > maxMembersPerPage {
		return nil, errors.New("member limit must be between 1 and 1000")
	}
	values := url.Values{}
	setQuery(values, "limit", limitString(query.Limit))
	setQuery(values, "after", query.After)
	var members []model.Member
	err := c.call(ctx, &request{method: http.MethodGet, path: path("guilds", url.PathEscape(guildID), "members"), query: values, dest: &members})
	return members, err
}

func (c *Client) ModifyMember(ctx context.Context, guildID, userID string, edit model.MemberEdit, options ...RequestOption) (model.Member, error) {
	if guildID == "" || userID == "" {
		return model.Member{}, errors.New("guild ID and user ID are required")
	}
	body := map[string]any{}
	if edit.Nick != nil {
		body["nick"] = *edit.Nick
	}
	if edit.Roles != nil {
		body["roles"] = *edit.Roles
	}
	if edit.Mute != nil {
		body["mute"] = *edit.Mute
	}
	if edit.Deaf != nil {
		body["deaf"] = *edit.Deaf
	}
	if edit.CommunicationDisabledUntil != nil {
		if *edit.CommunicationDisabledUntil == "" {
			body["communication_disabled_until"] = nil
		} else {
			body["communication_disabled_until"] = *edit.CommunicationDisabledUntil
		}
	}
	switch {
	case edit.Disconnect && edit.ChannelID != nil:
		return model.Member{}, errors.New("set either ChannelID or Disconnect, not both")
	case edit.Disconnect:
		body["channel_id"] = nil
	case edit.ChannelID != nil:
		body["channel_id"] = *edit.ChannelID
	}
	if edit.ConnectionID != "" {
		body["connection_id"] = edit.ConnectionID
	}
	if edit.TimeoutReason != "" {
		if edit.CommunicationDisabledUntil == nil || *edit.CommunicationDisabledUntil == "" {
			return model.Member{}, errors.New("a timeout reason needs CommunicationDisabledUntil")
		}
		if utf8.RuneCountInString(edit.TimeoutReason) > maxAuditReason {
			return model.Member{}, fmt.Errorf("timeout reason is longer than %d characters", maxAuditReason)
		}
		body["timeout_reason"] = edit.TimeoutReason
	}
	if len(body) == 0 {
		return model.Member{}, errors.New("member edit has no changes")
	}
	var member model.Member
	err := c.call(ctx, apply(&request{method: http.MethodPatch, path: memberPath(guildID, userID), body: body, dest: &member}, options))
	return member, err
}

func (c *Client) AddMemberRole(ctx context.Context, guildID, userID, roleID string, options ...RequestOption) error {
	return c.memberRole(ctx, http.MethodPut, guildID, userID, roleID, options)
}

func (c *Client) RemoveMemberRole(ctx context.Context, guildID, userID, roleID string, options ...RequestOption) error {
	return c.memberRole(ctx, http.MethodDelete, guildID, userID, roleID, options)
}

func (c *Client) KickMember(ctx context.Context, guildID, userID string, options ...RequestOption) error {
	if guildID == "" || userID == "" {
		return errors.New("guild ID and user ID are required")
	}
	return c.call(ctx, apply(&request{method: http.MethodDelete, path: memberPath(guildID, userID)}, options))
}

func (c *Client) BanMember(ctx context.Context, guildID, userID string, ban model.Ban, options ...RequestOption) error {
	if guildID == "" || userID == "" {
		return errors.New("guild ID and user ID are required")
	}
	if ban.DeleteMessageSeconds < 0 || ban.DeleteMessageSeconds > 604800 {
		return errors.New("ban message deletion must be 0 to 604800 seconds")
	}
	if ban.DurationSeconds != 0 && (ban.DurationSeconds < 60 || ban.DurationSeconds > 63072000) {
		return errors.New("ban duration must be 0 or 60 to 63072000 seconds")
	}
	body := map[string]any{}
	if ban.DeleteMessageSeconds != 0 {
		body["delete_message_seconds"] = ban.DeleteMessageSeconds
	}
	if ban.DurationSeconds != 0 {
		body["ban_duration_seconds"] = ban.DurationSeconds
	}
	if ban.Reason != "" {
		body["reason"] = ban.Reason
	}
	return c.call(ctx, apply(&request{method: http.MethodPut, path: path("guilds", url.PathEscape(guildID), "bans", url.PathEscape(userID)), body: body}, options))
}

func (c *Client) UnbanMember(ctx context.Context, guildID, userID string, options ...RequestOption) error {
	if guildID == "" || userID == "" {
		return errors.New("guild ID and user ID are required")
	}
	return c.call(ctx, apply(&request{method: http.MethodDelete, path: path("guilds", url.PathEscape(guildID), "bans", url.PathEscape(userID))}, options))
}

func (c *Client) memberRole(ctx context.Context, method, guildID, userID, roleID string, options []RequestOption) error {
	if guildID == "" || userID == "" || roleID == "" {
		return errors.New("guild ID, user ID and role ID are required")
	}
	return c.call(ctx, apply(&request{method: method, path: memberPath(guildID, userID) + "/roles/" + url.PathEscape(roleID)}, options))
}

func memberPath(guildID, userID string) string {
	return path("guilds", url.PathEscape(guildID), "members", url.PathEscape(userID))
}

func (c *Client) CurrentMember(ctx context.Context, guildID string) (model.Member, error) {
	if guildID == "" {
		return model.Member{}, errors.New("guild ID is required")
	}
	var member model.Member
	err := c.call(ctx, &request{method: http.MethodGet, path: path("guilds", url.PathEscape(guildID), "members", "@me"), dest: &member})
	return member, err
}

func (c *Client) ModifyCurrentMember(ctx context.Context, guildID string, edit model.CurrentMemberEdit, options ...RequestOption) (model.Member, error) {
	if guildID == "" {
		return model.Member{}, errors.New("guild ID is required")
	}
	if nick, ok := edit.Nick.Value(); ok && utf8.RuneCountInString(nick) > maxNickname {
		return model.Member{}, fmt.Errorf("nickname is longer than %d characters", maxNickname)
	}
	body := map[string]any{}
	edit.Nick.Apply(body, "nick")
	edit.Avatar.Apply(body, "avatar")
	edit.Banner.Apply(body, "banner")
	edit.Bio.Apply(body, "bio")
	edit.Pronouns.Apply(body, "pronouns")
	edit.AccentColor.Apply(body, "accent_color")
	edit.ProfileFlags.Apply(body, "profile_flags")
	edit.MentionFlags.Apply(body, "mention_flags")
	if len(body) == 0 {
		return model.Member{}, errors.New("member edit has no changes")
	}
	var member model.Member
	err := c.call(ctx, apply(&request{method: http.MethodPatch, path: path("guilds", url.PathEscape(guildID), "members", "@me"), body: body, dest: &member}, options))
	return member, err
}

func (c *Client) SetNickname(ctx context.Context, guildID, nick string, options ...RequestOption) (model.Member, error) {
	return c.ModifyCurrentMember(ctx, guildID, model.CurrentMemberEdit{Nick: model.Set(nick)}, options...)
}

func (c *Client) TimeoutMember(ctx context.Context, guildID, userID string, duration time.Duration, reason string, options ...RequestOption) (model.Member, error) {
	if duration <= 0 || duration > maxTimeout {
		return model.Member{}, errors.New("a timeout lasts more than 0 and at most 365 days")
	}
	until := time.Now().Add(duration).UTC().Format(time.RFC3339)
	return c.ModifyMember(ctx, guildID, userID, model.MemberEdit{CommunicationDisabledUntil: &until, TimeoutReason: reason}, options...)
}

func (c *Client) ClearTimeout(ctx context.Context, guildID, userID string, options ...RequestOption) (model.Member, error) {
	none := ""
	return c.ModifyMember(ctx, guildID, userID, model.MemberEdit{CommunicationDisabledUntil: &none}, options...)
}

func (c *Client) SearchMembers(ctx context.Context, guildID string, search model.MemberSearch) (model.MemberSearchPage, error) {
	if guildID == "" {
		return model.MemberSearchPage{}, errors.New("guild ID is required")
	}
	if search.Limit < 0 || search.Limit > maxMemberSearch || search.Offset < 0 {
		return model.MemberSearchPage{}, fmt.Errorf("member search takes a limit of 1 to %d and a non-negative offset", maxMemberSearch)
	}
	if utf8.RuneCountInString(search.Query) > maxMemberSearch || len(search.RoleIDs) > 10 || len(search.JoinSourceTypes) > 10 || len(search.SourceInviteCodes) > 10 {
		return model.MemberSearchPage{}, errors.New("member search takes a query of at most 100 characters and at most 10 values per filter")
	}
	body := map[string]any{}
	if search.Query != "" {
		body["query"] = search.Query
	}
	if search.Limit != 0 {
		body["limit"] = search.Limit
	}
	if search.Offset != 0 {
		body["offset"] = search.Offset
	}
	if len(search.RoleIDs) != 0 {
		body["role_ids"] = search.RoleIDs
	}
	if len(search.JoinSourceTypes) != 0 {
		body["join_source_type"] = search.JoinSourceTypes
	}
	if len(search.SourceInviteCodes) != 0 {
		body["source_invite_code"] = search.SourceInviteCodes
	}
	if search.IsBot != nil {
		body["is_bot"] = *search.IsBot
	}
	for key, value := range map[string]int64{"joined_at_gte": search.JoinedAtGTE, "joined_at_lte": search.JoinedAtLTE, "user_created_at_gte": search.UserCreatedAtGTE, "user_created_at_lte": search.UserCreatedAtLTE} {
		if value != 0 {
			body[key] = value
		}
	}
	if search.SortBy != "" {
		body["sort_by"] = search.SortBy
	}
	if search.SortOrder != "" {
		body["sort_order"] = search.SortOrder
	}
	var page model.MemberSearchPage
	err := c.call(ctx, &request{method: http.MethodPost, path: path("guilds", url.PathEscape(guildID), "members-search"), body: body, dest: &page, idempotent: true})
	return page, err
}
