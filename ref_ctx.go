package hctx

import (
	"context"
	"sync"
)

var (
	// Mutex to protect global variables
	refMutex sync.RWMutex

	// Default values
	RefIdContext = NewGContextType[string]("X-Correlation-Id")
	RefHeaderKey = RefIdContext.Key()
)

// NewRefContext updates the global reference ID context with a custom key
func NewRefContext(customRefIdKey string) HContext[string] {
	refMutex.Lock()
	defer refMutex.Unlock()

	if customRefIdKey == "" {
		RefIdContext = NewGContextType[string]("X-Correlation-Id")
	} else {
		RefIdContext = NewGContextType[string](customRefIdKey)
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
	return Claims(ctx, RefIdContext)
}

// NewReferenceIdContext creates a new context with the reference ID
func NewReferenceIdContext(ctx context.Context, refID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	refMutex.RLock()
	defer refMutex.RUnlock()

	return NewGContext(ctx, RefIdContext, refID)
}
