package chat

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"personal-assistant-backend/internal/models"
)

// CreateChat godoc
// @Summary Create a new chat
// @Description Creates a blank chat session for the logged-in user. Optionally accepts a title.
// @Tags Chats
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body models.CreateChatReq false "Optional chat title"
// @Success 201 {object} models.ChatCreateResponse
// @Failure 400 {object} map[string]string "Invalid payload"
// @Failure 500 {object} map[string]string "Database error"
// @Router /chats [post]
func (h *ChatHandler) CreateChat(c *gin.Context) {
	userID := c.GetString("userID")

	var req models.CreateChatReq
	if err := c.ShouldBindJSON(&req); err != nil {
		// If no body or invalid, just use default title
		req.Title = "New Chat"
	}

	title := req.Title
	if title == "" {
		title = "New Chat"
	}

	var chat models.Chat
	err := h.DB.QueryRow(`
		INSERT INTO chats (user_id, title)
		VALUES ($1, $2)
		RETURNING id, title, created_at
	`, userID, title).Scan(&chat.ID, &chat.Title, &chat.CreatedAt)
	if err != nil {
		// Show DB error details (for debugging)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "db error",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.ChatCreateResponse{Chat: chat})
}
