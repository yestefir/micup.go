package rest

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/yestefir/micup.go/src/model"
)

const maxMembersPerPage = 1000

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
