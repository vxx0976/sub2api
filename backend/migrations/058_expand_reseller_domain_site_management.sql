-- Expand reseller_domains for site management features
-- Adds homepage template, purchase page, per-site locale, SEO, login redirect

ALTER TABLE reseller_domains ADD COLUMN IF NOT EXISTS home_template VARCHAR(50) DEFAULT '' NOT NULL;
ALTER TABLE reseller_domains ADD COLUMN IF NOT EXISTS purchase_enabled BOOLEAN DEFAULT false NOT NULL;
ALTER TABLE reseller_domains ADD COLUMN IF NOT EXISTS purchase_url VARCHAR(500) DEFAULT '' NOT NULL;
ALTER TABLE reseller_domains ADD COLUMN IF NOT EXISTS default_locale VARCHAR(20) DEFAULT '' NOT NULL;
ALTER TABLE reseller_domains ADD COLUMN IF NOT EXISTS seo_title VARCHAR(255) DEFAULT '' NOT NULL;
ALTER TABLE reseller_domains ADD COLUMN IF NOT EXISTS seo_description TEXT DEFAULT '' NOT NULL;
ALTER TABLE reseller_domains ADD COLUMN IF NOT EXISTS seo_keywords VARCHAR(500) DEFAULT '' NOT NULL;
ALTER TABLE reseller_domains ADD COLUMN IF NOT EXISTS login_redirect VARCHAR(255) DEFAULT '' NOT NULL;
