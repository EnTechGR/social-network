package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"social-network/middleware"
	"social-network/models"
	"social-network/repository"
	"social-network/utils"
)

// ============================================================================
// Constants
// ============================================================================

const (
	// feedDefaultLimit is the number of posts returned when no limit is given.
	feedDefaultLimit = 20
	// feedMaxLimit caps the limit query parameter to prevent abuse.
	feedMaxLimit = 50
	// feedMinLimit is the minimum accepted value; anything below is clamped up.
	feedMinLimit = 1
	// feedCursorLayout is the time format used to encode / decode cursors.
	// Using RFC3339Nano preserves full sub-second precision for cursors that
	// sit very close together.
	feedCursorLayout = time.RFC3339Nano
)

// ============================================================================
// FeedHandler
// ============================================================================

// FeedHandler handles GET /api/v1/feed.
type FeedHandler struct {
	FeedRepo *repository.FeedRepository
}

// NewFeedHandler creates a new FeedHandler.
func NewFeedHandler(feedRepo *repository.FeedRepository) *FeedHandler {
	return &FeedHandler{FeedRepo: feedRepo}
}

// GetFeed returns a paginated, visibility-filtered list of posts for the
// authenticated user's feed.
//
// @Summary      Get feed
// @Description  Returns a cursor-paginated list of posts visible to the authenticated user.
//               Visibility rules: public posts from everyone; followers-only posts from
//               users the viewer follows; private posts the viewer has been explicitly
//               allowed to see; and always the viewer's own posts regardless of visibility.
// @Tags         Feed
// @Security     CookieAuth
// @Produce      json
// @Param        limit      query  integer  false  "Number of posts per page (1-50, default 20)"
// @Param        cursor     query  string   false  "RFC3339 timestamp cursor from a previous response; omit for the first page"
// @Param        direction  query  string   false  "Pagination direction: 'next' (default, older posts) or 'prev' (newer posts, for pull-to-refresh)"
// @Success      200  {object}  models.FeedResponse
// @Failure      400  {object}  models.ErrorResponse "Invalid limit or cursor"
// @Failure      401  {object}  models.ErrorResponse "Unauthorized"
// @Failure      500  {object}  models.ErrorResponse "Internal Server Error"
// @Router       /api/v1/feed [get]
func (h *FeedHandler) GetFeed(w http.ResponseWriter, r *http.Request) {
	// ── Method guard ─────────────────────────────────────────────────────────
	if r.Method != http.MethodGet {
		utils.ErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// ── Authentication ───────────────────────────────────────────────────────
	viewer := middleware.GetCurrentUser(r)
	if viewer == nil {
		utils.ErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// ── Parse & validate query parameters ────────────────────────────────────
	params, err := parseFeedParams(r, viewer.ID)
	if err != nil {
		utils.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	// ── Delegate to repository ───────────────────────────────────────────────
	posts, hasMore, err := h.FeedRepo.GetFeedPage(params)
	if err != nil {
		log.Printf("[FeedHandler] GetFeedPage error for viewer %s: %v", viewer.ID, err)
		utils.ErrorResponse(w, "Failed to load feed", http.StatusInternalServerError)
		return
	}

	// ── Build the paginated envelope ─────────────────────────────────────────
	resp := buildFeedResponse(posts, params, hasMore)
	utils.JSONResponse(w, resp, http.StatusOK)
}

// ============================================================================
// parseFeedParams
// ============================================================================

// parseFeedParams reads, validates, and normalises the query parameters for
// the feed endpoint.  It returns an error string suitable for sending directly
// to the client.
func parseFeedParams(r *http.Request, viewerID string) (models.FeedParams, error) {
	q := r.URL.Query()

	// ── limit ─────────────────────────────────────────────────────────────────
	limit := feedDefaultLimit
	if raw := q.Get("limit"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil {
			return models.FeedParams{}, fmt.Errorf("invalid limit: must be an integer")
		}
		limit = v
	}
	// Clamp to [feedMinLimit, feedMaxLimit].
	if limit < feedMinLimit {
		limit = feedMinLimit
	}
	if limit > feedMaxLimit {
		limit = feedMaxLimit
	}

	// ── cursor ────────────────────────────────────────────────────────────────
	var cursor time.Time // zero value = "no cursor, start from the top"
	if raw := q.Get("cursor"); raw != "" {
		t, err := time.Parse(feedCursorLayout, raw)
		if err != nil {
			// Try RFC3339 without nanoseconds as a fallback for clients that
			// strip the fractional seconds.
			t, err = time.Parse(time.RFC3339, raw)
			if err != nil {
				return models.FeedParams{}, fmt.Errorf("invalid cursor: expected RFC3339 timestamp")
			}
		}
		cursor = t
	}

	// ── direction ─────────────────────────────────────────────────────────────
	direction := "next"
	if raw := q.Get("direction"); raw == "prev" {
		direction = "prev"
	}

	return models.FeedParams{
		ViewerID:  viewerID,
		Cursor:    cursor,
		Limit:     limit,
		Direction: direction,
	}, nil
}

// ============================================================================
// buildFeedResponse
// ============================================================================

// buildFeedResponse assembles the FeedResponse envelope from the repository
// results.  It derives the next/prev cursors from the first and last posts in
// the slice so that the caller can navigate both directions.
func buildFeedResponse(posts []models.FeedPost, params models.FeedParams, hasMore bool) models.FeedResponse {
	resp := models.FeedResponse{
		Posts:   posts,
		HasMore: hasMore,
		Count:   len(posts),
	}

	if len(posts) == 0 {
		return resp
	}

	// posts is always newest-first, regardless of the requested direction.
	newest := posts[0].CreatedAt
	oldest := posts[len(posts)-1].CreatedAt

	// NextCursor allows the client to fetch older posts (next page forward in
	// time).  It points just before the oldest post on this page.
	if hasMore {
		resp.NextCursor = oldest.Format(feedCursorLayout)
	}

	// PrevCursor allows the client to refresh toward newer posts.  It is only
	// meaningful when the client already has a cursor (i.e., is not on the
	// very first page) — but we always provide it so the frontend can
	// poll for fresh content without extra logic.
	resp.PrevCursor = newest.Format(feedCursorLayout)

	return resp
}