//go:build unit

package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// DeepSeek 分时段计价：本 fork 的口径（合并 main 后重写）
//
// 上游本轮新增的 deepseek_pricing_test.go 钉的是**上游口径**：$ 低谷价常量 +
// deepseekPeakMultiplierAt 高峰 ×2 + applyModelSpecificPricingPolicy 整块强制覆盖。
// 本 fork 用的是 pricing_service.go 的官方 ¥ 表（deepSeekPricingTable）＋
// pricing_time_tier.go 的 deepSeekOfficialSchedule：
//
//   - 表价恒为**最贵档**（北京时间工作日 09:00-12:00 / 14:00-18:00 高峰）；
//   - 空闲档由 offPeakFactor=0.5 折算；周六/周日全天空闲（peakWeekdaysOnly，
//     本轮从上游吸收的唯一一项正确性改进）；
//   - 认不出档位的 deepseek-* 兜底到最贵档并告警，绝不静默少收。
//
// 两套口径换算关系是 ¥9 ÷ $1.32 = 6.818（工作日）。若哪天有人把上游那套搬回来，
// DeepSeek 收入会掉到约 14.5%，因此下面这几个用例是**收入护栏**，不要随手改绿。
// 时段档本身的纯函数覆盖在 pricing_time_tier_weekend_test.go，
// 计价链路端到端覆盖在 pricing_band_billing_test.go。
// ---------------------------------------------------------------------------

// deepSeekCNYPeak/OffPeak 是 deepSeekPricingTable 的 ¥ 价（汇率 1:1）换算出的
// 每 token 单价，直接写死以便与表格数值对照。
const (
	dsFlashPeakInput     = 3.0e-6 // ¥3.0 per MTok
	dsFlashPeakOutput    = 9.0e-6 // ¥9.0 per MTok
	dsFlashPeakCacheRead = 1.0e-7 // ¥0.10 per MTok
	dsProPeakInput       = 9.0e-6 // ¥9.0 per MTok
	dsProPeakOutput      = 2.7e-5 // ¥27.0 per MTok
	dsProPeakCacheRead   = 3.0e-7 // ¥0.30 per MTok
)

// bjAt 按北京时间构造任意日期的时刻（bj() 固定在 2026-08-17 周一，这里要跨周末）。
func bjAt(t *testing.T, day, hour int) time.Time {
	t.Helper()
	loc := deepSeekPricingLocation()
	require.NotNil(t, loc)
	return time.Date(2026, 8, day, hour, 0, 0, 0, loc)
}

// 默认价卡（Source=LiteLLM）：¥ 表价 = 高峰档，空闲档恰为其一半。
// 若上游的 deepseekPeakMultiplierAt ×2 被搬回计费链路，peak 会变成表价的 2 倍，本用例即红。
func TestCalculateCostUnified_DeepSeekDefaultCardUsesCNYTableBands(t *testing.T) {
	bs := NewBillingService(&config.Config{}, newCNYPricingService(1.0))
	resolver := NewModelPricingResolver(nil, bs)
	tokens := UsageTokens{InputTokens: 1000, OutputTokens: 500, CacheReadTokens: 1000}

	cases := []struct {
		model                    string
		input, output, cacheRead float64
	}{
		{"deepseek-v4-flash", dsFlashPeakInput, dsFlashPeakOutput, dsFlashPeakCacheRead},
		{"deepseek-v4-pro", dsProPeakInput, dsProPeakOutput, dsProPeakCacheRead},
		// 版本化名称按子串归档，与精确名同价。
		{"deepseek-v4-flash-0731", dsFlashPeakInput, dsFlashPeakOutput, dsFlashPeakCacheRead},
		{"deepseek-v4-pro-0813", dsProPeakInput, dsProPeakOutput, dsProPeakCacheRead},
		// 本轮上游新增的官方型号，与 flash 同价。
		{"deepseek-v4-flash-vision-exp", dsFlashPeakInput, dsFlashPeakOutput, dsFlashPeakCacheRead},
	}
	for _, tc := range cases {
		t.Run(tc.model, func(t *testing.T) {
			peakTotal := 1000*tc.input + 500*tc.output + 1000*tc.cacheRead

			peak, err := bs.CalculateCostUnified(CostInput{
				Ctx: context.Background(), Model: tc.model, Tokens: tokens,
				RateMultiplier: 1.0, Resolver: resolver,
				PricingAt: bjAt(t, 24, 10), // 周一 10:00 北京时间 → 高峰
			})
			require.NoError(t, err)
			require.InDelta(t, peakTotal, peak.TotalCost, 1e-12, "高峰档必须等于 ¥ 表价")
			require.Equal(t, PricingBandPeak, peak.PricingTimeBand)

			off, err := bs.CalculateCostUnified(CostInput{
				Ctx: context.Background(), Model: tc.model, Tokens: tokens,
				RateMultiplier: 1.0, Resolver: resolver,
				PricingAt: bjAt(t, 24, 20), // 周一 20:00 北京时间 → 空闲
			})
			require.NoError(t, err)
			require.InDelta(t, peakTotal/2, off.TotalCost, 1e-12, "空闲档 = 表价的一半")
			require.Equal(t, PricingBandOffPeak, off.PricingTimeBand)
		})
	}
}

