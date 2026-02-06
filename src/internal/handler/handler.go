package handler

import (
	"github.com/kevingruber/gradle-cache/internal/storage"
	"github.com/rs/zerolog"
	"io"
)

// CacheHandler handles Gradle build cache HTTP requests.
type CacheHandler struct {
	storage      storage.Storage
	maxEntrySize int64
	logger       zerolog.Logger
	metrics      *Metrics
}

// NewCacheHandler creates a new cache handler.
func NewCacheHandler(store storage.Storage, maxEntrySize int64, logger zerolog.Logger) (*CacheHandler, error) {
	metrics, err := NewMetrics()
	if err != nil {
		return nil, err
	}

	return &CacheHandler{
		storage:      store,
		maxEntrySize: maxEntrySize,
		logger:       logger,
		metrics:      metrics,
	}, nil
}

// bytesReaderAt implements io.ReaderAt for a byte slice.
type bytesReaderAt struct {
	data []byte
}

func (b *bytesReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	if off >= int64(len(b.data)) {
		return 0, io.EOF
	}
	n = copy(p, b.data[off:])
	if n < len(p) {
		err = io.EOF
	}
	return
}
