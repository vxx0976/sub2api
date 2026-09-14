package service

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// normalizeAnthropicDirectInputUsage 的口径契约：
//   - DeepSeek / Kimi（语义已实测：input_tokens 只报缓存未命中数）无条件加回缓存桶；
//   - 其他平台（Zhipu 等，语义未实测）仅在 input_tokens < cache_read+cache_creation（明显为
//     未命中口径）时加回，总量口径上游不得双重计费。
func TestNormalizeAnthropicDirectInputUsage(t *testing.T) {
	t.Run("DeepSeek 未命中口径无条件加回", func(t *testing.T) {
		// 实测样例：3015 token prompt → input=71, cache_read=2944
		u := OpenAIUsage{InputTokens: 71, CacheReadInputTokens: 2944}
		normalizeAnthropicDirectInputUsage(PlatformDeepseek, &u)
		require.Equal(t, 3015, u.InputTokens)
	})

	t.Run("DeepSeek 新增内容超过缓存前缀也加回", func(t *testing.T) {
		u := OpenAIUsage{InputTokens: 5000, CacheReadInputTokens: 2944}
		normalizeAnthropicDirectInputUsage(PlatformDeepseek, &u)
		require.Equal(t, 7944, u.InputTokens, "条件判断 (input < cache_read) 会在此漏计")
	})

	t.Run("Kimi 新增内容超过缓存前缀也加回", func(t *testing.T) {
		// 2026-09-14 实测 api.kimi.com/coding/v1/messages：2482 token 前缀已缓存，
		// 追加约 6250 token 新内容 → input=6430, cache_read=2304（两者之和才是完整 prompt）。
		// 旧的条件加回 (6430 < 2304 为假) 会漏计整段缓存前缀。
		u := OpenAIUsage{InputTokens: 6430, CacheReadInputTokens: 2304}
		normalizeAnthropicDirectInputUsage(PlatformKimi, &u)
		require.Equal(t, 8734, u.InputTokens)
	})

	t.Run("Kimi 明显未命中口径加回", func(t *testing.T) {
		u := OpenAIUsage{InputTokens: 71, CacheReadInputTokens: 2944}
		normalizeAnthropicDirectInputUsage(PlatformKimi, &u)
		require.Equal(t, 3015, u.InputTokens)
	})

	t.Run("Zhipu 总量口径不得双重计费", func(t *testing.T) {
		// 语义未实测的平台若按总量口径上报（input 已含全部输入），不应再加回
		u := OpenAIUsage{InputTokens: 3015, CacheReadInputTokens: 2944}
		normalizeAnthropicDirectInputUsage(PlatformZhipu, &u)
		require.Equal(t, 3015, u.InputTokens)
	})

	t.Run("Zhipu 明显未命中口径仍加回", func(t *testing.T) {
		u := OpenAIUsage{InputTokens: 71, CacheReadInputTokens: 2944}
		normalizeAnthropicDirectInputUsage(PlatformZhipu, &u)
		require.Equal(t, 3015, u.InputTokens)
	})

	t.Run("无缓存命中为 no-op", func(t *testing.T) {
		u := OpenAIUsage{InputTokens: 100}
		normalizeAnthropicDirectInputUsage(PlatformDeepseek, &u)
		require.Equal(t, 100, u.InputTokens)
	})

	t.Run("DeepSeek 一并加回 cache_creation", func(t *testing.T) {
		u := OpenAIUsage{InputTokens: 71, CacheReadInputTokens: 2944, CacheCreationInputTokens: 500}
		normalizeAnthropicDirectInputUsage(PlatformDeepseek, &u)
		require.Equal(t, 71+2944+500, u.InputTokens)
	})

	t.Run("首笔缓存写入 creation>input 也加回(修双减)", func(t *testing.T) {
		// 首次写缓存：cache_read=0、cache_creation 很大、input 很小。旧逻辑
		// 条件 (input < cache_read) = (50 < 0) 为假 → 不加回 → 下游多减一次
		// cache_creation → 新输入被夹成 0。现按 input < read+creation 加回。
		u := OpenAIUsage{InputTokens: 50, CacheReadInputTokens: 0, CacheCreationInputTokens: 3000}
		normalizeAnthropicDirectInputUsage(PlatformZhipu, &u)
		require.Equal(t, 3050, u.InputTokens)
	})

	t.Run("Zhipu 总量口径含 creation 不双计", func(t *testing.T) {
		// 总量口径下 input 恒 >= read+creation，不应再加回。
		u := OpenAIUsage{InputTokens: 3515, CacheReadInputTokens: 2944, CacheCreationInputTokens: 500}
		normalizeAnthropicDirectInputUsage(PlatformZhipu, &u)
		require.Equal(t, 3515, u.InputTokens)
	})
}

