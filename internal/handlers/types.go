package handlers

import (
	"database/sql"

	"github.com/golang-jwt/jwt/v5"
)

// AuthHandler holds DB connection
type AuthHandler struct {
	db *sql.DB
}

func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{db: db}
}

// Signup request payload
type signupReq struct {
	FirstName   string `json:"first_name" binding:"required"`
	LastName    string `json:"last_name" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8,max=128"`
	PhoneNumber string `json:"phone_number,omitempty"`
}

// Login request payload
type loginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// JWT claims
type Claims struct {
	UserID string `json:"sub"`
	jwt.RegisteredClaims
}
