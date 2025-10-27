package chat

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	openai "github.com/sashabaranov/go-openai"
	"personal-assistant-backend/internal/models"
)

// ✅ Mockable OpenAI client factory (allows test injection)
var openAIClientFactory = func(apiKey string) openAIClient {
	return openai.NewClient(apiKey)
}

// ✅ Interface to mock CreateChatCompletion
type openAIClient interface {
	CreateChatCompletion(ctx context.Context, req openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error)
}

// SendMessage godoc
// @Summary Send a message in a chat and get AI response
// @Description Sends a message to a chat. The last 20 messages are sent as context to the AI model.
// @Tags Chats
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param chat_id path string true "Chat ID"
// @Param payload body models.SendMessageReq true "User message content"
// @Success 200 {object} models.MessageResponse
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

	// Append user message
	history = append(history, openai.ChatCompletionMessage{
		Role:    "user",
		Content: req.Content,
	})

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "missing OPENAI_API_KEY in env"})
		return
	}

	client := openAIClientFactory(apiKey)
	ctx := context.Background()

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    "gpt-4o-mini",
		Messages: history,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "model error", "details": err.Error()})
		return
	}

	assistantMessage := resp.Choices[0].Message.Content

	// Save user message
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

	// Save assistant message
	var assistantMsg models.Message
	err = h.DB.QueryRow(`
		INSERT INTO messages (chat_id, role, content, created_at)
		VALUES ($1, 'assistant', $2, $3)
		RETURNING id, created_at
	`, chatID, assistantMessage, time.Now()).
		Scan(&assistantMsg.ID, &assistantMsg.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save assistant message"})
		return
	}
	assistantMsg.ChatID = chatID
	assistantMsg.Role = "assistant"
	assistantMsg.Content = assistantMessage

	c.JSON(http.StatusOK, models.MessageResponse{
		UserMessage:      userMsg,
		AssistantMessage: assistantMsg,
	})
}
