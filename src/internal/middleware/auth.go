package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kevingruber/gradle-cache/internal/config"
)

// CacheAuth creates a middleware that validates HTTP Basic Authentication
func CacheAuth(auth config.AuthConfig, requireWriter bool) gin.HandlerFunc {

	return func(c *gin.Context) {
		username, password, ok := c.Request.BasicAuth()
		if !ok {
			c.Header("WWW-Authenticate", `Basic realm="Gradle Build Cache"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Check credentials
		isReader := username == auth.Reader.Username &&
			subtle.ConstantTimeCompare([]byte(password), []byte(auth.Reader.Password)) == 1
		isWriter := username == auth.Writer.Username &&
			subtle.ConstantTimeCompare([]byte(password), []byte(auth.Writer.Password)) == 1

		if !isReader && !isWriter {
			c.Header("WWW-Authenticate", `Basic realm="Gradle Build Cache"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if requireWriter && !isWriter {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Next()
	}
}
