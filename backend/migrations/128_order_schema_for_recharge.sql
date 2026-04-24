-- orders 表：从"订阅/充值二合一"重构为"纯充值"系统（AliMPay 个人免签）
-- 1) 移除订阅关联：外键 + subscription_id 列
-- 2) 新增充值字段：credit_amount（到账余额）、multiplier（倍率快照）
-- 3) 新增审计字段：source_domain（下单来源域名；归属仍走 User.register_domain）
-- 说明：orders 表当前为空表（历史业务层代码已移除，schema 保留），无数据迁移风险。

ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_user_subscriptions_orders;
ALTER TABLE orders DROP COLUMN IF EXISTS subscription_id;

ALTER TABLE orders ADD COLUMN IF NOT EXISTS credit_amount decimal(10,2) NOT NULL DEFAULT 0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS multiplier    decimal(10,2) NOT NULL DEFAULT 1.0;
ALTER TABLE orders ADD COLUMN IF NOT EXISTS source_domain varchar(255);

ALTER TABLE orders ALTER COLUMN group_id DROP NOT NULL;

COMMENT ON COLUMN orders.credit_amount IS '实际到账余额（USD）';
COMMENT ON COLUMN orders.multiplier    IS '倍率快照';
COMMENT ON COLUMN orders.source_domain IS '下单时的来源域名（审计用，归属仍走 User.register_domain）';
