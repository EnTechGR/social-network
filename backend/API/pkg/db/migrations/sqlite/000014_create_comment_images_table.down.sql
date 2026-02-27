-- Migration 000014 Rollback: Drop comment_images relationship table

DROP VIEW IF EXISTS v_comment_images;

DROP INDEX IF EXISTS idx_comment_images_comment;
DROP INDEX IF EXISTS idx_comment_images_image;

DROP TABLE IF EXISTS comment_images;
