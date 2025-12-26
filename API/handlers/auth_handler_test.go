package handlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"social-network/handlers"
	"social-network/models"
	dbmigrate "social-network/pkg/db/sqlite"
	"social-network/repository"
	"social-network/repository/session"
	"social-network/repository/user"

	_ "github.com/mattn/go-sqlite3"
)

// TestDB holds the test database connection and manages cleanup
type TestDB struct {
	DB     *sql.DB
	DBPath string
}

// SetupTestDB creates a test database and applies all migrations
// This uses the SAME migration system as production, ensuring schema consistency
func SetupTestDB(t *testing.T) *TestDB {
	// Create temporary directory for test database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_social_network.db")

	// Open database connection with foreign keys enabled
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		t.Fatalf("Failed to ping test database: %v", err)
	}

	// Set connection pool limits
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	// Apply migrations using the SAME system as production
	// This automatically applies all migrations: 000001 through 000005
	if err := dbmigrate.Migrate(db); err != nil {
		db.Close()
		t.Fatalf("Failed to apply migrations: %v", err)
	}

	t.Logf("✅ Test database created with all migrations applied: %s", dbPath)

	return &TestDB{
		DB:     db,
		DBPath: dbPath,
	}
}

// TeardownTestDB closes the database connection and cleans up
func (tdb *TestDB) TeardownTestDB() {
	if tdb.DB != nil {
		tdb.DB.Close()
	}
	// Temp directory is automatically cleaned up by t.TempDir()
}

// createTestAvatar creates a temporary test image file
func createTestAvatar(t *testing.T, filename string, sizeKB int) string {
	// Create temp directory for test files
	tempDir := t.TempDir()
	filepath := filepath.Join(tempDir, filename)

	// Create a simple test image (1x1 PNG)
	// PNG header and minimal valid PNG data
	pngData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4,
		0x89, 0x00, 0x00, 0x00, 0x0A, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE,
		0x42, 0x60, 0x82,
	}

	// If we need a larger file, pad it
	if sizeKB > 0 {
		totalSize := sizeKB * 1024
		if totalSize > len(pngData) {
			padding := make([]byte, totalSize-len(pngData))
			pngData = append(pngData, padding...)
		}
	}

	if err := os.WriteFile(filepath, pngData, 0644); err != nil {
		t.Fatalf("Failed to create test avatar: %v", err)
	}

	return filepath
}

// createMultipartRequest creates a multipart form request for testing
func createMultipartRequest(t *testing.T, fields map[string]string, fileField, filePath string) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add form fields
	for key, val := range fields {
		if err := writer.WriteField(key, val); err != nil {
			return nil, err
		}
	}

	// Add file if provided
	if fileField != "" && filePath != "" {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		// Determine content type based on file extension
		contentType := "image/png"
		ext := filepath.Ext(filePath)
		switch ext {
		case ".jpg", ".jpeg":
			contentType = "image/jpeg"
		case ".gif":
			contentType = "image/gif"
		case ".png":
			contentType = "image/png"
		}

		// Create form file with proper content type
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition",
			fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
				fileField, filepath.Base(filePath)))
		h.Set("Content-Type", contentType)

		part, err := writer.CreatePart(h)
		if err != nil {
			return nil, err
		}

		if _, err := io.Copy(part, file); err != nil {
			return nil, err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	req := httptest.NewRequest("POST", "/api/auth/register", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, nil
}

// TestAuthHandler_Register_Success_NoAvatar tests successful registration without avatar
func TestAuthHandler_Register_Success_NoAvatar(t *testing.T) {
	// Setup: Create test database with real migrations
	testDB := SetupTestDB(t)
	defer testDB.TeardownTestDB()

	userRepo := user.NewUserRepository(testDB.DB)
	sessionRepo := session.NewSessionRepository(testDB.DB)
	imageRepo := repository.NewImageRepository(testDB.DB)
	authHandler := handlers.NewAuthHandler(userRepo, sessionRepo, imageRepo)

	// Create request
	fields := map[string]string{
		"email":         "test@example.com",
		"password":      "SecurePass123",
		"first_name":    "John",
		"last_name":     "Doe",
		"date_of_birth": "1990-01-01",
		"gender":        "male",
		"nickname":      "johndoe",
		"about_me":      "I love testing!",
		"is_private":    "false",
	}

	req, err := createMultipartRequest(t, fields, "", "")
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Execute
	w := httptest.NewRecorder()
	authHandler.Register(w, req)

	// Assert
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		t.Logf("Response body: %s", w.Body.String())
	}

	// Parse response
	var response models.LoginResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify user data
	if response.User.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", response.User.Email)
	}
	if response.User.Nickname != "johndoe" {
		t.Errorf("Expected nickname johndoe, got %s", response.User.Nickname)
	}
	if response.User.FirstName != "John" {
		t.Errorf("Expected first name John, got %s", response.User.FirstName)
	}
	if response.SessionID == "" {
		t.Error("Expected session_id to be set")
	}
	if response.CSRFToken == "" {
		t.Error("Expected csrf_token to be set")
	}

	// Verify user exists in database
	var count int
	err = testDB.DB.QueryRow("SELECT COUNT(*) FROM user WHERE email = ?", "test@example.com").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query database: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 user in database, found %d", count)
	}

	// Verify session exists
	err = testDB.DB.QueryRow("SELECT COUNT(*) FROM sessions WHERE user_id = ?", response.User.ID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query sessions: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 session in database, found %d", count)
	}
}

