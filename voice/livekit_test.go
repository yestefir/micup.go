package voice

import (
	"context"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/livekit/protocol/auth"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/pion/webrtc/v4"
)

func reservePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()
	return port
}
func TestLocalLiveKitReceivesOpusAudio(t *testing.T) {
	serverBinary := os.Getenv("LIVEKIT_TEST_BINARY")
	if serverBinary == "" {
		t.Skip("set LIVEKIT_TEST_BINARY to verify a real local LiveKit server")
	}
	ffmpeg := os.Getenv("FFMPEG_TEST_BINARY")
	if ffmpeg == "" {
		t.Fatal("FFMPEG_TEST_BINARY is required for local LiveKit test")
	}
	port, rtcPort := reservePort(t), reservePort(t)
	udp, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	udpPort := udp.LocalAddr().(*net.UDPAddr).Port
	_ = udp.Close()
	log, err := os.Create(filepath.Join(t.TempDir(), "livekit.log"))
	if err != nil {
		t.Fatal(err)
	}
	defer log.Close()
	cmd := exec.Command(serverBinary, "--dev", "--bind", "127.0.0.1", "--node-ip", "127.0.0.1", "--port", strconv.Itoa(port), "--rtc.tcp_port", strconv.Itoa(rtcPort), "--udp-port", strconv.Itoa(udpPort), "--rtc.enable_loopback_candidate", "--keys", "devkey: fixture-secret-at-least-32-characters-long")
	configure(cmd)
	cmd.Stdout = log
	cmd.Stderr = log
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cmd.Process.Kill(); _ = cmd.Wait() }()
	address := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
	deadline := time.Now().Add(8 * time.Second)
	up := false
	for time.Now().Before(deadline) {
		c, e := net.DialTimeout("tcp", address, 100*time.Millisecond)
		if e == nil {
			_ = c.Close()
			up = true
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !up {
		_ = log.Sync()
		data, _ := os.ReadFile(log.Name())
		t.Fatalf("local LiveKit failed startup: %s", data)
	}
	token := func(identity string) string {
		v, e := auth.NewAccessToken("devkey", "fixture-secret-at-least-32-characters-long").SetIdentity(identity).SetValidFor(time.Minute).SetVideoGrant(&auth.VideoGrant{RoomJoin: true, Room: "micup-fixture"}).ToJWT()
		if e != nil {
			t.Fatal(e)
		}
		return v
	}
	var received atomic.Int64
	callbacks := &lksdk.RoomCallback{ParticipantCallback: lksdk.ParticipantCallback{OnTrackSubscribed: func(track *webrtc.TrackRemote, _ *lksdk.RemoteTrackPublication, _ *lksdk.RemoteParticipant) {
		go func() {
			for {
				packet, _, e := track.ReadRTP()
				if e != nil {
					return
				}
				if len(packet.Payload) > 0 {
					received.Add(1)
				}
			}
		}()
	}}}
	listener, err := lksdk.ConnectToRoomWithToken("ws://"+address, token("listener"), callbacks, lksdk.WithConnectTimeout(5*time.Second))
	if err != nil {
		t.Fatal("local listener join failed")
	}
	defer listener.Disconnect()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	room, err := (LiveKit{FFmpeg: ffmpeg}).Connect(ctx, Grant{Endpoint: "ws://" + address, Token: token("bot")})
	if err != nil {
		t.Fatal(err)
	}
	defer room.Close()
	sink, err := room.OpenTrack(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer sink.Close()
	frame := make([]int16, FrameSamples)
	for i := range frame {
		frame[i] = int16(i % 1000)
	}
	for range 100 {
		if err = sink.Write(ctx, frame); err != nil {
			t.Fatal(err)
		}
		time.Sleep(20 * time.Millisecond)
	}
	deadline = time.Now().Add(2 * time.Second)
	for received.Load() < 10 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if received.Load() < 10 {
		t.Fatalf("receiver did not get Opus packets: %d", received.Load())
	}
	_ = sink.Close()
	room.Close()
	if room.Connected() {
		t.Fatal("voice room stayed connected")
	}
	t.Logf("received %d Opus RTP packets", received.Load())
}
