package message

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestGetConversation_WithoutChatImagesTable(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE user (
			user_id TEXT PRIMARY KEY,
			nickname TEXT NOT NULL
		);

		CREATE TABLE messages (
			message_id TEXT PRIMARY KEY,
			sender_id TEXT NOT NULL,
			receiver_id TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			is_read BOOLEAN NOT NULL DEFAULT 0
		);
	`); err != nil {
		t.Fatalf("create test schema: %v", err)
	}

	if _, err := db.Exec(`INSERT INTO user (user_id, nickname) VALUES (?, ?), (?, ?)`, "u1", "alice", "u2", "bob"); err != nil {
		t.Fatalf("seed users: %v", err)
	}

	if _, err := db.Exec(
		`INSERT INTO messages (message_id, sender_id, receiver_id, content, created_at, is_read) VALUES (?, ?, ?, ?, ?, ?)`,
		"m1", "u1", "u2", "hello", time.Now(), false,
	); err != nil {
		t.Fatalf("seed message: %v", err)
	}

	repo := NewMessageRepository(db)

	messages, err := repo.GetConversation("u1", "u2", 10, 0)
	if err != nil {
		t.Fatalf("GetConversation returned error: %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}

	if messages[0].Image != nil {
		t.Fatalf("expected nil image when chat_images table is missing, got %#v", messages[0].Image)
	}
}
