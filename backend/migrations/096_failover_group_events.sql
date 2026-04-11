-- 故障转移事件日志：仅追加表，记录每次 active_id 切换
CREATE TABLE IF NOT EXISTS failover_group_events (
    id               BIGSERIAL PRIMARY KEY,
    virtual_group_id BIGINT NOT NULL,
    from_member_id   BIGINT,
    to_member_id     BIGINT,
    reason           VARCHAR(40) NOT NULL,
    triggered_by     BIGINT,
    note             TEXT,
    occurred_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS failover_group_events_group_occurred_idx
  ON failover_group_events(virtual_group_id, occurred_at DESC);

CREATE INDEX IF NOT EXISTS failover_group_events_occurred_idx
  ON failover_group_events(occurred_at);
