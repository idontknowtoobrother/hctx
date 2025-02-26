package hctx

import (
	"context"
	"net/http"
	"testing"
)

func TestNewGContextType(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantKey string
	}{
		{
			name:    "create string context",
			key:     "test-key",
			wantKey: "test-key",
		},
		{
			name:    "create int context",
			key:     "int-key",
			wantKey: "int-key",
		},
		{
			name:    "empty key",
			key:     "",
			wantKey: "",
		},
		{
			name:    "special characters in key",
			key:     "test@key#123",
			wantKey: "test@key#123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gc := NewGContextType[string](tt.key)
			if gc.Key() != tt.wantKey {
				t.Errorf("NewGContextType().Key() = %v, want %v", gc.Key(), tt.wantKey)
			}
		})

		t.Run(tt.name+" with int", func(t *testing.T) {
			gc := NewGContextType[int](tt.key)
			if gc.Key() != tt.wantKey {
				t.Errorf("NewGContextType().Key() = %v, want %v", gc.Key(), tt.wantKey)
			}
		})

		// Test with a custom struct type
		t.Run(tt.name+" with struct", func(t *testing.T) {
			type customStruct struct {
				Value string
			}
			gc := NewGContextType[customStruct](tt.key)
			if gc.Key() != tt.wantKey {
				t.Errorf("NewGContextType().Key() = %v, want %v", gc.Key(), tt.wantKey)
			}
		})
	}
}

func TestClaims(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		value     interface{}
		wantValue interface{}
		wantOk    bool
	}{
		{
			name:      "string value",
			key:       "string-key",
			value:     "test-value",
			wantValue: "test-value",
			wantOk:    true,
		},
		{
			name:      "int value",
			key:       "int-key",
			value:     42,
			wantValue: 42,
			wantOk:    true,
		},
		{
			name:      "wrong type",
			key:       "wrong-type",
			value:     123,
			wantValue: "",
			wantOk:    false,
		},
		{
			name:      "nil context",
			key:       "nil-ctx",
			value:     nil,
			wantValue: "",
			wantOk:    false,
		},
		{
			name:      "empty string value",
			key:       "empty-string",
			value:     "",
			wantValue: "",
			wantOk:    true,
		},
		{
			name:      "zero int value",
			key:       "zero-int",
			value:     0,
			wantValue: 0,
			wantOk:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "nil context" {
				got, ok := Claims(nil, NewGContextType[string](tt.key))
				if ok != tt.wantOk || got != tt.wantValue {
					t.Errorf("Claims() got = %v, %v, want %v, %v", got, ok, tt.wantValue, tt.wantOk)
				}
				return
			}

			switch v := tt.value.(type) {
			case string:
				ctx := context.Background()
				gc := NewGContextType[string](tt.key)
				ctx = NewGContext(ctx, gc, v)
				got, ok := Claims(ctx, gc)
				if ok != tt.wantOk || got != tt.wantValue {
					t.Errorf("Claims() got = %v, %v, want %v, %v", got, ok, tt.wantValue, tt.wantOk)
				}
			case int:
				if tt.name == "wrong type" {
					ctx := context.Background()
					gc := NewGContextType[string](tt.key)
					ctx = context.WithValue(ctx, gc, v) // Deliberately store int when expecting string
					got, ok := Claims(ctx, gc)
					if ok != tt.wantOk || got != tt.wantValue {
						t.Errorf("Claims() got = %v, %v, want %v, %v", got, ok, tt.wantValue, tt.wantOk)
					}
				} else {
					ctx := context.Background()
					gc := NewGContextType[int](tt.key)
					ctx = NewGContext(ctx, gc, v)
					got, ok := Claims(ctx, gc)
					if ok != tt.wantOk || got != tt.wantValue {
						t.Errorf("Claims() got = %v, %v, want %v, %v", got, ok, tt.wantValue, tt.wantOk)
					}
				}
			}
		})
	}
}

func TestToHttpRequest(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{
			name:  "add string to request",
			key:   "test-key",
			value: "test-value",
		},
		{
			name:  "empty value",
			key:   "empty-key",
			value: "",
		},
		{
			name:  "special characters in value",
			key:   "special-key",
			value: "value@#$%^&*()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "http://example.com", nil)
			gc := NewGContextType[string](tt.key)

			req = ToHttpRequest(req, gc, tt.value)
			got, ok := Claims(req.Context(), gc)

			if !ok || got != tt.value {
				t.Errorf("ToHttpRequest() got = %v, %v, want %v, true", got, ok, tt.value)
			}
		})
	}
}

// TestNestedContext tests nested context values
func TestNestedContext(t *testing.T) {
	ctx := context.Background()

	// Create first context value
	key1 := NewGContextType[string]("key1")
	ctx = NewGContext(ctx, key1, "value1")

	// Create second context value
	key2 := NewGContextType[int]("key2")
	ctx = NewGContext(ctx, key2, 42)

	// Retrieve and verify both values
	val1, ok1 := Claims(ctx, key1)
	if !ok1 || val1 != "value1" {
		t.Errorf("Claims() for key1 got = %v, %v, want %v, true", val1, ok1, "value1")
	}

	val2, ok2 := Claims(ctx, key2)
	if !ok2 || val2 != 42 {
		t.Errorf("Claims() for key2 got = %v, %v, want %v, true", val2, ok2, 42)
	}
}

// TestStructContext tests using a struct as context value
func TestStructContext(t *testing.T) {
	type User struct {
		ID   int
		Name string
	}

	ctx := context.Background()
	userKey := NewGContextType[User]("user-key")
	user := User{ID: 1, Name: "Test User"}

	ctx = NewGContext(ctx, userKey, user)

	retrievedUser, ok := Claims(ctx, userKey)
	if !ok {
		t.Errorf("Claims() failed to retrieve User struct")
	}

	if retrievedUser.ID != user.ID || retrievedUser.Name != user.Name {
		t.Errorf("Claims() got = %+v, want %+v", retrievedUser, user)
	}
}
