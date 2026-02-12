-- Migration 000005 ROLLBACK: Revert Centralized Image Management
-- This migration reverts all changes made in 000005_centralize_image_management_up.sql
-- and restores the original image handling structure

-- ============================================================================
-- STEP 1: Drop helper views
-- ============================================================================

DROP VIEW IF EXISTS v_comment_images;
DROP VIEW IF EXISTS v_post_images;
DROP VIEW IF EXISTS v_user_avatars;

-- ============================================================================
-- STEP 2: Recreate old table structures with image columns
-- ============================================================================

-- 2A. Recreate user table WITH avatar columns
CREATE TABLE user_old (
    user_id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE CHECK (LENGTH(email) <= 100),
    nickname TEXT NOT NULL UNIQUE CHECK (LENGTH(nickname) <= 50),
    first_name TEXT NOT NULL CHECK (LENGTH(first_name) <= 100),
    last_name TEXT NOT NULL CHECK (LENGTH(last_name) <= 100),
    date_of_birth DATE NOT NULL,
    avatar_path TEXT,
    avatar_thumbnail_path TEXT,
    about_me TEXT,
    gender TEXT CHECK (gender IN ('male','female','other','prefer_not_to_say')),
    is_private BOOLEAN NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 2B. Recreate comments table WITH image columns
CREATE TABLE comments_old (
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
    FOREIGN KEY (user_id) REFERENCES user_old(user_id) ON DELETE CASCADE
);

-- 2C. Recreate old images table (for post images)
CREATE TABLE images_old (
    image_id TEXT PRIMARY KEY,
    post_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    file_path TEXT NOT NULL,
    thumbnail_path TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user_old(user_id) ON DELETE CASCADE
);

-- 2D. Recreate old chat_images table
CREATE TABLE chat_images_old (
    image_id TEXT PRIMARY KEY,
    message_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    filename TEXT NOT NULL,
    original_filename TEXT NOT NULL,
    file_path TEXT NOT NULL,
    thumbnail_path TEXT NOT NULL,
    file_size INTEGER NOT NULL CHECK (file_size > 0),
    mime_type TEXT NOT NULL CHECK (mime_type IN ('image/jpeg','image/png','image/gif','image/webp')),
    width INTEGER NOT NULL CHECK (width > 0),
    height INTEGER NOT NULL CHECK (height > 0),
    uploaded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (message_id) REFERENCES messages(message_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user_old(user_id) ON DELETE CASCADE
);

-- 2E. Recreate old group_message_images table
CREATE TABLE group_message_images_old (
    image_id TEXT PRIMARY KEY,
    message_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    filename TEXT NOT NULL,
    original_filename TEXT NOT NULL,
    file_path TEXT NOT NULL,
    thumbnail_path TEXT NOT NULL,
    file_size INTEGER NOT NULL CHECK (file_size > 0),
    mime_type TEXT NOT NULL CHECK (mime_type IN ('image/jpeg','image/png','image/gif','image/webp')),
    width INTEGER NOT NULL CHECK (width > 0),
    height INTEGER NOT NULL CHECK (height > 0),
    uploaded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (message_id) REFERENCES group_messages(message_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user_old(user_id) ON DELETE CASCADE
);

-- ============================================================================
-- STEP 3: Migrate data back from new schema to old schema
-- ============================================================================

-- 3A. Restore user data with avatar columns
INSERT INTO user_old (
    user_id, email, nickname, first_name, last_name, 
    date_of_birth, avatar_path, avatar_thumbnail_path, about_me, gender, is_private, created_at
)
SELECT 
    u.user_id,
    u.email,
    u.nickname,
    u.first_name,
    u.last_name,
    u.date_of_birth,
    ic.file_path as avatar_path,
    ic.thumbnail_path as avatar_thumbnail_path,
    u.about_me,
    u.gender,
    u.is_private,
    u.created_at
FROM user u
LEFT JOIN user_avatars ua ON u.user_id = ua.user_id
LEFT JOIN images_core ic ON ua.image_id = ic.image_id;

-- 3B. Restore comments with image columns
INSERT INTO comments_old (
    comment_id, post_id, user_id, content, 
    image_path, image_thumbnail_path, created_at, updated_at, deleted_at
)
SELECT 
    c.comment_id,
    c.post_id,
    c.user_id,
    c.content,
    ic.file_path as image_path,
    ic.thumbnail_path as image_thumbnail_path,
    c.created_at,
    c.updated_at,
    c.deleted_at
FROM comments c
LEFT JOIN comment_images ci ON c.comment_id = ci.comment_id
LEFT JOIN images_core ic ON ci.image_id = ic.image_id;

-- 3C. Restore post images table
INSERT INTO images_old (image_id, post_id, user_id, file_path, thumbnail_path, created_at)
SELECT 
    ic.image_id,
    pi.post_id,
    ic.uploader_user_id as user_id,
    ic.file_path,
    ic.thumbnail_path,
    ic.uploaded_at as created_at
FROM post_images pi
INNER JOIN images_core ic ON pi.image_id = ic.image_id;

-- 3D. Restore chat_images table
INSERT INTO chat_images_old (
    image_id, message_id, user_id, filename, original_filename,
    file_path, thumbnail_path, file_size, mime_type, width, height, uploaded_at
)
SELECT 
    ic.image_id,
    mi.message_id,
    ic.uploader_user_id as user_id,
    ic.filename,
    ic.original_filename,
    ic.file_path,
    ic.thumbnail_path,
    ic.file_size,
    ic.mime_type,
    ic.width,
    ic.height,
    ic.uploaded_at
FROM message_images mi
INNER JOIN images_core ic ON mi.image_id = ic.image_id;

-- 3E. Restore group_message_images table
INSERT INTO group_message_images_old (
    image_id, message_id, user_id, filename, original_filename,
    file_path, thumbnail_path, file_size, mime_type, width, height, uploaded_at
)
SELECT 
    ic.image_id,
    gmi.group_message_id as message_id,
    ic.uploader_user_id as user_id,
    ic.filename,
    ic.original_filename,
    ic.file_path,
    ic.thumbnail_path,
    ic.file_size,
    ic.mime_type,
    ic.width,
    ic.height,
    ic.uploaded_at
FROM group_message_images gmi
INNER JOIN images_core ic ON gmi.image_id = ic.image_id;

-- ============================================================================
-- STEP 4: Drop new schema tables
-- ============================================================================

-- Drop relationship tables
DROP TABLE IF EXISTS user_avatars;
DROP TABLE IF EXISTS post_images;
DROP TABLE IF EXISTS comment_images;
DROP TABLE IF EXISTS message_images;
DROP TABLE IF EXISTS group_message_images;

-- Drop core images table
DROP TABLE IF EXISTS images_core;

-- Drop indexes related to new tables (if they weren't cascade-deleted)
DROP INDEX IF EXISTS idx_images_core_uploader;
DROP INDEX IF EXISTS idx_images_core_uploaded;
DROP INDEX IF EXISTS idx_images_core_deleted;
DROP INDEX IF EXISTS idx_user_avatars_image;
DROP INDEX IF EXISTS idx_post_images_post;
DROP INDEX IF EXISTS idx_post_images_image;
DROP INDEX IF EXISTS idx_comment_images_comment;
DROP INDEX IF EXISTS idx_comment_images_image;
DROP INDEX IF EXISTS idx_message_images_message;
DROP INDEX IF EXISTS idx_message_images_image;
DROP INDEX IF EXISTS idx_group_message_images_new_message;
DROP INDEX IF EXISTS idx_group_message_images_new_image;

-- ============================================================================
-- STEP 5: Replace current tables with old structure
-- ============================================================================

-- 5A. Replace user table
DROP TABLE user;
ALTER TABLE user_old RENAME TO user;

-- 5B. Replace comments table
DROP TABLE comments;
ALTER TABLE comments_old RENAME TO comments;

-- 5C. Rename restored images tables to original names
ALTER TABLE images_old RENAME TO images;
ALTER TABLE chat_images_old RENAME TO chat_images;
ALTER TABLE group_message_images_old RENAME TO group_message_images;

-- ============================================================================
-- STEP 6: Recreate original indexes
-- ============================================================================

-- Comments indexes
CREATE INDEX IF NOT EXISTS idx_comments_post ON comments(post_id);
CREATE INDEX IF NOT EXISTS idx_comments_user ON comments(user_id);

-- Images indexes
CREATE INDEX IF NOT EXISTS idx_images_post ON images(post_id);

-- Chat images indexes
CREATE INDEX IF NOT EXISTS idx_chat_images_message ON chat_images(message_id);

-- Group messages indexes (if they were dropped)
CREATE INDEX IF NOT EXISTS idx_group_messages_group ON group_messages(group_id);
CREATE INDEX IF NOT EXISTS idx_group_messages_sender ON group_messages(sender_id);

-- ============================================================================
-- Rollback Complete
-- ============================================================================

-- Summary of rollback:
-- ✅ Dropped helper views
-- ✅ Recreated original table structures with image columns
-- ✅ Migrated all data back from centralized schema to original schema
-- ✅ Dropped all new centralized image management tables
-- ✅ Restored original table names
-- ✅ Recreated all original indexes
-- ⚠️  Note: Some metadata (file_size, width, height) may be 0 for old avatar/comment images