-- fork：把旧的分组多语言两列（093_group_i18n.sql 的 name_i18n / description_i18n，
-- 形如 {"<locale>": "<text>"}）回填进上游 242_group_i18n.sql 的 groups.i18n
-- （形如 {"<locale>": {"name": ..., "description": ...}}）。
--
-- 合并上游 0.2.16 时改用上游的分组多语言实现，代码不再读写旧两列；旧两列保留不删，
-- 便于回滚核对。
--
-- 规则：
--   - 只回填 i18n 仍为空对象的分组（不覆盖上游新界面已编辑过的译文；可重复执行）；
--   - 语言代码取两列键的并集；空白译文丢弃，某语言两项都为空则整条丢弃；
--   - 旧列不是 JSON 对象（NULL / 'null' 等）时视为无译文。

UPDATE groups AS g
SET i18n = src.i18n
FROM (
    SELECT t.id, jsonb_object_agg(t.locale, t.entry) AS i18n
    FROM (
        SELECT
            g2.id,
            k.locale,
            jsonb_strip_nulls(jsonb_build_object(
                'name', NULLIF(btrim(
                    CASE WHEN jsonb_typeof(g2.name_i18n) = 'object' THEN g2.name_i18n ->> k.locale END
                ), ''),
                'description', NULLIF(btrim(
                    CASE WHEN jsonb_typeof(g2.description_i18n) = 'object' THEN g2.description_i18n ->> k.locale END
                ), '')
            )) AS entry
        FROM groups AS g2
        CROSS JOIN LATERAL (
            SELECT jsonb_object_keys(g2.name_i18n) AS locale
            WHERE jsonb_typeof(g2.name_i18n) = 'object'
            UNION
            SELECT jsonb_object_keys(g2.description_i18n)
            WHERE jsonb_typeof(g2.description_i18n) = 'object'
        ) AS k
    ) AS t
    WHERE t.entry <> '{}'::jsonb AND btrim(t.locale) <> ''
    GROUP BY t.id
) AS src
WHERE g.id = src.id
  AND g.i18n = '{}'::jsonb;
