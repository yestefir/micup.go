# Channels, roles, threads and invites

All methods are on `b.REST`. Methods that change a guild accept `micup.WithReason("...")` for the audit log.

## Optional fields

Edit structs use `micup.Optional[T]`, which has three states: left out (the zero value, the field keeps its value), `micup.Set(v)` (the field changes to `v`, even when `v` is `0`, `false` or `""`), and `micup.Clear[T]()` (the field is sent as `null`, which Micup reads as "remove"):

```go
_, err := b.REST.ModifyChannel(ctx, channelID, micup.ChannelEdit{
	Topic:            micup.Clear[string](),
	RateLimitPerUser: micup.Set(0),
	ParentID:         micup.Set(categoryID),
}, micup.WithReason("tidy up"))
```

An edit with no fields returns an error before any request is sent.

## Channels

| Method | Endpoint and notes |
|---|---|
| `CreateChannel(ctx, guildID, micup.ChannelCreate{Type, Name, Options})` | `Options` takes every `ChannelEdit` field except `Name`, `Type`, `RTCRegion` and `RTCP2P`. Needs `MANAGE_CHANNELS` |
| `ModifyChannel(ctx, channelID, micup.ChannelEdit{...})` | `PermissionOverwrites` replaces every overwrite; use `SetChannelPermission` for one. `NSFW` sends `nsfw_override`, so `Clear` inherits the parent's setting |
| `DeleteChannel(ctx, channelID)` | Deletes a guild channel with its messages and threads, or closes a DM |
| `ModifyChannelPositions(ctx, guildID, []micup.ChannelPosition{...})` | Moves channels and categories; `ParentID: micup.Clear[string]()` moves to the top level |
| `SetChannelPermission(ctx, channelID, micup.PermissionOverwrite{ID, Type, Allow, Deny})`, `DeleteChannelPermission` | One role or member overwrite; needs `MANAGE_ROLES` |
| `FollowChannel(ctx, announcementID, targetID)`, `ChannelFollowerStats` | Announcement channel follows |
| `CrosspostMessage(ctx, channelID, messageID)` | Publishes an announcement to followers (10 in a row, then one per 6 minutes) |

Names are 1 to 100 characters, topics at most 4,096 and slowmode 0 to 21,600 seconds; forum tags are checked for count, length, duplicates and emoji fields.

## Roles

| Method | Notes |
|---|---|
| `CreateRole(ctx, guildID, micup.RoleCreate{Name, Color, Permissions, Icon, UnicodeEmoji})` | A nil `Permissions` copies the everyone role |
| `ModifyRole(ctx, guildID, roleID, micup.RoleEdit{...})` | Only `HoistPosition`, `Icon` and `UnicodeEmoji` can be cleared |
| `DeleteRole`, `ModifyRolePositions`, `ModifyRoleHoistPositions`, `ResetRoleHoistPositions` | The everyone role cannot be deleted or moved |

Micup refuses role icons (`Icon` and `UnicodeEmoji`) with `INVALID_FORM_BODY` until the guild reaches boost level 1.

`micup.ImageData("image/png", bytes)` builds the data URL Micup expects for role icons, emojis, stickers, webhook avatars and guild images.

## Guild settings

`ModifyGuild(ctx, guildID, micup.GuildEdit{...})` changes the name, images, system and AFK channels, verification, content filters and the features you may toggle. `Features` replaces the whole toggleable set: send the current list with your changes. `CurrentUserGuilds`, `LeaveGuild`, `VanityURL`, `ModifyVanityURL`, `Bans` and `AuditLog(ctx, guildID, micup.AuditLogQuery{Limit, Before, UserID, ActionType})` cover the rest; `micup.AuditMemberBanAdd` and the other `Audit…` constants name the actions.

## Threads and forums

| Method | Notes |
|---|---|
| `StartThread(ctx, channelID, micup.ThreadStart{Name, Type, AutoArchiveDuration, RateLimitPerUser, Invitable})` | `Type` defaults to a public thread |
| `StartThreadFromMessage(ctx, channelID, messageID, micup.ThreadStart{Name})` | |
| `CreateForumPost(ctx, forumID, micup.ForumPostStart{Name, Message, AppliedTags})` | The first message takes content, embeds, components, files and stickers |
| `ModifyThread`, `ArchiveThread`, `LockThread` | `DeleteChannel` deletes a thread |
| `JoinThread`, `LeaveThread`, `AddThreadMember`, `RemoveThreadMember`, `ThreadMember`, `ThreadMembers` | |
| `ActiveThreads`, `PublicArchivedThreads`, `PrivateArchivedThreads`, `JoinedPrivateArchivedThreads`, `SearchThreads` | |
| `CreateForumTag`, `ModifyForumTag`, `DeleteForumTag`, `ForumPostData` | `ModifyForumTag` replaces every field of the tag |

Archive durations are `micup.AutoArchiveHour`, `AutoArchiveDay`, `AutoArchiveThreeDays` and `AutoArchiveWeek`. A thread uses its parent's permission overwrites, and the state cache computes thread permissions that way.

## Invites

`CreateInvite(ctx, channelID, micup.InviteCreate{MaxUses, MaxAge, Unique, Temporary})` (0 to 100 uses, 0 to 604,800 seconds), `Invite`, `ChannelInvites`, `GuildInvites` and `DeleteInvite`. `micup.InviteURL(b.Endpoints().Invite, code)` builds a link.

## Emojis and stickers

`Emojis`, `CreateEmoji`, `CreateEmojis` (1 to 50), `CloneEmoji`, `ModifyEmoji`, `DeleteEmoji(ctx, guildID, emojiID, purge)` and the same set for stickers. Bulk creates return `Success` and `Failed` lists; a request succeeds even when every item fails. `ModifySticker` writes the description and tags every time, so send both.

## Search

`SearchMessages(ctx, micup.MessageSearch{GuildID, ChannelID, Content, Has, AuthorIDs, ...})` searches one guild or channel; bots cannot use the other scopes. When `Indexing` is true, Micup is building the index and the request should be retried. `SearchMembers(ctx, guildID, micup.MemberSearch{...})` needs a moderation permission.

## Paging

Iterators page through results and stop at the first error:

```go
for message, err := range b.REST.EachMessage(ctx, channelID, micup.MessageQuery{}) {
	if err != nil {
		return err
	}
	fmt.Println(message.ID, message.Content)
}
```

`EachMessage` runs newest first with `Before` (or oldest first when only `After` is set), and `EachMember`, `EachGuild`, `EachReactionUser`, `EachAuditLogEntry`, `EachPin` and `EachThreadMember` page the matching lists. Breaking out of the loop stops the requests.

## Errors

`micup.ErrorCode(err)`, `micup.IsErrorCode(err, micup.CodeUnknownMessage, ...)`, `micup.IsNotFound(err)` and `micup.IsForbidden(err)` read wrapped `*micup.APIError` values. `APIError.HasFieldCode("…")` checks one validation error.
