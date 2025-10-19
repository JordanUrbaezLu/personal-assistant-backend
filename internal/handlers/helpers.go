package handlers

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Generate a JWT token for given userID and TTL
func generateJWT(userID string, ttl time.Duration) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", errors.New("JWT_SECRET not set")
	}

	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// Parse and validate a JWT token
func parseJWT(tokenStr string) (*Claims, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, errors.New("JWT_SECRET not set")
	}

	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

// Access token TTL in minutes
func getAccessTTL() time.Duration {
	minStr := os.Getenv("ACCESS_TTL_MINUTES")
	if minStr == "" {
		return 15 * time.Minute
	}
	min, err := strconv.Atoi(minStr)
	if err != nil || min <= 0 {
		return 15 * time.Minute
	}
	return time.Duration(min) * time.Minute
}

// Refresh token TTL in days
func getRefreshTTL() time.Duration {
	daysStr := os.Getenv("REFRESH_TTL_DAYS")
	if daysStr == "" {
		return 30 * 24 * time.Hour
	}
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		return 30 * 24 * time.Hour
	}
	return time.Duration(days) * 24 * time.Hour
}
