package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"social-network/models"
)

// staticBase is the URL prefix used to turn a stored file path into a
// publicly accessible URL.  The static file server is mounted as:
//   http.FileServer(http.Dir("./uploads")) → Handle("/static/", StripPrefix("/static/", fs))
// so a request to /static/foo.jpg serves ./uploads/foo.jpg.
const staticBase = "http://localhost:8080/static/"

// toStaticURL converts a stored file_path from images_core (e.g. "uploads/abc.jpg")
// into a fully-qualified public URL (e.g. "http://localhost:8080/static/abc.jpg").
//
// The image repository stores paths as filepath.Join("uploads", filename), which
// includes the "uploads/" directory prefix.  The static handler already roots
// itself at ./uploads/, so we must strip that prefix before prepending staticBase
// to avoid the double-path "/static/uploads/abc.jpg" bug.
//
// Returns an empty string unchanged so callers do not need a nil-guard.
func toStaticURL(storedPath string) string {
	if storedPath == "" {
		return ""
	}
	return staticBase + strings.TrimPrefix(storedPath, "uploads/")
}

// ============================================================================
// FeedRepository
// ============================================================================

// FeedRepository executes all SQL required by the feed endpoint.
// It purposely knows nothing about HTTP — all inputs arrive via FeedParams
// and all outputs are []models.FeedPost.
type FeedRepository struct {
	db *sql.DB
}

// NewFeedRepository creates a new FeedRepository.
func NewFeedRepository(db *sql.DB) *FeedRepository {
	return &FeedRepository{db: db}
}

// ============================================================================
// GetFeedPage — the only public method
// ============================================================================

// GetFeedPage fetches one page of posts that params.ViewerID is allowed to see,
// ordered by created_at DESC (newest first).
//
// The implementation fires exactly TWO database queries per call:
//  1. The primary post query — JOINs user + avatar, embeds correlated
//     subquery aggregates (comment_count, like_count, dislike_count,
//     viewer_reaction), and applies the full visibility predicate in SQL
//     so that no post-fetch filtering is needed in Go.
//  2. A batch image query — fetches all images for the returned post IDs
//     in a single IN(...) query, then merges them into the FeedPost slice.
//
// Pagination is cursor-based on created_at.  To get the first page, pass a
// zero-value time.Time in params.Cursor.  For subsequent pages use the
// NextCursor / PrevCursor values from FeedResponse.
func (r *FeedRepository) GetFeedPage(params models.FeedParams) ([]models.FeedPost, bool, error) {
	// ── 1.  Build and run the primary query ──────────────────────────────────
	posts, hasMore, err := r.queryPosts(params)
	if err != nil {
		return nil, false, fmt.Errorf("feed: primary query: %w", err)
	}
	if len(posts) == 0 {
		return posts, false, nil
	}

	// ── 2.  Batch-fetch images for all returned post IDs ─────────────────────
	postIDs := make([]string, len(posts))
	for i := range posts {
		postIDs[i] = posts[i].ID
	}
	imagesByPost, err := r.batchFetchImages(postIDs)
	if err != nil {
		// Non-fatal: return posts without images rather than failing the whole
		// feed.  The caller can decide whether to surface a warning.
		imagesByPost = make(map[string][]models.FeedImage)
	}

	// ── 3.  Merge images into posts ───────────────────────────────────────────
	for i := range posts {
		if imgs, ok := imagesByPost[posts[i].ID]; ok {
			posts[i].Images = imgs
		}
		// Images is already initialised to []FeedImage{} in the scan helper,
		// so the frontend always receives an array, never null.
	}

	return posts, hasMore, nil
}

// ============================================================================
// GetPostByID — single-post detail
// ============================================================================

