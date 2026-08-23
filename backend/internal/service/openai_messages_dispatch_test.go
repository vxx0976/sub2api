package service

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/xai"
	"github.com/stretchr/testify/require"
)

func TestNormalizeOpenAIMessagesDispatchModelConfig(t *testing.T) {
	t.Parallel()

	cfg := normalizeOpenAIMessagesDispatchModelConfig(OpenAIMessagesDispatchModelConfig{
		OpusMappedModel:   " gpt-5.4-high ",
		SonnetMappedModel: "gpt-5.3-codex",
		HaikuMappedModel:  " gpt-5.4-mini-medium ",
		ExactModelMappings: map[string]string{
			" claude-sonnet-4-5-20250929 ": " gpt-5.2-high ",
			"":                             "gpt-5.4",
			"claude-opus-4-6":              " ",
		},
	})

	require.Equal(t, "gpt-5.4", cfg.OpusMappedModel)
	require.Equal(t, "gpt-5.3-codex", cfg.SonnetMappedModel)
	require.Equal(t, "gpt-5.4-mini", cfg.HaikuMappedModel)
	require.Equal(t, map[string]string{
		"claude-sonnet-4-5-20250929": "gpt-5.2",
	}, cfg.ExactModelMappings)
}

func TestGroupResolveMessagesDispatchModel_GrokRequiresCrossClientMapping(t *testing.T) {
	original := xai.RuntimeModelMappingOptions()
	t.Cleanup(func() { xai.SetRuntimeModelMappingOptions(original) })
	group := &Group{Platform: PlatformGrok}

	xai.SetRuntimeModelMappingOptions(xai.ModelMappingOptions{})
	require.Empty(t, group.ResolveMessagesDispatchModel("claude-sonnet-4-5"))

	xai.SetRuntimeModelMappingOptions(xai.ModelMappingOptions{
		DefaultText:          "grok-build-0.1",
		EnableCrossClientMap: true,
	})
	require.Equal(t, "grok-build-0.1", group.ResolveMessagesDispatchModel("claude-sonnet-4-5"))
	require.Equal(t, "grok-build-0.1", group.ResolveMessagesDispatchModel("claude-opus-4-6"))
	require.Equal(t, "grok-build-0.1", group.ResolveMessagesDispatchModel("claude-haiku-4-5"))
	require.Empty(t, group.ResolveMessagesDispatchModel("grok"))
	require.Empty(t, group.ResolveMessagesDispatchModel("gpt-5.3-codex"))
}

func TestSanitizeGroupMessagesDispatchFields_ClearsNonOpenAIPlatform(t *testing.T) {
	t.Parallel()

	group := &Group{
		Platform:              PlatformAnthropic,
		AllowMessagesDispatch: true,
		DefaultMappedModel:    "gpt-5.6-sol",
		MessagesDispatchModelConfig: OpenAIMessagesDispatchModelConfig{
			SonnetMappedModel: "gpt-5.3-codex",
			ExactModelMappings: map[string]string{
				"claude-fable-5": "gpt-5.6-sol",
			},
		},
	}

	sanitizeGroupMessagesDispatchFields(group)

	require.False(t, group.AllowMessagesDispatch)
	require.Empty(t, group.DefaultMappedModel)
	require.Equal(t, OpenAIMessagesDispatchModelConfig{}, group.MessagesDispatchModelConfig)
}

// OpenAI 分组必须整体保留调度配置。合并 origin/main 时这条早退曾被冲突块的
// 公共前缀吃掉（只剩 if g == nil），而当时的用例只覆盖 anthropic / composite，
// 测不出「每次保存都把 OpenAI 分组的映射清空」这个回归。
func TestSanitizeGroupMessagesDispatchFields_PreservesOpenAIConfig(t *testing.T) {
	t.Parallel()

	cfg := OpenAIMessagesDispatchModelConfig{
		SonnetMappedModel: "gpt-5.3-codex",
		ExactModelMappings: map[string]string{
			"claude-fable-5": "gpt-5.6-sol",
		},
	}
	group := &Group{
		Platform:                    PlatformOpenAI,
		AllowMessagesDispatch:       true,
		DefaultMappedModel:          "gpt-5.6-sol",
		MessagesDispatchModelConfig: cfg,
	}

	sanitizeGroupMessagesDispatchFields(group)

	require.True(t, group.AllowMessagesDispatch)
	require.Equal(t, "gpt-5.6-sol", group.DefaultMappedModel)
	require.Equal(t, cfg, group.MessagesDispatchModelConfig)
}

