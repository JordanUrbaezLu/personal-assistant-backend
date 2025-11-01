package chat

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	openai "github.com/sashabaranov/go-openai"
	"personal-assistant-backend/internal/models"
)

// ✅ Interface for streaming client
type openAIStreamClient interface {
	CreateChatCompletionStream(ctx context.Context, req openai.ChatCompletionRequest) (*openai.ChatCompletionStream, error)
}

// ✅ Factory (mockable + TLS-safe)
var openAIStreamFactory = func(apiKey string) openAIStreamClient {
	config := openai.DefaultConfig(apiKey)
	config.HTTPClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: false, // ✅ verifies properly if ca-certificates are installed
			},
		},
		Timeout: 0, // allow long-lived connections for streaming
	}
	return openai.NewClientWithConfig(config)
}

// SendMessage godoc
// @Summary Send a message in a chat and stream AI response
// @Description Sends a message to a chat. The last 20 messages are sent as context to the AI model, and tokens stream back in real time.
// @Tags Chats
// @Security BearerAuth
// @Accept json
// @Produce text/event-stream
// @Param chat_id path string true "Chat ID"
// @Param payload body models.SendMessageReq true "User message content"
// @Success 200 {string} string "Streamed AI response"
// @Failure 400 {object} map[string]string "Invalid payload"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Chat not found"
// @Failure 500 {object} map[string]string "Database or model error"
// @Router /chats/{chat_id}/messages [post]
func (h *ChatHandler) SendMessage(c *gin.Context) {
	userID := c.GetString("userID")
	chatID := c.Param("chat_id")

	var req models.SendMessageReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	// Verify chat ownership
	var exists bool
	err := h.DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM chats WHERE id = $1 AND user_id = $2
		)
	`, chatID, userID).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error", "details": err.Error()})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "chat not found"})
		return
	}

	// Fetch last 20 messages
	rows, err := h.DB.Query(`
		SELECT role, content
		FROM messages
		WHERE chat_id = $1
		ORDER BY created_at DESC
		LIMIT 20
	`, chatID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load chat history"})
		return
	}
	defer rows.Close()

	var history []openai.ChatCompletionMessage
	for rows.Next() {
		var role, content string
		if err := rows.Scan(&role, &content); err == nil {
			history = append([]openai.ChatCompletionMessage{{
				Role:    role,
				Content: content,
			}}, history...)
		}
	}

	// Append new user message
	history = append(history, openai.ChatCompletionMessage{
		Role:    "user",
		Content: req.Content,
	})

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "missing OPENAI_API_KEY in env"})
		return
	}

	client := openAIStreamFactory(apiKey)
	ctx := context.Background()

	stream, err := client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:    "gpt-4o-mini",
		Messages: history,
		Stream:   true,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "model error", "details": err.Error()})
		return
	}
	defer stream.Close()

	// Set SSE headers for streaming
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// Save user message before streaming
	var userMsg models.Message
	err = h.DB.QueryRow(`
		INSERT INTO messages (chat_id, role, content, created_at)
		VALUES ($1, 'user', $2, $3)
		RETURNING id, created_at
	`, chatID, req.Content, time.Now()).
		Scan(&userMsg.ID, &userMsg.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save user message"})
		return
	}

	userMsg.ChatID = chatID
	userMsg.Role = "user"
	userMsg.Content = req.Content

	// Stream assistant response tokens to client
	var fullResponse string
	c.Stream(func(w io.Writer) bool {
		for {
			resp, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				c.SSEvent("error", err.Error())
				return false
			}

			if len(resp.Choices) > 0 {
				delta := resp.Choices[0].Delta.Content
				if delta != "" {
					fullResponse += delta
					c.SSEvent("message", delta)
				}
			}
		}

		// Save assistant message once stream finishes
		var assistantMsg models.Message
		err = h.DB.QueryRow(`
			INSERT INTO messages (chat_id, role, content, created_at)
			VALUES ($1, 'assistant', $2, $3)
			RETURNING id, created_at
		`, chatID, fullResponse, time.Now()).
			Scan(&assistantMsg.ID, &assistantMsg.CreatedAt)
		if err == nil {
			assistantMsg.ChatID = chatID
			assistantMsg.Role = "assistant"
			assistantMsg.Content = fullResponse
		}

		c.SSEvent("done", "[DONE]")
		return false
	})
}
