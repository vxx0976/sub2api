package service

import (
	"context"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
)

// WithResolvedTargetPlatform stores the concrete provider chosen for a request
// made through a composite group.
func WithResolvedTargetPlatform(ctx context.Context, platform string) context.Context {
	platform = strings.TrimSpace(platform)
	if ctx == nil || platform == "" {
		return ctx
	}
	return context.WithValue(ctx, ctxkey.ResolvedTargetPlatform, platform)
}

// ResolvedTargetPlatformFromContext returns the concrete provider chosen for
// the current request, if one was resolved.
func ResolvedTargetPlatformFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	platform, ok := ctx.Value(ctxkey.ResolvedTargetPlatform).(string)
	platform = strings.TrimSpace(platform)
	if !ok || platform == "" {
		return "", false
	}
	return platform, true
}

func WithCompositeRouteDecision(ctx context.Context, decision CompositeRouteDecision) context.Context {
	if ctx == nil || !decision.Matched {
		return ctx
	}
	ctx = WithResolvedTargetPlatform(ctx, decision.TargetPlatform)
	if model := strings.TrimSpace(decision.UpstreamModel); model != "" {
		ctx = context.WithValue(ctx, ctxkey.ResolvedUpstreamModel, model)
	}
	if model := strings.TrimSpace(decision.PublicModel); model != "" {
		ctx = context.WithValue(ctx, ctxkey.RequestedPublicModel, model)
	}
	if source := strings.TrimSpace(decision.Source); source != "" {
		ctx = context.WithValue(ctx, ctxkey.CompositeRouteSource, source)
	}
	return ctx
}

func ResolvedUpstreamModelFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	model, ok := ctx.Value(ctxkey.ResolvedUpstreamModel).(string)
	model = strings.TrimSpace(model)
	if !ok || model == "" {
		return "", false
	}
	return model, true
}

func RequestedPublicModelFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	model, ok := ctx.Value(ctxkey.RequestedPublicModel).(string)
	model = strings.TrimSpace(model)
	if !ok || model == "" {
		return "", false
	}
	return model, true
}

func CompositeRouteSourceFromContext(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}
	source, ok := ctx.Value(ctxkey.CompositeRouteSource).(string)
	source = strings.TrimSpace(source)
	if !ok || source == "" {
		return "", false
	}
	return source, true
}

// DetectModelPlatform maps common public model IDs to the concrete provider
// platform used by sub2api. It intentionally returns false for ambiguous model
// names so composite groups fail closed instead of guessing.
func DetectModelPlatform(model string) (string, bool) {
	normalized := strings.ToLower(strings.TrimSpace(model))
	if normalized == "" {
		return "", false
	}

	normalized = strings.TrimPrefix(normalized, "models/")
	if slash := strings.IndexByte(normalized, '/'); slash > 0 {
		provider := strings.TrimSpace(normalized[:slash])
		rest := strings.TrimSpace(normalized[slash+1:])
		switch provider {
		case "anthropic", "claude":
			return PlatformAnthropic, true
		case "openai", "chatgpt":
			return PlatformOpenAI, true
		case "google", "google-ai-studio", "gemini":
			return PlatformGemini, true
		case "xai", "x-ai", "grok":
			return PlatformGrok, true
		}
		if rest != "" {
			normalized = strings.TrimPrefix(rest, "models/")
		}
	}

	switch {
	case strings.HasPrefix(normalized, "anthropic.claude-"),
		strings.HasPrefix(normalized, "claude-"):
		return PlatformAnthropic, true
	case strings.HasPrefix(normalized, "gpt-"),
		strings.HasPrefix(normalized, "chatgpt-"),
		strings.HasPrefix(normalized, "codex-"),
		strings.HasPrefix(normalized, "text-embedding-"),
		strings.HasPrefix(normalized, "text-moderation-"),
		strings.HasPrefix(normalized, "omni-moderation-"),
		strings.HasPrefix(normalized, "dall-e-"),
		strings.HasPrefix(normalized, "gpt-image-"),
		strings.HasPrefix(normalized, "tts-"),
		strings.HasPrefix(normalized, "whisper-"),
		hasOpenAISeriesPrefix(normalized):
		return PlatformOpenAI, true
	case strings.HasPrefix(normalized, "gemini-"),
		strings.HasPrefix(normalized, "learnlm-"):
		return PlatformGemini, true
	case normalized == "grok" || strings.HasPrefix(normalized, "grok-"):
		return PlatformGrok, true
	default:
		return "", false
	}
}

