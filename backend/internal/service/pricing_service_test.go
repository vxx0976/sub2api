package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

// bj 按**北京时间**构造时刻，供官方时段分档（peak/offpeak）的用例使用。
// 刻意用显式时区构造而非本地时间：档位判定硬锚 Asia/Shanghai，用例结果不得随
// 运行机器/CI 的时区变化。放在无 build tag 的文件里，unit 与默认构建都能用。
func bj(t *testing.T, hour, min, sec int) time.Time {
	t.Helper()
	loc, err := time.LoadLocation("Asia/Shanghai")
	require.NoError(t, err)
	return time.Date(2026, 8, 17, hour, min, sec, 0, loc)
}

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
		model     string
		inputCNY  float64
		outputCNY float64
		cacheCNY  float64 // 0 = 无缓存
	}{
		{"kimi-k3", 20.0, 100.0, 2.0},
		{"kimi-k3-preview", 20.0, 100.0, 2.0}, // k3 变体不得回落到 k2.6 少计
		{"moonshotai/kimi-k3", 20.0, 100.0, 2.0},
		{"kimi-k2.7-code", 6.5, 27.0, 1.3},
		{"kimi-k2.7", 6.5, 27.0, 1.3},
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
		model     string
		inputCNY  float64
		outputCNY float64
		cacheCNY  float64
	}{
		{"deepseek-v4-flash", 3.0, 9.0, 0.10},
		{"deepseek-v4-pro", 9.0, 27.0, 0.30},
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

// TestGetModelPricingAt_DeepSeekOffPeak 锁定空闲档：表价（高峰价）的一半。
// ⚠️ 上面几个既有用例走零值时刻（基准价路径），即使峰谷因子整体写反也照样全绿，
// 真正兜底的是本用例与 TestTimeTierFactorNeverExceedsOne。
func TestGetModelPricingAt_DeepSeekOffPeak(t *testing.T) {
	const rate = 1.0
	svc := newCNYPricingService(rate)
	offPeak := bj(t, 20, 0, 0) // 北京 20:00 → 空闲档

	tests := []struct {
		model     string
		inputCNY  float64
		outputCNY float64
		cacheCNY  float64
	}{
		{"deepseek-v4-flash", 1.5, 4.5, 0.05},
		{"deepseek-v4-pro", 4.5, 13.5, 0.15},
		// 已下线别名并入 ¥ 口径后同样跟随分档（按 flash 计价）
		{"deepseek-chat", 1.5, 4.5, 0.05},
		{"deepseek-reasoner", 1.5, 4.5, 0.05},
	}
	for _, tt := range tests {
		got := svc.GetModelPricingAt(tt.model, offPeak)
		require.NotNilf(t, got, "expected pricing for %q", tt.model)
		require.InDeltaf(t, tt.inputCNY/rate/1e6, got.InputCostPerToken, 1e-15, "%s input", tt.model)
		require.InDeltaf(t, tt.outputCNY/rate/1e6, got.OutputCostPerToken, 1e-15, "%s output", tt.model)
		require.InDeltaf(t, tt.cacheCNY/rate/1e6, got.CacheReadInputTokenCost, 1e-15, "%s cache", tt.model)
		require.Truef(t, got.SupportsPromptCaching, "%s should still support caching", tt.model)
		require.Equalf(t, "deepseek", got.LiteLLMProvider, "%s provider", tt.model)
		require.Equalf(t, PricingBandOffPeak, got.PricingTimeBand, "%s band", tt.model)
	}
}

// 高峰时刻必须等于表价，且标注 peak 档。
func TestGetModelPricingAt_DeepSeekPeakEqualsTablePrice(t *testing.T) {
	svc := newCNYPricingService(1.0)
	got := svc.GetModelPricingAt("deepseek-v4-flash", bj(t, 10, 0, 0))
	require.NotNil(t, got)
	require.InDelta(t, 3.0/1e6, got.InputCostPerToken, 1e-15)
	require.InDelta(t, 9.0/1e6, got.OutputCostPerToken, 1e-15)
	require.InDelta(t, 0.10/1e6, got.CacheReadInputTokenCost, 1e-15)
	require.Equal(t, PricingBandPeak, got.PricingTimeBand)
}

// 零值时刻 = 基准价（最贵档）且不标档位：所有未接线路径落在贵的一侧。
func TestGetModelPricingAt_ZeroTimeIsBaseline(t *testing.T) {
	svc := newCNYPricingService(1.0)
	zero := svc.GetModelPricingAt("deepseek-v4-flash", time.Time{})
	require.NotNil(t, zero)
	require.InDelta(t, 3.0/1e6, zero.InputCostPerToken, 1e-15)
	require.Equal(t, "", zero.PricingTimeBand)

	// 旧签名必须与零值时刻逐字节一致
	legacy := svc.GetModelPricing("deepseek-v4-flash")
	require.NotNil(t, legacy)
	require.Equal(t, *zero, *legacy)
}

// 名字变体（大小写、provider 前缀）在峰谷两侧都要正确分档。
func TestGetModelPricingAt_DeepSeekNameVariantsBothBands(t *testing.T) {
	svc := newCNYPricingService(1.0)
	for _, name := range []string{
		"deepseek-v4-flash", "DeepSeek-V4-Flash", "deepseek/deepseek-v4-flash",
	} {
		peak := svc.GetModelPricingAt(name, bj(t, 15, 0, 0))
		require.NotNilf(t, peak, "peak %q", name)
		require.InDeltaf(t, 3.0/1e6, peak.InputCostPerToken, 1e-15, "peak %q", name)
		require.Equalf(t, PricingBandPeak, peak.PricingTimeBand, "peak %q", name)

		off := svc.GetModelPricingAt(name, bj(t, 3, 0, 0))
		require.NotNilf(t, off, "offpeak %q", name)
		require.InDeltaf(t, 1.5/1e6, off.InputCostPerToken, 1e-15, "offpeak %q", name)
		require.Equalf(t, PricingBandOffPeak, off.PricingTimeBand, "offpeak %q", name)
	}
}

// 档位系数必须作用在汇率折算之后的同一侧：谷价 ÷ 汇率，而不是折进汇率。
func TestGetModelPricingAt_OffPeakRespectsCNYRate(t *testing.T) {
	got := newCNYPricingService(7.0).GetModelPricingAt("deepseek-v4-flash", bj(t, 22, 0, 0))
	require.NotNil(t, got)
	require.InDelta(t, 1.5/7.0/1e6, got.InputCostPerToken, 1e-15)
}

// 未知 deepseek-* 型号按最贵档（v4-pro）兜底，绝不少收；且同样跟随时段。
//
// 🔴 用例必须覆盖**带 v4 的未知型号**：下一代命名极可能是 deepseek-v4.5 / v4-max /
// v4.1-pro，早期版本用宽松的 Contains(m,"v4") 把这一族全归到最便宜的 flash 档，
// 还绕过了兜底与告警（kimi-k3 少收 ¥503 的同型）。只测 deepseek-v9-unknown
// 恰好挑中了唯一走到兜底分支的那类名字，为一条并不成立的不变量背书。
func TestGetModelPricingAt_UnknownDeepSeekFallsBackToMostExpensive(t *testing.T) {
	svc := newCNYPricingService(1.0)
	unknown := []string{
		"deepseek-v9-unknown",
		"deepseek-r2",
		"deepseek-v4.5",     // 带 v4 的未知型号
		"deepseek-v4-max",   // 同上
		"deepseek-v4.1-pro", // 含 pro 但不含 "v4-pro" 精确串
		"deepseek-v4-ultra", // 同上
		"deepseek-v4",       // 光杆 v4，档位未知
		"deepseek/deepseek-v4-turbo",
	}
	for _, name := range unknown {
		peak := svc.GetModelPricingAt(name, bj(t, 10, 0, 0))
		require.NotNilf(t, peak, "%s", name)
		require.InDeltaf(t, 9.0/1e6, peak.InputCostPerToken, 1e-15,
			"%s 是未知型号，必须按最贵档 v4-pro 计费（绝不少收）", name)
		require.InDeltaf(t, 27.0/1e6, peak.OutputCostPerToken, 1e-15, "%s output", name)
		require.Equalf(t, PricingBandPeak, peak.PricingTimeBand, "%s", name)
		require.Equalf(t, CurrencyCNY, ModelPriceCurrency(name), "%s 兜底后币种口径必须是 CNY", name)
	}
}

// 明确点名档位的变体仍按各自档位计费（不能被上面的兜底一刀切成最贵档）。
func TestGetModelPricingAt_DeepSeekNamedTierVariants(t *testing.T) {
	svc := newCNYPricingService(1.0)
	cases := []struct {
		model    string
		inputCNY float64
	}{
		{"deepseek-v4-flash-0731", 3.0},
		{"deepseek-v4-pro-260201", 9.0},
		{"DeepSeek-V4-Flash-Preview", 3.0},
	}
	for _, tt := range cases {
		got := svc.GetModelPricingAt(tt.model, bj(t, 10, 0, 0))
		require.NotNilf(t, got, "%s", tt.model)
		require.InDeltaf(t, tt.inputCNY/1e6, got.InputCostPerToken, 1e-15, "%s", tt.model)
	}
}

// Kimi / Qwen 未挂时段规则，任何时刻价格恒定且不标档位。
func TestGetModelPricingAt_KimiQwenUnaffectedByTimeBand(t *testing.T) {
	svc := newCNYPricingService(1.0)
	for _, model := range []string{"kimi-k3", "kimi-k2.6", "qwen3-max", "qwen-plus"} {
		base := svc.GetModelPricing(model)
		require.NotNilf(t, base, "%s", model)
		for _, at := range []time.Time{bj(t, 3, 0, 0), bj(t, 10, 0, 0), bj(t, 20, 0, 0)} {
			got := svc.GetModelPricingAt(model, at)
			require.NotNilf(t, got, "%s @%v", model, at)
			require.Equalf(t, *base, *got, "%s 不应随时刻变价", model)
			require.Equalf(t, "", got.PricingTimeBand, "%s 不应标注档位", model)
		}
	}
}

func TestGetModelPricing_DeepSeekV4RateConfigurable(t *testing.T) {
	// 汇率可配置
	got := newCNYPricingService(7.0).GetModelPricing("deepseek-v4-flash")
	require.NotNil(t, got)
	require.InDelta(t, 3.0/7.0/1e6, got.InputCostPerToken, 1e-15)

	// 配置缺失（0）时回退到兜底汇率
	got2 := newCNYPricingService(0).GetModelPricing("deepseek-v4-flash")
	require.NotNil(t, got2)
	require.InDelta(t, 3.0/defaultCNYToUSDRate/1e6, got2.InputCostPerToken, 1e-15)
}

func TestGetModelPricing_QwenAllModels(t *testing.T) {
	const rate = 1.0
	svc := newCNYPricingService(rate)

	tests := []struct {
		model     string
		inputCNY  float64
		outputCNY float64
		cacheCNY  float64
	}{
		{"qwen3-max", 6.0, 24.0, 1.2},
		{"qwen-max", 2.4, 9.6, 0.48},
		{"qwen-plus", 0.8, 2.0, 0.16},
		{"qwen-flash", 0.15, 1.5, 0.03},
		{"qwen-turbo", 0.3, 0.6, 0.06},
		{"qwen-long", 0.5, 2.0, 0.1},
		{"qwen3-coder-plus", 4.0, 16.0, 0.8},
		{"qwen3-coder-flash", 1.5, 6.0, 0.3},
	}
	for _, tt := range tests {
		got := svc.GetModelPricing(tt.model)
		require.NotNilf(t, got, "expected pricing for %q", tt.model)
		require.InDeltaf(t, tt.inputCNY/rate/1e6, got.InputCostPerToken, 1e-15, "%s input", tt.model)
		require.InDeltaf(t, tt.outputCNY/rate/1e6, got.OutputCostPerToken, 1e-15, "%s output", tt.model)
		require.InDeltaf(t, tt.cacheCNY/rate/1e6, got.CacheReadInputTokenCost, 1e-15, "%s cache", tt.model)
		require.True(t, got.SupportsPromptCaching, "%s should support caching", tt.model)
		require.Equal(t, "dashscope", got.LiteLLMProvider)
		require.Equal(t, CurrencyCNY, ModelPriceCurrency(tt.model), "%s currency", tt.model)
	}
}

func TestGetModelPricing_QwenNameVariantsAndFallback(t *testing.T) {
	const rate = 1.0
	svc := newCNYPricingService(rate)

	// 精确/前缀变体 + 按模式回退（dated 版本、provider 前缀、未知 qwen 兜底 qwen-plus）。
	cases := []struct {
		model    string
		inputCNY float64 // 期望命中的档位输入价
	}{
		{"qwen3-max-2026-01-23", 6.0},           // → qwen3-max
		{"qwen-max-latest", 2.4},                // → qwen-max（非 qwen3）
		{"dashscope/qwen3-coder-plus", 4.0},     // provider 前缀剥离 → coder-plus
		{"qwen3-coder-480b-a35b-instruct", 4.0}, // 含 coder → coder-plus
		{"qwen3-coder-flash-2025-07-22", 1.5},   // 含 coder+flash → coder-flash
		{"qwen3-235b-a22b", 0.8},                // 未知 → 兜底 qwen-plus
		{"qwq-32b", 0.8},                        // qwq 前缀 → 兜底 qwen-plus
	}
	for _, c := range cases {
		got := svc.GetModelPricing(c.model)
		require.NotNilf(t, got, "expected pricing for %q", c.model)
		require.InDeltaf(t, c.inputCNY/rate/1e6, got.InputCostPerToken, 1e-15, "%s input", c.model)
		require.Equal(t, "dashscope", got.LiteLLMProvider)
		require.Equal(t, CurrencyCNY, ModelPriceCurrency(c.model), "%s currency", c.model)
	}

	// 汇率可配置 + 缺失回退兜底汇率。
	require.InDelta(t, 0.8/7.0/1e6, newCNYPricingService(7.0).GetModelPricing("qwen-plus").InputCostPerToken, 1e-15)
	require.InDelta(t, 0.8/defaultCNYToUSDRate/1e6, newCNYPricingService(0).GetModelPricing("qwen-plus").InputCostPerToken, 1e-15)
}

func TestPricingSchedulerBlankRemoteURLDoesNotStart(t *testing.T) {
	svc := NewPricingService(&config.Config{Pricing: config.PricingConfig{RemoteURL: "  \t  "}}, nil)
	defer svc.Stop()

	svc.startUpdateScheduler()
	done := make(chan struct{})
	go func() {
		svc.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("blank remote URL must not start scheduler")
	}
}

func TestPricingNonEmptyInvalidRemoteURLStillReturnsValidationError(t *testing.T) {
	svc := NewPricingService(&config.Config{Pricing: config.PricingConfig{
		RemoteURL: "://invalid",
		DataDir:   t.TempDir(),
	}}, nil)

	err := svc.ForceUpdate()

	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid pricing url")
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
			"cache_creation_input_token_cost_priority": 0.000005,
			"cache_read_input_token_cost": 0.00000025,
			"cache_read_input_token_cost_priority": 0.0000005,
			"long_context_input_token_threshold": 272000,
			"long_context_input_cost_multiplier": 2,
			"long_context_output_cost_multiplier": 1.5,
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
	require.InDelta(t, 5e-6, pricing.CacheCreationInputTokenCostPriority, 1e-12)
	require.InDelta(t, 5e-7, pricing.CacheReadInputTokenCostPriority, 1e-12)
	require.Equal(t, 272000, pricing.LongContextInputTokenThreshold)
	require.InDelta(t, 2.0, pricing.LongContextInputCostMultiplier, 1e-12)
	require.InDelta(t, 1.5, pricing.LongContextOutputCostMultiplier, 1e-12)
	require.True(t, pricing.SupportsServiceTier)
}

const gpt6AstraCatalogJSON = `{
	"gpt-6-astra": {
		"litellm_provider": "openai",
		"mode": "chat",
		"input_cost_per_token": 1e-05,
		"input_cost_per_token_priority": 2e-05,
		"output_cost_per_token": 5e-05,
		"output_cost_per_token_priority": 1e-04,
		"cache_creation_input_token_cost": 1.25e-05,
		"cache_creation_input_token_cost_priority": 2.5e-05,
		"cache_read_input_token_cost": 1e-06,
		"cache_read_input_token_cost_priority": 2e-06,
		"input_cost_per_token_above_272k_tokens": 2e-05,
		"output_cost_per_token_above_272k_tokens": 7.5e-05,
		"cache_creation_input_token_cost_above_272k_tokens": 2.5e-05,
		"cache_read_input_token_cost_above_272k_tokens": 2e-06
	}
}`

func TestBillingServiceGPT6AstraUsesOfficialPricingAcrossTiersAndLongContext(t *testing.T) {
	svc := NewBillingService(&config.Config{}, newStubPricingServiceFromJSON(t, gpt6AstraCatalogJSON))
	boundaryTokens := UsageTokens{InputTokens: 100_000, CacheCreationTokens: 100_000, CacheReadTokens: 72_000, OutputTokens: 10}
	boundary, err := svc.CalculateCost("gpt-6-astra", boundaryTokens, 1)
	require.NoError(t, err)
	require.False(t, boundary.LongContextBillingApplied)
	require.InDelta(t, 100_000*10e-6, boundary.InputCost, 1e-12)
	require.InDelta(t, 100_000*12.5e-6, boundary.CacheCreationCost, 1e-12)
	require.InDelta(t, 72_000*1e-6, boundary.CacheReadCost, 1e-12)
	require.InDelta(t, 10*50e-6, boundary.OutputCost, 1e-12)

	tokens := UsageTokens{InputTokens: 100_000, CacheCreationTokens: 100_000, CacheReadTokens: 73_000, OutputTokens: 10}
	tiers := []struct {
		name        string
		serviceTier string
		priceScale  float64
	}{
		{name: "standard", priceScale: 1},
		{name: "fast", serviceTier: "priority", priceScale: 2},
		{name: "flex", serviceTier: "flex", priceScale: 0.5},
	}
	for _, tier := range tiers {
		t.Run(tier.name, func(t *testing.T) {
			cost, err := svc.CalculateCostWithServiceTier("gpt-6-astra", tokens, 1, tier.serviceTier)
			require.NoError(t, err)
			require.True(t, cost.LongContextBillingApplied)
			require.InDelta(t, 100_000*10e-6*tier.priceScale*2, cost.InputCost, 1e-12)
			require.InDelta(t, 100_000*12.5e-6*tier.priceScale*2, cost.CacheCreationCost, 1e-12)
			require.InDelta(t, 73_000*1e-6*tier.priceScale*2, cost.CacheReadCost, 1e-12)
			require.InDelta(t, 10*50e-6*tier.priceScale*1.5, cost.OutputCost, 1e-12)
		})
	}
}

func TestGPT6AstraDedicatedFallbacksUseOfficialRates(t *testing.T) {
	tests := []struct {
		name string
		svc  *BillingService
	}{
		{name: "pricing_service", svc: NewBillingService(&config.Config{}, &PricingService{pricingData: map[string]*LiteLLMModelPricing{}})},
		{name: "billing_service", svc: NewBillingService(&config.Config{}, nil)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pricing, err := tt.svc.GetModelPricing("gpt-6-astra")
			require.NoError(t, err)
			require.InDelta(t, 10e-6, pricing.InputPricePerToken, 1e-12)
			require.InDelta(t, 20e-6, pricing.InputPricePerTokenPriority, 1e-12)
			require.InDelta(t, 50e-6, pricing.OutputPricePerToken, 1e-12)
			require.InDelta(t, 100e-6, pricing.OutputPricePerTokenPriority, 1e-12)
			require.InDelta(t, 12.5e-6, pricing.CacheCreationPricePerToken, 1e-12)
			require.InDelta(t, 25e-6, pricing.CacheCreationPricePerTokenPriority, 1e-12)
			require.InDelta(t, 1e-6, pricing.CacheReadPricePerToken, 1e-12)
			require.InDelta(t, 2e-6, pricing.CacheReadPricePerTokenPriority, 1e-12)
			require.Equal(t, 272_000, pricing.LongContextInputThreshold)
			require.InDelta(t, 2.0, pricing.LongContextInputMultiplier, 1e-12)
			require.InDelta(t, 1.5, pricing.LongContextOutputMultiplier, 1e-12)
		})
	}
}

