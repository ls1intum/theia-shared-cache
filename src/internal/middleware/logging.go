package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// RequestLogger creates a middleware that logs HTTP requests.
func RequestLogger(logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)
		status := c.Writer.Status()
		size := c.Writer.Size()

		// Build log event
		event := logger.Info()
		if status >= 500 {
			event = logger.Error()
		} else if status >= 400 {
			event = logger.Warn()
		}

		event.
			Str("method", method).
			Str("path", path).
			Int("status", status).
			Int("size", size).
			Dur("latency", latency).
			Str("client_ip", c.ClientIP())

		// Add username if authenticated
		if username, exists := c.Get("username"); exists {
			event.Str("user", username.(string))
		}

		// Add cache key if present
		if key := c.Param("key"); key != "" {
			event.Str("cache_key", key)
		}

		// Add error if present
		if len(c.Errors) > 0 {
			event.Str("error", c.Errors.String())
		}

		event.Msg("request")
	}
}
