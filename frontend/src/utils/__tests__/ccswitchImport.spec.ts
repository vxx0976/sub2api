import { describe, expect, it } from 'vitest'
import {
  CC_SWITCH_USAGE_SCRIPT,
  DEEPSEEK_CC_SWITCH_CODEX_MODEL,
  GROK_CC_SWITCH_MODEL,
  KIMI_CC_SWITCH_CODEX_MODEL,
  MINIMAX_CC_SWITCH_CODEX_MODEL,
  OPENAI_CC_SWITCH_CODEX_MODEL,
  ZHIPU_CC_SWITCH_CODEX_MODEL,
  buildCcSwitchImportDeeplink
} from '@/utils/ccswitchImport'
import type { GroupPlatform } from '@/types'

function paramsFromDeeplink(deeplink: string): URLSearchParams {
  const query = deeplink.split('?')[1] || ''
  return new URLSearchParams(query)
}

describe('ccswitchImport utils', () => {
  it('defaults OpenAI CC Switch imports to the current Codex model', () => {
    expect(OPENAI_CC_SWITCH_CODEX_MODEL).toBe('gpt-5.5')
  })

  it('defaults Grok Build imports to the current Grok model', () => {
    expect(GROK_CC_SWITCH_MODEL).toBe('grok-4.5')
  })

  const baseInput = {
    baseUrl: 'https://api.example.com',
    providerName: 'Sub2API',
    apiKey: 'sk-test',
    usageScript: 'return true'
  }

  it.each([
    ['https://api.example.com', 'https://api.example.com'],
    ['https://api.example.com/', 'https://api.example.com'],
    ['https://api.example.com/v1', 'https://api.example.com/v1'],
    ['https://api.example.com/v1/', 'https://api.example.com/v1']
  ])('keeps Codex imports on the configured endpoint for base URL %s', (baseUrl, endpoint) => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        baseUrl,
        platform: 'openai',
        // fork(dev 8d53be42): openai 按 clientType 拆分协议，codex 客户端才导出 Codex 供应商；
        // 端点跟随上游 14483c925 保持根地址（CC Switch 的 Codex 供应商自己补路径，不再补 /v1）
        clientType: 'codex'
      })
    )

    expect(params.get('resource')).toBe('provider')
    expect(params.get('app')).toBe('codex')
    expect(params.get('endpoint')).toBe(endpoint)
    expect(params.get('model')).toBe(OPENAI_CC_SWITCH_CODEX_MODEL)
    expect(atob(params.get('usageScript') || '')).toBe(baseInput.usageScript)
  })

  it('exports OpenAI groups as a Claude provider when the client is Claude Code', () => {
    // dev 8d53be42：OpenAI 分组也可经 /v1/messages 派发给 Claude Code，按原样 baseUrl 导出、不带 Codex 模型
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({ ...baseInput, platform: 'openai', clientType: 'claude' })
    )

    expect(params.get('app')).toBe('claude')
    expect(params.get('endpoint')).toBe(baseInput.baseUrl)
    expect(params.has('model')).toBe(false)
  })

  it.each([
    'https://api.example.com',
    'https://api.example.com/',
    'https://api.example.com/v1',
    'https://api.example.com/v1/'
  ])('imports Grok Build with one /v1 suffix for base URL %s', (baseUrl) => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        baseUrl,
        platform: 'grok',
        clientType: 'claude'
      })
    )

    expect(params.get('app')).toBe('grokbuild')
    expect(params.get('endpoint')).toBe('https://api.example.com/v1')
    expect(params.get('model')).toBe(GROK_CC_SWITCH_MODEL)
  })

  // 支持 /v1/messages 派发的国产平台都要能导出 Codex 配置；漏一个的表现是
  // 该平台的 Key 点「导入 ccswitch」时静默按 Anthropic 协议导出，Codex 侧拿不到配置。
  it.each([
    { platform: 'deepseek' as GroupPlatform, model: DEEPSEEK_CC_SWITCH_CODEX_MODEL },
    { platform: 'kimi' as GroupPlatform, model: KIMI_CC_SWITCH_CODEX_MODEL },
    { platform: 'zhipu' as GroupPlatform, model: ZHIPU_CC_SWITCH_CODEX_MODEL },
    { platform: 'minimax' as GroupPlatform, model: MINIMAX_CC_SWITCH_CODEX_MODEL }
  ])('exports $platform as a Codex provider with its own model', ({ platform, model }) => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({ ...baseInput, platform, clientType: 'codex' })
    )

    expect(params.get('app')).toBe('codex')
    expect(params.get('endpoint')).toBe(baseInput.baseUrl)
    expect(params.get('model')).toBe(model)
  })

  it.each([
    { platform: 'zhipu' as GroupPlatform },
    { platform: 'minimax' as GroupPlatform }
  ])('exports $platform as a Claude provider when the client is Claude Code', ({ platform }) => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({ ...baseInput, platform, clientType: 'claude' })
    )

    expect(params.get('app')).toBe('claude')
    expect(params.get('model')).toBeNull()
  })

  it.each([
    { platform: 'anthropic' as GroupPlatform, clientType: 'claude' as const, app: 'claude' },
    { platform: 'gemini' as GroupPlatform, clientType: 'gemini' as const, app: 'gemini' }
  ])('does not add a model parameter for $platform imports', ({ platform, clientType, app }) => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        platform,
        clientType
      })
    )

    expect(params.get('app')).toBe(app)
    expect(params.get('endpoint')).toBe(baseInput.baseUrl)
    expect(params.has('model')).toBe(false)
  })

  it('keeps Antigravity imports on the selected client endpoint without a model parameter', () => {
    const params = paramsFromDeeplink(
      buildCcSwitchImportDeeplink({
        ...baseInput,
        platform: 'antigravity',
        clientType: 'gemini'
      })
    )

    expect(params.get('app')).toBe('gemini')
    expect(params.get('endpoint')).toBe(`${baseInput.baseUrl}/antigravity`)
    expect(params.has('model')).toBe(false)
  })
})

