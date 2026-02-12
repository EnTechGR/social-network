-- Migration 000010: Redesign Groups Schema (Simplified - Requirements Only)
-- This migration creates a clean, minimal schema that matches EXACTLY the requirements
-- No extra features, no over-engineering

-- ============================================================================
-- STEP 1: Drop unnecessary tables from original schema
-- ============================================================================

-- Drop group chat tables (not in requirements)
DROP TABLE IF EXISTS group_message_images;
DROP TABLE IF EXISTS group_messages;
DROP INDEX IF EXISTS idx_group_messages_group;
DROP INDEX IF EXISTS idx_group_messages_sender;

-- ============================================================================
-- STEP 2: Create simplified groups table
-- ============================================================================
-- Requirements: "title and a description"
-- Removed: is_public (all groups are browsable)

CREATE TABLE IF NOT EXISTS groups_new (
    group_id TEXT PRIMARY KEY,
    owner_id TEXT NOT NULL,
    title TEXT NOT NULL CHECK (LENGTH(title) <= 100),
    description TEXT CHECK (LENGTH(description) <= 500),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_id) REFERENCES user(user_id) ON DELETE CASCADE
);

-- Migrate existing groups data
INSERT INTO groups_new (group_id, owner_id, title, description, created_at)
SELECT group_id, owner_id, name, description, created_at
FROM groups;

-- Drop old table and rename
DROP TABLE IF EXISTS groups;
ALTER TABLE groups_new RENAME TO groups;

-- ============================================================================
-- STEP 3: Create simplified group_members table
-- ============================================================================
-- Requirements: Just track membership, no roles
-- Owner is identified by groups.owner_id, not by a role field

CREATE TABLE IF NOT EXISTS group_members_new (
    group_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (group_id, user_id),
    FOREIGN KEY (group_id) REFERENCES groups(group_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE
);

-- Migrate existing active members only (ignoring role)
INSERT INTO group_members_new (group_id, user_id, joined_at)
SELECT group_id, user_id, created_at
FROM group_members
WHERE status = 'active';

-- Drop old table and rename
DROP TABLE IF EXISTS group_members;
ALTER TABLE group_members_new RENAME TO group_members;

-- ============================================================================
-- STEP 4: Create group_invites table
-- ============================================================================
-- Requirements: "invite other users to join the group"
-- "invited users need to accept the invitation"

CREATE TABLE IF NOT EXISTS group_invites_new (
    invite_id TEXT PRIMARY KEY,
    group_id TEXT NOT NULL,
    from_user_id TEXT NOT NULL,
    to_user_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','accepted','declined')),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    responded_at TIMESTAMP,
    UNIQUE(group_id, to_user_id, status),
    FOREIGN KEY (group_id) REFERENCES groups(group_id) ON DELETE CASCADE,
    FOREIGN KEY (from_user_id) REFERENCES user(user_id) ON DELETE CASCADE,
    FOREIGN KEY (to_user_id) REFERENCES user(user_id) ON DELETE CASCADE
);

-- Migrate existing invites
INSERT OR IGNORE INTO group_invites_new (invite_id, group_id, from_user_id, to_user_id, status, created_at, responded_at)
SELECT invite_id, group_id, from_user_id, to_user_id, status, created_at, NULL
FROM group_invites;

-- Drop old table and rename
DROP TABLE IF EXISTS group_invites;
ALTER TABLE group_invites_new RENAME TO group_invites;

-- ============================================================================
-- STEP 5: Create group_join_requests table
-- ============================================================================
-- Requirements: "request to be in it"
-- "only the creator of the group would be allowed to accept or refuse"

CREATE TABLE IF NOT EXISTS group_join_requests_new (
    request_id TEXT PRIMARY KEY,
    group_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','approved','denied')),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    responded_at TIMESTAMP,
    UNIQUE(group_id, user_id, status),
    FOREIGN KEY (group_id) REFERENCES groups(group_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES user(user_id) ON DELETE CASCADE
);

-- Migrate existing requests
INSERT OR IGNORE INTO group_join_requests_new (request_id, group_id, user_id, status, created_at, responded_at)
SELECT request_id, group_id, user_id, status, created_at, NULL
FROM group_join_requests;

-- Drop old table and rename
DROP TABLE IF EXISTS group_join_requests;
ALTER TABLE group_join_requests_new RENAME TO group_join_requests;

-- ============================================================================
-- STEP 6: Events tables are already correct
-- ============================================================================
-- group_events: "Title, Description, Day/Time" ✓
-- group_event_options: "Going, Not going" ✓
-- group_event_votes: Tracks which user chose which option ✓

-- No changes needed for events tables, they match requirements perfectly

-- ============================================================================
-- STEP 7: Create indexes for performance
-- ============================================================================

-- Groups indexes
CREATE INDEX IF NOT EXISTS idx_groups_owner ON groups(owner_id);
CREATE INDEX IF NOT EXISTS idx_groups_created ON groups(created_at DESC);

-- Group members indexes
CREATE INDEX IF NOT EXISTS idx_group_members_group ON group_members(group_id);
CREATE INDEX IF NOT EXISTS idx_group_members_user ON group_members(user_id);

-- Group invites indexes
CREATE INDEX IF NOT EXISTS idx_group_invites_group ON group_invites(group_id);
CREATE INDEX IF NOT EXISTS idx_group_invites_to_user ON group_invites(to_user_id, status);
CREATE INDEX IF NOT EXISTS idx_group_invites_from_user ON group_invites(from_user_id);

-- Group join requests indexes
CREATE INDEX IF NOT EXISTS idx_group_join_requests_group ON group_join_requests(group_id, status);
CREATE INDEX IF NOT EXISTS idx_group_join_requests_user ON group_join_requests(user_id, status);

-- Group events indexes  
CREATE INDEX IF NOT EXISTS idx_group_events_group ON group_events(group_id);
CREATE INDEX IF NOT EXISTS idx_group_events_creator ON group_events(creator_id);
CREATE INDEX IF NOT EXISTS idx_group_events_time ON group_events(event_time);

-- Group event options indexes
CREATE INDEX IF NOT EXISTS idx_group_event_options_event ON group_event_options(event_id);

-- Group event votes indexes (THIS is how we track who's going/not going)
CREATE INDEX IF NOT EXISTS idx_group_event_votes_event ON group_event_votes(event_id);
CREATE INDEX IF NOT EXISTS idx_group_event_votes_user ON group_event_votes(user_id);
CREATE INDEX IF NOT EXISTS idx_group_event_votes_option ON group_event_votes(option_id);

-- ============================================================================
-- Migration Complete
-- ============================================================================

-- Summary of simplified schema:
-- ✅ groups: group_id, owner_id, title, description, created_at
-- ✅ group_members: group_id, user_id, joined_at (no roles)
-- ✅ group_invites: invite tracking with status
-- ✅ group_join_requests: join request tracking with status
-- ✅ group_events: event details
-- ✅ group_event_options: going, not going, maybe
-- ✅ group_event_votes: tracks WHO chose WHICH option (attendance tracking)
-- ✅ Removed group_messages and group_message_images (not in requirements)
-- ✅ Removed is_public (all groups browsable)
-- ✅ Removed role field (owner checked via owner_id)