package middleware

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

// APIKeyAuthMiddleware returns a middleware that checks for the given API key
func APIKeyAuthMiddleware(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientKey := c.GetHeader("X-API-Key")
		if clientKey == "" || clientKey != apiKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
			return
		}
		c.Next()
	}
}
