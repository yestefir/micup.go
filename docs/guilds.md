# Guilds, members and moderation

All methods are on `b.REST`. Routes that write the audit log accept `micup.WithReason("...")`.

`EachGuild(ctx)` fetches successive pages of the bot's guilds. `EachAuditLogEntry(ctx, guildID, query)` fetches audit entries newest first and accepts `Before`, but not `After`. Each traversal has its own cursor and starts from the supplied query. Both iterators stop when the next cursor is empty or unchanged; a repeated page may already have been yielded before this check. Break from the loop to stop fetching pages.

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

## Maintained member inventories

Create the session with `b.NewSession(micup.SessionOptions{TrackMembers: true})` and run it normally. `s.MemberInventories()` returns guild-ID-sorted `[]micup.GuildMembers` snapshots. Each snapshot has `GuildID`, `Members` (user ID to bot flag), `Complete`, and `Available`. Maps are independent copies; callers may modify them. With tracking disabled, the result is empty and no enumeration requests are made.

The session enumerates REST members in pages of 1,000 with two concurrent workers. Each guild enumeration has a 30-second deadline. Failed or malformed enumerations remain incomplete and retry after 10 seconds, in addition to REST rate-limit handling. Changes received during enumeration are merged into its result. READY, guild sync and guild rejoin invalidate older enumerations; RESUME keeps the inventory. Temporary guild outages set `Available` to false; leaving a guild removes it.

Keep member add/update/remove and guild lifecycle events enabled, and grant the bot the platform permissions needed to list members. `Complete` means an enumeration finished and subsequent received events were applied; missing platform events cannot be inferred. Only user IDs and bot flags are retained by this inventory. Role and voice state remain available through `s.State()`.

Standalone `micup.NewSession` users must supply `SessionConfig.Members`, implementing `micup.MemberLookup`. `b.NewSession` and `b.NewShardManager` supply it automatically. Background workers are cancelled and joined when `Run` returns. This API is separate from opcode-8 `RequestGuildMembers`, which returns requested chunks without maintaining a complete inventory.