// TestAuthHandler_Register_Success_WithAvatar tests successful registration with avatar
func TestAuthHandler_Register_Success_WithAvatar(t *testing.T) {
	// Setup
	testDB := SetupTestDB(t)
	defer testDB.TeardownTestDB()

	userRepo := user.NewUserRepository(testDB.DB)
	sessionRepo := session.NewSessionRepository(testDB.DB)
	imageRepo := repository.NewImageRepository(testDB.DB)
	authHandler := handlers.NewAuthHandler(userRepo, sessionRepo, imageRepo)

	// Create test avatar
	avatarPath := createTestAvatar(t, "test-avatar.png", 100) // 100KB

	// Create request with avatar
	fields := map[string]string{
		"email":         "avatar@example.com",
		"password":      "SecurePass123",
		"first_name":    "Jane",
		"last_name":     "Smith",
		"date_of_birth": "1995-05-15",
		"gender":        "female",
	}

	req, err := createMultipartRequest(t, fields, "avatar", avatarPath)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Execute
	w := httptest.NewRecorder()
	authHandler.Register(w, req)

	// Assert
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		t.Logf("Response body: %s", w.Body.String())
	}

	// Parse response
	var response models.LoginResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify avatar was uploaded to images_core (migration 000005)
	var imageCount int
	err = testDB.DB.QueryRow(`
		SELECT COUNT(*) 
		FROM images_core 
		WHERE uploader_user_id = ?
	`, response.User.ID).Scan(&imageCount)
	if err != nil {
		t.Fatalf("Failed to query images: %v", err)
	}
	if imageCount != 1 {
		t.Errorf("Expected 1 image in images_core, found %d", imageCount)
	}

	// Verify user_avatars relationship (migration 000005)
	var avatarCount int
	err = testDB.DB.QueryRow(`
		SELECT COUNT(*) 
		FROM user_avatars 
		WHERE user_id = ?
	`, response.User.ID).Scan(&avatarCount)
	if err != nil {
		t.Fatalf("Failed to query user_avatars: %v", err)
	}
	if avatarCount != 1 {
		t.Errorf("Expected 1 avatar link, found %d", avatarCount)
	}
}

