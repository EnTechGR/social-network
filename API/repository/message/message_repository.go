package message

import (
	"database/sql"
)

// MessageRepository handles message-related database operations
type MessageRepository struct {
	DB *sql.DB
}

// NewMessageRepository creates a new MessageRepository
func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{DB: db}
}