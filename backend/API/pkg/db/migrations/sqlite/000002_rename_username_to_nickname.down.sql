CREATE TABLE user_old (
    user_id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE CHECK (LENGTH(email) <= 100),
    username TEXT NOT NULL UNIQUE CHECK (LENGTH(username) <= 50),
    first_name TEXT NOT NULL CHECK (LENGTH(first_name) <= 100),
    last_name TEXT NOT NULL CHECK (LENGTH(last_name) <= 100),
    date_of_birth DATE NOT NULL,
    avatar_url TEXT,
    nickname TEXT UNIQUE, -- Putting the original optional nickname back
    about_me TEXT,
    gender TEXT CHECK (gender IN ('male','female','other','prefer_not_to_say')),
    is_private BOOLEAN NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO user_old (
    user_id, email, username, first_name, last_name, 
    date_of_birth, avatar_url, nickname, about_me, gender, is_private, created_at
)
SELECT 
    user_id, email, nickname, first_name, last_name, 
    date_of_birth, avatar_url, nickname, about_me, gender, is_private, created_at
FROM user;

DROP TABLE user;

ALTER TABLE user_old RENAME TO user;