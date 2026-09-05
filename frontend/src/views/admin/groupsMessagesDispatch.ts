import type { GroupPlatform, OpenAIMessagesDispatchModelConfig } from "@/types";

export interface MessagesDispatchMappingRow {
  claude_model: string;
  target_model: string;
}

export interface MessagesDispatchFormState {
  allow_messages_dispatch: boolean;
  opus_mapped_model: string;
  sonnet_mapped_model: string;
  haiku_mapped_model: string;
  exact_model_mappings: MessagesDispatchMappingRow[];
}

// 按平台的默认映射目标。gpt-5.x 是 OpenAI 专属型号，发给国产上游必然 400，
// 所以国产平台必须各有各的默认值。
function getDefaultModelsForPlatform(platform?: GroupPlatform | null): {
  opus: string;
  sonnet: string;
  haiku: string;
} {
  switch (platform) {
    case "deepseek":
      return { opus: "deepseek-v4-pro", sonnet: "deepseek-v4-pro", haiku: "deepseek-v4-flash" };
    case "kimi":
      return { opus: "kimi-k2.6", sonnet: "kimi-k2.6", haiku: "kimi-k2.6" };
    case "zhipu":
      return { opus: "glm-4.6", sonnet: "glm-4.6", haiku: "glm-4.5-air" };
    default:
      // Haiku 用 gpt-5.6-sol：gpt-5.4-mini 官方公告下架、gpt-5.6-terra 实测会卡
      // 30-45 秒无输出，详见后端 defaultOpenAIMessagesDispatchHaikuMappedModel 的说明。
      return { opus: "gpt-6-astra", sonnet: "gpt-5.6-sol", haiku: "gpt-5.6-sol" };
  }
}

// 支持 /v1/messages 派发配置的平台白名单。
//
// ⚠️ 上游这里只有 openai + composite，因为上游假定国产账号配 api_protocol=anthropic
// 原生直通、模型名交给账号级 model_mapping。本站不满足该前提：生产 Deepseek 分组近
// 30 天约 90% 请求用 claude-* 模型名，全靠这份**分组级**映射翻成 deepseek-v4-pro /
// deepseek-v4-flash（账号级 model_mapping 是恒等白名单，接不住）。
// 后端也是这么实现的——sanitizeGroupMessagesDispatchFields 对 CN 分组保留
// MessagesDispatchModelConfig，ResolveMessagesDispatchModel 对 CN 分组读它。
// 这里若跟随上游收窄，就会变成「后端在用、管理员却看不到也改不了」，
// 而且新建 CN 分组必然拿到空配置 → claude-* 原样透传上游 400。
const MESSAGES_DISPATCH_PLATFORMS: GroupPlatform[] = [
  "openai",
  "composite",
  "deepseek",
  "kimi",
  "zhipu",
];

export function supportsMessagesDispatchPlatform(
  platform?: GroupPlatform | string | null,
): boolean {
  return !!platform && MESSAGES_DISPATCH_PLATFORMS.includes(platform as GroupPlatform);
}

export function createDefaultMessagesDispatchFormState(
  platform?: GroupPlatform | null,
): MessagesDispatchFormState {
  const models = getDefaultModelsForPlatform(platform);
  return {
    allow_messages_dispatch: false,
    opus_mapped_model: models.opus,
    sonnet_mapped_model: models.sonnet,
    haiku_mapped_model: models.haiku,
    exact_model_mappings: [],
  };
}

export function messagesDispatchConfigToFormState(
  config?: OpenAIMessagesDispatchModelConfig | null,
  platform?: GroupPlatform | null,
): MessagesDispatchFormState {
  const defaults = createDefaultMessagesDispatchFormState(platform);
  const exactMappings = Object.entries(config?.exact_model_mappings || {})
    .sort(([left], [right]) => left.localeCompare(right))
    .map(([claude_model, target_model]) => ({ claude_model, target_model }));

  return {
    allow_messages_dispatch: false,
    opus_mapped_model:
      config?.opus_mapped_model?.trim() || defaults.opus_mapped_model,
    sonnet_mapped_model:
      config?.sonnet_mapped_model?.trim() || defaults.sonnet_mapped_model,
    haiku_mapped_model:
      config?.haiku_mapped_model?.trim() || defaults.haiku_mapped_model,
    exact_model_mappings: exactMappings,
  };
}

export function messagesDispatchFormStateToConfig(
  state: MessagesDispatchFormState,
): OpenAIMessagesDispatchModelConfig {
  const exactModelMappings = Object.fromEntries(
    state.exact_model_mappings
      .map((row) => [row.claude_model.trim(), row.target_model.trim()] as const)
      .filter(([claudeModel, targetModel]) => claudeModel && targetModel),
  );

  return {
    opus_mapped_model: state.opus_mapped_model.trim(),
    sonnet_mapped_model: state.sonnet_mapped_model.trim(),
    haiku_mapped_model: state.haiku_mapped_model.trim(),
    exact_model_mappings: exactModelMappings,
  };
}

export function resetMessagesDispatchFormState(
  target: MessagesDispatchFormState,
  platform?: GroupPlatform | null,
): void {
  const defaults = createDefaultMessagesDispatchFormState(platform);
  target.allow_messages_dispatch = defaults.allow_messages_dispatch;
  target.opus_mapped_model = defaults.opus_mapped_model;
  target.sonnet_mapped_model = defaults.sonnet_mapped_model;
  target.haiku_mapped_model = defaults.haiku_mapped_model;
  target.exact_model_mappings = [];
}
