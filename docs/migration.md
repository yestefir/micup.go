# Migrating from v0.3 to v0.4

The import path stays `github.com/yestefir/micup.go` and most names are unchanged. Internally the code moved into domain packages under `src/`; applications keep importing the root package and `github.com/yestefir/micup.go/voice`.

| v0.3 | v0.4 |
|---|---|
| `*micup.Client` | `*micup.Bot` |
| `client.BotID` | `b.ID` |
| `client.GatewayEndpoint()` | `b.Endpoints().Gateway` |
| `client.REST(ctx, method, path, body, dest)` | `b.REST.Do(ctx, method, path, body, dest)` |
| `client.SendMessage`, `SendMessageWithOptions`, `Channel` | `b.REST.SendMessage`, `b.REST.SendMessageWithOptions`, `b.REST.Channel` |
| `client.RespondToInteraction`, `DeferInteraction`, `SendInteractionFollowup` | same names on `b.REST` |
| `client.NewSession(options)` returns `*Session` | `b.NewSession(options)` returns `(*Session, error)` |
| `client.NewShardManager(micup.ShardManagerOptions{...})` | `b.NewShardManager(micup.ShardOptions{...})` |
| `micup.State.Apply(event)` | `micup.State.Apply(event.Name, event.Data)` |
| `micup.Interaction.Member`, `User` as raw JSON | `*micup.Member`, `*micup.User` |

Unchanged: `micup.New`, `micup.Config`, `micup.DefaultOrigin`, `micup.Session`, `micup.SessionOptions`, `micup.VoiceJoin`, `micup.Event`, `micup.EventHandler`, every `micup.Event*` and `micup.Op*` constant, `micup.CloseError`, every `micup.Err*` error, `micup.ShardForGuild`, `micup.ShardManager`, `micup.DefaultIdentifyInterval`, `micup.APIError`, all payload types (`micup.MessageCreate`, `micup.MessageOptions`, `micup.Embed`, ...), and the `voice` package.

New: `micup.NewREST`, `micup.RunGateway`, `micup.NewSession`, `micup.NewShardManager` and `micup.NewState` for building parts without `micup.New`, plus `b.SessionConfig()` and `b.Endpoints()`, and [buttons](components.md).
