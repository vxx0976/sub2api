-- +migrate Up
ALTER TABLE users ADD COLUMN IF NOT EXISTS register_domain VARCHAR(255) DEFAULT '';
