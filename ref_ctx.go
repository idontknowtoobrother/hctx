package hctx

import (
	"context"
	"sync"
)

var (
	// Mutex to protect global variables
	refMutex sync.RWMutex

	// Default values
	RefIdContext = NewHContextType[string]("X-Correlation-Id")
	RefHeaderKey = RefIdContext.Key()
)

// NewRefContext updates the global reference ID context with a custom key
func NewRefContext(customRefIdKey string) HContext[string] {
	refMutex.Lock()
	defer refMutex.Unlock()

	if customRefIdKey == "" {
		RefIdContext = NewHContextType[string]("X-Correlation-Id")
	} else {
		RefIdContext = NewHContextType[string](customRefIdKey)
	}

	RefHeaderKey = RefIdContext.Key()
	return RefIdContext
}

// ClaimsReferenceId safely retrieves the reference ID from context
func ClaimsReferenceId(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}

	refMutex.RLock()
	defer refMutex.RUnlock()

	// Get the reference ID from context
	return GetFromContext(ctx, RefIdContext)
}

// NewReferenceIdContext creates a new context with the reference ID
func NewReferenceIdContext(ctx context.Context, refID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	refMutex.RLock()
	defer refMutex.RUnlock()

	return WithValue(ctx, RefIdContext, refID)
}