// resolveAnthropicDirectInputUsage：带 prompt_tokens 的上游按显式总量拆分，不按平台加回双计。
func TestResolveAnthropicDirectInputUsage(t *testing.T) {
	t.Run("Kimi 实测未命中口径（无 prompt_tokens）无条件加回", func(t *testing.T) {
		got := resolveAnthropicDirectInputUsage(PlatformKimi, gjson.Parse(`{"input_tokens":6430,"cache_creation_input_tokens":0,"cache_read_input_tokens":2304,"output_tokens":16}`))
		require.Equal(t, anthropicDirectInputUsage{total: 8734, cacheRead: 2304}, got)
	})

	t.Run("Kimi 总量口径带 prompt_tokens 不双计", func(t *testing.T) {
		// 历史 Kimi 样例（kimi_anthropic_usage_test.go）：input=prompt=173306、cache_read=173056。
		got := resolveAnthropicDirectInputUsage(PlatformKimi, gjson.Parse(`{"input_tokens":173306,"output_tokens":166,"cache_read_input_tokens":173056,"prompt_tokens":173306,"cached_tokens":173056}`))
		require.Equal(t, anthropicDirectInputUsage{total: 173306, cacheRead: 173056}, got)
	})

	t.Run("DeepSeek prompt_cache_hit/miss 显式拆分", func(t *testing.T) {
		got := resolveAnthropicDirectInputUsage(PlatformDeepseek, gjson.Parse(`{"input_tokens":400,"output_tokens":300,"prompt_cache_hit_tokens":800,"prompt_cache_miss_tokens":400}`))
		require.Equal(t, anthropicDirectInputUsage{total: 1200, cacheRead: 800}, got)
	})

	t.Run("只带 prompt_cache_hit_tokens 不走拆分（防未命中被算成 0）", func(t *testing.T) {
		got := resolveAnthropicDirectInputUsage(PlatformDeepseek, gjson.Parse(`{"input_tokens":400,"prompt_cache_hit_tokens":800,"cache_read_input_tokens":800}`))
		require.Equal(t, anthropicDirectInputUsage{total: 1200, cacheRead: 800}, got)
	})

	t.Run("DeepSeek 无 prompt 字段沿用无条件加回", func(t *testing.T) {
		got := resolveAnthropicDirectInputUsage(PlatformDeepseek, gjson.Parse(`{"input_tokens":71,"cache_read_input_tokens":2944}`))
		require.Equal(t, anthropicDirectInputUsage{total: 3015, cacheRead: 2944}, got)
	})
}

