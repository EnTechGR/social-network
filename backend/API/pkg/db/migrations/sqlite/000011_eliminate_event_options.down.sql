-- Migration 000011 Rollback: restore group_event_options + option_id FK
-- ============================================================================
-- Reverses the up migration:
--   1. Recreates group_event_options with one row per (event, label).
--      UUIDs are generated in-place with randomblob — acceptable for a rollback.
--   2. Recreates group_event_votes with the original option_id FK, resolving
--      each choice label back to the newly-generated option_id.
--   3. Drops the choice-based votes table and restores indexes.
-- ============================================================================

-- ─── 1. Recreate group_event_options ─────────────────────────────────────────
CREATE TABLE IF NOT EXISTS group_event_options (
    option_id TEXT PRIMARY KEY,
    event_id  TEXT NOT NULL,
    label     TEXT NOT NULL CHECK (label IN ('going', 'not going', 'maybe')),
    FOREIGN KEY (event_id) REFERENCES group_events(event_id) ON DELETE CASCADE
);

-- Insert one row per (event, label).  hex(randomblob(16)) gives a unique 32-char
-- hex string — not RFC-4122 UUID format, but referentially sound for a rollback.
INSERT INTO group_event_options (option_id, event_id, label)
SELECT hex(randomblob(16)), event_id, 'going'     FROM group_events;

INSERT INTO group_event_options (option_id, event_id, label)
SELECT hex(randomblob(16)), event_id, 'not going' FROM group_events;

INSERT INTO group_event_options (option_id, event_id, label)
SELECT hex(randomblob(16)), event_id, 'maybe'     FROM group_events;

-- ─── 2. Recreate group_event_votes with option_id FK ────────────────────────
CREATE TABLE IF NOT EXISTS group_event_votes_old (
    event_id   TEXT      NOT NULL,
    user_id    TEXT      NOT NULL,
    option_id  TEXT      NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (event_id, user_id),
    FOREIGN KEY (event_id)  REFERENCES group_events(event_id)       ON DELETE CASCADE,
    FOREIGN KEY (user_id)   REFERENCES user(user_id)                ON DELETE CASCADE,
    FOREIGN KEY (option_id) REFERENCES group_event_options(option_id) ON DELETE CASCADE
);

-- Resolve choice label → option_id via the freshly-populated options table
INSERT INTO group_event_votes_old (event_id, user_id, option_id, created_at)
SELECT gev.event_id,
       gev.user_id,
       geo.option_id,
       gev.created_at
FROM   group_event_votes gev
INNER JOIN group_event_options geo
       ON  geo.event_id = gev.event_id
       AND geo.label    = gev.choice;

-- ─── 3. Swap tables ─────────────────────────────────────────────────────────
DROP  TABLE group_event_votes;
ALTER TABLE group_event_votes_old RENAME TO group_event_votes;

-- ─── 4. Restore indexes ─────────────────────────────────────────────────────
DROP INDEX IF EXISTS idx_group_event_votes_choice;   -- new index, remove it

CREATE INDEX IF NOT EXISTS idx_group_event_options_event  ON group_event_options(event_id);
CREATE INDEX IF NOT EXISTS idx_group_event_votes_event    ON group_event_votes(event_id);
CREATE INDEX IF NOT EXISTS idx_group_event_votes_user     ON group_event_votes(user_id);
CREATE INDEX IF NOT EXISTS idx_group_event_votes_option   ON group_event_votes(option_id);

-- ============================================================================
-- Rollback complete — schema is back to the 000010 state.
-- ============================================================================