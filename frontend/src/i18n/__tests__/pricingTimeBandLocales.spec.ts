import { describe, expect, it } from 'vitest'

import en from '../locales/en'
import ru from '../locales/ru'
import zh from '../locales/zh'

// 本仓是 zh/en/ru 三语，i18n 只加载当前 locale，漏 key 会直接影响该语种用户，
// 因此新增定价文案必须三份同时补齐（这里逐语种断言，缺哪份一眼可见）。
const locales = { zh, en, ru } as const

describe('pricing time band locale keys', () => {
  it('defines the off-peak usage badge in all three locales', () => {
    for (const [name, locale] of Object.entries(locales)) {
      expect(locale.usage.pricingTimeBandOffPeak, `${name}.usage.pricingTimeBandOffPeak`).toBeTruthy()
      expect(
        locale.usage.pricingTimeBandOffPeakHint,
        `${name}.usage.pricingTimeBandOffPeakHint`
      ).toBeTruthy()
    }
  })

  it('defines the model plaza band labels and note in all three locales', () => {
    for (const [name, locale] of Object.entries(locales)) {
      expect(locale.modelPlaza.table.bandPeak, `${name}.modelPlaza.table.bandPeak`).toBeTruthy()
      expect(locale.modelPlaza.table.bandOffPeak, `${name}.modelPlaza.table.bandOffPeak`).toBeTruthy()

      // 说明文案带三个插值变量，缺一个就会在页面上渲染出字面占位符。
      const note = locale.modelPlaza.table.timeTierNote
      expect(note, `${name}.modelPlaza.table.timeTierNote`).toBeTruthy()
      expect(note, `${name} timeTierNote must interpolate {windows}`).toContain('{windows}')
      expect(note, `${name} timeTierNote must interpolate {tz}`).toContain('{tz}')
      expect(note, `${name} timeTierNote must interpolate {percent}`).toContain('{percent}')
    }
  })

  it('keeps zh wording aligned with the official DeepSeek terminology', () => {
    expect(zh.modelPlaza.table.bandPeak).toBe('高峰')
    expect(zh.modelPlaza.table.bandOffPeak).toBe('空闲')
  })
})
