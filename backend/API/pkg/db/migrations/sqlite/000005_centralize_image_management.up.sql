-- Migration 000005: Centralize Image Management
-- This migration creates a unified image management system where all images are stored in a core 'images' table
-- with context-specific relationship tables linking images to their respective entities (users, posts, comments, messages)

-- ============================================================================
-- STEP 1: Create the core images table (single source of truth for all images)
-- ============================================================================
-- Note: CHECK constraints allow >= 0 instead of > 0 to accommodate migrated historical data
-- where file_size, width, and height weren't captured. New uploads will have proper values.

CREATE TABLE IF NOT EXISTS images_core (
    image_id TEXT PRIMARY KEY,
    uploader_user_id TEXT NOT NULL,
    filename TEXT NOT NULL,                    -- Generated unique filename
    original_filename TEXT NOT NULL,           -- Original name from user upload
    file_path TEXT NOT NULL,                   -- Full/original size path
    thumbnail_path TEXT NOT NULL,              -- Thumbnail version path
    file_size INTEGER NOT NULL CHECK (file_size >= 0),  -- Allow 0 for migrated historical data
    mime_type TEXT NOT NULL CHECK (mime_type IN ('image/jpeg','image/png','image/gif','image/webp')),
    width INTEGER NOT NULL CHECK (width >= 0),          -- Allow 0 for migrated historical data
    height INTEGER NOT NULL CHECK (height >= 0),        -- Allow 0 for migrated historical data
    uploaded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,                      -- Soft delete for cleanup jobs
    FOREIGN KEY (uploader_user_id) REFERENCES user(user_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_images_core_uploader ON images_core(uploader_user_id);
CREATE INDEX IF NOT EXISTS idx_images_core_uploaded ON images_core(uploaded_at);
CREATE INDEX IF NOT EXISTS idx_images_core_deleted ON images_core(deleted_at);

-- ============================================================================
-- STEP 2: Create relationship tables to link images to their contexts
-- ============================================================================

-- A. User Avatars (1:1 relationship - one current avatar per user)
CREATE TABLE IF NOT EXISTS user_avatars (
    user_id TEXT PRIMARY KEY,
    image_id TEXT NOT NULL,
    set_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE,
    FOREIGN KEY (image_id) REFERENCES images_core(image_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_user_avatars_image ON user_avatars(image_id);

-- B. Post Images (1:N relationship - multiple images per post possible)
-- We'll rename the existing 'images' table and convert it to a relationship table
CREATE TABLE IF NOT EXISTS post_images (
    post_image_id TEXT PRIMARY KEY,
    post_id TEXT NOT NULL,
    image_id TEXT NOT NULL,
    display_order INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE,
    FOREIGN KEY (image_id) REFERENCES images_core(image_id) ON DELETE CASCADE,
    UNIQUE(post_id, display_order)
);

CREATE INDEX IF NOT EXISTS idx_post_images_post ON post_images(post_id);
CREATE INDEX IF NOT EXISTS idx_post_images_image ON post_images(image_id);

-- C. Comment Images (1:1 relationship - one image per comment)
CREATE TABLE IF NOT EXISTS comment_images (
    comment_image_id TEXT PRIMARY KEY,
    comment_id TEXT NOT NULL UNIQUE,
    image_id TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (comment_id) REFERENCES comments(comment_id) ON DELETE CASCADE,
    FOREIGN KEY (image_id) REFERENCES images_core(image_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_comment_images_comment ON comment_images(comment_id);
CREATE INDEX IF NOT EXISTS idx_comment_images_image ON comment_images(image_id);

-- D. Message Images (1:N relationship - multiple images per message possible)
CREATE TABLE IF NOT EXISTS message_images (
    message_image_id TEXT PRIMARY KEY,
    message_id TEXT NOT NULL,
    image_id TEXT NOT NULL,
    display_order INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (message_id) REFERENCES messages(message_id) ON DELETE CASCADE,
    FOREIGN KEY (image_id) REFERENCES images_core(image_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_message_images_message ON message_images(message_id);
CREATE INDEX IF NOT EXISTS idx_message_images_image ON message_images(image_id);

-- E. Group Message Images (1:N relationship - multiple images per group message possible)
CREATE TABLE IF NOT EXISTS group_message_images_new (
    group_message_image_id TEXT PRIMARY KEY,
    group_message_id TEXT NOT NULL,
    image_id TEXT NOT NULL,
    display_order INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (group_message_id) REFERENCES group_messages(message_id) ON DELETE CASCADE,
    FOREIGN KEY (image_id) REFERENCES images_core(image_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_group_message_images_new_message ON group_message_images_new(group_message_id);
CREATE INDEX IF NOT EXISTS idx_group_message_images_new_image ON group_message_images_new(image_id);

-- ============================================================================
-- STEP 3: Migrate existing data from old schema to new schema
-- ============================================================================

-- 3A. Migrate user avatars
-- Only migrate users who have avatar_path set (not NULL)
INSERT INTO images_core (
    image_id,
    uploader_user_id,
    filename,
    original_filename,
    file_path,
    thumbnail_path,
    file_size,
    mime_type,
    width,
    height,
    uploaded_at
)
SELECT 
    'avatar_' || user_id as image_id,
    user_id as uploader_user_id,
    COALESCE(
        substr(avatar_path, instr(avatar_path, '/') + 1),  -- Extract filename from path
        'unknown.jpg'
    ) as filename,
    'avatar.jpg' as original_filename,                      -- We don't have the original name stored
    avatar_path as file_path,
    COALESCE(avatar_thumbnail_path, avatar_path) as thumbnail_path,
    0 as file_size,                                         -- Unknown, set to 0 (can be updated later)
    'image/jpeg' as mime_type,                              -- Default assumption
    0 as width,                                             -- Unknown, set to 0
    0 as height,                                            -- Unknown, set to 0
    created_at as uploaded_at
FROM user
WHERE avatar_path IS NOT NULL;

-- Link users to their avatar images
INSERT INTO user_avatars (user_id, image_id, set_at)
SELECT 
    user_id,
    'avatar_' || user_id as image_id,
    created_at as set_at
FROM user
WHERE avatar_path IS NOT NULL;

-- 3B. Migrate post images from existing 'images' table
-- First, copy the image metadata to images_core
INSERT INTO images_core (
    image_id,
    uploader_user_id,
    filename,
    original_filename,
    file_path,
    thumbnail_path,
    file_size,
    mime_type,
    width,
    height,
    uploaded_at
)
SELECT 
    image_id,
    user_id as uploader_user_id,
    COALESCE(
        substr(file_path, instr(file_path, '/') + 1),
        'unknown.jpg'
    ) as filename,
    'post_image.jpg' as original_filename,                  -- We don't have the original name stored
    file_path,
    thumbnail_path,
    0 as file_size,                                         -- Unknown from old schema
    'image/jpeg' as mime_type,                              -- Default assumption
    0 as width,                                             -- Unknown from old schema
    0 as height,                                            -- Unknown from old schema
    created_at as uploaded_at
FROM images;

-- Create the post-to-image relationships
INSERT INTO post_images (post_image_id, post_id, image_id, display_order, created_at)
SELECT 
    'post_img_' || image_id as post_image_id,
    post_id,
    image_id,
    1 as display_order,                                     -- Single image per post in old schema
    created_at
FROM images;

-- 3C. Migrate comment images
-- Only migrate comments that have image_path set (not NULL)
INSERT INTO images_core (
    image_id,
    uploader_user_id,
    filename,
    original_filename,
    file_path,
    thumbnail_path,
    file_size,
    mime_type,
    width,
    height,
    uploaded_at
)
SELECT 
    'comment_' || comment_id as image_id,
    user_id as uploader_user_id,
    COALESCE(
        substr(image_path, instr(image_path, '/') + 1),
        'unknown.jpg'
    ) as filename,
    'comment_image.jpg' as original_filename,
    image_path as file_path,
    COALESCE(image_thumbnail_path, image_path) as thumbnail_path,
    0 as file_size,
    'image/jpeg' as mime_type,
    0 as width,
    0 as height,
    created_at as uploaded_at
FROM comments
WHERE image_path IS NOT NULL;

-- Link comments to their images
INSERT INTO comment_images (comment_image_id, comment_id, image_id, created_at)
SELECT 
    'cmt_img_' || comment_id as comment_image_id,
    comment_id,
    'comment_' || comment_id as image_id,
    created_at
FROM comments
WHERE image_path IS NOT NULL;

-- 3D. Migrate chat images (1:1 messages)
-- Copy image metadata to images_core
INSERT INTO images_core (
    image_id,
    uploader_user_id,
    filename,
    original_filename,
    file_path,
    thumbnail_path,
    file_size,
    mime_type,
    width,
    height,
    uploaded_at
)
SELECT 
    image_id,
    user_id as uploader_user_id,
    filename,
    original_filename,
    file_path,
    thumbnail_path,
    file_size,
    mime_type,
    width,
    height,
    uploaded_at
FROM chat_images;

-- Create message-to-image relationships
INSERT INTO message_images (message_image_id, message_id, image_id, display_order, created_at)
SELECT 
    'msg_img_' || image_id as message_image_id,
    message_id,
    image_id,
    1 as display_order,
    uploaded_at as created_at
FROM chat_images;

-- 3E. Migrate group message images
-- Copy image metadata to images_core
INSERT INTO images_core (
    image_id,
    uploader_user_id,
    filename,
    original_filename,
    file_path,
    thumbnail_path,
    file_size,
    mime_type,
    width,
    height,
    uploaded_at
)
SELECT 
    image_id,
    user_id as uploader_user_id,
    filename,
    original_filename,
    file_path,
    thumbnail_path,
    file_size,
    mime_type,
    width,
    height,
    uploaded_at
FROM group_message_images;

-- Create group-message-to-image relationships
INSERT INTO group_message_images_new (group_message_image_id, group_message_id, image_id, display_order, created_at)
SELECT 
    'grp_msg_img_' || image_id as group_message_image_id,
    message_id as group_message_id,
    image_id,
    1 as display_order,
    uploaded_at as created_at
FROM group_message_images;

-- ============================================================================
-- STEP 4: Drop old image-related columns and tables
-- ============================================================================

-- 4A. Remove avatar columns from user table
CREATE TABLE user_temp (
    user_id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE CHECK (LENGTH(email) <= 100),
    nickname TEXT NOT NULL UNIQUE CHECK (LENGTH(nickname) <= 50),
    first_name TEXT NOT NULL CHECK (LENGTH(first_name) <= 100),
    last_name TEXT NOT NULL CHECK (LENGTH(last_name) <= 100),
    date_of_birth DATE NOT NULL,
    about_me TEXT,
    gender TEXT CHECK (gender IN ('male','female','other','prefer_not_to_say')),
    is_private BOOLEAN NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO user_temp SELECT 
    user_id, email, nickname, first_name, last_name, 
    date_of_birth, about_me, gender, is_private, created_at
FROM user;

DROP TABLE user;
ALTER TABLE user_temp RENAME TO user;

-- 4B. Remove image columns from comments table
CREATE TABLE comments_temp (
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

INSERT INTO comments_temp SELECT 
    comment_id, post_id, user_id, content, created_at, updated_at, deleted_at
FROM comments;

DROP TABLE comments;
ALTER TABLE comments_temp RENAME TO comments;

-- Recreate indexes for comments
CREATE INDEX IF NOT EXISTS idx_comments_post ON comments(post_id);
CREATE INDEX IF NOT EXISTS idx_comments_user ON comments(user_id);

-- 4C. Drop old images table (now replaced by images_core + post_images)
DROP TABLE images;
DROP INDEX IF EXISTS idx_images_post;

-- 4D. Drop old chat_images table (now replaced by images_core + message_images)
DROP TABLE chat_images;
DROP INDEX IF EXISTS idx_chat_images_message;

-- 4E. Drop old group_message_images table (now replaced by images_core + group_message_images_new)
DROP TABLE group_message_images;

-- Rename the new group_message_images table to the standard name
ALTER TABLE group_message_images_new RENAME TO group_message_images;

-- ============================================================================
-- STEP 5: Create additional helper views for easy querying (optional but useful)
-- ============================================================================

-- View to get user avatar with all image details
CREATE VIEW IF NOT EXISTS v_user_avatars AS
SELECT 
    u.user_id,
    u.nickname,
    u.first_name,
    u.last_name,
    ic.image_id,
    ic.file_path,
    ic.thumbnail_path,
    ic.mime_type,
    ua.set_at
FROM user u
LEFT JOIN user_avatars ua ON u.user_id = ua.user_id
LEFT JOIN images_core ic ON ua.image_id = ic.image_id;

-- View to get post images with all details
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

-- View to get comment images with all details
CREATE VIEW IF NOT EXISTS v_comment_images AS
SELECT 
    c.comment_id,
    c.post_id,
    c.user_id as comment_author_id,
    ci.comment_image_id,
    ic.image_id,
    ic.file_path,
    ic.thumbnail_path,
    ic.mime_type,
    ic.uploaded_at
FROM comments c
INNER JOIN comment_images ci ON c.comment_id = ci.comment_id
INNER JOIN images_core ic ON ci.image_id = ic.image_id;

-- ============================================================================
-- Migration Complete
-- ============================================================================

-- Summary of changes:
-- ✅ Created centralized images_core table for all image metadata
-- ✅ Created relationship tables: user_avatars, post_images, comment_images, message_images, group_message_images
-- ✅ Migrated all existing image data to new schema
-- ✅ Removed image columns from user and comments tables
-- ✅ Replaced old images, chat_images, group_message_images tables
-- ✅ Created helper views for easier querying
-- ✅ All foreign keys and indexes properly set up