package routes

import (
	"database/sql"
	"net/http"

	"social-network/handlers"
	"social-network/middleware"
	"social-network/repository"
	"social-network/repository/message"
	"social-network/repository/session"
	"social-network/repository/user_repository"
	"social-network/websocket"

	// IMPORTANT SWAGGER IMPORTS
	_ "social-network/docs" // <--- Ensure this points to your generated docs package!

	httpSwagger "github.com/swaggo/http-swagger"
)

func SetupRoutes(db *sql.DB) http.Handler {
	// Create repositories
	userRepo := user_repository.NewUserRepository(db)
	sessionRepo := session.NewSessionRepository(db)
	followRepo := repository.NewFollowRepository(db)
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
	authHandler := handlers.NewAuthHandler(userRepo, sessionRepo, imageRepo)
	oauthHandler := handlers.NewOAuthHandler(userRepo, sessionRepo, authHandler)
	userHandler := handlers.NewUserHandler(userRepo, imageRepo) // ✅ User handler for profile management
	followHandler := handlers.NewFollowHandler(followRepo, userRepo)
	postHandler := handlers.NewPostHandler(postRepo, imageRepo)
	myPostsHandler := handlers.NewMyPostsHandler(postRepo, commentRepo, reactionRepo, imageRepo)
	likedPostsHandler := handlers.NewLikedPostsHandler(postRepo, commentRepo, reactionRepo, imageRepo)
	commentHandler := handlers.NewCommentHandler(commentRepo, postRepo, notificationRepo, hub)
	commentHandler.SetImageRepo(imageRepo) // ✅ Enable image uploads for comments
	reactionHandler := handlers.NewReactionHandler(reactionRepo, postRepo, commentRepo, notificationRepo, hub)
	imageHTTPHandler := handlers.NewImageHTTPHandler(imageRepo, postRepo)

	notificationHandler := handlers.NewNotificationHandler(notificationRepo, hub)
	messageHandler := handlers.NewMessageHandler(messageRepo, hub)     // ✅ Pass hub to handler
	chatImageHandler := handlers.NewChatImageHandler(messageRepo, hub) // ✅ Chat image handler

	// Create middleware
	registerLimiter := middleware.NewRateLimiter()
	authMiddleware := middleware.NewAuthMiddleware(sessionRepo, userRepo)
	corsMiddleware := middleware.NewCORSMiddleware("http://localhost:8081")

	// 1. Create the API Mux (All application routes go here)
	apiMux := http.NewServeMux()

	// Serve uploaded images from the API container
	fs := http.FileServer(http.Dir("./uploads"))
	apiMux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Authentication routes (guest only)
	guestOnly := func(h http.Handler) http.Handler {
		return corsMiddleware.Handler(authMiddleware.RequireGuest(h))
	}

	apiMux.Handle("/api/v1/register", guestOnly(http.HandlerFunc(registerLimiter.Limit(authHandler.Register))))
	apiMux.Handle("/api/v1/login", guestOnly(http.HandlerFunc(authHandler.Login)))

	// OAuth routes (guest only)
	apiMux.Handle("/auth/google/login", guestOnly(http.HandlerFunc(oauthHandler.GoogleLogin)))
	// OAuth callbacks do not need RequireGuest or CSRF, as the 'state' parameter handles CSRF
	apiMux.Handle("/auth/google/callback", corsMiddleware.Handler(http.HandlerFunc(oauthHandler.GoogleCallback)))
	apiMux.Handle("/auth/github/login", guestOnly(http.HandlerFunc(oauthHandler.GitHubLogin)))
	apiMux.Handle("/oauth/github/callback", corsMiddleware.Handler(http.HandlerFunc(oauthHandler.GitHubCallback)))

	// Session management routes
	apiMux.Handle("/api/v1/logout", corsMiddleware.Handler(http.HandlerFunc(authHandler.Logout)))
	apiMux.Handle("/api/v1/verify", corsMiddleware.Handler(http.HandlerFunc(authHandler.VerifySession)))

	// ✅ WebSocket endpoint (requires authentication, no CSRF needed for WebSocket upgrade)
	apiMux.Handle("/ws", corsMiddleware.Handler(authMiddleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		websocket.HandleWebSocket(hub, w, r)
	}))))

	// Protected routes with CSRF
	protected := func(h http.Handler) http.Handler {
		// Ensure CSRF middleware is active for protected routes
		return corsMiddleware.Handler(authMiddleware.RequireAuth(authMiddleware.CSRF(h)))
	}

	// Protected user routes
	apiMux.Handle("/api/v1/posts/create", protected(http.HandlerFunc(postHandler.CreatePost)))
	apiMux.Handle("/api/v1/posts/delete/", protected(http.HandlerFunc(postHandler.DeletePost)))
	apiMux.Handle("/api/v1/posts/edit-title/", protected(http.HandlerFunc(postHandler.EditPostTitle)))
	apiMux.Handle("/api/v1/posts/edit-content/", protected(http.HandlerFunc(postHandler.EditPostContent)))
	apiMux.Handle("/api/v1/user/posts", protected(http.HandlerFunc(myPostsHandler.GetMyPosts)))
	apiMux.Handle("/api/v1/user/liked", protected(http.HandlerFunc(likedPostsHandler.GetLikedPosts)))
	apiMux.Handle("/api/v1/user/disliked", protected(http.HandlerFunc(likedPostsHandler.GetDislikedPosts)))
	apiMux.Handle("/api/v1/comments/create", protected(http.HandlerFunc(commentHandler.CreateComment)))
	apiMux.Handle("/api/v1/comments/edit/", protected(http.HandlerFunc(commentHandler.EditComment)))
	apiMux.Handle("/api/v1/comments/delete/", protected(http.HandlerFunc(commentHandler.DeleteComment)))
	apiMux.Handle("/api/v1/react", protected(http.HandlerFunc(reactionHandler.CreateReact)))
	apiMux.Handle("/api/v1/user/commented", protected(http.HandlerFunc(myPostsHandler.GetCommentedPosts)))
	// Image upload endpoints using new centralized handler
	apiMux.Handle("/api/v1/images/upload", protected(http.HandlerFunc(imageHTTPHandler.UploadPostImages)))
	apiMux.Handle("/api/v1/images/delete/", protected(http.HandlerFunc(imageHTTPHandler.DeletePostImages)))
	apiMux.Handle("/api/v1/user/avatar", protected(http.HandlerFunc(imageHTTPHandler.UploadAvatar)))

	// Additional protected routes for user management
	// User profile endpoint (handles both GET and PUT via method switching in handler)
	apiMux.Handle("/api/v1/user/profile", protected(http.HandlerFunc(userHandler.GetProfile)))
	apiMux.Handle("/api/v1/user/privacy", protected(http.HandlerFunc(userHandler.TogglePrivacy)))
	apiMux.Handle("/api/v1/users/", protected(http.HandlerFunc(userHandler.GetUserProfile)))

	apiMux.Handle("/api/v1/follow", protected(http.HandlerFunc(followHandler.FollowPublicUser)))
	apiMux.Handle("/api/v1/follow/accept/", protected(http.HandlerFunc(followHandler.AcceptFollowRequest)))
	apiMux.Handle("/api/v1/follow/requests", protected(http.HandlerFunc(followHandler.GetFollowRequests)))
	apiMux.Handle("/api/v1/follow/requests/pending", protected(http.HandlerFunc(followHandler.GetPendingFollowRequests)))
	apiMux.Handle("/api/v1/followers", protected(http.HandlerFunc(followHandler.GetFollowers)))
	apiMux.Handle("/api/v1/follower/delete/", protected(http.HandlerFunc(followHandler.Unfollow)))
	apiMux.Handle("/api/v1/followee/delete/", protected(http.HandlerFunc(followHandler.RemoveFollower)))
	apiMux.Handle("/api/v1/notifications", protected(http.HandlerFunc(notificationHandler.GetNotifications)))
	apiMux.Handle("/api/v1/notifications/delete/", protected(http.HandlerFunc(notificationHandler.HideNotification)))
	apiMux.Handle("/api/v1/logout-all", protected(http.HandlerFunc(authHandler.LogoutAll)))

	// Protected message routes
	apiMux.Handle("/api/v1/chat/send", protected(http.HandlerFunc(messageHandler.SendMessage)))
	apiMux.Handle("/api/v1/chat/conversation", protected(http.HandlerFunc(messageHandler.GetConversation)))
	apiMux.Handle("/api/v1/chat/conversations", protected(http.HandlerFunc(messageHandler.GetConversations)))
	apiMux.Handle("/api/v1/chat/users", protected(http.HandlerFunc(messageHandler.GetAllUsers)))
	apiMux.Handle("/api/v1/chat/users-for-chat", protected(http.HandlerFunc(messageHandler.GetUsersForChat)))
	apiMux.Handle("/api/v1/chat/unread-count", protected(http.HandlerFunc(messageHandler.GetUnreadCount)))
	apiMux.Handle("/api/v1/chat/mark-read/", protected(http.HandlerFunc(messageHandler.MarkAsRead)))
	apiMux.Handle("/api/v1/chat/delete/", protected(http.HandlerFunc(messageHandler.DeleteMessage)))

	// Protected chat image routes
	apiMux.Handle("/api/v1/chat/images/upload", protected(http.HandlerFunc(chatImageHandler.UploadChatImage)))
	apiMux.Handle("/api/v1/chat/images/message", protected(http.HandlerFunc(chatImageHandler.GetMessageImages)))
	apiMux.Handle("/api/v1/chat/images/gallery", protected(http.HandlerFunc(chatImageHandler.GetConversationGallery)))
	apiMux.Handle("/api/v1/chat/images/serve/", protected(http.HandlerFunc(chatImageHandler.ServeChatImage)))
	apiMux.Handle("/api/v1/chat/images/delete/", protected(http.HandlerFunc(chatImageHandler.DeleteChatImage)))
	apiMux.Handle("/api/v1/chat/images/stats", protected(http.HandlerFunc(chatImageHandler.GetUserImageStats)))

	// =========================================================================
	// 2. Wrap the API Mux with the authentication middleware
	apiHandler := authMiddleware.Authenticate(apiMux)

	// 3. Create the Root Mux (the final router that combines everything)
	rootMux := http.NewServeMux()

	// 4. Register the Swagger Routes (NOT WRAPPED by AuthMiddleware)
	// The Swagger UI needs to load static files and the doc.json file without auth checks.
	swaggerHandler := httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), // Pointer to where the JSON spec is served
	)

	// Handle the Swagger UI and static files
	rootMux.Handle("/swagger/", swaggerHandler)
	// Handle the JSON specification file
	rootMux.Handle("/swagger/doc.json", httpSwagger.WrapHandler)

	// 5. Register the Main API Handler
	// All requests not starting with /swagger/ are directed to the API handler,
	// which includes your authentication and CORS middlewares.
	rootMux.Handle("/", apiHandler)

	// 6. Return the final root handler
	return rootMux
	// =========================================================================
}
