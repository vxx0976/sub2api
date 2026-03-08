-- 商户代理系统：新增 reseller_withdrawals 表 + usage_logs 新增 merchant_rate_snapshot 字段

-- 1. 新增提现申请表
CREATE TABLE IF NOT EXISTS reseller_withdrawals (
    id               BIGSERIAL      PRIMARY KEY,
    reseller_id      BIGINT         NOT NULL,
    amount           DECIMAL(20,10) NOT NULL,
    status           VARCHAR(20)    NOT NULL DEFAULT 'pending',
    payment_method   VARCHAR(50)    NOT NULL,
    payment_account  VARCHAR(255)   NOT NULL,
    payment_name     VARCHAR(100)   NOT NULL,
    admin_notes      TEXT           NOT NULL DEFAULT '',
    admin_id         BIGINT,
    created_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    paid_at          TIMESTAMPTZ,
    rejected_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_reseller_withdrawals_reseller ON reseller_withdrawals(reseller_id);
CREATE INDEX IF NOT EXISTS idx_reseller_withdrawals_status   ON reseller_withdrawals(status);

-- 确保同一商户同时只有一笔 pending 提现（局部唯一索引）
CREATE UNIQUE INDEX IF NOT EXISTS idx_reseller_withdrawals_one_pending
    ON reseller_withdrawals(reseller_id)
    WHERE status = 'pending';

-- 2. usage_logs 新增 merchant_rate_snapshot 字段
ALTER TABLE usage_logs
    ADD COLUMN IF NOT EXISTS merchant_rate_snapshot DECIMAL(10,4) NULL;
