-- Migration 000007 ROLLBACK: Restore Categories
-- This rollback migration restores the categories functionality
-- FIXED: Now properly handles v_post_images view dependency
-- Note: This will restore the structure but not the data

-- ============================================================================
-- STEP 1: Drop views that depend on posts table
-- ============================================================================
DROP VIEW IF EXISTS v_post_images;

-- ============================================================================
-- STEP 2: Recreate categories table
-- ============================================================================
CREATE TABLE IF NOT EXISTS categories (
    category_id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

-- ============================================================================
-- STEP 3: Recreate post_categories junction table
-- ============================================================================
CREATE TABLE IF NOT EXISTS post_categories (
    post_id TEXT NOT NULL,
    category_id INTEGER NOT NULL,
    PRIMARY KEY (post_id, category_id),
    FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES categories(category_id) ON DELETE CASCADE
);

-- ============================================================================
-- STEP 4: Recreate indexes
-- ============================================================================
CREATE INDEX IF NOT EXISTS idx_post_categories_post ON post_categories(post_id);
CREATE INDEX IF NOT EXISTS idx_post_categories_cat ON post_categories(category_id);

-- ============================================================================
-- STEP 5: Optionally add back category_id to posts (if you had it)
-- ============================================================================
-- Note: This would require recreating the posts table with category_id column
-- For simplicity, we're skipping this in the rollback
-- You would need to manually add this if you want full rollback

-- ============================================================================
-- STEP 6: Recreate v_post_images view
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
-- STEP 7: Insert default categories (optional)
-- ============================================================================
INSERT OR IGNORE INTO categories (name) VALUES 
    ('General'),
    ('Technology'),
    ('Entertainment'),
    ('Sports'),
    ('News');

-- ============================================================================
-- Rollback Complete
-- ============================================================================