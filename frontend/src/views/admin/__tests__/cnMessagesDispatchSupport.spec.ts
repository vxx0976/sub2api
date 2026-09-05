import { describe, it, expect } from 'vitest'
import {
  supportsMessagesDispatchPlatform,
  createDefaultMessagesDispatchFormState,
  messagesDispatchConfigToFormState,
} from '../groupsMessagesDispatch'

// 国产分组必须保留 /v1/messages 派发配置的编辑入口。
//
// 上游这里只有 openai + composite，前提是国产账号配 api_protocol=anthropic 原生直通、
// 模型名交给账号级 model_mapping。本站不满足该前提：生产 Deepseek 分组近 30 天约 90%
// 请求用 claude-* 模型名，靠分组级映射翻成 deepseek-v4-pro / deepseek-v4-flash；
// 账号级 model_mapping 是恒等白名单，接不住。后端（sanitizeGroupMessagesDispatchFields
// 对 CN 保留配置、ResolveMessagesDispatchModel 对 CN 读它）也是这么实现的，
// 前端若跟随上游收窄，就变成「后端在用、管理员看不到也改不了」，新建 CN 分组还必然空配置。
describe('CN 分组的 messages dispatch 支持', () => {
  it('白名单含国产三家与 composite', () => {
    for (const p of ['openai', 'composite', 'deepseek', 'kimi', 'zhipu']) {
      expect(supportsMessagesDispatchPlatform(p), p).toBe(true)
    }
    for (const p of ['anthropic', 'gemini', 'grok', 'antigravity', '', null, undefined]) {
      expect(supportsMessagesDispatchPlatform(p as never), String(p)).toBe(false)
    }
  })

  it('默认映射按平台给，绝不把 gpt-5.x 塞给国产分组', () => {
    expect(createDefaultMessagesDispatchFormState('deepseek')).toMatchObject({
      opus_mapped_model: 'deepseek-v4-pro',
      sonnet_mapped_model: 'deepseek-v4-pro',
      haiku_mapped_model: 'deepseek-v4-flash',
    })
    expect(createDefaultMessagesDispatchFormState('kimi')).toMatchObject({
      opus_mapped_model: 'kimi-k2.6',
      sonnet_mapped_model: 'kimi-k2.6',
      haiku_mapped_model: 'kimi-k2.6',
    })
    expect(createDefaultMessagesDispatchFormState('zhipu')).toMatchObject({
      opus_mapped_model: 'glm-4.6',
      sonnet_mapped_model: 'glm-4.6',
      haiku_mapped_model: 'glm-4.5-air',
    })
    // OpenAI 默认走 GPT-5.6 家族
    expect(createDefaultMessagesDispatchFormState('openai')).toMatchObject({
      opus_mapped_model: 'gpt-6-astra',
      sonnet_mapped_model: 'gpt-5.6-sol',
      haiku_mapped_model: 'gpt-5.6-terra',
    })
    // 国产平台的默认值里不得出现任何 gpt-*
    for (const p of ['deepseek', 'kimi', 'zhipu'] as const) {
      const s = createDefaultMessagesDispatchFormState(p)
      for (const v of [s.opus_mapped_model, s.sonnet_mapped_model, s.haiku_mapped_model]) {
        expect(v.startsWith('gpt-'), `${p} 默认值不应是 OpenAI 型号: ${v}`).toBe(false)
      }
    }
  })

  it('回显已保存配置时保留库里的值，缺项才按平台补默认', () => {
    const state = messagesDispatchConfigToFormState(
      { opus_mapped_model: 'deepseek-v4-pro', sonnet_mapped_model: '', haiku_mapped_model: '' },
      'deepseek',
    )
    expect(state.opus_mapped_model).toBe('deepseek-v4-pro')
    expect(state.sonnet_mapped_model).toBe('deepseek-v4-pro')
    expect(state.haiku_mapped_model).toBe('deepseek-v4-flash')
  })
})
