//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"

	"github.com/stretchr/testify/require"
)

// newDeepSeekBillingService 构造一个走官方 ¥ 表（含峰谷分档）的计费服务，汇率 1:1。
func newDeepSeekBillingService() *BillingService {
	return NewBillingService(&config.Config{}, newCNYPricingService(1.0))
}

// 空闲档必须真正折半流到 CostBreakdown，并把档位标记带出来。
func TestCalculateCostAt_DeepSeekOffPeakHalvesCost(t *testing.T) {
	svc := newDeepSeekBillingService()
	tokens := UsageTokens{InputTokens: 1_000_000, OutputTokens: 1_000_000}

	peak, err := svc.CalculateCostAt("deepseek-v4-flash", tokens, 1.0, bj(t, 10, 0, 0))
	require.NoError(t, err)
	off, err := svc.CalculateCostAt("deepseek-v4-flash", tokens, 1.0, bj(t, 20, 0, 0))
	require.NoError(t, err)

	require.InDelta(t, 3.0+9.0, peak.TotalCost, 1e-9, "高峰档 = 表价")
	require.InDelta(t, (3.0+9.0)/2, off.TotalCost, 1e-9, "空闲档 = 表价的一半")
	require.Equal(t, PricingBandPeak, peak.PricingTimeBand)
	require.Equal(t, PricingBandOffPeak, off.PricingTimeBand)
}

// 🔒 计价时刻由调用方冻结（请求开始），不得回退到"现在"。
// 若未来有人把 at 参数忽略、改回 time.Now()，本用例会在非空闲时段跑时立刻失败。
func TestCalculateCostAt_UsesFrozenInstantNotNow(t *testing.T) {
	svc := newDeepSeekBillingService()
	tokens := UsageTokens{InputTokens: 1_000_000}

	// 固定传入一个空闲时刻；无论测试在一天中的哪个真实时刻运行，结果都必须是谷价。
	off, err := svc.CalculateCostAt("deepseek-v4-flash", tokens, 1.0, bj(t, 3, 0, 0))
	require.NoError(t, err)
	require.InDelta(t, 1.5, off.TotalCost, 1e-9)
	require.Equal(t, PricingBandOffPeak, off.PricingTimeBand)

	// 同理，高峰时刻恒为表价。
	peak, err := svc.CalculateCostAt("deepseek-v4-flash", tokens, 1.0, bj(t, 15, 30, 0))
	require.NoError(t, err)
	require.InDelta(t, 3.0, peak.TotalCost, 1e-9)
	require.Equal(t, PricingBandPeak, peak.PricingTimeBand)
}

// 未传时刻的旧签名 = 基准价（最贵档）且不标档位：漏接线只会多收。
func TestCalculateCost_WithoutInstantIsPeakBaseline(t *testing.T) {
	svc := newDeepSeekBillingService()
	tokens := UsageTokens{InputTokens: 1_000_000}

	cost, err := svc.CalculateCost("deepseek-v4-flash", tokens, 1.0)
	require.NoError(t, err)
	require.InDelta(t, 3.0, cost.TotalCost, 1e-9, "未接线路径必须落在贵的一侧")
	require.Equal(t, "", cost.PricingTimeBand, "未标档位可作为漏接线的监控信号")
}

// 分组倍率与官方峰谷是**正交相乘**的两套时段机制，绝不能互相折进对方。
func TestOffPeakStacksWithGroupRateMultiplier(t *testing.T) {
	svc := newDeepSeekBillingService()
	tokens := UsageTokens{InputTokens: 1_000_000}

	cost, err := svc.CalculateCostAt("deepseek-v4-flash", tokens, 3.0, bj(t, 20, 0, 0))
	require.NoError(t, err)
	require.InDelta(t, 1.5, cost.TotalCost, 1e-9, "TotalCost 是倍率前的口径")
	require.InDelta(t, 1.5*3.0, cost.ActualCost, 1e-9, "谷单价 × tokens × 分组倍率")
}