// TestAuthHandler_Register_DefaultNickname tests nickname defaults to email prefix
func TestAuthHandler_Register_DefaultNickname(t *testing.T) {
	testDB := SetupTestDB(t)
	defer testDB.TeardownTestDB()

	userRepo := user.NewUserRepository(testDB.DB)
	sessionRepo := session.NewSessionRepository(testDB.DB)
	imageRepo := repository.NewImageRepository(testDB.DB)
	authHandler := handlers.NewAuthHandler(userRepo, sessionRepo, imageRepo)

	fields := map[string]string{
		"email":         "myemail@example.com",
		"password":      "SecurePass123",
		"first_name":    "Test",
		"last_name":     "User",
		"date_of_birth": "1990-01-01",
		"gender":        "male",
		// No nickname provided
	}

	req, _ := createMultipartRequest(t, fields, "", "")
	w := httptest.NewRecorder()
	authHandler.Register(w, req)

	var response models.LoginResponse
	json.NewDecoder(w.Body).Decode(&response)

	if response.User.Nickname != "myemail" {
		t.Errorf("Expected nickname 'myemail', got '%s'", response.User.Nickname)
	}
}

// TestAuthHandler_Register_MissingRequiredFields tests validation for required fields
func TestAuthHandler_Register_MissingRequiredFields(t *testing.T) {
	testDB := SetupTestDB(t)
	defer testDB.TeardownTestDB()

	userRepo := user.NewUserRepository(testDB.DB)
	sessionRepo := session.NewSessionRepository(testDB.DB)
	imageRepo := repository.NewImageRepository(testDB.DB)
	authHandler := handlers.NewAuthHandler(userRepo, sessionRepo, imageRepo)

	testCases := []struct {
		name          string
		fields        map[string]string
		expectedError string
	}{
		{
			name: "Missing email",
			fields: map[string]string{
				"password":      "SecurePass123",
				"first_name":    "John",
				"last_name":     "Doe",
				"date_of_birth": "1990-01-01",
				"gender":        "male",
			},
			expectedError: "Required fields",
		},
		{
			name: "Missing password",
			fields: map[string]string{
				"email":         "test@example.com",
				"first_name":    "John",
				"last_name":     "Doe",
				"date_of_birth": "1990-01-01",
				"gender":        "male",
			},
			expectedError: "Required fields",
		},
		{
			name: "Missing first_name",
			fields: map[string]string{
				"email":         "test@example.com",
				"password":      "SecurePass123",
				"last_name":     "Doe",
				"date_of_birth": "1990-01-01",
				"gender":        "male",
			},
			expectedError: "Required fields",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := createMultipartRequest(t, tc.fields, "", "")
			w := httptest.NewRecorder()
			authHandler.Register(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
			}

			if !strings.Contains(w.Body.String(), tc.expectedError) {
				t.Errorf("Expected error containing '%s', got: %s", tc.expectedError, w.Body.String())
			}
		})
	}
}

// TestAuthHandler_Register_InvalidEmail tests email validation
func TestAuthHandler_Register_InvalidEmail(t *testing.T) {
	testDB := SetupTestDB(t)
	defer testDB.TeardownTestDB()

	userRepo := user.NewUserRepository(testDB.DB)
	sessionRepo := session.NewSessionRepository(testDB.DB)
	imageRepo := repository.NewImageRepository(testDB.DB)
	authHandler := handlers.NewAuthHandler(userRepo, sessionRepo, imageRepo)

	fields := map[string]string{
		"email":         "invalid-email", // Invalid email format
		"password":      "SecurePass123",
		"first_name":    "John",
		"last_name":     "Doe",
		"date_of_birth": "1990-01-01",
		"gender":        "male",
	}

	req, _ := createMultipartRequest(t, fields, "", "")
	w := httptest.NewRecorder()
	authHandler.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestAuthHandler_Register_WeakPassword tests password strength validation
func TestAuthHandler_Register_WeakPassword(t *testing.T) {
	testDB := SetupTestDB(t)
	defer testDB.TeardownTestDB()

	userRepo := user.NewUserRepository(testDB.DB)
	sessionRepo := session.NewSessionRepository(testDB.DB)
	imageRepo := repository.NewImageRepository(testDB.DB)
	authHandler := handlers.NewAuthHandler(userRepo, sessionRepo, imageRepo)

	testCases := []struct {
		name     string
		password string
	}{
		{"Too short", "Pass1"},
		{"No digits", "PasswordOnly"},
		{"No letters", "12345678"},
		{"Empty", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fields := map[string]string{
				"email":         "test@example.com",
				"password":      tc.password,
				"first_name":    "John",
				"last_name":     "Doe",
				"date_of_birth": "1990-01-01",
				"gender":        "male",
			}

			req, _ := createMultipartRequest(t, fields, "", "")
			w := httptest.NewRecorder()
			authHandler.Register(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d for weak password, got %d", http.StatusBadRequest, w.Code)
			}
		})
	}
}