// GetPostByID fetches one fully-enriched post by its ID, applying the same
// visibility rules as the feed so a viewer cannot access a post they are not
// permitted to see.
//
// Returns (nil, nil) when the post does not exist or the viewer is not
// allowed to see it — the handler converts both cases to 404 so that
// private post existence is never leaked to unauthorised callers.
func (r *FeedRepository) GetPostByID(postID, viewerID string) (*models.FeedPost, error) {
	// The query is the feed's primary query with two differences:
	//   1.  The cursor/pagination clause is replaced by AND p.post_id = ?
	//   2.  No LIMIT — we expect exactly one row.
	// The visibility predicate is word-for-word identical to queryPosts so
	// access rules are guaranteed to be consistent.
	query := `
		SELECT
		    p.post_id,
		    p.user_id,
		    u.nickname,
		    u.first_name,
		    u.last_name,
		    COALESCE(ic_av.file_path,      '')  AS avatar_url,
		    COALESCE(ic_av.thumbnail_path, '')  AS avatar_thumb_url,
		    p.visibility,
		    COALESCE(p.title,   '')             AS title,
		    COALESCE(p.content, '')             AS content,
		    p.created_at,
		    p.updated_at,
		    (
		        SELECT COUNT(*)
		        FROM   comments c
		        WHERE  c.post_id    = p.post_id
		        AND    c.deleted_at IS NULL
		    ) AS comment_count,
		    (
		        SELECT COUNT(*)
		        FROM   reactions rk
		        WHERE  rk.post_id      = p.post_id
		        AND    rk.reaction_type = 1
		    ) AS like_count,
		    (
		        SELECT COUNT(*)
		        FROM   reactions rd
		        WHERE  rd.post_id      = p.post_id
		        AND    rd.reaction_type = 2
		    ) AS dislike_count,
		    (
		        SELECT rv.reaction_type
		        FROM   reactions rv
		        WHERE  rv.post_id = p.post_id
		        AND    rv.user_id = ?
		        LIMIT  1
		    ) AS viewer_reaction
		FROM  posts p
		JOIN  user u
		      ON  u.user_id = p.user_id
		LEFT  JOIN user_avatars ua
		      ON  ua.user_id = u.user_id
		LEFT  JOIN images_core ic_av
		      ON  ic_av.image_id   = ua.image_id
		      AND ic_av.deleted_at IS NULL
		WHERE p.post_id = ?
		AND   p.group_id IS NULL
		AND   (
		          p.user_id = ?
		      OR  p.visibility = 'public'
		      OR  (
		              p.visibility = 'followers'
		          AND EXISTS (
		                  SELECT 1
		                  FROM   follow_relationships fr
		                  WHERE  fr.follower_id = ?
		                  AND    fr.followee_id = p.user_id
		                  AND    fr.status      = 'accepted'
		              )
		          )
		      OR  (
		              p.visibility = 'private'
		          AND EXISTS (
		                  SELECT 1
		                  FROM   post_allowed_users pau
		                  WHERE  pau.post_id = p.post_id
		                  AND    pau.user_id = ?
		              )
		          )
		      )
	`

	// Bind order:
	//   1  viewer_reaction subquery  → viewerID
	//   2  post identity             → postID
	//   3  rule 1 (own posts)        → viewerID
	//   4  rule 3 (followers)        → viewerID
	//   5  rule 4 (private allowed)  → viewerID
	row := r.db.QueryRow(query, viewerID, postID, viewerID, viewerID, viewerID)

	post, err := scanFeedPostRow(row)
	if err == sql.ErrNoRows {
		return nil, nil // not found or not visible — caller returns 404
	}
	if err != nil {
		return nil, fmt.Errorf("feed: get post by id: %w", err)
	}

	// Fetch images for this single post using the same batch helper.
	imagesByPost, err := r.batchFetchImages([]string{postID})
	if err == nil {
		if imgs, ok := imagesByPost[postID]; ok {
			post.Images = imgs
		}
	}

	return post, nil
}

// ============================================================================
// queryPosts — primary query
// ============================================================================

