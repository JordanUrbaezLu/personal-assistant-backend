package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Login godoc
// @Summary Login a user
// @Description Authenticates a user with email and password, returning JWT access and refresh tokens.
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param payload body loginReq true "User login credentials"
// @Success 200 {object} map[string]string "Access and refresh tokens"
// @Failure 400 {object} map[string]string "Invalid payload"
// @Failure 401 {object} map[string]string "Invalid credentials"
// @Failure 500 {object} map[string]string "Database or token generation error"
// @Router /login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid payload",
			"details": err.Error(),
		})
		return
	}

	var userID, hash string
	err := h.db.QueryRow(
		`SELECT id, password_hash FROM users WHERE email=$1`,
		req.Email,
	).Scan(&userID, &hash)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Generate tokens
	accessToken, err := generateJWT(userID, getAccessTTL())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create access token"})
		return
	}

	refreshToken, err := generateJWT(userID, getRefreshTTL())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}
