package voice

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

func TestPCMTrackEncodesOpus(t *testing.T) {
	path := os.Getenv("FFMPEG_TEST_BINARY")
	if path == "" {
		name := "ffmpeg"
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		found, err := exec.LookPath(name)
		if err != nil {
			t.Skip("FFmpeg is required for Opus encoder tests")
		}
		path = found
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var packets, released atomic.Int64
	track, err := openEncoder(ctx, path, 64000, func(packet []byte) error {
		if len(packet) > 0 {
			packets.Add(1)
		}
		return nil
	}, func() bool { return true }, func() { released.Add(1) })
	if err != nil {
		t.Fatal(err)
	}
	frame := make([]int16, FrameSamples)
	for i := range frame {
		frame[i] = int16(i % 2000)
	}
	for range 40 {
		if err = track.Write(ctx, frame); err != nil {
			t.Fatal(err)
		}
	}
	if err = track.Close(); err != nil {
		t.Fatal(err)
	}
	_ = track.Close()
	if packets.Load() == 0 || released.Load() != 1 {
		t.Fatalf("encoder produced %d packets and released %d tracks", packets.Load(), released.Load())
	}
}

func TestPCMTrackRejectsInvalidFrame(t *testing.T) {
	track := &pcmTrack{ctx: context.Background(), connected: func() bool { return true }}
	if err := track.Write(context.Background(), make([]int16, FrameSamples+1)); err == nil {
		t.Fatal("oversized PCM frame accepted")
	}
}

func TestClampBitrate(t *testing.T) {
	for in, want := range map[int]int{0: DefaultBitrate, -5: DefaultBitrate, 1000: MinBitrate, 64000: 64000, 900000: MaxBitrate} {
		if got := clampBitrate(in); got != want {
			t.Fatalf("clampBitrate(%d) = %d, want %d", in, got, want)
		}
	}
}
