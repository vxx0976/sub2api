-- Add reseller support: parent_id, token_version, role_version fields on users table
-- and create reseller_domains table

-- 1. Add parent_id column (FK to users.id, nullable, ON DELETE SET NULL)
ALTER TABLE users ADD COLUMN IF NOT EXISTS parent_id BIGINT;
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_users_parent_id'
    ) THEN
        ALTER TABLE users ADD CONSTRAINT fk_users_parent_id
            FOREIGN KEY (parent_id) REFERENCES users(id) ON DELETE SET NULL;
    END IF;
END
$$;
CREATE INDEX IF NOT EXISTS idx_users_parent_id ON users(parent_id);

-- 2. Add token_version column (persists JWT invalidation on password change)
ALTER TABLE users ADD COLUMN IF NOT EXISTS token_version BIGINT NOT NULL DEFAULT 0;

-- 3. Add role_version column (persists JWT invalidation on role change)
ALTER TABLE users ADD COLUMN IF NOT EXISTS role_version BIGINT NOT NULL DEFAULT 0;

-- 4. Create reseller_domains table
CREATE TABLE IF NOT EXISTS reseller_domains (
    id            BIGSERIAL PRIMARY KEY,
    reseller_id   BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    domain        VARCHAR(255) NOT NULL,
    site_name     VARCHAR(100) NOT NULL DEFAULT '',
    site_logo     TEXT NOT NULL DEFAULT '',
    brand_color   VARCHAR(20) NOT NULL DEFAULT '',
    custom_css    TEXT NOT NULL DEFAULT '',
    verify_token  VARCHAR(64) NOT NULL DEFAULT '',
    verified      BOOLEAN NOT NULL DEFAULT FALSE,
    verified_at   TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_reseller_domains_domain ON reseller_domains(domain);
CREATE INDEX IF NOT EXISTS idx_reseller_domains_reseller_id ON reseller_domains(reseller_id);
CREATE INDEX IF NOT EXISTS idx_reseller_domains_verified ON reseller_domains(verified);
