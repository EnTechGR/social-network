package handlers

import (
	"net/http"
	"strconv"

	"forum/models"
	"forum/repository"
	"forum/utils"
)

// CategoryHandler handles category related requests

type CategoryHandler struct {
	CategoryRepo *repository.CategoryRepository
	PostRepo     *repository.PostRepository
	ImageRepo    *repository.ImageRepository
}

// NewCategoryHandler creates a new CategoryHandler
func NewCategoryHandler(catRepo *repository.CategoryRepository, postRepo *repository.PostRepository, imageRepo *repository.ImageRepository) *CategoryHandler {
	return &CategoryHandler{
		CategoryRepo: catRepo,
		PostRepo:     postRepo,
		ImageRepo:    imageRepo,
	}
}

// GetCategories returns all categories as JSON
func (h *CategoryHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	categories, err := h.CategoryRepo.GetAll()
	if err != nil {
		utils.ErrorResponse(w, "Failed to load categories", http.StatusInternalServerError)
		return
	}

	utils.JSONResponse(w, categories, http.StatusOK)
}

func (h *CategoryHandler) GetCategoryByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		utils.ErrorResponse(w, "Missing category ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.ErrorResponse(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	category, err := h.CategoryRepo.GetCategoryByID(id)
	if err != nil {
		utils.ErrorResponse(w, "Failed to load category", http.StatusInternalServerError)
		return
	}

	if category == nil {
		utils.ErrorResponse(w, "Category not found", http.StatusNotFound)
		return
	}

	posts, err := h.PostRepo.GetPostsByCategoryWithUser(id)
	if err != nil {
		utils.ErrorResponse(w, "Failed to fetch posts", http.StatusInternalServerError)
		return
	}

	for i := range posts {
		imgs, err := h.ImageRepo.GetByPostID(posts[i].ID)
		if err != nil {
			utils.ErrorResponse(w, "Failed to load images", http.StatusInternalServerError)
			return
		}
		if len(imgs) > 0 {
			posts[i].ImageURL = apiStaticBase + imgs[0].FilePath
			posts[i].ThumbnailURL = apiStaticBase + imgs[0].ThumbnailPath
		}
	}

	categoryByID := models.CategoryWithPosts{
		ID:    category.ID,
		Name:  category.Name,
		Posts: posts,
	}

	utils.JSONResponse(w, categoryByID, http.StatusOK)
}
