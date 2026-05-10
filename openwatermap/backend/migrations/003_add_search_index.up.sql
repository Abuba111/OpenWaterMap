-- 003_add_search_index.up.sql
-- Добавить индексы для поиска

CREATE INDEX IF NOT EXISTS idx_wp_name_search
    ON water_points USING gin(to_tsvector('russian', name));

CREATE INDEX IF NOT EXISTS idx_wp_source_search
    ON water_points USING gin(to_tsvector('russian', source));
