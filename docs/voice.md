# Voice publishing

Micup voice is a two-step handshake. The bot sends Gateway opcode 4 with a guild and voice channel. It then observes its matching `VOICE_STATE_UPDATE` and `VOICE_SERVER_UPDATE`. The server update supplies a LiveKit endpoint, short-lived token, and connection ID. Those values are ephemeral credentials: do not store them, log them, or include them in application command records.

`Session.JoinVoice` (or `micup.ShardManager.JoinVoice`) performs this handshake (see [the Gateway guide](gateway.md#voice-channels)):

```go
update, err := s.JoinVoice(ctx, guildID, channelID, micup.VoiceJoin{SelfDeaf: true})
if err != nil {
	return err
}
```

Connect to the grant with `github.com/yestefir/micup.go/voice` and open an audio track:

```go
room, err := (voice.LiveKit{FFmpeg: "ffmpeg", TrackName: "My Bot"}).Connect(ctx, voice.Grant{
	Endpoint: update.Endpoint,
	Token: update.Token,
})
if err != nil {
	return err
}
defer room.Close()

track, err := room.OpenTrack(ctx)
if err != nil {
	return err
}
defer track.Close()
```

Write signed 16-bit little-endian PCM frames at 48 kHz stereo. Each frame may contain up to 20 ms, or 1,920 samples across both channels:

```go
if err := track.Write(ctx, frame); err != nil {
	return err
}
```

`voice.LiveKit.Bitrate` sets the Opus bitrate in bits per second (default `voice.DefaultBitrate`, 96,000, clamped to 8,000–512,000). Micup voice channels default to 64,000 and allow up to 96,000 (more with the guild's audio bitrate features); pass the channel's `Bitrate` from the state cache to match it:

```go
channel, _ := s.State().Channel(channelID)
room, err := (voice.LiveKit{Bitrate: channel.Bitrate}).Connect(ctx, grant)
```

The helper encodes PCM to Opus with FFmpeg and publishes an Opus microphone track named `TrackName` (default `micup.go`). `Track.Close` lets the encoder flush buffered audio for up to one second, then stops it and unpublishes the track. `Room.Close` disconnects LiveKit. On Windows the FFmpeg process starts without a console window.

When the bot leaves, call `Session.LeaveVoice(guildID, update.ConnectionID)` and close the room. When `micup.SessionOptions.OnVoiceLost` reports the same guild and connection, close the room as well. Retry and ownership across processes stay with the application.

To verify delivery against a real local LiveKit server, set `LIVEKIT_TEST_BINARY` and `FFMPEG_TEST_BINARY` and run `go test ./src/voice`.

A grant is valid for 600 seconds and is only needed to connect. When Micup moves the connection to another region or channel it sends a new grant; set `SessionOptions.OnVoiceGrant` and reconnect the room with it (see [the Gateway guide](gateway.md#voice-channels)).

Limits that the module does not work around:

- Channels with encrypted call media send an `e2ee_key` and need an E2EE-capable LiveKit client. This publisher does not encrypt, so it cannot play in those channels.
- Bots never join peer-to-peer calls; Micup always uses LiveKit for them.
- A channel at its `user_limit` refuses the join with `VOICE_CHANNEL_FULL`, returned as `*micup.VoiceRefusedError`.

This helper publishes outgoing audio only. It does not subscribe to or decode incoming participants. Install FFmpeg and verify `ffmpeg -encoders` lists `libopus`. The executable path can be set explicitly in `voice.LiveKit`.
