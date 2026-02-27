-- Migration 000014: Create comment_images relationship table
-- ============================================================================
-- Comments no longer store inline image columns. This table links images_core
-- records to comments so comment attachments can be uploaded and retrieved.
-- ============================================================================

CREATE TABLE IF NOT EXISTS comment_images (
    comment_image_id TEXT PRIMARY KEY,
    comment_id       TEXT      NOT NULL,
    image_id         TEXT      NOT NULL,
    display_order    INTEGER   NOT NULL DEFAULT 1,
    created_at       TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (comment_id) REFERENCES comments(comment_id) ON DELETE CASCADE,
    FOREIGN KEY (image_id)   REFERENCES images_core(image_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_comment_images_comment ON comment_images(comment_id);
CREATE INDEX IF NOT EXISTS idx_comment_images_image   ON comment_images(image_id);

DROP VIEW IF EXISTS v_comment_images;

CREATE VIEW IF NOT EXISTS v_comment_images AS
SELECT
    c.comment_id,
    c.post_id,
    ci.comment_image_id,
    ci.display_order,
    ic.image_id,
    ic.file_path,
    ic.thumbnail_path,
    ic.mime_type,
    ic.width,
    ic.height,
    ic.uploaded_at
FROM comments c
INNER JOIN comment_images ci ON c.comment_id = ci.comment_id
INNER JOIN images_core ic ON ci.image_id = ic.image_id
WHERE ic.deleted_at IS NULL
ORDER BY c.comment_id, ci.display_order;