// queryPosts builds the paginated, visibility-filtered SQL query and scans
// the results into a []models.FeedPost slice.  It fetches params.Limit + 1
// rows so we can determine hasMore without a separate COUNT query.
func (r *FeedRepository) queryPosts(params models.FeedParams) ([]models.FeedPost, bool, error) {
	// ── Cursor handling ───────────────────────────────────────────────────────
	// When no cursor is provided we want all posts up to now, so we use a
	// far-future sentinel that is always greater than any real created_at.
	cursorTime := params.Cursor
	if cursorTime.IsZero() {
		cursorTime = time.Now().Add(24 * time.Hour)
	}

	// For the "prev" direction we want rows *newer* than the cursor so that
	// pull-to-refresh works correctly.  We flip the operator and order, then
	// reverse the slice before returning so the caller always gets newest-first.
	var (
		cursorOp    string
		orderClause string
	)
	switch params.Direction {
	case "prev":
		cursorOp = ">"
		orderClause = "p.created_at ASC"
	default: // "next"
		cursorOp = "<"
		orderClause = "p.created_at DESC"
	}

	// ── SQL ───────────────────────────────────────────────────────────────────
	// Visibility predicate (all four rules expressed in pure SQL):
	//   Rule 1 — viewer always sees their own posts
	//   Rule 2 — public posts are open to everyone
	//   Rule 3 — followers-only posts: viewer must be an accepted follower
	//   Rule 4 — private posts: viewer must be in post_allowed_users
	//
	// Correlated subqueries for counts are evaluated once per row by SQLite's
	// query planner and are covered by the existing indexes on post_id.
	query := fmt.Sprintf(`
		SELECT
		    p.post_id,
		    p.user_id,
		    u.nickname,
		    u.first_name,
		    u.last_name,
		    COALESCE(ic_av.file_path,      '')  AS avatar_url,
		    COALESCE(ic_av.thumbnail_path, '')  AS avatar_thumb_url,
		    p.visibility,
		    COALESCE(p.title,   '')             AS title,
		    COALESCE(p.content, '')             AS content,
		    p.created_at,
		    p.updated_at,
		    (
		        SELECT COUNT(*)
		        FROM   comments c
		        WHERE  c.post_id    = p.post_id
		        AND    c.deleted_at IS NULL
		    ) AS comment_count,
		    (
		        SELECT COUNT(*)
		        FROM   reactions rk
		        WHERE  rk.post_id      = p.post_id
		        AND    rk.reaction_type = 1
		    ) AS like_count,
		    (
		        SELECT COUNT(*)
		        FROM   reactions rd
		        WHERE  rd.post_id      = p.post_id
		        AND    rd.reaction_type = 2
		    ) AS dislike_count,
		    (
		        SELECT rv.reaction_type
		        FROM   reactions rv
		        WHERE  rv.post_id = p.post_id
		        AND    rv.user_id = ?
		        LIMIT  1
		    ) AS viewer_reaction
		FROM  posts p
		JOIN  user u
		      ON  u.user_id = p.user_id
		LEFT  JOIN user_avatars ua
		      ON  ua.user_id = u.user_id
		LEFT  JOIN images_core ic_av
		      ON  ic_av.image_id   = ua.image_id
		      AND ic_av.deleted_at IS NULL
		WHERE p.group_id IS NULL
		AND   p.created_at %s ?
		AND   (
		          p.user_id = ?
		      OR  p.visibility = 'public'
		      OR  (
		              p.visibility = 'followers'
		          AND EXISTS (
		                  SELECT 1
		                  FROM   follow_relationships fr
		                  WHERE  fr.follower_id = ?
		                  AND    fr.followee_id = p.user_id
		                  AND    fr.status      = 'accepted'
		              )
		          )
		      OR  (
		              p.visibility = 'private'
		          AND EXISTS (
		                  SELECT 1
		                  FROM   post_allowed_users pau
		                  WHERE  pau.post_id = p.post_id
		                  AND    pau.user_id = ?
		              )
		          )
		      )
		ORDER BY %s
		LIMIT ?
	`, cursorOp, orderClause)

	// Bind parameters in the exact order the ? placeholders appear:
	//   1  viewer_reaction subquery  → ViewerID
	//   2  cursor comparison         → cursorTime
	//   3  rule 1 (own posts)        → ViewerID
	//   4  rule 3 (followers)        → ViewerID
	//   5  rule 4 (private allowed)  → ViewerID
	//   6  LIMIT                     → Limit + 1
	rows, err := r.db.Query(query,
		params.ViewerID,
		cursorTime,
		params.ViewerID,
		params.ViewerID,
		params.ViewerID,
		params.Limit+1,
	)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	var posts []models.FeedPost
	for rows.Next() {
		p, err := scanFeedPost(rows)
		if err != nil {
			return nil, false, fmt.Errorf("feed: scan row: %w", err)
		}
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("feed: rows iteration: %w", err)
	}

	// Detect whether more pages exist by checking for the extra row.
	hasMore := len(posts) > params.Limit
	if hasMore {
		posts = posts[:params.Limit]
	}

	// For "prev" direction the SQL returned rows oldest-first; reverse so that
	// the caller always receives newest-first.
	if params.Direction == "prev" {
		for i, j := 0, len(posts)-1; i < j; i, j = i+1, j-1 {
			posts[i], posts[j] = posts[j], posts[i]
		}
	}

	return posts, hasMore, nil
}

// ============================================================================
// batchFetchImages — second query
// ============================================================================