// CN 供应商分组的 /v1/messages 能力改由账号 api_protocol 表达，分组级开关一律清空；
// 同时 ResolveMessagesDispatchModel 对 CN 平台恒返回空（不得回落到 gpt-5.x 默认值）。
// 本用例由上游的 TestMessagesDispatchClearedForCNProviders 改写而来。
//
// 上游对 CN 分组是「开关 + 默认模型 + 模型映射」三样全清、且 ResolveMessagesDispatchModel
// 直接 return ""，前提是 CN 账号配 api_protocol=anthropic 原生直通、模型名交给账号级
// model_mapping。本站不满足该前提（生产 CN 分组约 90% 请求用 claude-* 模型名，靠分组级
// 映射翻成 kimi-* / deepseek-*；账号级 model_mapping 是恒等白名单），照抄会打掉主力流量。
//
// 本 fork 的口径：AllowMessagesDispatch 与 DefaultMappedModel 仍按上游清空
// （前者 handler 对 CN 已豁免，后者生产未使用），**只保留 MessagesDispatchModelConfig**；
// 同时保住上游真正想防的那件事——CN 分组绝不回落到 gpt-5.x 默认值。
// 逐条行为见 cn_messages_dispatch_mapping_test.go。
func TestMessagesDispatchSanitizeForCNProviders(t *testing.T) {
	t.Parallel()

	for _, platform := range []string{PlatformKimi, PlatformZhipu, PlatformDeepseek} {
		cfg := OpenAIMessagesDispatchModelConfig{SonnetMappedModel: "kimi-k2.6"}
		group := &Group{
			Platform:                    platform,
			AllowMessagesDispatch:       true,
			DefaultMappedModel:          "gpt-5.6-sol",
			MessagesDispatchModelConfig: cfg,
		}

		// 配了映射就必须命中（上游此处是 Empty）。
		require.Equal(t, "kimi-k2.6", group.ResolveMessagesDispatchModel("claude-sonnet-4-5"), platform)

		sanitizeGroupMessagesDispatchFields(group)
		require.False(t, group.AllowMessagesDispatch, platform)
		require.Empty(t, group.DefaultMappedModel, platform)
		require.Equal(t, cfg, group.MessagesDispatchModelConfig, platform,
			"CN 分组的模型映射不能被清掉——线上 Claude Code 流量靠它")
	}
}

// 上游那条约束的实质（CN 分组不得吃到 openai 专属的 gpt-5.x 默认值）必须继续成立。
func TestMessagesDispatchCNProvidersNeverInheritOpenAIDefaults(t *testing.T) {
	t.Parallel()

	// 本 fork 不是靠「返回空」达成的，而是按平台给各自兜底——返回空会让 claude-*
	// 原样透传国产上游（选号失败，或上游接受后被 filterCNProviderBillingModelCandidates
	// 滤空候选 → 零成本落账）。要守的实质是「不得吃到 gpt-5.x」。
	for _, platform := range []string{PlatformKimi, PlatformZhipu, PlatformDeepseek} {
		group := &Group{Platform: platform, AllowMessagesDispatch: true} // 未配任何映射
		for _, model := range []string{"claude-sonnet-4-5", "claude-opus-5"} {
			got := group.ResolveMessagesDispatchModel(model)
			require.NotEmpty(t, got, "%s/%s 应按平台兜底而不是原样透传", platform, model)
			require.NotContains(t, got, "gpt-", "%s/%s 不得回落 OpenAI 型号：%s", platform, model, got)
		}
	}
}

func TestSanitizeGroupMessagesDispatchFields_PreservesCompositeDispatchToggle(t *testing.T) {
	t.Parallel()

	group := &Group{
		Platform:              PlatformComposite,
		AllowMessagesDispatch: true,
		DefaultMappedModel:    "gpt-5.6-sol",
		MessagesDispatchModelConfig: OpenAIMessagesDispatchModelConfig{
			SonnetMappedModel: "gpt-5.3-codex",
			ExactModelMappings: map[string]string{
				"claude-fable-5": "gpt-5.6-sol",
			},
		},
	}

	sanitizeGroupMessagesDispatchFields(group)

	require.True(t, group.AllowMessagesDispatch)
	require.Empty(t, group.DefaultMappedModel)
	require.Equal(t, OpenAIMessagesDispatchModelConfig{}, group.MessagesDispatchModelConfig)
}
