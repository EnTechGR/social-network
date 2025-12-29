-- Migration: Rename avatar_url to avatar_path and add avatar_thumbnail_path
-- This aligns the user table with the naming convention used in images and chat_images tables

-- 1. Create a new table with the updated schema
CREATE TABLE user_new (
    user_id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE CHECK (LENGTH(email) <= 100),
    nickname TEXT NOT NULL UNIQUE CHECK (LENGTH(nickname) <= 50),
    first_name TEXT NOT NULL CHECK (LENGTH(first_name) <= 100),
    last_name TEXT NOT NULL CHECK (LENGTH(last_name) <= 100),
    date_of_birth DATE NOT NULL,
    avatar_path TEXT,                    -- Changed from avatar_url
    avatar_thumbnail_path TEXT,          -- New field for consistency with other image tables
    about_me TEXT,
    gender TEXT CHECK (gender IN ('male','female','other','prefer_not_to_say')),
    is_private BOOLEAN NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 2. Copy existing data from the old table to the new one
-- Note: We map avatar_url to avatar_path, and avatar_thumbnail_path will be NULL initially
INSERT INTO user_new (
    user_id, email, nickname, first_name, last_name, 
    date_of_birth, avatar_path, avatar_thumbnail_path, about_me, gender, is_private, created_at
)
SELECT 
    user_id, email, nickname, first_name, last_name, 
    date_of_birth, avatar_url, NULL, about_me, gender, is_private, created_at
FROM user;

-- 3. Drop the old table
DROP TABLE user;

-- 4. Rename the new table to the original name
ALTER TABLE user_new RENAME TO user;