func TestPricingServiceBareGPT6AliasUsesAstra(t *testing.T) {
	astraPricing := &LiteLLMModelPricing{InputCostPerToken: 123e-6, OutputCostPerToken: 456e-6}
	pricingSvc := &PricingService{pricingData: map[string]*LiteLLMModelPricing{"gpt-6-astra": astraPricing}}
	for _, model := range []string{"gpt-6", "openai/gpt-6"} {
		pricing := pricingSvc.GetModelPricing(model)
		require.Same(t, astraPricing, pricing)
	}
}

func TestBillingService_GPT56CacheWritePricingUsesOfficialMultiplier(t *testing.T) {
	tests := []struct {
		model             string
		input             float64
		inputPriority     float64
		output            float64
		outputPriority    float64
		cacheRead         float64
		cacheReadPriority float64
	}{
		{model: "gpt-5.6-sol", input: 5e-6, inputPriority: 10e-6, output: 30e-6, outputPriority: 60e-6, cacheRead: 0.5e-6, cacheReadPriority: 1e-6},
		{model: "gpt-5.6-terra", input: 2e-6, inputPriority: 4e-6, output: 12e-6, outputPriority: 24e-6, cacheRead: 0.2e-6, cacheReadPriority: 0.4e-6},
		{model: "gpt-5.6-luna", input: 0.2e-6, inputPriority: 0.4e-6, output: 1.2e-6, outputPriority: 2.4e-6, cacheRead: 0.02e-6, cacheReadPriority: 0.04e-6},
	}
	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			pricingSvc := &PricingService{pricingData: map[string]*LiteLLMModelPricing{
				tt.model: {
					InputCostPerToken:               tt.input,
					InputCostPerTokenPriority:       tt.inputPriority,
					OutputCostPerToken:              tt.output,
					OutputCostPerTokenPriority:      tt.outputPriority,
					CacheReadInputTokenCost:         tt.cacheRead,
					CacheReadInputTokenCostPriority: tt.cacheReadPriority,
				},
			}}
			svc := NewBillingService(&config.Config{}, pricingSvc)

			pricing, err := svc.GetModelPricing(tt.model)
			require.NoError(t, err)
			require.InDelta(t, tt.input*1.25, pricing.CacheCreationPricePerToken, 1e-12)
			require.InDelta(t, tt.inputPriority*1.25, pricing.CacheCreationPricePerTokenPriority, 1e-12)
			// 阶梯由目录数据驱动：条目无 above/long_context 字段时不再由策略强补。
			require.Zero(t, pricing.LongContextInputThreshold)

			tokens := UsageTokens{InputTokens: 700, OutputTokens: 50, CacheCreationTokens: 200, CacheReadTokens: 100}
			standard, err := svc.CalculateCostWithServiceTier(tt.model, tokens, 1, "")
			require.NoError(t, err)
			require.InDelta(t, 200*tt.input*1.25, standard.CacheCreationCost, 1e-12)

			priority, err := svc.CalculateCostWithServiceTier(tt.model, tokens, 1, "priority")
			require.NoError(t, err)
			require.InDelta(t, 200*tt.inputPriority*1.25, priority.CacheCreationCost, 1e-12)

			flex, err := svc.CalculateCostWithServiceTier(tt.model, tokens, 1, "flex")
			require.NoError(t, err)
			require.InDelta(t, 200*tt.input*1.25*0.5, flex.CacheCreationCost, 1e-12)
		})
	}
}