// batchFetchImages retrieves all non-deleted images for the given post IDs in
// a single query and returns them grouped by post_id.
func (r *FeedRepository) batchFetchImages(postIDs []string) (map[string][]models.FeedImage, error) {
	if len(postIDs) == 0 {
		return make(map[string][]models.FeedImage), nil
	}

	// Build the IN(?, ?, …) placeholder string.
	placeholders := strings.Repeat("?,", len(postIDs))
	placeholders = placeholders[:len(placeholders)-1] // strip trailing comma

	query := fmt.Sprintf(`
		SELECT
		    pi.post_id,
		    ic.image_id,
		    ic.file_path,
		    ic.thumbnail_path,
		    pi.display_order
		FROM  post_images pi
		JOIN  images_core ic
		      ON  ic.image_id   = pi.image_id
		      AND ic.deleted_at IS NULL
		WHERE pi.post_id IN (%s)
		ORDER BY pi.post_id, pi.display_order ASC
	`, placeholders)

	// Convert []string to []interface{} for db.Query variadic args.
	args := make([]interface{}, len(postIDs))
	for i, id := range postIDs {
		args[i] = id
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]models.FeedImage, len(postIDs))
	for rows.Next() {
		var (
			postID       string
			imageID      string
			filePath     string
			thumbPath    string
			displayOrder int
		)
		if err := rows.Scan(&postID, &imageID, &filePath, &thumbPath, &displayOrder); err != nil {
			return nil, fmt.Errorf("feed: scan image row: %w", err)
		}
		result[postID] = append(result[postID], models.FeedImage{
			ImageID:      imageID,
			URL:          toStaticURL(filePath),
			ThumbnailURL: toStaticURL(thumbPath),
			DisplayOrder: displayOrder,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("feed: image rows iteration: %w", err)
	}

	return result, nil
}

// ============================================================================
// scanFeedPost — row scanner
// ============================================================================

// scanFeedPost scans one row from the primary post query into a FeedPost.
// viewer_reaction is nullable (NULL when the viewer has no reaction), so it
// is scanned into sql.NullInt64 and converted to *int afterwards.
func scanFeedPost(rows *sql.Rows) (models.FeedPost, error) {
	var (
		p              models.FeedPost
		viewerReaction sql.NullInt64
	)

	// Images is always an allocated slice so the frontend never receives null.
	p.Images = []models.FeedImage{}

	err := rows.Scan(
		&p.ID,
		&p.AuthorID,
		&p.AuthorNickname,
		&p.AuthorFirstName,
		&p.AuthorLastName,
		&p.AuthorAvatarURL,
		&p.AuthorAvatarThumbURL,
		&p.Visibility,
		&p.Title,
		&p.Content,
		&p.CreatedAt,
		&p.UpdatedAt,
		&p.CommentCount,
		&p.LikeCount,
		&p.DislikeCount,
		&viewerReaction,
	)
	if err != nil {
		return p, err
	}

	if viewerReaction.Valid {
		v := int(viewerReaction.Int64)
		p.ViewerReaction = &v
	}

	// The avatar paths are stored as "uploads/filename.jpg" in images_core.
	// Convert them to fully-qualified static URLs here, consistent with
	// how post images are handled in batchFetchImages.
	p.AuthorAvatarURL = toStaticURL(p.AuthorAvatarURL)
	p.AuthorAvatarThumbURL = toStaticURL(p.AuthorAvatarThumbURL)

	return p, nil
}

// scanFeedPostRow is identical to scanFeedPost but accepts *sql.Row (QueryRow)
// instead of *sql.Rows (Query).  The two types share the same column contract
// but expose different interfaces, so both helpers are required.
func scanFeedPostRow(row *sql.Row) (*models.FeedPost, error) {
	var (
		p              models.FeedPost
		viewerReaction sql.NullInt64
	)

	p.Images = []models.FeedImage{}

	err := row.Scan(
		&p.ID,
		&p.AuthorID,
		&p.AuthorNickname,
		&p.AuthorFirstName,
		&p.AuthorLastName,
		&p.AuthorAvatarURL,
		&p.AuthorAvatarThumbURL,
		&p.Visibility,
		&p.Title,
		&p.Content,
		&p.CreatedAt,
		&p.UpdatedAt,
		&p.CommentCount,
		&p.LikeCount,
		&p.DislikeCount,
		&viewerReaction,
	)
	if err != nil {
		return nil, err
	}

	if viewerReaction.Valid {
		v := int(viewerReaction.Int64)
		p.ViewerReaction = &v
	}

	p.AuthorAvatarURL = toStaticURL(p.AuthorAvatarURL)
	p.AuthorAvatarThumbURL = toStaticURL(p.AuthorAvatarThumbURL)

	return &p, nil
}