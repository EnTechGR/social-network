package routes

import (
	"database/sql"
	"net/http"

	"forum/handlers"
	"forum/middleware"
	"forum/repository"
	"forum/repository/message"
	"forum/repository/session"
	"forum/repository/user"
	"forum/websocket"
)

func SetupRoutes(db *sql.DB) http.Handler {
	// Create repositories
	userRepo := user.NewUserRepository(db)
	sessionRepo := session.NewSessionRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	postRepo := repository.NewPostRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	reactionRepo := repository.NewReactionRepository(db)
	imageRepo := repository.NewImageRepository(db)
	notificationRepo := repository.NewNotificationRepository(db)
	messageRepo := message.NewMessageRepository(db)

	// ✅ Create and start WebSocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// Create handlers
	authHandler := handlers.NewAuthHandler(userRepo, sessionRepo)
	oauthHandler := handlers.NewOAuthHandler(userRepo, sessionRepo, authHandler)
	categoryHandler := handlers.NewCategoryHandler(categoryRepo, postRepo, imageRepo)
	postHandler := handlers.NewPostHandler(postRepo)
	myPostsHandler := handlers.NewMyPostsHandler(postRepo, commentRepo, reactionRepo, imageRepo)
	likedPostsHandler := handlers.NewLikedPostsHandler(postRepo, commentRepo, reactionRepo, imageRepo)
	commentHandler := handlers.NewCommentHandler(commentRepo, postRepo, notificationRepo)
	reactionHandler := handlers.NewReactionHandler(reactionRepo, postRepo, commentRepo, notificationRepo)
	imageHandler := handlers.NewImageHandler(imageRepo, postRepo)
	guestHandler := handlers.NewGuestHandler(categoryRepo, postRepo, commentRepo, reactionRepo, imageRepo)
	notificationHandler := handlers.NewNotificationHandler(notificationRepo)
	messageHandler := handlers.NewMessageHandler(messageRepo, hub) // ✅ Pass hub to handler

	// Create middleware
	registerLimiter := middleware.NewRateLimiter()
	authMiddleware := middleware.NewAuthMiddleware(sessionRepo, userRepo)
	corsMiddleware := middleware.NewCORSMiddleware("http://localhost:8081")

	// Create router
	mux := http.NewServeMux()

	// Serve uploaded images from the API container
	fs := http.FileServer(http.Dir("./uploads"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Public routes
	mux.Handle("/forum/api/categories", corsMiddleware.Handler(http.HandlerFunc(categoryHandler.GetCategories)))
	mux.Handle("/forum/api/category", corsMiddleware.Handler(http.HandlerFunc(categoryHandler.GetCategoryByID)))
	mux.Handle("/forum/api/feed", corsMiddleware.Handler(http.HandlerFunc(guestHandler.GetGuestData)))

	// Authentication routes (guest only)
	guestOnly := func(h http.Handler) http.Handler {
		return corsMiddleware.Handler(authMiddleware.RequireGuest(h))
	}

	mux.Handle("/forum/api/register", guestOnly(http.HandlerFunc(registerLimiter.Limit(authHandler.Register))))
	mux.Handle("/forum/api/session/login", guestOnly(http.HandlerFunc(authHandler.Login)))

	// OAuth routes (guest only)
	mux.Handle("/auth/google/login", guestOnly(http.HandlerFunc(oauthHandler.GoogleLogin)))
	// OAuth callbacks do not need RequireGuest or CSRF, as the 'state' parameter handles CSRF
	mux.Handle("/auth/google/callback", corsMiddleware.Handler(http.HandlerFunc(oauthHandler.GoogleCallback)))
	mux.Handle("/auth/github/login", guestOnly(http.HandlerFunc(oauthHandler.GitHubLogin)))
	mux.Handle("/oauth/github/callback", corsMiddleware.Handler(http.HandlerFunc(oauthHandler.GitHubCallback)))

	// Session management routes
	mux.Handle("/forum/api/session/logout", corsMiddleware.Handler(http.HandlerFunc(authHandler.Logout)))
	mux.Handle("/forum/api/session/verify", corsMiddleware.Handler(http.HandlerFunc(authHandler.VerifySession)))

	// ✅ WebSocket endpoint (requires authentication, no CSRF needed for WebSocket upgrade)
	mux.Handle("/ws", corsMiddleware.Handler(authMiddleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		websocket.HandleWebSocket(hub, w, r)
	}))))

	// Protected routes with CSRF
	protected := func(h http.Handler) http.Handler {
		// Ensure CSRF middleware is active for protected routes
		return corsMiddleware.Handler(authMiddleware.RequireAuth(authMiddleware.CSRF(h)))
	}

	// Protected user routes
	mux.Handle("/forum/api/posts/create", protected(http.HandlerFunc(postHandler.CreatePost)))
	mux.Handle("/forum/api/posts/delete/", protected(http.HandlerFunc(postHandler.DeletePost)))
	mux.Handle("/forum/api/posts/edit-title/", protected(http.HandlerFunc(postHandler.EditPostTitle)))
	mux.Handle("/forum/api/posts/edit-content/", protected(http.HandlerFunc(postHandler.EditPostContent)))
	mux.Handle("/forum/api/user/posts", protected(http.HandlerFunc(myPostsHandler.GetMyPosts)))
	mux.Handle("/forum/api/user/liked", protected(http.HandlerFunc(likedPostsHandler.GetLikedPosts)))
	mux.Handle("/forum/api/user/disliked", protected(http.HandlerFunc(likedPostsHandler.GetDislikedPosts)))
	mux.Handle("/forum/api/comments/create", protected(http.HandlerFunc(commentHandler.CreateComment)))
	mux.Handle("/forum/api/comments/edit/", protected(http.HandlerFunc(commentHandler.EditComment)))
	mux.Handle("/forum/api/comments/delete/", protected(http.HandlerFunc(commentHandler.DeleteComment)))
	mux.Handle("/forum/api/react", protected(http.HandlerFunc(reactionHandler.CreateReact)))
	mux.Handle("/forum/api/images/upload", protected(http.HandlerFunc(imageHandler.Upload)))
	mux.Handle("/forum/api/user/commented", protected(http.HandlerFunc(myPostsHandler.GetCommentedPosts)))
	mux.Handle("/forum/api/images/delete/", protected(http.HandlerFunc(imageHandler.DeleteImagesByPost)))

	// Additional protected routes for user management
	mux.Handle("/forum/api/user/profile", protected(http.HandlerFunc(authHandler.GetProfile)))
	mux.Handle("/forum/api/notifications", protected(http.HandlerFunc(notificationHandler.GetNotifications)))
	mux.Handle("/forum/api/notifications/delete/", protected(http.HandlerFunc(notificationHandler.HideNotification)))
	mux.Handle("/forum/api/session/logout-all", protected(http.HandlerFunc(authHandler.LogoutAll)))

	// Protected message routes
	mux.Handle("/forum/api/messages/send", protected(http.HandlerFunc(messageHandler.SendMessage)))
	mux.Handle("/forum/api/messages/conversation", protected(http.HandlerFunc(messageHandler.GetConversation)))
	mux.Handle("/forum/api/messages/conversations", protected(http.HandlerFunc(messageHandler.GetConversations)))
	mux.Handle("/forum/api/messages/users", protected(http.HandlerFunc(messageHandler.GetAllUsers)))
	mux.Handle("/forum/api/messages/users-for-chat", protected(http.HandlerFunc(messageHandler.GetUsersForChat)))
	mux.Handle("/forum/api/messages/unread-count", protected(http.HandlerFunc(messageHandler.GetUnreadCount)))
	mux.Handle("/forum/api/messages/mark-read/", protected(http.HandlerFunc(messageHandler.MarkAsRead)))
	mux.Handle("/forum/api/messages/delete/", protected(http.HandlerFunc(messageHandler.DeleteMessage)))

	return authMiddleware.Authenticate(mux)
}