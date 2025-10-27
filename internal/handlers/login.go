package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"personal-assistant-backend/internal/models"
)

// Login godoc
// @Summary Login a user
// @Description Authenticates a user with email and password, returning account info with JWT access and refresh tokens.
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param payload body loginReq true "User login credentials"
// @Success 200 {object} models.AuthWithTokensResponse "Authenticated user with access and refresh tokens"
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

	var user models.User
	var hash string

	// ✅ Fetch user from DB
	err := h.db.QueryRow(`
		SELECT id, first_name, last_name, email, phone_number, password_hash, created_at
		FROM users WHERE email=$1
	`, req.Email).Scan(
		&user.ID, &user.FirstName, &user.LastName,
		&user.Email, &user.PhoneNumber, &hash, &user.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	// ✅ Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// ✅ Generate tokens using injected dependencies
	accessToken, err := h.generateJWT(user.ID, h.getAccessTTL())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create access token"})
		return
	}

	refreshToken, err := h.generateJWT(user.ID, h.getRefreshTTL())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create refresh token"})
		return
	}

	// ✅ Build response
	response := models.AuthWithTokensResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	c.JSON(http.StatusOK, response)
}