// message_delta 最终输入拆分覆盖 message_start 的契约。
func TestMergeAnthropicDirectDeltaInputUsage(t *testing.T) {
	startOf := func(platform, raw string) (anthropicDirectInputUsage, OpenAIUsage) {
		start := resolveAnthropicDirectInputUsage(platform, gjson.Parse(raw))
		var u OpenAIUsage
		start.apply(&u)
		return start, u
	}

	t.Run("Kimi start 只报总量、delta 给真实拆分 → 以 delta 为准", func(t *testing.T) {
		// 实测：message_start 8736/0 → message_delta input=32, cache_read=8704
		start, u := startOf(PlatformKimi, `{"input_tokens":8736,"cache_read_input_tokens":0}`)
		delta := gjson.Parse(`{"input_tokens":32,"cache_creation_input_tokens":0,"cache_read_input_tokens":8704,"output_tokens":16}`)
		require.True(t, mergeAnthropicDirectDeltaInputUsage(PlatformKimi, delta, start, &u))
		require.Equal(t, 8736, u.InputTokens, "归一后总量不变")
		require.Equal(t, 8704, u.CacheReadInputTokens)
		require.Equal(t, 32, u.InputTokens-u.CacheReadInputTokens-u.CacheCreationInputTokens, "新输入桶只剩未命中部分")
	})

	t.Run("Kimi 带 prompt_tokens 的 delta 也以 delta 为准", func(t *testing.T) {
		start, u := startOf(PlatformKimi, `{"input_tokens":173306,"cache_read_input_tokens":0,"prompt_tokens":173306,"cached_tokens":0}`)
		delta := gjson.Parse(`{"input_tokens":250,"cache_read_input_tokens":173056,"output_tokens":166,"prompt_tokens":173306,"cached_tokens":173056}`)
		require.True(t, mergeAnthropicDirectDeltaInputUsage(PlatformKimi, delta, start, &u))
		require.Equal(t, 173306, u.InputTokens)
		require.Equal(t, 173056, u.CacheReadInputTokens)
	})

	t.Run("Zhipu 条件口径下 delta 归一后总量变小 → 保持 start（防少计）", func(t *testing.T) {
		// start 8736/0；delta 6430/2306 在条件加回口径下归一为 6430 < 8736，不得覆盖。
		start, u := startOf(PlatformZhipu, `{"input_tokens":8736,"cache_read_input_tokens":0}`)
		before := u
		delta := gjson.Parse(`{"input_tokens":6430,"cache_read_input_tokens":2306,"output_tokens":16}`)
		require.False(t, mergeAnthropicDirectDeltaInputUsage(PlatformZhipu, delta, start, &u))
		require.Equal(t, before, u)
	})

	t.Run("delta 只报 output_tokens（DeepSeek 形态）保持 start 口径", func(t *testing.T) {
		start, u := startOf(PlatformDeepseek, `{"input_tokens":71,"cache_read_input_tokens":2944}`)
		before := u
		require.False(t, mergeAnthropicDirectDeltaInputUsage(PlatformDeepseek, gjson.Parse(`{"output_tokens":120}`), start, &u))
		require.Equal(t, before, u)
	})

	t.Run("残缺 delta 只报 input_tokens=0 不得清零", func(t *testing.T) {
		start, u := startOf(PlatformKimi, `{"input_tokens":8736}`)
		before := u
		require.False(t, mergeAnthropicDirectDeltaInputUsage(PlatformKimi, gjson.Parse(`{"input_tokens":0,"output_tokens":16}`), start, &u))
		require.Equal(t, before, u)
	})

	t.Run("delta 总量小于 start 不覆盖（防少计）", func(t *testing.T) {
		start, u := startOf(PlatformDeepseek, `{"input_tokens":71,"cache_read_input_tokens":2944}`)
		before := u
		require.False(t, mergeAnthropicDirectDeltaInputUsage(PlatformDeepseek, gjson.Parse(`{"input_tokens":71,"output_tokens":16}`), start, &u))
		require.Equal(t, before, u)
	})

	t.Run("delta 与 start 相同为 no-op", func(t *testing.T) {
		start, u := startOf(PlatformDeepseek, `{"input_tokens":71,"cache_read_input_tokens":2944}`)
		require.False(t, mergeAnthropicDirectDeltaInputUsage(PlatformDeepseek, gjson.Parse(`{"input_tokens":71,"cache_read_input_tokens":2944,"output_tokens":16}`), start, &u))
		require.Equal(t, 3015, u.InputTokens)
	})
}

