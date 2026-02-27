package models

import "time"

// ============================================================================
// Feed Models
// These types are the canonical shapes returned by GET /api/v1/feed.
// They are deliberately separate from PostWithUser so that the feed
// can evolve its shape without breaking other endpoints.
// ============================================================================

// FeedImage represents a single image attached to a feed post,
// ordered by display_order ascending.
//
// swagger:model FeedImage
type FeedImage struct {
	// Internal image identifier
	// example: 3fa85f64-5717-4562-b3fc-2c963f66afa6
	ImageID string `json:"image_id"`
	// Fully-qualified URL to the original image
	// example: http://localhost:8080/static/uploads/3fa85f64.jpg
	URL string `json:"url"`
	// Fully-qualified URL to the 200×200 thumbnail
	// example: http://localhost:8080/static/uploads/3fa85f64_thumb.jpg
	ThumbnailURL string `json:"thumbnail_url"`
	// Fully-qualified URL to a smaller responsive thumbnail (when available)
	// example: http://localhost:8080/static/uploads/3fa85f64_thumb_sm.jpg
	SmallThumbnailURL string `json:"small_thumbnail_url,omitempty"`
	// Fully-qualified URL to a medium responsive thumbnail (when available)
	// example: http://localhost:8080/static/uploads/3fa85f64_thumb_md.jpg
	MediumThumbnailURL string `json:"medium_thumbnail_url,omitempty"`
	// 1-based position for ordered display
	// example: 1
	DisplayOrder int `json:"display_order"`
}

// FeedPost is the fully-enriched post object returned inside FeedResponse.
// It carries denormalised author data, pre-aggregated engagement counts, and
// the viewer's own reaction state so the frontend needs no additional calls.
//
// swagger:model FeedPost
type FeedPost struct {
	// ── Identity ─────────────────────────────────────────────────────────────
	// example: 7b1a2c3d-4e5f-6a7b-8c9d-0e1f2a3b4c5d
	ID string `json:"id"`
	// example: 2025-02-17T10:30:00Z
	CreatedAt time.Time `json:"created_at"`
	// example: 2025-02-17T11:00:00Z
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	// Visibility at the time the feed was fetched: public | followers | private
	// example: public
	Visibility string `json:"visibility"`

	// ── Author ───────────────────────────────────────────────────────────────
	// example: a1b2c3d4-e5f6-7890-abcd-ef1234567890
	AuthorID string `json:"author_id"`
	// example: johndoe
	AuthorNickname string `json:"author_nickname"`
	// example: John
	AuthorFirstName string `json:"author_first_name"`
	// example: Doe
	AuthorLastName string `json:"author_last_name"`
	// URL to the author's full-size avatar; empty string when no avatar is set
	// example: http://localhost:8080/static/uploads/avatar_abc.jpg
	AuthorAvatarURL string `json:"author_avatar_url"`
	// URL to the author's thumbnail avatar; empty string when no avatar is set
	// example: http://localhost:8080/static/uploads/avatar_abc_thumb.jpg
	AuthorAvatarThumbURL string `json:"author_avatar_thumb_url"`

	// ── Content ──────────────────────────────────────────────────────────────
	// example: How to containerise a Go application
	Title string `json:"title"`
	// example: I've been trying to use Docker with Go and ran into a few issues…
	Content string `json:"content"`
	// Ordered list of images; always present (empty slice, never null)
	Images []FeedImage `json:"images"`

	// ── Engagement counts ────────────────────────────────────────────────────
	// Number of likes (reaction_type = 1)
	// example: 42
	LikeCount int `json:"like_count"`
	// Number of dislikes (reaction_type = 2)
	// example: 3
	DislikeCount int `json:"dislike_count"`
	// Total non-deleted comments
	// example: 17
	CommentCount int `json:"comment_count"`

	// ── Viewer state ─────────────────────────────────────────────────────────
	// The authenticated viewer's reaction type on this post.
	// null  = no reaction
	// 1     = like
	// 2     = dislike
	// 3     = love (or any other type you add later)
	// example: 1
	ViewerReaction *int `json:"viewer_reaction"`
}

