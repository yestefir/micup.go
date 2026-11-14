package session

import (
	"context"
	"errors"
	"strconv"
	"testing"

	"github.com/yestefir/micup.go/src/gateway"
	"github.com/yestefir/micup.go/src/model"
)

func TestConsumedVoiceGrantCannotCommitAfterRunStops(t *testing.T) {
	for _, buffered := range []bool{false, true} {
		name := "empty_queue"
		if buffered {
			name = "buffered_queue"
		}
		t.Run(name, func(t *testing.T) {
			s, err := New(Config{Endpoint: "wss://localhost", Token: "fixture"}, Options{})
			if err != nil {
				t.Fatal(err)
			}
			pending, err := s.beginJoin("g")
			if err != nil {
				t.Fatal(err)
			}
			pending <- joinResult{grant: model.VoiceServerUpdate{GuildID: "g", ChannelID: "v", ConnectionID: "old", Endpoint: "wss://voice.invalid", Token: "fixture-grant"}}
			consumed := <-pending
			if buffered {
				pending <- consumed
			}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if err := s.Run(ctx); err != nil {
				t.Fatal(err)
			}
			if _, err := s.finishJoin("g", "v", pending, consumed); !errors.Is(err, gateway.ErrClosed) {
				t.Errorf("consumed grant committed after shutdown: %v", err)
			}
			if _, ok := s.Voice("g"); ok {
				t.Error("shutdown was followed by storing an old grant")
			}
			if _, err := s.beginJoin("g"); !errors.Is(err, ErrVoiceJoinPending) {
				t.Errorf("shutdown released the reservation before caller cleanup: %v", err)
			}
			s.endJoin("g", pending)
			fresh, err := s.beginJoin("g")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := s.finishJoin("g", "v", fresh, consumed); err != nil {
				t.Errorf("fresh reservation inherited an earlier shutdown: %v", err)
			}
			s.endJoin("g", fresh)
		})
	}
}

var benchmarkVoiceSession *Session

func BenchmarkVoiceJoinReservations(b *testing.B) {
	keys := make([]string, 512)
	for i := range keys {
		keys[i] = strconv.Itoa(i)
	}
	b.ReportAllocs()
	for b.Loop() {
		s, err := New(Config{Endpoint: "wss://localhost", Token: "fixture"}, Options{})
		if err != nil {
			b.Fatal(err)
		}
		for _, guild := range keys {
			if _, err := s.beginJoin(guild); err != nil {
				b.Fatal(err)
			}
		}
		benchmarkVoiceSession = s
	}
	b.ReportMetric(512, "reservations/op")
}
