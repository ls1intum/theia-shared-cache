package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kevingruber/gradle-cache/internal/storage"
	"github.com/rs/zerolog"
)

// CacheHandler handles Gradle build cache HTTP requests.
type CacheHandler struct {
	storage      storage.Storage
	maxEntrySize int64
	logger       zerolog.Logger
}

// NewCacheHandler creates a new cache handler.
func NewCacheHandler(store storage.Storage, maxEntrySize int64, logger zerolog.Logger) *CacheHandler {
	return &CacheHandler{
		storage:      store,
		maxEntrySize: maxEntrySize,
		logger:       logger,
	}
}

// Get handles GET requests to retrieve cache entries.
// Gradle expects: 200 with body on hit, 404 on miss.
func (h *CacheHandler) Get(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	reader, size, err := h.storage.Get(c.Request.Context(), key)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			c.Status(http.StatusNotFound)
			return
		}
		h.logger.Error().Err(err).Str("key", key).Msg("failed to get cache entry")
		c.Status(http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", string(rune(size)))

	c.DataFromReader(http.StatusOK, size, "application/octet-stream", reader, nil)
}

// Put handles PUT requests to store cache entries.
// Gradle expects: 2xx on success, 413 if too large.
func (h *CacheHandler) Put(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	// Check Content-Length header for size validation
	contentLength := c.Request.ContentLength
	if contentLength > h.maxEntrySize {
		h.logger.Warn().
			Str("key", key).
			Int64("size", contentLength).
			Int64("max_size", h.maxEntrySize).
			Msg("cache entry too large")
		c.Status(http.StatusRequestEntityTooLarge)
		return
	}

	// Handle Expect: 100-continue
	// Gin/Go handles this automatically, but we validate size first

	// For chunked transfers or unknown size, we need to handle differently
	if contentLength < 0 {
		// Read with size limit
		limitedReader := io.LimitReader(c.Request.Body, h.maxEntrySize+1)
		data, err := io.ReadAll(limitedReader)
		if err != nil {
			h.logger.Error().Err(err).Str("key", key).Msg("failed to read request body")
			c.Status(http.StatusInternalServerError)
			return
		}

		if int64(len(data)) > h.maxEntrySize {
			c.Status(http.StatusRequestEntityTooLarge)
			return
		}

		contentLength = int64(len(data))
		c.Request.Body = io.NopCloser(io.NewSectionReader(
			&bytesReaderAt{data: data}, 0, contentLength,
		))
	}

	err := h.storage.Put(c.Request.Context(), key, c.Request.Body, contentLength)
	if err != nil {
		h.logger.Error().Err(err).Str("key", key).Msg("failed to store cache entry")
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusCreated)
}

// Head handles HEAD requests to check cache entry existence.
func (h *CacheHandler) Head(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.Status(http.StatusBadRequest)
		return
	}

	exists, err := h.storage.Exists(c.Request.Context(), key)
	if err != nil {
		h.logger.Error().Err(err).Str("key", key).Msg("failed to check cache entry existence")
		c.Status(http.StatusInternalServerError)
		return
	}

	if !exists {
		c.Status(http.StatusNotFound)
		return
	}

	c.Status(http.StatusOK)
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
