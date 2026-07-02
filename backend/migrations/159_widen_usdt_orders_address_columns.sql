-- 放宽 usdt_orders 的地址列长度：64 → 128
--
-- 起因：TON 付款方地址是 raw 形态 0:<64位hex> = 66 字符，超过原 VARCHAR(64)。
-- 监控器扫到并匹配到链上转账后，MarkPaid 写 from_address 触发 ent 校验器
-- （"value is greater than the required length"），确认入账每轮失败，订单永不入账、
-- 最终过期。见 ent/schema/usdt_order.go 与 ent/migrate/schema.go（Size: 128）。
--
-- 列类型严格对齐 ent 生成结果（避免将来若启用 ent auto-migrate 时类型漂移）。
-- from_address / receiving_address 均无索引，ALTER TYPE 不触发索引重建；本表数据量极小，瞬时完成。
-- 幂等：仅当当前长度不为 128 时才 ALTER（IS DISTINCT FROM 处理 NULL；本表两列恒为 VARCHAR，不会出现 NULL 长度）。
-- 下面 DO $$ ... BEGIN ... END $$ 中的 BEGIN 是 PL/pgSQL 匿名块语法，不是事务控制语句，
-- 因此本文件是普通事务型迁移（由 runner 在单个事务内执行），切勿改名为 *_notx.sql。

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'usdt_orders'
          AND column_name = 'from_address'
          AND character_maximum_length IS DISTINCT FROM 128
    ) THEN
        ALTER TABLE usdt_orders ALTER COLUMN from_address TYPE VARCHAR(128);
    END IF;

    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'usdt_orders'
          AND column_name = 'receiving_address'
          AND character_maximum_length IS DISTINCT FROM 128
    ) THEN
        ALTER TABLE usdt_orders ALTER COLUMN receiving_address TYPE VARCHAR(128);
    END IF;
END $$;