// TestAuthHandler_Register_InvalidAge tests age validation
func TestAuthHandler_Register_InvalidAge(t *testing.T) {
	testDB := SetupTestDB(t)
	defer testDB.TeardownTestDB()

	userRepo := user.NewUserRepository(testDB.DB)
	sessionRepo := session.NewSessionRepository(testDB.DB)
	imageRepo := repository.NewImageRepository(testDB.DB)
	authHandler := handlers.NewAuthHandler(userRepo, sessionRepo, imageRepo)

	testCases := []struct {
		name        string
		dateOfBirth string
	}{
		{"Too young", time.Now().AddDate(-10, 0, 0).Format("2006-01-02")}, // 10 years old
		{"Too old", time.Now().AddDate(-130, 0, 0).Format("2006-01-02")},  // 130 years old
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fields := map[string]string{
				"email":         "test@example.com",
				"password":      "SecurePass123",
				"first_name":    "John",
				"last_name":     "Doe",
				"date_of_birth": tc.dateOfBirth,
				"gender":        "male",
			}

			req, _ := createMultipartRequest(t, fields, "", "")
			w := httptest.NewRecorder()
			authHandler.Register(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("Expected status %d for invalid age, got %d", http.StatusBadRequest, w.Code)
			}

			if !strings.Contains(w.Body.String(), "between 13 and 120") {
				t.Errorf("Expected age error message, got: %s", w.Body.String())
			}
		})
	}
}

// TestAuthHandler_Register_DuplicateEmail tests duplicate email rejection
func TestAuthHandler_Register_DuplicateEmail(t *testing.T) {
	testDB := SetupTestDB(t)
	defer testDB.TeardownTestDB()

	userRepo := user.NewUserRepository(testDB.DB)
	sessionRepo := session.NewSessionRepository(testDB.DB)
	imageRepo := repository.NewImageRepository(testDB.DB)
	authHandler := handlers.NewAuthHandler(userRepo, sessionRepo, imageRepo)

	// Create first user
	fields := map[string]string{
		"email":         "duplicate@example.com",
		"password":      "SecurePass123",
		"first_name":    "John",
		"last_name":     "Doe",
		"date_of_birth": "1990-01-01",
		"gender":        "male",
		"nickname":      "john1",
	}

	req, _ := createMultipartRequest(t, fields, "", "")
	w := httptest.NewRecorder()
	authHandler.Register(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("First registration failed: %d", w.Code)
	}

	// Try to create second user with same email
	fields["nickname"] = "john2" // Different nickname
	req, _ = createMultipartRequest(t, fields, "", "")
	w = httptest.NewRecorder()
	authHandler.Register(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status %d for duplicate email, got %d", http.StatusConflict, w.Code)
	}

	if !strings.Contains(w.Body.String(), "Email is already taken") {
		t.Errorf("Expected 'Email is already taken' error, got: %s", w.Body.String())
	}
}

// TestAuthHandler_Register_DuplicateNickname tests duplicate nickname rejection
func TestAuthHandler_Register_DuplicateNickname(t *testing.T) {
	testDB := SetupTestDB(t)
	defer testDB.TeardownTestDB()

	userRepo := user.NewUserRepository(testDB.DB)
	sessionRepo := session.NewSessionRepository(testDB.DB)
	imageRepo := repository.NewImageRepository(testDB.DB)
	authHandler := handlers.NewAuthHandler(userRepo, sessionRepo, imageRepo)

	// Create first user
	fields := map[string]string{
		"email":         "user1@example.com",
		"password":      "SecurePass123",
		"first_name":    "John",
		"last_name":     "Doe",
		"date_of_birth": "1990-01-01",
		"gender":        "male",
		"nickname":      "coolnickname",
	}

	req, _ := createMultipartRequest(t, fields, "", "")
	w := httptest.NewRecorder()
	authHandler.Register(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("First registration failed: %d", w.Code)
	}

	// Try to create second user with same nickname
	fields["email"] = "user2@example.com" // Different email
	req, _ = createMultipartRequest(t, fields, "", "")
	w = httptest.NewRecorder()
	authHandler.Register(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status %d for duplicate nickname, got %d", http.StatusConflict, w.Code)
	}

	if !strings.Contains(w.Body.String(), "Nickname is already taken") {
		t.Errorf("Expected 'Nickname is already taken' error, got: %s", w.Body.String())
	}
}

