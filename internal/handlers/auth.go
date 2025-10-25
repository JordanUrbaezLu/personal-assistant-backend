package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"personal-assistant-backend/internal/models"
)

// AuthCheck godoc
// @Summary Check user session (Auth validation)
// @Description Validates the user's access token and returns their account information if still logged in. This endpoint does not issue new tokens.
// @Tags Auth
// @Security BearerAuth
// @Produce  json
// @Success 200 {object} models.AuthCheckResponse "Valid token and user info (no tokens returned)"
// @Failure 401 {object} map[string]string "Invalid or expired token"
// @Failure 500 {object} map[string]string "Database error"
// @Router /auth [get]
func (h *AuthHandler) AuthCheck(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var user models.User
	err := h.db.QueryRow(`
		SELECT id, first_name, last_name, email, phone_number, created_at
		FROM users WHERE id = $1
	`, userID).Scan(
		&user.ID, &user.FirstName, &user.LastName,
		&user.Email, &user.PhoneNumber, &user.CreatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		return
	}

	// Build token-free response
	response := models.AuthCheckResponse{
		User: user,
	}

	c.JSON(http.StatusOK, response)
}
