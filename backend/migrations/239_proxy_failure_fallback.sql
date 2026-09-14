-- proxies: 故障回退（与到期回退 fallback_mode 相互独立）+ 健康状态
-- 代理健康检查连续失败达到阈值后标记 degraded，并按 failure_fallback_mode 把账号改投备用代理/直连；
-- 连续恢复后标记 healthy，并把仍处于回退状态的账号改回原代理。
ALTER TABLE proxies ADD COLUMN IF NOT EXISTS failure_fallback_mode varchar(20) NOT NULL DEFAULT 'none';
ALTER TABLE proxies ADD COLUMN IF NOT EXISTS failure_backup_proxy_id BIGINT REFERENCES proxies(id) ON DELETE SET NULL;
ALTER TABLE proxies ADD COLUMN IF NOT EXISTS health_status varchar(20) NOT NULL DEFAULT 'healthy';
ALTER TABLE proxies ADD COLUMN IF NOT EXISTS health_changed_at timestamptz;
ALTER TABLE proxies ADD COLUMN IF NOT EXISTS health_last_error varchar(500);
CREATE INDEX IF NOT EXISTS proxies_failure_backup_proxy_id_idx ON proxies (failure_backup_proxy_id);
-- 连续失败/成功计数落库：多实例轮流拿健康检查 leader 锁，内存计数会被拆散。
ALTER TABLE proxies ADD COLUMN IF NOT EXISTS health_fail_streak INT NOT NULL DEFAULT 0;
ALTER TABLE proxies ADD COLUMN IF NOT EXISTS health_ok_streak INT NOT NULL DEFAULT 0;
-- accounts: 故障回退来源（与到期回退/手动恢复用的 proxy_fallback_origin_id 分开记录）。
-- 故障改投时记最初所在代理，该代理恢复后按此列改回；到期回退与故障回退叠加时互不覆盖。
ALTER TABLE accounts ADD COLUMN IF NOT EXISTS proxy_failure_origin_id BIGINT;
CREATE INDEX IF NOT EXISTS accounts_proxy_failure_origin_id_idx ON accounts (proxy_failure_origin_id);
