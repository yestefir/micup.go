# Voice publishing

Micup voice is a two-step handshake. The bot sends Gateway opcode 4 with a guild and voice channel. It then observes its matching `VOICE_STATE_UPDATE` and `VOICE_SERVER_UPDATE`. The server update supplies a LiveKit endpoint, short-lived token, and connection ID. Those values are ephemeral credentials: do not store them, log them, or include them in application command records.

`Session.JoinVoice` (or `shard.Manager.JoinVoice`) performs this handshake (see [the Gateway guide](gateway.md#voice-channels)):

```go
update, err := s.JoinVoice(ctx, guildID, channelID, session.VoiceJoin{SelfDeaf: true})
if err != nil {
	return err
}
```

Connect to the grant with `github.com/yestefir/micup.go/src/voice` and open an audio track:

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

The helper encodes PCM to Opus with FFmpeg and publishes an Opus microphone track named `TrackName` (default `micup.go`). `Track.Close` lets the encoder flush buffered audio for up to one second, then stops it and unpublishes the track. `Room.Close` disconnects LiveKit. On Windows the FFmpeg process starts without a console window.

When the bot leaves, call `Session.LeaveVoice(guildID, update.ConnectionID)` and close the room. When `session.Options.OnVoiceLost` reports the same guild and connection, close the room as well. Retry and ownership across processes stay with the application.

To verify delivery against a real local LiveKit server, set `LIVEKIT_TEST_BINARY` and `FFMPEG_TEST_BINARY` and run `go test ./src/voice`.

This helper publishes outgoing audio only. It does not subscribe to or decode incoming participants. Install FFmpeg and verify `ffmpeg -encoders` lists `libopus`. The executable path can be set explicitly in `voice.LiveKit`.
