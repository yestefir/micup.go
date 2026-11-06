# Slash and context menu commands

Micup supports application commands: slash commands with typed options, and user and message context menu commands. A command use arrives as an `INTERACTION_CREATE` event, the same way a button click does.

## Registering commands

Commands belong to the bot's application. `b.ApplicationID` holds its ID (on Micup it equals the bot user ID).

```go
commands := []micup.ApplicationCommand{
	{
		Name:        "echo",
		Description: "Repeat text",
		Options: []micup.CommandOption{
			{Type: micup.OptionTypeString, Name: "text", Description: "What to say", Required: true},
			{Type: micup.OptionTypeInteger, Name: "times", Description: "How many times", Choices: []micup.CommandChoice{{Name: "once", Value: 1}, {Name: "twice", Value: 2}}},
			{Type: micup.OptionTypeString, Name: "fruit", Description: "Pick a fruit", Autocomplete: true},
		},
	},
	{Type: micup.CommandTypeUser, Name: "Inspect user"},
	{Type: micup.CommandTypeMessage, Name: "Quote message"},
}
_, err := b.REST.SetGuildCommands(ctx, b.ApplicationID, guildID, commands)
```

| Method | Request |
|---|---|
| `GlobalCommands`, `GuildCommands` | List commands |
| `GlobalCommand`, `GuildCommand` | Read one command |
| `CreateGlobalCommand`, `CreateGuildCommand` | Create one command |
| `EditGlobalCommand`, `EditGuildCommand` | Change only the non-empty fields of the given command |
| `DeleteGlobalCommand`, `DeleteGuildCommand` | Delete a command |
| `SetGlobalCommands`, `SetGuildCommands` | Replace the whole list in one request (up to 100); `nil` removes every command |

Prefer `Set…Commands` at startup: it is one request, keeps command IDs for unchanged commands, and removes commands that are no longer in the list. Guild commands are visible at once; use them while developing. When the list changes, Micup sends `APPLICATION_COMMANDS_UPDATE` (`micup.ApplicationCommandsUpdate`) with the new version.

Create and set requests are checked before they are sent, and an invalid command returns an error without contacting Micup:

- slash command and option names are 1–32 lowercase letters, digits, `-` or `_` (letters outside ASCII are allowed);
- slash commands and options need a 1–100 character description;
- at most 25 options per level and 25 choices per option; choices and `Autocomplete` exclude each other;
- required options come before optional ones; option names are unique per level;
- subcommand groups contain subcommands, which contain options, and nothing is nested deeper;
- user and message commands have a 1–32 character name and no description or options.

`DefaultMemberPermissions` hides a command from members without those permissions (`micup.PermissionAdministrator` for example), and `DMPermission` controls whether it appears in DMs. `NameLocalizations` and `DescriptionLocalizations` map locales such as `tr` to translated text. Micup has no per-command permission overrides: `…/commands/permissions` returns 400.

## Handling a command

```go
mux.On(micup.EventInteractionCreate, micup.Listen(func(ctx context.Context, i micup.Interaction) error {
	data, ok := i.CommandData()
	if !ok {
		return nil
	}
	switch {
	case i.IsAutocomplete():
		focused, _ := data.Focused()
		prefix, _ := focused.String()
		return b.REST.RespondAutocomplete(ctx, i, fruitsStartingWith(prefix))
	case data.Name == "echo":
		text, _ := data.String("text")
		times, ok := data.Int("times")
		if !ok {
			times = 1
		}
		return b.REST.RespondToInteraction(ctx, i, micup.MessageOptions{Content: strings.Repeat(text+" ", int(times))}, false)
	case data.Name == "Inspect user":
		user, _ := data.TargetUser()
		return b.REST.RespondToInteraction(ctx, i, micup.MessageOptions{Content: "ID " + user.ID}, true)
	}
	return nil
}))
```

`i.IsCommand()` is true for a command use and `i.IsAutocomplete()` while the user types into an autocomplete option. Both carry `CommandData`:

| `CommandData` | Returns |
|---|---|
| `Name`, `Type`, `ID`, `GuildID` | The command that was used |
| `Path()` | Command, group and subcommand names, for example `["admin", "role", "add"]` |
| `Option(name)` | The raw option of the innermost subcommand |
| `String`, `Int`, `Float`, `Bool` | A typed option value; `false` when the option was not given |
| `User`, `Member`, `Role`, `Channel`, `Attachment` | The object Micup resolved for an option; `Member` fills in its `User` |
| `Focused()` | The option the user is typing into, during autocomplete |
| `TargetUser()`, `TargetMessage()` | The target of a user or message context menu command |

Resolved channels and the invoking `Member` carry the `Permissions` Micup computed for that channel. The interaction also holds `Channel`, `Guild` (ID, locale, features), `Context` (`micup.InteractionContextGuild`, `…BotDM`, `…PrivateChannel`), `Locale`, `AppPermissions` and `AttachmentSizeLimit`.

## Answering

A command is answered with the same calls as a button, except that it has no message to update:

| Call | Use |
|---|---|
| `RespondToInteraction(ctx, i, options, ephemeral)` | Reply at once; `true` shows the reply only to the user |
| `DeferInteraction(ctx, i, ephemeral)` then `EditInteractionResponse` | Work that takes longer than a moment |
| `SendInteractionFollowup` | More messages after the first answer |
| `RespondAutocomplete(ctx, i, choices)` | Up to 25 suggestions for an autocomplete interaction; `nil` shows none |

`UpdateInteractionMessage` and `DeferInteractionUpdate` are only for buttons. Answer every command promptly, or the user sees it fail.

Micup expires an interaction about three seconds after sending it: measured on Micup, an autocomplete answer sent after 2.5 seconds was accepted and one sent after 3.5 seconds returned `410 INTERACTION_EXPIRED`. Autocomplete cannot be deferred, so a slow lookup (a web search, for example) should answer from a cache or with a placeholder choice before that limit and finish in the background. A command can be acknowledged with `DeferInteraction` first and edited later.
