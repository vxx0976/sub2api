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
		normalizeAnthropicDirectInputUsage(PlatformDeepSeek, &u)
		require.Equal(t, 3015, u.InputTokens)
	})

	t.Run("DeepSeek 新增内容超过缓存前缀也加回", func(t *testing.T) {
		u := OpenAIUsage{InputTokens: 5000, CacheReadInputTokens: 2944}
		normalizeAnthropicDirectInputUsage(PlatformDeepSeek, &u)
		require.Equal(t, 7944, u.InputTokens, "条件判断 (input < cache_read) 会在此漏计")
	})

	t.Run("Moonshot 总量口径不得双重计费", func(t *testing.T) {
		// 若 Kimi 按 Anthropic 总量口径上报（input 已含全部输入），不应再加回
		u := OpenAIUsage{InputTokens: 3015, CacheReadInputTokens: 2944}
		normalizeAnthropicDirectInputUsage(PlatformMoonshot, &u)
		require.Equal(t, 3015, u.InputTokens)
	})

	t.Run("Moonshot 明显未命中口径仍加回", func(t *testing.T) {
		u := OpenAIUsage{InputTokens: 71, CacheReadInputTokens: 2944}
		normalizeAnthropicDirectInputUsage(PlatformMoonshot, &u)
		require.Equal(t, 3015, u.InputTokens)
	})

	t.Run("无缓存命中为 no-op", func(t *testing.T) {
		u := OpenAIUsage{InputTokens: 100}
		normalizeAnthropicDirectInputUsage(PlatformDeepSeek, &u)
		require.Equal(t, 100, u.InputTokens)
	})
}
