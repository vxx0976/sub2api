-- 放宽 user_platform_quotas.platform 的 CHECK 约束。
-- 142 建表时 CHECK 只允许 anthropic/openai/gemini/antigravity 四个平台，
-- 但 service.AllowedQuotaPlatforms（单一权威来源）与 ent schema 已支持 8 个平台
-- （新增 deepseek/moonshot/glm/seedance）。导致给后四者配置 per-user 配额时
-- ent + service 校验通过、但 INSERT 撞 DB CHECK 约束失败，功能静默不可用。
-- 这里删除旧约束、加回包含全部 8 个平台的新约束，并以稳定的约束名命名便于后续维护。
-- 平台列表保持与 backend/internal/service/domain_constants.go 的 AllowedQuotaPlatforms 一致。

-- 删除 142 内联 CHECK 自动生成的约束（Postgres 内联列 CHECK 命名为 <表>_<列>_check）。
ALTER TABLE user_platform_quotas
    DROP CONSTRAINT IF EXISTS user_platform_quotas_platform_check;

-- 加回包含全部 8 个平台的 CHECK。用 DO 块守卫，避免重复执行时报“约束已存在”。
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'user_platform_quotas_platform_check'
          AND conrelid = 'user_platform_quotas'::regclass
    ) THEN
        ALTER TABLE user_platform_quotas
            ADD CONSTRAINT user_platform_quotas_platform_check
            CHECK (platform IN (
                'anthropic', 'openai', 'gemini', 'antigravity',
                'deepseek', 'moonshot', 'glm', 'seedance'
            ));
    END IF;
END $$;
