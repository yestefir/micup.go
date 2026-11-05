package voice

import "github.com/yestefir/micup.go/src/voice"

const (
	SampleRate     = voice.SampleRate
	Channels       = voice.Channels
	FrameDuration  = voice.FrameDuration
	FrameSamples   = voice.FrameSamples
	DefaultBitrate = voice.DefaultBitrate
	MinBitrate     = voice.MinBitrate
	MaxBitrate     = voice.MaxBitrate
)

type (
	Grant   = voice.Grant
	Room    = voice.Room
	Track   = voice.Track
	LiveKit = voice.LiveKit
)
