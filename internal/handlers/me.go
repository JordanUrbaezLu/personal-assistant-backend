package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

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
