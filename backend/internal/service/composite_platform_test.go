package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/stretchr/testify/require"
)

func TestDetectModelPlatform(t *testing.T) {
	tests := []struct {
		name     string
		model    string
		platform string
		ok       bool
	}{
		{name: "claude", model: "claude-sonnet-4-5", platform: PlatformAnthropic, ok: true},
		{name: "anthropic prefix", model: "anthropic/claude-opus-4-5", platform: PlatformAnthropic, ok: true},
		{name: "gpt", model: "gpt-5.1", platform: PlatformOpenAI, ok: true},
		{name: "o series", model: "o3-mini", platform: PlatformOpenAI, ok: true},
		{name: "embedding", model: "text-embedding-3-large", platform: PlatformOpenAI, ok: true},
		{name: "gemini", model: "gemini-3-pro", platform: PlatformGemini, ok: true},
		{name: "gemini models prefix", model: "models/gemini-2.5-flash", platform: PlatformGemini, ok: true},
		{name: "learnlm", model: "learnlm-2.0-flash-experimental", platform: PlatformGemini, ok: true},
		{name: "grok", model: "grok-4", platform: PlatformGrok, ok: true},
		{name: "xai prefix", model: "xai/grok-4", platform: PlatformGrok, ok: true},
		{name: "unknown", model: "llama-4-maverick", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			platform, ok := DetectModelPlatform(tt.model)
			require.Equal(t, tt.ok, ok)
			require.Equal(t, tt.platform, platform)
		})
	}
}

func TestQuotaPlatformCompositeUsesResolvedOrForceOnly(t *testing.T) {
	apiKey := &APIKey{Group: &Group{Platform: PlatformComposite}}

	require.Equal(t, "", QuotaPlatform(context.Background(), apiKey))
	require.Equal(t, PlatformGemini, QuotaPlatform(WithResolvedTargetPlatform(context.Background(), PlatformGemini), apiKey))
	require.Equal(t, PlatformAntigravity, QuotaPlatform(context.WithValue(context.Background(), ctxkey.ForcePlatform, PlatformAntigravity), apiKey))

	ctx := WithResolvedTargetPlatform(context.Background(), PlatformAnthropic)
	ctx = context.WithValue(ctx, ctxkey.ForcePlatform, PlatformAntigravity)
	require.Equal(t, PlatformAntigravity, QuotaPlatform(ctx, apiKey))
}

// 调度快照桶必须覆盖 AllowedQuotaPlatforms 全部 8 个平台。
//
// 这是所有分组通用的桶集合（schedulerBucketsForGroup 对任意 groupID 都调用），
// 8 个平台各自的单平台分组都需要桶。⚠️ 它**不是**「复合分组支持 8 个平台」的
// 证据——复合分组能承载的平台集是 compositeRequestPlatforms（5 个），两者互不相干。
// 原用例名与注释把二者混为一谈，曾据此得出「isConcreteRequestPlatform 少列了国产
// 平台」的错误结论。
func TestSchedulerCanonicalBucketsCoverAllQuotaPlatforms(t *testing.T) {
	seen := make(map[string]struct{})
	for _, bucket := range schedulerCanonicalBuckets(99) {
		seen[bucket.Platform] = struct{}{}
	}
	platforms := make([]string, 0, len(seen))
	for platform := range seen {
		platforms = append(platforms, platform)
	}
	require.ElementsMatch(t, AllowedQuotaPlatforms, platforms)
}

// 复合分组的平台集刻意窄于 AllowedQuotaPlatforms，理由见 compositeRequestPlatforms。
// 这批用例的作用是：谁想放宽它，必须先把三处运行时缺口一起补上，否则这里会红。
func TestCompositeRequestPlatformsStaysNarrowerThanQuotaPlatforms(t *testing.T) {
	// 用有序断言而非 ElementsMatch：次序决定复合分组下同名模型取哪个平台的定价/映射
	// （lookupPricingAcrossPlatforms 取首个命中），属行为契约。
	require.Equal(t,
		[]string{PlatformAnthropic, PlatformGemini, PlatformOpenAI, PlatformAntigravity, PlatformGrok},
		compositeRequestPlatforms(),
	)

	// fork 多出的 3 个国产 openai-compat 平台只能做单平台分组，不能进复合分组。
	for _, platform := range []string{
		PlatformKimi, PlatformZhipu, PlatformDeepseek,
	} {
		require.False(t, isConcreteRequestPlatform(platform),
			"%s 复合路由未实现端到端支持（DetectModelPlatform 不认其模型名、"+
				"openAICompatibleRequestPlatform 会把它压成 openai、端点白名单也没有它）；"+
				"放宽这里只会配出静默误路由", platform)
	}

	// 集合必须是配额平台全集的真子集：新增平台默认不进复合分组，要进得显式补齐。
	for _, platform := range compositeRequestPlatforms() {
		require.Contains(t, AllowedQuotaPlatforms, platform)
	}
	require.Less(t, len(compositeRequestPlatforms()), len(AllowedQuotaPlatforms))
}

// DetectModelPlatform 的产出必须落在复合分组能承载的平台集内——它解析出来的平台
// 会被直接当作请求目标，一旦产出集合外的平台就是误路由。
func TestDetectModelPlatformNeverEscapesCompositePlatformSet(t *testing.T) {
	allowed := make(map[string]struct{}, len(compositeRequestPlatforms()))
	for _, p := range compositeRequestPlatforms() {
		allowed[p] = struct{}{}
	}
	for _, model := range []string{
		"claude-sonnet-4.5", "anthropic/claude-opus", "gpt-5.6", "o3-mini", "codex-mini",
		"text-embedding-3-large", "gemini-2.5-pro", "models/gemini-2.0-flash", "grok-4",
		"openai/gpt-4o", "xai/grok-3",
	} {
		platform, ok := DetectModelPlatform(model)
		require.True(t, ok, "%s 应能被识别", model)
		require.Contains(t, allowed, platform, "%s 解析出集合外的平台 %s", model, platform)
	}

	// 国产模型不参与自动探测：探测不到就走不进复合路由，与平台集口径一致。
	for _, model := range []string{
		"deepseek-chat", "deepseek-reasoner", "kimi-k2.6", "glm-4.6", "qwen-max", "seedance-1.0",
	} {
		_, ok := DetectModelPlatform(model)
		require.False(t, ok, "%s 不应被自动探测为复合路由目标", model)
	}
}

// matchingPlatforms 与 isConcreteRequestPlatform 曾各抄一份平台集，会各自漂移。
// 现在共用 compositeRequestPlatforms，这里钉住两者不再分家。
func TestMatchingPlatformsCompositeSharesSingleSource(t *testing.T) {
	require.Equal(t, compositeRequestPlatforms(), matchingPlatforms(PlatformComposite),
		"含次序：matchingPlatforms 的遍历顺序决定定价/映射的首个命中")
	for _, platform := range matchingPlatforms(PlatformComposite) {
		require.True(t, isConcreteRequestPlatform(platform))
	}
	require.Equal(t, []string{PlatformDeepseek}, matchingPlatforms(PlatformDeepseek),
		"具体平台分组只匹配自身")
}
