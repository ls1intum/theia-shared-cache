package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kevingruber/gradle-cache/internal/config"
)

type userCredential struct {
	Password string
	Role     string
}

// BasicAuth creates a middleware that validates HTTP Basic Authentication
// and stores the user's role in the gin context.
func BasicAuth(users []config.UserAuth) gin.HandlerFunc {
	// Build a map for O(1) lookup
	credentials := make(map[string]userCredential, len(users))
	for _, user := range users {
		credentials[user.Username] = userCredential{
			Password: user.Password,
			Role:     user.Role,
		}
	}

	return func(c *gin.Context) {
		username, password, ok := c.Request.BasicAuth()
		if !ok {
			c.Header("WWW-Authenticate", `Basic realm="Gradle Build Cache"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		cred, userExists := credentials[username]
		if !userExists {
			c.Header("WWW-Authenticate", `Basic realm="Gradle Build Cache"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Use constant-time comparison to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(password), []byte(cred.Password)) != 1 {
			c.Header("WWW-Authenticate", `Basic realm="Gradle Build Cache"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Store username and role in context for logging/metrics and authorization
		c.Set("username", username)
		c.Set("role", cred.Role)
		c.Next()
	}
}

// RequireRole creates a middleware that checks if the authenticated user
// has the required role. Returns 403 Forbidden if not.
func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role.(string) != requiredRole {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	}
}
