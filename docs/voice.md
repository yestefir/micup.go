# Voice publishing

Micup voice is a two-step handshake. The bot sends Gateway opcode 4 with a guild and voice channel. It then observes its matching `VOICE_STATE_UPDATE` and `VOICE_SERVER_UPDATE`. The server update supplies a LiveKit endpoint, short-lived token, and connection ID. Those values are ephemeral credentials: do not store them, log them, or include them in application command records.

Connect to the grant and open an audio track:

```go
room, err := (voice.LiveKit{FFmpeg: "ffmpeg"}).Connect(ctx, voice.Grant{
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

The helper encodes PCM to Opus with FFmpeg and publishes an Opus microphone track. `Track.Close` stops the encoder and unpublishes that track. `Room.Close` disconnects LiveKit. The calling application remains responsible for Micup Gateway voice state, connection IDs, join/leave opcodes, retry, and ownership.

This helper publishes outgoing audio only. It does not subscribe to or decode incoming participants. Install FFmpeg and verify `ffmpeg -encoders` lists `libopus`. The executable path can be set explicitly in `voice.LiveKit`.