// gpt56LadderCatalogJSON 三个 5.6 模型的目录条目：above_272k 绝对价 + priority 平价，
// cache_write 缺失由策略按 1.25 倍输入价补齐。
const gpt56LadderCatalogJSON = `{
	"gpt-5.6-sol": {"litellm_provider": "openai", "mode": "chat",
		"input_cost_per_token": 5e-06, "input_cost_per_token_priority": 1e-05,
		"output_cost_per_token": 3e-05, "output_cost_per_token_priority": 6e-05,
		"cache_read_input_token_cost": 5e-07, "cache_read_input_token_cost_priority": 1e-06,
		"input_cost_per_token_above_272k_tokens": 1e-05,
		"output_cost_per_token_above_272k_tokens": 4.5e-05,
		"cache_read_input_token_cost_above_272k_tokens": 1e-06},
	"gpt-5.6-terra": {"litellm_provider": "openai", "mode": "chat",
		"input_cost_per_token": 2e-06, "input_cost_per_token_priority": 4e-06,
		"output_cost_per_token": 1.2e-05, "output_cost_per_token_priority": 2.4e-05,
		"cache_read_input_token_cost": 2e-07, "cache_read_input_token_cost_priority": 4e-07,
		"input_cost_per_token_above_272k_tokens": 4e-06,
		"output_cost_per_token_above_272k_tokens": 1.8e-05,
		"cache_read_input_token_cost_above_272k_tokens": 4e-07},
	"gpt-5.6-luna": {"litellm_provider": "openai", "mode": "chat",
		"input_cost_per_token": 2e-07, "input_cost_per_token_priority": 4e-07,
		"output_cost_per_token": 1.2e-06, "output_cost_per_token_priority": 2.4e-06,
		"cache_read_input_token_cost": 2e-08, "cache_read_input_token_cost_priority": 4e-08,
		"input_cost_per_token_above_272k_tokens": 4e-07,
		"output_cost_per_token_above_272k_tokens": 1.8e-06,
		"cache_read_input_token_cost_above_272k_tokens": 4e-08}
}`

