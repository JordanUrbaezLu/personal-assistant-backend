package chat

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"personal-assistant-backend/internal/models"
)

// ListMessages godoc
// @Summary Get all messages in a chat
// @Description Returns the full conversation history for a given chat ID.
// @Tags Chats
// @Security BearerAuth
// @Produce  json
// @Param chat_id path string true "Chat ID"
// @Success 200 {array} models.Message
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Chat not found"
// @Failure 500 {object} map[string]string "Database error"
// @Router /chats/{chat_id}/messages [get]
func (h *ChatHandler) ListMessages(c *gin.Context) {
	userID := c.GetString("userID")
	chatID := c.Param("chat_id")

	// Verify chat ownership
	var exists bool
	err := h.DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM chats WHERE id = $1 AND user_id = $2
		)
	`, chatID, userID).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "chat not found"})
		return
	}

	// Fetch messages
	rows, err := h.DB.Query(`
		SELECT id, chat_id, role, content, created_at
		FROM messages
		WHERE chat_id = $1
		ORDER BY created_at ASC
	`, chatID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		if err := rows.Scan(&msg.ID, &msg.ChatID, &msg.Role, &msg.Content, &msg.CreatedAt); err == nil {
			messages = append(messages, msg)
		}
	}

	c.JSON(http.StatusOK, messages)
}
