package handler

import (
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

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
