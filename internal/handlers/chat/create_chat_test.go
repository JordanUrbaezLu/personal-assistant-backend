package chat

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"personal-assistant-backend/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// setupChatRouter sets up Gin + sqlmock for ChatHandler
func setupChatRouter(t *testing.T) (*gin.Engine, sqlmock.Sqlmock) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}

	h := &ChatHandler{DB: db}
	r := gin.Default()

	// Add fake userID in context middleware
	r.Use(func(c *gin.Context) {
		c.Set("userID", "user123")
		c.Next()
	})

	r.POST("/chats", h.CreateChat)
	return r, mock
}

// --- TESTS ---

func TestCreateChat_Success(t *testing.T) {
	router, mock := setupChatRouter(t)

	now := time.Now()
	mock.ExpectQuery(`INSERT INTO chats \(user_id, title\) VALUES \(\$1, \$2\) RETURNING id, title, created_at`).
		WithArgs("user123", "My First Chat").
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "created_at"}).
			AddRow("chat-123", "My First Chat", now))

	body := `{"title":"My First Chat"}`
	req, _ := http.NewRequest("POST", "/chats", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.ChatCreateResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "My First Chat", resp.Chat.Title)
	assert.NotEmpty(t, resp.Chat.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateChat_InvalidJSON_DefaultTitle(t *testing.T) {
	router, mock := setupChatRouter(t)

	now := time.Now()
	mock.ExpectQuery(`INSERT INTO chats \(user_id, title\) VALUES \(\$1, \$2\) RETURNING id, title, created_at`).
		WithArgs("user123", "New Chat").
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "created_at"}).
			AddRow("chat-999", "New Chat", now))

	req, _ := http.NewRequest("POST", "/chats", strings.NewReader(`{invalid}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.ChatCreateResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "New Chat", resp.Chat.Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateChat_EmptyTitle_DefaultTitle(t *testing.T) {
	router, mock := setupChatRouter(t)

	now := time.Now()
	mock.ExpectQuery(`INSERT INTO chats \(user_id, title\) VALUES \(\$1, \$2\) RETURNING id, title, created_at`).
		WithArgs("user123", "New Chat").
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "created_at"}).
			AddRow("chat-111", "New Chat", now))

	body := `{"title":""}`
	req, _ := http.NewRequest("POST", "/chats", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp models.ChatCreateResponse
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "New Chat", resp.Chat.Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateChat_DBError(t *testing.T) {
	router, mock := setupChatRouter(t)

	mock.ExpectQuery(`INSERT INTO chats`).WillReturnError(errors.New("db exploded"))

	body := `{"title":"Boom"}`
	req, _ := http.NewRequest("POST", "/chats", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "db error")
	assert.NoError(t, mock.ExpectationsWereMet())
}
