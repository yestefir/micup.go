# Formatting, media URLs and snowflakes

## Mentions and markdown

| Helper | Result |
|---|---|
| `micup.MentionUser(id)`, `user.Mention()` | `<@id>` |
| `micup.MentionRole(id)`, `role.Mention()` | `<@&id>` |
| `micup.MentionChannel(id)`, `channel.Mention()` | `<#id>` |
| `micup.FormatTimestamp(t, micup.TimestampRelative)` | `<t:unix:R>`; the other styles are `TimestampShortTime` … `TimestampLongDateTime` |
| `emoji.String()`, `emoji.Reaction()` | `<:name:id>` for messages and `name:id` for reaction routes |
| `micup.EscapeMarkdown(text)` | Text with markdown characters escaped |
| `user.DisplayName()`, `member.DisplayName()`, `user.Tag()` | Global name or nickname, falling back to the username |

Mentions in content notify nobody unless `AllowedMentions` allows them.

## Embeds

```go
embed := micup.NewEmbed().
	WithTitle("Now playing").
	WithDescription(micup.EscapeMarkdown(title)).
	WithColor(0x5865F2).
	WithField("Requested by", micup.MentionUser(userID), true).
	WithTimestamp(time.Now())
```

Every send checks Micup's embed limits (title 256, description 4,096, 25 fields, field name 256 and value 1,024, footer 2,048) before the request. `micup.AttachmentURL("cover.png")` points an image at a file sent in the same message.

## Media URLs

`b.Media()` uses the media base URL from discovery:

```go
avatar := b.Media().UserAvatar(user, micup.ImageOptions{Size: 256})
icon := b.Media().GuildIcon(guild, micup.ImageOptions{Format: micup.ImagePNG})
```

`MemberAvatar` prefers the guild avatar. `UserBanner`, `GuildBanner`, `GuildSplash`, `RoleIcon`, `EmojiImage` and `StickerImage` work the same way. A missing hash returns an empty string. Media URLs need no token.

## Snowflakes

Micup IDs are decimal strings. `micup.SnowflakeTime(id)` returns the creation time, `micup.SnowflakeAt(t)` builds a cursor for `Before` or `After`, and `micup.CompareSnowflakes(a, b)` orders IDs numerically. Never compare IDs as plain strings: `"9" > "10"`.

## Permissions

`permissions.String()` and `Names()` list the set flags, `Missing(required)` returns what is lacking (nothing for `ADMINISTRATOR`), and `micup.ParsePermission("BAN_MEMBERS")` reads a name. `micup.ElevatedPermissions` are the bits that need two-factor authentication in guilds with the elevated MFA level.
