package user_repository

import (
	"database/sql"
	"testing"
	"time"

	"social-network/models"
	dbmigrate "social-network/pkg/db/sqlite"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB creates a temporary test database
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Run migrations
	if err := dbmigrate.Migrate(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return db
}

// createTestUser creates a user for testing
func createTestUser(t *testing.T, db *sql.DB, nickname string, isPrivate bool) *models.User {
	userID := "test-user-" + nickname
	email := nickname + "@test.com"

	_, err := db.Exec(`
		INSERT INTO user (user_id, nickname, email, first_name, last_name, date_of_birth, 
		                  gender, is_private, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, userID, nickname, email, "Test", "User", time.Now().AddDate(-25, 0, 0), "male", isPrivate, time.Now())

	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create password hash
	_, err = db.Exec(`
		INSERT INTO user_auth (user_id, password_hash)
		VALUES (?, ?)
	`, userID, "$2a$10$testHash")

	if err != nil {
		t.Fatalf("Failed to create user auth: %v", err)
	}

	return &models.User{
		ID:        userID,
		Nickname:  nickname,
		Email:     email,
		FirstName: "Test",
		LastName:  "User",
		IsPrivate: isPrivate,
	}
}

// createFollowRequest creates a follow request between two users
func createFollowRequest(t *testing.T, db *sql.DB, followerID, followeeID, status string) {
	_, err := db.Exec(`
		INSERT INTO follow_relationships (follower_id, followee_id, status, created_at)
		VALUES (?, ?, ?, ?)
	`, followerID, followeeID, status, time.Now())

	if err != nil {
		t.Fatalf("Failed to create follow request: %v", err)
	}
}

// TestUpdatePrivacy_PrivateToPublic tests changing from private to public
func TestUpdatePrivacy_PrivateToPublic(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)

	// Create a private user
	user := createTestUser(t, db, "privateuser", true)

	// Update to public
	err := repo.UpdatePrivacy(user.ID, false)
	if err != nil {
		t.Fatalf("Failed to update privacy: %v", err)
	}

	// Verify the change
	var isPrivate bool
	err = db.QueryRow(`SELECT is_private FROM user WHERE user_id = ?`, user.ID).Scan(&isPrivate)
	if err != nil {
		t.Fatalf("Failed to query user: %v", err)
	}

	if isPrivate {
		t.Errorf("Expected is_private to be false, got true")
	}
}

// TestUpdatePrivacy_PublicToPrivate tests changing from public to private
func TestUpdatePrivacy_PublicToPrivate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)

	// Create a public user
	user := createTestUser(t, db, "publicuser", false)

	// Update to private
	err := repo.UpdatePrivacy(user.ID, true)
	if err != nil {
		t.Fatalf("Failed to update privacy: %v", err)
	}

	// Verify the change
	var isPrivate bool
	err = db.QueryRow(`SELECT is_private FROM user WHERE user_id = ?`, user.ID).Scan(&isPrivate)
	if err != nil {
		t.Fatalf("Failed to query user: %v", err)
	}

	if !isPrivate {
		t.Errorf("Expected is_private to be true, got false")
	}
}

// TestUpdatePrivacy_CleanupPendingRequests tests the critical security feature
func TestUpdatePrivacy_CleanupPendingRequests(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)

	// Create users
	userA := createTestUser(t, db, "usera", true)  // private profile
	userB := createTestUser(t, db, "userb", false) // public profile
	userC := createTestUser(t, db, "userc", false) // public profile

	// Create pending follow requests to user A
	createFollowRequest(t, db, userB.ID, userA.ID, "pending")
	createFollowRequest(t, db, userC.ID, userA.ID, "pending")

	// Verify pending requests exist
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM follow_relationships 
		WHERE followee_id = ? AND status = 'pending'
	`, userA.ID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query follow requests: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 pending requests, got %d", count)
	}

	// CRITICAL TEST: Change user A from private to public
	err = repo.UpdatePrivacy(userA.ID, false)
	if err != nil {
		t.Fatalf("Failed to update privacy: %v", err)
	}

	// SECURITY CHECK: Verify pending requests were deleted
	err = db.QueryRow(`
		SELECT COUNT(*) FROM follow_relationships 
		WHERE followee_id = ? AND status = 'pending'
	`, userA.ID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query follow requests: %v", err)
	}

	if count != 0 {
		t.Errorf("SECURITY VIOLATION: Expected 0 pending requests after going public, got %d", count)
	}
}

// TestUpdatePrivacy_PreserveAcceptedFollows tests that accepted follows are not deleted
func TestUpdatePrivacy_PreserveAcceptedFollows(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)

	// Create users
	userA := createTestUser(t, db, "usera", true)  // private profile
	userB := createTestUser(t, db, "userb", false)
	userC := createTestUser(t, db, "userc", false)

	// Create one accepted and one pending follow request
	createFollowRequest(t, db, userB.ID, userA.ID, "accepted")
	createFollowRequest(t, db, userC.ID, userA.ID, "pending")

	// Change to public
	err := repo.UpdatePrivacy(userA.ID, false)
	if err != nil {
		t.Fatalf("Failed to update privacy: %v", err)
	}

	// Verify accepted request still exists
	var acceptedCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM follow_relationships 
		WHERE followee_id = ? AND status = 'accepted'
	`, userA.ID).Scan(&acceptedCount)
	if err != nil {
		t.Fatalf("Failed to query accepted follows: %v", err)
	}

	if acceptedCount != 1 {
		t.Errorf("Expected 1 accepted follow, got %d", acceptedCount)
	}

	// Verify pending request was deleted
	var pendingCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM follow_relationships 
		WHERE followee_id = ? AND status = 'pending'
	`, userA.ID).Scan(&pendingCount)
	if err != nil {
		t.Fatalf("Failed to query pending follows: %v", err)
	}

	if pendingCount != 0 {
		t.Errorf("Expected 0 pending follows, got %d", pendingCount)
	}
}

// TestUpdatePrivacy_UserNotFound tests error handling
func TestUpdatePrivacy_UserNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)

	// Try to update non-existent user
	err := repo.UpdatePrivacy("non-existent-user", true)
	if err == nil {
		t.Error("Expected error for non-existent user, got nil")
	}
}

// TestUpdateProfile tests profile update functionality
func TestUpdateProfile(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewUserRepository(db)

	// Create a user
	user := createTestUser(t, db, "testuser", false)

	// Update profile
	err := repo.UpdateProfile(user.ID, "Updated", "Name", "New bio", "female")
	if err != nil {
		t.Fatalf("Failed to update profile: %v", err)
	}

	// Verify updates
	var firstName, lastName, aboutMe, gender string
	err = db.QueryRow(`
		SELECT first_name, last_name, about_me, gender 
		FROM user 
		WHERE user_id = ?
	`, user.ID).Scan(&firstName, &lastName, &aboutMe, &gender)

	if err != nil {
		t.Fatalf("Failed to query user: %v", err)
	}

	if firstName != "Updated" {
		t.Errorf("Expected first_name 'Updated', got '%s'", firstName)
	}
	if lastName != "Name" {
		t.Errorf("Expected last_name 'Name', got '%s'", lastName)
	}
	if aboutMe != "New bio" {
		t.Errorf("Expected about_me 'New bio', got '%s'", aboutMe)
	}
	if gender != "female" {
		t.Errorf("Expected gender 'female', got '%s'", gender)
	}
}