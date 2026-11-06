# Slash echo bot

A complete guild-scoped slash command using only the public `micup` package: register `/echo`, decode a typed interaction, and send an ephemeral response visible to the invoking user.

## Run

Use Go 1.26.8 or newer, invite a test bot with application-command access to a test guild, and copy that guild's ID. Run from the repository root. This uses the checked-out library, including APIs not yet in a tagged release. No message-content grant or FFmpeg is needed.

PowerShell:

```powershell
$env:MICUP_BOT_TOKEN = '<your test bot token>'
$env:MICUP_GUILD_ID = '<your test guild ID>'
go run ./examples/slash-echo
```

Bash:

```sh
export MICUP_BOT_TOKEN='<your test bot token>'
export MICUP_GUILD_ID='<your test guild ID>'
go run ./examples/slash-echo
```

The example reads the environment directly, without loading `.env` files. Missing configuration stops startup before network access. Keep tokens out of source code and commits.

Use `/echo text:hello` in the configured guild; expect a private `hello` response. Text is limited to 1,000 characters by the command definition. Whitespace-only input receives a helpful response. Other commands, autocomplete events, and interactions from other guilds are ignored. The library disables mention parsing by default.

## Registration and lifecycle

Use a dedicated test bot and guild: starting this program creates or updates the guild command named `echo`. `CreateGuildCommand` leaves differently named commands in place, unlike `SetGuildCommands`, which replaces the complete list. The registration persists after Ctrl+C; remove the `echo` command with `DeleteGuildCommand` when you no longer need it. The example deliberately leaves it registered between runs.

Ctrl+C cancels the session and closes the Gateway connection. Guild commands should be available immediately; if `/echo` is missing, check the guild ID and application-command access, and make sure the bot has reached Gateway readiness. The startup log announces that connection is starting, not that it is already ready.

The handler responds directly without background work. Micup requires the first response or deferral within 3 seconds. For slower operations, defer first and then edit the response; see [answering commands](../../docs/commands.md#answering).
