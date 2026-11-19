# Slash and context menu commands

Micup supports application commands: slash commands with typed options, and user and message context menu commands. A command use arrives as an `INTERACTION_CREATE` event, the same way a button click does.

## Registering commands

For a complete program with environment setup, guild-scoped registration, typed interaction handling, and Ctrl+C shutdown, see the [runnable slash echo bot](../examples/slash-echo/README.md).

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

`DefaultMemberPermissions` hides a command from members without those permissions (`micup.PermissionAdministrator` for example), and `DMPermission` controls whether it appears in DMs. `NameLocalizations` and `DescriptionLocalizations` map locales such as `tr` to translated text; see [Localization](#localization). Micup has no per-command permission overrides: `…/commands/permissions` returns 400.

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

## Routing interactions

`micup.NewEventRouter()` maps commands, autocomplete requests, components and modals to handlers by name, so a bot does not need one large type switch.

```go
router := micup.NewEventRouter()
router.Command("ping", func(ctx context.Context, i micup.Interaction) error {
	return b.REST.RespondToInteraction(ctx, i, micup.MessageOptions{Content: "pong"}, false)
})
router.Command("queue add", addToQueue)
router.Autocomplete("queue add", suggestTracks)
router.Component("skip", skipTrack)
router.Component("page:*", changePage)
router.Modal("feedback", saveFeedback)

mux := micup.NewEventMux()
mux.On(micup.EventInteractionCreate, router.Listener())
```

| Method | Matches |
|---|---|
| `Command(path, handler)` | Slash command path; subcommands and groups are joined with spaces, such as `queue add` |
| `Autocomplete(path, handler)` | The same paths for autocomplete interactions |
| `Component(customID, handler)` | A button or select menu `custom_id` |
| `Modal(customID, handler)` | A modal submit `custom_id` |

A key ending in `*` matches every value with that prefix, and the longest prefix wins. An exact key always wins over a prefix. Registering an empty key or a nil handler panics. `router.Handle(ctx, interaction)` returns `micup.ErrNoRoute` when nothing matches, while `router.Listener()` ignores unrouted interactions and returns every handler error.

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

## Modals

A command or a button can answer with a form. Each text input gets its own row; a modal holds 1–5 inputs.

```go
modal := micup.Modal{CustomID: "song-request", Title: "Request a song", Inputs: []micup.TextInput{
	micup.ShortInput("query", "Song").WithLength(2, 100).WithPlaceholder("Artist and title"),
	micup.ParagraphInput("note", "Note").WithRequired(false),
}}
err := b.REST.RespondWithModal(ctx, i, modal)
```

The submitted form arrives as another `INTERACTION_CREATE` with `i.IsModalSubmit()`:

```go
if data, ok := i.ModalData(); ok && data.CustomID == "song-request" {
	query, _ := data.Value("query")
	return b.REST.RespondToInteraction(ctx, i, micup.MessageOptions{Content: "Requested " + query}, false)
}
```

`Modal.Validate` rejects a title over 45 characters, a missing or duplicate custom ID, more than 5 inputs, a label over 45 characters, an unknown style and lengths outside 0–4000. `RespondWithModal` refuses autocomplete and modal submit interactions, which cannot open another modal. A modal submit is answered with `RespondToInteraction`, `DeferInteraction` or `SendInteractionFollowup`.

## Localization

Micup's verified-bot rules ask for commands and replies in the user's language. `ApplicationCommand`, `CommandOption` and `CommandChoice` carry `NameLocalizations` (`CommandChoice` has no description), and commands and options also carry `DescriptionLocalizations`. Every interaction carries `Locale` (the user's language, for example `tr`) and `GuildLocale`. `ApplicationCommand.Validate` checks the localized names and descriptions against the same limits as the originals: names 1-32 lowercase letters, digits, `-` or `_`, descriptions 1-100 characters, and locale keys that look like language tags.

`micup.Catalog` keeps the texts in one place and serves both uses. The fallback locale holds every message; other locales may translate any subset, and a key missing from a locale falls back to the fallback text:

```go
catalog, err := micup.NewCatalog("en-US", micup.Messages{
	"en-US": {"play.name": "play", "play.description": "Play a song", "playing": "Now playing %s", "no-voice": "Join a voice channel first."},
	"tr":    {"play.name": "oynat", "play.description": "Bir şarkı çal", "playing": "%s çalınıyor", "no-voice": "Önce bir ses kanalına katıl."},
})
if err != nil {
	return err
}

play := micup.ApplicationCommand{
	Name:                     catalog.TextIn("en-US", "play.name"),
	NameLocalizations:        catalog.Localizations("play.name"),
	Description:              catalog.TextIn("en-US", "play.description"),
	DescriptionLocalizations: catalog.Localizations("play.description"),
	Options: []micup.CommandOption{{
		Type: micup.OptionTypeString, Name: "query", Description: "Song name or link", Required: true,
		NameLocalizations:        map[string]string{"tr": "şarkı"},
		DescriptionLocalizations: map[string]string{"tr": "Şarkı adı veya bağlantısı"},
	}},
}

err = b.REST.RespondToInteraction(ctx, interaction,
	micup.MessageOptions{Content: catalog.Text(interaction, "no-voice")}, true)
```

| Method | Behavior |
|---|---|
| `NewCatalog(fallback, messages)` | Builds a catalog. It returns an error when the fallback locale has no messages, when a locale is not a language tag, when a locale is listed twice, or when another locale has a key the fallback lacks (usually a typo) |
| `Text(interaction, key, args...)` | The message in the interaction's `Locale`, or in its `GuildLocale` when `Locale` is empty. `args` are passed to `fmt.Sprintf` when present |
| `TextIn(locale, key, args...)` | The same for a locale you already hold |
| `Localizations(key)` | A `locale → text` map of every translation except the fallback's, for the localization fields above. It is `nil` when nothing is translated |
| `Fallback()`, `Locales()` | The fallback locale and every locale, sorted |

Locales match without regard to case or `_` against `-`. `tr_TR` and `TR` find `tr`; `pt` finds `pt-BR` when no plain `pt` exists; `en-GB` finds `en`. A locale with no match, and an unknown key, give the fallback text and the key itself, so a missing translation shows up as a visible key. Translate error messages too, not only the success path.

micup.go checks only that a locale key looks like a language tag. Micup decides which locales it accepts, so use the codes it documents.

