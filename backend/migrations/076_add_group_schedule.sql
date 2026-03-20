-- 为分组添加定时上线时间窗口字段
-- active_start_time 和 active_end_time 均为 HH:MM 格式，两者都设置时生效

ALTER TABLE groups
    ADD COLUMN IF NOT EXISTS active_start_time VARCHAR(5),
    ADD COLUMN IF NOT EXISTS active_end_time   VARCHAR(5);
