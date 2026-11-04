# Gateway events and replies

Call `Client.Gateway(ctx, handler)` to connect. Micup sends HELLO before the client identifies. The runner sends IDENTIFY with the configured shard pair, answers heartbeat requests, and forwards dispatch packets as `Event` values. The handler receives the event name, sequence number, and raw JSON payload.

The runner owns its WebSocket. `Event.Send` sends a native Gateway opcode while handling that event. This supports operations such as joining a voice channel:

```go
payload := map[string]any{
	"guild_id": guildID,
	"channel_id": channelID,
	"self_mute": false,
	"self_deaf": true,
}
if err := event.Send(ctx, 4, payload); err != nil {
	return err
}
```

Handle `VOICE_STATE_UPDATE` and `VOICE_SERVER_UPDATE` to confirm the bot joined and collect the voice endpoint and short-lived token. Keep that token in memory and pair it with the matching guild and connection ID. See [the voice guide](voice.md).

Decode message payloads using `MessageCreate`. Ignore messages authored by bots and use `Client.SendMessage` for plain replies or `Client.SendMessageWithOptions` for embeds. Automatic mentions are disabled for both methods.

Ephemeral messages are interaction responses, not regular channel messages. Decode an `INTERACTION_CREATE` event as `Interaction`, then answer it over REST with `RespondToInteraction`. Pass `true` for the `ephemeral` argument to make the response visible only to the user who invoked the interaction. Use `DeferInteraction` when work needs more time; the ephemeral choice is made when deferring and cannot be changed by later responses. A follow-up can be sent with `SendInteractionFollowup` while the interaction token remains valid.

```go
if event.Name == "INTERACTION_CREATE" {
	var interaction micup.Interaction
	if err := json.Unmarshal(event.Data, &interaction); err != nil {
		return err
	}
	return client.RespondToInteraction(ctx, interaction, micup.MessageOptions{
		Embeds: []micup.Embed{{
			Title:       "Queue",
			Description: "Your track was added.",
			Color:       0x5865F2,
		}},
	}, true)
}
```

The connection runner currently returns when the socket disconnects, the session is invalidated, or the event handler returns an error. Applications own reconnect backoff, resume policy, and shard supervision. Do not start multiple runners with the same shard ID unless the application has an explicit connection ownership policy.