func TestBillingService_GPT56UsesLongContextPricingAcrossModelsAndTiers(t *testing.T) {
	models := []struct {
		name               string
		input, cached      float64
		cacheWrite, output float64
	}{
		{name: "gpt-5.6-sol", input: 5e-6, cached: 0.5e-6, cacheWrite: 6.25e-6, output: 30e-6},
		{name: "gpt-5.6-terra", input: 2e-6, cached: 0.2e-6, cacheWrite: 2.5e-6, output: 12e-6},
		{name: "gpt-5.6-luna", input: 0.2e-6, cached: 0.02e-6, cacheWrite: 0.25e-6, output: 1.2e-6},
	}
	tiers := []struct {
		name       string
		priceScale float64
	}{
		{name: "standard", priceScale: 1},
		{name: "priority", priceScale: 2},
		{name: "flex", priceScale: 0.5},
	}
	tokens := UsageTokens{
		InputTokens:         100000,
		CacheCreationTokens: 100000,
		CacheReadTokens:     73000,
		OutputTokens:        10,
	}

	for _, model := range models {
		for _, tier := range tiers {
			t.Run(model.name+"/"+tier.name, func(t *testing.T) {
				svc := NewBillingService(&config.Config{}, newStubPricingServiceFromJSON(t, gpt56LadderCatalogJSON))
				serviceTier := ""
				if tier.name != "standard" {
					serviceTier = tier.name
				}
				cost, err := svc.CalculateCostWithServiceTier(model.name, tokens, 1, serviceTier)
				require.NoError(t, err)
				require.InDelta(t, float64(tokens.InputTokens)*model.input*tier.priceScale*2, cost.InputCost, 1e-12)
				require.InDelta(t, float64(tokens.CacheCreationTokens)*model.cacheWrite*tier.priceScale*2, cost.CacheCreationCost, 1e-12)
				require.InDelta(t, float64(tokens.CacheReadTokens)*model.cached*tier.priceScale*2, cost.CacheReadCost, 1e-12)
				require.InDelta(t, float64(tokens.OutputTokens)*model.output*tier.priceScale*1.5, cost.OutputCost, 1e-12)
			})
		}
	}
}

func TestBillingService_GPT56LongContextBoundaryIsExclusive(t *testing.T) {
	svc := NewBillingService(&config.Config{}, newStubPricingServiceFromJSON(t, gpt56LadderCatalogJSON))
	tokens := UsageTokens{InputTokens: 100000, CacheCreationTokens: 100000, CacheReadTokens: 72000, OutputTokens: 10}

	cost, err := svc.CalculateCost("gpt-5.6-sol", tokens, 1)
	require.NoError(t, err)
	require.InDelta(t, 100000*5e-6, cost.InputCost, 1e-12)
	require.InDelta(t, 100000*6.25e-6, cost.CacheCreationCost, 1e-12)
	require.InDelta(t, 72000*0.5e-6, cost.CacheReadCost, 1e-12)
	require.InDelta(t, 10*30e-6, cost.OutputCost, 1e-12)
}

func TestPricingService_BareGPT56AliasDeterministicallyUsesSol(t *testing.T) {
	pricingSvc := &PricingService{pricingData: map[string]*LiteLLMModelPricing{
		"gpt-5.6-sol":   {InputCostPerToken: 5e-6},
		"gpt-5.6-terra": {InputCostPerToken: 2e-6},
		"gpt-5.6-luna":  {InputCostPerToken: 0.2e-6},
		"gpt-5.4":       {InputCostPerToken: 2.5e-6},
	}}

	for i := 0; i < 100; i++ {
		for _, alias := range []string{"gpt-5.6", "openai/gpt-5.6"} {
			pricing := pricingSvc.GetModelPricing(alias)
			require.NotNil(t, pricing)
			require.InDelta(t, 5e-6, pricing.InputCostPerToken, 1e-12, "iteration=%d alias=%s", i, alias)
		}
	}

	billingSvc := NewBillingService(&config.Config{}, pricingSvc)
	for _, alias := range []string{"gpt-5.6", "openai/gpt-5.6"} {
		pricing, err := billingSvc.GetModelPricing(alias)
		require.NoError(t, err)
		require.InDelta(t, 5e-6, pricing.InputPricePerToken, 1e-12)
		require.InDelta(t, 6.25e-6, pricing.CacheCreationPricePerToken, 1e-12)
	}
}

