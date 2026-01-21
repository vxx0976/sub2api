-- 1. 创建充值订单表
CREATE TABLE recharge_orders (
    id BIGSERIAL PRIMARY KEY,
    order_no VARCHAR(50) UNIQUE NOT NULL,
    trade_no VARCHAR(100),
    user_id BIGINT NOT NULL REFERENCES users(id),

    -- 支付金额（用户实际支付的金额）
    amount DECIMAL(10,2) NOT NULL,

    -- 到账金额（实际充入余额 = amount × multiplier）
    credit_amount DECIMAL(10,2) NOT NULL,

    -- 倍率快照（记录创建订单时的倍率，防止配置修改影响历史订单）
    multiplier DECIMAL(10,2) NOT NULL DEFAULT 1.0,

    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    pay_type VARCHAR(20),
    paid_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expired_at TIMESTAMP
);

CREATE INDEX idx_recharge_orders_user_id ON recharge_orders(user_id);
CREATE INDEX idx_recharge_orders_status ON recharge_orders(status);
CREATE INDEX idx_recharge_orders_order_no ON recharge_orders(order_no);

COMMENT ON TABLE recharge_orders IS '充值订单表';
COMMENT ON COLUMN recharge_orders.amount IS '用户实际支付金额';
COMMENT ON COLUMN recharge_orders.credit_amount IS '实际到账余额 = amount × multiplier';
COMMENT ON COLUMN recharge_orders.multiplier IS '充值倍率快照';

-- 2. 插入阶梯倍率配置到 settings 表
INSERT INTO settings (key, value, updated_at)
VALUES
  (
    'recharge_tiers',
    '[{"min":10,"max":49.99,"multiplier":1.0},{"min":50,"max":99.99,"multiplier":1.2},{"min":100,"max":499.99,"multiplier":1.5},{"min":500,"max":null,"multiplier":2.0}]',
    NOW()
  ),
  (
    'recharge_enabled',
    'true',
    NOW()
  ),
  (
    'recharge_min_amount',
    '10',
    NOW()
  ),
  (
    'recharge_max_amount',
    '10000',
    NOW()
  )
ON CONFLICT (key) DO NOTHING;
