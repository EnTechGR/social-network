-- Migration 000006 ROLLBACK: Remove absolute session timeout
-- This reverts the changes made in 000006_add_absolute_session_timeout.up.sql

-- 1. Create table with old schema (without absolute_expires_at)
CREATE TABLE sessions_old (
    session_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL UNIQUE,
    csrf_token TEXT NOT NULL,
    ip_address TEXT,
    user_agent TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE
);

-- 2. Copy data back (dropping absolute_expires_at column)
INSERT INTO sessions_old (
    session_id, user_id, csrf_token, ip_address, user_agent,
    created_at, expires_at
)
SELECT 
    session_id, user_id, csrf_token, ip_address, user_agent,
    created_at, expires_at
FROM sessions;

-- 3. Drop current table
DROP TABLE sessions;

-- 4. Rename old table back
ALTER TABLE sessions_old RENAME TO sessions;

-- 5. Recreate the original index
CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);