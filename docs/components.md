# Buttons

Bots can put buttons under their messages and receive a click as an `INTERACTION_CREATE` event. This works for ordinary bot applications, not only official Micup apps. The payloads follow the Fluxer (Discord-compatible) component format. Slash and context menu commands arrive the same way; see [Slash and context menu commands](commands.md).

## Sending buttons

Buttons sit in action rows. A message holds up to 5 rows and a row holds up to 5 buttons.

```go
rows := []micup.ActionRow{
	micup.NewActionRow(
		micup.PrimaryButton("player:pause", "Pause").WithEmoji(micup.Emoji{Name: "⏯️"}),
		micup.SecondaryButton("player:skip", "Skip"),
		micup.DangerButton("player:stop", "Stop"),
	),
	micup.NewActionRow(micup.LinkButton("https://micup.gg", "Micup")),
}
msg, err := b.REST.SendMessageWithOptions(ctx, channelID, micup.MessageOptions{
	Embeds:     []micup.Embed{{Title: "Now playing"}},
	Components: rows,
})
```

| Constructor | Style | Click |
|---|---|---|
| `PrimaryButton(customID, label)` | `ButtonStylePrimary` (blue) | `INTERACTION_CREATE` with `customID` |
| `SecondaryButton` | `ButtonStyleSecondary` (grey) | same |
| `SuccessButton` | `ButtonStyleSuccess` (green) | same |
| `DangerButton` | `ButtonStyleDanger` (red) | same |
| `LinkButton(url, label)` | `ButtonStyleLink` | opens the URL; no event |

`WithEmoji` and `WithDisabled` return a changed copy. A button needs a label (up to 80 characters), an emoji, or both. Every non-link button needs a custom ID of 1 to 100 characters that is unique within the message; link buttons need a URL and no custom ID. The client checks these rules before sending and returns an error without calling Micup.

A message with only buttons is allowed. Webhook messages cannot carry buttons.

## Changing buttons

`EditMessage` replaces the whole button list when `Components` is set, and leaves it alone when it is nil:

```go
rows = micup.DisableButtons(msg.Components)
_, err = b.REST.EditMessage(ctx, channelID, msg.ID, micup.MessageEdit{Components: &rows})

none := []micup.ActionRow{}
_, err = b.REST.EditMessage(ctx, channelID, msg.ID, micup.MessageEdit{Components: &none})
```

`DisableButtons` returns a copy with every non-link button disabled, which suits a panel that has timed out. An empty list removes all buttons. Received messages carry their buttons in `Message.Components`.

## Handling clicks

A click arrives as `micup.EventInteractionCreate`. Decode it as `micup.Interaction`:

```go
case micup.EventInteractionCreate:
	var i micup.Interaction
	if err := event.Decode(&i); err != nil {
		return err
	}
	if !i.IsButton() {
		return nil
	}
	switch i.CustomID() {
	case "player:pause":
		content := "Paused by " + i.Invoker().Username
		return b.REST.UpdateInteractionMessage(ctx, i, micup.MessageEdit{Content: &content})
	case "player:stop":
		rows := micup.DisableButtons(i.Message.Components)
		return b.REST.UpdateInteractionMessage(ctx, i, micup.MessageEdit{Components: &rows})
	}
```

| Field or method | Meaning |
|---|---|
| `IsButton()` | The interaction is a button click (`InteractionTypeMessageComponent` with `ComponentTypeButton`) |
| `CustomID()` | The clicked button's custom ID |
| `ComponentData()` | `CustomID` and `ComponentType`; `false` for other interaction types |
| `Invoker()` | The user who clicked (`Member.User` in a guild, `User` in a DM) |
| `Member` | The clicker's guild member, with role IDs |
| `Message` | The message the button is on, with its current `Components` |
| `GuildID`, `ChannelID`, `Locale`, `GuildLocale`, `AppPermissions` | Where it was clicked and what the bot may do there |

Routing by custom ID stays in the application; a prefix such as `player:` keeps a bot's panels apart. Check who clicked before acting: anyone who can see the message can press its buttons.

## Answering a click

Answer every click promptly, or the user sees the click fail. When the work may take more than a moment, acknowledge first with `DeferInteractionUpdate` and edit afterwards. The library does not enforce Micup's time limits.

| Method | Callback | Effect |
|---|---|---|
| `UpdateInteractionMessage(ctx, i, micup.MessageEdit{...})` | 7 | Edits the message the button is on, in the same request |
| `DeferInteractionUpdate(ctx, i)` | 6 | Acknowledges the click and changes nothing yet |
| `EditInteractionResponse(ctx, i, micup.MessageEdit{...})` | `PATCH …/messages/@original` | Edits that message later, after a deferred update |
| `RespondToInteraction(ctx, i, micup.MessageOptions{...}, ephemeral)` | 4 | Posts a new reply; `true` shows it only to the clicker |
| `DeferInteraction(ctx, i, ephemeral)` | 5 | Shows a "thinking" reply to fill in with `EditInteractionResponse` |
| `SendInteractionFollowup(ctx, i, options, ephemeral)` | webhook | Posts another message after the first answer |

Deferring and editing afterwards:

```go
if err := b.REST.DeferInteractionUpdate(ctx, i); err != nil {
	return err
}
content := skipTrack()
_, err := b.REST.EditInteractionResponse(ctx, i, micup.MessageEdit{Content: &content})
```

`UpdateInteractionMessage` and `DeferInteractionUpdate` accept only component interactions, and an interaction update cannot carry files.
