package hctx

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestClaimsFile(t *testing.T) {
	tests := []struct {
		name      string
		files     []*File
		wantFiles bool
		wantNil   bool
		nilCtx    bool
	}{
		{
			name: "valid files",
			files: []*File{
				{
					Content:       []byte("test content"),
					FileExtension: "txt",
					ContentType:   "text/plain",
				},
			},
			wantFiles: true,
			wantNil:   false,
		},
		{
			name:      "nil files",
			files:     nil,
			wantFiles: false,
			wantNil:   true,
		},
		{
			name:      "nil context",
			nilCtx:    true,
			wantFiles: false,
			wantNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ctx context.Context
			if !tt.nilCtx {
				ctx = context.Background()
				if tt.files != nil {
					ctx = NewFileContext(ctx, tt.files)
				}
			}

			got, ok := ClaimsFile(ctx)
			assert.Equal(t, tt.wantFiles, ok, "ClaimsFile() ok value")
			if tt.wantNil {
				assert.Nil(t, got, "ClaimsFile() should return nil")
			} else {
				assert.NotNil(t, got, "ClaimsFile() should not return nil")
				assert.Equal(t, len(tt.files), len(got), "ClaimsFile() returned files length")
			}
		})
	}
}

func TestNewFileContext(t *testing.T) {
	tests := []struct {
		name      string
		files     []*File
		wantFiles bool
		wantLen   int
	}{
		{
			name: "set single file",
			files: []*File{
				{
					Content:       []byte("test content"),
					FileExtension: "txt",
					ContentType:   "text/plain",
				},
			},
			wantFiles: true,
			wantLen:   1,
		},
		{
			name: "set multiple files",
			files: []*File{
				{
					Content:       []byte("test1"),
					FileExtension: "txt",
					ContentType:   "text/plain",
				},
				{
					Content:       []byte("test2"),
					FileExtension: "pdf",
					ContentType:   "application/pdf",
				},
			},
			wantFiles: true,
			wantLen:   2,
		},
		{
			name:      "set nil files",
			files:     nil,
			wantFiles: false,
			wantLen:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			newCtx := NewFileContext(ctx, tt.files)

			got, ok := ClaimsFile(newCtx)
			assert.Equal(t, tt.wantFiles, ok, "NewFileContext() files presence")
			if tt.wantFiles {
				assert.NotNil(t, got, "NewFileContext() should return files")
				assert.Equal(t, tt.wantLen, len(got), "NewFileContext() files length")
			} else {
				assert.Nil(t, got, "NewFileContext() should return nil for nil files")
			}
		})
	}
}

func TestFile_IsAllowedFileType(t *testing.T) {
	tests := []struct {
		name        string
		file        *File
		wantAllowed bool
	}{
		{
			name: "allowed type",
			file: &File{
				FileExtension: "txt",
				AllowedTypes:  []string{"txt", "pdf"},
			},
			wantAllowed: true,
		},
		{
			name: "not allowed type",
			file: &File{
				FileExtension: "exe",
				AllowedTypes:  []string{"txt", "pdf"},
			},
			wantAllowed: false,
		},
		{
			name: "empty allowed types",
			file: &File{
				FileExtension: "any",
				AllowedTypes:  []string{},
			},
			wantAllowed: true,
		},
		{
			name: "nil allowed types",
			file: &File{
				FileExtension: "any",
				AllowedTypes:  nil,
			},
			wantAllowed: true,
		},
		{
			name: "case sensitive match",
			file: &File{
				FileExtension: "TXT",
				AllowedTypes:  []string{"txt", "pdf"},
			},
			wantAllowed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.file.IsAllowedFileType()
			assert.Equal(t, tt.wantAllowed, got, "IsAllowedFileType() result")
		})
	}
}

func TestValidateFiles(t *testing.T) {
	tests := []struct {
		name      string
		files     []*File
		wantValid bool
	}{
		{
			name: "all files valid",
			files: []*File{
				{
					FileExtension: "txt",
					AllowedTypes:  []string{"txt", "pdf"},
				},
				{
					FileExtension: "pdf",
					AllowedTypes:  []string{"txt", "pdf"},
				},
			},
			wantValid: true,
		},
		{
			name: "one file invalid",
			files: []*File{
				{
					FileExtension: "txt",
					AllowedTypes:  []string{"txt", "pdf"},
				},
				{
					FileExtension: "exe",
					AllowedTypes:  []string{"txt", "pdf"},
				},
			},
			wantValid: false,
		},
		{
			name: "empty allowed types",
			files: []*File{
				{
					FileExtension: "anything",
					AllowedTypes:  []string{},
				},
			},
			wantValid: true,
		},
		{
			name:      "nil files",
			files:     nil,
			wantValid: true,
		},
		{
			name:      "empty files array",
			files:     []*File{},
			wantValid: true,
		},
		{
			name: "mixed validation rules",
			files: []*File{
				{
					FileExtension: "txt",
					AllowedTypes:  []string{"txt", "pdf"},
				},
				{
					FileExtension: "anything",
					AllowedTypes:  []string{},
				},
			},
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateFiles(tt.files)
			assert.Equal(t, tt.wantValid, got, "ValidateFiles() result")
		})
	}
}

func TestNewContextKey(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()

	tests := []struct {
		name    string
		keyName string
		wantKey HContext[[]*File]
	}{
		{
			name:    "create key",
			keyName: "test-key",
			wantKey: NewGContextType[[]*File]("test-key"),
		},
		{
			name:    "empty key",
			keyName: "",
			wantKey: NewGContextType[[]*File](""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use unique path for each test case
			path := fmt.Sprintf("/test-key-%s", tt.name)
			router.GET(path, func(c *gin.Context) {
				got := NewGContextType[[]*File](tt.keyName)
				assert.Equal(t, tt.wantKey, got)
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
