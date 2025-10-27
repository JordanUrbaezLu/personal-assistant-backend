package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// setupGreetRouter creates a Gin router with the GreetHandler mounted.
func setupGreetRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/greet", GreetHandler)
	return r
}

func TestGreetHandler_Success(t *testing.T) {
	router := setupGreetRouter()

	req, _ := http.NewRequest("GET", "/greet?first=Jordan&last=Urbaez", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "Hello, Jordan Urbaez", resp["message"])
}

func TestGreetHandler_MissingFirst(t *testing.T) {
	router := setupGreetRouter()

	req, _ := http.NewRequest("GET", "/greet?last=Urbaez", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Missing first or last name")
}

func TestGreetHandler_MissingLast(t *testing.T) {
	router := setupGreetRouter()

	req, _ := http.NewRequest("GET", "/greet?first=Jordan", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Missing first or last name")
}

func TestGreetHandler_NoParams(t *testing.T) {
	router := setupGreetRouter()

	req, _ := http.NewRequest("GET", "/greet", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Missing first or last name")
}
