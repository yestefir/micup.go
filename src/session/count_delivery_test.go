package session

import (
	"bytes"
	"context"
	"sync"
	"testing"

	"github.com/yestefir/micup.go/src/gateway"
)

func TestCountResponseOwnsQueuedEventData(t *testing.T) {
	answer := make(chan []byte, 1)
	s := &Session{countRequests: map[string]chan []byte{"request": answer}}
	data := []byte(`{"nonce":"request","counts":[]}`)
	want := bytes.Clone(data)
	s.handleCounts(gateway.Event{Data: data})
	data[0] = 'x'
	if got := <-answer; !bytes.Equal(got, want) {
		t.Fatal("editing event data changed the queued count response")
	}
}

func TestSessionShutdownClosesEmptyAndBufferedCountAnswers(t *testing.T) {
	empty, full := make(chan []byte, 1), make(chan []byte, 1)
	full <- []byte(`{"counts":[]}`)
	s := &Session{countRequests: map[string]chan []byte{"empty": empty, "full": full}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := s.Run(ctx); err != nil {
		t.Fatal(err)
	}
	for name, answer := range map[string]chan []byte{"empty": empty, "full": full} {
		select {
		case data, open := <-answer:
			if open || data != nil {
				t.Errorf("%s answer kept a buffered result instead of closing: %s", name, data)
			}
		default:
			t.Errorf("%s answer received no shutdown signal", name)
		}
	}
	if len(s.countRequests) != 0 {
		t.Error("shutdown retained count request registrations")
	}
}

func TestConcurrentCountDeliveryAndShutdown(t *testing.T) {
	for range 100 {
		answer := make(chan []byte, 1)
		s := &Session{countRequests: map[string]chan []byte{"request": answer}}
		start := make(chan struct{})
		var work sync.WaitGroup
		work.Add(2)
		go func() {
			defer work.Done()
			<-start
			for range 10 {
				s.handleCounts(gateway.Event{Data: []byte(`{"nonce":"request","counts":[]}`)})
			}
		}()
		go func() {
			defer work.Done()
			<-start
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			if err := s.Run(ctx); err != nil {
				t.Error(err)
			}
		}()
		close(start)
		work.Wait()
		if _, open := <-answer; open || len(s.countRequests) != 0 {
			t.Fatal("concurrent delivery left a pending answer after shutdown")
		}
	}
}

var benchmarkCountResponse []byte

func BenchmarkCountResponseDelivery(b *testing.B) {
	answer := make(chan []byte, 1)
	s := &Session{countRequests: map[string]chan []byte{"request": answer}}
	event := gateway.Event{Data: []byte(`{"nonce":"request","counts":[{"guild_id":"g","member_count":10,"online_count":3}]}`)}
	b.ReportAllocs()
	for b.Loop() {
		s.handleCounts(event)
		benchmarkCountResponse = <-answer
	}
}
