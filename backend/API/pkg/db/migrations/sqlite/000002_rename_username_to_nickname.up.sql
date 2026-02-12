-- 1. Create a new table with the desired schema
CREATE TABLE user_new (
    user_id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE CHECK (LENGTH(email) <= 100),
    nickname TEXT NOT NULL UNIQUE CHECK (LENGTH(nickname) <= 50), -- Changed from username
    first_name TEXT NOT NULL CHECK (LENGTH(first_name) <= 100),
    last_name TEXT NOT NULL CHECK (LENGTH(last_name) <= 100),
    date_of_birth DATE NOT NULL,
    avatar_url TEXT,
    about_me TEXT,
    gender TEXT CHECK (gender IN ('male','female','other','prefer_not_to_say')),
    is_private BOOLEAN NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 2. Copy data from the old table to the new one
-- Note: We map the old 'username' values into the new 'nickname' column
INSERT INTO user_new (
    user_id, email, nickname, first_name, last_name, 
    date_of_birth, avatar_url, about_me, gender, is_private, created_at
)
SELECT 
    user_id, email, username, first_name, last_name, 
    date_of_birth, avatar_url, about_me, gender, is_private, created_at
FROM user;

-- 3. Drop the old table
DROP TABLE user;

-- 4. Rename the new table to the original name
ALTER TABLE user_new RENAME TO user;