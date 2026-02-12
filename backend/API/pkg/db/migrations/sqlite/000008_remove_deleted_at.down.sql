-- Migration 000008 Rollback: Add soft delete (deleted_at column) back to posts table

-- ============================================================================
-- STEP 1: Drop views that depend on posts table
-- ============================================================================
DROP VIEW IF EXISTS v_post_images;

-- ============================================================================
-- STEP 2: Recreate posts table with deleted_at column
-- ============================================================================

-- Create new posts table with deleted_at column
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

-- Copy data from old table (deleted_at will be NULL for all posts)
INSERT INTO posts_new (post_id, user_id, group_id, visibility, title, content, created_at, updated_at, deleted_at)
SELECT post_id, user_id, group_id, visibility, title, content, created_at, updated_at, NULL
FROM posts;

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
-- STEP 4: Recreate v_post_images view
-- ============================================================================
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
-- Rollback Complete
-- ============================================================================
