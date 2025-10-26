package chat

import "database/sql"

type ChatHandler struct {
	DB *sql.DB
}

func NewChatHandler(db *sql.DB) *ChatHandler {
	return &ChatHandler{DB: db}
}
