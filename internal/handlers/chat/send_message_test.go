package chat

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)

// --- mock AI client implementing openAIClient interface ---
type mockAIClient struct {
	resp openai.ChatCompletionResponse
	err  error
}

func (m *mockAIClient) CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	if m.err != nil {
		return openai.ChatCompletionResponse{}, m.err
	}
	return m.resp, nil
}

// --- router setup ---
func setupSendMessageRouter(t *testing.T, aiClient openAIClient) (*gin.Engine, sqlmock.Sqlmock) {
	gin.SetMode(gin.TestMode)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}

	h := &ChatHandler{DB: db}
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Set("userID", "user123")
		c.Next()
	})

	// ✅ Override factory for test
	openAIClientFactory = func(apiKey string) openAIClient {
		return aiClient
	}

	r.POST("/chats/:chat_id/messages", h.SendMessage)
	return r, mock
}

// --- TESTS ---

func TestSendMessage_Success(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "mock-key")

	aiClient := &mockAIClient{
		resp: openai.ChatCompletionResponse{
			Choices: []openai.ChatCompletionChoice{
				{Message: openai.ChatCompletionMessage{Role: "assistant", Content: "Hi there!"}},
			},
		},
	}

	router, mock := setupSendMessageRouter(t, aiClient)
	now := time.Now()

	mock.ExpectQuery(`SELECT EXISTS`).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery(`SELECT role, content`).WillReturnRows(sqlmock.NewRows([]string{"role", "content"}))
	mock.ExpectQuery(`INSERT INTO messages`).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow("1", now))
	mock.ExpectQuery(`INSERT INTO messages`).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).AddRow("2", now))

	req, _ := http.NewRequest("POST", "/chats/abc/messages", strings.NewReader(`{"content":"Hello"}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Hi there!")
}

func TestSendMessage_InvalidPayload(t *testing.T) {
	router, _ := setupSendMessageRouter(t, &mockAIClient{})
	req, _ := http.NewRequest("POST", "/chats/abc/messages", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSendMessage_ChatNotFound(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "mock-key")

	router, mock := setupSendMessageRouter(t, &mockAIClient{})
	mock.ExpectQuery(`SELECT EXISTS`).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	req, _ := http.NewRequest("POST", "/chats/abc/messages", strings.NewReader(`{"content":"yo"}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSendMessage_DBError(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "mock-key")

	router, mock := setupSendMessageRouter(t, &mockAIClient{})
	mock.ExpectQuery(`SELECT EXISTS`).WillReturnError(sql.ErrConnDone)

	req, _ := http.NewRequest("POST", "/chats/abc/messages", strings.NewReader(`{"content":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSendMessage_MissingAPIKey(t *testing.T) {
	os.Unsetenv("OPENAI_API_KEY")

	router, mock := setupSendMessageRouter(t, &mockAIClient{})
	mock.ExpectQuery(`SELECT EXISTS`).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery(`SELECT role, content`).WillReturnRows(sqlmock.NewRows([]string{"role", "content"}))

	req, _ := http.NewRequest("POST", "/chats/abc/messages", strings.NewReader(`{"content":"Hello"}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "missing OPENAI_API_KEY")
}

func TestSendMessage_ModelError(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "mock-key")

	aiClient := &mockAIClient{err: errors.New("model failed")}
	router, mock := setupSendMessageRouter(t, aiClient)

	mock.ExpectQuery(`SELECT EXISTS`).WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery(`SELECT role, content`).WillReturnRows(sqlmock.NewRows([]string{"role", "content"}))

	req, _ := http.NewRequest("POST", "/chats/abc/messages", strings.NewReader(`{"content":"Hello"}`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "model error")
}
