-- 恢复 user_platform_quotas.platform 的 CHECK 约束为「全部 10 个平台」。
--
-- 背景：上游迁移 157_user_platform_quotas_add_grok.sql 把约束 DROP 后重建为仅
-- 5 个平台（anthropic/openai/gemini/antigravity/grok），无意中丢掉了本 fork 经
-- 150（+deepseek/moonshot/glm/seedance）和 156（+qwen）放进白名单的 5 个国产平台。
-- 由于 migrations_runner 按文件名 sort.Strings 顺序执行，157 是这条约束的终态，
-- 导致：service.AllowedQuotaPlatforms / ent schema 校验均放行 deepseek 等国产平台，
-- 但 INSERT 撞 DB CHECK → 管理员配 per-user 国产平台配额 500、自助注册写国产默认
-- 配额行被 fail-open 静默丢弃。与 150/156 注释描述的是同一类问题，这里再合并修一次。
--
-- 不在 157 内原地改（157 已发布、checksum 受防篡改校验保护，改动会导致已应用库启动报错），
-- 而是新增本迁移：DROP 旧约束、重建包含全部 10 个平台（150+156 的 9 个国产/通用 ∪ 157 的 grok）。
-- 约束名沿用 user_platform_quotas_platform_check。
-- 平台列表保持与 backend/internal/service/domain_constants.go 的 AllowedQuotaPlatforms 一致。

ALTER TABLE user_platform_quotas
    DROP CONSTRAINT IF EXISTS user_platform_quotas_platform_check;

-- 加回包含全部 10 个平台的 CHECK。用 DO 块守卫，避免重复执行时报“约束已存在”。
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
                'deepseek', 'moonshot', 'glm', 'qwen', 'seedance', 'grok'
            ));
    END IF;
END $$;
