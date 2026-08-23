package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// 账号级长上下文开关必须能**否决**分组开关（显式 false 时）。
//
// 上游把它改成了「只加开、false 不否决」。两侧默认值叠起来会让那个语义在生产上全线多收：
// 迁移 222 把 groups.long_context_pricing_enabled 设成 NOT NULL DEFAULT TRUE 并把存量分组
// 全刷成 TRUE，而 OpenAI 账号的 openai_long_context_billing_enabled 在
// admin_account.go 的 normalize 里缺省写 false。「分组 true × 账号 false」正是绝大多数
// 生产账号的组合：按上游语义这类请求 input ×2、output ×1.5（实测 0.900 → 1.725，多收 ~92%），
// 且后台把开关关掉也不生效——变成一个看起来能关、实际关不掉的开关。
// openAILongContextBillingGate() 返回 *bool 且对非 OpenAI 平台返回 nil 而非 &false
// （其注释明写「否则会否决官方模型阶梯」）也只有在否决语义下才讲得通。
//
// 本 fork 的终态是**双向覆盖**：显式 true 可在分组关时单独开（上游新增的能力），
// 显式 false 可否决分组的开（fork 原有语义），nil 完全跟随分组。
func TestLongContextAccountGateOverridesGroupBothWays(t *testing.T) {
	svc := &BillingService{}
	resolver := &ModelPricingResolver{}
	tokens := UsageTokens{InputTokens: 200}

	newResolved := func(groupEnabled bool) *ResolvedPricing {
		return &ResolvedPricing{
			BasePricing:               &ModelPricing{InputPricePerToken: 1e-6},
			Intervals:                 []PricingInterval{{MinTokens: 100, InputMultiplier: pricingMultiplier(2)}},
			longContextPricingEnabled: groupEnabled,
		}
	}
	const baseCost, liftedCost = 200e-6, 400e-6

	yes, no := true, false
	for _, tc := range []struct {
		name    string
		groupOn bool
		gate    *bool
		want    float64
	}{
		{"分组开 + 账号未设 → 生效", true, nil, liftedCost},
		{"分组开 + 账号显式 true → 生效", true, &yes, liftedCost},
		{"分组开 + 账号显式 false → **被否决**", true, &no, baseCost},
		{"分组关 + 账号未设 → 不生效", false, nil, baseCost},
		{"分组关 + 账号显式 true → 单独开启", false, &yes, liftedCost},
		{"分组关 + 账号显式 false → 不生效", false, &no, baseCost},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cost, err := svc.calculateTokenCost(newResolved(tc.groupOn), CostInput{
				Model: "custom", Tokens: tokens, RateMultiplier: 1, Resolver: resolver,
				LongContextBillingEnabled: tc.gate,
			})
			require.NoError(t, err)
			require.InDelta(t, tc.want, cost.TotalCost, 1e-12)
		})
	}
}
