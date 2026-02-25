-- Fix recharge_orders time columns: TIMESTAMP -> TIMESTAMPTZ
-- The original migration (046) used TIMESTAMP (without timezone), causing
-- lib/pq to misinterpret CST values as UTC when reading back.
-- PostgreSQL will interpret existing values using the session timezone
-- (Asia/Shanghai), so existing data is preserved correctly.

ALTER TABLE recharge_orders
    ALTER COLUMN paid_at      TYPE TIMESTAMPTZ USING paid_at AT TIME ZONE current_setting('timezone'),
    ALTER COLUMN created_at   TYPE TIMESTAMPTZ USING created_at AT TIME ZONE current_setting('timezone'),
    ALTER COLUMN updated_at   TYPE TIMESTAMPTZ USING updated_at AT TIME ZONE current_setting('timezone'),
    ALTER COLUMN expired_at   TYPE TIMESTAMPTZ USING expired_at AT TIME ZONE current_setting('timezone');
