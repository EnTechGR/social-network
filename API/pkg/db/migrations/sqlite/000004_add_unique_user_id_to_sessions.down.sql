-- Migration Rollback: Remove UNIQUE constraint from sessions.user_id
-- This reverts the changes made in 000004_add_unique_user_id_to_sessions_up.sql

-- 1. Create a table with the old schema (without UNIQUE on user_id)
CREATE TABLE sessions_old (
    session_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,                  -- UNIQUE constraint removed
    csrf_token TEXT NOT NULL,
    ip_address TEXT,
    user_agent TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE
);

-- 2. Copy all data back
INSERT INTO sessions_old (session_id, user_id, csrf_token, ip_address, user_agent, created_at, expires_at)
SELECT session_id, user_id, csrf_token, ip_address, user_agent, created_at, expires_at
FROM sessions;

-- 3. Drop the current table
DROP TABLE sessions;

-- 4. Rename the old table back to sessions
ALTER TABLE sessions_old RENAME TO sessions;

-- 5. Recreate the index
CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);