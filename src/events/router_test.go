package events

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/yestefir/micup.go/src/gateway"
	"github.com/yestefir/micup.go/src/model"
)

func interactionOf(t *testing.T, kind int, data any) model.Interaction {
	t.Helper()
	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	return model.Interaction{Type: kind, Data: raw}
}

func eventOf(t *testing.T, name string, payload any) gateway.Event {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return gateway.Event{Name: name, Data: raw}
}

func TestRouterDispatchesByInteractionKind(t *testing.T) {
	var got []string
	record := func(name string) InteractionHandler {
		return func(context.Context, model.Interaction) error { got = append(got, name); return nil }
	}
	r := NewRouter()
	r.Command("ping", record("ping"))
	r.Command("queue add", record("queue add"))
	r.Autocomplete("queue add", record("queue add autocomplete"))
	r.Component("skip", record("skip"))
	r.Component("page:*", record("page"))
	r.Component("page:next:*", record("page next"))
	r.Modal("feedback", record("feedback"))

	subcommand := map[string]any{"name": "queue", "options": []map[string]any{{"type": model.OptionTypeSubcommand, "name": "add"}}}
	cases := []struct {
		name        string
		interaction model.Interaction
		want        string
	}{
		{"command", interactionOf(t, model.InteractionTypeApplicationCommand, map[string]any{"name": "ping"}), "ping"},
		{"subcommand", interactionOf(t, model.InteractionTypeApplicationCommand, subcommand), "queue add"},
		{"autocomplete", interactionOf(t, model.InteractionTypeAutocomplete, subcommand), "queue add autocomplete"},
		{"exact component", interactionOf(t, model.InteractionTypeMessageComponent, map[string]any{"custom_id": "skip"}), "skip"},
		{"prefix component", interactionOf(t, model.InteractionTypeMessageComponent, map[string]any{"custom_id": "page:3"}), "page"},
		{"longest prefix", interactionOf(t, model.InteractionTypeMessageComponent, map[string]any{"custom_id": "page:next:3"}), "page next"},
		{"modal", interactionOf(t, model.InteractionTypeModalSubmit, map[string]any{"custom_id": "feedback"}), "feedback"},
	}
	for _, tc := range cases {
		got = nil
		if err := r.Handle(context.Background(), tc.interaction); err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if len(got) != 1 || got[0] != tc.want {
			t.Errorf("%s routed to %v, want %q", tc.name, got, tc.want)
		}
	}
}

func TestRouterReportsMissingRoutesAndPropagatesErrors(t *testing.T) {
	r := NewRouter()
	boom := errors.New("boom")
	r.Command("fail", func(context.Context, model.Interaction) error { return boom })
	missing := interactionOf(t, model.InteractionTypeApplicationCommand, map[string]any{"name": "other"})
	if err := r.Handle(context.Background(), missing); !errors.Is(err, ErrNoRoute) {
		t.Fatalf("missing route error = %v", err)
	}
	failing := interactionOf(t, model.InteractionTypeApplicationCommand, map[string]any{"name": "fail"})
	if err := r.Handle(context.Background(), failing); !errors.Is(err, boom) {
		t.Fatalf("handler error = %v", err)
	}
	mux := NewMux()
	mux.On("INTERACTION_CREATE", r.Listener())
	if err := mux.Handle(context.Background(), eventOf(t, "INTERACTION_CREATE", missing)); err != nil {
		t.Fatalf("listener must ignore unrouted interactions: %v", err)
	}
	if err := mux.Handle(context.Background(), eventOf(t, "INTERACTION_CREATE", failing)); !errors.Is(err, boom) {
		t.Fatalf("listener error = %v", err)
	}
}

func TestRouterRejectsInvalidRoutes(t *testing.T) {
	r := NewRouter()
	for name, fn := range map[string]func(){
		"empty key":   func() { r.Command("", func(context.Context, model.Interaction) error { return nil }) },
		"nil handler": func() { r.Command("x", nil) },
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s did not panic", name)
				}
			}()
			fn()
		}()
	}
}
