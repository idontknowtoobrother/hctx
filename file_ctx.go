package hctx

import (
	"context"
	"mime/multipart"
)

// * File context key
var fileContextKey = NewGContextType[[]*File]("File-Ctx-Key")

// * File holds file-related information
type File struct {
	Content       []byte
	FileHeader    *multipart.FileHeader
	AllowedTypes  []string
	MaxSize       int64
	ContentType   string
	FileExtension string
}

// * ClaimsFile retrieves file context from gin context
func ClaimsFile(ctx context.Context) (files []*File, found bool) {
	files, found = Claims(ctx, fileContextKey)
	if !found {
		return nil, false
	}
	return files, true
}

// * NewFileContext sets file context to gin context
func NewFileContext(ctx context.Context, files []*File) context.Context {
	if files == nil {
		return ctx
	}
	return NewGContext(ctx, fileContextKey, files)
}

// * IsAllowedFileType checks if the file extension is allowed
func (fc *File) IsAllowedFileType() bool {
	if len(fc.AllowedTypes) == 0 {
		return true
	}
	for _, allowedType := range fc.AllowedTypes {
		if allowedType == fc.FileExtension {
			return true
		}
	}
	return false
}

// * ValidateFiles checks if all files in the array are valid
func ValidateFiles(files []*File) bool {
	for _, file := range files {
		if !file.IsAllowedFileType() {
			return false
		}
	}
	return true
}