// 谷价 ×0.5 与长上下文输入 ×2.0 叠乘后恰好等于基准价——这正是对账不能只靠
// cost÷tokens 反算档位的原因，必须结合 long_context_billing_applied 一起看。
func TestOffPeakStacksWithLongContextMultiplier(t *testing.T) {
	svc := newDeepSeekBillingService()
	const threshold = 100_000
	tokens := UsageTokens{InputTokens: 200_000}

	off, err := svc.CalculateCostWithLongContextAt(
		"deepseek-v4-flash", tokens, 1.0, threshold, 2.0, bj(t, 20, 0, 0))
	require.NoError(t, err)
	require.True(t, off.LongContextBillingApplied)
	require.Equal(t, PricingBandOffPeak, off.PricingTimeBand,
		"叠加长上下文后档位标记仍须保留，否则对账无法拆分这两个乘数")

	// 谷价 1.5/MTok。TotalCost 是**倍率前**口径：范围内 100k + 范围外 100k 各按 1.5 计 = 0.30；
	// 长上下文的 ×2 只体现在 ActualCost：0.15 + 0.30 = 0.45。
	require.InDelta(t, 0.30, off.TotalCost, 1e-9)
	require.InDelta(t, 0.45, off.ActualCost, 1e-9)

	// 🔴 对账盲区实证：谷价(×0.5) × 长上下文(×2.0) 的实收金额，与「高峰档、无长上下文」
	// 的 150k 请求完全相同。仅凭 cost÷tokens 反算无法区分这两种情况，
	// 对账 SQL 必须同时看 pricing_time_band 与 long_context_billing_applied。
	peakNoLongCtx, err := svc.CalculateCostAt("deepseek-v4-flash", UsageTokens{InputTokens: 150_000}, 1.0, bj(t, 10, 0, 0))
	require.NoError(t, err)
	require.InDelta(t, off.ActualCost, peakNoLongCtx.ActualCost, 1e-9,
		"谷价×长上下文 与 峰价 的实收金额可能完全相同，对账必须结合 long_context_billing_applied")
	require.False(t, peakNoLongCtx.LongContextBillingApplied)
}

// Kimi/Qwen 未挂时段规则，任何时刻都不受影响。
func TestCalculateCostAt_NonBandedModelsUnaffected(t *testing.T) {
	svc := newDeepSeekBillingService()
	tokens := UsageTokens{InputTokens: 1_000_000}

	base, err := svc.CalculateCost("kimi-k3", tokens, 1.0)
	require.NoError(t, err)
	for _, at := range []time.Time{bj(t, 3, 0, 0), bj(t, 10, 0, 0), bj(t, 20, 0, 0)} {
		got, err := svc.CalculateCostAt("kimi-k3", tokens, 1.0, at)
		require.NoError(t, err)
		require.InDelta(t, base.TotalCost, got.TotalCost, 1e-12)
		require.Equal(t, "", got.PricingTimeBand)
	}
}

// 优先级：admin 覆盖表命中 → 时段完全失效（覆盖价是绝对价），且不得标注档位。
func TestPricingBandPrecedence_OverrideTableWins(t *testing.T) {
	ps := newCNYPricingService(1.0)
	ps.SetSettingRepository(&writableSettingRepo{store: map[string]string{
		SettingKeyModelPricingOverrides: `[{"model":"deepseek-v4","currency":"CNY","input":1,"output":2,"enabled":true}]`,
	}})
	svc := NewBillingService(&config.Config{}, ps)

	tokens := UsageTokens{InputTokens: 1_000_000, OutputTokens: 1_000_000}
	for _, at := range []time.Time{bj(t, 10, 0, 0), bj(t, 20, 0, 0)} {
		cost, err := svc.CalculateCostAt("deepseek-v4-pro", tokens, 1.0, at)
		require.NoError(t, err)
		require.InDelta(t, 1.0+2.0, cost.TotalCost, 1e-9, "覆盖价是绝对价，不随时段打折")
		require.Equal(t, "", cost.PricingTimeBand, "覆盖价成交时不得标注官方档位")
	}
}