// TestAuthHandler_Register_AvatarTooLarge tests avatar size limit
func TestAuthHandler_Register_AvatarTooLarge(t *testing.T) {
	testDB := SetupTestDB(t)
	defer testDB.TeardownTestDB()

	userRepo := user.NewUserRepository(testDB.DB)
	sessionRepo := session.NewSessionRepository(testDB.DB)
	imageRepo := repository.NewImageRepository(testDB.DB)
	authHandler := handlers.NewAuthHandler(userRepo, sessionRepo, imageRepo)

	// Create large avatar (6MB - exceeds 5MB limit)
	avatarPath := createTestAvatar(t, "large-avatar.png", 6*1024)

	fields := map[string]string{
		"email":         "large@example.com",
		"password":      "SecurePass123",
		"first_name":    "John",
		"last_name":     "Doe",
		"date_of_birth": "1990-01-01",
		"gender":        "male",
	}

	req, _ := createMultipartRequest(t, fields, "avatar", avatarPath)
	w := httptest.NewRecorder()
	authHandler.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d for large avatar, got %d", http.StatusBadRequest, w.Code)
	}

	if !strings.Contains(w.Body.String(), "exceeds 5 MB limit") {
		t.Errorf("Expected size limit error, got: %s", w.Body.String())
	}

	// Verify user was NOT created
	var count int
	testDB.DB.QueryRow("SELECT COUNT(*) FROM user WHERE email = ?", "large@example.com").Scan(&count)
	if count != 0 {
		t.Errorf("User should not have been created, found %d users", count)
	}
}

// TestAuthHandler_Register_InvalidGender tests gender validation
func TestAuthHandler_Register_InvalidGender(t *testing.T) {
	testDB := SetupTestDB(t)
	defer testDB.TeardownTestDB()

	userRepo := user.NewUserRepository(testDB.DB)
	sessionRepo := session.NewSessionRepository(testDB.DB)
	imageRepo := repository.NewImageRepository(testDB.DB)
	authHandler := handlers.NewAuthHandler(userRepo, sessionRepo, imageRepo)

	fields := map[string]string{
		"email":         "test@example.com",
		"password":      "SecurePass123",
		"first_name":    "John",
		"last_name":     "Doe",
		"date_of_birth": "1990-01-01",
		"gender":        "invalid_gender",
	}

	req, _ := createMultipartRequest(t, fields, "", "")
	w := httptest.NewRecorder()
	authHandler.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	if !strings.Contains(w.Body.String(), "Gender must be one of") {
		t.Errorf("Expected gender validation error, got: %s", w.Body.String())
	}
}

// TestAuthHandler_Register_AboutMeTooLong tests about_me length validation
func TestAuthHandler_Register_AboutMeTooLong(t *testing.T) {
	testDB := SetupTestDB(t)
	defer testDB.TeardownTestDB()

	userRepo := user.NewUserRepository(testDB.DB)
	sessionRepo := session.NewSessionRepository(testDB.DB)
	imageRepo := repository.NewImageRepository(testDB.DB)
	authHandler := handlers.NewAuthHandler(userRepo, sessionRepo, imageRepo)

	// Create about_me text longer than 500 characters
	longText := strings.Repeat("a", 501)

	fields := map[string]string{
		"email":         "test@example.com",
		"password":      "SecurePass123",
		"first_name":    "John",
		"last_name":     "Doe",
		"date_of_birth": "1990-01-01",
		"gender":        "male",
		"about_me":      longText,
	}

	req, _ := createMultipartRequest(t, fields, "", "")
	w := httptest.NewRecorder()
	authHandler.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	if !strings.Contains(w.Body.String(), "About me must be under 500") {
		t.Errorf("Expected about_me length error, got: %s", w.Body.String())
	}
}

