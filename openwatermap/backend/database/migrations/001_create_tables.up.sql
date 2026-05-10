-- 001_create_tables.up.sql
-- Создание всех базовых таблиц

CREATE TABLE IF NOT EXISTS water_points (
    id            SERIAL PRIMARY KEY,
    name          TEXT    NOT NULL,
    lat           REAL    NOT NULL,
    lng           REAL    NOT NULL,
    ph            REAL    NOT NULL,
    turbidity     REAL    NOT NULL DEFAULT 0,
    chlorine      REAL    NOT NULL DEFAULT 0,
    tds           REAL    NOT NULL DEFAULT 0,
    status        TEXT    NOT NULL CHECK(status IN ('good','warning','danger')),
    source        TEXT    NOT NULL DEFAULT 'Неизвестно',
    checked_at    TEXT    NOT NULL,
    review_status TEXT    NOT NULL DEFAULT 'approved'
                  CHECK(review_status IN ('pending','approved','rejected')),
    submitted_by  INTEGER,
    reviewed_by   INTEGER,
    reject_reason TEXT    NOT NULL DEFAULT '',
    deleted_at    TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_wp_status        ON water_points(status);
CREATE INDEX IF NOT EXISTS idx_wp_lat           ON water_points(lat);
CREATE INDEX IF NOT EXISTS idx_wp_lng           ON water_points(lng);
CREATE INDEX IF NOT EXISTS idx_wp_review_status ON water_points(review_status);
CREATE INDEX IF NOT EXISTS idx_wp_deleted_at    ON water_points(deleted_at);
CREATE INDEX IF NOT EXISTS idx_wp_name          ON water_points USING gin(to_tsvector('russian', name));

CREATE TABLE IF NOT EXISTS users (
    id            SERIAL PRIMARY KEY,
    name          TEXT    NOT NULL,
    email         TEXT    NOT NULL UNIQUE,
    password_hash TEXT    NOT NULL,
    role          TEXT    NOT NULL DEFAULT 'user'
                  CHECK(role IN ('admin','dl1','dl2','user')),
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

CREATE TABLE IF NOT EXISTS comments (
    id         SERIAL PRIMARY KEY,
    point_id   INTEGER NOT NULL REFERENCES water_points(id) ON DELETE CASCADE,
    user_id    INTEGER NOT NULL REFERENCES users(id),
    text       TEXT    NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_comments_point ON comments(point_id);

CREATE TABLE IF NOT EXISTS photos (
    id         SERIAL PRIMARY KEY,
    point_id   INTEGER NOT NULL REFERENCES water_points(id) ON DELETE CASCADE,
    user_id    INTEGER NOT NULL REFERENCES users(id),
    type       TEXT    NOT NULL DEFAULT 'photo'
               CHECK(type IN ('photo','certificate')),
    file_name  TEXT    NOT NULL,
    file_path  TEXT    NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_photos_point ON photos(point_id);

-- Таблица версий миграций (служебная)
CREATE TABLE IF NOT EXISTS schema_migrations (
    version    BIGINT  NOT NULL PRIMARY KEY,
    dirty      BOOLEAN NOT NULL DEFAULT FALSE,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
