# Migrating from v0.3 to v0.4

v0.4 moves all code from the root `micup` package into domain packages under `src/`. Behavior is unchanged; names and import paths change.

```go
import (
	"github.com/yestefir/micup.go/src/bot"
	"github.com/yestefir/micup.go/src/gateway"
	"github.com/yestefir/micup.go/src/model"
	"github.com/yestefir/micup.go/src/rest"
	"github.com/yestefir/micup.go/src/session"
	"github.com/yestefir/micup.go/src/shard"
	"github.com/yestefir/micup.go/src/voice"
)
```

| v0.3 | v0.4 |
|---|---|
| `micup.New(ctx, micup.Config{...})` | `bot.New(ctx, bot.Config{...})` |
| `client.BotID` | `b.ID` |
| `client.MessageContentEnabled` | `b.MessageContentEnabled` |
| `client.GatewayEndpoint()` | `b.Endpoints().Gateway` |
| `client.REST(ctx, method, path, body, dest)` | `b.REST.Do(ctx, method, path, body, dest)` |
| `client.SendMessage`, `SendMessageWithOptions`, `Channel` | `b.REST.SendMessage`, `SendMessageWithOptions`, `Channel` |
| `client.RespondToInteraction`, `DeferInteraction`, `SendInteractionFollowup` | same names on `b.REST` |
| `client.Gateway(ctx, handler)` | `b.Gateway(ctx, handler)` |
| `client.NewSession(micup.SessionOptions{...})` | `b.NewSession(session.Options{...})`, now also returns an error |
| `client.NewShardManager(micup.ShardManagerOptions{...})` | `b.NewShardManager(shard.Options{...})` |
| `micup.Session`, `micup.VoiceJoin` | `session.Session`, `session.VoiceJoin` |
| `micup.ShardManager` | `shard.Manager` |
| `micup.State`, `micup.NewState` | `state.Cache`, `state.New` |
| `State.Apply(event)` | `Cache.Apply(event.Name, event.Data)` |
| `micup.ShardForGuild` | `shard.ForGuild` |
| `micup.DefaultIdentifyInterval` | `shard.DefaultIdentifyInterval` |
| `micup.Event`, `micup.EventHandler` | `gateway.Event`, `gateway.Handler` |
| `micup.EventReady`, ... | `gateway.EventReady`, ... |
| `micup.Op*` | `gateway.Op*` |
| `micup.CloseError`, `ErrReconnectRequested`, `ErrSessionInvalidated` | same names in `gateway` |
| `micup.ErrGatewayClosed` | `gateway.ErrClosed` |
| `micup.ErrNotReady`, `ErrVoiceJoinPending`, `ErrVoiceGrant`, `ErrVoiceMoved` | same names in `session` |
| `micup.ErrShardNotOwned`, `ErrShardManagerRunning` | `shard.ErrNotOwned`, `shard.ErrRunning` |
| `micup.APIError` | `rest.APIError` |
| `micup.User`, `Channel`, `Message`, `MessageCreate`, `MessageOptions`, `Embed*`, `Interaction`, `VoiceState`, `VoiceServerUpdate`, `MessageType*` | same names in `model` |
| `micup.DefaultOrigin` | `discovery.DefaultOrigin` |
| `micup.Version` | `bot.Version` |
| `github.com/yestefir/micup.go/voice` | `github.com/yestefir/micup.go/src/voice` |