// FeedResponse is the top-level envelope returned by GET /api/v1/feed.
//
// Pagination is cursor-based on created_at so new posts arriving between
// page fetches never cause duplicates or gaps.
//
// swagger:model FeedResponse
type FeedResponse struct {
	// The posts for this page, newest-first.
	Posts []FeedPost `json:"posts"`
	// Cursor to pass as ?cursor=… to fetch the NEXT (older) page.
	// Empty string when this is the last page.
	// example: 2025-01-14T08:00:00Z
	NextCursor string `json:"next_cursor,omitempty"`
	// Cursor to pass as ?cursor=…&direction=prev to refresh toward newer posts.
	// Empty string on the very first page.
	// example: 2025-02-17T10:30:00Z
	PrevCursor string `json:"prev_cursor,omitempty"`
	// True when there is at least one more page in the requested direction.
	HasMore bool `json:"has_more"`
	// Number of posts in this page (convenience field — equals len(Posts)).
	// example: 20
	Count int `json:"count"`
}

// FeedComment is the enriched comment shape returned by GET /api/v1/posts/{id}/comments.
// It mirrors the design of FeedPost: denormalised author info, pre-aggregated
// reaction counts, and the viewer's own reaction state so the frontend needs
// no secondary calls.
//
// swagger:model FeedComment
type FeedComment struct {
	// ── Identity ─────────────────────────────────────────────────────────────
	// example: 550e8400-e29b-41d4-a716-446655440000
	ID string `json:"id"`
	// Post this comment belongs to
	// example: 7b1a2c3d-4e5f-6a7b-8c9d-0e1f2a3b4c5d
	PostID string `json:"post_id"`
	// example: 2025-02-17T10:30:00Z
	CreatedAt time.Time `json:"created_at"`
	// example: 2025-02-17T11:00:00Z
	UpdatedAt *time.Time `json:"updated_at,omitempty"`

	// ── Author ───────────────────────────────────────────────────────────────
	// example: a1b2c3d4-e5f6-7890-abcd-ef1234567890
	AuthorID string `json:"author_id"`
	// example: johndoe
	AuthorNickname string `json:"author_nickname"`
	// example: John
	AuthorFirstName string `json:"author_first_name"`
	// example: Doe
	AuthorLastName string `json:"author_last_name"`
	// URL to the author's full-size avatar; empty string when no avatar is set
	AuthorAvatarURL string `json:"author_avatar_url"`
	// URL to the author's thumbnail avatar; empty string when no avatar is set
	AuthorAvatarThumbURL string `json:"author_avatar_thumb_url"`

	// ── Content ──────────────────────────────────────────────────────────────
	// Text body of the comment; empty string when the comment has been deleted
	// example: Great post, really helpful!
	Content string `json:"content"`
	// Fully-qualified URL of the first attached image (by display_order), if any
	// example: http://localhost:8080/static/3fa85f64.jpg
	ImageURL string `json:"image_url,omitempty"`
	// Fully-qualified URL of the first attached thumbnail, if any
	// example: http://localhost:8080/static/3fa85f64_thumb.jpg
	ImageThumbnailURL string `json:"image_thumbnail_url,omitempty"`

	// ── Engagement counts ────────────────────────────────────────────────────
	// example: 5
	LikeCount int `json:"like_count"`
	// example: 0
	DislikeCount int `json:"dislike_count"`

	// ── Viewer state ─────────────────────────────────────────────────────────
	// null = no reaction, 1 = like, 2 = dislike, 3 = love
	// example: 1
	ViewerReaction *int `json:"viewer_reaction"`
}

// FeedParams bundles all inputs the repository needs to execute a paginated
// feed query. Constructed by the handler after validating query parameters.
type FeedParams struct {
	// ID of the authenticated user making the request.
	ViewerID string
	// Upper-bound (next) or lower-bound (prev) created_at for pagination.
	// Zero value means "start from the very beginning" (newest post).
	Cursor time.Time
	// Maximum number of posts to return. Clamped to [1, 50] by the handler.
	Limit int
	// "next" (default) fetches older posts; "prev" fetches newer posts.
	Direction string
}
