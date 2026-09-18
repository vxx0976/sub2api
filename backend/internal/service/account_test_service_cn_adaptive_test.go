//go:build unit

package service

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func adaptiveCNAccountTestAccount(id int64, platform string) *Account {
	return &Account{
		ID:          id,
		Name:        "adaptive-cn-test",
		Platform:    platform,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Concurrency: 1,
		Credentials: map[string]any{
			"api_key":      "sk-adaptive-test",
			"api_protocol": APIProtocolAdaptive,
			"api_base_urls": map[string]any{
				APIProtocolChatCompletions: "http://chat.example/v1",
				APIProtocolAnthropic:       "http://anthropic.example",
				APIProtocolResponses:       "http://responses.example",
			},
		},
	}
}

func adaptiveCNAccountTestService(account *Account, responses ...*http.Response) (*AccountTestService, *httpUpstreamRecorder) {
	repo := &openAIAccountTestRepo{
		mockAccountRepoForGemini: mockAccountRepoForGemini{
			accountsByID: map[int64]*Account{account.ID: account},
		},
	}
	upstream := &httpUpstreamRecorder{responses: responses}
	return &AccountTestService{
		accountRepo:  repo,
		httpUpstream: upstream,
		cfg:          rawChatCompletionsTestConfig(),
	}, upstream
}

func adaptiveCNChatTestResponse() *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body: io.NopCloser(strings.NewReader(`data: {"choices":[{"delta":{"content":"chat ok"},"finish_reason":"stop"}]}

data: [DONE]

`)),
	}
}

func adaptiveCNAnthropicTestResponse() *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body: io.NopCloser(strings.NewReader(`data: {"type":"content_block_delta","delta":{"text":"anthropic ok"}}

data: {"type":"message_stop"}

`)),
	}
}

func adaptiveCNResponsesTestResponse() *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body: io.NopCloser(strings.NewReader(`data: {"type":"response.output_text.delta","delta":"responses ok"}

data: {"type":"response.completed"}

`)),
	}
}

func TestAccountTestService_AdaptiveChatOnlyProvidersTestChatAndAnthropicEndpoints(t *testing.T) {
	account := adaptiveCNAccountTestAccount(301, PlatformZhipu)
	svc, upstream := adaptiveCNAccountTestService(
		account,
		adaptiveCNChatTestResponse(),
		adaptiveCNAnthropicTestResponse(),
	)
	c, recorder := newTestContext()

	err := svc.TestAccountConnection(c, account.ID, "glm-4.7", "hello", AccountTestModeDefault)

	require.NoError(t, err)
	require.Len(t, upstream.requests, 2)
	require.Equal(t, "http://chat.example/v1/chat/completions", upstream.requests[0].URL.String())
	require.Equal(t, "http://anthropic.example/v1/messages", upstream.requests[1].URL.String())
	require.Equal(t, "Bearer sk-adaptive-test", upstream.requests[0].Header.Get("Authorization"))
	require.Equal(t, "sk-adaptive-test", upstream.requests[1].Header.Get("x-api-key"))
	require.Equal(t, 1, strings.Count(recorder.Body.String(), `"type":"test_start"`))
	require.Equal(t, 1, strings.Count(recorder.Body.String(), `"type":"test_complete"`))
	require.Contains(t, recorder.Body.String(), "已通过原生 /v1/messages 验证")
}

