package voice

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"sync"
	"time"
)

const (
	MinVolume = 0.0
	MaxVolume = 2.0
)

var (
	ErrSkipped = errors.New("playback skipped")
	ErrBusy    = errors.New("player is already playing")
)

type Player struct {
	track    Track
	interval time.Duration
	mu       sync.Mutex
	volume   float64
	paused   bool
	wake     chan struct{}
	skip     chan struct{}
	playing  bool
}

func NewPlayer(track Track) *Player {
	return &Player{track: track, interval: FrameDuration, volume: 1, wake: make(chan struct{})}
}

func (p *Player) Volume() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.volume
}

func (p *Player) SetVolume(volume float64) float64 {
	if math.IsNaN(volume) {
		volume = 1
	}
	volume = min(max(volume, MinVolume), MaxVolume)
	p.mu.Lock()
	p.volume = volume
	p.mu.Unlock()
	return volume
}

func (p *Player) Paused() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.paused
}

func (p *Player) Pause() {
	p.mu.Lock()
	p.paused = true
	p.mu.Unlock()
}

func (p *Player) Resume() {
	p.mu.Lock()
	if p.paused {
		p.paused = false
		close(p.wake)
		p.wake = make(chan struct{})
	}
	p.mu.Unlock()
}

func (p *Player) Skip() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.skip == nil {
		return false
	}
	close(p.skip)
	p.skip = nil
	return true
}

func (p *Player) Playing() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.playing
}

func (p *Player) begin() (chan struct{}, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.playing {
		return nil, ErrBusy
	}
	p.playing = true
	p.skip = make(chan struct{})
	return p.skip, nil
}

func (p *Player) end() {
	p.mu.Lock()
	p.playing = false
	p.skip = nil
	p.mu.Unlock()
}

func (p *Player) state() (volume float64, paused bool, wake chan struct{}) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.volume, p.paused, p.wake
}

func (p *Player) Play(ctx context.Context, pcm io.Reader) error {
	skip, err := p.begin()
	if err != nil {
		return err
	}
	defer p.end()
	buffer := make([]byte, FrameSamples*2)
	frame := make([]int16, FrameSamples)
	start := time.Now()
	sent := 0
	for {
		if err := checkStop(ctx, skip); err != nil {
			return err
		}
		volume, paused, wake := p.state()
		if paused {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-skip:
				return ErrSkipped
			case <-wake:
			}
			start, sent = time.Now(), 0
			continue
		}
		n, readErr := io.ReadFull(pcm, buffer)
		samples := n / (2 * Channels) * Channels
		if samples > 0 {
			for i := range samples {
				frame[i] = int16(binary.LittleEndian.Uint16(buffer[i*2:]))
			}
			ScaleVolume(frame[:samples], volume)
			if err := p.track.Write(ctx, frame[:samples]); err != nil {
				return err
			}
			sent++
			if err := p.wait(ctx, skip, start.Add(time.Duration(sent)*p.interval)); err != nil {
				return err
			}
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) || errors.Is(readErr, io.ErrUnexpectedEOF) {
				return nil
			}
			return readErr
		}
	}
}

func (p *Player) wait(ctx context.Context, skip <-chan struct{}, until time.Time) error {
	delay := time.Until(until)
	if delay <= 0 {
		return checkStop(ctx, skip)
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-skip:
		return ErrSkipped
	case <-timer.C:
		return nil
	}
}

func checkStop(ctx context.Context, skip <-chan struct{}) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	select {
	case <-skip:
		return ErrSkipped
	default:
		return nil
	}
}

func ScaleVolume(frame []int16, factor float64) {
	if factor == 1 {
		return
	}
	for i, sample := range frame {
		frame[i] = int16(min(max(math.Round(float64(sample)*factor), math.MinInt16), math.MaxInt16))
	}
}
