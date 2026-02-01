-- Add external_buy_url field to groups table
-- This allows admins to configure an external purchase link (e.g., Taobao)
ALTER TABLE groups ADD COLUMN IF NOT EXISTS external_buy_url TEXT;
