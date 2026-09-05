package service

import (
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/xai"
)

const (
	defaultOpenAIMessagesDispatchOpusMappedModel   = "gpt-6-astra"
	defaultOpenAIMessagesDispatchSonnetMappedModel = "gpt-5.6-sol"
	// Haiku 落在 gpt-5.6-sol 而不是更便宜的同代型号，是两次实测排除法的结果：
	// gpt-5.4-mini 官方已公告下架（gpt-5.4 早已对 ChatGPT 账号返回 400）；
	// gpt-5.6-terra 在生产号池里按账号分化严重（2026-09-05 实测三个账号
	// 10/10、6/10、5/10），失败形态是卡 30-45 秒无输出；剩下能稳定秒回的
	// 便宜档只有 gpt-5.6-luna，但站长明确不用 luna。Haiku 承接 Claude Code
	// 的后台高频调用，宁可贵也不能卡。terra 稳定或另有便宜型号后再调。
	defaultOpenAIMessagesDispatchHaikuMappedModel = "gpt-5.6-sol"
)

func normalizeOpenAIMessagesDispatchMappedModel(model string) string {
	model = NormalizeOpenAICompatRequestedModel(strings.TrimSpace(model))
	return strings.TrimSpace(model)
}

func normalizeOpenAIMessagesDispatchModelConfig(cfg OpenAIMessagesDispatchModelConfig) OpenAIMessagesDispatchModelConfig {
	out := OpenAIMessagesDispatchModelConfig{
		OpusMappedModel:   normalizeOpenAIMessagesDispatchMappedModel(cfg.OpusMappedModel),
		SonnetMappedModel: normalizeOpenAIMessagesDispatchMappedModel(cfg.SonnetMappedModel),
		HaikuMappedModel:  normalizeOpenAIMessagesDispatchMappedModel(cfg.HaikuMappedModel),
	}

	if len(cfg.ExactModelMappings) > 0 {
		out.ExactModelMappings = make(map[string]string, len(cfg.ExactModelMappings))
		for requestedModel, mappedModel := range cfg.ExactModelMappings {
			requestedModel = strings.TrimSpace(requestedModel)
			mappedModel = normalizeOpenAIMessagesDispatchMappedModel(mappedModel)
			if requestedModel == "" || mappedModel == "" {
				continue
			}
			out.ExactModelMappings[requestedModel] = mappedModel
		}
		if len(out.ExactModelMappings) == 0 {
			out.ExactModelMappings = nil
		}
	}

	return out
}

func claudeMessagesDispatchFamily(model string) string {
	normalized := strings.ToLower(strings.TrimSpace(model))
	if !strings.HasPrefix(normalized, "claude") {
		return ""
	}
	switch {
	case strings.Contains(normalized, "opus"):
		return "opus"
	case strings.Contains(normalized, "sonnet"):
		return "sonnet"
	case strings.Contains(normalized, "haiku"):
		return "haiku"
	default:
		return ""
	}
}

// cnDefaultMessagesDispatchModels 是国产平台在未配置分组级映射时的兜底目标。
// 与前端 groupsMessagesDispatch.ts 的 getDefaultModelsForPlatform 保持一致。
func cnDefaultMessagesDispatchModels(platform string) (opus, sonnet, haiku string) {
	switch platform {
	case PlatformDeepseek:
		return "deepseek-v4-pro", "deepseek-v4-pro", "deepseek-v4-flash"
	case PlatformKimi:
		return "kimi-k2.6", "kimi-k2.6", "kimi-k2.6"
	case PlatformZhipu:
		return "glm-4.6", "glm-4.6", "glm-4.5-air"
	default:
		return "", "", ""
	}
}

