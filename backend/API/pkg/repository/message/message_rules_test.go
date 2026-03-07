package message

import (
	"database/sql"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupRulesDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}

	if _, err := db.Exec(`
		CREATE TABLE user (
			user_id TEXT PRIMARY KEY,
			is_private BOOLEAN NOT NULL DEFAULT 0,
			nickname TEXT NOT NULL
		);

		CREATE TABLE follow_relationships (
			follower_id TEXT NOT NULL,
			followee_id TEXT NOT NULL,
			status TEXT NOT NULL
		);
	`); err != nil {
		db.Close()
		t.Fatalf("create schema: %v", err)
	}

	return db
}

func TestCanDeliverMessageRealtime(t *testing.T) {
	const senderID = "sender"
	const receiverID = "receiver"

	t.Run("receiver_public", func(t *testing.T) {
		db := setupRulesDB(t)
		defer db.Close()

		if _, err := db.Exec(`INSERT INTO user (user_id, is_private, nickname) VALUES (?, ?, ?), (?, ?, ?)`,
			senderID, false, "sender",
			receiverID, false, "receiver",
		); err != nil {
			t.Fatalf("seed users: %v", err)
		}

		repo := NewMessageRepository(db)
		ok, err := repo.CanDeliverMessageRealtime(senderID, receiverID)
		if err != nil {
			t.Fatalf("CanDeliverMessageRealtime returned error: %v", err)
		}
		if !ok {
			t.Fatalf("expected realtime delivery for public receiver")
		}
	})

	t.Run("receiver_private_and_follows_sender", func(t *testing.T) {
		db := setupRulesDB(t)
		defer db.Close()

		if _, err := db.Exec(`INSERT INTO user (user_id, is_private, nickname) VALUES (?, ?, ?), (?, ?, ?)`,
			senderID, false, "sender",
			receiverID, true, "receiver",
		); err != nil {
			t.Fatalf("seed users: %v", err)
		}
		if _, err := db.Exec(`INSERT INTO follow_relationships (follower_id, followee_id, status) VALUES (?, ?, 'accepted')`,
			receiverID, senderID,
		); err != nil {
			t.Fatalf("seed follow relationship: %v", err)
		}

		repo := NewMessageRepository(db)
		ok, err := repo.CanDeliverMessageRealtime(senderID, receiverID)
		if err != nil {
			t.Fatalf("CanDeliverMessageRealtime returned error: %v", err)
		}
		if !ok {
			t.Fatalf("expected realtime delivery when receiver follows sender")
		}
	})

	t.Run("receiver_private_and_only_sender_follows", func(t *testing.T) {
		db := setupRulesDB(t)
		defer db.Close()

		if _, err := db.Exec(`INSERT INTO user (user_id, is_private, nickname) VALUES (?, ?, ?), (?, ?, ?)`,
			senderID, false, "sender",
			receiverID, true, "receiver",
		); err != nil {
			t.Fatalf("seed users: %v", err)
		}
		if _, err := db.Exec(`INSERT INTO follow_relationships (follower_id, followee_id, status) VALUES (?, ?, 'accepted')`,
			senderID, receiverID,
		); err != nil {
			t.Fatalf("seed follow relationship: %v", err)
		}

		repo := NewMessageRepository(db)
		ok, err := repo.CanDeliverMessageRealtime(senderID, receiverID)
		if err != nil {
			t.Fatalf("CanDeliverMessageRealtime returned error: %v", err)
		}
		if ok {
			t.Fatalf("expected realtime delivery to be disabled when private receiver does not follow sender")
		}
	})
}

func TestCreateEmojiLengthUsesCharacters(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`
		CREATE TABLE messages (
			message_id TEXT PRIMARY KEY,
			sender_id TEXT NOT NULL,
			receiver_id TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			is_read BOOLEAN NOT NULL DEFAULT 0
		);
	`); err != nil {
		t.Fatalf("create messages table: %v", err)
	}

	repo := NewMessageRepository(db)

	okContent := strings.Repeat("😀", 1000)
	if _, err := repo.Create("u1", "u2", okContent); err != nil {
		t.Fatalf("expected 1000 emoji characters to be accepted, got: %v", err)
	}

	tooLong := strings.Repeat("😀", 1001)
	if _, err := repo.Create("u1", "u2", tooLong); err == nil {
		t.Fatalf("expected content above 1000 characters to fail")
	}
}

func TestCreateGroupMessageEmojiLengthUsesCharacters(t *testing.T) {
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

		CREATE TABLE group_messages (
			message_id TEXT PRIMARY KEY,
			group_id TEXT NOT NULL,
			sender_id TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL
		);
	`); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	if _, err := db.Exec(`INSERT INTO user (user_id, nickname) VALUES (?, ?)`, "u1", "alice"); err != nil {
		t.Fatalf("seed sender: %v", err)
	}

	repo := NewMessageRepository(db)

	okContent := strings.Repeat("😀", 1000)
	if _, err := repo.CreateGroupMessage("g1", "u1", okContent); err != nil {
		t.Fatalf("expected 1000 emoji characters to be accepted, got: %v", err)
	}

	tooLong := strings.Repeat("😀", 1001)
	if _, err := repo.CreateGroupMessage("g1", "u1", tooLong); err == nil {
		t.Fatalf("expected group content above 1000 characters to fail")
	}
}
