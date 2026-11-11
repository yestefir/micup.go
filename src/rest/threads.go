package rest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"unicode/utf8"

	"github.com/yestefir/micup.go/src/model"
)

const (
	maxThreadsPerPage       = 100
	maxThreadSearchResults  = 25
	maxThreadMembersPerPage = 100
	maxPostDataThreads      = 100
)

func (c *Client) StartThread(ctx context.Context, channelID string, start model.ThreadStart, options ...RequestOption) (model.Channel, error) {
	if channelID == "" {
		return model.Channel{}, errors.New("channel ID is required")
	}
	switch start.Type {
	case 0:
		start.Type = model.ChannelTypePublicThread
	case model.ChannelTypePublicThread, model.ChannelTypePrivateThread, model.ChannelTypeAnnouncementThread:
	default:
		return model.Channel{}, fmt.Errorf("channel type %d is not a thread type", start.Type)
	}
	body, err := threadStartBody(start.Name, start.AutoArchiveDuration, start.RateLimitPerUser)
	if err != nil {
		return model.Channel{}, err
	}
	body["type"] = start.Type
	start.Invitable.Apply(body, "invitable")
	var thread model.Channel
	err = c.call(ctx, apply(&request{method: http.MethodPost, path: path("channels", url.PathEscape(channelID), "threads"), body: body, dest: &thread}, options))
	return thread, err
}

func (c *Client) StartThreadFromMessage(ctx context.Context, channelID, messageID string, start model.ThreadStart, options ...RequestOption) (model.Channel, error) {
	if channelID == "" || messageID == "" {
		return model.Channel{}, errors.New("channel ID and message ID are required")
	}
	if start.Type != 0 || start.Invitable.IsSet() {
		return model.Channel{}, errors.New("a thread started from a message takes only a name, an archive duration and slowmode")
	}
	body, err := threadStartBody(start.Name, start.AutoArchiveDuration, start.RateLimitPerUser)
	if err != nil {
		return model.Channel{}, err
	}
	var thread model.Channel
	err = c.call(ctx, apply(&request{method: http.MethodPost, path: messagePath(channelID, messageID) + "/threads", body: body, dest: &thread}, options))
	return thread, err
}

func (c *Client) CreateForumPost(ctx context.Context, channelID string, post model.ForumPostStart, options ...RequestOption) (model.ForumPost, error) {
	if channelID == "" {
		return model.ForumPost{}, errors.New("channel ID is required")
	}
	if len(post.AppliedTags) > model.MaxAppliedTags {
		return model.ForumPost{}, fmt.Errorf("a forum post takes at most %d tags", model.MaxAppliedTags)
	}
	message := post.Message
	if message.ReplyTo != "" || message.Nonce != "" || message.TTS || message.Forward != nil {
		return model.ForumPost{}, errors.New("the first message of a forum post cannot be a reply, a forward, text-to-speech or carry a nonce")
	}
	if message.Content == "" && len(message.Embeds) == 0 && len(message.Files) == 0 && len(message.Attachments) == 0 && len(message.StickerIDs) == 0 {
		return model.ForumPost{}, errors.New("a forum post needs content, an embed, a file or a sticker")
	}
	first, err := messageBody(message.Content, message.Embeds, message.Components, message.AllowedMentions, message.Files, message.Attachments, "forum post")
	if err != nil {
		return model.ForumPost{}, err
	}
	if message.Flags != 0 {
		first["flags"] = message.Flags
	}
	if len(message.StickerIDs) != 0 {
		first["sticker_ids"] = message.StickerIDs
	}
	body, err := threadStartBody(post.Name, post.AutoArchiveDuration, post.RateLimitPerUser)
	if err != nil {
		return model.ForumPost{}, err
	}
	body["message"] = first
	if len(post.AppliedTags) != 0 {
		body["applied_tags"] = post.AppliedTags
	}
	var created model.ForumPost
	err = c.call(ctx, apply(&request{method: http.MethodPost, path: path("channels", url.PathEscape(channelID), "threads"), body: body, files: message.Files, dest: &created}, options))
	return created, err
}

