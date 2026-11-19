// Package voice joins Micup voice channels' LiveKit rooms and publishes Opus
// audio.
//
// It is a separate import so that bots without voice skip LiveKit and WebRTC.
// Joining is two steps: ask the Gateway for a grant with Session.JoinVoice
// from the micup package, then connect to the room with this package.
//
// # Publishing audio
//
// Connect with the grant and open a track:
//
//	update, err := s.JoinVoice(ctx, guildID, channelID, micup.VoiceJoin{SelfDeaf: true})
//	if err != nil {
//		return err
//	}
//	room, err := (voice.LiveKit{FFmpeg: "ffmpeg", TrackName: "My Bot"}).Connect(ctx, voice.Grant{
//		Endpoint: update.Endpoint,
//		Token:    update.Token,
//	})
//	if err != nil {
//		return err
//	}
//	defer room.Close()
//
//	track, err := room.OpenTrack(ctx)
//	if err != nil {
//		return err
//	}
//	defer track.Close()
//
// Write signed 16-bit little-endian PCM at 48 kHz stereo to the track. Each
// frame carries up to 20 ms, or 1,920 samples across both channels. FFmpeg
// encodes it to Opus and publishes a microphone track. LiveKit.Bitrate sets
// the Opus bitrate (default [DefaultBitrate]); pass the voice channel's own
// bitrate to match it.
//
// # Playing files and URLs
//
// [OpenSource] starts FFmpeg and returns a reader of PCM from a file path or
// an http(s) URL. [NewPlayer] reads any io.Reader, paces it in 20 ms frames
// and writes it to the track:
//
//	source, err := voice.OpenSource(ctx, url, voice.SourceOptions{})
//	if err != nil {
//		return err
//	}
//	defer source.Close()
//
//	player := voice.NewPlayer(track)
//	player.SetVolume(0.8)
//	err = player.Play(ctx, source)
//
// A [Player] supports Pause, Resume, Skip and SetVolume while Play runs. Play
// returns [ErrSkipped] after a skip and [ErrBusy] if the player is already
// playing. [ScaleVolume] applies the same saturating gain to frames you
// manage yourself.
//
// # Limits
//
// Grants are valid for 600 seconds and are only needed to connect; Micup
// sends a new grant when it moves the connection, and the session reports it
// through SessionOptions.OnVoiceGrant. Channels with encrypted call media
// need an E2EE-capable client and are not supported, and bots never join
// peer-to-peer calls. The package publishes outgoing audio only; it does not
// receive or decode other participants. Queues, track metadata and what to
// play next stay with the application.
//
// FFmpeg with libopus is required; check that ffmpeg -encoders lists libopus.
package voice
