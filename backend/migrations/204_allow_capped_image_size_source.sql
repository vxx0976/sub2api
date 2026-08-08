-- 允许 image_size_source = 'capped'。
-- 上游忽略请求里显式指定的图片尺寸（如 ChatGPT OAuth 把 size 归一成 auto 后恒定出
-- ~1.57MP 的图）时，实际出图档位会高于客户请求的档位；此时按请求档封顶计费，
-- 并用 'capped' 标记该笔计费的尺寸来源，便于与 'output'（如实按出图档计费）区分。

ALTER TABLE usage_logs
    DROP CONSTRAINT IF EXISTS usage_logs_image_size_source_check;

ALTER TABLE usage_logs
    ADD CONSTRAINT usage_logs_image_size_source_check
    CHECK (
        image_size_source IS NULL
        OR image_size_source IN ('output', 'input', 'default', 'legacy', 'capped')
    ) NOT VALID;
