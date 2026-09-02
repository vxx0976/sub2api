package service

import "testing"

// 内置价目表的展示币种必须与实际计费同源。
//
// ListBuiltinPricing 按 deepSeekPricingTable 的 **map key** 枚举 ¥ 表；只靠
// matchDeepSeekCNY 里 Contains(m, "v4-flash") 兜底命中的型号不是 key，会掉进
// LiteLLM 分支被渲染成 USD（deepseek-v4-flash-vision-exp 曾显示 $0.22，实收 ¥3，差 13.6 倍）。
// 危害不止于展示：admin「模型定价」页的「覆盖」按钮会把那个错价预填成 enabled 的永久
// override，而覆盖表在 GetModelPricingAt 里短路在 ¥ 表之前、还会让官方峰谷整体失效 ——
// 与 kimi-k3 少收 ¥503 同一形态，也违反「漏接线只会多收、绝不静默少收」这条底座。
func TestBuiltinPricingCurrencyMatchesActualBilling(t *testing.T) {
	ps := &PricingService{}
	for _, entry := range ps.ListBuiltinPricing() {
		cny, ok := matchDeepSeekCNY(entry.Model)
		if !ok {
			continue // 非 DeepSeek ¥ 计价模型，不在本用例范围内
		}
		if entry.Currency != CurrencyCNY {
			t.Errorf("%s 实际按 ¥ 计费（in=%v），内置视图却标 %s —— 展示与计费不同源",
				entry.Model, cny.inputCNY, entry.Currency)
		}
		if entry.InputPerM != cny.inputCNY {
			t.Errorf("%s 内置视图 input=%v，实际 ¥%v", entry.Model, entry.InputPerM, cny.inputCNY)
		}
	}
}

// 所有能被 matchDeepSeekCNY 命中的官方型号都必须是 ¥ 表的显式 key，
// 否则就会重演上面那条（兜底命中 → 不进内置视图 → 被渲染成 USD）。
func TestDeepSeekOfficialModelsAreExplicitTableKeys(t *testing.T) {
	for _, model := range []string{
		"deepseek-v4-flash",
		"deepseek-v4-pro",
		"deepseek-v4-flash-vision-exp",
	} {
		if _, ok := deepSeekPricingTable[model]; !ok {
			t.Errorf("%s 未列为 deepSeekPricingTable 的显式 key —— "+
				"只靠 Contains 兜底会让它在 admin 内置价目表里显示成美元价", model)
		}
	}
}