// TestAuthHandler_Register_PrivacyFlag tests is_private field
func TestAuthHandler_Register_PrivacyFlag(t *testing.T) {
	testDB := SetupTestDB(t)
	defer testDB.TeardownTestDB()

	userRepo := user.NewUserRepository(testDB.DB)
	sessionRepo := session.NewSessionRepository(testDB.DB)
	imageRepo := repository.NewImageRepository(testDB.DB)
	authHandler := handlers.NewAuthHandler(userRepo, sessionRepo, imageRepo)

	testCases := []struct {
		name      string
		isPrivate string
		expected  bool
	}{
		{"Private true", "true", true},
		{"Private false", "false", false},
		{"Private 1", "1", true},
		{"Private 0", "0", false},
		{"Private empty", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Generate unique email without spaces
			cleanName := strings.ReplaceAll(strings.ToLower(tc.name), " ", "_")
			fields := map[string]string{
				"email":         "test_" + cleanName + "@example.com",
				"password":      "SecurePass123",
				"first_name":    "John",
				"last_name":     "Doe",
				"date_of_birth": "1990-01-01",
				"gender":        "male",
				"nickname":      "user_" + cleanName,
			}
			if tc.isPrivate != "" {
				fields["is_private"] = tc.isPrivate
			}

			req, _ := createMultipartRequest(t, fields, "", "")
			w := httptest.NewRecorder()
			authHandler.Register(w, req)

			if w.Code != http.StatusCreated {
				t.Fatalf("Registration failed: %d", w.Code)
			}

			var response models.LoginResponse
			json.NewDecoder(w.Body).Decode(&response)

			if response.User.IsPrivate != tc.expected {
				t.Errorf("Expected is_private=%v, got %v", tc.expected, response.User.IsPrivate)
			}
		})
	}
}

// TestAuthHandler_Register_MethodNotAllowed tests non-POST requests
func TestAuthHandler_Register_MethodNotAllowed(t *testing.T) {
	testDB := SetupTestDB(t)
	defer testDB.TeardownTestDB()

	userRepo := user.NewUserRepository(testDB.DB)
	sessionRepo := session.NewSessionRepository(testDB.DB)
	imageRepo := repository.NewImageRepository(testDB.DB)
	authHandler := handlers.NewAuthHandler(userRepo, sessionRepo, imageRepo)

	methods := []string{"GET", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/auth/register", nil)
			w := httptest.NewRecorder()
			authHandler.Register(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Expected status %d for %s method, got %d",
					http.StatusMethodNotAllowed, method, w.Code)
			}
		})
	}
}

// BenchmarkAuthHandler_Register benchmarks the registration handler
func BenchmarkAuthHandler_Register(b *testing.B) {
	// Note: This creates one test DB for all benchmark iterations
	// For true isolation, each iteration would need its own DB (slower but more realistic)
	testDB := &TestDB{}
	tempDir := b.TempDir()
	dbPath := filepath.Join(tempDir, "bench_test.db")

	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	if err := dbmigrate.Migrate(db); err != nil {
		b.Fatal(err)
	}

	testDB.DB = db
	userRepo := user.NewUserRepository(testDB.DB)
	sessionRepo := session.NewSessionRepository(testDB.DB)
	imageRepo := repository.NewImageRepository(testDB.DB)
	authHandler := handlers.NewAuthHandler(userRepo, sessionRepo, imageRepo)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		fields := map[string]string{
			"email":         fmt.Sprintf("bench%d@example.com", i),
			"password":      "SecurePass123",
			"first_name":    "Bench",
			"last_name":     "User",
			"date_of_birth": "1990-01-01",
			"gender":        "male",
			"nickname":      fmt.Sprintf("bench%d", i),
		}

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		for key, val := range fields {
			writer.WriteField(key, val)
		}
		writer.Close()

		req := httptest.NewRequest("POST", "/api/auth/register", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()
		authHandler.Register(w, req)
	}
}