package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgconn"
	"golang.org/x/crypto/bcrypt"
	"personal-assistant-backend/internal/models"
)

// Signup godoc
// @Summary Register a new user
// @Description Creates a new user account in PostgreSQL and returns account info with JWT access + refresh tokens.
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param payload body signupReq true "User signup data"
// @Success 201 {object} models.AuthWithTokensResponse "User created successfully with access and refresh tokens"
// @Failure 400 {object} map[string]string "Invalid payload"
// @Failure 409 {object} map[string]string "Email already exists"
// @Failure 500 {object} map[string]string "Database or hashing error"
// @Router /signup [post]
func (h *AuthHandler) Signup(c *gin.Context) {
	var req signupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid payload",
			"details": err.Error(),
		})
		return
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// Insert into DB
	var user models.User
	err = h.db.QueryRow(`
		INSERT INTO users (first_name, last_name, email, password_hash, phone_number, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`, req.FirstName, req.LastName, req.Email, string(hash), req.PhoneNumber, time.Now()).
		Scan(&user.ID, &user.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "db error",
			"details": err.Error(),
		})
		return
	}

	// Generate tokens
	accessToken, err := generateJWT(user.ID, getAccessTTL())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create access token"})
		return
	}

	refreshToken, err := generateJWT(user.ID, getRefreshTTL())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create refresh token"})
		return
	}

	// Populate user struct
	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.Email = req.Email
	user.PhoneNumber = req.PhoneNumber

	// Build and return typed response
	response := models.AuthWithTokensResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	c.JSON(http.StatusCreated, response)
}
