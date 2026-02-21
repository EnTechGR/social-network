-- Reintroduce group chat persistence for realtime room messaging.
CREATE TABLE IF NOT EXISTS group_messages (
    message_id TEXT PRIMARY KEY,
    group_id TEXT NOT NULL,
    sender_id TEXT NOT NULL,
    content TEXT NOT NULL CHECK (LENGTH(content) <= 1000),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES groups(group_id) ON DELETE CASCADE,
    FOREIGN KEY (sender_id) REFERENCES user(user_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_group_messages_group_created
    ON group_messages(group_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_group_messages_sender
    ON group_messages(sender_id);
