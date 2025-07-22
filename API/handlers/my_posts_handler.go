package handlers

import (
	"net/http"
	"time"

	"forum/middleware"
	"forum/repository"
	"forum/utils"
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

type CategoryInfo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type MyPostResponse struct {
	ID           string             `json:"id"`
	UserID       string             `json:"user_id"`
	Username     string             `json:"username"`
	Categories   []CategoryInfo     `json:"categories"`
	Title        string             `json:"title"`
	Content      string             `json:"content"`
	ImageURL     string             `json:"image_url,omitempty"`
	ThumbnailURL string             `json:"thumbnail_url,omitempty"`
	CreatedAt    time.Time          `json:"created_at"`
	Comments     []CommentResponse  `json:"comments,omitempty"`
	Reactions    []ReactionResponse `json:"reactions,omitempty"`
}

func (h *MyPostsHandler) GetMyPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
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
				Username:  c.Username,
				Content:   c.Content,
				CreatedAt: c.CreatedAt,
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
					Username:     r.Username,
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
				Username:     r.Username,
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
			Username:     post.Username,
			Categories:   catInfo,
			Title:        post.Title,
			Content:      post.Content,
			ImageURL:     imgURL,
			ThumbnailURL: thumbURL,
			CreatedAt:    post.CreatedAt,
			Comments:     commentResp,
			Reactions:    reactResp,
		})
	}

	utils.JSONResponse(w, response, http.StatusOK)
}
