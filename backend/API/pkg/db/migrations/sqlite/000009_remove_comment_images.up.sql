-- Migration 000009: Remove comment image columns (use centralized image management)
-- This migration removes the inline image columns from the comments table
-- since we now use the centralized images_core table for all image management

-- ============================================================================
-- STEP 1: Drop views that depend on comments table
-- ============================================================================
DROP VIEW IF EXISTS v_comment_images;

-- ============================================================================
-- STEP 2: Remove image columns from comments table
-- ============================================================================
-- SQLite doesn't support DROP COLUMN directly, so we need to recreate the table

-- Create new comments table without image columns
CREATE TABLE IF NOT EXISTS comments_new (
    comment_id TEXT PRIMARY KEY,
    post_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    content TEXT CHECK (LENGTH(content) <= 1000),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE
);

-- Copy data from old table to new table (without image columns)
INSERT INTO comments_new (comment_id, post_id, user_id, content, created_at, updated_at, deleted_at)
SELECT comment_id, post_id, user_id, content, created_at, updated_at, deleted_at
FROM comments;

-- Drop old table
DROP TABLE comments;

-- Rename new table to comments
ALTER TABLE comments_new RENAME TO comments;

-- ============================================================================
-- STEP 3: Recreate indexes for comments table
-- ============================================================================
CREATE INDEX IF NOT EXISTS idx_comments_post ON comments(post_id);
CREATE INDEX IF NOT EXISTS idx_comments_user ON comments(user_id);

-- ============================================================================
-- Migration Complete
-- ============================================================================
-- Summary:
-- ✅ Dropped v_comment_images view (dependency)
-- ✅ Removed image_path column from comments
-- ✅ Removed image_thumbnail_path column from comments
-- ✅ All image management now uses centralized images_core table
-- ✅ Comments can still have images via post_images relationship
-- ✅ Recreated comments indexes

