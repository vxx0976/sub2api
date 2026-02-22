-- +migrate Up
ALTER TABLE orders ADD COLUMN IF NOT EXISTS payment_amount DECIMAL(10,2);

-- +migrate Down
ALTER TABLE orders DROP COLUMN IF EXISTS payment_amount;
