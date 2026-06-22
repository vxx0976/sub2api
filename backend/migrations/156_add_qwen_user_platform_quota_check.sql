-- 把 'qwen' 补进 user_platform_quotas.platform 的 CHECK 约束。
-- 150_relax_user_platform_quota_platform_check.sql 把白名单从 4 个放宽到 8 个
-- （anthropic/openai/gemini/antigravity + deepseek/moonshot/glm/seedance），
-- 但漏了后来新增的 qwen 平台。而 service.AllowedQuotaPlatforms（单一权威来源）
-- 与 ent schema 的 CHECK 均已包含 qwen，导致给 qwen 账号配置 per-user 配额时
-- ent + service 校验通过、但 INSERT 撞 DB CHECK 约束失败，功能静默不可用——
-- 与 150 注释描述的是同一类问题，这里对 qwen 再修一次。
--
-- 不在 150 内原地改（150 已发布、checksum 受防篡改校验保护，改动会导致已应用库启动报错），
-- 而是新增本迁移：DROP 旧约束、重建包含全部 9 个平台的 CHECK。
-- 约束名沿用 150 的 user_platform_quotas_platform_check 以便后续维护。
-- 平台列表保持与 backend/internal/service/domain_constants.go 的 AllowedQuotaPlatforms 一致。

-- 删除现有约束（150 或 142 内联列 CHECK 命名均为 user_platform_quotas_platform_check）。
ALTER TABLE user_platform_quotas
    DROP CONSTRAINT IF EXISTS user_platform_quotas_platform_check;

-- 加回包含全部 9 个平台的 CHECK。用 DO 块守卫，避免重复执行时报“约束已存在”。
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
                'deepseek', 'moonshot', 'glm', 'qwen', 'seedance'
            ));
    END IF;
END $$;
