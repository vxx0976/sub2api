package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func newKimiPricingService(rate float64) *PricingService {
	return &PricingService{
		cfg:         &config.Config{Pricing: config.PricingConfig{CNYToUSDRate: rate}},
		pricingData: map[string]*LiteLLMModelPricing{},
	}
}

func TestGetModelPricing_KimiK26OfficialPriceConvertedToUSD(t *testing.T) {
	const rate = 1.0
	got := newKimiPricingService(rate).GetModelPricing("kimi-k2.6")
	require.NotNil(t, got)
	// 官方人民币价（每 100 万 token）÷ 汇率 ÷ 1e6 = 每 token 美元价。
	require.InDelta(t, 6.5/rate/1e6, got.InputCostPerToken, 1e-15)
	require.InDelta(t, 27.0/rate/1e6, got.OutputCostPerToken, 1e-15)
	require.InDelta(t, 1.1/rate/1e6, got.CacheReadInputTokenCost, 1e-15)
	require.Equal(t, "moonshot", got.LiteLLMProvider)
	require.True(t, got.SupportsPromptCaching)
}

func TestGetModelPricing_KimiK26NameVariants(t *testing.T) {
	svc := newKimiPricingService(1.0)
	for _, name := range []string{"kimi-k2.6", "kimi-k2-6", "Kimi-K2.6", "moonshotai/Kimi-K2.6", "moonshot/kimi-k2-6"} {
		require.NotNilf(t, svc.GetModelPricing(name), "expected pricing for %q", name)
	}
}

func TestGetModelPricing_KimiMoonshotAllModels(t *testing.T) {
	const rate = 1.0
	svc := newKimiPricingService(rate)

	tests := []struct {
		model    string
		inputCNY float64
		outputCNY float64
		cacheCNY float64 // 0 = 无缓存
	}{
		{"kimi-k2.6", 6.5, 27.0, 1.1},
		{"kimi-k2.5", 4.0, 21.0, 0.7},
		{"kimi-for-coding", 6.5, 27.0, 1.1},
		{"moonshot-v1-8k", 2.0, 10.0, 0},
		{"moonshot-v1-32k", 5.0, 20.0, 0},
		{"moonshot-v1-128k", 10.0, 30.0, 0},
		{"moonshot-v1-8k-vision-preview", 2.0, 10.0, 0},
		{"moonshot-v1-32k-vision-preview", 5.0, 20.0, 0},
		{"moonshot-v1-128k-vision-preview", 10.0, 30.0, 0},
	}
	for _, tt := range tests {
		got := svc.GetModelPricing(tt.model)
		require.NotNilf(t, got, "expected pricing for %q", tt.model)
		require.InDeltaf(t, tt.inputCNY/rate/1e6, got.InputCostPerToken, 1e-15, "%s input", tt.model)
		require.InDeltaf(t, tt.outputCNY/rate/1e6, got.OutputCostPerToken, 1e-15, "%s output", tt.model)
		if tt.cacheCNY > 0 {
			require.InDeltaf(t, tt.cacheCNY/rate/1e6, got.CacheReadInputTokenCost, 1e-15, "%s cache", tt.model)
			require.True(t, got.SupportsPromptCaching, "%s should support caching", tt.model)
		}
		require.Equal(t, "moonshot", got.LiteLLMProvider)
	}
}

func TestGetModelPricing_KimiK26RateConfigurableAndFallback(t *testing.T) {
	// 汇率可配置：改汇率即改美元价。
	got1 := newKimiPricingService(7.0).GetModelPricing("kimi-k2.6")
	require.NotNil(t, got1)
	require.InDelta(t, 6.5/7.0/1e6, got1.InputCostPerToken, 1e-15)

	// 配置缺失/非法（0 或负）时回退到兜底汇率 defaultCNYToUSDRate。
	got2 := newKimiPricingService(0).GetModelPricing("kimi-k2.6")
	require.NotNil(t, got2)
	require.InDelta(t, 6.5/defaultCNYToUSDRate/1e6, got2.InputCostPerToken, 1e-15)

	// 所有 kimi-* 模型兜底到 kimi-k2.6 计费（含未来新模型）。
	got3 := newKimiPricingService(1.0).GetModelPricing("kimi-k2-thinking")
	require.NotNil(t, got3)
	require.InDelta(t, 6.5/1.0/1e6, got3.InputCostPerToken, 1e-15)
}

