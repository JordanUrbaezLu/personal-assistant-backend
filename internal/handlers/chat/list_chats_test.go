package chat

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"personal-assistant-backend/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// setupListChatsRouter sets up Gin + sqlmock for ChatHandler
func setupListChatsRouter(t *testing.T) (*gin.Engine, sqlmock.Sqlmock) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}

	h := &ChatHandler{DB: db}
	r := gin.Default()

	// Middleware injects userID
	r.Use(func(c *gin.Context) {
		c.Set("userID", "user123")
		c.Next()
	})

	r.GET("/chats", h.ListChats)
	return r, mock
}

// --- TESTS ---

func TestListChats_Success(t *testing.T) {
	router, mock := setupListChatsRouter(t)

	now := time.Now()

	mock.ExpectQuery(`SELECT id, title, created_at FROM chats WHERE user_id = \$1 ORDER BY created_at DESC`).
		WithArgs("user123").
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "created_at"}).
			AddRow("chat1", "First Chat", now).
			AddRow("chat2", "Second Chat", now.Add(-time.Hour)))

	req, _ := http.NewRequest("GET", "/chats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp models.ChatListResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp.Chats, 2)
	assert.Equal(t, "First Chat", resp.Chats[0].Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListChats_DBError(t *testing.T) {
	router, mock := setupListChatsRouter(t)

	mock.ExpectQuery(`SELECT id, title, created_at FROM chats WHERE user_id = \$1 ORDER BY created_at DESC`).
		WithArgs("user123").
		WillReturnError(errors.New("db exploded"))

	req, _ := http.NewRequest("GET", "/chats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "db error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListChats_ScanError(t *testing.T) {
	router, mock := setupListChatsRouter(t)

	// Simulate a broken row (extra column value will trigger Scan error)
	mockRows := sqlmock.NewRows([]string{"id", "title"}).AddRow("chat1", "Broken Chat")
	mock.ExpectQuery(`SELECT id, title, created_at FROM chats`).
		WithArgs("user123").
		WillReturnRows(mockRows)

	req, _ := http.NewRequest("GET", "/chats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "scan error")
}
