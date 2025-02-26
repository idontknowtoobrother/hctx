package hctx

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGlobalSetRefIDHeaderKey(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	tests := []struct {
		name          string
		customRefKey  string
		wantHeaderKey string
	}{
		{
			name:          "custom_ref_id_key",
			customRefKey:  "Custom-Ref-ID",
			wantHeaderKey: "Custom-Ref-ID",
		},
		{
			name:          "empty_ref_id_key",
			customRefKey:  "",
			wantHeaderKey: "X-Correlation-Id", // Should keep default
		},
		{
			name:          "special_characters_in_key",
			customRefKey:  "X-Custom$Ref@ID",
			wantHeaderKey: "X-Custom$Ref@ID",
		},
		{
			name:          "lowercase_key",
			customRefKey:  "x-correlation-id",
			wantHeaderKey: "x-correlation-id",
		},
	}

	// Store original values to restore after tests
	originalRefIdContext := RefIdContext
	originalRefHeaderKey := RefHeaderKey
	defer func() {
		RefIdContext = originalRefIdContext
		RefHeaderKey = originalRefHeaderKey
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset to default before each test
			RefIdContext = NewGContextType[string]("X-Correlation-Id")
			RefHeaderKey = RefIdContext.Key()

			// Use unique path for each test case
			path := fmt.Sprintf("/test-ref-key-%s", tt.name)
			router.GET(path, func(c *gin.Context) {
				got := NewRefContext(tt.customRefKey)
				assert.Equal(t, tt.wantHeaderKey, got.Key(), "Header key should match")
				assert.Equal(t, tt.wantHeaderKey, RefHeaderKey, "Global header key should match")

				// Test context with new key
				newCtx := NewReferenceIdContext(c.Request.Context(), "test-value")
				c.Request = c.Request.WithContext(newCtx)

				// Verify the value can be retrieved with new key
				val, ok := ClaimsReferenceId(c.Request.Context())
				assert.Equal(t, "test-value", val, "Should retrieve value with new key")
				assert.True(t, ok, "Should retrieve value with new key")

				c.Status(http.StatusOK)
			})

			// Make request with unique path
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestGetRefIDFromContext(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	tests := []struct {
		name      string
		refID     string
		wantID    string
		setInCtx  bool
		nilCtx    bool
		customKey string
	}{
		{
			name:     "existing_ref_id",
			refID:    "test-ref-id",
			wantID:   "test-ref-id",
			setInCtx: true,
		},
		{
			name:     "missing_ref_id",
			refID:    "",
			wantID:   "",
			setInCtx: false,
		},
		{
			name:   "nil_context",
			refID:  "test-ref-id",
			wantID: "",
			nilCtx: true,
		},
		{
			name:      "custom_key_with_value",
			refID:     "custom-ref-id",
			wantID:    "custom-ref-id",
			setInCtx:  true,
			customKey: "X-Custom-Ref",
		},
		{
			name:     "empty_string_ref_id",
			refID:    "",
			wantID:   "",
			setInCtx: true,
		},
		{
			name:     "special_characters_in_ref_id",
			refID:    "ref@id#123$%^",
			wantID:   "ref@id#123$%^",
			setInCtx: true,
		},
	}

	// Store original values to restore after tests
	originalRefIdContext := RefIdContext
	originalRefHeaderKey := RefHeaderKey
	defer func() {
		RefIdContext = originalRefIdContext
		RefHeaderKey = originalRefHeaderKey
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset to default before each test
			RefIdContext = NewGContextType[string]("X-Correlation-Id")
			RefHeaderKey = RefIdContext.Key()

			// Use unique path for each test case
			path := fmt.Sprintf("/test-get-ref-%s", tt.name)
			router.GET(path, func(c *gin.Context) {
				if tt.customKey != "" {
					NewRefContext(tt.customKey)
				}

				var ctx context.Context
				if tt.nilCtx {
					ctx = nil
				} else {
					ctx = c.Request.Context()
					if tt.setInCtx {
						ctx = NewReferenceIdContext(ctx, tt.refID)
						c.Request = c.Request.WithContext(ctx)
					}
				}

				got, ok := ClaimsReferenceId(ctx)

				if tt.nilCtx || !tt.setInCtx {
					assert.Equal(t, tt.wantID, got, "RefID should match")
					assert.False(t, ok, "Should not retrieve value")
				} else {
					assert.Equal(t, tt.wantID, got, "RefID should match")
					assert.True(t, ok, "Should retrieve value")
				}

				// Test header consistency
				if tt.setInCtx && !tt.nilCtx {
					// Set header after getting from context
					c.Header(RefHeaderKey, got)
				}

				c.Status(http.StatusOK)
			})

			// Make request with unique path
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			if tt.setInCtx {
				assert.Equal(t, tt.wantID, w.Header().Get(RefHeaderKey), "Header value should match context value")
			}
		})
	}
}

