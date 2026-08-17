-- 记录每笔用量命中的官方「时段档」，用于事后对账。
--
-- 背景：部分上游按时段公布两套价（DeepSeek 高峰价 + 空闲时段半价，高峰为北京时间
-- 09:00-12:00 与 14:00-18:00）。计费按请求开始冻结的时刻定档，事后必须能逐笔证明
-- 「当时是按哪一档收的」。
--
-- 为什么不能靠 cost ÷ tokens 反算：空闲档(×0.5) 叠加长上下文输入倍率(×2.0) 之后，
-- 反算出的单价与「高峰档 + 无长上下文」完全相同，两者无法区分。
--
-- priced_at 记录定档所用的那个时刻，使「按 priced_at 重算的档位必须等于 pricing_time_band」
-- 成为可机械回放的对账不变量（用 created_at - duration_ms 近似，恰在档位边界最不准）。
--
-- 两列均可空、无默认值：PG 下是元数据级变更，不重写表；CHECK 用 NOT VALID 避免全表扫描。
-- pricing_time_band 为 NULL 表示该笔未走官方内置分档价（渠道价/分组价卡/按次计费），
-- 或计价时刻未接线——后者正是漏接线的监控信号。

ALTER TABLE usage_logs
    ADD COLUMN IF NOT EXISTS pricing_time_band VARCHAR(16);

ALTER TABLE usage_logs
    DROP CONSTRAINT IF EXISTS usage_logs_pricing_time_band_check;

ALTER TABLE usage_logs
    ADD CONSTRAINT usage_logs_pricing_time_band_check
    CHECK (
        pricing_time_band IS NULL
        OR pricing_time_band IN ('peak', 'offpeak')
    ) NOT VALID;

ALTER TABLE usage_logs
    ADD COLUMN IF NOT EXISTS priced_at TIMESTAMPTZ;

COMMENT ON COLUMN usage_logs.pricing_time_band IS
    '官方时段档：peak=高峰价，offpeak=空闲档（半价）；NULL=未走内置分档价或未接线';
COMMENT ON COLUMN usage_logs.priced_at IS
    '本笔计费定档所用的时刻（请求开始冻结），供回放校验 pricing_time_band 是否正确';
