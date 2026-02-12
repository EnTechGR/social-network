-- Migration 000011: Eliminate group_event_options
-- ============================================================================
-- The three RSVP labels (going / not going / maybe) are fixed for every event.
-- Storing them as per-event rows in group_event_options is pure redundancy:
--   • 3 unnecessary INSERTs on every event creation
--   • an extra JOIN on every detail / vote query
--   • meaningless UUIDs exposed through the API
--
-- This migration:
--   1. Recreates group_event_votes with a `choice` TEXT column (CHECK-constrained)
--      that replaces the old option_id FK.
--   2. Migrates existing vote data by resolving each option_id → its label.
--   3. Drops group_event_options entirely.
--   4. Adjusts indexes accordingly.
-- ============================================================================

-- ─── 1. New votes table with choice column ──────────────────────────────────
CREATE TABLE IF NOT EXISTS group_event_votes_new (
    event_id   TEXT      NOT NULL,
    user_id    TEXT      NOT NULL,
    choice     TEXT      NOT NULL CHECK (choice IN ('going', 'not going', 'maybe')),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (event_id, user_id),
    FOREIGN KEY (event_id) REFERENCES group_events(event_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id)  REFERENCES user(user_id)          ON DELETE CASCADE
);

-- ─── 2. Carry over existing votes, resolving option_id → label ───────────────
INSERT INTO group_event_votes_new (event_id, user_id, choice, created_at)
SELECT gev.event_id,
       gev.user_id,
       geo.label,          -- resolve the UUID to the human-readable label
       gev.created_at
FROM   group_event_votes gev
INNER JOIN group_event_options geo ON gev.option_id = geo.option_id;

-- ─── 3. Swap tables ─────────────────────────────────────────────────────────
DROP  TABLE group_event_votes;
ALTER TABLE group_event_votes_new RENAME TO group_event_votes;

-- ─── 4. Drop the now-unused options table ───────────────────────────────────
DROP TABLE IF EXISTS group_event_options;

-- ─── 5. Clean up stale indexes, create new ones ────────────────────────────
DROP INDEX IF EXISTS idx_group_event_options_event;   -- table gone
DROP INDEX IF EXISTS idx_group_event_votes_option;    -- column gone

CREATE INDEX IF NOT EXISTS idx_group_event_votes_event  ON group_event_votes(event_id);
CREATE INDEX IF NOT EXISTS idx_group_event_votes_user   ON group_event_votes(user_id);
CREATE INDEX IF NOT EXISTS idx_group_event_votes_choice ON group_event_votes(event_id, choice);

-- ============================================================================
-- Result:
--   group_event_votes  →  event_id | user_id | choice ('going'|…) | created_at
--   group_event_options →  GONE
-- ============================================================================