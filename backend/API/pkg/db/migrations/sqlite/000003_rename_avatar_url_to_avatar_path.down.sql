-- Migration Rollback: Revert avatar_path back to avatar_url and remove avatar_thumbnail_path
-- This reverts the changes made in 000003_rename_avatar_url_to_avatar_path_up.sql

-- 1. Create a table with the old schema
CREATE TABLE user_old (
    user_id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE CHECK (LENGTH(email) <= 100),
    nickname TEXT NOT NULL UNIQUE CHECK (LENGTH(nickname) <= 50),
    first_name TEXT NOT NULL CHECK (LENGTH(first_name) <= 100),
    last_name TEXT NOT NULL CHECK (LENGTH(last_name) <= 100),
    date_of_birth DATE NOT NULL,
    avatar_url TEXT,                     -- Reverted back from avatar_path
    about_me TEXT,
    gender TEXT CHECK (gender IN ('male','female','other','prefer_not_to_say')),
    is_private BOOLEAN NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 2. Copy data back, mapping avatar_path to avatar_url
-- Note: avatar_thumbnail_path will be lost in the rollback
INSERT INTO user_old (
    user_id, email, nickname, first_name, last_name, 
    date_of_birth, avatar_url, about_me, gender, is_private, created_at
)
SELECT 
    user_id, email, nickname, first_name, last_name, 
    date_of_birth, avatar_path, about_me, gender, is_private, created_at
FROM user;

-- 3. Drop the current table
DROP TABLE user;

-- 4. Rename the old table back to user
ALTER TABLE user_old RENAME TO user;