func TestAccountTestService_AdaptiveDeepSeekAlsoTestsResponsesEndpoint(t *testing.T) {
	account := adaptiveCNAccountTestAccount(302, PlatformDeepseek)
	svc, upstream := adaptiveCNAccountTestService(
		account,
		adaptiveCNChatTestResponse(),
		adaptiveCNAnthropicTestResponse(),
		adaptiveCNResponsesTestResponse(),
	)
	c, recorder := newTestContext()

	err := svc.TestAccountConnection(c, account.ID, "deepseek-chat", "", AccountTestModeDefault)

	require.NoError(t, err)
	require.Len(t, upstream.requests, 3)
	require.Equal(t, "http://responses.example/responses", upstream.requests[2].URL.String())
	require.Equal(t, HTTPUpstreamProfileOpenAI, HTTPUpstreamProfileFromContext(upstream.requests[2].Context()))
	require.Equal(t, "Bearer sk-adaptive-test", upstream.requests[2].Header.Get("Authorization"))
	require.True(t, gjson.GetBytes(upstream.bodies[2], "stream").Bool())
	require.False(t, gjson.GetBytes(upstream.bodies[2], "store").Bool())
	require.False(t, gjson.GetBytes(upstream.bodies[2], "instructions").Exists())
	require.Equal(t, 1, strings.Count(recorder.Body.String(), `"type":"test_complete"`))
	require.Contains(t, recorder.Body.String(), "已通过原生 /responses 验证")
}

func TestAccountTestService_AdaptiveKimiAlsoTestsResponsesEndpoint(t *testing.T) {
	account := adaptiveCNAccountTestAccount(306, PlatformKimi)
	svc, upstream := adaptiveCNAccountTestService(
		account,
		adaptiveCNChatTestResponse(),
		adaptiveCNAnthropicTestResponse(),
		adaptiveCNResponsesTestResponse(),
	)
	c, recorder := newTestContext()

	err := svc.TestAccountConnection(c, account.ID, "k3-256k", "", AccountTestModeDefault)

	require.NoError(t, err)
	require.Len(t, upstream.requests, 3)
	require.Equal(t, "http://responses.example/v1/responses", upstream.requests[2].URL.String())
	require.Equal(t, HTTPUpstreamProfileOpenAI, HTTPUpstreamProfileFromContext(upstream.requests[2].Context()))
	require.Equal(t, "Bearer sk-adaptive-test", upstream.requests[2].Header.Get("Authorization"))
	require.True(t, gjson.GetBytes(upstream.bodies[2], "stream").Bool())
	require.False(t, gjson.GetBytes(upstream.bodies[2], "store").Bool())
	require.False(t, gjson.GetBytes(upstream.bodies[2], "instructions").Exists())
	require.Equal(t, 1, strings.Count(recorder.Body.String(), `"type":"test_complete"`))
	require.Contains(t, recorder.Body.String(), "已通过原生 /responses 验证")
}

func TestAccountTestService_AdaptiveStopsAndNamesFailingEndpoint(t *testing.T) {
	account := adaptiveCNAccountTestAccount(303, PlatformDeepseek)
	svc, upstream := adaptiveCNAccountTestService(
		account,
		adaptiveCNChatTestResponse(),
		newJSONResponse(http.StatusNotFound, `{"error":{"message":"missing messages route"}}`),
	)
	c, recorder := newTestContext()

	err := svc.TestAccountConnection(c, account.ID, "deepseek-chat", "", AccountTestModeDefault)

	require.Error(t, err)
	require.Contains(t, err.Error(), "Adaptive Anthropic endpoint returned 404")
	require.Len(t, upstream.requests, 2)
	require.Contains(t, recorder.Body.String(), `"type":"error"`)
	require.NotContains(t, recorder.Body.String(), `"type":"test_complete"`)
}

func TestAccountTestService_AdaptiveRejectsInvalidAnthropicSuccessBody(t *testing.T) {
	account := adaptiveCNAccountTestAccount(305, PlatformKimi)
	svc, upstream := adaptiveCNAccountTestService(
		account,
		adaptiveCNChatTestResponse(),
		newJSONResponse(http.StatusOK, `<html>not an Anthropic stream</html>`),
	)
	c, recorder := newTestContext()

	err := svc.TestAccountConnection(c, account.ID, "kimi-k2.5", "", AccountTestModeDefault)

	require.Error(t, err)
	require.Contains(t, err.Error(), "Adaptive Anthropic stream ended before message_stop")
	require.Len(t, upstream.requests, 2)
	require.Contains(t, recorder.Body.String(), `"type":"error"`)
	require.NotContains(t, recorder.Body.String(), `"type":"test_complete"`)
}

