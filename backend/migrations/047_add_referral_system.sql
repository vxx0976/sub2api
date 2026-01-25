-- 047_add_referral_system.sql
-- Add referral/invitation reward system

-- Add referral fields to users table
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS referral_code VARCHAR(8) UNIQUE,
    ADD COLUMN IF NOT EXISTS referred_by BIGINT,
    ADD COLUMN IF NOT EXISTS referral_rewarded BOOLEAN NOT NULL DEFAULT FALSE;

-- Create index on referral_code for fast lookups
CREATE INDEX IF NOT EXISTS idx_users_referral_code ON users(referral_code) WHERE referral_code IS NOT NULL;

-- Create index on referred_by for querying invitees
CREATE INDEX IF NOT EXISTS idx_users_referred_by ON users(referred_by) WHERE referred_by IS NOT NULL;

-- Create referral_rewards table to track reward distributions
CREATE TABLE IF NOT EXISTS referral_rewards (
    id BIGSERIAL PRIMARY KEY,
    referrer_id BIGINT NOT NULL REFERENCES users(id),
    invitee_id BIGINT NOT NULL UNIQUE REFERENCES users(id),
    trigger_order_id BIGINT NOT NULL REFERENCES orders(id),
    referrer_reward DECIMAL(10, 2) NOT NULL DEFAULT 0,
    invitee_reward DECIMAL(10, 2) NOT NULL DEFAULT 0,
    skip_referrer_reason VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for referral_rewards
CREATE INDEX IF NOT EXISTS idx_referral_rewards_referrer_id ON referral_rewards(referrer_id);
CREATE INDEX IF NOT EXISTS idx_referral_rewards_invitee_id ON referral_rewards(invitee_id);
CREATE INDEX IF NOT EXISTS idx_referral_rewards_trigger_order_id ON referral_rewards(trigger_order_id);
CREATE INDEX IF NOT EXISTS idx_referral_rewards_created_at ON referral_rewards(created_at);

-- Add comment
COMMENT ON TABLE referral_rewards IS 'Records of referral rewards distributed when invitees make their first qualifying payment';
COMMENT ON COLUMN users.referral_code IS 'User unique referral code (format: R + 7 characters)';
COMMENT ON COLUMN users.referred_by IS 'ID of the user who referred this user';
COMMENT ON COLUMN users.referral_rewarded IS 'Whether referral reward has been distributed for this user';
