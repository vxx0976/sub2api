package service

import (
	"strings"
	"time"
)

// optionalTimePtr 把零值时刻映射成 nil（对应 DB 里的 NULL）。
func optionalTimePtr(v time.Time) *time.Time {
	if v.IsZero() {
		return nil
	}
	return &v
}

// costPricingTimeBand 取出本次计费命中的官方时段档；cost 为 nil 时返回空串。
// 空串经 optionalTrimmedStringPtr 变成 nil → 落库为 NULL，正是「未走内置分档价
// 或计价时刻未接线」的监控信号。
func costPricingTimeBand(cost *CostBreakdown) string {
	if cost == nil {
		return ""
	}
	return cost.PricingTimeBand
}

func optionalTrimmedStringPtr(raw string) *string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func optionalStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func forwardResultBillingModel(requestedModel, upstreamModel string) string {
	if trimmed := strings.TrimSpace(requestedModel); trimmed != "" {
		return trimmed
	}
	return strings.TrimSpace(upstreamModel)
}

func optionalInt64Ptr(v int64) *int64 {
	if v == 0 {
		return nil
	}
	return &v
}
