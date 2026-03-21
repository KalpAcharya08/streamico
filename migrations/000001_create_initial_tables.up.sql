-- =============================================================================
-- 📖 Migration: Create Initial Tables (MySQL)
-- =============================================================================
-- MySQL differences from PostgreSQL:
--   - No UUID type → use CHAR(36) with UUID() function
--   - No "SERIAL" → use AUTO_INCREMENT for integer IDs
--   - TIMESTAMP has different behavior → we use DATETIME
--   - No CREATE EXTENSION → UUID() is built into MySQL 8+
--   - Indexes must have explicit names
--   - TEXT columns can't have defaults in some MySQL versions
-- =============================================================================

-- ─── Users Table ─────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS users (
    id            CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    email         VARCHAR(255) NOT NULL UNIQUE,
    username      VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    avatar_url    TEXT,
    role          VARCHAR(20) NOT NULL DEFAULT 'viewer',
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_users_email (email),
    INDEX idx_users_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ─── Channels Table ──────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS channels (
    id               CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    owner_id         CHAR(36) NOT NULL,
    name             VARCHAR(255) NOT NULL UNIQUE,
    description      TEXT,
    banner_url       TEXT,
    subscriber_count BIGINT NOT NULL DEFAULT 0,
    created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_channels_owner (owner_id),
    CONSTRAINT fk_channels_owner FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ─── Videos Table ────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS videos (
    id               CHAR(36) PRIMARY KEY DEFAULT (UUID()),
    channel_id       CHAR(36) NOT NULL,
    title            VARCHAR(500) NOT NULL,
    description      TEXT,
    status           VARCHAR(30) NOT NULL DEFAULT 'uploading',
    duration_seconds INT DEFAULT 0,
    thumbnail_url    TEXT,
    storage_path     TEXT,
    view_count       BIGINT NOT NULL DEFAULT 0,
    visibility       VARCHAR(20) NOT NULL DEFAULT 'private',
    published_at     DATETIME NULL,
    created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_videos_channel (channel_id),
    INDEX idx_videos_status (status),
    INDEX idx_videos_visibility (visibility),
    CONSTRAINT fk_videos_channel FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
