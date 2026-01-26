-- 048_add_channels.sql
-- Add channels table for balance query management

CREATE TABLE IF NOT EXISTS channels (
    id BIGSERIAL PRIMARY KEY,

    -- 基本信息
    name VARCHAR(100) NOT NULL,
    description TEXT,
    platform VARCHAR(50),
    status VARCHAR(20) NOT NULL DEFAULT 'active',

    -- 余额查询配置
    balance_url VARCHAR(500),
    balance_method VARCHAR(10) NOT NULL DEFAULT 'GET',
    balance_headers JSONB,
    balance_body TEXT,
    balance_path VARCHAR(200),
    balance_unit VARCHAR(20) NOT NULL DEFAULT '$',

    -- 缓存的余额信息
    cached_balance DECIMAL(20,6),
    last_check_at TIMESTAMPTZ,
    last_error TEXT,

    -- 时间戳
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_channels_name ON channels(name);
CREATE INDEX IF NOT EXISTS idx_channels_platform ON channels(platform) WHERE platform IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_channels_status ON channels(status);
CREATE INDEX IF NOT EXISTS idx_channels_deleted_at ON channels(deleted_at);

-- Add comments
COMMENT ON TABLE channels IS '渠道管理表，用于配置多个渠道并查询余额';
COMMENT ON COLUMN channels.name IS '渠道名称';
COMMENT ON COLUMN channels.description IS '渠道描述';
COMMENT ON COLUMN channels.platform IS '所属平台（anthropic/openai/gemini/other）';
COMMENT ON COLUMN channels.status IS '状态：active/inactive/error';
COMMENT ON COLUMN channels.balance_url IS '余额查询API URL';
COMMENT ON COLUMN channels.balance_method IS '请求方式：GET/POST';
COMMENT ON COLUMN channels.balance_headers IS '请求头（JSON格式，包含认证信息）';
COMMENT ON COLUMN channels.balance_body IS '请求体（POST时使用）';
COMMENT ON COLUMN channels.balance_path IS '响应中余额字段路径（如 data.balance）';
COMMENT ON COLUMN channels.balance_unit IS '余额单位';
COMMENT ON COLUMN channels.cached_balance IS '缓存的余额值';
COMMENT ON COLUMN channels.last_check_at IS '上次查询时间';
COMMENT ON COLUMN channels.last_error IS '上次查询错误';
