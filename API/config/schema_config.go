package config

const CreatePostCategoriesTable = `CREATE TABLE IF NOT EXISTS post_categories (
            post_id TEXT NOT NULL,
            category_id INTEGER NOT NULL,
            PRIMARY KEY (post_id, category_id),
            FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE,
            FOREIGN KEY (category_id) REFERENCES categories(category_id) ON DELETE CASCADE
        );`

const CreateUserTable = `CREATE TABLE IF NOT EXISTS user (
            user_id TEXT PRIMARY KEY,
            username TEXT NOT NULL UNIQUE CHECK (LENGTH(username) <= 50),
            email TEXT NOT NULL UNIQUE CHECK (LENGTH(email) <= 100),
            created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
        );`

const CreateUserAuthTable = `CREATE TABLE IF NOT EXISTS user_auth (
            user_id TEXT PRIMARY KEY,
            password_hash TEXT NOT NULL CHECK (LENGTH(password_hash) <= 255),
            FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE
        );`

const CreateSessionsTable = `CREATE TABLE IF NOT EXISTS sessions (
            user_id TEXT PRIMARY KEY,
            session_id TEXT NOT NULL UNIQUE,
            csrf_token TEXT NOT NULL,             -- ✅ new column
            ip_address TEXT,
            created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
            expires_at TIMESTAMP NOT NULL,
            FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE
        );`

const CreateCategoriesTable = `CREATE TABLE IF NOT EXISTS categories (
            category_id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL UNIQUE CHECK (LENGTH(name) <= 100)
        );`

const CreatePostsTable = `CREATE TABLE IF NOT EXISTS posts (
        post_id TEXT PRIMARY KEY,
        user_id TEXT NOT NULL,
        title TEXT NOT NULL CHECK (LENGTH(title) <= 200),
        content TEXT NOT NULL CHECK (LENGTH(content) <= 2000),
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP,
        FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE
    );`

const CreateCommentsTable = `CREATE TABLE IF NOT EXISTS comments (
            comment_id TEXT PRIMARY KEY,
            post_id TEXT NOT NULL,
            user_id TEXT NOT NULL,
            content TEXT NOT NULL CHECK (LENGTH(content) <= 1000),
            created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP,
            FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE,
            FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE
        );`

const CreateReactionsTable = `CREATE TABLE IF NOT EXISTS reactions (
            user_id TEXT NOT NULL,
            reaction_type INTEGER NOT NULL CHECK (reaction_type IN (1, 2, 3)),
            comment_id TEXT,
            post_id TEXT,
            created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
            PRIMARY KEY (user_id, comment_id, post_id),
            FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE,
            FOREIGN KEY (comment_id) REFERENCES comments(comment_id) ON DELETE CASCADE,
            FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE,
            CHECK (
                (post_id IS NULL AND comment_id IS NOT NULL) OR
                (post_id IS NOT NULL AND comment_id IS NULL)
            )
        );`

// CreateImagesTable stores image uploads linked to posts
const CreateImagesTable = `CREATE TABLE IF NOT EXISTS images (
        image_id TEXT PRIMARY KEY,
        post_id TEXT NOT NULL,
        user_id TEXT NOT NULL,
        file_path TEXT NOT NULL,
        thumbnail_path TEXT NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE,
        FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE
    );`

// -- OAuth providers table to store OAuth account information
const CreateOAuthTable = `CREATE TABLE IF NOT EXISTS oauth_accounts (
    oauth_id TEXT PRIMARY KEY,                    -- Unique identifier for this OAuth record
    user_id TEXT NOT NULL,                        -- Links to your existing user table
    provider TEXT NOT NULL CHECK (provider IN ('google', 'github', 'discord', 'facebook')), -- OAuth provider
    provider_user_id TEXT NOT NULL,               -- User ID from the OAuth provider
    provider_username TEXT,                       -- Username from provider (optional)
    provider_email TEXT,                          -- Email from provider
    provider_avatar_url TEXT,                     -- Avatar URL from provider (optional)
    access_token TEXT,                            -- OAuth access token (encrypted in production)
    refresh_token TEXT,                           -- OAuth refresh token (encrypted in production)
    token_expires_at TIMESTAMP,                   -- When the access token expires
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    
    -- Composite unique constraint to prevent duplicate provider accounts
    UNIQUE(provider, provider_user_id),
    
    -- Foreign key to link with existing user
    FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE
);`
