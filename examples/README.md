# Runnable examples

These programs use only the public `micup` package. Run them from the repository root with Go 1.26.8 or newer. They use the checked-out library, so they also demonstrate APIs newer than the latest release. No FFmpeg or voice setup is needed.

## Ping bot

Invite your bot to a test guild, enable its message-content grant, and allow it to view the test channel, read message history, and send messages.

Set the token in your terminal; the example does not load `.env` files. Keep real tokens out of source files, screenshots, and commits.

PowerShell:

```powershell
$env:MICUP_BOT_TOKEN = '<your test bot token>'
go run ./examples/ping
```

Bash:

```sh
export MICUP_BOT_TOKEN='<your test bot token>'
go run ./examples/ping
```

Send `!ping` as a human user in the test channel. The bot replies `pong` to that message and ignores messages from bots. Press Ctrl+C to cancel the session and close the Gateway connection.

An empty token fails before a network request. A known disabled message-content grant also stops startup; if the grant is unknown, check the application settings when `!ping` receives no response. Check channel permissions as well.

The library handles Gateway reconnects and session resumption. See [Gateway sessions](../docs/gateway.md) and [typed events](../docs/events.md) for extending this example.
