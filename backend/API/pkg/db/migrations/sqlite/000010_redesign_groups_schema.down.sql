-- Migration 000010 Rollback: Restore original groups schema

-- ============================================================================
-- STEP 1: Drop simplified indexes
-- ============================================================================

DROP INDEX IF EXISTS idx_groups_created;
DROP INDEX IF EXISTS idx_group_invites_to_user;
DROP INDEX IF EXISTS idx_group_invites_from_user;
DROP INDEX IF EXISTS idx_group_join_requests_group;
DROP INDEX IF EXISTS idx_group_join_requests_user;
DROP INDEX IF EXISTS idx_group_events_creator;
DROP INDEX IF EXISTS idx_group_events_time;
DROP INDEX IF EXISTS idx_group_event_options_event;
DROP INDEX IF EXISTS idx_group_event_votes_event;
DROP INDEX IF EXISTS idx_group_event_votes_user;
DROP INDEX IF EXISTS idx_group_event_votes_option;

-- ============================================================================
-- STEP 2: Restore original groups table (with name and is_public)
-- ============================================================================

CREATE TABLE IF NOT EXISTS groups_old (
    group_id TEXT PRIMARY KEY,
    owner_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    is_public BOOLEAN NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_id) REFERENCES user(user_id) ON DELETE CASCADE
);

INSERT INTO groups_old (group_id, owner_id, name, description, is_public, created_at)
SELECT group_id, owner_id, title, description, 0, created_at
FROM groups;

DROP TABLE groups;
ALTER TABLE groups_old RENAME TO groups;

CREATE INDEX IF NOT EXISTS idx_groups_public ON groups(is_public);

-- ============================================================================
-- STEP 3: Restore original group_members table (with role and status)
-- ============================================================================

CREATE TABLE IF NOT EXISTS group_members_old (
    group_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('owner','admin','member')),
    status TEXT NOT NULL CHECK (status IN ('active','invited','requested')),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (group_id, user_id),
    FOREIGN KEY (group_id) REFERENCES groups(group_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE
);

-- Restore members with default role='member' and status='active'
INSERT INTO group_members_old (group_id, user_id, role, status, created_at)
SELECT 
    gm.group_id, 
    gm.user_id, 
    CASE WHEN g.owner_id = gm.user_id THEN 'owner' ELSE 'member' END as role,
    'active' as status,
    gm.joined_at
FROM group_members gm
INNER JOIN groups g ON gm.group_id = g.group_id;

DROP TABLE group_members;
ALTER TABLE group_members_old RENAME TO group_members;

CREATE INDEX IF NOT EXISTS idx_group_members_status ON group_members(group_id, status);
CREATE INDEX IF NOT EXISTS idx_group_members_user ON group_members(user_id);

-- ============================================================================
-- STEP 4: Restore original group_invites table (without responded_at)
-- ============================================================================

CREATE TABLE IF NOT EXISTS group_invites_old (
    invite_id TEXT PRIMARY KEY,
    group_id TEXT NOT NULL,
    from_user_id TEXT NOT NULL,
    to_user_id TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending','accepted','declined')),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES groups(group_id) ON DELETE CASCADE,
    FOREIGN KEY (from_user_id) REFERENCES user(user_id) ON DELETE CASCADE,
    FOREIGN KEY (to_user_id) REFERENCES user(user_id) ON DELETE CASCADE
);

INSERT INTO group_invites_old (invite_id, group_id, from_user_id, to_user_id, status, created_at)
SELECT invite_id, group_id, from_user_id, to_user_id, status, created_at
FROM group_invites;

DROP TABLE group_invites;
ALTER TABLE group_invites_old RENAME TO group_invites;

-- ============================================================================
-- STEP 5: Restore original group_join_requests table
-- ============================================================================

CREATE TABLE IF NOT EXISTS group_join_requests_old (
    request_id TEXT PRIMARY KEY,
    group_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending','approved','denied')),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (group_id) REFERENCES groups(group_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE
);

INSERT INTO group_join_requests_old (request_id, group_id, user_id, status, created_at)
SELECT request_id, group_id, user_id, status, created_at
FROM group_join_requests;

DROP TABLE group_join_requests;
ALTER TABLE group_join_requests_old RENAME TO group_join_requests;

-- ============================================================================
-- STEP 6: Restore group_messages and group_message_images tables
-- ============================================================================

CREATE TABLE IF NOT EXISTS group_messages (
    message_id TEXT PRIMARY KEY,
    group_id TEXT NOT NULL,
    sender_id TEXT NOT NULL,
    content TEXT CHECK (LENGTH(content) <= 1000),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    is_read BOOLEAN NOT NULL DEFAULT 0,
    FOREIGN KEY (group_id) REFERENCES groups(group_id) ON DELETE CASCADE,
    FOREIGN KEY (sender_id) REFERENCES user(user_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS group_message_images (
    image_id TEXT PRIMARY KEY,
    message_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    filename TEXT NOT NULL,
    original_filename TEXT NOT NULL,
    file_path TEXT NOT NULL,
    thumbnail_path TEXT NOT NULL,
    file_size INTEGER NOT NULL CHECK (file_size > 0),
    mime_type TEXT NOT NULL CHECK (mime_type IN ('image/jpeg','image/png','image/gif','image/webp')),
    width INTEGER NOT NULL CHECK (width > 0),
    height INTEGER NOT NULL CHECK (height > 0),
    uploaded_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (message_id) REFERENCES group_messages(message_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE
);

-- ============================================================================
-- STEP 7: Restore original indexes
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_group_messages_group ON group_messages(group_id);
CREATE INDEX IF NOT EXISTS idx_group_messages_sender ON group_messages(sender_id);
CREATE INDEX IF NOT EXISTS idx_group_events_group ON group_events(group_id);

-- ============================================================================
-- Rollback Complete
-- ============================================================================