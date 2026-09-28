import type { GroupPlatform } from '@/types'

export const OPENAI_CC_SWITCH_CODEX_MODEL = 'gpt-5.5'
export const DEEPSEEK_CC_SWITCH_CODEX_MODEL = 'deepseek-flash'
export const KIMI_CC_SWITCH_CODEX_MODEL = 'kimi-for-coding'
// zhipu / minimax 的 Codex 目标型号与后端 cnDefaultMessagesDispatchModels
// （openai_messages_dispatch.go）、前端 groupsMessagesDispatch.ts 的兜底保持一致。
export const ZHIPU_CC_SWITCH_CODEX_MODEL = 'glm-4.6'
export const MINIMAX_CC_SWITCH_CODEX_MODEL = 'MiniMax-M2.7'
export const GROK_CC_SWITCH_MODEL = 'grok-4.5'

export type CcSwitchClientType = 'claude' | 'gemini' | 'codex'

export interface CcSwitchImportConfig {
  app: string
  endpoint: string
  model?: string
}

export interface CcSwitchImportDeeplinkInput {
  baseUrl: string
  platform?: GroupPlatform | null
  clientType: CcSwitchClientType
  providerName: string
  apiKey: string
  usageScript: string
}

/**
 * Balance query CC Switch runs against the imported provider. CC Switch fills
 * `{{baseUrl}}` with the provider's base URL as stored — Codex and Grok imports
 * carry a trailing `/v1` (see `withV1Endpoint`), Claude ones do not, and users
 * may edit it either way afterwards — then evaluates the script, so the URL
 * strips an existing `/v1` instead of blindly appending one (`/v1/v1/usage`
 * is a 404 and CC Switch shows "query failed").
 */
export const CC_SWITCH_USAGE_SCRIPT = `({
    request: {
      url: "{{baseUrl}}".replace(/\\/+$/, "").replace(/\\/v1$/, "") + "/v1/usage",
      method: "GET",
      headers: { "Authorization": "Bearer {{apiKey}}" }
    },
    extractor: function(response) {
      const remaining = response?.remaining ?? response?.quota?.remaining ?? response?.balance;
      const unit = response?.unit ?? response?.quota?.unit ?? "USD";
      return {
        isValid: response?.is_active ?? response?.isValid ?? true,
        remaining,
        unit
      };
    }
  })`

function withV1Endpoint(baseUrl: string): string {
  const normalizedBaseUrl = baseUrl.replace(/\/+$/, '')
  return normalizedBaseUrl.endsWith('/v1') ? normalizedBaseUrl : `${normalizedBaseUrl}/v1`
}

function withoutTrailingSlashes(baseUrl: string): string {
  return baseUrl.replace(/\/+$/, '')
}

export function resolveCcSwitchImportConfig(
  platform: GroupPlatform | undefined | null,
  clientType: CcSwitchClientType,
  baseUrl: string
): CcSwitchImportConfig {
  switch (platform || 'anthropic') {
    case 'antigravity':
      return {
        app: clientType === 'gemini' ? 'gemini' : 'claude',
        endpoint: `${baseUrl.replace(/\/+$/, '')}/antigravity`
      }
    case 'openai':
      if (clientType === 'codex') {
        // CC Switch's Codex provider appends the OpenAI-compatible path itself.
        // Passing /v1 here can make the client request /v1/v1/.... (upstream 14483c925)
        return { app: 'codex', endpoint: withoutTrailingSlashes(baseUrl), model: OPENAI_CC_SWITCH_CODEX_MODEL }
      }
      return { app: 'claude', endpoint: baseUrl }
    case 'deepseek':
      if (clientType === 'codex') {
        return { app: 'codex', endpoint: baseUrl, model: DEEPSEEK_CC_SWITCH_CODEX_MODEL }
      }
      return { app: 'claude', endpoint: baseUrl }
    case 'kimi':
      if (clientType === 'codex') {
        return { app: 'codex', endpoint: baseUrl, model: KIMI_CC_SWITCH_CODEX_MODEL }
      }
      return { app: 'claude', endpoint: baseUrl }
    case 'zhipu':
      if (clientType === 'codex') {
        return { app: 'codex', endpoint: baseUrl, model: ZHIPU_CC_SWITCH_CODEX_MODEL }
      }
      return { app: 'claude', endpoint: baseUrl }
    case 'minimax':
      if (clientType === 'codex') {
        return { app: 'codex', endpoint: baseUrl, model: MINIMAX_CC_SWITCH_CODEX_MODEL }
      }
      return { app: 'claude', endpoint: baseUrl }
    case 'gemini':
      return {
        app: 'gemini',
        endpoint: baseUrl
      }
    case 'grok':
      return {
        app: 'grokbuild',
        endpoint: withV1Endpoint(baseUrl),
        model: GROK_CC_SWITCH_MODEL
      }
    default:
      return {
        app: 'claude',
        endpoint: baseUrl
      }
  }
}

export function buildCcSwitchImportDeeplink(input: CcSwitchImportDeeplinkInput): string {
  const config = resolveCcSwitchImportConfig(input.platform, input.clientType, input.baseUrl)
  const entries: [string, string][] = [
    ['resource', 'provider'],
    ['app', config.app],
    ['name', input.providerName],
    ['homepage', input.baseUrl],
    ['endpoint', config.endpoint],
    ['apiKey', input.apiKey],
    ['configFormat', 'json'],
    ['usageEnabled', 'true'],
    ['usageScript', btoa(input.usageScript)],
    ['usageAutoInterval', '30']
  ]

  if (config.model) {
    entries.splice(2, 0, ['model', config.model])
  }

  return `ccswitch://v1/import?${new URLSearchParams(entries).toString()}`
}
