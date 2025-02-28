# GCTX - Go Context Utilities

[![Go Version](https://img.shields.io/badge/Go-1.24.0-blue.svg)](https://golang.org/doc/go1.24)
[![Coverage](https://img.shields.io/badge/Coverage-100.0%25-brightgreen.svg)](coverage.out)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

A lightweight, type-safe context management library for Go applications, providing utilities for handling context values with generic type support.

## Features

- **Type-safe context management** using Go generics
- **File context handling** for HTTP file uploads with validation
- **Reference ID tracking** for request correlation
- **100% test coverage** ensuring reliability

## Installation

```bash
go get github.com/idontknowtoobrother/hctx
```

## Usage

### Generic Context

```go
// Create a typed context key
userCtx := hctx.NewGContextType[User]("user-context")

// Store a value in context
ctx = hctx.NewGContext(ctx, userCtx, user)

// Retrieve the value with type safety
user, found := hctx.Claims(ctx, userCtx)
if found {
    // Use user...
}

// Use with HTTP requests
req = hctx.ToHttpRequest(req, userCtx, user)
```

### File Context

```go
// Define file constraints
file := &hctx.File{
    Content:       fileBytes,
    FileHeader:    fileHeader,
    AllowedTypes:  []string{"jpg", "png", "pdf"},
    MaxSize:       5 * 1024 * 1024, // 5MB
    ContentType:   "image/jpeg",
    FileExtension: "jpg",
}

// Store files in context
ctx = hctx.NewFileContext(ctx, []*hctx.File{file})

// Retrieve files from context
files, found := hctx.ClaimsFile(ctx)
if found {
    // Validate files
    if hctx.ValidateFiles(files) {
        // Process files...
    }
}
```

### Reference ID Context

```go
// Configure a custom reference ID key
refCtx := hctx.NewRefContext("X-Request-ID")

// Store a reference ID in context
ctx = hctx.NewReferenceIdContext(ctx, "req-123456")

// Retrieve the reference ID
refID, found := hctx.ClaimsReferenceId(ctx)
if found {
    // Use reference ID for logging, tracing, etc.
}
```

## Thread Safety

The reference ID context utilities are thread-safe, using a mutex to protect global variables.

## License

MIT License
