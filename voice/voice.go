package voice

import (
	"context"
	"io"

	"github.com/yestefir/micup.go/src/voice"
)

const (
	SampleRate     = voice.SampleRate
	Channels       = voice.Channels
	FrameDuration  = voice.FrameDuration
	FrameSamples   = voice.FrameSamples
	DefaultBitrate = voice.DefaultBitrate
	MinBitrate     = voice.MinBitrate
	MaxBitrate     = voice.MaxBitrate
	MinVolume      = voice.MinVolume
	MaxVolume      = voice.MaxVolume
)

var (
	ErrSkipped = voice.ErrSkipped
	ErrBusy    = voice.ErrBusy

	ScaleVolume = voice.ScaleVolume
)

type (
	Grant         = voice.Grant
	Room          = voice.Room
	Track         = voice.Track
	LiveKit       = voice.LiveKit
	Player        = voice.Player
	SourceOptions = voice.SourceOptions
)

func NewPlayer(track Track) *Player { return voice.NewPlayer(track) }

func OpenSource(ctx context.Context, input string, options SourceOptions) (io.ReadCloser, error) {
	return voice.OpenSource(ctx, input, options)
}
