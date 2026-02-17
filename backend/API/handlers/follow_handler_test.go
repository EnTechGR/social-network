package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"social-network/handlers"
	"social-network/models"
	"social-network/repository"
	"social-network/repository/user_repository"
	"social-network/websocket"

	"github.com/google/uuid"
)

func createFollowTestUser(t *testing.T, testDB *TestDB, nickname string, isPrivate bool) string {
	t.Helper()

	userID := uuid.NewString()
	email := nickname + "@example.com"

	_, err := testDB.DB.Exec(`
		INSERT INTO user (user_id, email, nickname, first_name, last_name, date_of_birth, about_me, gender, is_private, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, userID, email, nickname, "Test", "User", "1990-01-01", "", "other", isPrivate, time.Now())
	if err != nil {
		t.Fatalf("failed to create user %s: %v", nickname, err)
	}

	return userID
}

func withAuthUser(req *http.Request, userID, nickname string) *http.Request {
	user := &models.User{
		ID:       userID,
		Nickname: nickname,
	}
	ctx := context.WithValue(req.Context(), "user", user)
	return req.WithContext(ctx)
}

func TestFollowPublicUser_PrivateTarget_CreatesAndPushesFollowRequestNotification(t *testing.T) {
	testDB := SetupTestDB(t)
	defer testDB.TeardownTestDB()

	followerID := createFollowTestUser(t, testDB, "follower_private_case", false)
	followeeID := createFollowTestUser(t, testDB, "private_target_case", true)

	followRepo := repository.NewFollowRepository(testDB.DB)
	userRepo := user_repository.NewUserRepository(testDB.DB)
	notificationRepo := repository.NewNotificationRepository(testDB.DB)
	hub := websocket.NewHub()

	targetClient := &websocket.Client{
		UserID:   followeeID,
		Nickname: "private_target_case",
		Send:     make(chan []byte, 1),
		Hub:      hub,
	}
	hub.Clients[followeeID] = targetClient

	handler := handlers.NewFollowHandler(followRepo, userRepo, notificationRepo, hub)

	body := `{"followee_id":"` + followeeID + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/follow", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = withAuthUser(req, followerID, "follower_private_case")

	w := httptest.NewRecorder()
	handler.FollowPublicUser(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d (body: %s)", http.StatusCreated, w.Code, w.Body.String())
	}

	var notifType, notifUserID, notifFromUserID string
	err := testDB.DB.QueryRow(`
		SELECT type, user_id, from_user_id
		FROM notifications
		LIMIT 1
	`).Scan(&notifType, &notifUserID, &notifFromUserID)
	if err != nil {
		t.Fatalf("failed to query notification row: %v", err)
	}

	if notifType != "follow_request" {
		t.Fatalf("expected notification type follow_request, got %s", notifType)
	}
	if notifUserID != followeeID {
		t.Fatalf("expected notification user_id %s, got %s", followeeID, notifUserID)
	}
	if notifFromUserID != followerID {
		t.Fatalf("expected notification from_user_id %s, got %s", followerID, notifFromUserID)
	}

	select {
	case raw := <-targetClient.Send:
		var msg struct {
			Type string                  `json:"type"`
			Data models.NotificationView `json:"data"`
		}
		if err := json.Unmarshal(raw, &msg); err != nil {
			t.Fatalf("failed to decode websocket payload: %v", err)
		}
		if msg.Type != "notification" {
			t.Fatalf("expected websocket type notification, got %s", msg.Type)
		}
		if msg.Data.Type != "follow_request" {
			t.Fatalf("expected websocket notification type follow_request, got %s", msg.Data.Type)
		}
		if msg.Data.Nickname != "follower_private_case" {
			t.Fatalf("expected websocket notification nickname follower_private_case, got %s", msg.Data.Nickname)
		}
	default:
		t.Fatal("expected websocket notification to be sent to followee")
	}

	notifs, err := notificationRepo.GetByUser(followeeID)
	if err != nil {
		t.Fatalf("GetByUser failed: %v", err)
	}
	if len(notifs) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(notifs))
	}
	if notifs[0].Type != "follow_request" {
		t.Fatalf("expected persisted notification type follow_request, got %s", notifs[0].Type)
	}
	if notifs[0].PostID != "" {
		t.Fatalf("expected empty post_id for follow notification, got %s", notifs[0].PostID)
	}
}

func TestAcceptFollowRequest_SendsFollowAcceptNotificationToFollower(t *testing.T) {
	testDB := SetupTestDB(t)
	defer testDB.TeardownTestDB()

	followerID := createFollowTestUser(t, testDB, "pending_follower_case", false)
	followeeID := createFollowTestUser(t, testDB, "private_owner_case", true)

	followRepo := repository.NewFollowRepository(testDB.DB)
	userRepo := user_repository.NewUserRepository(testDB.DB)
	notificationRepo := repository.NewNotificationRepository(testDB.DB)
	hub := websocket.NewHub()

	if _, err := followRepo.CreateFollowRequest(followerID, followeeID); err != nil {
		t.Fatalf("failed to seed pending follow request: %v", err)
	}

	followerClient := &websocket.Client{
		UserID:   followerID,
		Nickname: "pending_follower_case",
		Send:     make(chan []byte, 1),
		Hub:      hub,
	}
	hub.Clients[followerID] = followerClient

	handler := handlers.NewFollowHandler(followRepo, userRepo, notificationRepo, hub)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/follow/accept/"+followerID, nil)
	req = withAuthUser(req, followeeID, "private_owner_case")

	w := httptest.NewRecorder()
	handler.AcceptFollowRequest(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d (body: %s)", http.StatusOK, w.Code, w.Body.String())
	}

	var notifType, notifUserID, notifFromUserID string
	err := testDB.DB.QueryRow(`
		SELECT type, user_id, from_user_id
		FROM notifications
		ORDER BY created_at DESC
		LIMIT 1
	`).Scan(&notifType, &notifUserID, &notifFromUserID)
	if err != nil {
		t.Fatalf("failed to query notification row: %v", err)
	}

	if notifType != "follow_accept" {
		t.Fatalf("expected notification type follow_accept, got %s", notifType)
	}
	if notifUserID != followerID {
		t.Fatalf("expected notification user_id %s, got %s", followerID, notifUserID)
	}
	if notifFromUserID != followeeID {
		t.Fatalf("expected notification from_user_id %s, got %s", followeeID, notifFromUserID)
	}

	select {
	case raw := <-followerClient.Send:
		var msg struct {
			Type string                  `json:"type"`
			Data models.NotificationView `json:"data"`
		}
		if err := json.Unmarshal(raw, &msg); err != nil {
			t.Fatalf("failed to decode websocket payload: %v", err)
		}
		if msg.Type != "notification" {
			t.Fatalf("expected websocket type notification, got %s", msg.Type)
		}
		if msg.Data.Type != "follow_accept" {
			t.Fatalf("expected websocket notification type follow_accept, got %s", msg.Data.Type)
		}
		if msg.Data.Nickname != "private_owner_case" {
			t.Fatalf("expected websocket notification nickname private_owner_case, got %s", msg.Data.Nickname)
		}
	default:
		t.Fatal("expected websocket follow_accept notification to be sent to follower")
	}
}