// 周六/周日全天空闲：即便落在工作日的高峰时刻。
func TestCalculateCostUnified_DeepSeekWeekendIsOffPeak(t *testing.T) {
	bs := NewBillingService(&config.Config{}, newCNYPricingService(1.0))
	resolver := NewModelPricingResolver(nil, bs)
	tokens := UsageTokens{InputTokens: 1_000_000}

	for _, tc := range []struct {
		name string
		day  int
	}{
		{"周六 10:00", 22},
		{"周日 15:00", 23},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cost, err := bs.CalculateCostUnified(CostInput{
				Ctx: context.Background(), Model: "deepseek-v4-flash", Tokens: tokens,
				RateMultiplier: 1.0, Resolver: resolver, PricingAt: bjAt(t, tc.day, 10),
			})
			require.NoError(t, err)
			require.InDelta(t, 3.0/2, cost.TotalCost, 1e-12)
			require.Equal(t, PricingBandOffPeak, cost.PricingTimeBand)
		})
	}
}

// 零值 PricingAt = 基准价（最贵档），**不**回退到"现在"。
// 上游的 TestCalculateCostUnified_DeepseekPricingAtZeroFallsBackToNow 钉的是相反语义：
// fork 的全局约定是漏接线只会多收（可发现可退款），不能因为半夜跑测试就变成谷价。
func TestCalculateCostUnified_DeepSeekZeroPricingAtIsPeakBaseline(t *testing.T) {
	bs := NewBillingService(&config.Config{}, newCNYPricingService(1.0))
	resolver := NewModelPricingResolver(nil, bs)
	tokens := UsageTokens{InputTokens: 1_000_000}

	cost, err := bs.CalculateCostUnified(CostInput{
		Ctx: context.Background(), Model: "deepseek-v4-flash", Tokens: tokens,
		RateMultiplier: 1.0, Resolver: resolver, // PricingAt 零值
	})
	require.NoError(t, err)
	require.InDelta(t, 3.0, cost.TotalCost, 1e-12, "未接线路径必须落在贵的一侧")
	require.Equal(t, "", cost.PricingTimeBand, "未标档位可作为漏接线的监控信号")
}

// 认不出档位的 deepseek-* 兜底到**最贵档** v4-pro（上游兜底到最便宜的 flash，差 40.9×）。
// 与 kimi-k3 三天少收 ¥503 的事故同型：宁可多收也绝不静默少收。
func TestGetModelPricing_UnknownDeepSeekFallsBackToMostExpensiveTier(t *testing.T) {
	bs := NewBillingService(&config.Config{}, newCNYPricingService(1.0))

	for _, m := range []string{"deepseek-foo", "deepseek-v4.5", "deepseek-v5-max", "deepseek-v3-2-251201"} {
		t.Run(m, func(t *testing.T) {
			pricing, err := bs.GetModelPricing(m)
			require.NoError(t, err)
			require.InDelta(t, dsProPeakInput, pricing.InputPricePerToken, 1e-15)
			require.InDelta(t, dsProPeakOutput, pricing.OutputPricePerToken, 1e-15)
			require.InDelta(t, dsProPeakCacheRead, pricing.CacheReadPricePerToken, 1e-15)
		})
	}
}

// ¥ 表必须压过 LiteLLM 目录里的陈旧 $ 价（远端仓库不可改，生产会先拉到旧价）。
func TestGetModelPricing_DeepSeekCNYTableWinsOverCatalogJSON(t *testing.T) {
	ps := newCNYPricingService(1.0)
	ps.pricingData = map[string]*LiteLLMModelPricing{
		"deepseek-v4-flash": {InputCostPerToken: 2.2e-7, OutputCostPerToken: 6.6e-7, CacheReadInputTokenCost: 7e-9},
		"deepseek-v4-pro":   {InputCostPerToken: 6.6e-7, OutputCostPerToken: 1.98e-6, CacheReadInputTokenCost: 2.2e-8},
	}
	bs := NewBillingService(&config.Config{}, ps)

	flash, err := bs.GetModelPricing("deepseek-v4-flash")
	require.NoError(t, err)
	require.InDelta(t, dsFlashPeakInput, flash.InputPricePerToken, 1e-15,
		"¥ 表必须压过目录的 $ 价，否则收入掉到约 14.5%%")

	pro, err := bs.GetModelPricing("deepseek-v4-pro")
	require.NoError(t, err)
	require.InDelta(t, dsProPeakInput, pro.InputPricePerToken, 1e-15)
}

