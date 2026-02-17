-- Migration 000012: Create post_images relationship table
-- ============================================================================
-- The post_images table was referenced in migration 000005's down script and
-- in the v_post_images view (migration 000007), but was never actually created
-- in any up migration. This left UploadPostImages writing to images_core with
-- no link back to the post entity.
-- ============================================================================

-- ─── 1. Create the post_images relationship table ───────────────────────────
CREATE TABLE IF NOT EXISTS post_images (
    post_image_id  TEXT      PRIMARY KEY,
    post_id        TEXT      NOT NULL,
    image_id       TEXT      NOT NULL,
    display_order  INTEGER   NOT NULL DEFAULT 1,
    created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (post_id)  REFERENCES posts(post_id)       ON DELETE CASCADE,
    FOREIGN KEY (image_id) REFERENCES images_core(image_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_post_images_post  ON post_images(post_id);
CREATE INDEX IF NOT EXISTS idx_post_images_image ON post_images(image_id);

-- ─── 2. Recreate v_post_images view (was broken without the table) ───────────
DROP VIEW IF EXISTS v_post_images;

CREATE VIEW IF NOT EXISTS v_post_images AS
SELECT
    p.post_id,
    p.user_id          AS post_author_id,
    pi.post_image_id,
    pi.display_order,
    ic.image_id,
    ic.file_path,
    ic.thumbnail_path,
    ic.mime_type,
    ic.width,
    ic.height,
    ic.uploaded_at
FROM posts p
INNER JOIN post_images pi ON p.post_id = pi.post_id
INNER JOIN images_core ic ON pi.image_id = ic.image_id
ORDER BY p.post_id, pi.display_order;

-- ============================================================================
-- Result:
--   post_images  →  post_image_id | post_id | image_id | display_order | created_at
--   v_post_images → fully functional join view
-- ============================================================================