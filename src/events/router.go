package events

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/yestefir/micup.go/src/model"
)

var ErrNoRoute = errors.New("no route for interaction")

type InteractionHandler func(context.Context, model.Interaction) error

type routeTable struct {
	exact    map[string]InteractionHandler
	prefixes []prefixRoute
}

type prefixRoute struct {
	prefix  string
	handler InteractionHandler
}

func (t *routeTable) add(key string, handler InteractionHandler) {
	if handler == nil {
		panic("events: route handler must not be nil")
	}
	if prefix, ok := strings.CutSuffix(key, "*"); ok {
		t.prefixes = append(t.prefixes, prefixRoute{prefix: prefix, handler: handler})
		return
	}
	if t.exact == nil {
		t.exact = map[string]InteractionHandler{}
	}
	t.exact[key] = handler
}

func (t *routeTable) find(key string) InteractionHandler {
	if handler, ok := t.exact[key]; ok {
		return handler
	}
	var best InteractionHandler
	longest := -1
	for _, route := range t.prefixes {
		if strings.HasPrefix(key, route.prefix) && len(route.prefix) > longest {
			best, longest = route.handler, len(route.prefix)
		}
	}
	return best
}

type Router struct {
	mu           sync.RWMutex
	commands     routeTable
	autocomplete routeTable
	components   routeTable
	modals       routeTable
}

func NewRouter() *Router { return &Router{} }

func (r *Router) Command(path string, handler InteractionHandler) {
	r.register(&r.commands, path, handler)
}

func (r *Router) Autocomplete(path string, handler InteractionHandler) {
	r.register(&r.autocomplete, path, handler)
}

func (r *Router) Component(customID string, handler InteractionHandler) {
	r.register(&r.components, customID, handler)
}

func (r *Router) Modal(customID string, handler InteractionHandler) {
	r.register(&r.modals, customID, handler)
}

func (r *Router) register(table *routeTable, key string, handler InteractionHandler) {
	if key == "" {
		panic("events: route key must not be empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	table.add(key, handler)
}

func (r *Router) Handle(ctx context.Context, interaction model.Interaction) error {
	handler := r.lookup(interaction)
	if handler == nil {
		return ErrNoRoute
	}
	return handler(ctx, interaction)
}

func (r *Router) lookup(interaction model.Interaction) InteractionHandler {
	r.mu.RLock()
	defer r.mu.RUnlock()
	switch interaction.Type {
	case model.InteractionTypeApplicationCommand, model.InteractionTypeAutocomplete:
		data, ok := interaction.CommandData()
		if !ok {
			return nil
		}
		table := &r.commands
		if interaction.IsAutocomplete() {
			table = &r.autocomplete
		}
		return table.find(strings.Join(data.Path(), " "))
	case model.InteractionTypeMessageComponent:
		data, ok := interaction.ComponentData()
		if !ok {
			return nil
		}
		return r.components.find(data.CustomID)
	case model.InteractionTypeModalSubmit:
		data, ok := interaction.ModalData()
		if !ok {
			return nil
		}
		return r.modals.find(data.CustomID)
	}
	return nil
}

func (r *Router) Listener() Listener {
	return Listen(func(ctx context.Context, interaction model.Interaction) error {
		if err := r.Handle(ctx, interaction); !errors.Is(err, ErrNoRoute) {
			return err
		}
		return nil
	})
}
