-- 商户域名小写化存量回填。
-- CreateDomain/GetByDomain 与域名识别中间件均按小写精确匹配，但历史数据可能存有
-- 大小写混合的 domain（早期创建路径未归一）：这些行在服务端永远匹配不到（品牌/归属
-- 失效），且不阻止同一主机名以小写重复注册，造成一个主机名两行并存。
-- 回填策略（保证每步都不会撞 idx_reseller_domains_domain 唯一索引）：
--   1. 硬删与其它行小写同名的"已软删"行（与 CreateDomain 前置的 Purge 语义一致）；
--   2. 同一小写域名仍有多行 live 的（混合大小写并存），保留 verified 优先、其次最新
--      的一行，其余软删；
--   3. 对小写化后不与任何其它行冲突的行执行 domain = lower(domain)。

-- 1. 清理小写同名的软删行
DELETE FROM reseller_domains a
WHERE a.deleted_at IS NOT NULL
  AND EXISTS (
      SELECT 1 FROM reseller_domains b
      WHERE b.id <> a.id AND lower(b.domain) = lower(a.domain)
  );

-- 2. live 重复行去重：保留 verified 优先、id 最大（最新）的一行
WITH ranked AS (
    SELECT id,
           ROW_NUMBER() OVER (
               PARTITION BY lower(domain)
               ORDER BY verified DESC, id DESC
           ) AS rn
    FROM reseller_domains
    WHERE deleted_at IS NULL
)
UPDATE reseller_domains rd
SET deleted_at = NOW(), updated_at = NOW()
FROM ranked
WHERE rd.id = ranked.id AND ranked.rn > 1;

-- 3. 安全小写化（仅当小写化后在全表内唯一）
UPDATE reseller_domains rd
SET domain = lower(rd.domain), updated_at = NOW()
WHERE rd.domain <> lower(rd.domain)
  AND NOT EXISTS (
      SELECT 1 FROM reseller_domains o
      WHERE o.id <> rd.id AND lower(o.domain) = lower(rd.domain)
  );
