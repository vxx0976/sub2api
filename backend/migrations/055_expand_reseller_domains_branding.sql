-- 055: Expand reseller_domains with per-domain branding fields
-- Move subtitle, home_content, doc_url from reseller_settings (global) to reseller_domains (per-domain)

ALTER TABLE reseller_domains ADD COLUMN IF NOT EXISTS subtitle VARCHAR(255) DEFAULT '' NOT NULL;
ALTER TABLE reseller_domains ADD COLUMN IF NOT EXISTS home_content TEXT DEFAULT '' NOT NULL;
ALTER TABLE reseller_domains ADD COLUMN IF NOT EXISTS doc_url VARCHAR(500) DEFAULT '' NOT NULL;
