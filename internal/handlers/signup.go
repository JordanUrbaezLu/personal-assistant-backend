package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgconn"
	"golang.org/x/crypto/bcrypt"
)

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
	var userID string
	err = h.db.QueryRow(`
		INSERT INTO users (first_name, last_name, email, password_hash, phone_number)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, req.FirstName, req.LastName, req.Email, string(hash), req.PhoneNumber).Scan(&userID)

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

	// Response
	c.JSON(http.StatusCreated, gin.H{
		"user_id":      userID,
		"first_name":   req.FirstName,
		"last_name":    req.LastName,
		"email":        req.Email,
		"phone_number": req.PhoneNumber,
	})
}
