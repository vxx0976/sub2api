-- 把"定时上线时间窗口"从分组级别迁移到账号级别
-- 1) 在 accounts 表新增字段
ALTER TABLE accounts
    ADD COLUMN IF NOT EXISTS active_start_time VARCHAR(5),
    ADD COLUMN IF NOT EXISTS active_end_time   VARCHAR(5);

-- 2) 删除 groups 表的旧字段（无运行时使用方依赖，直接 DROP）
ALTER TABLE groups
    DROP COLUMN IF EXISTS active_start_time,
    DROP COLUMN IF EXISTS active_end_time;
