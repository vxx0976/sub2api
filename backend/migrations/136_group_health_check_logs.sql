-- 分组健康检查历史日志：每次探测后写入一条记录，用于公开状态页按天聚合可用性。
-- 明细只保留约 31 天，由后台清理任务分批物理删（日志类无恢复需求，不用软删）。

CREATE TABLE IF NOT EXISTS group_health_check_logs (
    id         BIGSERIAL PRIMARY KEY,
    group_id   BIGINT      NOT NULL,
    status     VARCHAR(20) NOT NULL,
    checked_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_group_health_check_logs_group_checked
    ON group_health_check_logs(group_id, checked_at DESC);

CREATE INDEX IF NOT EXISTS idx_group_health_check_logs_checked_at
    ON group_health_check_logs(checked_at);
