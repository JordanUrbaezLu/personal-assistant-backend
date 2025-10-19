package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)
// Refresh godoc
// @Summary Refresh access token
// @Description Takes a valid refresh token and issues a new access token. The refresh token remains the same.
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param payload body RefreshRequest true "Refresh token payload"
// @Success 200 {object} map[string]string "New access token issued"
// @Failure 400 {object} map[string]string "Invalid payload"
// @Failure 401 {object} map[string]string "Invalid or expired refresh token"
// @Failure 500 {object} map[string]string "Failed to generate new access token"
// @Router /token/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var body RefreshRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	claims, err := parseJWT(body.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	accessToken, err := generateJWT(claims.UserID, getAccessTTL())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create access token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": body.RefreshToken,
	})
}


