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
		case "kimi", "moonshot":
			return PlatformKimi, true
		case "zhipu", "glm", "bigmodel":
			return PlatformZhipu, true
		case "deepseek":
			return PlatformDeepseek, true
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
	case normalized == "k3",
		normalized == "k3-256k",
		strings.HasPrefix(normalized, "kimi-"),
		strings.HasPrefix(normalized, "moonshot-"):
		return PlatformKimi, true
	case strings.HasPrefix(normalized, "glm-"):
		return PlatformZhipu, true
	case strings.HasPrefix(normalized, "deepseek-"):
		return PlatformDeepseek, true
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
// 本集合与 AllowedQuotaPlatforms 现已同为 8 个，但仍是两条独立不变量：调度桶
// （schedulerCanonicalBuckets / schedulerBucketsForGroup）对任意分组通用，本集合
// 只描述「复合分组能承载什么」。别把两者当同一件事，也别用其中一个去证明另一个。
//
// 国产三家（kimi / zhipu / deepseek）曾因三处运行时缺口被刻意排除在外，现已全部补齐：
//
//  1. DetectModelPlatform 认 kimi/moonshot、zhipu/glm/bigmodel、deepseek 三组
//     provider 前缀，以及 kimi- / moonshot- / glm- / deepseek- 四组模型名前缀。
//  2. handler 侧 openAICompatibleRequestPlatform 改走
//     service.NormalizeOpenAICompatiblePlatform：grok/kimi/zhipu/deepseek 原样保留，
//     不再被压成 PlatformOpenAI（压平会进错号池、错计费平台）。
//  3. 文本类端点白名单 openAICompatibleTextTargetAllowed 含 Kimi/Zhipu/Deepseek，
//     覆盖 /v1/chat/completions、/v1/responses、/v1/messages 及两个 count_tokens 端点。
//
// ⚠️ 仍**刻意**不放开的窄口（这是当前设计，不是待办事项，解冲突/重构时别顺手补齐）：
//   - /v1/embeddings、/v1/images/*、/v1/alpha/search、/v1/realtime 仍只允许
//     PlatformOpenAI（各自硬写死）。
//   - Responses WebSocket 走 isResponsesWebSocketCompositePlatform，只有
//     openai + grok：CN 账号过不了 WSv2 ingress 的 transport 过滤，且 WS HTTP 桥
//     没有面向 CN 的 Responses 转换，放行只会把明确的策略拒绝变成误导性的
//     "no available account"。
//
// 注意 PlatformAntigravity 在集合内但 DetectModelPlatform 不会产出它——它只经
// /antigravity 路由的 ForcePlatform 进来，属于正常情况。
//
// ⚠️ 次序是**行为契约**，勿随手调整：matchingPlatforms(PlatformComposite) 直接返回
// 本函数的结果，lookupPricingAcrossPlatforms / lookupMappingAcrossPlatforms 先整轮
// 精确匹配、再整轮通配匹配，两轮都按此顺序取**首个命中**。复合分组下同名模型在多个
// 平台都配了定价/映射时，由这个顺序决定用哪一份。
// 国产三家一律**追加在队尾**，理由是：前 5 个平台之间的既有命中结果因此一字不变，
// 从 5 扩到 8 对存量复合分组零回归。把它们插进前 5 个中间会改变既有命中。
func compositeRequestPlatforms() []string {
	return []string{
		PlatformAnthropic,
		PlatformGemini,
		PlatformOpenAI,
		PlatformAntigravity,
		PlatformGrok,
		PlatformKimi,
		PlatformZhipu,
		PlatformDeepseek,
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
