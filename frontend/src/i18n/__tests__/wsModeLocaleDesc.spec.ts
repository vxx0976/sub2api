import { describe, expect, it } from 'vitest'

import en from '../locales/en'
import zh from '../locales/zh'
import ru from '../locales/ru'

describe('OpenAI WS mode locale descriptions', () => {
  it('documents the global v2 router requirement for account WS modes', () => {
    // 上游 0.2.5 把 wsModeDesc 里的 http_bridge 字样拆进了按模式分述的三条 hint，
    // 描述本身只保留开关条件，所以这里只断言 mode_router_v2_enabled=true。
    // fork 的 locale 是扁平三份，路径带 admin. 前缀，且 ru 也要一起钉。
    expect(zh.admin.accounts.openai.wsModeDesc).toContain('mode_router_v2_enabled=true')
    expect(en.admin.accounts.openai.wsModeDesc).toContain('mode_router_v2_enabled=true')
    expect(ru.admin.accounts.openai.wsModeDesc).toContain('mode_router_v2_enabled=true')
  })
})
