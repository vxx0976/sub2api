-- +migrate Up
ALTER TABLE users ADD COLUMN IF NOT EXISTS register_domain VARCHAR(255) DEFAULT '';

-- +migrate Down
ALTER TABLE users DROP COLUMN IF EXISTS register_domain;