func newCNYPricingService(rate float64) *PricingService {
	return &PricingService{
		cfg:         &config.Config{Pricing: config.PricingConfig{CNYToUSDRate: rate}},
		pricingData: map[string]*LiteLLMModelPricing{},
	}
}

func TestGetModelPricing_DeepSeekV4AllModels(t *testing.T) {
	const rate = 1.0
	svc := newCNYPricingService(rate)

	tests := []struct {
		model    string
		inputCNY float64
		outputCNY float64
		cacheCNY float64
	}{
		{"deepseek-v4-flash", 1.0, 2.0, 0.02},
		{"deepseek-v4-pro", 3.0, 6.0, 0.025},
	}
	for _, tt := range tests {
		got := svc.GetModelPricing(tt.model)
		require.NotNilf(t, got, "expected pricing for %q", tt.model)
		require.InDeltaf(t, tt.inputCNY/rate/1e6, got.InputCostPerToken, 1e-15, "%s input", tt.model)
		require.InDeltaf(t, tt.outputCNY/rate/1e6, got.OutputCostPerToken, 1e-15, "%s output", tt.model)
		require.InDeltaf(t, tt.cacheCNY/rate/1e6, got.CacheReadInputTokenCost, 1e-15, "%s cache", tt.model)
		require.True(t, got.SupportsPromptCaching, "%s should support caching", tt.model)
		require.Equal(t, "deepseek", got.LiteLLMProvider)
	}
}

func TestGetModelPricing_DeepSeekV4NameVariants(t *testing.T) {
	svc := newCNYPricingService(1.0)
	// All V4 variants should resolve
	for _, name := range []string{
		"deepseek-v4-flash",
		"deepseek-v4-pro",
		"DeepSeek-V4-Flash",
		"deepseek/deepseek-v4-flash",
		"deepseek/deepseek-v4-pro",
	} {
		got := svc.GetModelPricing(name)
		require.NotNilf(t, got, "expected pricing for %q", name)
		require.Equal(t, "deepseek", got.LiteLLMProvider)
	}
}

func TestGetModelPricing_DeepSeekV4RateConfigurable(t *testing.T) {
	// 汇率可配置
	got := newCNYPricingService(7.0).GetModelPricing("deepseek-v4-flash")
	require.NotNil(t, got)
	require.InDelta(t, 1.0/7.0/1e6, got.InputCostPerToken, 1e-15)

	// 配置缺失（0）时回退到兜底汇率
	got2 := newCNYPricingService(0).GetModelPricing("deepseek-v4-flash")
	require.NotNil(t, got2)
	require.InDelta(t, 1.0/defaultCNYToUSDRate/1e6, got2.InputCostPerToken, 1e-15)
}

func TestParsePricingData_ParsesPriorityAndServiceTierFields(t *testing.T) {
	svc := &PricingService{}
	body := []byte(`{
		"gpt-5.4": {
			"input_cost_per_token": 0.0000025,
			"input_cost_per_token_priority": 0.000005,
			"output_cost_per_token": 0.000015,
			"output_cost_per_token_priority": 0.00003,
			"cache_creation_input_token_cost": 0.0000025,
			"cache_read_input_token_cost": 0.00000025,
			"cache_read_input_token_cost_priority": 0.0000005,
			"supports_service_tier": true,
			"supports_prompt_caching": true,
			"litellm_provider": "openai",
			"mode": "chat"
		}
	}`)

	data, err := svc.parsePricingData(body)
	require.NoError(t, err)
	pricing := data["gpt-5.4"]
	require.NotNil(t, pricing)
	require.InDelta(t, 5e-6, pricing.InputCostPerTokenPriority, 1e-12)
	require.InDelta(t, 3e-5, pricing.OutputCostPerTokenPriority, 1e-12)
	require.InDelta(t, 5e-7, pricing.CacheReadInputTokenCostPriority, 1e-12)
	require.True(t, pricing.SupportsServiceTier)
}

