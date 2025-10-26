package models

// Message represents an individual chat message (user or assistant).
type Message struct {
	ID        string `json:"id"`
	ChatID    string `json:"chat_id"`
	Role      string `json:"role"`   // "user" or "assistant"
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// Request body when sending a message
type SendMessageReq struct {
	Content string `json:"content" binding:"required"`
}

// Response after sending a message
type MessageResponse struct {
	UserMessage      Message `json:"user_message"`
	AssistantMessage Message `json:"assistant_message"`
}
