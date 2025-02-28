package hctx

import (
	"context"
	"net/http"
)

// HContext is a generic context key with type parameter
type HContext[T any] struct {
	key string
}

// NewHContextType creates a new generic context with key and specific type
func NewHContextType[T any](key string) HContext[T] {
	return HContext[T]{key: key}
}

// Key returns the string key of the context
func (c HContext[T]) Key() string {
	return c.key
}

// GetFromContext retrieves a value from context with the given key and type
func GetFromContext[T any](ctx context.Context, gc HContext[T]) (T, bool) {
	if ctx == nil {
		var zero T
		return zero, false
	}
	value, ok := ctx.Value(gc).(T)
	return value, ok
}

// WithValue  adds a value to context with the given key and type
func WithValue[T any](ctx context.Context, key HContext[T], data T) context.Context {
	return context.WithValue(ctx, key, data)
}

// ToHttpRequest adds a value to http request context with the given key and type
func ToHttpRequest[T any](c *http.Request, key HContext[T], data T) *http.Request {
	return c.WithContext(WithValue(c.Context(), key, data))
}
