package middleware

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"personal-assistant-backend/internal/handlers"
)

// JWTAuthMiddleware ensures requests have a valid JWT access token
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Println("❌ Missing Authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}

		// Expect "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			log.Printf("❌ Invalid Authorization header format: %q\n", authHeader)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			c.Abort()
			return
		}

		tokenStr := parts[1]
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			log.Println("❌ JWT_SECRET not set in environment")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "server misconfigured"})
			c.Abort()
			return
		}

		// Parse and validate
		token, err := jwt.ParseWithClaims(tokenStr, &handlers.Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err != nil {
			log.Printf("❌ Token parse error: %v\n", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		if !token.Valid {
			log.Println("❌ Token parsed but marked as invalid")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*handlers.Claims)
		if !ok {
			log.Println("❌ Failed to cast token claims to expected type")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			c.Abort()
			return
		}

		log.Printf("✅ Valid token for user_id=%s, expires=%v\n", claims.UserID, claims.ExpiresAt)

		// Store user ID in context
		c.Set("user_id", claims.UserID)

		c.Next()
	}
}
