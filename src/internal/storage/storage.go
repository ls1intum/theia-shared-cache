package storage

import (
	"context"
	"errors"
	"io"
)

var (
	ErrNotFound = errors.New("cache entry not found")
)

// Storage defines the interface for cache storage backends.
// This abstraction allows for different implementations (MinIO, S3, filesystem, etc.)
type Storage interface {
	// Get retrieves a cache entry by key.
	// Returns the content reader, content size, and any error.
	// Returns ErrNotFound if the entry does not exist.
	Get(ctx context.Context, key string) (io.ReadCloser, int64, error)

	// Put stores a cache entry.
	// The size parameter is the content length for the upload.
	Put(ctx context.Context, key string, reader io.Reader, size int64) error

	// Exists checks if a cache entry exists.
	Exists(ctx context.Context, key string) (bool, error)

	// Delete removes a cache entry.
	// Returns nil if the entry does not exist.
	Delete(ctx context.Context, key string) error

	// Ping checks if the storage backend is reachable.
	Ping(ctx context.Context) error
}

// NamespacedStorage extends Storage with namespace support.
// This will be used for per-exercise cache isolation in the future.
type NamespacedStorage interface {
	Storage
	// WithNamespace returns a new Storage instance scoped to the given namespace.
	WithNamespace(namespace string) Storage
}
