package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"forum/models"
	"forum/repository"
	"forum/repository/session"
	"forum/repository/user"
	"forum/utils"
)

var (
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	GitHubClientID     string
	GitHubClientSecret string
	GitHubRedirectURL  string // e.g., "http://localhost:8080/auth/github/callback"
)

// OAuthHandler handles OAuth authentication
type OAuthHandler struct {
	UserRepo    *user.UserRepository
	SessionRepo *session.SessionRepository
	AuthHandler *AuthHandler
}

// NewOAuthHandler creates a new OAuthHandler
func NewOAuthHandler(userRepo *user.UserRepository, sessionRepo *session.SessionRepository, authHandler *AuthHandler) *OAuthHandler {
	return &OAuthHandler{
		UserRepo:    userRepo,
		SessionRepo: sessionRepo,
		AuthHandler: authHandler,
	}
}

// Google OAuth handlers
func (h *OAuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	state := h.generateState()

	// Store state in session for verification
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   300, // 5 minutes
		HttpOnly: true,
		Secure:   false, // true in production
		SameSite: http.SameSiteLaxMode,
	})
	GoogleClientID = os.Getenv("GOOGLE_CLIENT_ID")
	GoogleRedirectURL = os.Getenv("GOOGLE_REDIRECT_URL")
	authURL := fmt.Sprintf(
		"https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=email+profile&state=%s&access_type=offline&prompt=consent",
		GoogleClientID,
		url.QueryEscape(GoogleRedirectURL),
		state,
	)

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (h *OAuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	// Verify state parameter
	if !h.verifyState(r) {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Authorization code not provided", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	// This now returns accessToken, refreshToken, and expiresIn
	tokenResp, err := h.exchangeGoogleCode(code)
	if err != nil {
		log.Printf("Failed to exchange Google code: %v", err)
		http.Error(w, "Failed to exchange authorization code", http.StatusInternalServerError)
		return
	}

	tokenExpiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	// Get user info from Google
	userInfo, err := h.getGoogleUserInfo(tokenResp.AccessToken) // Pass the access token for user info
	if err != nil {
		log.Printf("Failed to get Google user info: %v", err)
		http.Error(w, "Failed to get user information", http.StatusInternalServerError)
		return
	}

	// Handle user creation/login - NOW PASSING ALL REQUIRED PARAMETERS
	user, err := h.handleOAuthUser(userInfo, "google", tokenResp.AccessToken, tokenResp.RefreshToken, tokenExpiresAt)
	if err != nil {
		log.Printf("Failed to handle OAuth user: %v", err)
		http.Error(w, "Failed to process user", http.StatusInternalServerError)
		return
	}

	// Create session and redirect
	_, err = h.AuthHandler.createUserSession(w, r, user)
	if err != nil {
		utils.ErrorResponse(w, "Failed to create session", http.StatusInternalServerError)
		return
	}
	log.Printf("Redirecting to /user/feed for user: %s", user.Email)
	http.Redirect(w, r, "http://localhost:8081/user/feed", http.StatusFound)
}

// GitHub OAuth handlers
func (h *OAuthHandler) GitHubLogin(w http.ResponseWriter, r *http.Request) {
	state := h.generateState()

	// Store state in session for verification
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   300, // 5 minutes
		HttpOnly: true,
		Secure:   false, // true in production
		SameSite: http.SameSiteLaxMode,
	})
	GitHubClientID = os.Getenv("GITHUB_CLIENT_ID")
	GitHubRedirectURL = os.Getenv("GITHUB_REDIRECT_URL")
	authURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=user:email%%20offline_access&state=%s",
		GitHubClientID,
		url.QueryEscape(GitHubRedirectURL),
		state,
	)

	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