func TestAccountTestService_FixedCNChatProtocolStillTestsOnlyChatEndpoint(t *testing.T) {
	account := adaptiveCNAccountTestAccount(304, PlatformZhipu)
	account.Credentials["api_protocol"] = APIProtocolChatCompletions
	account.Credentials["base_url"] = "http://fixed-chat.example/v1"
	svc, upstream := adaptiveCNAccountTestService(account, adaptiveCNChatTestResponse())
	c, recorder := newTestContext()

	err := svc.TestAccountConnection(c, account.ID, "glm-4.7", "", AccountTestModeDefault)

	require.NoError(t, err)
	require.Len(t, upstream.requests, 1)
	require.Equal(t, "http://fixed-chat.example/v1/chat/completions", upstream.requests[0].URL.String())
	require.Equal(t, 1, strings.Count(recorder.Body.String(), `"type":"test_complete"`))
}

func anthropicProtocolCNAccount(id int64, platform string, credentials map[string]any) *Account {
	base := map[string]any{
		"api_key":      "sk-anthropic-test",
		"api_protocol": APIProtocolAnthropic,
	}
	for key, value := range credentials {
		base[key] = value
	}
	return &Account{
		ID:          id,
		Name:        "anthropic-protocol-cn-test",
		Platform:    platform,
		Type:        AccountTypeAPIKey,
		Status:      StatusActive,
		Concurrency: 1,
		Credentials: base,
	}
}

func TestAccountTestService_AnthropicProtocolProbesNativeEndpointWithoutBetaQuery(t *testing.T) {
	account := anthropicProtocolCNAccount(311, PlatformZhipu, map[string]any{
		"base_url": "https://open.bigmodel.cn/api/anthropic",
	})
	svc, upstream := adaptiveCNAccountTestService(account, adaptiveCNAnthropicTestResponse())
	c, recorder := newTestContext()

	err := svc.TestAccountConnection(c, account.ID, "glm-4.7", "", AccountTestModeDefault)

	require.NoError(t, err)
	require.Len(t, upstream.requests, 1)
	req := upstream.requests[0]
	// Native Anthropic path without the ?beta=true suffix the generic Claude tester appends.
	require.Equal(t, "https://open.bigmodel.cn/api/anthropic/v1/messages", req.URL.String())
	require.Empty(t, req.URL.RawQuery)
	require.Equal(t, "sk-anthropic-test", req.Header.Get("x-api-key"))
	require.Equal(t, "2023-06-01", req.Header.Get("anthropic-version"))
	require.Contains(t, recorder.Body.String(), `"type":"test_complete"`)
}

func TestAccountTestService_AnthropicProtocolFallsBackToProviderDefaultNotAnthropicDotCom(t *testing.T) {
	// base_url intentionally absent: forwarding resolves the per-platform default
	// Anthropic endpoint. The old fall-through probed https://api.anthropic.com
	// with the provider's API key.
	account := anthropicProtocolCNAccount(312, PlatformZhipu, nil)
	svc, upstream := adaptiveCNAccountTestService(account, adaptiveCNAnthropicTestResponse())
	c, _ := newTestContext()

	err := svc.TestAccountConnection(c, account.ID, "glm-4.7", "", AccountTestModeDefault)

	require.NoError(t, err)
	require.Len(t, upstream.requests, 1)
	require.Equal(t, "https://open.bigmodel.cn/api/anthropic/v1/messages", upstream.requests[0].URL.String())
}

func TestAccountTestService_AnthropicProtocolRejectsOpenAICompatBaseURL(t *testing.T) {
	account := anthropicProtocolCNAccount(313, PlatformZhipu, map[string]any{
		"base_url": "https://open.bigmodel.cn/api/paas/v4",
	})
	svc, upstream := adaptiveCNAccountTestService(account)
	c, recorder := newTestContext()

	err := svc.TestAccountConnection(c, account.ID, "glm-4.7", "", AccountTestModeDefault)

	require.Error(t, err)
	// Fails fast locally: no upstream request with the wrong endpoint shape.
	require.Empty(t, upstream.requests)
	require.Contains(t, recorder.Body.String(), "looks like an OpenAI-compatible endpoint")
	require.Contains(t, recorder.Body.String(), "https://open.bigmodel.cn/api/anthropic")
}

