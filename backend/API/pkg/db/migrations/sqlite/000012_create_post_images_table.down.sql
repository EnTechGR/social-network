-- Migration 000012 Rollback: Drop post_images relationship table

DROP VIEW IF EXISTS v_post_images;

DROP INDEX IF EXISTS idx_post_images_post;
DROP INDEX IF EXISTS idx_post_images_image;

DROP TABLE IF EXISTS post_images;