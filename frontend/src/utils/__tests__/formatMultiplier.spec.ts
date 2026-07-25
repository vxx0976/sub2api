import { describe, expect, it } from 'vitest'
import { formatMultiplier } from '../formatters'

describe('formatMultiplier', () => {
  it('keeps significant decimals instead of rounding to 2 places', () => {
    expect(formatMultiplier(0.035)).toBe('0.035')
    expect(formatMultiplier(0.125)).toBe('0.125')
    expect(formatMultiplier(0.015)).toBe('0.015')
    expect(formatMultiplier(0.0625)).toBe('0.0625')
  })

  // fork：dev 的展示约定是裁掉无意义尾随 0（见 utils/formatters.ts trimTrailingZeros
  // 与 formatters.spec.ts），与上游「补齐到 2 位小数」相反。倍率展示统一走裁零，
  // 故此处按 dev 约定断言；上游「保留有效小数」的修复已在上一条用例中保留。
  it('trims trailing zeros for round values (fork convention)', () => {
    expect(formatMultiplier(0.3)).toBe('0.3')
    expect(formatMultiplier(1)).toBe('1')
    expect(formatMultiplier(1.5)).toBe('1.5')
    expect(formatMultiplier(2)).toBe('2')
  })

  it('handles small values down to 4 decimals', () => {
    expect(formatMultiplier(0.001)).toBe('0.001')
    expect(formatMultiplier(0.0001)).toBe('0.0001')
  })

  it('falls back to 2 significant digits below 0.0001', () => {
    expect(formatMultiplier(0.00005)).toBe('0.00005')
  })
})