func (c *Client) ModifyThread(ctx context.Context, threadID string, edit model.ThreadEdit, options ...RequestOption) (model.Channel, error) {
	if threadID == "" {
		return model.Channel{}, errors.New("thread ID is required")
	}
	if name, ok := edit.Name.Value(); edit.Name.IsSet() && (!ok || name == "" || utf8.RuneCountInString(name) > model.MaxChannelName) {
		return model.Channel{}, fmt.Errorf("thread name must be 1 to %d characters", model.MaxChannelName)
	}
	if duration, ok := edit.AutoArchiveDuration.Value(); ok && !validAutoArchive(duration) {
		return model.Channel{}, errors.New("auto archive duration must be 60, 1440, 4320 or 10080 minutes")
	}
	if tags, ok := edit.AppliedTags.Value(); ok && len(tags) > model.MaxAppliedTags {
		return model.Channel{}, fmt.Errorf("a forum post takes at most %d tags", model.MaxAppliedTags)
	}
	body := map[string]any{}
	edit.Name.Apply(body, "name")
	edit.Archived.Apply(body, "archived")
	edit.AutoArchiveDuration.Apply(body, "auto_archive_duration")
	edit.Locked.Apply(body, "locked")
	edit.RateLimitPerUser.Apply(body, "rate_limit_per_user")
	edit.Flags.Apply(body, "flags")
	if tags, ok := edit.AppliedTags.Value(); ok && tags == nil {
		body["applied_tags"] = []string{}
	} else {
		edit.AppliedTags.Apply(body, "applied_tags")
	}
	edit.Invitable.Apply(body, "invitable")
	if len(body) == 0 {
		return model.Channel{}, errors.New("thread edit has no changes")
	}
	var thread model.Channel
	err := c.call(ctx, apply(&request{method: http.MethodPatch, path: path("channels", url.PathEscape(threadID)), body: body, dest: &thread}, options))
	return thread, err
}

func (c *Client) ArchiveThread(ctx context.Context, threadID string, archived bool, options ...RequestOption) (model.Channel, error) {
	return c.ModifyThread(ctx, threadID, model.ThreadEdit{Archived: model.Set(archived)}, options...)
}

func (c *Client) LockThread(ctx context.Context, threadID string, locked bool, options ...RequestOption) (model.Channel, error) {
	return c.ModifyThread(ctx, threadID, model.ThreadEdit{Locked: model.Set(locked)}, options...)
}

func (c *Client) JoinThread(ctx context.Context, threadID string) error {
	return c.threadMember(ctx, http.MethodPut, threadID, "@me")
}

func (c *Client) LeaveThread(ctx context.Context, threadID string) error {
	return c.threadMember(ctx, http.MethodDelete, threadID, "@me")
}

func (c *Client) AddThreadMember(ctx context.Context, threadID, userID string) error {
	if userID == "" {
		return errors.New("user ID is required")
	}
	return c.threadMember(ctx, http.MethodPut, threadID, url.PathEscape(userID))
}

func (c *Client) RemoveThreadMember(ctx context.Context, threadID, userID string) error {
	if userID == "" {
		return errors.New("user ID is required")
	}
	return c.threadMember(ctx, http.MethodDelete, threadID, url.PathEscape(userID))
}

func (c *Client) ThreadMember(ctx context.Context, threadID, userID string, withMember bool) (model.ThreadMember, error) {
	if threadID == "" || userID == "" {
		return model.ThreadMember{}, errors.New("thread ID and user ID are required")
	}
	values := url.Values{}
	if withMember {
		values.Set("with_member", "true")
	}
	var member model.ThreadMember
	err := c.call(ctx, &request{method: http.MethodGet, path: path("channels", url.PathEscape(threadID), "thread-members", url.PathEscape(userID)), query: values, dest: &member})
	return member, err
}

