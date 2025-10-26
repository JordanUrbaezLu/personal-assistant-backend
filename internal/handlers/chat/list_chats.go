package chat

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"personal-assistant-backend/internal/models"
)

// ListChats godoc
// @Summary List all chats for current user
// @Description Returns all chat sessions created by the logged-in user.
// @Tags Chats
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.ChatListResponse
// @Failure 500 {object} map[string]string "Database error"
// @Router /chats [get]
func (h *ChatHandler) ListChats(c *gin.Context) {
	userID := c.GetString("userID")

	rows, err := h.DB.Query(`
		SELECT id, title, created_at
		FROM chats
		WHERE user_id = $1
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}
	defer rows.Close()

	var chats []models.Chat
	for rows.Next() {
		var chat models.Chat
		if err := rows.Scan(&chat.ID, &chat.Title, &chat.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "scan error"})
			return
		}
		chats = append(chats, chat)
	}

	c.JSON(http.StatusOK, models.ChatListResponse{Chats: chats})
}