// 优先级：分组价卡命中 → 时段失效，band 为空。
func TestPricingBandPrecedence_GroupPricingCardWins(t *testing.T) {
	svc := newDeepSeekBillingService()
	resolver := NewModelPricingResolver(nil, svc)
	inputPrice, outputPrice := 10.0/1e6, 20.0/1e6
	group := &Group{
		ID: 7,
		ModelPricing: []ChannelModelPricing{{
			Models:      []string{"deepseek-v4-flash"},
			BillingMode: BillingModeToken,
			InputPrice:  &inputPrice,
			OutputPrice: &outputPrice,
		}},
		LongContextPricingEnabled: true,
	}

	tokens := UsageTokens{InputTokens: 1_000_000, OutputTokens: 1_000_000}
	var costs []float64
	for _, at := range []time.Time{bj(t, 10, 0, 0), bj(t, 20, 0, 0)} {
		gid := group.ID
		cost, err := svc.CalculateCostUnified(CostInput{
			Ctx: context.Background(), Model: "deepseek-v4-flash",
			GroupID: &gid, Group: group, Tokens: tokens,
			RequestCount: 1, RateMultiplier: 1.0, Resolver: resolver, PricingAt: at,
		})
		require.NoError(t, err)
		require.Equal(t, "", cost.PricingTimeBand, "分组价卡成交时不得标注官方档位")
		costs = append(costs, cost.TotalCost)
	}
	require.InDelta(t, costs[0], costs[1], 1e-12, "分组价卡是绝对价，峰谷两时刻必须同价")
	require.InDelta(t, 30.0, costs[0], 1e-9)
}

// 🔴 半张价卡：只配了 input 的卡，未被覆盖的 output/cache 也必须按**基准价**成交。
//
// 这是最容易漏的一种形态——价卡只覆盖显式配置的字段，未配字段回落底价；若底价带了
// 谷价折扣，就会出现「一张只配 input 的卡在空闲时段把 output 悄悄按半价卖出，而
// Source=channel 又把 band 清空」的状态：真按谷价收、审计列却记 NULL，事后无法对账。
// 上面那个 GroupPricingCardWins 用例恰好把用到的字段都配满了，抓不到这一形态。
func TestPricingBandPrecedence_PartialGroupCardStillUsesBaselineForUncoveredFields(t *testing.T) {
	svc := newDeepSeekBillingService()
	resolver := NewModelPricingResolver(nil, svc)
	inputPrice := 10.0 / 1e6 // 只配 input，output / cache_read 留空（常规配法）
	group := &Group{
		ID: 11,
		ModelPricing: []ChannelModelPricing{{
			Models:      []string{"deepseek-v4-flash"},
			BillingMode: BillingModeToken,
			InputPrice:  &inputPrice,
		}},
		LongContextPricingEnabled: true,
	}
	tokens := UsageTokens{InputTokens: 1_000_000, OutputTokens: 1_000_000, CacheReadTokens: 1_000_000}

	var costs []float64
	for _, at := range []time.Time{bj(t, 10, 0, 0), bj(t, 20, 0, 0)} {
		gid := group.ID
		cost, err := svc.CalculateCostUnified(CostInput{
			Ctx: context.Background(), Model: "deepseek-v4-flash",
			GroupID: &gid, Group: group, Tokens: tokens,
			RequestCount: 1, RateMultiplier: 1.0, Resolver: resolver, PricingAt: at,
		})
		require.NoError(t, err)
		require.Equal(t, "", cost.PricingTimeBand, "有价卡即不标档位")
		costs = append(costs, cost.TotalCost)
	}
	require.InDelta(t, costs[0], costs[1], 1e-12,
		"半张价卡在峰谷两时刻必须同价：未被卡覆盖的字段也走基准价，不得混入谷价折扣")
	// input 走卡价 10；output 9 + cache_read 0.10 走基准（高峰）价
	require.InDelta(t, 10.0+9.0+0.10, costs[0], 1e-9)
}

