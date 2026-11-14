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

An explicit empty `AllowedMentions` also disables mentions and serializes as `{"parse":[]}`. Empty policies never become `{}` or `{"parse":null}`, which Micup rejects in interaction message updates. Explicit user IDs, role IDs, and `RepliedUser` are preserved.

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

The whole multipart body is built in memory before it is sent. The client checks the request context before encoding and between file reads. A reader that can block should honor cancellation itself. For large files, upload first; Micup recommends it whenever the instance allows presigned uploads:

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

For multipart uploads, the client checks the part count and consecutive numbering against the declared file size before reading file data or uploading parts. It sends file segments in part-number order, even if the plan lists the parts in a different order.

Plans must provide a nonempty upload key and, for multipart uploads, an upload ID. Multipart part URLs are checked before reading file data. Upload URLs must use HTTP or HTTPS with a host and no embedded username/password; signed query strings are preserved.

## Reading and editing

| Method | Endpoint |
|---|---|
| `Message(ctx, channelID, messageID)` | One message |
| `Messages(ctx, channelID, micup.MessageQuery{Limit, Before, After, Around})` | Newest first, up to 100 |
| `EditMessage(ctx, channelID, messageID, micup.MessageEdit{...})` | Nil fields stay the same; `Embeds`, `Components` and attachments replace the whole list |
| `DeleteMessage(ctx, channelID, messageID, micup.WithReason("..."))` | Delete one message |
| `BulkDeleteMessages(ctx, channelID, ids, ...)` | 1 to 100 messages of any age, needs `MANAGE_MESSAGES` |
| `TriggerTyping(ctx, channelID)` | Show the typing indicator |
| `CreateDM(ctx, userID)`, `SendDM(ctx, userID, options)` | Open (or return) the DM channel with a user, or send to it. Micup does not let bots list their DMs |
| `ForwardMessage(ctx, channelID, micup.MessageForward{ChannelID, MessageID})` | Forward a message, optionally only some `AttachmentIDs` or `EmbedIndices` |
| `CrosspostMessage`, `CrosspostSource` | Publish an announcement, or find where a published copy came from |
| `DeleteAttachment(ctx, channelID, messageID, attachmentID)` | Remove one attachment from the bot's message |
| `SuppressEmbeds(ctx, channelID, messageID, true)` | Hide or show link previews |
| `RefreshAttachmentURLs(ctx, urls)` | Re-sign up to 50 attachment URLs that are about to expire |

`MessageEdit.KeepAttachments` lists existing attachments to keep (with a new title or alt text); attachments left out are removed, and an empty, non-nil list removes all of them. Without `KeepAttachments`, `Files` or `Attachments`, an edit leaves attachments alone. `File` and `UploadedAttachment` take `Title`, `Description` and `Spoiler`; a voice message upload also takes `Duration` and `Waveform`.

Only `DEFAULT` and `REPLY` messages can be edited, pinned or replied to.

`EachMessage(ctx, channelID, query)` fetches successive pages lazily. With `After`, messages are yielded in ascending ID order; otherwise they are yielded newest first. `Around` is not supported for iteration. Each traversal starts from the supplied query, including when the same iterator is used concurrently. Break from the loop to stop fetching pages. Iteration stops if the next cursor is empty or unchanged; a repeated page may already have been yielded before this check.

## Reactions and pins

Pass a Unicode emoji as is, or `name:id` for a custom emoji (`micup.EmojiString(emoji)` builds it from a `micup.Emoji`).

| Method | Effect |
|---|---|
| `AddReaction`, `RemoveReaction` | The bot's own reaction |
| `RemoveUserReaction(ctx, ch, msg, emoji, userID)` | Another user's reaction, needs `MANAGE_MESSAGES` |
| `RemoveEmojiReactions`, `RemoveAllReactions` | Clear one emoji or every reaction |
| `PinMessage`, `UnpinMessage` | Needs `PIN_MESSAGES`; accept `WithReason` |
| `PinnedMessages(ctx, ch, micup.PinQuery{Limit, Before})` | A page of pins; pass the last `PinnedAt` as `Before` for the next page |
| `ReactionUsers(ctx, ch, msg, emoji, micup.ReactionUsersQuery{Limit, After})` | Who reacted; `EachReactionUser` pages through all of them |

## Gateway payloads

`micup.MessageCreate` (and `micup.MessageUpdate`) embeds `micup.Message` and adds `ChannelType`, `MentionHere` and, in guild channels, the author's `Member` with role IDs. `Message` carries attachments, embeds, buttons and select menus (`Components`), reactions, `MessageReference` and `ReferencedMessage` for replies.

## Webhooks

| Method | Effect |
|---|---|
| `CreateWebhook(ctx, channelID, name, ...)` | Creates a webhook and returns it with its token; needs `MANAGE_WEBHOOKS` |
| `ChannelWebhooks(ctx, channelID)`, `GuildWebhooks(ctx, guildID)` | List webhooks |
| `DeleteWebhook(ctx, webhookID, ...)` | Delete a webhook |
| `Webhook`, `ModifyWebhook(ctx, id, micup.WebhookEdit{Name, Avatar, ChannelID})`, `CreateWebhookWithAvatar` | Read, rename, change the avatar or move a webhook |
| `WebhookWithToken`, `ModifyWebhookWithToken`, `DeleteWebhookWithToken` | The same with only the webhook token; no bot token is sent |
| `WebhookMessage`, `EditWebhookMessage`, `DeleteWebhookMessage` | Manage a message the webhook posted; edits change content, embeds and flags |
| `ExecuteWebhook(ctx, id, token, micup.WebhookMessage{...})` | Post as the webhook; returns the message only when `Wait` is true |

`WebhookMessage` embeds `MessageOptions` (without components) and adds `Username` and `AvatarURL` for that message only, and `ThreadID` to post in a thread. In a forum, `ThreadName` and `AppliedTags` create a new post. The webhook token is a secret: anyone with the ID and token can post.

`CreateWebhook` and `CreateWebhookWithAvatar` accept names of 1 to 80 Unicode code points. Both methods use the same name validation; the avatar is optional.
