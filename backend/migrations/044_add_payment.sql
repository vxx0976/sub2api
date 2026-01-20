-- +migrate Up

-- Extend groups table with payment-related fields
-- Note: validity_days uses existing default_validity_days column
ALTER TABLE groups ADD COLUMN IF NOT EXISTS price DECIMAL(10,2);
ALTER TABLE groups ADD COLUMN IF NOT EXISTS is_purchasable BOOLEAN DEFAULT FALSE;
ALTER TABLE groups ADD COLUMN IF NOT EXISTS sort_order INT DEFAULT 0;

-- Orders table for payment tracking
CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    order_no VARCHAR(64) NOT NULL UNIQUE,          -- Merchant order number
    trade_no VARCHAR(64),                          -- Payment platform order number
    user_id BIGINT NOT NULL REFERENCES users(id),
    group_id BIGINT NOT NULL REFERENCES groups(id),
    amount DECIMAL(10,2) NOT NULL,                 -- Order amount
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending/paid/expired/refunded
    pay_type VARCHAR(20),                          -- alipay/wxpay
    paid_at TIMESTAMPTZ,                           -- Payment time
    subscription_id BIGINT REFERENCES user_subscriptions(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expired_at TIMESTAMPTZ                         -- Order expiration time
);

CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_order_no ON orders(order_no);
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at);

-- +migrate Down
DROP TABLE IF EXISTS orders;
ALTER TABLE groups DROP COLUMN IF EXISTS price;
ALTER TABLE groups DROP COLUMN IF EXISTS is_purchasable;
ALTER TABLE groups DROP COLUMN IF EXISTS sort_order;
