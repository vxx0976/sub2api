-- Add platform_cost_snapshot to usage_logs for correct commission calculation.
-- Records the merchant's platform cost at the time of each request.
ALTER TABLE usage_logs ADD COLUMN IF NOT EXISTS platform_cost_snapshot DECIMAL(10,4);