func TestGetModelPricing_Gpt53CodexSparkUsesGpt51CodexPricing(t *testing.T) {
	sparkPricing := &LiteLLMModelPricing{InputCostPerToken: 1}
	gpt53Pricing := &LiteLLMModelPricing{InputCostPerToken: 9}

	svc := &PricingService{
		pricingData: map[string]*LiteLLMModelPricing{
			"gpt-5.1-codex": sparkPricing,
			"gpt-5.3":       gpt53Pricing,
		},
	}

	got := svc.GetModelPricing("gpt-5.3-codex-spark")
	require.Same(t, sparkPricing, got)
}

func TestGetModelPricing_Gpt53CodexFallbackStillUsesGpt52Codex(t *testing.T) {
	gpt52CodexPricing := &LiteLLMModelPricing{InputCostPerToken: 2}

	svc := &PricingService{
		pricingData: map[string]*LiteLLMModelPricing{
			"gpt-5.2-codex": gpt52CodexPricing,
		},
	}

	got := svc.GetModelPricing("gpt-5.3-codex")
	require.Same(t, gpt52CodexPricing, got)
}

func TestGetModelPricing_OpenAIFallbackMatchedLoggedAsInfo(t *testing.T) {
	logSink, restore := captureStructuredLog(t)
	defer restore()

	gpt52CodexPricing := &LiteLLMModelPricing{InputCostPerToken: 2}
	svc := &PricingService{
		pricingData: map[string]*LiteLLMModelPricing{
			"gpt-5.2-codex": gpt52CodexPricing,
		},
	}

	got := svc.GetModelPricing("gpt-5.3-codex")
	require.Same(t, gpt52CodexPricing, got)

	require.True(t, logSink.ContainsMessageAtLevel("[Pricing] OpenAI fallback matched gpt-5.3-codex -> gpt-5.2-codex", "info"))
	require.False(t, logSink.ContainsMessageAtLevel("[Pricing] OpenAI fallback matched gpt-5.3-codex -> gpt-5.2-codex", "warn"))
}

func TestGetModelPricing_Gpt54UsesStaticFallbackWhenRemoteMissing(t *testing.T) {
	svc := &PricingService{
		pricingData: map[string]*LiteLLMModelPricing{
			"gpt-5.1-codex": &LiteLLMModelPricing{InputCostPerToken: 1.25e-6},
		},
	}

	got := svc.GetModelPricing("gpt-5.4")
	require.NotNil(t, got)
	require.InDelta(t, 2.5e-6, got.InputCostPerToken, 1e-12)
	require.InDelta(t, 1.5e-5, got.OutputCostPerToken, 1e-12)
	require.InDelta(t, 2.5e-7, got.CacheReadInputTokenCost, 1e-12)
	require.Equal(t, 272000, got.LongContextInputTokenThreshold)
	require.InDelta(t, 2.0, got.LongContextInputCostMultiplier, 1e-12)
	require.InDelta(t, 1.5, got.LongContextOutputCostMultiplier, 1e-12)
}

func TestGetModelPricing_OpenAICompactAliasUsesStaticFallback(t *testing.T) {
	svc := &PricingService{
		pricingData: map[string]*LiteLLMModelPricing{
			"gpt-5.1-codex": {InputCostPerToken: 1.25e-6},
		},
	}

	got := svc.GetModelPricing("openai/gpt5.5")
	require.NotNil(t, got)
	require.InDelta(t, 2.5e-6, got.InputCostPerToken, 1e-12)
	require.InDelta(t, 1.5e-5, got.OutputCostPerToken, 1e-12)
}

