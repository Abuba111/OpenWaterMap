-- 001_create_tables.down.sql
-- Откат всех базовых таблиц

DROP TABLE IF EXISTS photos CASCADE;
DROP TABLE IF EXISTS comments CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS water_points CASCADE;
DROP TABLE IF EXISTS schema_migrations CASCADE;