func (c *Client) ThreadMembers(ctx context.Context, threadID string, query model.ThreadMemberQuery) ([]model.ThreadMember, error) {
	if threadID == "" {
		return nil, errors.New("thread ID is required")
	}
	if query.Limit < 0 || query.Limit > maxThreadMembersPerPage {
		return nil, fmt.Errorf("thread member limit must be between 1 and %d", maxThreadMembersPerPage)
	}
	values := url.Values{}
	setQuery(values, "limit", limitString(query.Limit))
	setQuery(values, "after", query.After)
	if query.WithMember {
		values.Set("with_member", "true")
	}
	var members []model.ThreadMember
	err := c.call(ctx, &request{method: http.MethodGet, path: path("channels", url.PathEscape(threadID), "thread-members"), query: values, dest: &members})
	return members, err
}

func (c *Client) ActiveThreads(ctx context.Context, guildID string) (model.ThreadList, error) {
	if guildID == "" {
		return model.ThreadList{}, errors.New("guild ID is required")
	}
	var list model.ThreadList
	err := c.call(ctx, &request{method: http.MethodGet, path: path("guilds", url.PathEscape(guildID), "threads", "active"), dest: &list})
	return list, err
}

func (c *Client) PublicArchivedThreads(ctx context.Context, channelID string, query model.ArchivedThreadQuery) (model.ThreadList, error) {
	return c.archivedThreads(ctx, channelID, "threads/archived/public", query)
}

func (c *Client) PrivateArchivedThreads(ctx context.Context, channelID string, query model.ArchivedThreadQuery) (model.ThreadList, error) {
	return c.archivedThreads(ctx, channelID, "threads/archived/private", query)
}

func (c *Client) JoinedPrivateArchivedThreads(ctx context.Context, channelID string, query model.ArchivedThreadQuery) (model.ThreadList, error) {
	return c.archivedThreads(ctx, channelID, "users/@me/threads/archived/private", query)
}

func (c *Client) SearchThreads(ctx context.Context, channelID string, search model.ThreadSearch) (model.ThreadList, error) {
	if channelID == "" {
		return model.ThreadList{}, errors.New("channel ID is required")
	}
	if search.Limit < 0 || search.Limit > maxThreadSearchResults || search.Offset < 0 {
		return model.ThreadList{}, fmt.Errorf("thread search takes a limit of 1 to %d and a non-negative offset", maxThreadSearchResults)
	}
	if utf8.RuneCountInString(search.Name) > model.MaxChannelName || len(search.Tags) > model.MaxForumTags {
		return model.ThreadList{}, fmt.Errorf("thread search takes a name of at most %d characters and at most %d tags", model.MaxChannelName, model.MaxForumTags)
	}
	values := url.Values{}
	setQuery(values, "name", search.Name)
	for _, tag := range search.Tags {
		values.Add("tag", tag)
	}
	setQuery(values, "tag_setting", search.TagSetting)
	if search.Archived != nil {
		values.Set("archived", strconv.FormatBool(*search.Archived))
	}
	setQuery(values, "sort_by", search.SortBy)
	setQuery(values, "sort_order", search.SortOrder)
	setQuery(values, "limit", limitString(search.Limit))
	setQuery(values, "offset", limitString(search.Offset))
	setQuery(values, "max_id", search.MaxID)
	setQuery(values, "min_id", search.MinID)
	var list model.ThreadList
	err := c.call(ctx, &request{method: http.MethodGet, path: path("channels", url.PathEscape(channelID), "threads", "search"), query: values, dest: &list})
	return list, err
}

func (c *Client) CreateForumTag(ctx context.Context, channelID string, tag model.ForumTag, options ...RequestOption) (model.Channel, error) {
	if channelID == "" {
		return model.Channel{}, errors.New("channel ID is required")
	}
	if err := validateForumTag(tag); err != nil {
		return model.Channel{}, err
	}
	tag.ID = ""
	var channel model.Channel
	err := c.call(ctx, apply(&request{method: http.MethodPost, path: path("channels", url.PathEscape(channelID), "tags"), body: tag, dest: &channel}, options))
	return channel, err
}

