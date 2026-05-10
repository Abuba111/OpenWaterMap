-- 002_add_soft_delete.up.sql
-- Добавить soft delete для точек и комментариев

ALTER TABLE water_points
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

ALTER TABLE comments
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_wp_deleted_at       ON water_points(deleted_at);
CREATE INDEX IF NOT EXISTS idx_comments_deleted_at ON comments(deleted_at);