func (g *Group) ResolveMessagesDispatchModel(requestedModel string) string {
	if g == nil {
		return ""
	}
	requestedModel = strings.TrimSpace(requestedModel)
	if requestedModel == "" {
		return ""
	}

	if g.Platform == PlatformGrok {
		if claudeMessagesDispatchFamily(requestedModel) == "" {
			return ""
		}
		opts := xai.RuntimeModelMappingOptions()
		if !opts.EnableCrossClientMap {
			return ""
		}
		return xai.ModelMappingWithOptions(opts)["claude-*"]
	}

	cfg := normalizeOpenAIMessagesDispatchModelConfig(g.MessagesDispatchModelConfig)

	// 国产供应商分组:用管理员配置的映射，但**绝不回落**下方的 gpt-5.x 默认值
	// （那是 OpenAI 专属型号，发给 kimi / zhipu / deepseek 上游必然 400）。
	//
	// ⚠️ 上游在此处是无条件 `if IsCNProvider(g.Platform) { return "" }`，理由是
	// CN 账号配 api_protocol=anthropic 时原生直通、模型名交给账号级 model_mapping。
	// 那个前提在本站不成立，照抄会直接打掉线上主力流量：
	//   - 生产 Deepseek 分组近 30 天 8300 次请求里约 7500 次用的是 claude-* 模型名
	//     （claude-opus-5 5487 / claude-opus-4-6 1179 / claude-sonnet-5 490 …），
	//     全靠这里的分组级映射翻成 deepseek-v4-pro / deepseek-v4-flash；
	//   - 两个 CN 账号都没有 api_protocol（默认 chat_completions）、base_url 指向
	//     api.deepseek.com / api.kimi.com，不是原生 Anthropic 端点；
	//   - 账号级 model_mapping 是恒等白名单（deepseek-v4-pro→deepseek-v4-pro），
	//     没有任何 claude-* 条目，接不住。
	// 所以这里保留映射能力，只把「没配置时的兜底」改成不改写（返回空 = 原样透传），
	// 而不是套用 OpenAI 的 gpt-5.x 默认值。
	if IsCNProvider(g.Platform) {
		if mapped := strings.TrimSpace(cfg.ExactModelMappings[requestedModel]); mapped != "" {
			return mapped
		}
		// 未配置时按平台兜底，而不是回落下方的 gpt-5.x（OpenAI 专属型号，发给国产上游必 400）。
		// 兜底值与前端 groupsMessagesDispatch.ts 的 getDefaultModelsForPlatform 一致。
		// 没有兜底会让新建的 CN 分组把 claude-opus-5 原样发给上游：轻则选号失败，
		// 重则上游接受了但 filterCNProviderBillingModelCandidates 滤空候选 → 零成本落账。
		opus, sonnet, haiku := cnDefaultMessagesDispatchModels(g.Platform)
		switch claudeMessagesDispatchFamily(requestedModel) {
		case "opus":
			if m := strings.TrimSpace(cfg.OpusMappedModel); m != "" {
				return m
			}
			return opus
		case "sonnet":
			if m := strings.TrimSpace(cfg.SonnetMappedModel); m != "" {
				return m
			}
			return sonnet
		case "haiku":
			if m := strings.TrimSpace(cfg.HaikuMappedModel); m != "" {
				return m
			}
			return haiku
		default:
			return ""
		}
	}

	if mappedModel := strings.TrimSpace(cfg.ExactModelMappings[requestedModel]); mappedModel != "" {
		return mappedModel
	}

	switch claudeMessagesDispatchFamily(requestedModel) {
	case "opus":
		if mappedModel := strings.TrimSpace(cfg.OpusMappedModel); mappedModel != "" {
			return mappedModel
		}
		return defaultOpenAIMessagesDispatchOpusMappedModel
	case "sonnet":
		if mappedModel := strings.TrimSpace(cfg.SonnetMappedModel); mappedModel != "" {
			return mappedModel
		}
		return defaultOpenAIMessagesDispatchSonnetMappedModel
	case "haiku":
		if mappedModel := strings.TrimSpace(cfg.HaikuMappedModel); mappedModel != "" {
			return mappedModel
		}
		return defaultOpenAIMessagesDispatchHaikuMappedModel
	default:
		return ""
	}
}

func sanitizeGroupMessagesDispatchFields(g *Group) {
	// OpenAI 分组的调度配置完整保留（含 AllowMessagesDispatch 与模型映射）。
	// Composite 分组只保留开关，映射交给复合路由解析出的目标平台。
	// CN 供应商分组保留**模型映射**：线上 Claude Code 用户就是靠它把 claude-*
	// 翻成 kimi-* / deepseek-*（理由见 ResolveMessagesDispatchModel 里的长注释）。
	// 其 AllowMessagesDispatch 仍按上游置 false —— handler 的
	// allowOpenAICompatibleMessagesDispatch 对 CN 分组直接豁免这个开关，
	// 留不留都不影响放行，跟随上游可减少下轮合并冲突。
	// 其余平台一律清空。
	if g == nil || g.Platform == PlatformOpenAI {
		return
	}
	if g.Platform != PlatformComposite {
		g.AllowMessagesDispatch = false
	}
	g.DefaultMappedModel = ""
	// CN 供应商只保留 MessagesDispatchModelConfig（线上 Claude Code 靠它把 claude-*
	// 翻成 kimi-* / deepseek-*）。AllowMessagesDispatch 与 DefaultMappedModel 仍按上游
	// 清空：前者 handler 已对 CN 豁免，后者生产 CN 分组本就未使用。
	if IsCNProvider(g.Platform) {
		return
	}
	g.MessagesDispatchModelConfig = OpenAIMessagesDispatchModelConfig{}
}
