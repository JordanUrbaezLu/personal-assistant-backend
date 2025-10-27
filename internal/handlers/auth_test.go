package handlers

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// setupAuthCheckRouter initializes the test environment for AuthCheck
func setupAuthCheckRouter(t *testing.T) (*gin.Engine, sqlmock.Sqlmock) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}

	authHandler := NewAuthHandler(db)
	r := gin.Default()

	// Simulate JWT middleware setting userID in context
	r.GET("/auth", func(c *gin.Context) {
		c.Set("userID", "user-123")
		authHandler.AuthCheck(c)
	})

	return r, mock
}

func TestAuthCheck_Success(t *testing.T) {
	router, mock := setupAuthCheckRouter(t)

	// Mock DB response
	now := time.Now()
	mock.ExpectQuery(`SELECT id, first_name, last_name, email, phone_number, created_at FROM users WHERE id = \$1`).
		WithArgs("user-123").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "first_name", "last_name", "email", "phone_number", "created_at",
		}).AddRow("user-123", "John", "Doe", "john@example.com", "1234567890", now))

	req, _ := http.NewRequest("GET", "/auth", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "john@example.com")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAuthCheck_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, _, _ := sqlmock.New()
	authHandler := NewAuthHandler(db)
	r := gin.Default()

	// Missing userID in context
	r.GET("/auth", authHandler.AuthCheck)

	req, _ := http.NewRequest("GET", "/auth", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")
}

func TestAuthCheck_UserNotFound(t *testing.T) {
	router, mock := setupAuthCheckRouter(t)

	mock.ExpectQuery(`SELECT id, first_name, last_name, email, phone_number, created_at FROM users WHERE id = \$1`).
		WithArgs("user-123").
		WillReturnError(sql.ErrNoRows)

	req, _ := http.NewRequest("GET", "/auth", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "user not found")
}
