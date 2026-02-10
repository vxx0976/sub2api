-- Add Telegram chat ID to api_keys for user binding
ALTER TABLE api_keys ADD COLUMN IF NOT EXISTS tg_chat_id BIGINT;

-- Index for looking up keys by tg_chat_id
CREATE INDEX IF NOT EXISTS idx_api_keys_tg_chat_id ON api_keys (tg_chat_id) WHERE tg_chat_id IS NOT NULL;
