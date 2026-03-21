-- =============================================================================
-- 📖 Migration: DROP Initial Tables (Rollback)
-- =============================================================================
-- Drop in reverse dependency order:
--   videos → depends on channels → depends on users
-- =============================================================================

DROP TABLE IF EXISTS videos;
DROP TABLE IF EXISTS channels;
DROP TABLE IF EXISTS users;
