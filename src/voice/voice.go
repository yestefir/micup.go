package voice

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"log/slog"
	"net/url"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/livekit/server-sdk-go/v2/pkg/oggreader"
	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
)

const SampleRate = 48000
const Channels = 2
const FrameDuration = 20 * time.Millisecond
const FrameSamples = SampleRate * Channels / 50

const (
	DefaultBitrate = 96000
	MinBitrate     = 8000
	MaxBitrate     = 512000
)

type Grant struct {
	Endpoint string
	Token    string
}

type Room interface {
	Connected() bool
	OpenTrack(context.Context) (Track, error)
	Close()
}

type Track interface {
	Write(context.Context, []int16) error
	Close() error
}

type LiveKit struct {
	FFmpeg    string
	TrackName string
	Bitrate   int
}

type room struct {
	client    *lksdk.Room
	ffmpeg    string
	trackName string
	bitrate   int
	connected atomic.Bool
	once      sync.Once
}

func (l LiveKit) Connect(ctx context.Context, grant Grant) (Room, error) {
	endpoint := grant.Endpoint
	if !strings.Contains(endpoint, "://") {
		endpoint = "wss://" + endpoint
	}
	u, err := url.Parse(endpoint)
	if err != nil || (u.Scheme != "wss" && u.Scheme != "ws") || u.User != nil || u.Host == "" || grant.Token == "" {
		return nil, errors.New("invalid LiveKit grant")
	}
	timeout := 10 * time.Second
	if deadline, ok := ctx.Deadline(); ok {
		timeout = min(timeout, time.Until(deadline))
	}
	if timeout <= 0 {
		return nil, context.DeadlineExceeded
	}
	if l.FFmpeg == "" {
		l.FFmpeg = "ffmpeg"
	}
	if l.TrackName == "" {
		l.TrackName = "micup.go"
	}
	r := &room{ffmpeg: l.FFmpeg, trackName: l.TrackName, bitrate: clampBitrate(l.Bitrate)}
	client, err := lksdk.ConnectToRoomWithToken(endpoint, grant.Token, &lksdk.RoomCallback{OnDisconnected: func() { r.connected.Store(false) }}, lksdk.WithAutoSubscribe(false), lksdk.WithConnectTimeout(timeout))
	if err != nil {
		return nil, errors.New("LiveKit join failed")
	}
	if ctx.Err() != nil {
		client.Disconnect()
		return nil, ctx.Err()
	}
	r.client = client
	r.connected.Store(true)
	return r, nil
}

func (r *room) Connected() bool { return r.connected.Load() }
func (r *room) Close()          { r.once.Do(func() { r.connected.Store(false); r.client.Disconnect() }) }

func (r *room) OpenTrack(parent context.Context) (Track, error) {
	if !r.Connected() {
		return nil, errors.New("LiveKit room is disconnected")
	}
	track, err := lksdk.NewLocalSampleTrack(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus, ClockRate: SampleRate, Channels: Channels, SDPFmtpLine: "minptime=10;useinbandfec=1;stereo=1;sprop-stereo=1"})
	if err != nil {
		return nil, err
	}
	publication, err := r.client.LocalParticipant.PublishTrack(track, &lksdk.TrackPublicationOptions{Name: r.trackName, Source: livekit.TrackSource_MICROPHONE, Stereo: true, DisableDTX: true})
	if err != nil {
		_ = track.Close()
		return nil, errors.New("LiveKit track publish failed")
	}
	first := true
	t, err := openEncoder(parent, r.ffmpeg, r.bitrate, func(packet []byte) error {
		err := track.WriteSample(media.Sample{Data: packet, Duration: FrameDuration}, nil)
		if err == nil && first {
			first = false
			slog.Info("Micup audio packets started", "track_id", publication.SID())
		}
		return err
	}, r.Connected, func() { _ = track.Close(); _ = r.client.LocalParticipant.UnpublishTrack(publication.SID()) })
	if err != nil {
		_ = track.Close()
		_ = r.client.LocalParticipant.UnpublishTrack(publication.SID())
		return nil, err
	}
	return t, nil
}

func clampBitrate(bitrate int) int {
	if bitrate <= 0 {
		return DefaultBitrate
	}
	return min(max(bitrate, MinBitrate), MaxBitrate)
}

func openEncoder(parent context.Context, binaryPath string, bitrate int, writePacket func([]byte) error, connected func() bool, release func()) (*pcmTrack, error) {
	ctx, cancel := context.WithCancel(parent)
	cmd := exec.CommandContext(ctx, binaryPath, "-hide_banner", "-loglevel", "error", "-nostdin", "-f", "s16le", "-ar", "48000", "-ac", "2", "-probesize", "32", "-analyzeduration", "0", "-i", "pipe:0", "-c:a", "libopus", "-application", "audio", "-b:a", strconv.Itoa(clampBitrate(bitrate)), "-frame_duration", "20", "-page_duration", "20000", "-flush_packets", "1", "-f", "ogg", "pipe:1")
	configure(cmd)
	cmd.WaitDelay = time.Second
	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		_ = stdin.Close()
		return nil, err
	}
	cmd.Stderr = io.Discard
	if err = cmd.Start(); err != nil {
		cancel()
		_ = stdin.Close()
		_ = stdout.Close()
		return nil, errors.New("Opus encoder could not start")
	}
	t := &pcmTrack{ctx: ctx, cancel: cancel, cmd: cmd, stdin: stdin, stdout: stdout, connected: connected, release: release, done: make(chan struct{})}
	go func() {
		defer close(t.done)
		reader, _, err := oggreader.NewOggReader(stdout)
		if err == nil {
			for {
				packet, e := reader.ReadPacket()
				if e != nil {
					err = e
					break
				}
				if strings.HasPrefix(string(packet), "OpusTags") {
					continue
				}
				if e = writePacket(packet); e != nil {
					err = e
					break
				}
			}
		}
		if err != nil && !errors.Is(err, io.EOF) && ctx.Err() == nil {
			t.failed.Store(true)
		}
		cancel()
	}()
	return t, nil
}

type pcmTrack struct {
	ctx       context.Context
	cancel    context.CancelFunc
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	stdout    io.ReadCloser
	connected func() bool
	release   func()
	done      chan struct{}
	failed    atomic.Bool
	once      sync.Once
}

func (t *pcmTrack) Write(ctx context.Context, frame []int16) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(frame) == 0 || len(frame) > FrameSamples || len(frame)%Channels != 0 {
		return errors.New("PCM frame must contain up to 20 ms of stereo audio")
	}
	if !t.connected() || t.failed.Load() || t.ctx.Err() != nil {
		return errors.New("LiveKit audio stream ended")
	}
	stop := context.AfterFunc(ctx, func() { t.cancel(); _ = t.stdin.Close() })
	defer stop()
	data := make([]byte, len(frame)*2)
	for i, v := range frame {
		binary.LittleEndian.PutUint16(data[i*2:], uint16(v))
	}
	for len(data) > 0 {
		n, err := t.stdin.Write(data)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return errors.New("Opus encoder input failed")
		}
		if n == 0 {
			return io.ErrShortWrite
		}
		data = data[n:]
	}
	return nil
}

func (t *pcmTrack) Close() error {
	t.once.Do(func() {
		_ = t.stdin.Close()
		if t.ctx.Err() == nil {
			timer := time.NewTimer(time.Second)
			select {
			case <-t.done:
				timer.Stop()
			case <-timer.C:
			}
		}
		t.cancel()
		_ = t.stdout.Close()
		_ = t.cmd.Wait()
		<-t.done
		t.release()
	})
	return nil
}