// 分组自定义定价是运营方的绝对价，不受官方时段档影响（档位标记也作废）。
func TestCalculateCostUnified_DeepSeekGroupPricingNotScaledByBand(t *testing.T) {
	bs := NewBillingService(&config.Config{}, newCNYPricingService(1.0))
	resolver := NewModelPricingResolver(nil, bs)

	inputPrice, outputPrice, cacheReadPrice := 1e-6, 2e-6, 5e-8
	group := &Group{
		ID: 1, Name: "ds-group", Platform: PlatformDeepseek, Status: StatusActive,
		ModelPricing: []ChannelModelPricing{{
			Models: []string{"deepseek-v4-flash"}, BillingMode: BillingModeToken,
			InputPrice: &inputPrice, OutputPrice: &outputPrice, CacheReadPrice: &cacheReadPrice,
		}},
	}
	resolved := resolver.Resolve(context.Background(), PricingInput{Model: "deepseek-v4-flash", Group: group})
	require.Equal(t, PricingSourceGroup, resolved.Source)

	tokens := UsageTokens{InputTokens: 1000, OutputTokens: 500, CacheReadTokens: 1000}
	groupTotal := 1000*inputPrice + 500*outputPrice + 1000*cacheReadPrice

	for _, at := range []time.Time{bjAt(t, 24, 10), bjAt(t, 24, 20), bjAt(t, 22, 10)} {
		cost, err := bs.CalculateCostUnified(CostInput{
			Ctx: context.Background(), Model: "deepseek-v4-flash", Group: group,
			Tokens: tokens, RateMultiplier: 1.0, Resolver: resolver, PricingAt: at,
		})
		require.NoError(t, err)
		require.InDelta(t, groupTotal, cost.TotalCost, 1e-12,
			"分组自定义定价不应受官方时段档影响（pricingAt=%v）", at)
	}
}

// 非 DeepSeek 模型不受任何时段档影响。
func TestCalculateCostUnified_NonDeepSeekNotScaledByBand(t *testing.T) {
	bs := newTestBillingService()
	resolver := NewModelPricingResolver(nil, bs)

	tokens := UsageTokens{InputTokens: 1000, OutputTokens: 500}
	total := 1000*3e-6 + 500*15e-6 // claude-sonnet-4 fallback

	for _, at := range []time.Time{bjAt(t, 24, 10), bjAt(t, 24, 20)} {
		cost, err := bs.CalculateCostUnified(CostInput{
			Ctx: context.Background(), Model: "claude-sonnet-4", Tokens: tokens,
			RateMultiplier: 1.0, Resolver: resolver, PricingAt: at,
		})
		require.NoError(t, err)
		require.InDelta(t, total, cost.TotalCost, 1e-12)
		require.Equal(t, "", cost.PricingTimeBand)
	}
}

// ---------------------------------------------------------------------------
// 出厂兜底目录（resources/model-pricing）：与官方 $ 低谷价一致。
// 这一条与计费口径正交——DeepSeek 的实际计价走 ¥ 表，目录只是 fork 未覆盖模型的底座；
// 保留上游用例是为了守住「$0 占位条目 / 已停服别名不得回流目录」这条数据契约。
// ---------------------------------------------------------------------------

func TestDeepseekPricingFileMatchesOfficialRates(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "resources", "model-pricing", "model_prices_and_context_window.json"))
	require.NoError(t, err)

	pricingSvc := &PricingService{}
	pricingData, err := pricingSvc.parsePricingData(data)
	require.NoError(t, err)

	_, ok := pricingData["deepseek-v3-2-251201"]
	require.False(t, ok, "deepseek-v3-2-251201（$0 占位条目）必须从价格表中移除")
	for _, discontinued := range []string{"deepseek-chat", "deepseek-reasoner"} {
		_, ok := pricingData[discontinued]
		require.False(t, ok, "%s 已停止服务，必须从价格表中移除", discontinued)
	}

	tests := []struct {
		model                    string
		input, output, cacheRead float64
	}{
		{"deepseek-v4-flash", 2.2e-7, 6.6e-7, 7e-9},
		{"deepseek-v4-flash-vision-exp", 2.2e-7, 6.6e-7, 7e-9},
		{"deepseek-v4-pro", 6.6e-7, 1.98e-6, 2.2e-8},
	}
	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			entry, ok := pricingData[tt.model]
			require.True(t, ok, "model %s must exist in pricing file", tt.model)
			require.InDelta(t, tt.input, entry.InputCostPerToken, 1e-15)
			require.InDelta(t, tt.output, entry.OutputCostPerToken, 1e-15)
			require.InDelta(t, tt.cacheRead, entry.CacheReadInputTokenCost, 1e-15)
		})
	}
}
