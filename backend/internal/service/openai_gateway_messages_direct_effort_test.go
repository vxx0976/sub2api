//go:build unit

package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// 存量 CN 直通（usesLegacyCNAnthropicDirect，未配 api_protocol）也必须把推理档写进
// result.ReasoningEffort：计费的 reasoning_effort_multipliers 与 usage_logs.reasoning_effort
// 都只认这个字段。之前 forwardAnthropicDirect 的所有 return 都不带它，价卡倍率恒按 1x。

func legacyCNDirectEffortTestAccount(platform string) *Account {
	return &Account{
		ID:          9301,
		Name:        "legacy-cn-direct",
		Platform:    platform,
		Type:        AccountTypeAPIKey,
		Concurrency: 1,
		Credentials: map[string]any{
			"api_key":  "sk-test",
			"base_url": "http://cn-upstream.example",
		},
	}
}

func legacyCNDirectEffortResponse(model string, mode string) *http.Response {
	switch mode {
	case "stream", "buffered-sse":
		sse := "event: message_start\n" +
			`data: {"type":"message_start","message":{"id":"msg_1","type":"message","role":"assistant","model":"` + model + `","content":[],"stop_reason":null,"usage":{"input_tokens":1000000,"output_tokens":1}}}` + "\n\n" +
			"event: content_block_start\n" +
			`data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}` + "\n\n" +
			"event: content_block_delta\n" +
			`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"ok"}}` + "\n\n" +
			"event: content_block_stop\n" +
			`data: {"type":"content_block_stop","index":0}` + "\n\n" +
			"event: message_delta\n" +
			`data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":2}}` + "\n\n" +
			"event: message_stop\n" +
			`data: {"type":"message_stop"}` + "\n\n"
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(sse)),
		}
	default:
		body := `{"id":"msg_1","type":"message","role":"assistant","model":"` + model + `","content":[{"type":"text","text":"ok"}],"stop_reason":"end_turn","usage":{"input_tokens":1000000,"output_tokens":2}}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(body)),
		}
	}
}

func TestLegacyCNAnthropicDirectSetsReasoningEffort(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		name       string
		platform   string
		model      string
		extra      string
		wantEffort string
	}{
		// 显式 output_config.effort：原样记录。
		{"deepseek explicit max", PlatformDeepseek, "deepseek-v4-pro", `,"output_config":{"effort":"max"}`, "max"},
		{"zhipu explicit max", PlatformZhipu, "glm-5.2", `,"output_config":{"effort":"max"}`, "max"},
		// thinking 已启用且无 effort：passback-required 国产模型兜底 high。
		{"zhipu thinking fallback", PlatformZhipu, "glm-5.2", `,"thinking":{"type":"enabled","budget_tokens":1024}`, "high"},
		{"kimi thinking fallback", PlatformKimi, "kimi-k2.6", `,"thinking":{"type":"enabled","budget_tokens":1024}`, "high"},
		// DeepSeek 不做兜底（有原生 effort），与原生路径口径一致。
		{"deepseek thinking no fallback", PlatformDeepseek, "deepseek-v4-pro", `,"thinking":{"type":"enabled","budget_tokens":1024}`, ""},
		// 无任何 effort 信号：保持 nil，倍率 1x、不新增记录。
		{"zhipu no signal", PlatformZhipu, "glm-5.2", ``, ""},
		{"deepseek no signal", PlatformDeepseek, "deepseek-v4-pro", ``, ""},
	}
	for _, tc := range cases {
		for _, mode := range []string{"stream", "buffered", "buffered-sse"} {
			t.Run(fmt.Sprintf("%s/%s", tc.name, mode), func(t *testing.T) {
				stream := mode == "stream"
				body := []byte(fmt.Sprintf(`{"model":%q,"stream":%t,"max_tokens":32,"messages":[{"role":"user","content":"hello"}]%s}`, tc.model, stream, tc.extra))
				account := legacyCNDirectEffortTestAccount(tc.platform)
				require.True(t, usesLegacyCNAnthropicDirect(account))
				upstream := &httpUpstreamRecorder{resp: legacyCNDirectEffortResponse(tc.model, mode)}
				svc := &OpenAIGatewayService{cfg: rawChatCompletionsTestConfig(), httpUpstream: upstream}

				result, err := svc.ForwardAsAnthropic(context.Background(), adaptiveProtocolTestContext("/v1/messages", body), account, body, "", "")
				require.NoError(t, err)
				require.NotNil(t, upstream.lastReq)
				require.True(t, strings.HasSuffix(upstream.lastReq.URL.Path, "/messages"), upstream.lastReq.URL.Path)
				require.NotNil(t, result)
				require.Equal(t, 1_000_000, result.Usage.InputTokens)
				if tc.wantEffort == "" {
					require.Nil(t, result.ReasoningEffort)
					return
				}
				require.NotNil(t, result.ReasoningEffort)
				require.Equal(t, tc.wantEffort, *result.ReasoningEffort)
			})
		}
	}
}

// 价卡 reasoning_effort_multipliers {"max": 2}：直通结果带 max 时 token 成本翻倍，无 effort 时 1x。
func TestLegacyCNAnthropicDirectReasoningEffortMultiplierBilling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	billing := NewBillingService(rawChatCompletionsTestConfig(), nil)
	for _, tc := range []struct {
		platform string
		model    string
	}{
		{PlatformDeepseek, "deepseek-v4-pro"},
		{PlatformZhipu, "glm-5.2"},
	} {
		t.Run(tc.platform, func(t *testing.T) {
			costFor := func(extra string) (float64, *OpenAIForwardResult) {
				body := []byte(fmt.Sprintf(`{"model":%q,"stream":false,"max_tokens":32,"messages":[{"role":"user","content":"hello"}]%s}`, tc.model, extra))
				upstream := &httpUpstreamRecorder{resp: legacyCNDirectEffortResponse(tc.model, "buffered")}
				svc := &OpenAIGatewayService{cfg: rawChatCompletionsTestConfig(), httpUpstream: upstream}
				result, err := svc.ForwardAsAnthropic(context.Background(), adaptiveProtocolTestContext("/v1/messages", body), legacyCNDirectEffortTestAccount(tc.platform), body, "", "")
				require.NoError(t, err)
				require.NotNil(t, result)
				group := &Group{ID: 1, Platform: tc.platform, ModelPricing: []ChannelModelPricing{{
					Models: []string{result.BillingModel}, BillingMode: BillingModeToken,
					InputPrice: testPtrFloat64(1e-6), OutputPrice: testPtrFloat64(0),
					ReasoningEffortMultipliers: map[string]float64{"max": 2},
				}}}
				cost, err := billing.CalculateTokenCostForRequest(TokenCostRequest{
					Ctx: context.Background(), Model: result.BillingModel, Group: group,
					Tokens:          UsageTokens{InputTokens: result.Usage.InputTokens, OutputTokens: result.Usage.OutputTokens},
					ReasoningEffort: optionalStringValue(result.ReasoningEffort), RateMultiplier: 1,
					Resolver: NewModelPricingResolver(nil, billing),
				})
				require.NoError(t, err)
				return cost.ActualCost, result
			}

			maxCost, maxResult := costFor(`,"output_config":{"effort":"max"}`)
			baseCost, baseResult := costFor(``)
			require.Equal(t, "max", optionalStringValue(maxResult.ReasoningEffort))
			require.Nil(t, baseResult.ReasoningEffort)
			require.InDelta(t, 1.0, baseCost, 1e-12)
			require.InDelta(t, 2.0, maxCost, 1e-12)
		})
	}
}
