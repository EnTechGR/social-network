-- Migration 000008: Remove soft delete (deleted_at column) from posts table
-- This migration converts soft deletes to hard deletes
-- Summary: Remove deleted_at column from posts table

-- ============================================================================
-- STEP 1: Drop views that depend on posts table
-- ============================================================================
DROP VIEW IF EXISTS v_post_images;

-- ============================================================================
-- STEP 2: Remove deleted_at column and recreate posts table
-- ============================================================================
-- SQLite doesn't support DROP COLUMN directly, so we need to recreate the table

-- Create new posts table without deleted_at
CREATE TABLE IF NOT EXISTS posts_new (
    post_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    group_id TEXT,
    visibility TEXT NOT NULL DEFAULT 'public' CHECK (visibility IN ('public','followers','private')),
    title TEXT CHECK (LENGTH(title) <= 200),
    content TEXT CHECK (LENGTH(content) <= 2000),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES groups(group_id) ON DELETE CASCADE
);

-- Copy data from old table to new table (excluding deleted_at)
-- Only copy posts where deleted_at IS NULL to keep active posts
INSERT INTO posts_new (post_id, user_id, group_id, visibility, title, content, created_at, updated_at)
SELECT post_id, user_id, group_id, visibility, title, content, created_at, updated_at
FROM posts
WHERE deleted_at IS NULL;

-- Drop old table
DROP TABLE posts;

-- Rename new table to posts
ALTER TABLE posts_new RENAME TO posts;

-- ============================================================================
-- STEP 3: Recreate indexes for posts table
-- ============================================================================
CREATE INDEX IF NOT EXISTS idx_posts_user ON posts(user_id);
CREATE INDEX IF NOT EXISTS idx_posts_visibility ON posts(visibility);
CREATE INDEX IF NOT EXISTS idx_posts_group ON posts(group_id);

-- ============================================================================
-- Migration Complete
-- ============================================================================
-- Summary:
-- ✅ Dropped v_post_images view (dependency)
-- ✅ Recreated posts table without deleted_at column
-- ✅ Migrated all active posts (deleted_at IS NULL)
-- ✅ Deleted posts are permanently removed
-- ✅ Recreated posts indexes
-- ✅ v_post_images view no longer needed (using images_core directly)
