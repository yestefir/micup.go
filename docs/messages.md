# Messages, reactions and files

All methods are on `b.REST` (`*micup.RESTClient`) and take a context.

## Sending

```go
msg, err := b.REST.SendMessage(ctx, channelID, "pong", replyToID)

msg, err = b.REST.SendMessageWithOptions(ctx, channelID, micup.MessageOptions{
	Content: "Now playing",
	Embeds:  []micup.Embed{{Title: "Track", Color: 0x5865F2}},
	ReplyTo: messageID,
	Nonce:   messageID,
	Flags:   micup.MessageFlagSuppressNotifications,
})
```

| Field | Meaning |
|---|---|
| `Content`, `Embeds` (max 10), `StickerIDs` (max 3) | What the message shows. A message needs text, an embed, a component, a file or a sticker |
| `Components` | Rows of buttons or a select menu; see [Buttons and select menus](components.md) |
| `ReplyTo` | The message ID to reply to |
| `Nonce` | Your ID for the message (up to 32 characters). Micup remembers it for 5 minutes per channel and returns the first message when you send it again, so a send with a nonce is safe to retry and the client retries it on 5xx |
| `Flags` | `MessageFlagSuppressEmbeds`, `MessageFlagSuppressNotifications` or `MessageFlagVoiceMessage` |
| `AllowedMentions` | Which mentions notify. When nil, no mention notifies anyone (`parse: []`). Do not combine `Parse` with `Users` or `Roles` |
| `Files` | Files sent in the same request as multipart |
| `Attachments` | Files uploaded earlier with `UploadAttachment` |

Bots can send at least 4,000 characters per message.

## Files

Small files go in the same request:

```go
f, _ := os.Open("cover.png")
defer f.Close()
_, err := b.REST.SendMessageWithOptions(ctx, channelID, micup.MessageOptions{
	Content: "Cover art",
	Files:   []micup.File{{Name: "cover.png", ContentType: "image/png", Reader: f}},
})
```

The whole multipart body is built in memory before it is sent. For large files, upload first; Micup recommends it whenever the instance allows presigned uploads:

```go
uploaded, err := b.REST.UploadAttachment(ctx, channelID, micup.Upload{
	Name: "mix.ogg", ContentType: "audio/ogg", Size: info.Size(), Reader: f,
})
if err != nil {
	return err
}
_, err = b.REST.SendMessageWithOptions(ctx, channelID, micup.MessageOptions{
	Attachments: []micup.UploadedAttachment{uploaded},
})
```

`UploadAttachment` asks Micup for an upload plan, sends the bytes to the returned URLs (in 10 MiB parts above 10 MiB) without the bot token, and completes multipart uploads. `Size` must be the exact byte count. Upload URLs are secrets until they expire; the library never logs them. A bot can attach files up to 50 MiB unless the instance changes the limit, and at most 10 per message.

## Reading and editing

| Method | Endpoint |
|---|---|
| `Message(ctx, channelID, messageID)` | One message |
| `Messages(ctx, channelID, micup.MessageQuery{Limit, Before, After, Around})` | Newest first, up to 100 |
| `EditMessage(ctx, channelID, messageID, micup.MessageEdit{...})` | Nil fields stay the same; `Embeds`, `Components` and attachments replace the whole list |
| `DeleteMessage(ctx, channelID, messageID, micup.WithReason("..."))` | Delete one message |
| `BulkDeleteMessages(ctx, channelID, ids, ...)` | 1 to 100 messages of any age, needs `MANAGE_MESSAGES` |
| `TriggerTyping(ctx, channelID)` | Show the typing indicator |
| `CreateDM(ctx, userID)` | Open (or return) the DM channel with a user |

Only `DEFAULT` and `REPLY` messages can be edited, pinned or replied to.

## Reactions and pins

Pass a Unicode emoji as is, or `name:id` for a custom emoji (`micup.EmojiString(emoji)` builds it from a `micup.Emoji`).

| Method | Effect |
|---|---|
| `AddReaction`, `RemoveReaction` | The bot's own reaction |
| `RemoveUserReaction(ctx, ch, msg, emoji, userID)` | Another user's reaction, needs `MANAGE_MESSAGES` |
| `RemoveEmojiReactions`, `RemoveAllReactions` | Clear one emoji or every reaction |
| `PinMessage`, `UnpinMessage` | Needs `PIN_MESSAGES`; accept `WithReason` |
| `PinnedMessages(ctx, ch, micup.PinQuery{Limit, Before})` | A page of pins; pass the last `PinnedAt` as `Before` for the next page |

## Gateway payloads

`micup.MessageCreate` (and `micup.MessageUpdate`) embeds `micup.Message` and adds `ChannelType`, `MentionHere` and, in guild channels, the author's `Member` with role IDs. `Message` carries attachments, embeds, buttons and select menus (`Components`), reactions, `MessageReference` and `ReferencedMessage` for replies.

## Webhooks

| Method | Effect |
|---|---|
| `CreateWebhook(ctx, channelID, name, ...)` | Creates a webhook and returns it with its token; needs `MANAGE_WEBHOOKS` |
| `ChannelWebhooks(ctx, channelID)`, `GuildWebhooks(ctx, guildID)` | List webhooks |
| `DeleteWebhook(ctx, webhookID, ...)` | Delete a webhook |
| `ExecuteWebhook(ctx, id, token, micup.WebhookMessage{...})` | Post as the webhook; returns the message only when `Wait` is true |

`WebhookMessage` embeds `MessageOptions` (without components) and adds `Username` and `AvatarURL` for that message only, and `ThreadID` to post in a thread. The webhook token is a secret: anyone with the ID and token can post.
