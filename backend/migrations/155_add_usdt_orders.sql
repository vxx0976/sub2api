-- USDT (TRC20) 收款订单表
-- 与 AliMPay 的 orders 表并列：CNY 计价（amount/credit_amount，1:1 入账），
-- usdt_amount 是按下单冻结汇率换算的、带唯一尾数的链上应付金额（6 位小数，独立列避免 decimal(10,2) 精度坑）。
-- 无 webhook，由 UsdtMonitor 轮询 TronGrid 按 (chain, usdt_amount) 唯一金额匹配；
-- trade_no = 链上交易哈希，唯一约束保证同一笔转账只确认一个订单（幂等）。
--
-- 列类型与索引名严格对齐 ent 生成结果（ent/migrate/schema.go 的 UsdtOrdersTable），
-- 避免将来若启用 ent auto-migrate 时重复建索引或类型漂移。

CREATE TABLE IF NOT EXISTS usdt_orders (
    id                BIGSERIAL      PRIMARY KEY,
    order_no          VARCHAR(64)    NOT NULL UNIQUE,
    trade_no          VARCHAR(80),
    user_id           BIGINT         NOT NULL,
    amount            DECIMAL(10,2)  NOT NULL,
    credit_amount     DECIMAL(10,2)  NOT NULL DEFAULT 0,
    multiplier        DECIMAL(10,2)  NOT NULL DEFAULT 1,
    chain             VARCHAR(20)    NOT NULL DEFAULT 'trc20',
    receiving_address VARCHAR(64)    NOT NULL,
    usdt_rate         DECIMAL(20,8)  NOT NULL DEFAULT 0,
    usdt_amount       DECIMAL(20,6)  NOT NULL,
    paid_usdt_amount  DECIMAL(20,6),
    from_address      VARCHAR(64),
    block_number      BIGINT,
    status            VARCHAR(20)    NOT NULL DEFAULT 'pending',
    pay_type          VARCHAR(20)    DEFAULT 'usdt',
    paid_at           TIMESTAMPTZ,
    source_domain     VARCHAR(255),
    created_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    expired_at        TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS usdtorder_user_id    ON usdt_orders (user_id);
CREATE INDEX IF NOT EXISTS usdtorder_status     ON usdt_orders (status);
CREATE INDEX IF NOT EXISTS usdtorder_order_no   ON usdt_orders (order_no);
CREATE INDEX IF NOT EXISTS usdtorder_created_at ON usdt_orders (created_at);

-- 唯一应付金额：同一链下，pending 订单的 usdt_amount 不可重复（唯一归属的基础）。
CREATE UNIQUE INDEX IF NOT EXISTS usdtorder_chain_usdt_amount
    ON usdt_orders (chain, usdt_amount)
    WHERE status = 'pending';

-- tx_hash 幂等：一笔链上转账只能确认一个订单。
CREATE UNIQUE INDEX IF NOT EXISTS usdtorder_trade_no
    ON usdt_orders (trade_no)
    WHERE trade_no IS NOT NULL;
