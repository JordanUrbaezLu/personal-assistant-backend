package handlers

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

// --- generateJWT tests ---

func TestGenerateJWT_Success(t *testing.T) {
	os.Setenv("JWT_SECRET", "testsecret")
	defer os.Unsetenv("JWT_SECRET")

	token, err := generateJWT("user123", time.Minute*10)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate token structure
	parsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte("testsecret"), nil
	})
	assert.NoError(t, err)
	assert.True(t, parsed.Valid)
}

func TestGenerateJWT_MissingSecret(t *testing.T) {
	os.Unsetenv("JWT_SECRET")

	token, err := generateJWT("user123", time.Minute*10)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET not set")
	assert.Empty(t, token)
}

// --- parseJWT tests ---

func TestParseJWT_Success(t *testing.T) {
	os.Setenv("JWT_SECRET", "parseSecret")
	defer os.Unsetenv("JWT_SECRET")

	// create valid token
	token, _ := generateJWT("validUser", time.Minute*5)

	claims, err := parseJWT(token)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "validUser", claims.UserID)
}

func TestParseJWT_InvalidToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "parseSecret")
	defer os.Unsetenv("JWT_SECRET")

	claims, err := parseJWT("this-is-not-a-valid-token")
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestParseJWT_MissingSecret(t *testing.T) {
	os.Unsetenv("JWT_SECRET")

	claims, err := parseJWT("whatever")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET not set")
	assert.Nil(t, claims)
}

// --- getAccessTTL tests ---

func TestGetAccessTTL_Default(t *testing.T) {
	os.Unsetenv("ACCESS_TTL_MINUTES")
	ttl := getAccessTTL()
	assert.Equal(t, 15*time.Minute, ttl)
}

func TestGetAccessTTL_CustomValid(t *testing.T) {
	os.Setenv("ACCESS_TTL_MINUTES", "45")
	defer os.Unsetenv("ACCESS_TTL_MINUTES")
	ttl := getAccessTTL()
	assert.Equal(t, 45*time.Minute, ttl)
}

func TestGetAccessTTL_InvalidValue(t *testing.T) {
	os.Setenv("ACCESS_TTL_MINUTES", "abc")
	defer os.Unsetenv("ACCESS_TTL_MINUTES")
	ttl := getAccessTTL()
	assert.Equal(t, 15*time.Minute, ttl)
}

// --- getRefreshTTL tests ---

func TestGetRefreshTTL_Default(t *testing.T) {
	os.Unsetenv("REFRESH_TTL_DAYS")
	ttl := getRefreshTTL()
	assert.Equal(t, 30*24*time.Hour, ttl)
}

func TestGetRefreshTTL_CustomValid(t *testing.T) {
	os.Setenv("REFRESH_TTL_DAYS", "10")
	defer os.Unsetenv("REFRESH_TTL_DAYS")
	ttl := getRefreshTTL()
	assert.Equal(t, 10*24*time.Hour, ttl)
}

func TestGetRefreshTTL_InvalidValue(t *testing.T) {
	os.Setenv("REFRESH_TTL_DAYS", "-3")
	defer os.Unsetenv("REFRESH_TTL_DAYS")
	ttl := getRefreshTTL()
	assert.Equal(t, 30*24*time.Hour, ttl)
}
