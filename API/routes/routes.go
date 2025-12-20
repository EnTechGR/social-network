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
	
	// IMPORTANT SWAGGER IMPORTS
	httpSwagger "github.com/swaggo/http-swagger"
	_ "forum/docs" // <--- Ensure this points to your generated docs package!
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
	commentHandler := handlers.NewCommentHandler(commentRepo, postRepo, notificationRepo, hub)
	reactionHandler := handlers.NewReactionHandler(reactionRepo, postRepo, commentRepo, notificationRepo, hub)
	imageHandler := handlers.NewImageHandler(imageRepo, postRepo)
	guestHandler := handlers.NewGuestHandler(categoryRepo, postRepo, commentRepo, reactionRepo, imageRepo)
	notificationHandler := handlers.NewNotificationHandler(notificationRepo, hub)
	messageHandler := handlers.NewMessageHandler(messageRepo, hub) // ✅ Pass hub to handler
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

	// Public routes
	apiMux.Handle("/forum/api/categories", corsMiddleware.Handler(http.HandlerFunc(categoryHandler.GetCategories)))
	apiMux.Handle("/forum/api/category", corsMiddleware.Handler(http.HandlerFunc(categoryHandler.GetCategoryByID)))
	apiMux.Handle("/forum/api/feed", corsMiddleware.Handler(http.HandlerFunc(guestHandler.GetGuestData)))

	// Authentication routes (guest only)
	guestOnly := func(h http.Handler) http.Handler {
		return corsMiddleware.Handler(authMiddleware.RequireGuest(h))
	}

	apiMux.Handle("/forum/api/register", guestOnly(http.HandlerFunc(registerLimiter.Limit(authHandler.Register))))
	apiMux.Handle("/forum/api/session/login", guestOnly(http.HandlerFunc(authHandler.Login)))

	// OAuth routes (guest only)
	apiMux.Handle("/auth/google/login", guestOnly(http.HandlerFunc(oauthHandler.GoogleLogin)))
	// OAuth callbacks do not need RequireGuest or CSRF, as the 'state' parameter handles CSRF
	apiMux.Handle("/auth/google/callback", corsMiddleware.Handler(http.HandlerFunc(oauthHandler.GoogleCallback)))
	apiMux.Handle("/auth/github/login", guestOnly(http.HandlerFunc(oauthHandler.GitHubLogin)))
	apiMux.Handle("/oauth/github/callback", corsMiddleware.Handler(http.HandlerFunc(oauthHandler.GitHubCallback)))

	// Session management routes
	apiMux.Handle("/forum/api/session/logout", corsMiddleware.Handler(http.HandlerFunc(authHandler.Logout)))
	apiMux.Handle("/forum/api/session/verify", corsMiddleware.Handler(http.HandlerFunc(authHandler.VerifySession)))

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
	apiMux.Handle("/forum/api/posts/create", protected(http.HandlerFunc(postHandler.CreatePost)))
	apiMux.Handle("/forum/api/posts/delete/", protected(http.HandlerFunc(postHandler.DeletePost)))
	apiMux.Handle("/forum/api/posts/edit-title/", protected(http.HandlerFunc(postHandler.EditPostTitle)))
	apiMux.Handle("/forum/api/posts/edit-content/", protected(http.HandlerFunc(postHandler.EditPostContent)))
	apiMux.Handle("/forum/api/user/posts", protected(http.HandlerFunc(myPostsHandler.GetMyPosts)))
	apiMux.Handle("/forum/api/user/liked", protected(http.HandlerFunc(likedPostsHandler.GetLikedPosts)))
	apiMux.Handle("/forum/api/user/disliked", protected(http.HandlerFunc(likedPostsHandler.GetDislikedPosts)))
	apiMux.Handle("/forum/api/comments/create", protected(http.HandlerFunc(commentHandler.CreateComment)))
	apiMux.Handle("/forum/api/comments/edit/", protected(http.HandlerFunc(commentHandler.EditComment)))
	apiMux.Handle("/forum/api/comments/delete/", protected(http.HandlerFunc(commentHandler.DeleteComment)))
	apiMux.Handle("/forum/api/react", protected(http.HandlerFunc(reactionHandler.CreateReact)))
	apiMux.Handle("/forum/api/images/upload", protected(http.HandlerFunc(imageHandler.Upload)))
	apiMux.Handle("/forum/api/user/commented", protected(http.HandlerFunc(myPostsHandler.GetCommentedPosts)))
	apiMux.Handle("/forum/api/images/delete/", protected(http.HandlerFunc(imageHandler.DeleteImagesByPost)))

	// Additional protected routes for user management
	apiMux.Handle("/forum/api/user/profile", protected(http.HandlerFunc(authHandler.GetProfile)))
	apiMux.Handle("/forum/api/notifications", protected(http.HandlerFunc(notificationHandler.GetNotifications)))
	apiMux.Handle("/forum/api/notifications/delete/", protected(http.HandlerFunc(notificationHandler.HideNotification)))
	apiMux.Handle("/forum/api/session/logout-all", protected(http.HandlerFunc(authHandler.LogoutAll)))

	// Protected message routes
	apiMux.Handle("/forum/api/messages/send", protected(http.HandlerFunc(messageHandler.SendMessage)))
	apiMux.Handle("/forum/api/messages/conversation", protected(http.HandlerFunc(messageHandler.GetConversation)))
	apiMux.Handle("/forum/api/messages/conversations", protected(http.HandlerFunc(messageHandler.GetConversations)))
	apiMux.Handle("/forum/api/messages/users", protected(http.HandlerFunc(messageHandler.GetAllUsers)))
	apiMux.Handle("/forum/api/messages/users-for-chat", protected(http.HandlerFunc(messageHandler.GetUsersForChat)))
	apiMux.Handle("/forum/api/messages/unread-count", protected(http.HandlerFunc(messageHandler.GetUnreadCount)))
	apiMux.Handle("/forum/api/messages/mark-read/", protected(http.HandlerFunc(messageHandler.MarkAsRead)))
	apiMux.Handle("/forum/api/messages/delete/", protected(http.HandlerFunc(messageHandler.DeleteMessage)))

	// Protected chat image routes
	apiMux.Handle("/forum/api/chat/images/upload", protected(http.HandlerFunc(chatImageHandler.UploadChatImage)))
	apiMux.Handle("/forum/api/chat/images/message", protected(http.HandlerFunc(chatImageHandler.GetMessageImages)))
	apiMux.Handle("/forum/api/chat/images/gallery", protected(http.HandlerFunc(chatImageHandler.GetConversationGallery)))
	apiMux.Handle("/forum/api/chat/images/serve/", protected(http.HandlerFunc(chatImageHandler.ServeChatImage)))
	apiMux.Handle("/forum/api/chat/images/delete/", protected(http.HandlerFunc(chatImageHandler.DeleteChatImage)))
	apiMux.Handle("/forum/api/chat/images/stats", protected(http.HandlerFunc(chatImageHandler.GetUserImageStats)))
	
    
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