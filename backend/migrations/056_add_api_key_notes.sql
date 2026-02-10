-- Add notes field to api_keys table for customer identification
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS notes VARCHAR(500) NOT NULL DEFAULT '';
