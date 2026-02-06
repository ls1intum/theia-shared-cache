package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

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
