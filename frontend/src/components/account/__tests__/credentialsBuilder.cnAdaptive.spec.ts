import { describe, expect, it } from 'vitest'

import {
  cnSupportsNativeResponses,
  defaultCNAdaptiveBaseUrls,
  inferCNAccountModeFromBaseUrl
} from '../credentialsBuilder'

describe('cnSupportsNativeResponses', () => {
  it('is true for DeepSeek, Kimi, and MiniMax', () => {
    expect(cnSupportsNativeResponses('deepseek')).toBe(true)
    expect(cnSupportsNativeResponses('kimi')).toBe(true)
    expect(cnSupportsNativeResponses('minimax')).toBe(true)
    expect(cnSupportsNativeResponses('zhipu')).toBe(false)
    expect(cnSupportsNativeResponses('openai')).toBe(false)
  })
})

describe('defaultCNAdaptiveBaseUrls', () => {
  it('resolves Kimi endpoints by account mode', () => {
    expect(defaultCNAdaptiveBaseUrls('kimi', 'payg')).toEqual({
      chat_completions: 'https://api.moonshot.cn/v1',
      anthropic: 'https://api.moonshot.cn/anthropic',
      responses: 'https://api.moonshot.cn/v1'
    })
    expect(defaultCNAdaptiveBaseUrls('kimi', 'coding')).toEqual({
      chat_completions: 'https://api.kimi.com/coding/v1',
      anthropic: 'https://api.kimi.com/coding',
      responses: 'https://api.kimi.com/coding/v1'
    })
  })

  it('resolves GLM endpoints by account mode', () => {
    expect(defaultCNAdaptiveBaseUrls('zhipu', 'payg')).toEqual({
      chat_completions: 'https://open.bigmodel.cn/api/paas/v4',
      anthropic: 'https://open.bigmodel.cn/api/anthropic',
      responses: ''
    })
    expect(defaultCNAdaptiveBaseUrls('zhipu', 'coding')).toEqual({
      chat_completions: 'https://open.bigmodel.cn/api/coding/paas/v4',
      anthropic: 'https://open.bigmodel.cn/api/anthropic',
      responses: ''
    })
  })

  it('includes all three native DeepSeek endpoints', () => {
    expect(defaultCNAdaptiveBaseUrls('deepseek', 'payg')).toEqual({
      chat_completions: 'https://api.deepseek.com',
      anthropic: 'https://api.deepseek.com/anthropic',
      responses: 'https://api.deepseek.com'
    })
  })

  it('uses the same MiniMax CN endpoints for payg and coding', () => {
    const expected = {
      chat_completions: 'https://api.minimaxi.com/v1',
      anthropic: 'https://api.minimaxi.com/anthropic',
      responses: 'https://api.minimaxi.com/v1'
    }
    expect(defaultCNAdaptiveBaseUrls('minimax', 'payg')).toEqual(expected)
    expect(defaultCNAdaptiveBaseUrls('minimax', 'coding')).toEqual(expected)
  })
})

describe('inferCNAccountModeFromBaseUrl', () => {
  it('官方域名 + 独立 /coding 路径段才算编程套餐', () => {
    expect(inferCNAccountModeFromBaseUrl('kimi', 'https://api.kimi.com/coding/v1')).toBe('coding')
    expect(inferCNAccountModeFromBaseUrl('kimi', ' https://API.Kimi.com/Coding ')).toBe('coding')
    expect(inferCNAccountModeFromBaseUrl('zhipu', 'https://open.bigmodel.cn/api/coding/paas/v4')).toBe('coding')
  })

  it('按量付费端点、第三方中转、无编程套餐的平台一律返回空串（调用方回落 payg）', () => {
    expect(inferCNAccountModeFromBaseUrl('kimi', 'https://api.moonshot.cn/v1')).toBe('')
    expect(inferCNAccountModeFromBaseUrl('zhipu', 'https://open.bigmodel.cn/api/paas/v4')).toBe('')
    // 第三方中转的 /coding-proxy 与套了 coding 路径的非官方域名都不算
    expect(inferCNAccountModeFromBaseUrl('kimi', 'https://relay.example/coding-proxy/v1')).toBe('')
    expect(inferCNAccountModeFromBaseUrl('kimi', 'https://relay.example/coding/v1')).toBe('')
    // DeepSeek / MiniMax 的表单没有 coding 选项，不参与推断
    expect(inferCNAccountModeFromBaseUrl('deepseek', 'https://api.kimi.com/coding/v1')).toBe('')
    expect(inferCNAccountModeFromBaseUrl('kimi', '')).toBe('')
    expect(inferCNAccountModeFromBaseUrl('kimi', 'not a url')).toBe('')
  })
})
