-- Migration 000006: Add absolute session timeout
-- This adds absolute_expires_at column to prevent indefinite session lifetime
-- OWASP recommends limiting total session lifetime regardless of activity

-- SQLite doesn't support ALTER TABLE ADD COLUMN with constraints easily,
-- so we need to recreate the table

-- 1. Create new table with absolute_expires_at column
CREATE TABLE sessions_new (
    session_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL UNIQUE,
    csrf_token TEXT NOT NULL,
    ip_address TEXT,
    user_agent TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,                    -- Idle timeout
    absolute_expires_at TIMESTAMP NOT NULL,           -- NEW: Absolute timeout
    FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE
);

-- 2. Copy existing data, setting absolute_expires_at to 12 hours from created_at
INSERT INTO sessions_new (
    session_id, user_id, csrf_token, ip_address, user_agent, 
    created_at, expires_at, absolute_expires_at
)
SELECT 
    session_id, user_id, csrf_token, ip_address, user_agent,
    created_at, expires_at,
    -- Set absolute_expires_at to 12 hours from creation time
    datetime(created_at, '+12 hours') as absolute_expires_at
FROM sessions;

-- 3. Drop old table
DROP TABLE sessions;

-- 4. Rename new table
ALTER TABLE sessions_new RENAME TO sessions;

-- 5. Recreate index
CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);

-- 6. Create index on absolute_expires_at for cleanup queries
CREATE INDEX IF NOT EXISTS idx_sessions_absolute_expiry ON sessions(absolute_expires_at);