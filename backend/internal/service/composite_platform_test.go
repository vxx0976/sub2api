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
		{name: "kimi", model: "kimi-k2-thinking", platform: PlatformKimi, ok: true},
		{name: "moonshot prefix", model: "moonshot/moonshot-v1-32k", platform: PlatformKimi, ok: true},
		{name: "zhipu", model: "glm-5.2", platform: PlatformZhipu, ok: true},
		{name: "deepseek", model: "deepseek-v4-pro", platform: PlatformDeepseek, ok: true},
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
// 8 个平台各自的单平台分组都需要桶。⚠️ 它与 compositeRequestPlatforms 现在虽然
// 同为 8 个，但仍是两条**独立不变量**：调度桶对任意分组通用，与「复合分组能承载
// 什么」无关。别用其中一个去证明另一个——两者恰好相等是巧合，不是契约。
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

// 复合分组的平台集与 AllowedQuotaPlatforms 现已同为 8 个，但**次序**仍是行为契约。
// 这批用例钉住两件事：全列表的字面次序，以及国产三家只能追加在队尾。
func TestCompositeRequestPlatformsOrderIsPricingContract(t *testing.T) {
	// 用有序断言而非 ElementsMatch：次序决定复合分组下同名模型取哪个平台的定价/映射
	// （lookupPricingAcrossPlatforms / lookupMappingAcrossPlatforms 取首个命中），属行为契约。
	require.Equal(t,
		[]string{
			PlatformAnthropic, PlatformGemini, PlatformOpenAI, PlatformAntigravity, PlatformGrok,
			PlatformKimi, PlatformZhipu, PlatformDeepseek,
		},
		compositeRequestPlatforms(),
	)

	// 尾追加不变量：原有 5 个平台的相对次序必须一字不动。国产三家插到前 5 个中间会
	// 改变既有复合分组的定价/映射命中结果（把某个平台的精确/通配行抢到前面）。
	require.Equal(t,
		[]string{PlatformAnthropic, PlatformGemini, PlatformOpenAI, PlatformAntigravity, PlatformGrok},
		compositeRequestPlatforms()[:5],
		"国产三家只能排在队尾；插进前 5 个中间 = 改变存量复合分组的定价/映射命中")

	// 与配额平台全集互为同一集合（仅次序不同）：新增平台时两处必须同步，
	// 否则复合分组会出现「有配额、进不了复合路由」或反之的裂缝。
	require.ElementsMatch(t, AllowedQuotaPlatforms, compositeRequestPlatforms())
	for _, platform := range compositeRequestPlatforms() {
		require.Contains(t, AllowedQuotaPlatforms, platform)
	}
}

// DetectModelPlatform 的产出必须落在复合分组能承载的平台集内——它解析出来的平台
// 会被直接当作请求目标，一旦产出集合外的平台就是误路由。
func TestDetectModelPlatformNeverEscapesCompositePlatformSet(t *testing.T) {
	allowed := make(map[string]struct{}, len(compositeRequestPlatforms()))
	for _, p := range compositeRequestPlatforms() {
		allowed[p] = struct{}{}
	}
	// 断言具体平台而不只是「落在集合内」，以捕捉 kimi / zhipu 之类的串台。
	for model, want := range map[string]string{
		"claude-sonnet-4.5":       PlatformAnthropic,
		"anthropic/claude-opus":   PlatformAnthropic,
		"gpt-5.6":                 PlatformOpenAI,
		"o3-mini":                 PlatformOpenAI,
		"codex-mini":              PlatformOpenAI,
		"text-embedding-3-large":  PlatformOpenAI,
		"openai/gpt-4o":           PlatformOpenAI,
		"gemini-2.5-pro":          PlatformGemini,
		"models/gemini-2.0-flash": PlatformGemini,
		"grok-4":                  PlatformGrok,
		"xai/grok-3":              PlatformGrok,
		// 国产三家已被上游补齐端到端支持，必须探测得到（见 compositeRequestPlatforms 注释）。
		"kimi-k2.6":            PlatformKimi,
		"kimi-k2.7":            PlatformKimi,
		"moonshot-v1-8k":       PlatformKimi,
		"moonshotai/kimi-k2.6": PlatformKimi,
		"glm-4.6":              PlatformZhipu,
		"z-ai/glm-5.1":         PlatformZhipu,
		"deepseek-chat":        PlatformDeepseek,
		"deepseek-v4-pro":      PlatformDeepseek,
	} {
		platform, ok := DetectModelPlatform(model)
		require.True(t, ok, "%s 应能被识别", model)
		require.Equal(t, want, platform, "%s 解析出的平台不对", model)
		require.Contains(t, allowed, platform, "%s 解析出集合外的平台 %s", model, platform)
	}

	// qwen / seedance 平台已下线，llama 从未支持：必须探测不到，
	// 防止被误当成复合路由目标（fail-closed 优于猜）。
	for _, model := range []string{"qwen-max", "seedance-1.0", "llama-4-maverick"} {
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

func TestCompositeConcretePlatformsIncludeCNProviders(t *testing.T) {
	for _, platform := range []string{PlatformKimi, PlatformZhipu, PlatformDeepseek} {
		require.True(t, isConcreteRequestPlatform(platform))
		require.True(t, canCopyAccountsFromGroupPlatform(PlatformComposite, platform))
	}
}