func TestDefaultPricingIncludesOfficialGPT56Rates(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "resources", "model-pricing", "model_prices_and_context_window.json"))
	require.NoError(t, err)

	pricingSvc := &PricingService{}
	pricingData, err := pricingSvc.parsePricingData(data)
	require.NoError(t, err)
	pricingSvc.pricingData = pricingData
	billingSvc := NewBillingService(&config.Config{}, pricingSvc)

	tests := []struct {
		model                                                             string
		input, cached, cacheWrite, output                                 float64
		inputPriority, cachedPriority, cacheWritePriority, outputPriority float64
	}{
		{model: "gpt-5.6-sol", input: 5e-6, cached: 0.5e-6, cacheWrite: 6.25e-6, output: 30e-6, inputPriority: 10e-6, cachedPriority: 1e-6, cacheWritePriority: 12.5e-6, outputPriority: 60e-6},
		{model: "gpt-5.6-terra", input: 2e-6, cached: 0.2e-6, cacheWrite: 2.5e-6, output: 12e-6, inputPriority: 4e-6, cachedPriority: 0.4e-6, cacheWritePriority: 5e-6, outputPriority: 24e-6},
		{model: "gpt-5.6-luna", input: 0.2e-6, cached: 0.02e-6, cacheWrite: 0.25e-6, output: 1.2e-6, inputPriority: 0.4e-6, cachedPriority: 0.04e-6, cacheWritePriority: 0.5e-6, outputPriority: 2.4e-6},
	}
	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			pricing, err := billingSvc.GetModelPricing(tt.model)
			require.NoError(t, err)
			require.InDelta(t, tt.input, pricing.InputPricePerToken, 1e-12)
			require.InDelta(t, tt.cached, pricing.CacheReadPricePerToken, 1e-12)
			require.InDelta(t, tt.cacheWrite, pricing.CacheCreationPricePerToken, 1e-12)
			require.InDelta(t, tt.output, pricing.OutputPricePerToken, 1e-12)
			require.InDelta(t, tt.inputPriority, pricing.InputPricePerTokenPriority, 1e-12)
			require.InDelta(t, tt.cachedPriority, pricing.CacheReadPricePerTokenPriority, 1e-12)
			require.InDelta(t, tt.cacheWritePriority, pricing.CacheCreationPricePerTokenPriority, 1e-12)
			require.InDelta(t, tt.outputPriority, pricing.OutputPricePerTokenPriority, 1e-12)
			require.Equal(t, 272000, pricing.LongContextInputThreshold)
			require.InDelta(t, 2.0, pricing.LongContextInputMultiplier, 1e-12)
			require.InDelta(t, 1.5, pricing.LongContextOutputMultiplier, 1e-12)
		})
	}
}

func TestGPT56DedicatedFallbacksUseOfficialRates(t *testing.T) {
	tests := []struct {
		model                             string
		input, cached, cacheWrite, output float64
	}{
		{model: "gpt-5.6-sol", input: 5e-6, cached: 0.5e-6, cacheWrite: 6.25e-6, output: 30e-6},
		{model: "gpt-5.6-terra", input: 2e-6, cached: 0.2e-6, cacheWrite: 2.5e-6, output: 12e-6},
		{model: "gpt-5.6-luna", input: 0.2e-6, cached: 0.02e-6, cacheWrite: 0.25e-6, output: 1.2e-6},
	}

	for _, tt := range tests {
		t.Run(tt.model+"/pricing_service", func(t *testing.T) {
			pricingSvc := &PricingService{pricingData: map[string]*LiteLLMModelPricing{
				"gpt-5.1-codex": {InputCostPerToken: 1.25e-6},
			}}
			svc := NewBillingService(&config.Config{}, pricingSvc)
			pricing, err := svc.GetModelPricing(tt.model + "-preview")
			require.NoError(t, err)
			assertGPT56FallbackPricing(t, pricing, tt.input, tt.cached, tt.cacheWrite, tt.output)
		})

		t.Run(tt.model+"/billing_service", func(t *testing.T) {
			svc := NewBillingService(&config.Config{}, nil)
			pricing, err := svc.GetModelPricing(tt.model)
			require.NoError(t, err)
			assertGPT56FallbackPricing(t, pricing, tt.input, tt.cached, tt.cacheWrite, tt.output)
		})
	}
}

func assertGPT56FallbackPricing(t *testing.T, pricing *ModelPricing, input, cached, cacheWrite, output float64) {
	t.Helper()
	require.InDelta(t, input, pricing.InputPricePerToken, 1e-12)
	require.InDelta(t, cached, pricing.CacheReadPricePerToken, 1e-12)
	require.InDelta(t, cacheWrite, pricing.CacheCreationPricePerToken, 1e-12)
	require.InDelta(t, output, pricing.OutputPricePerToken, 1e-12)
	// 静态兜底只兜基础价；阶梯由目录数据（above_272k 折算或显式字段）驱动。
	require.Zero(t, pricing.LongContextInputThreshold)
}

func TestParsePricingData_KeepsImageOnlyPricing(t *testing.T) {
	svc := &PricingService{}
	body := []byte(`{
		"image-only-model": {
			"output_cost_per_image": 0.034,
			"litellm_provider": "vertex_ai-language-models",
			"mode": "image_generation"
		}
	}`)

	data, err := svc.parsePricingData(body)
	require.NoError(t, err)
	pricing := data["image-only-model"]
	require.NotNil(t, pricing)
	require.InDelta(t, 0.034, pricing.OutputCostPerImage, 1e-12)
	require.Equal(t, "image_generation", pricing.Mode)
	// 仅有图片价的条目必须标记 token 价缺失，供 token 计费路径 fail-closed。
	require.True(t, pricing.TokenPricingAbsent)
}

func TestBillingService_GetModelPricing_FailsClosedForImageOnlyEntries(t *testing.T) {
	pricingSvc := &PricingService{}
	data, err := pricingSvc.parsePricingData([]byte(`{
		"imagen-9.0-generate": {
			"output_cost_per_image": 0.04,
			"litellm_provider": "vertex_ai-image-models",
			"mode": "image_generation"
		},
		"gemini-image-with-token-price": {
			"input_cost_per_token": 0.0,
			"output_cost_per_token": 0.0,
			"output_cost_per_image": 0.034,
			"litellm_provider": "vertex_ai-language-models",
			"mode": "image_generation"
		}
	}`))
	require.NoError(t, err)
	pricingSvc.pricingData = data
	billingSvc := NewBillingService(&config.Config{}, pricingSvc)

	// image-only 条目不得进入 token 计费（否则 token 流量按 $0 计费），
	// 必须落到 fallback / ErrModelPricingUnavailable 的 fail-closed 路径。
	_, err = billingSvc.GetModelPricing("imagen-9.0-generate")
	require.ErrorIs(t, err, ErrModelPricingUnavailable)

	// 显式 0 token 价的免费条目保持历史行为：正常返回。
	pricing, err := billingSvc.GetModelPricing("gemini-image-with-token-price")
	require.NoError(t, err)
	require.Zero(t, pricing.InputPricePerToken)

	// 图片计费路径不受影响：仍能读到 image-only 条目的图片单价。
	raw := pricingSvc.GetModelPricing("imagen-9.0-generate")
	require.NotNil(t, raw)
	require.InDelta(t, 0.04, raw.OutputCostPerImage, 1e-12)
}

