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
    oauth_id TEXT PRIMARY KEY,                                                                  -- Unique identifier for this OAuth record
    user_id TEXT NOT NULL,                                                                      -- Links to your existing user table
    provider TEXT NOT NULL CHECK (provider IN ('google', 'github', 'discord', 'facebook')),     -- OAuth provider
    provider_user_id TEXT NOT NULL,                                                             -- User ID from the OAuth provider
    provider_username TEXT,                                                                     -- Username from provider (optional)
    provider_email TEXT,                                                                        -- Email from provider
    provider_avatar_url TEXT,                                                                   -- Avatar URL from provider (optional)
    access_token TEXT,                                                                          -- OAuth access token (encrypted in production)
    refresh_token TEXT,                                                                         -- OAuth refresh token (encrypted in production)
    token_expires_at TIMESTAMP,                                                                 -- When the access token expires
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    UNIQUE(provider, provider_user_id),
    FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE
);`

/**
 * @name messages
 * @description Stores all private message records exchanged between users.
 * This table uses CASCADE on deletion for both sender and receiver
 * to ensure message integrity if a user account is removed.
 */
const CreateMessagesTable = `CREATE TABLE IF NOT EXISTS messages (
    message_id TEXT PRIMARY KEY,                                            -- Unique identifier for the message (e.g., a UUID).
    sender_id TEXT NOT NULL,                                                -- ID of the user who sent the message.
    receiver_id TEXT NOT NULL,                                              -- ID of the user intended to receive the message.
    content TEXT NOT NULL CHECK (LENGTH(content) <= 1000),                  -- The message text, limited to 1000 characters to prevent abuse/excessive storage.
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,                -- Timestamp when the message was recorded.
    is_read BOOLEAN NOT NULL DEFAULT 0,                                     -- Status indicating if the receiver has viewed the message (0=unread, 1=read).
    FOREIGN KEY (sender_id) REFERENCES user(user_id) ON DELETE CASCADE,     -- Ensures sender exists and handles cleanup if sender is deleted.
    FOREIGN KEY (receiver_id) REFERENCES user(user_id) ON DELETE CASCADE,   -- Ensures receiver exists and handles cleanup if receiver is deleted.
    CHECK (sender_id != receiver_id)                                        -- Constraint: A user cannot send a message to themselves.
);`


/**
 * @name chat_images
 * @description Stores metadata for image files attached to chat messages.
 * This table is designed for scalability by separating large file metadata
 * from the core 'messages' table. It supports a constrained set of common image types.
 */
const CreateChatImagesTable = `CREATE TABLE IF NOT EXISTS chat_images (
    image_id TEXT PRIMARY KEY,                                                                              -- Unique identifier for the image record.
    message_id TEXT NOT NULL,                                                                               -- The message this image is attached to (links back to the 'messages' table).
    user_id TEXT NOT NULL,                                                                                  -- The user who uploaded the image (should match messages.sender_id).
    filename TEXT NOT NULL,                                                                                 -- The unique, server-generated name of the stored file (e.g., a UUID).
    original_filename TEXT NOT NULL,                                                                        -- The name the user originally gave the file.
    file_path TEXT NOT NULL,                                                                                -- Full path to the original, high-resolution image file on the storage system.
    thumbnail_path TEXT NOT NULL,                                                                           -- Path to the smaller, optimized thumbnail for quick loading in chat previews.
    file_size INTEGER NOT NULL CHECK (file_size > 0),                                                       -- Size of the original file in bytes. Must be positive.
    mime_type TEXT NOT NULL CHECK (mime_type IN ('image/jpeg', 'image/png', 'image/gif', 'image/webp')),    -- Restricted list of allowed image MIME types.
    width INTEGER NOT NULL CHECK (width > 0),                                                               -- Width of the image in pixels.
    height INTEGER NOT NULL CHECK (height > 0),                                                             -- Height of the image in pixels.
    uploaded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,                                               -- Timestamp when the image record was created.
    FOREIGN KEY (message_id) REFERENCES messages(message_id) ON DELETE CASCADE,                             -- If the message is deleted, its associated image metadata is also removed.
    FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE                                        -- Ensures the uploader exists and handles cleanup if user is deleted.
);`