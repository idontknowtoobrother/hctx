package hctx

import (
	"context"
	"net/http"
)

// * HContext is a generic context key with type parameter
type HContext[T any] struct {
	key string
}

// * NewGContextType creates a new generic context with key and specific type
func NewGContextType[T any](key string) HContext[T] {
	return HContext[T]{key: key}
}

// * Key returns the string key of the context
func (c HContext[T]) Key() string {
	return c.key
}

// * Claims retrieves a value from context with the given key and type
func Claims[T any](ctx context.Context, gc HContext[T]) (T, bool) {
	if ctx == nil {
		var zero T
		return zero, false
	}
	value, ok := ctx.Value(gc).(T)
	return value, ok
}

// * NewGContext adds a value to context with the given key and type
func NewGContext[T any](ctx context.Context, key HContext[T], data T) context.Context {
	return context.WithValue(ctx, key, data)
}

// * ToHttpRequest adds a value to http request context with the given key and type
func ToHttpRequest[T any](c *http.Request, key HContext[T], data T) *http.Request {
	return c.WithContext(NewGContext(c.Context(), key, data))
}