func TestPricingService_MergesFallbackOnlyModels(t *testing.T) {
	dir := t.TempDir()
	fallbackFile := filepath.Join(dir, "fallback.json")
	require.NoError(t, os.WriteFile(fallbackFile, []byte(`{
		"remote-model": {
			"input_cost_per_token": 0.000001,
			"litellm_provider": "test",
			"mode": "chat"
		},
		"gemini-3.1-flash-lite-image": {
			"output_cost_per_image": 0.034,
			"litellm_provider": "vertex_ai-language-models",
			"mode": "image_generation"
		}
	}`), 0644))

	svc := &PricingService{cfg: &config.Config{}}
	svc.cfg.Pricing.FallbackFile = fallbackFile
	remoteData, err := svc.parsePricingData([]byte(`{
		"remote-model": {
			"input_cost_per_token": 0.000002,
			"litellm_provider": "test",
			"mode": "chat"
		}
	}`))
	require.NoError(t, err)

	merged := svc.mergeFallbackPricingData(remoteData)
	require.InDelta(t, 0.000002, merged["remote-model"].InputCostPerToken, 1e-12)
	require.NotNil(t, merged["gemini-3.1-flash-lite-image"])
	require.InDelta(t, 0.034, merged["gemini-3.1-flash-lite-image"].OutputCostPerImage, 1e-12)
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
	// 静态兜底只兜基础价，不携带长上下文阶梯（阶梯由目录数据驱动）。
	require.Zero(t, got.LongContextInputTokenThreshold)
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

func TestPricingService_Gemini36FlashThinkingTiersUseBasePricing(t *testing.T) {
	basePricing := &LiteLLMModelPricing{
		InputCostPerToken:       1.5e-6,
		OutputCostPerToken:      7.5e-6,
		CacheReadInputTokenCost: 0.15e-6,
	}
	svc := &PricingService{pricingData: map[string]*LiteLLMModelPricing{
		"gemini-3.6-flash": basePricing,
	}}

	for _, model := range []string{
		"gemini-3.6-flash",
		"gemini-3.6-flash-high",
		"gemini-3.6-flash-low",
		"gemini-3.6-flash-medium",
		"gemini-3.6-flash-tiered",
	} {
		t.Run(model, func(t *testing.T) {
			require.Same(t, basePricing, svc.GetModelPricing(model))
		})
	}
}

func TestPricingService_Gemini36FlashTierSpecificPricingTakesPrecedence(t *testing.T) {
	basePricing := &LiteLLMModelPricing{InputCostPerToken: 1.5e-6}
	tierPricing := &LiteLLMModelPricing{InputCostPerToken: 2e-6}
	svc := &PricingService{pricingData: map[string]*LiteLLMModelPricing{
		"gemini-3.6-flash":     basePricing,
		"gemini-3.6-flash-low": tierPricing,
	}}

	require.Same(t, tierPricing, svc.GetModelPricing("models/gemini-3.6-flash-low"))
}

func TestBillingService_Gemini36FlashThinkingTierFallbacksAreBillable(t *testing.T) {
	svc := NewBillingService(&config.Config{}, nil)
	tokens := UsageTokens{InputTokens: 1_000_000, OutputTokens: 1_000_000, CacheReadTokens: 1_000_000}

	for _, model := range []string{
		"gemini-3.6-flash",
		"gemini-3.6-flash-high",
		"gemini-3.6-flash-low",
		"gemini-3.6-flash-medium",
		"gemini-3.6-flash-tiered",
	} {
		t.Run(model, func(t *testing.T) {
			cost, err := svc.CalculateCost(model, tokens, 1)
			require.NoError(t, err)
			require.InDelta(t, 1.5, cost.InputCost, 1e-12)
			require.InDelta(t, 7.5, cost.OutputCost, 1e-12)
			require.InDelta(t, 0.15, cost.CacheReadCost, 1e-12)
			require.InDelta(t, 9.15, cost.TotalCost, 1e-12)
		})
	}
}

func TestDefaultPricingIncludesGemini36FlashRates(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "resources", "model-pricing", "model_prices_and_context_window.json"))
	require.NoError(t, err)

	pricingSvc := &PricingService{}
	pricingData, err := pricingSvc.parsePricingData(data)
	require.NoError(t, err)
	pricingSvc.pricingData = pricingData
	billingSvc := NewBillingService(&config.Config{}, pricingSvc)

	for _, model := range []string{"gemini-3.6-flash", "gemini-3.6-flash-low", "gemini-3.6-flash-high"} {
		t.Run(model, func(t *testing.T) {
			pricing, err := billingSvc.GetModelPricing(model)
			require.NoError(t, err)
			require.InDelta(t, 1.5e-6, pricing.InputPricePerToken, 1e-12)
			require.InDelta(t, 7.5e-6, pricing.OutputPricePerToken, 1e-12)
			require.InDelta(t, 0.15e-6, pricing.CacheReadPricePerToken, 1e-12)
		})
	}
}

func TestDefaultPricingUsesCurrentCodexAutoReviewBaseRates(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "resources", "model-pricing", "model_prices_and_context_window.json"))
	require.NoError(t, err)

	svc := &PricingService{}
	pricingData, err := svc.parsePricingData(data)
	require.NoError(t, err)
	svc.pricingData = pricingData

	got := svc.GetModelPricing("codex-auto-review")
	require.NotNil(t, got)
	require.InDelta(t, 0.2e-6, got.InputCostPerToken, 1e-12)
	require.InDelta(t, 1.2e-6, got.OutputCostPerToken, 1e-12)
	require.InDelta(t, 0.02e-6, got.CacheReadInputTokenCost, 1e-12)

	// Auto-review is an internal Codex model. Do not infer public GPT-5.6 API
	// service-tier, cache-write, or long-context pricing without an upstream
	// usage contract for this dedicated model.
	require.Zero(t, got.InputCostPerTokenPriority)
	require.Zero(t, got.OutputCostPerTokenPriority)
	require.Zero(t, got.CacheReadInputTokenCostPriority)
	require.Zero(t, got.CacheCreationInputTokenCost)
	require.Zero(t, got.CacheCreationInputTokenCostPriority)
	require.Zero(t, got.LongContextInputTokenThreshold)
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

func TestGetModelPricing_Fable5ExactMatchPrefersPricingData(t *testing.T) {
	svc := &PricingService{
		cfg: &config.Config{},
		pricingData: map[string]*LiteLLMModelPricing{
			"claude-fable-5": {LiteLLMProvider: "anthropic", InputCostPerToken: 1e-05, OutputCostPerToken: 5e-05},
		},
	}

	got := svc.GetModelPricing("claude-fable-5")
	require.NotNil(t, got)
	require.Same(t, svc.pricingData["claude-fable-5"], got)
}

func TestGetModelPricing_Fable5StaticFallbackWhenDataMissing(t *testing.T) {
	// 远程定价库尚未收录 claude-fable-5 时，按官方静态价兜底（$10/$50 per MTok），避免计费为 $0。
	svc := &PricingService{
		cfg:         &config.Config{},
		pricingData: map[string]*LiteLLMModelPricing{},
	}

	for _, name := range []string{"claude-fable-5", "claude-fable-5-20260610", "claude-fable-5[1m]"} {
		got := svc.GetModelPricing(name)
		require.NotNilf(t, got, "expected pricing for %q", name)
		require.InDelta(t, 1e-05, got.InputCostPerToken, 1e-15)
		require.InDelta(t, 5e-05, got.OutputCostPerToken, 1e-15)
		require.InDelta(t, 1.25e-05, got.CacheCreationInputTokenCost, 1e-15)
		require.InDelta(t, 1e-06, got.CacheReadInputTokenCost, 1e-15)
		require.Equal(t, "anthropic", got.LiteLLMProvider)
		require.True(t, got.SupportsPromptCaching)
	}
}

// --- above_XXXk 绝对价字段折算为阈值+倍率 ---

func TestParsePricingData_DerivesLongContextFromAboveTierFields(t *testing.T) {
	svc := &PricingService{}
	data, err := svc.parsePricingData([]byte(`{
		"gpt-above": {"litellm_provider": "openai", "mode": "chat",
			"input_cost_per_token": 5e-06, "output_cost_per_token": 3e-05,
			"cache_read_input_token_cost": 5e-07,
			"input_cost_per_token_above_272k_tokens": 1e-05,
			"output_cost_per_token_above_272k_tokens": 4.5e-05,
			"cache_read_input_token_cost_above_272k_tokens": 1e-06,
			"input_cost_per_token_above_272k_tokens_flex": 5e-06,
			"output_cost_per_token_above_272k_tokens_flex": 2.25e-05},
		"gemini-above": {"litellm_provider": "vertex_ai-language-models", "mode": "chat",
			"input_cost_per_token": 1.25e-06, "output_cost_per_token": 1e-05,
			"input_cost_per_token_above_200k_tokens": 2.5e-06,
			"output_cost_per_token_above_200k_tokens": 1.5e-05},
		"explicit-wins": {"litellm_provider": "openai", "mode": "chat",
			"input_cost_per_token": 5e-06, "output_cost_per_token": 3e-05,
			"long_context_input_cost_multiplier": 1,
			"long_context_output_cost_multiplier": 1,
			"input_cost_per_token_above_272k_tokens": 1e-05,
			"output_cost_per_token_above_272k_tokens": 4.5e-05},
		"no-surcharge": {"litellm_provider": "openai", "mode": "chat",
			"input_cost_per_token": 5e-06, "output_cost_per_token": 3e-05,
			"input_cost_per_token_above_272k_tokens": 5e-06,
			"output_cost_per_token_above_272k_tokens": 3e-05},
		"cache-only-above": {"litellm_provider": "openai", "mode": "chat",
			"input_cost_per_token": 5e-06, "output_cost_per_token": 3e-05,
			"cache_read_input_token_cost_above_272k_tokens": 1e-06},
		"multi-threshold": {"litellm_provider": "openai", "mode": "chat",
			"input_cost_per_token": 1e-06, "output_cost_per_token": 2e-06,
			"input_cost_per_token_above_128k_tokens": 2e-06,
			"input_cost_per_token_above_272k_tokens": 4e-06}
	}`))
	require.NoError(t, err)

	openai := data["gpt-above"]
	require.Equal(t, 272000, openai.LongContextInputTokenThreshold, "阈值取自字段名（_flex 变体不参与）")
	require.InDelta(t, 2.0, openai.LongContextInputCostMultiplier, 1e-12)
	require.InDelta(t, 1.5, openai.LongContextOutputCostMultiplier, 1e-12)

	gemini := data["gemini-above"]
	require.Equal(t, 200000, gemini.LongContextInputTokenThreshold)
	require.InDelta(t, 2.0, gemini.LongContextInputCostMultiplier, 1e-12)
	require.InDelta(t, 1.5, gemini.LongContextOutputCostMultiplier, 1e-12)

	explicit := data["explicit-wins"]
	require.Zero(t, explicit.LongContextInputTokenThreshold, "显式 long_context_* 字段优先，不做折算")
	require.InDelta(t, 1.0, explicit.LongContextInputCostMultiplier, 1e-12)

	require.Zero(t, data["no-surcharge"].LongContextInputTokenThreshold, "above 价不高于基础价视为无附加费")
	require.Zero(t, data["cache-only-above"].LongContextInputTokenThreshold, "仅 cache 侧 above 字段不构成阶梯")
	require.Equal(t, 128000, data["multi-threshold"].LongContextInputTokenThreshold, "多阈值取最小")
}

