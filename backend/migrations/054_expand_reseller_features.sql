-- 054: Expand reseller features
-- Adds owner_id to groups, announcements, redeem_codes for reseller ownership
-- Adds reseller_template flag to groups
-- Creates reseller_settings table for per-reseller key-value settings

-- 1A. groups: add owner_id (NULL=admin, non-NULL=reseller)
ALTER TABLE groups ADD COLUMN IF NOT EXISTS owner_id BIGINT;
ALTER TABLE groups ADD COLUMN IF NOT EXISTS source_group_id BIGINT;
ALTER TABLE groups ADD COLUMN IF NOT EXISTS reseller_template BOOLEAN NOT NULL DEFAULT FALSE;
CREATE INDEX IF NOT EXISTS idx_groups_owner_id ON groups(owner_id);

-- 1B. announcements: add owner_id
ALTER TABLE announcements ADD COLUMN IF NOT EXISTS owner_id BIGINT;
CREATE INDEX IF NOT EXISTS idx_announcements_owner_id ON announcements(owner_id);

-- 1C. redeem_codes: add owner_id
ALTER TABLE redeem_codes ADD COLUMN IF NOT EXISTS owner_id BIGINT;
CREATE INDEX IF NOT EXISTS idx_redeem_codes_owner_id ON redeem_codes(owner_id);

-- 1D. reseller_settings table
CREATE TABLE IF NOT EXISTS reseller_settings (
    id            BIGSERIAL PRIMARY KEY,
    reseller_id   BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key           VARCHAR(100) NOT NULL,
    value         TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_reseller_settings_reseller_key
    ON reseller_settings(reseller_id, key);
