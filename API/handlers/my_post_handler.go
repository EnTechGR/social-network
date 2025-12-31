package handlers

import (
	"net/http"
	"time"

	"social-network/middleware"
	"social-network/repository"
	"social-network/utils"
)

type MyPostsHandler struct {
	PostRepo     *repository.PostRepository
	CommentRepo  *repository.CommentRepository
	ReactionRepo *repository.ReactionRepository
	ImageRepo    *repository.ImageRepository
}

func NewMyPostsHandler(postRepo *repository.PostRepository, commentRepo *repository.CommentRepository, reactionRepo *repository.ReactionRepository, imageRepo *repository.ImageRepository) *MyPostsHandler {
	return &MyPostsHandler{PostRepo: postRepo, CommentRepo: commentRepo, ReactionRepo: reactionRepo, ImageRepo: imageRepo}
}

// CategoryInfo represents a simplified category object
// swagger:model CategoryInfo
type CategoryInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// MyPostResponse represents a post with full context for the user's engagement view
// swagger:model MyPostResponse
type MyPostResponse struct {
	ID           string             `json:"id"`
	UserID       string             `json:"user_id"`
	Nickname     string             `json:"nickname"`
	Categories   []CategoryInfo     `json:"categories"`
	Title        string             `json:"title"`
	Content      string             `json:"content"`
	ImageURL     string             `json:"image_url,omitempty"`
	ThumbnailURL string             `json:"thumbnail_url,omitempty"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    *time.Time         `json:"updated_at,omitempty"`
	Comments     []CommentResponse  `json:"comments,omitempty"`
	Reactions    []ReactionResponse `json:"reactions,omitempty"`
}

// GetMyPosts retrieves all posts created by the authenticated user
// @Summary      Get own posts
// @Description  Retrieves a full list of posts created by the logged-in user, including nested categories, comments, and reactions.
// @Tags         User Activity
// @Security     CookieAuth
// @Produce      json
// @Success      200
// @Failure      401      {object}  models.ErrorResponse "Unauthorized"
// @Failure      500      {object}  models.ErrorResponse "Failed to load posts or associated data"
// @Router       /api/my-posts [get]
func (h *MyPostsHandler) GetMyPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	posts, err := h.PostRepo.GetPostsByUser(user.ID)
	if err != nil {
		utils.ErrorResponse(w, "Failed to load posts", http.StatusInternalServerError)
		return
	}

	var response []MyPostResponse
	for _, post := range posts {
		categories, err := h.PostRepo.GetCategoriesByPostID(post.ID)
		if err != nil {
			utils.ErrorResponse(w, "Failed to load categories", http.StatusInternalServerError)
			return
		}
		var catInfo []CategoryInfo
		for _, c := range categories {
			catInfo = append(catInfo, CategoryInfo{ID: c.ID, Name: c.Name})
		}

		comments, err := h.CommentRepo.GetCommentsByPostWithUser(post.ID)
		if err != nil {
			utils.ErrorResponse(w, "Failed to load comments", http.StatusInternalServerError)
			return
		}
		var commentResp []CommentResponse
		for _, c := range comments {
			cr := CommentResponse{
				ID:        c.ID,
				UserID:    c.UserID,
				Nickname:  c.Nickname,
				Content:   utils.DerefString(c.Content),
				CreatedAt: c.CreatedAt,
				UpdatedAt: c.UpdatedAt,
				Reactions: []ReactionResponse{},
			}
			reactions, err := h.ReactionRepo.GetReactionsByCommentWithUser(c.ID)
			if err != nil {
				utils.ErrorResponse(w, "Failed to load reactions", http.StatusInternalServerError)
				return
			}
			for _, r := range reactions {
				cr.Reactions = append(cr.Reactions, ReactionResponse{
					UserID:       r.UserID,
					Nickname:     r.Nickname,
					ReactionType: r.ReactionType,
					CreatedAt:    r.CreatedAt,
				})
			}
			commentResp = append(commentResp, cr)
		}

		reactions, err := h.ReactionRepo.GetReactionsByPostWithUser(post.ID)
		if err != nil {
			utils.ErrorResponse(w, "Failed to load reactions", http.StatusInternalServerError)
			return
		}
		var reactResp []ReactionResponse
		for _, r := range reactions {
			reactResp = append(reactResp, ReactionResponse{
				UserID:       r.UserID,
				Nickname:     r.Nickname,
				ReactionType: r.ReactionType,
				CreatedAt:    r.CreatedAt,
			})
		}

		imgs, err := h.ImageRepo.GetByPostID(post.ID)
		if err != nil {
			utils.ErrorResponse(w, "Failed to load images", http.StatusInternalServerError)
			return
		}
		var imgURL, thumbURL string
		if len(imgs) > 0 {
			imgURL = apiStaticBase + imgs[0].FilePath
			thumbURL = apiStaticBase + imgs[0].ThumbnailPath
		}

		response = append(response, MyPostResponse{
			ID:           post.ID,
			UserID:       post.UserID,
			Nickname:     post.Nickname,
			Categories:   catInfo,
			Title:        utils.DerefString(post.Title),
			Content:      utils.DerefString(post.Content),
			ImageURL:     imgURL,
			ThumbnailURL: thumbURL,
			CreatedAt:    post.CreatedAt,
			UpdatedAt:    post.UpdatedAt,
			Comments:     commentResp,
			Reactions:    reactResp,
		})
	}

	utils.JSONResponse(w, response, http.StatusOK)
}

// GetCommentedPosts retrieves posts where the authenticated user has left a comment
// @Summary      Get posts commented on
// @Description  Retrieves a list of posts that the current user has interacted with via comments. Useful for activity history.
// @Tags         User Activity
// @Security     CookieAuth
// @Produce      json
// @Success      200
// @Failure      401      {object}  models.ErrorResponse "Unauthorized"
// @Failure      500      {object}  models.ErrorResponse "Internal Server Error"
// @Router       /api/my-commented-posts [get]
func (h *MyPostsHandler) GetCommentedPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	posts, err := h.PostRepo.GetPostsCommentedByUser(user.ID)
	if err != nil {
		utils.ErrorResponse(w, "Failed to load posts", http.StatusInternalServerError)
		return
	}

	var response []MyPostResponse
	for _, post := range posts {
		categories, err := h.PostRepo.GetCategoriesByPostID(post.ID)
		if err != nil {
			utils.ErrorResponse(w, "Failed to load categories", http.StatusInternalServerError)
			return
		}
		var catInfo []CategoryInfo
		for _, c := range categories {
			catInfo = append(catInfo, CategoryInfo{ID: c.ID, Name: c.Name})
		}

		comments, err := h.CommentRepo.GetCommentsByPostWithUser(post.ID)
		if err != nil {
			utils.ErrorResponse(w, "Failed to load comments", http.StatusInternalServerError)
			return
		}
		var commentResp []CommentResponse
		for _, c := range comments {
			cr := CommentResponse{
				ID:        c.ID,
				UserID:    c.UserID,
				Nickname:  c.Nickname,
				Content:   utils.DerefString(c.Content),
				CreatedAt: c.CreatedAt,
				UpdatedAt: c.UpdatedAt,
				Reactions: []ReactionResponse{},
			}
			reactions, err := h.ReactionRepo.GetReactionsByCommentWithUser(c.ID)
			if err != nil {
				utils.ErrorResponse(w, "Failed to load reactions", http.StatusInternalServerError)
				return
			}
			for _, r := range reactions {
				cr.Reactions = append(cr.Reactions, ReactionResponse{
					UserID:       r.UserID,
					Nickname:     r.Nickname,
					ReactionType: r.ReactionType,
					CreatedAt:    r.CreatedAt,
				})
			}
			commentResp = append(commentResp, cr)
		}

		reactions, err := h.ReactionRepo.GetReactionsByPostWithUser(post.ID)
		if err != nil {
			utils.ErrorResponse(w, "Failed to load reactions", http.StatusInternalServerError)
			return
		}
		var reactResp []ReactionResponse
		for _, r := range reactions {
			reactResp = append(reactResp, ReactionResponse{
				UserID:       r.UserID,
				Nickname:     r.Nickname,
				ReactionType: r.ReactionType,
				CreatedAt:    r.CreatedAt,
			})
		}

		imgs, err := h.ImageRepo.GetByPostID(post.ID)
		if err != nil {
			utils.ErrorResponse(w, "Failed to load images", http.StatusInternalServerError)
			return
		}
		var imgURL, thumbURL string
		if len(imgs) > 0 {
			imgURL = apiStaticBase + imgs[0].FilePath
			thumbURL = apiStaticBase + imgs[0].ThumbnailPath
		}

		response = append(response, MyPostResponse{
			ID:           post.ID,
			UserID:       post.UserID,
			Nickname:     post.Nickname,
			Categories:   catInfo,
			Title:        utils.DerefString(post.Title),
			Content:      utils.DerefString(post.Content),
			ImageURL:     imgURL,
			ThumbnailURL: thumbURL,
			CreatedAt:    post.CreatedAt,
			UpdatedAt:    post.UpdatedAt,
			Comments:     commentResp,
			Reactions:    reactResp,
		})
	}

	utils.JSONResponse(w, response, http.StatusOK)
}