describe('CC Switch usage script', () => {
  // Mirrors CC Switch: substitute the template vars as text, evaluate, read request.url.
  function usageUrlFor(baseUrl: string): string {
    const script = CC_SWITCH_USAGE_SCRIPT.split('{{baseUrl}}').join(baseUrl).split('{{apiKey}}').join('sk-test')
    // eslint-disable-next-line no-new-func
    const config = new Function(`return ${script}`)() as { request: { url: string } }
    return config.request.url
  }

  it.each([
    'https://api.example.com',
    'https://api.example.com/',
    'https://api.example.com/v1',
    'https://api.example.com/v1/'
  ])('queries exactly one /v1/usage for base URL %s', (baseUrl) => {
    expect(usageUrlFor(baseUrl)).toBe('https://api.example.com/v1/usage')
  })

  it('works against the endpoint every platform import stores', () => {
    for (const platform of ['anthropic', 'openai', 'grok', 'gemini'] as GroupPlatform[]) {
      const endpoint = paramsFromDeeplink(
        buildCcSwitchImportDeeplink({
          baseUrl: 'https://api.example.com',
          platform,
          clientType: platform === 'gemini' ? 'gemini' : 'claude',
          providerName: 'Sub2API',
          apiKey: 'sk-test',
          usageScript: CC_SWITCH_USAGE_SCRIPT
        })
      ).get('endpoint') as string
      expect(usageUrlFor(endpoint)).toBe('https://api.example.com/v1/usage')
    }
  })

  // fork: 国产平台 + OpenAI 的 Codex 导出也走同一个余额脚本，同样只能查到一个 /v1/usage
  it.each(['openai', 'deepseek', 'kimi', 'zhipu', 'minimax'] as GroupPlatform[])(
    'works against the endpoint a %s Codex import stores',
    (platform) => {
      for (const baseUrl of ['https://api.example.com', 'https://api.example.com/']) {
        const endpoint = paramsFromDeeplink(
          buildCcSwitchImportDeeplink({
            baseUrl,
            platform,
            clientType: 'codex',
            providerName: 'Sub2API',
            apiKey: 'sk-test',
            usageScript: CC_SWITCH_USAGE_SCRIPT
          })
        ).get('endpoint') as string
        expect(usageUrlFor(endpoint)).toBe('https://api.example.com/v1/usage')
      }
    }
  )
})