func TestNewRefContext(t *testing.T) {
	// Store original values to restore after tests
	originalRefIdContext := RefIdContext
	originalRefHeaderKey := RefHeaderKey
	defer func() {
		RefIdContext = originalRefIdContext
		RefHeaderKey = originalRefHeaderKey
	}()

	tests := []struct {
		name          string
		customRefKey  string
		wantHeaderKey string
	}{
		{
			name:          "custom_ref_id_key",
			customRefKey:  "Custom-Ref-ID",
			wantHeaderKey: "Custom-Ref-ID",
		},
		{
			name:          "empty_ref_id_key",
			customRefKey:  "",
			wantHeaderKey: "X-Correlation-Id", // Should keep default
		},
		{
			name:          "special_characters_in_key",
			customRefKey:  "X-Custom$Ref@ID",
			wantHeaderKey: "X-Custom$Ref@ID",
		},
		{
			name:          "lowercase_key",
			customRefKey:  "x-correlation-id",
			wantHeaderKey: "x-correlation-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset to default before each test
			RefIdContext = NewGContextType[string]("X-Correlation-Id")
			RefHeaderKey = RefIdContext.Key()

			got := NewRefContext(tt.customRefKey)
			assert.Equal(t, tt.wantHeaderKey, got.Key(), "Header key should match")
			assert.Equal(t, tt.wantHeaderKey, RefHeaderKey, "Global header key should match")

			// Test context with new key
			ctx := context.Background()
			newCtx := NewReferenceIdContext(ctx, "test-value")

			// Verify the value can be retrieved with new key
			val, ok := ClaimsReferenceId(newCtx)
			assert.Equal(t, "test-value", val, "Should retrieve value with new key")
			assert.True(t, ok, "Should retrieve value with new key")
		})
	}
}

