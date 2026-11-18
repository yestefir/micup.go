package voice

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"sync"
	"testing"
	"time"
)

type recordingTrack struct {
	mu     sync.Mutex
	frames [][]int16
	got    chan struct{}
}

func newRecordingTrack() *recordingTrack { return &recordingTrack{got: make(chan struct{}, 1024)} }

func (r *recordingTrack) Write(_ context.Context, frame []int16) error {
	r.mu.Lock()
	r.frames = append(r.frames, append([]int16(nil), frame...))
	r.mu.Unlock()
	r.got <- struct{}{}
	return nil
}

func (r *recordingTrack) Close() error { return nil }

func (r *recordingTrack) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.frames)
}

func pcmBytes(samples int, value int16) []byte {
	data := make([]byte, samples*2)
	for i := range samples {
		binary.LittleEndian.PutUint16(data[i*2:], uint16(value))
	}
	return data
}

func fastPlayer(track Track) *Player {
	p := NewPlayer(track)
	p.interval = time.Millisecond
	return p
}

func TestPlayerSendsWholeAndPartialFrames(t *testing.T) {
	track := newRecordingTrack()
	data := pcmBytes(FrameSamples*2+10, 100)
	if err := fastPlayer(track).Play(context.Background(), bytes.NewReader(data)); err != nil {
		t.Fatal(err)
	}
	if len(track.frames) != 3 || len(track.frames[0]) != FrameSamples || len(track.frames[2]) != 10 {
		t.Fatalf("got %d frames", len(track.frames))
	}
}

func TestPlayerDropsTrailingHalfSample(t *testing.T) {
	track := newRecordingTrack()
	data := append(pcmBytes(FrameSamples, 5), 1, 2, 3)
	if err := fastPlayer(track).Play(context.Background(), bytes.NewReader(data)); err != nil {
		t.Fatal(err)
	}
	if len(track.frames) != 1 {
		t.Fatalf("frames = %d, want 1", len(track.frames))
	}
}

func TestPlayerAppliesVolume(t *testing.T) {
	track := newRecordingTrack()
	p := fastPlayer(track)
	if got := p.SetVolume(0.5); got != 0.5 {
		t.Fatalf("volume = %v", got)
	}
	if err := p.Play(context.Background(), bytes.NewReader(pcmBytes(FrameSamples, 1000))); err != nil {
		t.Fatal(err)
	}
	if track.frames[0][0] != 500 {
		t.Fatalf("sample = %d, want 500", track.frames[0][0])
	}
	if p.SetVolume(99) != MaxVolume || p.SetVolume(-1) != MinVolume {
		t.Fatal("volume was not clamped")
	}
}

func TestScaleVolumeSaturates(t *testing.T) {
	frame := []int16{30000, -30000, 10}
	ScaleVolume(frame, 2)
	if frame[0] != 32767 || frame[1] != -32768 || frame[2] != 20 {
		t.Fatalf("frame = %v", frame)
	}
}

func TestPlayerPauseResumeAndSkip(t *testing.T) {
	track := newRecordingTrack()
	p := fastPlayer(track)
	p.Pause()
	done := make(chan error, 1)
	go func() { done <- p.Play(context.Background(), bytes.NewReader(pcmBytes(FrameSamples*3, 1))) }()
	select {
	case <-track.got:
		t.Fatal("paused player wrote audio")
	case <-time.After(50 * time.Millisecond):
	}
	if !p.Paused() || !p.Playing() {
		t.Fatal("player state is wrong while paused")
	}
	p.Resume()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if track.count() != 3 {
		t.Fatalf("frames = %d, want 3", track.count())
	}

	p.Pause()
	go func() { done <- p.Play(context.Background(), bytes.NewReader(pcmBytes(FrameSamples, 1))) }()
	for !p.Playing() {
		time.Sleep(time.Millisecond)
	}
	if !p.Skip() {
		t.Fatal("skip reported no active track")
	}
	if err := <-done; !errors.Is(err, ErrSkipped) {
		t.Fatalf("play error = %v, want ErrSkipped", err)
	}
	if p.Skip() {
		t.Fatal("skip succeeded with nothing playing")
	}
}

func TestPlayerRejectsConcurrentPlayAndHonorsContext(t *testing.T) {
	track := newRecordingTrack()
	p := fastPlayer(track)
	p.Pause()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- p.Play(ctx, bytes.NewReader(pcmBytes(FrameSamples, 1))) }()
	for !p.Playing() {
		time.Sleep(time.Millisecond)
	}
	if err := p.Play(context.Background(), bytes.NewReader(nil)); !errors.Is(err, ErrBusy) {
		t.Fatalf("second play error = %v", err)
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("play error = %v", err)
	}
	if p.Playing() {
		t.Fatal("player still marked as playing")
	}
}

func TestOpenSourceRejectsUnsafeInput(t *testing.T) {
	for _, input := range []string{"", "-f", "--version"} {
		if _, err := OpenSource(context.Background(), input, SourceOptions{}); err == nil {
			t.Errorf("input %q was accepted", input)
		}
	}
	if _, err := OpenSource(context.Background(), "a.mp3", SourceOptions{Start: -time.Second}); err == nil {
		t.Error("negative start was accepted")
	}
	if _, err := OpenSource(context.Background(), "a.mp3", SourceOptions{FFmpeg: "definitely-not-installed-ffmpeg"}); err == nil {
		t.Error("missing decoder was accepted")
	}
}
