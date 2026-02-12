-- Migration 000007: Remove Categories from Social Network
-- This migration removes all category-related tables and columns
-- FIXED: Now properly handles v_post_images view dependency

-- ============================================================================
-- STEP 1: Drop views that depend on posts table
-- ============================================================================
-- The v_post_images view references the posts table, so we need to drop it
-- before we can drop and recreate the posts table
DROP VIEW IF EXISTS v_post_images;

-- ============================================================================
-- STEP 2: Drop category-related tables
-- ============================================================================
-- Drop junction table first (due to foreign key constraints)
DROP TABLE IF EXISTS post_categories;

-- Drop the categories table
DROP TABLE IF EXISTS categories;

-- ============================================================================
-- STEP 3: Recreate posts table without category_id column
-- ============================================================================
-- SQLite doesn't support DROP COLUMN directly, so we need to recreate the table

-- Create new posts table without category_id
CREATE TABLE IF NOT EXISTS posts_new (
    post_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    group_id TEXT,
    visibility TEXT NOT NULL DEFAULT 'public' CHECK (visibility IN ('public','followers','private')),
    title TEXT CHECK (LENGTH(title) <= 200),
    content TEXT CHECK (LENGTH(content) <= 2000),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES groups(group_id) ON DELETE CASCADE
);

-- Copy data from old table to new table (excluding category_id if it exists)
INSERT INTO posts_new (post_id, user_id, group_id, visibility, title, content, created_at, updated_at, deleted_at)
SELECT post_id, user_id, group_id, visibility, title, content, created_at, updated_at, deleted_at
FROM posts;

-- Drop old table
DROP TABLE posts;

-- Rename new table to posts
ALTER TABLE posts_new RENAME TO posts;

-- ============================================================================
-- STEP 4: Recreate indexes for posts table
-- ============================================================================
CREATE INDEX IF NOT EXISTS idx_posts_user ON posts(user_id);
CREATE INDEX IF NOT EXISTS idx_posts_visibility ON posts(visibility);
CREATE INDEX IF NOT EXISTS idx_posts_group ON posts(group_id);

-- ============================================================================
-- STEP 5: Recreate v_post_images view (from migration 000005)
-- ============================================================================
-- Recreate the view with the same definition as in 000005
CREATE VIEW IF NOT EXISTS v_post_images AS
SELECT 
    p.post_id,
    p.user_id as post_author_id,
    pi.post_image_id,
    pi.display_order,
    ic.image_id,
    ic.file_path,
    ic.thumbnail_path,
    ic.mime_type,
    ic.width,
    ic.height,
    ic.uploaded_at
FROM posts p
INNER JOIN post_images pi ON p.post_id = pi.post_id
INNER JOIN images_core ic ON pi.image_id = ic.image_id
ORDER BY p.post_id, pi.display_order;

-- ============================================================================
-- STEP 6: Clean up any remaining category-related indexes
-- ============================================================================
DROP INDEX IF EXISTS idx_post_categories_post;
DROP INDEX IF EXISTS idx_post_categories_cat;

-- ============================================================================
-- Migration Complete
-- ============================================================================
-- Summary:
-- ✅ Dropped v_post_images view (dependency)
-- ✅ Dropped post_categories junction table
-- ✅ Dropped categories table  
-- ✅ Recreated posts table without category_id
-- ✅ Preserved all existing post data
-- ✅ Recreated posts indexes
-- ✅ Recreated v_post_images view
-- ✅ Cleaned up category indexes