func (h *OAuthHandler) GitHubCallback(w http.ResponseWriter, r *http.Request) {
	// Verify state parameter
	if !h.verifyState(r) {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Authorization code not provided", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	tokenResp, err := h.exchangeGitHubCode(code)
	if err != nil {
		log.Printf("Failed to exchange GitHub code: %v", err)
		http.Error(w, "Failed to exchange authorization code", http.StatusInternalServerError)
		return
	}

	tokenExpiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	// Get user info from GitHub
	userInfo, err := h.getGitHubUserInfo(tokenResp.AccessToken) // Pass the access token for user info
	if err != nil {
		log.Printf("Failed to get GitHub user info: %v", err)
		http.Error(w, "Failed to get user information", http.StatusInternalServerError)
		return
	}

	// Handle user creation/login - NOW PASSING ALL REQUIRED PARAMETERS
	user, err := h.handleOAuthUser(userInfo, "github", tokenResp.AccessToken, tokenResp.RefreshToken, tokenExpiresAt)
	if err != nil {
		log.Printf("Failed to handle OAuth user: %v", err)
		http.Error(w, "Failed to process user", http.StatusInternalServerError)
		return
	}

	// Create session and redirect
	_, err = h.AuthHandler.createUserSession(w, r, user)
	if err != nil {
		utils.ErrorResponse(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	log.Printf("Redirecting to /user/feed for user: %s", user.Email)
	http.Redirect(w, r, "http://localhost:8081/user/feed", http.StatusFound)
}

// OAuthTokenResponse holds the common fields from OAuth token endpoints
type OAuthTokenResponse struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int // in seconds
}

// Helper methods
func (h *OAuthHandler) generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func (h *OAuthHandler) verifyState(r *http.Request) bool {
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil {
		return false
	}

	stateParam := r.URL.Query().Get("state")
	return stateCookie.Value == stateParam
}

// exchangeGoogleCode now returns OAuthTokenResponse
func (h *OAuthHandler) exchangeGoogleCode(code string) (*OAuthTokenResponse, error) {
	GoogleClientID = os.Getenv("GOOGLE_CLIENT_ID")
	GoogleClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	GoogleRedirectURL = os.Getenv("GOOGLE_REDIRECT_URL")
	data := url.Values{
		"client_id":     {GoogleClientID},
		"client_secret": {GoogleClientSecret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {GoogleRedirectURL},
	}

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	accessToken, ok := result["access_token"].(string)
	if !ok {
		return nil, fmt.Errorf("access token not found in response")
	}

	refreshToken, _ := result["refresh_token"].(string) // May be empty for initial exchange

	var expiresIn int
	if expiresInFloat, ok := result["expires_in"].(float64); ok {
		expiresIn = int(expiresInFloat)
	} else {
		// Default expiry if not provided, or handle error
		log.Println("Warning: expires_in not found in Google response, defaulting to 3600 seconds.")
		expiresIn = 3600 // Default value, adjust as per Google's actual token validity
	}

	return &OAuthTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

func (h *OAuthHandler) getGoogleUserInfo(token string) (*models.OAuthUserInfo, error) {
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var googleUser struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		VerifiedEmail bool   `json:"verified_email"`
	}

	if err := json.Unmarshal(body, &googleUser); err != nil {
		return nil, err
	}

	return &models.OAuthUserInfo{
		ID:        googleUser.ID,
		Email:     googleUser.Email,
		Name:      googleUser.Name,
		Username:  "", // Will be generated
		AvatarURL: googleUser.Picture,
	}, nil
}

// exchangeGitHubCode now returns OAuthTokenResponse
func (h *OAuthHandler) exchangeGitHubCode(code string) (*OAuthTokenResponse, error) {
	GitHubClientID = os.Getenv("GITHUB_CLIENT_ID")
	GitHubClientSecret = os.Getenv("GITHUB_CLIENT_SECRET")
	data := url.Values{
		"client_id":     {GitHubClientID},
		"client_secret": {GitHubClientSecret},
		"code":          {code},
	}

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Body = io.NopCloser(strings.NewReader(data.Encode()))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	accessToken, ok := result["access_token"].(string)
	if !ok {
		return nil, fmt.Errorf("access token not found in response")
	}

	refreshToken, _ := result["refresh_token"].(string) // May or may not be provided by GitHub depending on scope/setup

	var expiresIn int
	if expiresInFloat, ok := result["expires_in"].(float64); ok {
		expiresIn = int(expiresInFloat)
	} else {
		log.Println("Warning: expires_in not found in GitHub response, defaulting to 3600 seconds.")
		expiresIn = 3600 // Default value, adjust as per GitHub's actual token validity
	}

	return &OAuthTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}

func (h *OAuthHandler) getGitHubUserInfo(token string) (*models.OAuthUserInfo, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var githubUser struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := json.Unmarshal(body, &githubUser); err != nil {
		return nil, err
	}

	// GitHub doesn't always provide email in the user endpoint
	email := githubUser.Email
	if email == "" {
		email, _ = h.getGitHubUserEmail(token)
	}

	return &models.OAuthUserInfo{
		ID:        fmt.Sprintf("%d", githubUser.ID),
		Email:     email,
		Name:      githubUser.Name,
		Username:  githubUser.Login,
		AvatarURL: githubUser.AvatarURL,
	}, nil
}

func (h *OAuthHandler) getGitHubUserEmail(token string) (string, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "token "+token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var emails []struct {
		Email   string `json:"email"`
		Primary bool   `json:"primary"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	for _, email := range emails {
		if email.Primary {
			return email.Email, nil
		}
	}

	if len(emails) > 0 {
		return emails[0].Email, nil
	}

	return "", fmt.Errorf("no email found")
}

// handleOAuthUser now accepts accessToken, refreshToken, and tokenExpiresAt
func (h *OAuthHandler) handleOAuthUser(userInfo *models.OAuthUserInfo, provider, accessToken, refreshToken string, tokenExpiresAt time.Time) (*models.User, error) {
	// Check if user exists by email
	user, err := h.UserRepo.GetByEmail(userInfo.Email)
	if err != nil && err != repository.ErrUserNotFound {
		return nil, err
	}

	// If user exists
	if err == nil {
		// Check if provider is already linked
		providerLinked, err := h.UserRepo.IsProviderLinked(user.ID, provider)
		if err != nil {
			return nil, fmt.Errorf("error checking provider link: %w", err)
		}

		// If not linked, link the new provider
		if !providerLinked {
			err := h.UserRepo.LinkOAuthProvider(user.ID, provider, userInfo.ID, accessToken, refreshToken, tokenExpiresAt)
			if err != nil {
				return nil, fmt.Errorf("failed to link new OAuth provider: %w", err)
			}
		}

		// Optionally update tokens here if needed
		return user, nil
	}

	// User does not exist, create new one
	username := userInfo.Username
	if username == "" {
		username = h.generateUsernameFromEmail(userInfo.Email)
	}
	username = h.ensureUniqueUsername(username)

	reg := models.UserRegistration{
		Username: username,
		Email:    userInfo.Email,
		Password: "", // No password for OAuth
	}

	return h.UserRepo.CreateOAuthUser(reg, provider, userInfo.ID, userInfo.AvatarURL, accessToken, refreshToken, tokenExpiresAt)
}

func (h *OAuthHandler) generateUsernameFromEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) > 0 {
		// Clean up username: remove special characters, limit length
		username := strings.ReplaceAll(parts[0], ".", "")
		username = strings.ReplaceAll(username, "+", "")
		if len(username) > 20 {
			username = username[:20]
		}
		return username
	}
	return "user" + fmt.Sprintf("%d", time.Now().Unix())
}

func (h *OAuthHandler) ensureUniqueUsername(baseUsername string) string {
	username := baseUsername
	counter := 1

	for {
		// Check if username exists
		_, err := h.UserRepo.GetByUsername(username)
		if err == repository.ErrUserNotFound {
			// Username is available
			return username
		}

		// Username taken, try with counter
		username = fmt.Sprintf("%s%d", baseUsername, counter)
		counter++

		// Prevent infinite loop (though highly unlikely with a good UUID/timestamp strategy)
		if counter > 1000 {
			return fmt.Sprintf("user%d", time.Now().Unix())
		}
	}
}
