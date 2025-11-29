package message

import (
	"database/sql"
)

// MessageRepository handles message-related database operations.
// It encapsulates the logic for interacting with the 'messages' and 'chat_images' tables.
type MessageRepository struct {
	// DB is the underlying database connection pool used for executing SQL queries.
	DB *sql.DB
}

// NewMessageRepository creates and returns a new MessageRepository instance.
//
// It takes an active database connection pool (*sql.DB) and initializes the repository,
// allowing access to all message-specific CRUD and query methods.
func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{DB: db}
}