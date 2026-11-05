# Guilds, members and moderation

All methods are on `b.REST`. Routes that write the audit log accept `micup.WithReason("...")`.

| Method | Endpoint and notes |
|---|---|
| `Guild(ctx, guildID)`, `GuildChannels(ctx, guildID)`, `Roles(ctx, guildID)` | Read a guild, its channels and its roles |
| `Member(ctx, guildID, userID)` | One member |
| `Members(ctx, guildID, micup.MemberQuery{Limit, After})` | Members in ascending user ID order, up to 1,000 per page; pass the last user ID as `After` |
| `ModifyMember(ctx, guildID, userID, micup.MemberEdit{...}, ...)` | Nickname, roles (the full replacement set), mute, deafen, timeout, voice move or disconnect |
| `AddMemberRole`, `RemoveMemberRole` | One role, needs `MANAGE_ROLES` and a higher role |
| `KickMember(ctx, guildID, userID, ...)` | Needs `KICK_MEMBERS` |
| `BanMember(ctx, guildID, userID, micup.Ban{DeleteMessageSeconds, DurationSeconds, Reason}, ...)`, `UnbanMember` | `DeleteMessageSeconds` 0 to 604,800; `DurationSeconds` 0 (permanent) or 60 to 63,072,000 |

`MemberEdit` uses pointers so that unset fields stay unchanged:

```go
channel := "1189375284394692610"
_, err := b.REST.ModifyMember(ctx, guildID, userID, micup.MemberEdit{ChannelID: &channel}, micup.WithReason("moved to music"))

_, err = b.REST.ModifyMember(ctx, guildID, userID, micup.MemberEdit{Disconnect: true})

until := time.Now().Add(10 * time.Minute).UTC().Format(time.RFC3339)
_, err = b.REST.ModifyMember(ctx, guildID, userID, micup.MemberEdit{CommunicationDisabledUntil: &until})
```

`ChannelID` moves the member (needs `MOVE_MEMBERS`, and `VIEW_CHANNEL` and `CONNECT` in the destination) and `Disconnect` removes them from voice; set at most one. An empty `CommunicationDisabledUntil` clears a timeout. `ConnectionID` limits a voice move to one connection.

Check permissions before calling with the [state cache](gateway.md#state-cache), for example `s.State().BotPermissions(guildID, "")` for guild-wide permissions.
