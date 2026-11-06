package events

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"

	"github.com/yestefir/micup.go/src/gateway"
)

type payload struct {
	typ    reflect.Type
	decode func(json.RawMessage) (any, error)
}

func of[T any]() payload {
	return payload{
		typ: reflect.TypeFor[T](),
		decode: func(data json.RawMessage) (any, error) {
			var value T
			err := json.Unmarshal(data, &value)
			return value, err
		},
	}
}

// Decode returns the typed payload for a known event and the raw JSON for any other name.
func Decode(event gateway.Event) (any, error) {
	p, ok := registry[event.Name]
	if !ok {
		return event.Data, nil
	}
	value, err := p.decode(event.Data)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", event.Name, err)
	}
	return value, nil
}

func As[T any](event gateway.Event) (T, error) {
	var value T
	if p, ok := registry[event.Name]; ok && p.typ != reflect.TypeFor[T]() {
		return value, fmt.Errorf("%s carries %s, not %s", event.Name, p.typ, reflect.TypeFor[T]())
	}
	if err := json.Unmarshal(event.Data, &value); err != nil {
		return value, fmt.Errorf("decode %s: %w", event.Name, err)
	}
	return value, nil
}

// Parse decodes like json.Unmarshal but keeps the fields it could decode when a value has an unexpected type.
func Parse[T any](data json.RawMessage) (T, bool) {
	var value T
	err := json.Unmarshal(data, &value)
	var typeErr *json.UnmarshalTypeError
	return value, err == nil || errors.As(err, &typeErr)
}

func PayloadType(name string) (reflect.Type, bool) {
	p, ok := registry[name]
	return p.typ, ok
}

func Names() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