func TestGetModelPricing_XAIThresholdInclusive(t *testing.T) {
	svc := NewBillingService(&config.Config{}, newStubPricingServiceFromJSON(t, `{
		"grok-4.5": {"litellm_provider": "xai", "mode": "chat",
			"input_cost_per_token": 2e-06, "output_cost_per_token": 6e-06,
			"input_cost_per_token_above_200k_tokens": 4e-06,
			"output_cost_per_token_above_200k_tokens": 1.2e-05}
	}`))
	pricing, err := svc.GetModelPricing("grok-4.5")
	require.NoError(t, err)
	require.Equal(t, 200000, pricing.LongContextInputThreshold)
	require.True(t, pricing.LongContextThresholdInclusive, "xAI 阈值语义为达到即进高档")
}

// F3：显式 long_context 字段以"字段存在"为准——显式 0 也能压住 above 折算，关闭阶梯。
func TestParsePricingData_ExplicitZeroThresholdDisablesLadder(t *testing.T) {
	svc := &PricingService{}
	data, err := svc.parsePricingData([]byte(`{
		"gpt-5.5": {"litellm_provider": "openai", "mode": "chat",
			"input_cost_per_token": 5e-06, "output_cost_per_token": 3e-05,
			"long_context_input_token_threshold": 0,
			"input_cost_per_token_above_272k_tokens": 1e-05,
			"output_cost_per_token_above_272k_tokens": 4.5e-05}
	}`))
	require.NoError(t, err)
	require.Zero(t, data["gpt-5.5"].LongContextInputTokenThreshold)
	require.Zero(t, data["gpt-5.5"].LongContextInputCostMultiplier)
}

// cache 侧 above 档随输入倍率计费、不单独折算；缺基础价的 cache above 字段无法参与计费，
// 该缓存分项按 0 计，属于数据契约违规，必须有哨兵 WARN。服务档变体缺基础价时回落
// 标准基础价，不算孤儿。
func TestParsePricingData_WarnsOrphanCacheTierFields(t *testing.T) {
	logSink, restore := captureStructuredLog(t)
	defer restore()

	svc := &PricingService{}
	data, err := svc.parsePricingData([]byte(`{
		"gemini-orphan": {"litellm_provider": "vertex_ai-language-models", "mode": "chat",
			"input_cost_per_token": 1.25e-06, "output_cost_per_token": 1e-05,
			"cache_read_input_token_cost": 1.25e-07,
			"input_cost_per_token_above_200k_tokens": 2.5e-06,
			"output_cost_per_token_above_200k_tokens": 1.5e-05,
			"cache_read_input_token_cost_above_200k_tokens": 2.5e-07,
			"cache_creation_input_token_cost_above_200k_tokens": 2.5e-07},
		"gemini-complete": {"litellm_provider": "vertex_ai-language-models", "mode": "chat",
			"input_cost_per_token": 1.25e-06, "output_cost_per_token": 1e-05,
			"cache_read_input_token_cost": 1.25e-07,
			"cache_creation_input_token_cost": 1.25e-06,
			"input_cost_per_token_above_200k_tokens": 2.5e-06,
			"output_cost_per_token_above_200k_tokens": 1.5e-05,
			"cache_read_input_token_cost_above_200k_tokens": 2.5e-07,
			"cache_creation_input_token_cost_above_200k_tokens": 2.5e-06},
		"priority-variant-without-own-base": {"litellm_provider": "openai", "mode": "chat",
			"input_cost_per_token": 5e-06, "output_cost_per_token": 3e-05,
			"cache_read_input_token_cost": 5e-07,
			"input_cost_per_token_above_272k_tokens": 1e-05,
			"output_cost_per_token_above_272k_tokens": 4.5e-05,
			"cache_read_input_token_cost_above_272k_tokens_priority": 2e-06},
		"priority-variant-orphan": {"litellm_provider": "openai", "mode": "chat",
			"input_cost_per_token": 5e-06, "output_cost_per_token": 3e-05,
			"cache_creation_input_token_cost_above_272k_tokens_priority": 2.5e-05},
		"hourly-tier-with-5m-base": {"litellm_provider": "anthropic", "mode": "chat",
			"input_cost_per_token": 3e-06, "output_cost_per_token": 1.5e-05,
			"cache_creation_input_token_cost": 3.75e-06,
			"cache_creation_input_token_cost_above_1hr_above_200k_tokens": 1.2e-05},
		"hourly-tier-orphan": {"litellm_provider": "anthropic", "mode": "chat",
			"input_cost_per_token": 3e-06, "output_cost_per_token": 1.5e-05,
			"cache_creation_input_token_cost_above_1hr_above_200k_tokens": 1.2e-05}
	}`))
	require.NoError(t, err)

	require.Equal(t, 200000, data["gemini-orphan"].LongContextInputTokenThreshold, "孤儿 cache 字段不影响 input/output 阶梯折算")
	require.Zero(t, data["gemini-orphan"].CacheCreationInputTokenCost)
	require.InDelta(t, 1.25e-6, data["gemini-complete"].CacheCreationInputTokenCost, 1e-12)

	require.True(t, logSink.ContainsMessageAtLevel("gemini-orphan(cache_creation_input_token_cost_above_200k_tokens)", "warn"))
	require.True(t, logSink.ContainsMessage("priority-variant-orphan(cache_creation_input_token_cost_above_272k_tokens_priority)"))
	require.True(t, logSink.ContainsMessage("hourly-tier-orphan(cache_creation_input_token_cost_above_1hr_above_200k_tokens)"))
	require.False(t, logSink.ContainsMessage("gemini-complete"))
	require.False(t, logSink.ContainsMessage("priority-variant-without-own-base"))
	require.False(t, logSink.ContainsMessage("hourly-tier-with-5m-base"), "1h 档缺 above_1hr 基础价时计费回落 5m 价，不算孤儿")
}

