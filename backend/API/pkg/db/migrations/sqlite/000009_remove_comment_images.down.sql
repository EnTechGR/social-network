-- Migration 000009 Rollback: Add comment image columns back

-- ============================================================================
-- STEP 1: Restore image columns to comments table
-- ============================================================================

-- Create new comments table with image columns
CREATE TABLE IF NOT EXISTS comments_new (
    comment_id TEXT PRIMARY KEY,
    post_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    content TEXT CHECK (LENGTH(content) <= 1000),
    image_path TEXT,
    image_thumbnail_path TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE
);

-- Copy data from old table to new table (image columns will be NULL)
INSERT INTO comments_new (comment_id, post_id, user_id, content, image_path, image_thumbnail_path, created_at, updated_at, deleted_at)
SELECT comment_id, post_id, user_id, content, NULL, NULL, created_at, updated_at, deleted_at
FROM comments;

-- Drop old table
DROP TABLE comments;

-- Rename new table to comments
ALTER TABLE comments_new RENAME TO comments;

-- ============================================================================
-- STEP 2: Recreate indexes for comments table
-- ============================================================================
CREATE INDEX IF NOT EXISTS idx_comments_post ON comments(post_id);
CREATE INDEX IF NOT EXISTS idx_comments_user ON comments(user_id);

-- ============================================================================
-- STEP 3: Recreate v_comment_images view
-- ============================================================================
-- This view is used to join comment images with their metadata
CREATE VIEW IF NOT EXISTS v_comment_images AS
SELECT 
    c.comment_id,
    c.post_id,
    c.user_id as comment_author_id,
    c.image_path,
    c.image_thumbnail_path
FROM comments c
WHERE c.image_path IS NOT NULL;

-- ============================================================================
-- Rollback Complete
-- ============================================================================
