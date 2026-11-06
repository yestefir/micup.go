package events

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/yestefir/micup.go/src/gateway"
)

func TestMuxDispatchesTypedListeners(t *testing.T) {
	m := NewMux()
	var got MessageDelete
	var all []string
	m.On(gateway.EventMessageDelete, Listen(func(_ context.Context, e MessageDelete) error {
		got = e
		return nil
	}))
	m.HandleAll(func(_ context.Context, e gateway.Event) error {
		all = append(all, e.Name)
		return nil
	})
	ctx := context.Background()
	if err := m.Handle(ctx, gateway.Event{Name: gateway.EventMessageDelete, Data: json.RawMessage(`{"id":"m","channel_id":"c"}`)}); err != nil {
		t.Fatal(err)
	}
	if err := m.Handle(ctx, gateway.Event{Name: "SOMETHING_NEW", Data: json.RawMessage(`{}`)}); err != nil {
		t.Fatal(err)
	}
	if got.ID != "m" || got.ChannelID != "c" || len(all) != 2 || all[1] != "SOMETHING_NEW" {
		t.Fatalf("dispatch: %+v %v", got, all)
	}
}

func TestMuxJoinsErrors(t *testing.T) {
	m := NewMux()
	first, second := errors.New("first"), errors.New("second")
	m.HandleRaw(gateway.EventReady, func(context.Context, gateway.Event) error { return first })
	m.HandleRaw(gateway.EventReady, func(context.Context, gateway.Event) error { return second })
	err := m.Handle(context.Background(), gateway.Event{Name: gateway.EventReady})
	if !errors.Is(err, first) || !errors.Is(err, second) {
		t.Fatalf("errors were not joined: %v", err)
	}
}

func TestMuxRejectsMismatchedListener(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("mismatched listener accepted")
		}
	}()
	NewMux().On(gateway.EventMessageCreate, Listen(func(context.Context, MessageDelete) error { return nil }))
}

func TestMuxRejectsZeroListener(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("zero listener accepted")
		}
	}()
	NewMux().On(gateway.EventReady, Listener{})
}
