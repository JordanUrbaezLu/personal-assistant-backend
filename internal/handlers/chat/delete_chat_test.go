package chat

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// setupDeleteChatRouter sets up Gin + sqlmock for DeleteChat tests
func setupDeleteChatRouter(t *testing.T) (*gin.Engine, sqlmock.Sqlmock) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}

	h := &ChatHandler{DB: db}
	r := gin.Default()

	// Fake userID in context
	r.Use(func(c *gin.Context) {
		c.Set("userID", "user123")
		c.Next()
	})

	r.DELETE("/chats/:chat_id", h.DeleteChat)
	return r, mock
}

//
// --- TESTS ---
//

// 1️⃣ SUCCESS: chat deleted
func TestDeleteChat_Success(t *testing.T) {
	router, mock := setupDeleteChatRouter(t)

	mock.ExpectExec(`DELETE FROM chats WHERE id = \$1 AND user_id = \$2`).
		WithArgs("chat-123", "user123").
		WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row affected

	req, _ := http.NewRequest("DELETE", "/chats/chat-123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Chat deleted")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 2️⃣ NOT FOUND: zero rows affected
func TestDeleteChat_NotFound(t *testing.T) {
	router, mock := setupDeleteChatRouter(t)

	mock.ExpectExec(`DELETE FROM chats WHERE id = \$1 AND user_id = \$2`).
		WithArgs("missing-id", "user123").
		WillReturnResult(sqlmock.NewResult(0, 0)) // no rows

	req, _ := http.NewRequest("DELETE", "/chats/missing-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Chat not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// 3️⃣ DB ERROR
func TestDeleteChat_DBError(t *testing.T) {
	router, mock := setupDeleteChatRouter(t)

	mock.ExpectExec(`DELETE FROM chats`).
		WillReturnError(assert.AnError)

	req, _ := http.NewRequest("DELETE", "/chats/chat-err", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "db error")
	assert.NoError(t, mock.ExpectationsWereMet())
}