// 端到端：两条流式处理函数都必须采用 message_delta 的最终拆分。
func TestAnthropicDirectStreamHandlers_UseDeltaInputSplit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	sse := strings.Join([]string{
		`event: message_start`,
		`data: {"type":"message_start","message":{"id":"msg_1","type":"message","role":"assistant","content":[],"model":"kimi-k3","usage":{"input_tokens":8736,"cache_creation_input_tokens":0,"cache_read_input_tokens":0,"output_tokens":0}}}`,
		``,
		`event: content_block_start`,
		`data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`,
		``,
		`event: content_block_delta`,
		`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"hi"}}`,
		``,
		`event: message_delta`,
		`data: {"type":"message_delta","delta":{"stop_reason":"end_turn","stop_sequence":null},"usage":{"input_tokens":32,"cache_creation_input_tokens":0,"cache_read_input_tokens":8704,"output_tokens":16}}`,
		``,
		`event: message_stop`,
		`data: {"type":"message_stop"}`,
		``,
	}, "\n")
	newResp := func() *http.Response {
		return &http.Response{StatusCode: http.StatusOK, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(sse))}
	}
	check := func(t *testing.T, result *OpenAIForwardResult, err error) {
		t.Helper()
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, 8736, result.Usage.InputTokens)
		require.Equal(t, 8704, result.Usage.CacheReadInputTokens)
		require.Equal(t, 16, result.Usage.OutputTokens)
	}
	svc := &OpenAIGatewayService{}

	t.Run("streaming", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
		result, err := svc.handleAnthropicDirectStreamingResponse(newResp(), c, PlatformKimi, "kimi-k3", "kimi-k3", "kimi-k3", time.Now())
		check(t, result, err)
	})

	t.Run("buffered", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
		account := &Account{Platform: PlatformKimi, Type: AccountTypeAPIKey}
		result, err := svc.handleAnthropicDirectBufferedSSE(newResp(), c, account, "kimi-k3", "kimi-k3", "kimi-k3", time.Now())
		check(t, result, err)
		// 回给客户端的是 Anthropic 口径：input_tokens 只含未命中部分，不能把缓存再算一遍。
		require.Equal(t, int64(32), gjson.Get(rec.Body.String(), "usage.input_tokens").Int())
		require.Equal(t, int64(8704), gjson.Get(rec.Body.String(), "usage.cache_read_input_tokens").Int())
	})

	t.Run("json body with prompt_tokens (Kimi total semantics)", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)
		c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
		account := &Account{Platform: PlatformKimi, Type: AccountTypeAPIKey}
		body := `{"id":"msg_2","type":"message","content":[],"usage":{"input_tokens":173306,"output_tokens":166,"cache_read_input_tokens":173056,"prompt_tokens":173306,"cached_tokens":173056}}`
		resp := &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(body))}
		result, err := svc.handleAnthropicDirectBufferedResponse(resp, c, account, "kimi-k3", "kimi-k3", "kimi-k3", time.Now())
		require.NoError(t, err)
		require.Equal(t, 173306, result.Usage.InputTokens, "显式 prompt_tokens 总量不得被再加一遍缓存")
		require.Equal(t, 173056, result.Usage.CacheReadInputTokens)
		require.Equal(t, 166, result.Usage.OutputTokens)
	})
}