// 基础价与 above 档来自不同价格版本时（如基础价被手工 pin、above 档随上游更新）会折算出
// 只有一侧带附加费的阶梯，必须有哨兵 WARN；显式 long_context_* 字段是部署方意图，不告警。
func TestParsePricingData_WarnsLopsidedLongContextLadder(t *testing.T) {
	logSink, restore := captureStructuredLog(t)
	defer restore()

	svc := &PricingService{}
	data, err := svc.parsePricingData([]byte(`{
		"mixed-versions": {"litellm_provider": "openai", "mode": "chat",
			"input_cost_per_token": 5e-06, "output_cost_per_token": 3e-05,
			"input_cost_per_token_above_272k_tokens": 8e-06,
			"output_cost_per_token_above_272k_tokens": 3e-05},
		"consistent": {"litellm_provider": "openai", "mode": "chat",
			"input_cost_per_token": 4e-06, "output_cost_per_token": 2e-05,
			"input_cost_per_token_above_272k_tokens": 8e-06,
			"output_cost_per_token_above_272k_tokens": 3e-05},
		"explicit-input-only": {"litellm_provider": "openai", "mode": "chat",
			"input_cost_per_token": 4e-06, "output_cost_per_token": 2e-05,
			"long_context_input_token_threshold": 272000,
			"long_context_input_cost_multiplier": 2}
	}`))
	require.NoError(t, err)

	require.Equal(t, 272000, data["mixed-versions"].LongContextInputTokenThreshold, "单侧阶梯仍按折算结果计费，只告警不丢弃")
	require.InDelta(t, 1.6, data["mixed-versions"].LongContextInputCostMultiplier, 1e-12)
	require.True(t, logSink.ContainsMessageAtLevel("mixed-versions(input x1.60, output x1.00)", "warn"))
	require.False(t, logSink.ContainsMessage("consistent"))
	require.False(t, logSink.ContainsMessage("explicit-input-only"))
}

// 出厂回退快照必须满足数据契约：没有孤儿 cache above 字段、没有单侧阶梯，且 Gemini pro 系的
// 缓存写入基础价等于标准输入价（含 priority 变体）。快照是随目录同步刷新的文本，这里防止刷新时静默回退。
func TestDefaultCatalogSnapshot_CacheTierContract(t *testing.T) {
	logSink, restore := captureStructuredLog(t)
	defer restore()

	body, err := os.ReadFile(filepath.Join("..", "..", "resources", "model-pricing", "model_prices_and_context_window.json"))
	require.NoError(t, err)

	var rawEntries map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(body, &rawEntries))
	for name, raw := range rawEntries {
		require.Empty(t, orphanCacheTierFields(raw), "快照条目 %s 带孤儿 cache above 字段", name)
	}

	svc := &PricingService{}
	data, err := svc.parsePricingData(body)
	require.NoError(t, err)
	require.False(t, logSink.ContainsMessage("carry cache above-tier prices"), "快照不应触发孤儿 cache 字段哨兵")
	require.False(t, logSink.ContainsMessage("one-sided long-context ladder"), "快照不应触发单侧阶梯哨兵")
	for _, model := range []string{
		"gemini-2.5-pro", "gemini-3-pro-preview", "gemini-3.1-pro-preview",
		"gemini-3.1-pro-high", "gemini-3.1-pro-low", "gemini-3.1-pro-preview-customtools",
	} {
		pricing := data[model]
		require.NotNil(t, pricing, model)
		require.Positive(t, pricing.InputCostPerToken, model)
		require.InDelta(t, pricing.InputCostPerToken, pricing.CacheCreationInputTokenCost, 1e-15, "%s 缓存写入基础价应等于标准输入价", model)
		require.Equal(t, 200000, pricing.LongContextInputTokenThreshold, model)
		if pricing.InputCostPerTokenPriority > 0 {
			require.InDelta(t, pricing.InputCostPerTokenPriority, pricing.CacheCreationInputTokenCostPriority, 1e-15, "%s priority 缓存写入价应等于 priority 输入价", model)
		}
	}
}

// F1：显式字段只写了一侧倍率时，缺失侧按 1 计而不是乘 0 免费。
func TestCalculateCost_PartialLongContextMultiplierDefaultsToOne(t *testing.T) {
	tokens := UsageTokens{InputTokens: 300000, OutputTokens: 1000, CacheReadTokens: 10000}

	t.Run("only input multiplier", func(t *testing.T) {
		svc := NewBillingService(&config.Config{}, newStubPricingServiceFromJSON(t, `{
			"partial-in": {"litellm_provider": "openai", "mode": "chat",
				"input_cost_per_token": 2e-06, "output_cost_per_token": 1e-05,
				"cache_read_input_token_cost": 2e-07,
				"long_context_input_token_threshold": 272000,
				"long_context_input_cost_multiplier": 2.0}
		}`))
		cost, err := svc.CalculateCost("partial-in", tokens, 1.0)
		require.NoError(t, err)
		require.True(t, cost.LongContextBillingApplied)
		require.InDelta(t, 300000*2e-6*2, cost.InputCost, 1e-10)
		require.InDelta(t, 1000*1e-5, cost.OutputCost, 1e-10, "缺失的 output 倍率按 1 计，不得为 0")
		require.InDelta(t, 10000*2e-7*2, cost.CacheReadCost, 1e-10)
	})

	t.Run("only output multiplier", func(t *testing.T) {
		svc := NewBillingService(&config.Config{}, newStubPricingServiceFromJSON(t, `{
			"partial-out": {"litellm_provider": "openai", "mode": "chat",
				"input_cost_per_token": 2e-06, "output_cost_per_token": 1e-05,
				"cache_read_input_token_cost": 2e-07,
				"long_context_input_token_threshold": 272000,
				"long_context_output_cost_multiplier": 1.5}
		}`))
		cost, err := svc.CalculateCost("partial-out", tokens, 1.0)
		require.NoError(t, err)
		require.True(t, cost.LongContextBillingApplied)
		require.InDelta(t, 300000*2e-6, cost.InputCost, 1e-10, "缺失的 input 倍率按 1 计，不得为 0")
		require.InDelta(t, 1000*1e-5*1.5, cost.OutputCost, 1e-10)
		require.InDelta(t, 10000*2e-7, cost.CacheReadCost, 1e-10, "cache_read 跟随 input 倍率，同样按 1 计")
	})
}

// 行为声明：目录带 above_200k 的 Claude sonnet 条目同样获得数据驱动的整单阶梯
// （与 Anthropic 官方 1M 长上下文定价一致），受分组长上下文开关约束。
func TestCalculateCost_ClaudeSonnetCatalogLadderIsDataDriven(t *testing.T) {
	svc := NewBillingService(&config.Config{}, newStubPricingServiceFromJSON(t, `{
		"claude-sonnet-4-5": {"litellm_provider": "anthropic", "mode": "chat",
			"input_cost_per_token": 3e-06, "output_cost_per_token": 1.5e-05,
			"cache_read_input_token_cost": 3e-07,
			"input_cost_per_token_above_200k_tokens": 6e-06,
			"output_cost_per_token_above_200k_tokens": 2.25e-05,
			"cache_read_input_token_cost_above_200k_tokens": 6e-07}
	}`))

	pricing, err := svc.GetModelPricing("claude-sonnet-4-5")
	require.NoError(t, err)
	require.Equal(t, 200000, pricing.LongContextInputThreshold)
	require.False(t, pricing.LongContextThresholdInclusive, "anthropic 为严格大于")

	over := UsageTokens{InputTokens: 250000, OutputTokens: 1000}
	cost, err := svc.CalculateCost("claude-sonnet-4-5", over, 1.0)
	require.NoError(t, err)
	require.True(t, cost.LongContextBillingApplied)
	require.InDelta(t, 250000*3e-6*2, cost.InputCost, 1e-10)
	require.InDelta(t, 1000*1.5e-5*1.5, cost.OutputCost, 1e-10)

	under := UsageTokens{InputTokens: 200000, OutputTokens: 1000}
	cost, err = svc.CalculateCost("claude-sonnet-4-5", under, 1.0)
	require.NoError(t, err)
	require.False(t, cost.LongContextBillingApplied, "恰好 200000 不进高档（严格大于）")
}
