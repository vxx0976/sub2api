-- 智能路由（故障转移）分组：虚拟分组持有有序成员列表和当前激活指针
ALTER TABLE groups ADD COLUMN IF NOT EXISTS is_failover_group BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE groups ADD COLUMN IF NOT EXISTS failover_member_ids JSONB;
ALTER TABLE groups ADD COLUMN IF NOT EXISTS failover_active_member_id BIGINT;
ALTER TABLE groups ADD COLUMN IF NOT EXISTS failover_active_version BIGINT NOT NULL DEFAULT 0;
ALTER TABLE groups ADD COLUMN IF NOT EXISTS failover_pin_member_id BIGINT;
ALTER TABLE groups ADD COLUMN IF NOT EXISTS failover_pin_expires_at TIMESTAMPTZ;

-- 用量日志：记录请求所经过的虚拟分组（NULL 表示非故障转移流量）
ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS requested_group_id BIGINT;

-- 仅对故障转移流量建部分索引，节省空间
CREATE INDEX IF NOT EXISTS usage_logs_requested_group_id_idx
  ON usage_logs(requested_group_id)
  WHERE requested_group_id IS NOT NULL;

-- 快速定位虚拟分组
CREATE INDEX IF NOT EXISTS groups_is_failover_idx
  ON groups(is_failover_group)
  WHERE is_failover_group = TRUE AND deleted_at IS NULL;
