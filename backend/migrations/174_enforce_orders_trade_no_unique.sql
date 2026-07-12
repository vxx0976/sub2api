-- AliMPay orders.trade_no 部分唯一索引：防同一笔第三方支付账单被二次入账。
--
-- 背景：ConfirmOrderPaid 入账去重此前仅依赖 AlipayMonitor.matchedBills 进程内存，
-- 多实例部署 / 进程重启后该内存清零，同一笔账单（trade_no）可能被两个订单先后
-- 确认并各入一次账。对齐 usdt_orders.trade_no 的幂等约束（见 155_add_usdt_orders.sql），
-- 在 orders 表补一个 DB 侧唯一约束兜底。
--
-- 口径：仅约束「真实第三方支付渠道（alipay/wxpay）的已入账订单」，且排除占位/哨兵
-- trade_no。生产存在两类不能纳入唯一约束的合法数据：
--   1) 管理员手工确认的订单共用占位 trade_no = 'manual'（多条合法，非二次入账）；
--   2) pay_type='manual_adjust' 等非第三方渠道单。
-- 若不排除，(a) 下方 precheck 会被 'manual'×N 误判为重复而 RAISE 阻断，(b) 即便清洗
-- 历史，索引生效后也会打断未来的手工确认。故谓词按 pay_type 白名单 + trade_no 非
-- 占位来界定，pending/expired 阶段 trade_no 为 NULL/'' 的订单同样不受约束。
--
-- ⚠️ 上线前务必先确认无历史二次入账（下方 DO 块自动 precheck 并在有重复时 RAISE
--    EXCEPTION，报错含冲突 trade_no 与订单号，便于精确核账；也可手动先跑）：
--      SELECT trade_no, count(*), string_agg(order_no || '(u' || user_id || ')', ', ')
--        FROM orders
--       WHERE status='paid' AND pay_type IN ('alipay','wxpay')
--         AND trade_no IS NOT NULL AND trade_no NOT IN ('manual','')
--       GROUP BY trade_no HAVING count(*) > 1;
--    若有结果，说明已发生二次入账，需人工核账：保留真实付款订单、把幻影订单
--    status 置为终态非 paid（如 voided），并退回/核销多入的余额后再放行本迁移。
--
-- 本文件是普通事务型迁移（DO $$ 内的 BEGIN 是 PL/pgSQL 匿名块语法，非事务控制），
-- CREATE UNIQUE INDEX 走事务；orders 为充值订单表，数据量小，瞬时完成。

DO $$
DECLARE
    dup_detail text;
BEGIN
    SELECT string_agg(trade_no || ' x' || cnt || ' [' || orders_list || ']', '; ')
      INTO dup_detail
      FROM (
        SELECT trade_no,
               count(*) AS cnt,
               string_agg(order_no || '(u' || user_id || ')', ', ' ORDER BY created_at) AS orders_list
        FROM orders
        WHERE status = 'paid'
          AND pay_type IN ('alipay', 'wxpay')
          AND trade_no IS NOT NULL
          AND trade_no NOT IN ('manual', '')
        GROUP BY trade_no
        HAVING count(*) > 1
      ) d;

    IF dup_detail IS NOT NULL THEN
        RAISE EXCEPTION
            'orders.trade_no has duplicated paid third-party bill(s); resolve double-credited orders before applying unique index (keep the real payment, void the phantom(s), reconcile balance). Offending: %',
            dup_detail;
    END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS order_trade_no_paid_unique
    ON orders (trade_no)
    WHERE status = 'paid'
      AND pay_type IN ('alipay', 'wxpay')
      AND trade_no IS NOT NULL
      AND trade_no NOT IN ('manual', '');
