package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/kevingruber/gradle-cache/internal/storage"
	"net/http"
)

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
			h.metrics.CacheMisses.Add(c.Request.Context(), 1)
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
	h.metrics.CacheHits.Add(c.Request.Context(), 1)
	c.DataFromReader(http.StatusOK, size, "application/octet-stream", reader, nil)
}
