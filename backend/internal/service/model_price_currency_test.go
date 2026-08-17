package service

import (
	"strings"
	"testing"
	"time"
)

// TestModelPriceCurrency 锁定计价币种判定：仅走官方人民币定价路径（Kimi/Moonshot/DeepSeek
// 的 cnyModelPricing override）的模型返回 CNY，其余（含走美元口径 JSON/fallback 的 GLM/MiniMax）返回 USD。
func TestModelPriceCurrency(t *testing.T) {
	cnyModels := []string{
		"deepseek-v4-pro", "deepseek-v4-flash", "deepseek/deepseek-v4-pro", "DeepSeek-V4-Pro",
		"kimi-k2.6", "kimi-k2.5", "kimi-for-coding", "moonshotai/kimi-k2-6",
		"kimi-some-future-model", // 所有 kimi-* 兜底按 kimi-k2.6
		"moonshot-v1-8k", "moonshot-v1-32k", "moonshot-v1-128k",
	}
	usdModels := []string{
		// 国产但走美元口径 JSON/fallback：成本数值本就是美元，应保持 USD（不可误标 ¥）。
		"glm-5.1", "glm-4.6", "minimax-m2", "doubao-embedding-vision",
		// 海外模型
		"claude-sonnet-4-6", "claude-opus-4-5", "gpt-5.4", "gemini-3.1-pro",
		// 边界
		"", "unknown-model",
	}
	for _, m := range cnyModels {
		if got := ModelPriceCurrency(m); got != CurrencyCNY {
			t.Errorf("ModelPriceCurrency(%q) = %q, want %q", m, got, CurrencyCNY)
		}
	}
	for _, m := range usdModels {
		if got := ModelPriceCurrency(m); got != CurrencyUSD {
			t.Errorf("ModelPriceCurrency(%q) = %q, want %q", m, got, CurrencyUSD)
		}
	}
}

// TestModelPriceCurrencyMatchesOverride 保证币种判定与计费 override 同源、永不漂移：
// 凡 ModelPriceCurrency 返回 CNY 的模型，其对应 override 必返回非 nil（反之亦然）。
func TestModelPriceCurrencyMatchesOverride(t *testing.T) {
	s := &PricingService{}
	models := []string{
		"deepseek-v4-pro", "deepseek-v4-flash", "kimi-k2.6", "kimi-future", "moonshot-v1-32k",
		"glm-5.1", "minimax-m2", "claude-sonnet-4-6", "gpt-5.4", "gemini-3.1-pro", "",
	}
	for _, m := range models {
		ml := strings.ToLower(strings.TrimSpace(m))
		isCNY := ModelPriceCurrency(m) == CurrencyCNY
		// 零值时刻 = 基准价路径；档位不影响「是否命中 ¥ 口径」这一membership 判定。
		hitOverride := s.kimiMoonshotPricingOverrideAt(ml, time.Time{}) != nil ||
			s.deepSeekPricingOverrideAt(ml, time.Time{}) != nil ||
			s.qwenPricingOverrideAt(ml, time.Time{}) != nil
		if isCNY != hitOverride {
			t.Errorf("model %q: ModelPriceCurrency CNY=%v but override hit=%v (must match)", m, isCNY, hitOverride)
		}
	}
}
