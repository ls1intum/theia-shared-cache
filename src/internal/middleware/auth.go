package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kevingruber/gradle-cache/internal/config"
)

// BasicAuth creates a middleware that validates HTTP Basic Authentication.
func BasicAuth(users []config.UserAuth) gin.HandlerFunc {
	// Build a map for O(1) lookup
	credentials := make(map[string]string, len(users))
	for _, user := range users {
		credentials[user.Username] = user.Password
	}

	return func(c *gin.Context) {
		username, password, ok := c.Request.BasicAuth()
		if !ok {
			c.Header("WWW-Authenticate", `Basic realm="Gradle Build Cache"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		expectedPassword, userExists := credentials[username]
		if !userExists {
			c.Header("WWW-Authenticate", `Basic realm="Gradle Build Cache"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Use constant-time comparison to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(password), []byte(expectedPassword)) != 1 {
			c.Header("WWW-Authenticate", `Basic realm="Gradle Build Cache"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Store username in context for logging/metrics
		c.Set("username", username)
		c.Next()
	}
}
