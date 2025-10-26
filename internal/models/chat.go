package models

// Chat represents a conversation container owned by a user.
type Chat struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
}

// Request body when creating a new chat
type CreateChatReq struct {
	Title string `json:"title" binding:"omitempty,max=120"`
}

// Response for creating a chat
type ChatCreateResponse struct {
	Chat Chat `json:"chat"`
}

// Response for listing chats
type ChatListResponse struct {
	Chats []Chat `json:"chats"`
}