func (c *Client) ModifyForumTag(ctx context.Context, channelID string, tag model.ForumTag, options ...RequestOption) (model.Channel, error) {
	if channelID == "" || tag.ID == "" {
		return model.Channel{}, errors.New("channel ID and tag ID are required")
	}
	if err := validateForumTag(tag); err != nil {
		return model.Channel{}, err
	}
	tagID := tag.ID
	tag.ID = ""
	var channel model.Channel
	err := c.call(ctx, apply(&request{method: http.MethodPut, path: path("channels", url.PathEscape(channelID), "tags", url.PathEscape(tagID)), body: tag, dest: &channel}, options))
	return channel, err
}

func (c *Client) DeleteForumTag(ctx context.Context, channelID, tagID string, options ...RequestOption) (model.Channel, error) {
	if channelID == "" || tagID == "" {
		return model.Channel{}, errors.New("channel ID and tag ID are required")
	}
	var channel model.Channel
	err := c.call(ctx, apply(&request{method: http.MethodDelete, path: path("channels", url.PathEscape(channelID), "tags", url.PathEscape(tagID)), dest: &channel}, options))
	return channel, err
}

func (c *Client) ForumPostData(ctx context.Context, channelID string, threadIDs []string) (map[string]model.ForumPostData, error) {
	if channelID == "" {
		return nil, errors.New("channel ID is required")
	}
	if len(threadIDs) == 0 || len(threadIDs) > maxPostDataThreads {
		return nil, fmt.Errorf("post data takes 1 to %d post IDs", maxPostDataThreads)
	}
	var result struct {
		Threads map[string]model.ForumPostData `json:"threads"`
	}
	err := c.call(ctx, &request{method: http.MethodPost, path: path("channels", url.PathEscape(channelID), "post-data"), body: map[string]any{"thread_ids": threadIDs}, dest: &result, idempotent: true})
	return result.Threads, err
}

func (c *Client) threadMember(ctx context.Context, method, threadID, target string) error {
	if threadID == "" {
		return errors.New("thread ID is required")
	}
	return c.call(ctx, &request{method: method, path: path("channels", url.PathEscape(threadID), "thread-members", target)})
}

func (c *Client) archivedThreads(ctx context.Context, channelID, suffix string, query model.ArchivedThreadQuery) (model.ThreadList, error) {
	if channelID == "" {
		return model.ThreadList{}, errors.New("channel ID is required")
	}
	if query.Limit != 0 && (query.Limit < 2 || query.Limit > maxThreadsPerPage) {
		return model.ThreadList{}, fmt.Errorf("archived thread limit must be between 2 and %d", maxThreadsPerPage)
	}
	values := url.Values{}
	setQuery(values, "limit", limitString(query.Limit))
	setQuery(values, "before", query.Before)
	var list model.ThreadList
	err := c.call(ctx, &request{method: http.MethodGet, path: path("channels", url.PathEscape(channelID)) + "/" + suffix, query: values, dest: &list})
	return list, err
}

func threadStartBody(name string, autoArchive int, slowmode model.Optional[int]) (map[string]any, error) {
	if n := utf8.RuneCountInString(name); n == 0 || n > model.MaxChannelName {
		return nil, fmt.Errorf("thread name must be 1 to %d characters", model.MaxChannelName)
	}
	body := map[string]any{"name": name}
	if autoArchive != 0 {
		if !validAutoArchive(autoArchive) {
			return nil, errors.New("auto archive duration must be 60, 1440, 4320 or 10080 minutes")
		}
		body["auto_archive_duration"] = autoArchive
	}
	if seconds, ok := slowmode.Value(); ok {
		if seconds < 0 || seconds > model.MaxSlowmodeSeconds {
			return nil, fmt.Errorf("slowmode must be 0 to %d seconds", model.MaxSlowmodeSeconds)
		}
		body["rate_limit_per_user"] = seconds
	}
	return body, nil
}

func validAutoArchive(minutes int) bool {
	switch minutes {
	case model.AutoArchiveHour, model.AutoArchiveDay, model.AutoArchiveThreeDays, model.AutoArchiveWeek:
		return true
	}
	return false
}
