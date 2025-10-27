package handlers

import (
	"database/sql"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AuthHandler handles authentication-related endpoints.
type AuthHandler struct {
	db           *sql.DB
	generateJWT  func(userID string, ttl time.Duration) (string, error)
	getAccessTTL func() time.Duration
	getRefreshTTL func() time.Duration
	parseJWT     func(token string) (*Claims, error)
}

// NewAuthHandler creates a new AuthHandler with default dependencies.
func NewAuthHandler(db *sql.DB) *AuthHandler {
	return &AuthHandler{
		db:            db,
		generateJWT:   generateJWT,
		getAccessTTL:  getAccessTTL,
		getRefreshTTL: getRefreshTTL,
		parseJWT:      parseJWT,
	}
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

// RefreshRequest represents the expected JSON body for /token/refresh
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