// 预解析（GatewayService 侧 resolveChannelPricing 不传时刻）与内联解析（OpenAI 侧
// CalculateCostUnified 内部 Resolve）对同一张卡、同一时刻必须给出同一个金额。
// 两条记账链一旦分叉，同一配置在不同入口会差出一倍。
func TestPricingBandPrecedence_PreResolvedAndInlineAgree(t *testing.T) {
	svc := newDeepSeekBillingService()
	resolver := NewModelPricingResolver(nil, svc)
	inputPrice := 10.0 / 1e6
	group := &Group{
		ID: 12,
		ModelPricing: []ChannelModelPricing{{
			Models:      []string{"deepseek-v4-flash"},
			BillingMode: BillingModeToken,
			InputPrice:  &inputPrice,
		}},
		LongContextPricingEnabled: true,
	}
	tokens := UsageTokens{InputTokens: 1_000_000, OutputTokens: 1_000_000}
	at := bj(t, 20, 0, 0)
	gid := group.ID

	// 内联：不传 Resolved，CalculateCostUnified 内部带时刻 Resolve
	inline, err := svc.CalculateCostUnified(CostInput{
		Ctx: context.Background(), Model: "deepseek-v4-flash",
		GroupID: &gid, Group: group, Tokens: tokens,
		RequestCount: 1, RateMultiplier: 1.0, Resolver: resolver, PricingAt: at,
	})
	require.NoError(t, err)

	// 预解析：模拟 GatewayService.resolveChannelPricing（刻意不传 At）
	preResolved := resolver.Resolve(context.Background(), PricingInput{
		Model: "deepseek-v4-flash", GroupID: &gid, Group: group,
	})
	require.NotNil(t, preResolved)
	pre, err := svc.CalculateCostUnified(CostInput{
		Ctx: context.Background(), Model: "deepseek-v4-flash",
		GroupID: &gid, Group: group, Tokens: tokens,
		RequestCount: 1, RateMultiplier: 1.0, Resolver: resolver,
		Resolved: preResolved, PricingAt: at,
	})
	require.NoError(t, err)

	require.InDelta(t, pre.TotalCost, inline.TotalCost, 1e-12,
		"两条记账链对同一张价卡、同一时刻必须收敛到同一金额")
	require.Equal(t, pre.PricingTimeBand, inline.PricingTimeBand)
}

// 纯内置价路径（无分组/渠道覆盖）才会带上档位，且金额确实分档。
func TestPricingBandPrecedence_BuiltinPathCarriesBand(t *testing.T) {
	svc := newDeepSeekBillingService()
	resolver := NewModelPricingResolver(nil, svc)
	group := &Group{ID: 9, LongContextPricingEnabled: true}
	tokens := UsageTokens{InputTokens: 1_000_000}

	gid := group.ID
	off, err := svc.CalculateCostUnified(CostInput{
		Ctx: context.Background(), Model: "deepseek-v4-flash",
		GroupID: &gid, Group: group, Tokens: tokens,
		RequestCount: 1, RateMultiplier: 1.0, Resolver: resolver, PricingAt: bj(t, 20, 0, 0),
	})
	require.NoError(t, err)
	require.Equal(t, PricingBandOffPeak, off.PricingTimeBand)
	require.InDelta(t, 1.5, off.TotalCost, 1e-9)
}

// 🔒 band 与实际成交单价不允许漂移：标了 offpeak 就必须真的按谷价成交。
func TestPricingBandMatchesActualUnitPrice(t *testing.T) {
	svc := newDeepSeekBillingService()
	const mtok = 1_000_000

	for _, at := range []time.Time{
		bj(t, 0, 0, 0), bj(t, 9, 0, 0), bj(t, 11, 59, 0),
		bj(t, 12, 0, 0), bj(t, 14, 0, 0), bj(t, 18, 0, 0), bj(t, 23, 0, 0),
	} {
		cost, err := svc.CalculateCostAt("deepseek-v4-flash", UsageTokens{InputTokens: mtok}, 1.0, at)
		require.NoError(t, err)
		impliedPerM := cost.InputCost // 1M tokens，故 InputCost 即每百万单价
		switch cost.PricingTimeBand {
		case PricingBandPeak:
			require.InDeltaf(t, 3.0, impliedPerM, 1e-9, "标为 peak 却按 %v 成交 @%v", impliedPerM, at)
		case PricingBandOffPeak:
			require.InDeltaf(t, 1.5, impliedPerM, 1e-9, "标为 offpeak 却按 %v 成交 @%v", impliedPerM, at)
		default:
			t.Fatalf("内置价路径必须标注档位，得到 %q @%v", cost.PricingTimeBand, at)
		}
	}
}
