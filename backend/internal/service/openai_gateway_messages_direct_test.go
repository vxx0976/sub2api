package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// normalizeAnthropicDirectInputUsage 的口径契约：
//   - DeepSeek（语义已实测：input_tokens 只报缓存未命中数）无条件加回 cache_read；
//   - 其他平台（Kimi 等，语义未实测）仅在 input_tokens < cache_read（明显为未命中
//     口径）时加回，总量口径上游不得双重计费。
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

	t.Run("Moonshot 总量口径不得双重计费", func(t *testing.T) {
		// 若 Kimi 按 Anthropic 总量口径上报（input 已含全部输入），不应再加回
		u := OpenAIUsage{InputTokens: 3015, CacheReadInputTokens: 2944}
		normalizeAnthropicDirectInputUsage(PlatformKimi, &u)
		require.Equal(t, 3015, u.InputTokens)
	})

	t.Run("Moonshot 明显未命中口径仍加回", func(t *testing.T) {
		u := OpenAIUsage{InputTokens: 71, CacheReadInputTokens: 2944}
		normalizeAnthropicDirectInputUsage(PlatformKimi, &u)
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
		normalizeAnthropicDirectInputUsage(PlatformKimi, &u)
		require.Equal(t, 3050, u.InputTokens)
	})

	t.Run("Moonshot 总量口径含 creation 不双计", func(t *testing.T) {
		// 总量口径下 input 恒 >= read+creation，不应再加回。
		u := OpenAIUsage{InputTokens: 3515, CacheReadInputTokens: 2944, CacheCreationInputTokens: 500}
		normalizeAnthropicDirectInputUsage(PlatformKimi, &u)
		require.Equal(t, 3515, u.InputTokens)
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
