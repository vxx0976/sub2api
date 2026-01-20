-- +migrate Up
ALTER TABLE groups ADD COLUMN IF NOT EXISTS is_recommended BOOLEAN DEFAULT FALSE;

-- +migrate Down
ALTER TABLE groups DROP COLUMN IF EXISTS is_recommended;
