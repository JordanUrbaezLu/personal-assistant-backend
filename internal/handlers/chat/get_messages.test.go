package chat

import (
	"database/sql"
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

func setupListMessagesRouter(t *testing.T) (*gin.Engine, sqlmock.Sqlmock) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	h := &ChatHandler{DB: db}
	r := gin.Default()

	// Middleware adds fake userID
	r.Use(func(c *gin.Context) {
		c.Set("userID", "user123")
		c.Next()
	})

	r.GET("/chats/:chat_id/messages", h.ListMessages)
	return r, mock
}

func TestListMessages_Success(t *testing.T) {
	router, mock := setupListMessagesRouter(t)

	now := time.Now()

	// Ownership check returns true
	mock.ExpectQuery(`SELECT EXISTS \(.*FROM chats WHERE id = \$1 AND user_id = \$2\)`).
		WithArgs("chat123", "user123").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// Return two messages
	mock.ExpectQuery(`SELECT id, chat_id, role, content, created_at FROM messages WHERE chat_id = \$1 ORDER BY created_at ASC`).
		WithArgs("chat123").
		WillReturnRows(sqlmock.NewRows([]string{"id", "chat_id", "role", "content", "created_at"}).
			AddRow("msg1", "chat123", "user", "Hello", now).
			AddRow("msg2", "chat123", "assistant", "Hi there!", now))

	req, _ := http.NewRequest("GET", "/chats/chat123/messages", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var msgs []models.Message
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &msgs))
	assert.Len(t, msgs, 2)
	assert.Equal(t, "Hello", msgs[0].Content)
	assert.Equal(t, "assistant", msgs[1].Role)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListMessages_ChatNotFound(t *testing.T) {
	router, mock := setupListMessagesRouter(t)

	mock.ExpectQuery(`SELECT EXISTS \(.*FROM chats WHERE id = \$1 AND user_id = \$2\)`).
		WithArgs("chat404", "user123").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	req, _ := http.NewRequest("GET", "/chats/chat404/messages", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "chat not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListMessages_DBErrorOnOwnershipCheck(t *testing.T) {
	router, mock := setupListMessagesRouter(t)

	mock.ExpectQuery(`SELECT EXISTS \(.*FROM chats WHERE id = \$1 AND user_id = \$2\)`).
		WithArgs("chat500", "user123").
		WillReturnError(errors.New("db exploded"))

	req, _ := http.NewRequest("GET", "/chats/chat500/messages", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "db error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListMessages_DBErrorOnMessagesQuery(t *testing.T) {
	router, mock := setupListMessagesRouter(t)

	mock.ExpectQuery(`SELECT EXISTS \(.*FROM chats WHERE id = \$1 AND user_id = \$2\)`).
		WithArgs("chat999", "user123").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	mock.ExpectQuery(`SELECT id, chat_id, role, content, created_at FROM messages WHERE chat_id = \$1 ORDER BY created_at ASC`).
		WithArgs("chat999").
		WillReturnError(sql.ErrConnDone)

	req, _ := http.NewRequest("GET", "/chats/chat999/messages", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "db error")
	assert.NoError(t, mock.ExpectationsWereMet())
}
