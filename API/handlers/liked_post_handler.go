package handlers

import (
	"net/http"
	"social-network/middleware"
	"social-network/repository"
	"social-network/utils"
)

type LikedPostsHandler struct {
	PostRepo     *repository.PostRepository
	CommentRepo  *repository.CommentRepository
	ReactionRepo *repository.ReactionRepository
	ImageRepo    *repository.ImageRepository
}

func NewLikedPostsHandler(postRepo *repository.PostRepository, commentRepo *repository.CommentRepository, reactionRepo *repository.ReactionRepository, imageRepo *repository.ImageRepository) *LikedPostsHandler {
	return &LikedPostsHandler{PostRepo: postRepo, CommentRepo: commentRepo, ReactionRepo: reactionRepo, ImageRepo: imageRepo}
}

// GetLikedPosts retrieves all posts liked by the current user
// @Summary      Get liked posts
// @Description  Returns a list of all posts the authenticated user has reacted to positively (liked).
// @Tags         User Engagement
// @Security     CookieAuth
// @Produce      json
// @Success      200  
// @Failure      401  {object}  models.ErrorResponse "Unauthorized"
// @Failure      500  {object}  models.ErrorResponse "Internal server error"
// @Router       /api/posts/liked [get]
func (h *LikedPostsHandler) GetLikedPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	posts, err := h.PostRepo.GetPostsReactedByUser(user.ID)
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
			Comments:     commentResp,
			Reactions:    reactResp,
		})
	}

	utils.JSONResponse(w, response, http.StatusOK)
}

// GetDislikedPosts retrieves all posts disliked by the current user
// @Summary      Get disliked posts
// @Description  Returns a list of all posts the authenticated user has reacted to negatively (disliked).
// @Tags         User Engagement
// @Security     CookieAuth
// @Produce      json
// @Success      200  {array}   handlers.MyPostResponse
// @Failure      401  {object}  models.ErrorResponse "Unauthorized"
// @Failure      500  {object}  models.ErrorResponse "Internal server error"
// @Router       /api/posts/disliked [get]
func (h *LikedPostsHandler) GetDislikedPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user := middleware.GetCurrentUser(r)
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	posts, err := h.PostRepo.GetPostsDislikedByUser(user.ID)
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
			Comments:     commentResp,
			Reactions:    reactResp,
		})
	}

	utils.JSONResponse(w, response, http.StatusOK)
}