func TestClaimsReferenceId(t *testing.T) {
	// Store original values to restore after tests
	originalRefIdContext := RefIdContext
	originalRefHeaderKey := RefHeaderKey
	defer func() {
		RefIdContext = originalRefIdContext
		RefHeaderKey = originalRefHeaderKey
	}()

	tests := []struct {
		name      string
		refID     string
		wantID    string
		setInCtx  bool
		nilCtx    bool
		customKey string
	}{
		{
			name:     "existing_ref_id",
			refID:    "test-ref-id",
			wantID:   "test-ref-id",
			setInCtx: true,
		},
		{
			name:     "missing_ref_id",
			refID:    "",
			wantID:   "",
			setInCtx: false,
		},
		{
			name:   "nil_context",
			refID:  "test-ref-id",
			wantID: "",
			nilCtx: true,
		},
		{
			name:      "custom_key_with_value",
			refID:     "custom-ref-id",
			wantID:    "custom-ref-id",
			setInCtx:  true,
			customKey: "X-Custom-Ref",
		},
		{
			name:     "empty_string_ref_id",
			refID:    "",
			wantID:   "",
			setInCtx: true,
		},
		{
			name:     "special_characters_in_ref_id",
			refID:    "ref@id#123$%^",
			wantID:   "ref@id#123$%^",
			setInCtx: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset to default before each test
			RefIdContext = NewGContextType[string]("X-Correlation-Id")
			RefHeaderKey = RefIdContext.Key()

			// Set custom key if specified
			if tt.customKey != "" {
				NewRefContext(tt.customKey)
			}

			var ctx context.Context
			if !tt.nilCtx {
				ctx = context.Background()
				if tt.setInCtx {
					ctx = NewReferenceIdContext(ctx, tt.refID)
				}
			}

			got, ok := ClaimsReferenceId(ctx)
			if tt.nilCtx || (!tt.setInCtx && tt.wantID == "") {
				assert.False(t, ok, "ClaimsReferenceId() should not find value")
				assert.Equal(t, "", got, "ClaimsReferenceId() should return empty string")
			} else {
				assert.True(t, ok, "ClaimsReferenceId() should find value")
				assert.Equal(t, tt.wantID, got, "ClaimsReferenceId() value should match")
			}
		})
	}
}

func TestNewReferenceIdContext(t *testing.T) {
	// Store original values to restore after tests
	originalRefIdContext := RefIdContext
	originalRefHeaderKey := RefHeaderKey
	defer func() {
		RefIdContext = originalRefIdContext
		RefHeaderKey = originalRefHeaderKey
	}()

	tests := []struct {
		name      string
		refID     string
		nilCtx    bool
		customKey string
	}{
		{
			name:   "set_ref_id_with_context",
			refID:  "test-ref-id",
			nilCtx: false,
		},
		{
			name:   "set_ref_id_with_nil_context",
			refID:  "test-ref-id-nil-ctx",
			nilCtx: true,
		},
		{
			name:      "set_ref_id_with_custom_key",
			refID:     "custom-key-ref-id",
			nilCtx:    false,
			customKey: "X-Custom-Ref",
		},
		{
			name:   "set_empty_ref_id",
			refID:  "",
			nilCtx: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset to default before each test
			RefIdContext = NewGContextType[string]("X-Correlation-Id")
			RefHeaderKey = RefIdContext.Key()

			// Set custom key if specified
			if tt.customKey != "" {
				NewRefContext(tt.customKey)
			}

			var ctx context.Context
			if !tt.nilCtx {
				ctx = context.Background()
			}

			newCtx := NewReferenceIdContext(ctx, tt.refID)
			assert.NotNil(t, newCtx, "NewReferenceIdContext() should not return nil")

			got, ok := ClaimsReferenceId(newCtx)
			assert.True(t, ok, "ClaimsReferenceId() should find value")
			assert.Equal(t, tt.refID, got, "ClaimsReferenceId() value should match")
		})
	}
}

func TestRefContextConcurrency(t *testing.T) {
	// Store original values to restore after tests
	originalRefIdContext := RefIdContext
	originalRefHeaderKey := RefHeaderKey
	defer func() {
		RefIdContext = originalRefIdContext
		RefHeaderKey = originalRefHeaderKey
	}()

	// Reset to default before test
	RefIdContext = NewGContextType[string]("X-Correlation-Id")
	RefHeaderKey = RefIdContext.Key()

	const numGoroutines = 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Create contexts with different reference IDs concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			// Create a context with a unique reference ID
			refID := "ref-id-" + string(rune('A'+id))
			ctx := context.Background()
			ctx = NewReferenceIdContext(ctx, refID)

			// Verify the reference ID can be retrieved
			got, ok := ClaimsReferenceId(ctx)
			assert.True(t, ok, "ClaimsReferenceId() should find value")
			assert.Equal(t, refID, got, "ClaimsReferenceId() value should match")
		}(i)
	}

	wg.Wait()
}
