-- 002_add_soft_delete.down.sql

ALTER TABLE water_points DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE comments     DROP COLUMN IF EXISTS deleted_at;

DROP INDEX IF EXISTS idx_wp_deleted_at;
DROP INDEX IF EXISTS idx_comments_deleted_at;
