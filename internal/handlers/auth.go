package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	db *sql.DB
}

func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

// Matches the new DB schema
type signupReq struct {
	FirstName   string `json:"first_name" binding:"required"`
	LastName    string `json:"last_name" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8,max=128"`
	PhoneNumber string `json:"phone_number,omitempty"`
}

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

	// Insert user
	var userID string
	err = h.db.QueryRow(`
		INSERT INTO users (first_name, last_name, email, password_hash, phone_number)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, req.FirstName, req.LastName, req.Email, string(hash), req.PhoneNumber).Scan(&userID)

	if err != nil {
		// Handle duplicate email
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

	// Respond with user info
	c.JSON(http.StatusCreated, gin.H{
		"user_id":      userID,
		"first_name":   req.FirstName,
		"last_name":    req.LastName,
		"email":        req.Email,
		"phone_number": req.PhoneNumber,
	})
}
