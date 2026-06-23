package service

import (
	"context"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

// previewFakeRemoteClient 是 PreviewRemotePricing 的离线远程客户端桩。
type previewFakeRemoteClient struct{ body []byte }

func (f previewFakeRemoteClient) FetchPricingJSON(_ context.Context, _ string) ([]byte, error) {
	return f.body, nil
}
func (f previewFakeRemoteClient) FetchHashText(_ context.Context, _ string) (string, error) {
	return "", nil
}

func TestListBuiltinPricing_ExposesLiteLLMAndCNY(t *testing.T) {
	svc := &PricingService{
		pricingData: map[string]*LiteLLMModelPricing{
			// 0.000003/token → 3 /百万；0.000015/token → 15 /百万
			"gpt-5.3": {InputCostPerToken: 0.000003, OutputCostPerToken: 0.000015, LiteLLMProvider: "openai", Mode: "chat"},
		},
	}

	list := svc.ListBuiltinPricing()
	byModel := make(map[string]BuiltinPricingEntry, len(list))
	for _, e := range list {
		byModel[strings.ToLower(e.Model)] = e
	}

	// ③ LiteLLM 条目:USD,每百万换算正确
	gpt, ok := byModel["gpt-5.3"]
	require.True(t, ok, "Claude/GPT 类内置模型应出现在 builtin 列表")
	require.Equal(t, CurrencyUSD, gpt.Currency)
	require.InDelta(t, 3.0, gpt.InputPerM, 1e-9)
	require.InDelta(t, 15.0, gpt.OutputPerM, 1e-9)

	// ② 国产¥表也摊平进来(以 kimi-k2.6 为锚)
	kimi, ok := byModel["kimi-k2.6"]
	require.True(t, ok, "国产¥表应出现在 builtin 列表")
	require.Equal(t, CurrencyCNY, kimi.Currency)
	require.InDelta(t, 6.5, kimi.InputPerM, 1e-9)
	require.Contains(t, kimi.Source, "cny")

	// 按模型名升序
	for i := 1; i < len(list); i++ {
		require.LessOrEqual(t, list[i-1].Model, list[i].Model)
	}
}

func TestListBuiltinPricing_CNYWinsOverLiteLLMDuplicate(t *testing.T) {
	svc := &PricingService{
		pricingData: map[string]*LiteLLMModelPricing{
			// 与国产¥表同名:应被去重,保留 CNY 那条(真实计费口径)
			"kimi-k2.6": {InputCostPerToken: 0.001, OutputCostPerToken: 0.002, LiteLLMProvider: "openai"},
		},
	}
	for _, e := range svc.ListBuiltinPricing() {
		if strings.ToLower(e.Model) == "kimi-k2.6" {
			require.Equal(t, CurrencyCNY, e.Currency)
			require.Contains(t, e.Source, "cny")
			return
		}
	}
	t.Fatal("kimi-k2.6 未出现在 builtin 列表")
}

func TestPreviewRemotePricing_DiffCounts(t *testing.T) {
	cfg := &config.Config{}
	cfg.Pricing.RemoteURL = "https://example.com/prices.json"
	// URLAllowlist 默认 Enabled=false → 走 ValidateURLFormat,https 直接通过

	remote := `{
		"model-keep":      {"input_cost_per_token":0.000001,"output_cost_per_token":0.000002},
		"model-changed":   {"input_cost_per_token":0.000004,"output_cost_per_token":0.000008},
		"Model-CacheOnly": {"input_cost_per_token":0.000001,"output_cost_per_token":0.000002,"cache_read_input_token_cost":0.0000009},
		"model-added":     {"input_cost_per_token":0.000005,"output_cost_per_token":0.000010}
	}`

	svc := &PricingService{
		cfg:          cfg,
		remoteClient: previewFakeRemoteClient{body: []byte(remote)},
		pricingData: map[string]*LiteLLMModelPricing{
			"model-keep":      {InputCostPerToken: 0.000001, OutputCostPerToken: 0.000002},
			"model-changed":   {InputCostPerToken: 0.000003, OutputCostPerToken: 0.000008}, // input 不同
			"Model-CacheOnly": {InputCostPerToken: 0.000001, OutputCostPerToken: 0.000002}, // 仅缓存读取价新增 → 必须算 changed
			"model-removed":   {InputCostPerToken: 0.000009, OutputCostPerToken: 0.000009},
		},
	}

	pv, err := svc.PreviewRemotePricing(context.Background())
	require.NoError(t, err)
	require.Equal(t, 4, pv.RemoteCount)
	require.Equal(t, 4, pv.CurrentCount)
	require.Equal(t, 1, pv.Added, "model-added")
	require.Equal(t, 2, pv.Changed, "model-changed + 仅缓存价变化的 Model-CacheOnly")
	require.Equal(t, 1, pv.Removed, "model-removed")
	require.False(t, pv.Truncated)
	require.Len(t, pv.Changes, 4)

	// 缓存价变化被捕获,且明细 model 名保留原始大小写。
	var cacheCh *PricingChange
	for i := range pv.Changes {
		if strings.EqualFold(pv.Changes[i].Model, "Model-CacheOnly") {
			cacheCh = &pv.Changes[i]
		}
	}
	require.NotNil(t, cacheCh, "仅缓存价变化应出现在 changes")
	require.Equal(t, "Model-CacheOnly", cacheCh.Model, "model 名应保留原始大小写")
	require.InDelta(t, 0.9, cacheCh.NewCache, 1e-9) // 0.0000009 * 1e6
	require.InDelta(t, 0.0, cacheCh.OldCache, 1e-9)
}