func TestDefaultPricingIncludesCodexAutoReview(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "resources", "model-pricing", "model_prices_and_context_window.json"))
	require.NoError(t, err)

	svc := &PricingService{}
	pricingData, err := svc.parsePricingData(data)
	require.NoError(t, err)
	svc.pricingData = pricingData

	got := svc.GetModelPricing("codex-auto-review")
	require.NotNil(t, got)
	require.InDelta(t, 5e-6, got.InputCostPerToken, 1e-12)
	require.InDelta(t, 3e-5, got.OutputCostPerToken, 1e-12)
	require.InDelta(t, 5e-7, got.CacheReadInputTokenCost, 1e-12)
}

func TestGetModelPricing_Gpt54MiniUsesDedicatedStaticFallbackWhenRemoteMissing(t *testing.T) {
	svc := &PricingService{
		pricingData: map[string]*LiteLLMModelPricing{
			"gpt-5.1-codex": {InputCostPerToken: 1.25e-6},
		},
	}

	got := svc.GetModelPricing("gpt-5.4-mini")
	require.NotNil(t, got)
	require.InDelta(t, 7.5e-7, got.InputCostPerToken, 1e-12)
	require.InDelta(t, 4.5e-6, got.OutputCostPerToken, 1e-12)
	require.InDelta(t, 7.5e-8, got.CacheReadInputTokenCost, 1e-12)
	require.Zero(t, got.LongContextInputTokenThreshold)
}

func TestGetModelPricing_Gpt54NanoUsesDedicatedStaticFallbackWhenRemoteMissing(t *testing.T) {
	svc := &PricingService{
		pricingData: map[string]*LiteLLMModelPricing{
			"gpt-5.1-codex": {InputCostPerToken: 1.25e-6},
		},
	}

	got := svc.GetModelPricing("gpt-5.4-nano")
	require.NotNil(t, got)
	require.InDelta(t, 2e-7, got.InputCostPerToken, 1e-12)
	require.InDelta(t, 1.25e-6, got.OutputCostPerToken, 1e-12)
	require.InDelta(t, 2e-8, got.CacheReadInputTokenCost, 1e-12)
	require.Zero(t, got.LongContextInputTokenThreshold)
}

func TestGetModelPricing_ImageModelDoesNotFallbackToTextModel(t *testing.T) {
	imagePricing := &LiteLLMModelPricing{InputCostPerToken: 3}
	textPricing := &LiteLLMModelPricing{InputCostPerToken: 9}

	svc := &PricingService{
		pricingData: map[string]*LiteLLMModelPricing{
			"gpt-image-2": imagePricing,
			"gpt-5.4":     textPricing,
		},
	}

	got := svc.GetModelPricing("gpt-image-3")
	require.Same(t, imagePricing, got)
}

func TestParsePricingData_PreservesPriorityAndServiceTierFields(t *testing.T) {
	raw := map[string]any{
		"gpt-5.4": map[string]any{
			"input_cost_per_token":                 2.5e-6,
			"input_cost_per_token_priority":        5e-6,
			"output_cost_per_token":                15e-6,
			"output_cost_per_token_priority":       30e-6,
			"cache_read_input_token_cost":          0.25e-6,
			"cache_read_input_token_cost_priority": 0.5e-6,
			"supports_service_tier":                true,
			"supports_prompt_caching":              true,
			"litellm_provider":                     "openai",
			"mode":                                 "chat",
		},
	}
	body, err := json.Marshal(raw)
	require.NoError(t, err)

	svc := &PricingService{}
	pricingMap, err := svc.parsePricingData(body)
	require.NoError(t, err)

	pricing := pricingMap["gpt-5.4"]
	require.NotNil(t, pricing)
	require.InDelta(t, 2.5e-6, pricing.InputCostPerToken, 1e-12)
	require.InDelta(t, 5e-6, pricing.InputCostPerTokenPriority, 1e-12)
	require.InDelta(t, 15e-6, pricing.OutputCostPerToken, 1e-12)
	require.InDelta(t, 30e-6, pricing.OutputCostPerTokenPriority, 1e-12)
	require.InDelta(t, 0.25e-6, pricing.CacheReadInputTokenCost, 1e-12)
	require.InDelta(t, 0.5e-6, pricing.CacheReadInputTokenCostPriority, 1e-12)
	require.True(t, pricing.SupportsServiceTier)
}

