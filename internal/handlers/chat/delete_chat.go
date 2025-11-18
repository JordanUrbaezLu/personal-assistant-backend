package chat

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// DeleteChat godoc
// @Summary Delete a chat
// @Description Deletes a chat if it belongs to the logged-in user.
// @Tags Chats
// @Security BearerAuth
// @Produce json
// @Param chat_id path int true "Chat ID"
// @Success 200 {object} map[string]string "Chat deleted"
// @Failure 404 {object} map[string]string "Chat not found"
// @Failure 500 {object} map[string]string "Database error"
// @Router /chats/{chat_id} [delete]
func (h *ChatHandler) DeleteChat(c *gin.Context) {
	userID := c.GetString("userID")
	chatID := c.Param("chat_id")

	// Only delete if chat belongs to user
	result, err := h.DB.Exec(`
		DELETE FROM chats
		WHERE id = $1 AND user_id = $2
	`, chatID, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "db error",
			"details": err.Error(),
		})
		return
	}

	rows, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "db error",
			"details": err.Error(),
		})
		return
	}

	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chat not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Chat deleted"})
}
