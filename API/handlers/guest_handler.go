package handlers

import (
	"forum/repository"
	"forum/utils"
	"net/http"
	"time"
)

// apiStaticBase is the base URL for serving static files like uploaded images.
const apiStaticBase = "http://localhost:8080/static/"

type GuestHandler struct {
	categoryRepo *repository.CategoryRepository
	postRepo     *repository.PostRepository
	commentRepo  *repository.CommentRepository
	reactionRepo *repository.ReactionRepository
	imageRepo    *repository.ImageRepository
}

type ReactionResponse struct {
	UserID       string    `json:"user_id"`
	Username     string    `json:"username"`
	ReactionType int       `json:"reaction_type"`
	CreatedAt    time.Time `json:"created_at"`
}

type CommentResponse struct {
	ID        string             `json:"id"`
	UserID    string             `json:"user_id"`
	Username  string             `json:"username"`
	Content   string             `json:"content"`
	CreatedAt time.Time          `json:"created_at"`
	Reactions []ReactionResponse `json:"reactions,omitempty"`
}

type PostResponse struct {
	ID           string             `json:"id"`
	UserID       string             `json:"user_id"`
	Username     string             `json:"username"`
	CategoryID   int                `json:"category_id"`
	CategoryName string             `json:"category_name"` // NEW FIELD
	Title        string             `json:"title"`         // Optional title field
	Content      string             `json:"content"`
	CreatedAt    time.Time          `json:"created_at"`
	ImageURL     string             `json:"image_url,omitempty"`
	ThumbnailURL string             `json:"thumbnail_url,omitempty"`
	Comments     []CommentResponse  `json:"comments,omitempty"`
	Reactions    []ReactionResponse `json:"reactions,omitempty"`
}

type CategoryResponse struct {
	ID    int            `json:"id"`
	Name  string         `json:"name"`
	Posts []PostResponse `json:"posts"`
}

type GuestResponse struct {
	Categories []CategoryResponse `json:"categories"`
}

func NewGuestHandler(
	categoryRepo *repository.CategoryRepository,
	postRepo *repository.PostRepository,
	commentRepo *repository.CommentRepository,
	reactionRepo *repository.ReactionRepository,
	imageRepo *repository.ImageRepository,
) *GuestHandler {
	return &GuestHandler{
		categoryRepo: categoryRepo,
		postRepo:     postRepo,
		commentRepo:  commentRepo,
		reactionRepo: reactionRepo,
		imageRepo:    imageRepo,
	}
}

func (h *GuestHandler) GuestView(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		//http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		utils.ErrorResponse(w, "Only POST requests are allowed for registration.", http.StatusMethodNotAllowed)
		return
	}

	posts, err := h.postRepo.GetAllPosts()
	if err != nil {
		utils.ErrorResponse(w, "Failed to fetch posts.", http.StatusInternalServerError)
		return
	}

	comments, err := h.commentRepo.GetAllComments()
	if err != nil {
		utils.ErrorResponse(w, "Failed to fetch comments.", http.StatusInternalServerError)
		return
	}

	reactions, err := h.reactionRepo.GetAllReactions()
	if err != nil {
		utils.ErrorResponse(w, "Failed to fetch reactions.", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"posts":     posts,
		"comments":  comments,
		"reactions": reactions,
	}

	w.Header().Set("Content-Type", "application/json")
	utils.JSONResponse(w, response, http.StatusOK)
}

func (h *GuestHandler) GetGuestData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	categories, err := h.categoryRepo.GetAll()
	if err != nil {
		utils.ErrorResponse(w, "Failed to load categories", http.StatusInternalServerError)
		return
	}

	var response GuestResponse
	for _, cat := range categories {
		catResp := CategoryResponse{
			ID:    cat.ID,
			Name:  cat.Name,
			Posts: []PostResponse{}, // ✅ always initialized to avoid null
		}

		posts, err := h.postRepo.GetPostsByCategoryWithUser(cat.ID)

		if err != nil {
			utils.ErrorResponse(w, "Failed to load posts", http.StatusInternalServerError)
			return
		}

		for _, post := range posts {
			postResp := PostResponse{
				ID:           post.ID,
				UserID:       post.UserID,
				Username:     post.Username,
				CategoryID:   post.CategoryID,
				CategoryName: cat.Name,   // ✅ inject category name
				Title:        post.Title, // Optional title field
				Content:      post.Content,
				CreatedAt:    post.CreatedAt,
				Comments:     []CommentResponse{},  // ✅ avoid null
				Reactions:    []ReactionResponse{}, // ✅ avoid null
			}
			imgs, err := h.imageRepo.GetByPostID(post.ID)
			if err != nil {
				utils.ErrorResponse(w, "Failed to load images", http.StatusInternalServerError)
				return
			}
			if len(imgs) > 0 {
				postResp.ImageURL = apiStaticBase + imgs[0].FilePath
				postResp.ThumbnailURL = apiStaticBase + imgs[0].ThumbnailPath
			}

			comments, err := h.commentRepo.GetCommentsByPostWithUser(post.ID)
			if err != nil {
				utils.ErrorResponse(w, "Failed to load comments", http.StatusInternalServerError)
				return
			}

			for _, comment := range comments {
				commentResp := CommentResponse{
					ID:        comment.ID,
					UserID:    comment.UserID,
					Username:  comment.Username,
					Content:   comment.Content,
					CreatedAt: comment.CreatedAt,
					Reactions: []ReactionResponse{}, // ✅ avoid null
				}

				reactions, err := h.reactionRepo.GetReactionsByCommentWithUser(comment.ID)
				if err != nil {
					utils.ErrorResponse(w, "Failed to load reactions", http.StatusInternalServerError)
					return
				}
				for _, reaction := range reactions {
					commentResp.Reactions = append(commentResp.Reactions, ReactionResponse{
						UserID:       reaction.UserID,
						Username:     reaction.Username,
						ReactionType: reaction.ReactionType,
						CreatedAt:    reaction.CreatedAt,
					})
				}

				postResp.Comments = append(postResp.Comments, commentResp)
			}

			reactions, err := h.reactionRepo.GetReactionsByPostWithUser(post.ID)
			if err != nil {
				utils.ErrorResponse(w, "Failed to load reactions", http.StatusInternalServerError)
				return
			}
			for _, reaction := range reactions {
				postResp.Reactions = append(postResp.Reactions, ReactionResponse{
					UserID:       reaction.UserID,
					Username:     reaction.Username,
					ReactionType: reaction.ReactionType,
					CreatedAt:    reaction.CreatedAt,
				})
			}

			catResp.Posts = append(catResp.Posts, postResp)
		}

		response.Categories = append(response.Categories, catResp)
	}

	utils.JSONResponse(w, response, http.StatusOK)
}