func TestParsePricingData_PreservesServiceTierPriorityFields(t *testing.T) {
	svc := &PricingService{}
	pricingData, err := svc.parsePricingData([]byte(`{
		"gpt-5.4": {
			"input_cost_per_token": 0.0000025,
			"input_cost_per_token_priority": 0.000005,
			"output_cost_per_token": 0.000015,
			"output_cost_per_token_priority": 0.00003,
			"cache_read_input_token_cost": 0.00000025,
			"cache_read_input_token_cost_priority": 0.0000005,
			"supports_service_tier": true,
			"litellm_provider": "openai",
			"mode": "chat"
		}
	}`))
	require.NoError(t, err)

	pricing := pricingData["gpt-5.4"]
	require.NotNil(t, pricing)
	require.InDelta(t, 0.0000025, pricing.InputCostPerToken, 1e-12)
	require.InDelta(t, 0.000005, pricing.InputCostPerTokenPriority, 1e-12)
	require.InDelta(t, 0.000015, pricing.OutputCostPerToken, 1e-12)
	require.InDelta(t, 0.00003, pricing.OutputCostPerTokenPriority, 1e-12)
	require.InDelta(t, 0.00000025, pricing.CacheReadInputTokenCost, 1e-12)
	require.InDelta(t, 0.0000005, pricing.CacheReadInputTokenCostPriority, 1e-12)
	require.True(t, pricing.SupportsServiceTier)
}

// ---------------------------------------------------------------------------
// ListModelNamesByProvider
// ---------------------------------------------------------------------------

func TestListModelNamesByProvider_ReturnsMatchingModels(t *testing.T) {
	svc := &PricingService{
		pricingData: map[string]*LiteLLMModelPricing{
			"claude-opus-4-5-20251101": {LiteLLMProvider: "anthropic", InputCostPerToken: 1.5e-5},
			"claude-sonnet-4-5":        {LiteLLMProvider: "anthropic", InputCostPerToken: 3e-6},
			"gpt-4o":                   {LiteLLMProvider: "openai", InputCostPerToken: 5e-6},
			"gemini-2.5-pro":           {LiteLLMProvider: "google", InputCostPerToken: 1.25e-6},
		},
	}

	got := svc.ListModelNamesByProvider("anthropic")
	require.ElementsMatch(t, []string{"claude-opus-4-5-20251101", "claude-sonnet-4-5"}, got)
	// Must be sorted
	require.Equal(t, "claude-opus-4-5-20251101", got[0])
	require.Equal(t, "claude-sonnet-4-5", got[1])
}

func TestListModelNamesByProvider_CaseInsensitive(t *testing.T) {
	svc := &PricingService{
		pricingData: map[string]*LiteLLMModelPricing{
			"gpt-4o": {LiteLLMProvider: "OpenAI", InputCostPerToken: 5e-6},
		},
	}

	got := svc.ListModelNamesByProvider("openai")
	require.Equal(t, []string{"gpt-4o"}, got)

	got2 := svc.ListModelNamesByProvider("OPENAI")
	require.Equal(t, []string{"gpt-4o"}, got2)
}

func TestListModelNamesByProvider_NoMatch(t *testing.T) {
	svc := &PricingService{
		pricingData: map[string]*LiteLLMModelPricing{
			"gpt-4o": {LiteLLMProvider: "openai", InputCostPerToken: 5e-6},
		},
	}

	got := svc.ListModelNamesByProvider("anthropic")
	require.NotNil(t, got)
	require.Empty(t, got)
}

func TestListModelNamesByProvider_EmptyCatalog(t *testing.T) {
	svc := &PricingService{
		pricingData: map[string]*LiteLLMModelPricing{},
	}

	got := svc.ListModelNamesByProvider("openai")
	require.NotNil(t, got)
	require.Empty(t, got)
}
