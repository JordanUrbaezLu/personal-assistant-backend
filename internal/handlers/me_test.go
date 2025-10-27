package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestMe_Success ensures Me returns user_id when present in context
func TestMe_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &AuthHandler{}
	router := gin.Default()
	router.GET("/me", h.Me)

	req, _ := http.NewRequest("GET", "/me", nil)
	w := httptest.NewRecorder()

	// Simulate middleware setting "userID"
	c, _ := gin.CreateTestContext(w)
	c.Request = req
	c.Set("userID", "user-123")

	h.Me(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "user-123", resp["user_id"])
}

// TestMe_Unauthorized ensures Me returns 401 if userID missing
func TestMe_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &AuthHandler{}
	router := gin.Default()
	router.GET("/me", h.Me)

	req, _ := http.NewRequest("GET", "/me", nil)
	w := httptest.NewRecorder()

	// No userID in context
	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h.Me(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "userID not found")
}
