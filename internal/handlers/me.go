package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Me godoc
// @Summary Get current user info
// @Description Returns the authenticated user's ID extracted from the JWT access token.
// @Tags Auth
// @Security BearerAuth
// @Produce  json
// @Success 200 {object} map[string]interface{} "User ID returned successfully"
// @Failure 401 {object} map[string]string "Unauthorized or missing user ID"
// @Router /me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id not found in context"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
	})
}
