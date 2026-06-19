-- 分销商 M2M 服务令牌表
-- 用于分销商后端非交互式调用 key 管理接口（重置配额/启用-禁用/创建 key）。
-- 明文只在创建时返回一次，库内仅保存 SHA-256 哈希。

CREATE TABLE IF NOT EXISTS reseller_api_tokens (
    id           BIGSERIAL    PRIMARY KEY,
    reseller_id  BIGINT       NOT NULL,
    name         VARCHAR(100) NOT NULL DEFAULT '',
    token_prefix VARCHAR(20)  NOT NULL,
    token_hash   VARCHAR(64)  NOT NULL,
    status       VARCHAR(20)  NOT NULL DEFAULT 'active',
    last_used_at TIMESTAMPTZ,
    expires_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- 索引名与 ent 生成的名称保持一致，避免将来若启用 ent auto-migrate 时重复建索引。
CREATE UNIQUE INDEX IF NOT EXISTS resellerapitoken_token_hash
    ON reseller_api_tokens (token_hash);
CREATE INDEX IF NOT EXISTS resellerapitoken_reseller_id
    ON reseller_api_tokens (reseller_id);
