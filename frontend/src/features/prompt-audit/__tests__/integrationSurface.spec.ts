import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'
import en from '@/i18n/locales/en'
import ru from '@/i18n/locales/ru'
import zh from '@/i18n/locales/zh'

const here = dirname(fileURLToPath(import.meta.url))
const read = (path: string) => readFileSync(resolve(here, path), 'utf8')

describe('Prompt Audit integration surface', () => {
  it('registers an admin and risk-control guarded route', () => {
    const router = read('../../../router/index.ts')
    expect(router).toContain("path: '/admin/prompt-audit'")
    const route = router.slice(router.indexOf("path: '/admin/prompt-audit'"), router.indexOf("path: '/admin/usage'"))
    expect(route).toContain('requiresAuth: true')
    expect(route).toContain('requiresAdmin: true')
    expect(route).toContain('requiresRiskControl: true')
  })

  it('keeps the legacy content moderation route and adds both pages under an expand-only security group', () => {
    const sidebar = read('../../../components/layout/AppSidebar.vue')
    const group = sidebar.slice(sidebar.indexOf("path: '/admin/security-audit'"), sidebar.indexOf("path: '/admin/redeem'"))
    expect(group).toContain('expandOnly: true')
    expect(group).toContain("path: '/admin/risk-control'")
    expect(group).toContain("path: '/admin/prompt-audit'")
  })

  it('keeps Prompt Audit locale trees symmetric and all operational controls named', () => {
    expect(Object.keys(zh.admin.promptAudit)).toEqual(Object.keys(en.admin.promptAudit))
    expect(Object.keys(ru.admin.promptAudit).sort()).toEqual(Object.keys(zh.admin.promptAudit).sort())
    expect(zh.nav.securityAudit).toBeTruthy()
    expect(en.nav.securityAudit).toBeTruthy()
    const endpoint = read('../components/EndpointPool.vue')
    const events = read('../components/EventWorkspace.vue')
    expect(endpoint).toContain('aria-label')
    expect(events).toContain('aria-label')
    expect(events).toContain('overflow-x-auto')
    expect(events).toContain('sm:grid-cols-2')
  })

  // Only the active locale is loaded at runtime, so a key missing from ru
  // renders the literal key path to the user instead of falling back to en.
  it('names both latest-turn switches in every shipped locale and separates their scopes', () => {
    for (const bar of [zh.admin.promptAudit.saveBar, en.admin.promptAudit.saveBar, ru.admin.promptAudit.saveBar]) {
      for (const key of ['auditLatestTurnOnly', 'auditLatestTurnOnlyHint', 'blockingLatestTurnOnly', 'blockingLatestTurnOnlyHint'] as const) {
        expect(typeof (bar as Record<string, string>)[key]).toBe('string')
        expect((bar as Record<string, string>)[key].length).toBeGreaterThan(0)
      }
      expect((bar as Record<string, string>).auditLatestTurnOnly).not.toBe((bar as Record<string, string>).blockingLatestTurnOnly)
      expect((bar as Record<string, string>).auditLatestTurnOnlyHint).not.toBe((bar as Record<string, string>).blockingLatestTurnOnlyHint)
    }
    expect(zh.admin.promptAudit.saveBar.auditLatestTurnOnlyHint).toContain('异步审计')
    expect(zh.admin.promptAudit.saveBar.auditLatestTurnOnlyHint).toContain('完整对话')
    expect(zh.admin.promptAudit.saveBar.blockingLatestTurnOnlyHint).toContain('同步阻止')
    expect(en.admin.promptAudit.saveBar.auditLatestTurnOnlyHint).toContain('asynchronous audit path')
    expect(en.admin.promptAudit.saveBar.blockingLatestTurnOnlyHint).toContain('synchronous blocking path')
    expect(ru.admin.promptAudit.saveBar.auditLatestTurnOnlyHint).toContain('асинхронного аудита')
    expect(ru.admin.promptAudit.saveBar.blockingLatestTurnOnlyHint).toContain('синхронной блокировки')
  })
})
