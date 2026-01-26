-- 049_add_channel_icon.sql
-- Add icon_url column to channels table

ALTER TABLE channels ADD COLUMN IF NOT EXISTS icon_url VARCHAR(500);

COMMENT ON COLUMN channels.icon_url IS '渠道图标URL';
