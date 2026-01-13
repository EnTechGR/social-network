-- Migration: Add UNIQUE constraint to sessions.user_id
-- This allows ON CONFLICT(user_id) to work properly in session creation
-- Ensures only one active session per user at a time

-- 1. Create a new table with the updated schema
CREATE TABLE sessions_new (
    session_id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL UNIQUE,           -- Added UNIQUE constraint
    csrf_token TEXT NOT NULL,
    ip_address TEXT,
    user_agent TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE
);

-- 2. Copy existing data from the old table
-- If there are duplicate user_ids, only the most recent session is kept
INSERT INTO sessions_new (session_id, user_id, csrf_token, ip_address, user_agent, created_at, expires_at)
SELECT session_id, user_id, csrf_token, ip_address, user_agent, created_at, expires_at
FROM sessions
WHERE session_id IN (
    SELECT session_id
    FROM sessions s1
    WHERE created_at = (
        SELECT MAX(created_at)
        FROM sessions s2
        WHERE s2.user_id = s1.user_id
    )
);

-- 3. Drop the old table
DROP TABLE sessions;

-- 4. Rename the new table to the original name
ALTER TABLE sessions_new RENAME TO sessions;

-- 5. Recreate the index
CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);