// TestNormalizeAnthropicDirectInputUsage_BucketMath 验证归一后 InputTokens 经下游
// 三桶互斥拆分（actualInput = InputTokens - cache_read - cache_creation，见
// openai_gateway_service.go RecordUsage）后，三桶最终量各自等于真实量。
func TestNormalizeAnthropicDirectInputUsage_BucketMath(t *testing.T) {
	// 下游拆分口径（与 RecordUsage line ~6916 一致）。
	actualInput := func(u OpenAIUsage) int {
		v := u.InputTokens - u.CacheReadInputTokens - u.CacheCreationInputTokens
		if v < 0 {
			v = 0
		}
		return v
	}

	cases := []struct {
		name       string
		platform   string
		raw        OpenAIUsage // 上游原始上报（Anthropic 未命中口径）
		wantInput  int         // 期望真实新输入
		wantCreate int
		wantRead   int
	}{
		{
			name:       "DeepSeek 读+写缓存",
			platform:   PlatformDeepseek,
			raw:        OpenAIUsage{InputTokens: 71, CacheReadInputTokens: 2944, CacheCreationInputTokens: 500},
			wantInput:  71,
			wantCreate: 500,
			wantRead:   2944,
		},
		{
			name:       "首笔缓存写入 creation>input 不再夹成 0",
			platform:   PlatformKimi,
			raw:        OpenAIUsage{InputTokens: 50, CacheReadInputTokens: 0, CacheCreationInputTokens: 3000},
			wantInput:  50,
			wantCreate: 3000,
			wantRead:   0,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			u := tc.raw
			normalizeAnthropicDirectInputUsage(tc.platform, &u)
			require.Equal(t, tc.wantInput, actualInput(u), "新输入桶")
			require.Equal(t, tc.wantCreate, u.CacheCreationInputTokens, "cache_creation 桶")
			require.Equal(t, tc.wantRead, u.CacheReadInputTokens, "cache_read 桶")
		})
	}
}

// buildAnthropicDirectMessagesURL 的逐平台 URL 约定契约。
// GLM 端点根因上游而异：智谱官方/z.ai 在 /api/anthropic 下，NewAPI 中转在根。
func TestBuildAnthropicDirectMessagesURL(t *testing.T) {
	apikey := func(platform, baseURL string) *Account {
		return &Account{Platform: platform, Type: AccountTypeAPIKey, Credentials: map[string]any{"base_url": baseURL}}
	}
	cases := []struct {
		name    string
		account *Account
		want    string
	}{
		{"DeepSeek 默认", &Account{Platform: PlatformDeepseek, Type: AccountTypeAPIKey}, "https://api.deepseek.com/anthropic/v1/messages"},
		{"Kimi 默认", &Account{Platform: PlatformKimi, Type: AccountTypeAPIKey}, "https://api.kimi.com/coding/v1/messages"},
		{"Zhipu 官方默认 base 补 /api/anthropic", &Account{Platform: PlatformZhipu, Type: AccountTypeAPIKey}, "https://open.bigmodel.cn/api/anthropic/v1/messages"},
		{"Zhipu 官方 paas 根也归一到 /api/anthropic", apikey(PlatformZhipu, "https://open.bigmodel.cn/api/paas/v4"), "https://open.bigmodel.cn/api/anthropic/v1/messages"},
		{"Zhipu 官方已含 /api/anthropic 不重复", apikey(PlatformZhipu, "https://open.bigmodel.cn/api/anthropic"), "https://open.bigmodel.cn/api/anthropic/v1/messages"},
		{"Zhipu z.ai 补 /api/anthropic", apikey(PlatformZhipu, "https://api.z.ai"), "https://api.z.ai/api/anthropic/v1/messages"},
		{"Zhipu NewAPI 中转根直挂 /v1/messages", apikey(PlatformZhipu, "https://relay.orbitai.cc"), "https://relay.orbitai.cc/v1/messages"},
		{"Zhipu 中转 base 带 /v1 归一", apikey(PlatformZhipu, "https://relay.orbitai.cc/v1"), "https://relay.orbitai.cc/v1/messages"},
		{"Zhipu 中转 base 带末尾斜杠", apikey(PlatformZhipu, "https://relay.orbitai.cc/"), "https://relay.orbitai.cc/v1/messages"},
		{"未支持平台返回空", &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, buildAnthropicDirectMessagesURL(tc.account))
		})
	}
}
