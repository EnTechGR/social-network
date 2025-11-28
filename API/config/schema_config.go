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
            first_name TEXT NOT NULL CHECK (LENGTH(first_name) <= 100),
            last_name TEXT NOT NULL CHECK (LENGTH(last_name) <= 100),
            age INTEGER NOT NULL CHECK (age >= 13 AND age <= 120),
            gender TEXT NOT NULL CHECK (gender IN ('male', 'female', 'other', 'prefer_not_to_say')),
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
        title TEXT CHECK (LENGTH(title) <= 200),
        content TEXT CHECK (LENGTH(content) <= 2000),
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP,
        FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE
    );`

const CreateCommentsTable = `CREATE TABLE IF NOT EXISTS comments (
            comment_id TEXT PRIMARY KEY,
            post_id TEXT NOT NULL,
            user_id TEXT NOT NULL,
            content TEXT CHECK (LENGTH(content) <= 1000),
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

const CreateNotificationsTable = `CREATE TABLE IF NOT EXISTS notifications (
        notification_id TEXT PRIMARY KEY,
        user_id TEXT NOT NULL,
        from_user_id TEXT NOT NULL,
        type TEXT NOT NULL,
        post_id TEXT NOT NULL,
        comment_id TEXT,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        is_read BOOLEAN NOT NULL DEFAULT 0,
        is_visible BOOLEAN NOT NULL DEFAULT 1,
        FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE,
        FOREIGN KEY (from_user_id) REFERENCES user(user_id) ON DELETE CASCADE,
        FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE,
        FOREIGN KEY (comment_id) REFERENCES comments(comment_id) ON DELETE CASCADE
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
    UNIQUE(provider, provider_user_id),
    FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE
);`

// CreateMessagesTable stores private messages between users
const CreateMessagesTable = `CREATE TABLE IF NOT EXISTS messages (
    message_id TEXT PRIMARY KEY,
    sender_id TEXT NOT NULL,
    receiver_id TEXT NOT NULL,
    content TEXT NOT NULL CHECK (LENGTH(content) <= 1000),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_read BOOLEAN NOT NULL DEFAULT 0,
    FOREIGN KEY (sender_id) REFERENCES user(user_id) ON DELETE CASCADE,
    FOREIGN KEY (receiver_id) REFERENCES user(user_id) ON DELETE CASCADE,
    CHECK (sender_id != receiver_id)
);`

// CreateChatImagesTable stores images sent through chat messages
const CreateChatImagesTable = `CREATE TABLE IF NOT EXISTS chat_images (
    image_id TEXT PRIMARY KEY,
    message_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    filename TEXT NOT NULL,
    original_filename TEXT NOT NULL,
    file_path TEXT NOT NULL,
    thumbnail_path TEXT NOT NULL,
    file_size INTEGER NOT NULL CHECK (file_size > 0),
    mime_type TEXT NOT NULL CHECK (mime_type IN ('image/jpeg', 'image/png', 'image/gif', 'image/webp')),
    width INTEGER NOT NULL CHECK (width > 0),
    height INTEGER NOT NULL CHECK (height > 0),
    uploaded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (message_id) REFERENCES messages(message_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE
);`