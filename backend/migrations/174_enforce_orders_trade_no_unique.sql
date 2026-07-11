-- AliMPay orders.trade_no 部分唯一索引：防同一笔支付宝账单被二次入账。
--
-- 背景：ConfirmOrderPaid 入账去重此前仅依赖 AlipayMonitor.matchedBills 进程内存，
-- 多实例部署 / 进程重启后该内存清零，同一笔账单（trade_no）可能被两个订单先后
-- 确认并各入一次账。对齐 usdt_orders.trade_no 的幂等约束（见 155_add_usdt_orders.sql），
-- 在 orders 表补一个 DB 侧唯一约束兜底。
--
-- 口径：仅约束「已入账」的订单（status='paid' 且 trade_no 非空），不误伤 pending/expired
-- 阶段 trade_no 为 NULL/'' 的订单，也不影响退款(refunded)后可能的历史遗留。
--
-- ⚠️ 上线前务必先在生产库确认无历史重复（本迁移下方 DO 块会自动 precheck 并在有重复时
--    RAISE EXCEPTION 阻断，给出可操作报错；也可手动先跑）：
--      SELECT trade_no, count(*) FROM orders
--       WHERE status='paid' AND trade_no IS NOT NULL AND trade_no <> ''
--       GROUP BY trade_no HAVING count(*) > 1;
--    若有结果，说明已发生二次入账，需人工核账 / 退回多入的余额后再放行本迁移。
--
-- 本文件是普通事务型迁移（DO $$ 内的 BEGIN 是 PL/pgSQL 匿名块语法，非事务控制），
-- CREATE UNIQUE INDEX 走事务；orders 为 AliMPay 充值订单表，数据量小，瞬时完成。

DO $$
DECLARE
    dup_count integer;
BEGIN
    SELECT count(*) INTO dup_count FROM (
        SELECT trade_no
        FROM orders
        WHERE status = 'paid'
          AND trade_no IS NOT NULL
          AND trade_no <> ''
        GROUP BY trade_no
        HAVING count(*) > 1
    ) d;

    IF dup_count > 0 THEN
        RAISE EXCEPTION
            'orders.trade_no has % duplicated paid trade_no(s); resolve double-credited orders before applying unique index (see 174_enforce_orders_trade_no_unique.sql)',
            dup_count;
    END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS order_trade_no_paid_unique
    ON orders (trade_no)
    WHERE status = 'paid' AND trade_no IS NOT NULL AND trade_no <> '';