func TestAccountTestService_AnthropicProtocol401MarksAccountError(t *testing.T) {
	account := anthropicProtocolCNAccount(314, PlatformKimi, map[string]any{
		"base_url": "https://api.moonshot.cn/anthropic",
	})
	svc, _ := adaptiveCNAccountTestService(
		account,
		newJSONResponse(http.StatusUnauthorized, `{"error":{"message":"invalid key"}}`),
	)
	c, _ := newTestContext()

	err := svc.TestAccountConnection(c, account.ID, "kimi-k2.5", "", AccountTestModeDefault)

	require.Error(t, err)
	require.Contains(t, err.Error(), "Anthropic endpoint returned 401")
	repo := svc.accountRepo.(*openAIAccountTestRepo)
	require.Equal(t, account.ID, repo.setErrorID)
}

// 自适应账号的单条通道 401 不得停掉整个账号：adaptive 有三条独立通道，
// 一条端点配错（2026-09-18 线上：编程套餐 key 配了 PayG Anthropic 端点）
// 不应连同正常的 chat_completions 一起 SetError。
func TestAccountTestService_Adaptive401DoesNotDisableAccount(t *testing.T) {
	account := adaptiveCNAccountTestAccount(316, PlatformKimi)
	svc, _ := adaptiveCNAccountTestService(
		account,
		adaptiveCNChatTestResponse(),
		newJSONResponse(http.StatusUnauthorized, `{"error":{"message":"Invalid Authentication","type":"invalid_authentication_error"}}`),
	)
	c, recorder := newTestContext()

	err := svc.TestAccountConnection(c, account.ID, "kimi-k2.6", "", AccountTestModeDefault)

	require.Error(t, err)
	require.Contains(t, recorder.Body.String(), "Adaptive Anthropic endpoint returned 401")
	repo := svc.accountRepo.(*openAIAccountTestRepo)
	require.Zero(t, repo.setErrorID, "单条通道探测失败不得把账号整体置 error")
}

