package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// --- Mock helpers for Refresh tests ---

var mockParseJWT = func(token string) (*Claims, error) {
	if token == "valid-refresh" {
		return &Claims{UserID: "user123"}, nil
	}
	if token == "expired-refresh" {
		return nil, errors.New("token expired")
	}
	return nil, errors.New("invalid token")
}

var mockGenerateJWT = func(userID string, _ time.Duration) (string, error) {
	if userID == "fail-token" {
		return "", errors.New("token error")
	}
	return "mock-token-" + userID, nil
}

var mockGetAccessTTL = func() time.Duration { return 15 * time.Minute }

func setupRefreshRouter(t *testing.T) *gin.Engine {
	gin.SetMode(gin.TestMode)

	h := &AuthHandler{
		parseJWT:     mockParseJWT,
		generateJWT:  mockGenerateJWT,
		getAccessTTL: mockGetAccessTTL,
	}
	r := gin.Default()
	r.POST("/token/refresh", h.Refresh)
	return r
}

// --- Tests ---

func TestRefresh_Success(t *testing.T) {
	router := setupRefreshRouter(t)

	body := `{"refresh_token":"valid-refresh"}`
	req, _ := http.NewRequest("POST", "/token/refresh", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "mock-token-user123", resp["access_token"])
	assert.Equal(t, "valid-refresh", resp["refresh_token"])
}

func TestRefresh_InvalidToken(t *testing.T) {
	router := setupRefreshRouter(t)

	body := `{"refresh_token":"bad-token"}`
	req, _ := http.NewRequest("POST", "/token/refresh", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid refresh token")
}

func TestRefresh_ExpiredToken(t *testing.T) {
	router := setupRefreshRouter(t)

	body := `{"refresh_token":"expired-refresh"}`
	req, _ := http.NewRequest("POST", "/token/refresh", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid refresh token")
}

func TestRefresh_InvalidPayload(t *testing.T) {
	router := setupRefreshRouter(t)

	req, _ := http.NewRequest("POST", "/token/refresh", strings.NewReader(`{invalid-json}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid payload")
}

func TestRefresh_TokenGenError(t *testing.T) {
	// Override generateJWT to force error
	errorJWT := func(_ string, _ time.Duration) (string, error) {
		return "", errors.New("token creation failed")
	}

	h := &AuthHandler{
		parseJWT:     mockParseJWT,
		generateJWT:  errorJWT,
		getAccessTTL: mockGetAccessTTL,
	}
	r := gin.Default()
	r.POST("/token/refresh", h.Refresh)

	body := `{"refresh_token":"valid-refresh"}`
	req, _ := http.NewRequest("POST", "/token/refresh", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to create access token")
}