func hasOpenAISeriesPrefix(model string) bool {
	for _, prefix := range []string{"o1", "o3", "o4", "o5"} {
		if model == prefix || strings.HasPrefix(model, prefix+"-") {
			return true
		}
	}
	return false
}

func (s *GatewayService) resolveCompositeRouteDecision(ctx context.Context, group *Group, requestedModel, endpoint string) (CompositeRouteDecision, bool, error) {
	if group == nil || group.Platform != PlatformComposite {
		return CompositeRouteDecision{}, false, nil
	}
	if platform, ok := ResolvedTargetPlatformFromContext(ctx); ok {
		upstreamModel := requestedModel
		if resolvedModel, modelOK := ResolvedUpstreamModelFromContext(ctx); modelOK {
			upstreamModel = resolvedModel
		}
		source := CompositeRouteSourceDetector
		if resolvedSource, sourceOK := CompositeRouteSourceFromContext(ctx); sourceOK {
			source = resolvedSource
		}
		return CompositeRouteDecision{
			Matched:        true,
			Source:         source,
			GroupID:        group.ID,
			PublicModel:    requestedModel,
			TargetPlatform: platform,
			UpstreamModel:  upstreamModel,
			Endpoint:       normalizeCompositeRouteEndpoint(endpoint),
		}, true, nil
	}
	decision, err := s.compositeResolver.Resolve(ctx, group.ID, requestedModel, endpoint)
	if err != nil {
		return decision, false, err
	}
	return decision, decision.Matched, nil
}

// compositeRequestPlatforms 是复合分组能够真正承载的具体平台集合。
//
// ⚠️ 这里**刻意**只有 5 个，不是 AllowedQuotaPlatforms 的 10 个。dev fork 多出的
// deepseek / moonshot / glm / qwen / seedance 只能作为「单平台分组」使用，复合分组
// 对它们没有实现端到端支持。放宽这个集合不会让它们跑起来，只会让管理员配出静默
// 误路由的复合路由（比现在干净地拒绝更糟）。三处运行时缺口，缺一不可：
//
//  1. DetectModelPlatform 只认 claude- / gpt- / gemini- / grok- 这几族模型名，
//     对 deepseek-* / kimi-* / glm-* / qwen-* / seedance-* 一律返回 ok=false，
//     自动探测这条路永远解析不到这些平台。
//  2. openAICompatibleRequestPlatform（handler 侧）把「非 Grok 的已解析平台」一律
//     压成 PlatformOpenAI。即便管理员显式配了 target_platform=deepseek 的路由，
//     请求也会被当成 openai 走进 OpenAI 账号池——错号池、错计费平台。
//  3. 各端点的 compositeTargetPlatformAllowed(...) 白名单只写了 OpenAI（Responses
//     另加 Grok），国产平台到不了转发层。
//
// 也就是说：要支持它们得同时补齐上面三处 + matchingPlatforms，不是改这一个谓词。
// 注意 PlatformAntigravity 在集合内但 DetectModelPlatform 不会产出它——它只经
// /antigravity 路由的 ForcePlatform 进来，属于正常情况。
//
// 另注：schedulerCanonicalBuckets 覆盖全部 10 个平台与此**不矛盾**。那是所有分组
// 通用的调度快照桶（schedulerBucketsForGroup 对任意 groupID 都调用），10 个平台各自
// 的单平台分组都需要桶，与复合分组是否支持它们无关。
// ⚠️ 次序有行为意义，勿随手调整：lookupPricingAcrossPlatforms /
// lookupMappingAcrossPlatforms 按此顺序取**首个命中**，复合分组下同名模型在多个平台
// 都配了定价/映射时，由这个顺序决定用哪一份。这里沿用 matchingPlatforms 的历史顺序。
func compositeRequestPlatforms() []string {
	return []string{
		PlatformAnthropic,
		PlatformGemini,
		PlatformOpenAI,
		PlatformAntigravity,
		PlatformGrok,
	}
}

// isConcreteRequestPlatform 报告 platform 是否为复合分组可承载的具体平台。
// 集合与取舍理由见 compositeRequestPlatforms。
func isConcreteRequestPlatform(platform string) bool {
	for _, p := range compositeRequestPlatforms() {
		if p == platform {
			return true
		}
	}
	return false
}