// account_mode 缺失时按 base_url 推断接入模式：编程套餐 base_url 必须解析出编程套餐的
// Anthropic / Responses 端点，否则 adaptive 会去打 PayG 端点而恒 401。
func TestEffectiveCNAccountModeInfersCodingFromBaseURL(t *testing.T) {
	codingKimi := &Account{
		ID: 400, Platform: PlatformKimi, Type: AccountTypeAPIKey, Status: StatusActive,
		Credentials: map[string]any{"api_key": "sk-x", "base_url": "https://api.kimi.com/coding/v1"},
	}
	require.Equal(t, AccountModeCoding, codingKimi.effectiveCNAccountMode())
	require.Equal(t, DefaultKimiCodingAnthropicBaseURL, codingKimi.defaultCNProtocolBaseURL(APIProtocolAnthropic))
	require.Equal(t, DefaultKimiCodingBaseURL, codingKimi.defaultCNProtocolBaseURL(APIProtocolResponses))

	paygKimi := &Account{
		ID: 401, Platform: PlatformKimi, Type: AccountTypeAPIKey, Status: StatusActive,
		Credentials: map[string]any{"api_key": "sk-x", "base_url": "https://api.moonshot.cn/v1"},
	}
	require.Equal(t, AccountModePayG, paygKimi.effectiveCNAccountMode())
	require.Equal(t, DefaultKimiPayGAnthropicBaseURL, paygKimi.defaultCNProtocolBaseURL(APIProtocolAnthropic))

	// 显式 account_mode 优先于推断。
	explicitPayG := &Account{
		ID: 402, Platform: PlatformKimi, Type: AccountTypeAPIKey, Status: StatusActive,
		Credentials: map[string]any{"api_key": "sk-x", "base_url": "https://api.kimi.com/coding/v1", "account_mode": AccountModePayG},
	}
	require.Equal(t, AccountModePayG, explicitPayG.effectiveCNAccountMode())

	codingZhipu := &Account{
		ID: 403, Platform: PlatformZhipu, Type: AccountTypeAPIKey, Status: StatusActive,
		Credentials: map[string]any{"api_key": "sk-x", "base_url": "https://open.bigmodel.cn/api/coding/paas/v4"},
	}
	require.Equal(t, DefaultZhipuCodingBaseURL, codingZhipu.defaultCNProtocolBaseURL(APIProtocolChatCompletions))

	// 纯 API 建号可能只填了 api_base_urls.chat_completions。
	noLegacyBaseURL := &Account{
		ID: 404, Platform: PlatformKimi, Type: AccountTypeAPIKey, Status: StatusActive,
		Credentials: map[string]any{"api_key": "sk-x", "api_base_urls": map[string]any{
			APIProtocolChatCompletions: "https://api.kimi.com/coding/v1",
		}},
	}
	require.Equal(t, AccountModeCoding, noLegacyBaseURL.effectiveCNAccountMode())

	for _, bad := range []string{
		"", "not a url", "https://api.kimi.com/v1",
		// 第三方中转：/coding-proxy 与非官方域名下的 /coding 都不得推断成编程套餐，
		// 否则会连带改掉额度/余额监控口径。
		"https://relay.example/coding-proxy/v1", "https://relay.example/coding/v1",
	} {
		require.False(t, isCNCodingPlanBaseURL(bad), bad)
	}
	require.True(t, isCNCodingPlanBaseURL("https://api.kimi.com/coding"))
	require.True(t, isCNCodingPlanBaseURL("https://API.Kimi.com/Coding/v1/"))
	require.True(t, isCNCodingPlanBaseURL("https://open.bigmodel.cn/api/coding/paas/v4"))
}

// 自适应整轮探测里 chat_completions 这条通道的 401 同样不得停用账号：
// 它复用 testOpenAIChatCompletionsConnection，那里原本无条件 SetError。
func TestAccountTestService_AdaptiveChatChannel401DoesNotDisableAccount(t *testing.T) {
	account := adaptiveCNAccountTestAccount(317, PlatformKimi)
	svc, _ := adaptiveCNAccountTestService(
		account,
		newJSONResponse(http.StatusUnauthorized, `{"error":{"message":"Invalid Authentication"}}`),
	)
	c, recorder := newTestContext()

	err := svc.TestAccountConnection(c, account.ID, "kimi-k2.6", "", AccountTestModeDefault)

	require.Error(t, err)
	require.Contains(t, recorder.Body.String(), "401")
	repo := svc.accountRepo.(*openAIAccountTestRepo)
	require.Zero(t, repo.setErrorID, "chat 通道探测 401 也不得把账号整体置 error")
}

// 固定单协议账号（非 adaptive）的 chat 探测 401 仍然停用账号：本次改动只放开 adaptive。
func TestAccountTestService_FixedChatProtocol401StillDisablesAccount(t *testing.T) {
	account := adaptiveCNAccountTestAccount(318, PlatformKimi)
	account.Credentials["api_protocol"] = APIProtocolChatCompletions
	account.Credentials["base_url"] = "http://chat.example/v1"
	svc, _ := adaptiveCNAccountTestService(
		account,
		newJSONResponse(http.StatusUnauthorized, `{"error":{"message":"Invalid Authentication"}}`),
	)
	c, _ := newTestContext()

	err := svc.TestAccountConnection(c, account.ID, "kimi-k2.6", "", AccountTestModeDefault)

	require.Error(t, err)
	repo := svc.accountRepo.(*openAIAccountTestRepo)
	require.Equal(t, account.ID, repo.setErrorID)
}
