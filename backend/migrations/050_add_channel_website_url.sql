-- Add website_url field to channels table
ALTER TABLE channels ADD COLUMN IF NOT EXISTS website_url VARCHAR(500);
