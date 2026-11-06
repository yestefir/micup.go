package events

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/yestefir/micup.go/src/gateway"
)

type Listener struct {
	typ    reflect.Type
	handle gateway.Handler
}

func Listen[T any](fn func(context.Context, T) error) Listener {
	return Listener{
		typ: reflect.TypeFor[T](),
		handle: func(ctx context.Context, event gateway.Event) error {
			value, err := As[T](event)
			if err != nil {
				return err
			}
			return fn(ctx, value)
		},
	}
}

type Mux struct {
	mu       sync.RWMutex
	byName   map[string][]gateway.Handler
	fallback []gateway.Handler
}

func NewMux() *Mux { return &Mux{byName: map[string][]gateway.Handler{}} }

func (m *Mux) On(name string, listener Listener) {
	if listener.handle == nil {
		panic("events: Listener must be built with Listen")
	}
	if want, ok := PayloadType(name); ok && want != listener.typ {
		panic(fmt.Sprintf("events: %s carries %s, listener takes %s", name, want, listener.typ))
	}
	m.HandleRaw(name, listener.handle)
}

func (m *Mux) HandleRaw(name string, handler gateway.Handler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.byName[name] = append(m.byName[name], handler)
}

func (m *Mux) HandleAll(handler gateway.Handler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fallback = append(m.fallback, handler)
}

func (m *Mux) Handle(ctx context.Context, event gateway.Event) error {
	m.mu.RLock()
	handlers := append(append([]gateway.Handler(nil), m.byName[event.Name]...), m.fallback...)
	m.mu.RUnlock()
	var errs []error
	for _, handler := range handlers {
		if err := handler(ctx, event); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
