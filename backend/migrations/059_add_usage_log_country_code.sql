-- Add country_code field to usage_logs for geo-distribution analytics
ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS country_code VARCHAR(2);

-- Composite index for geo-distribution queries (country + time range)
CREATE INDEX IF NOT EXISTS idx_usage_logs_country_code_created_at ON usage_logs(country_code, created_at);
