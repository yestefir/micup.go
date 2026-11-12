package model

import (
	"encoding/base64"
	"errors"
)

type Optional[T any] struct {
	value T
	set   bool
	null  bool
}

func Set[T any](value T) Optional[T] { return Optional[T]{value: value, set: true} }

func Clear[T any]() Optional[T] { return Optional[T]{set: true, null: true} }

func (o Optional[T]) IsSet() bool { return o.set }

func (o Optional[T]) IsNull() bool { return o.set && o.null }

func (o Optional[T]) Value() (T, bool) { return o.value, o.set && !o.null }

func (o Optional[T]) Apply(body map[string]any, key string) {
	switch {
	case !o.set:
	case o.null:
		body[key] = nil
	default:
		body[key] = o.value
	}
}

func ImageData(contentType string, data []byte) (string, error) {
	if contentType == "" || len(data) == 0 {
		return "", errors.New("image data needs a content type and bytes")
	}
	return "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(data), nil
}
