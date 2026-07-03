import { describe, it, expect } from 'vitest'
import { trimTrailingZeros, formatMultiplier } from './formatters'

describe('trimTrailingZeros', () => {
  it('去掉定点小数末尾无意义的 0', () => {
    expect(trimTrailingZeros('100.000000')).toBe('100')
    expect(trimTrailingZeros('100.500000')).toBe('100.5')
    expect(trimTrailingZeros('1.230000')).toBe('1.23')
    expect(trimTrailingZeros('0.000000')).toBe('0')
  })

  it('保留有效小数（含链上匹配的唯一尾数）', () => {
    expect(trimTrailingZeros('100.001230')).toBe('100.00123')
    expect(trimTrailingZeros('7.2')).toBe('7.2')
    expect(trimTrailingZeros('-3.500000')).toBe('-3.5')
  })

  it('无小数点的整数字符串原样返回，不被误删末尾 0', () => {
    expect(trimTrailingZeros('100')).toBe('100')
    expect(trimTrailingZeros('0')).toBe('0')
    expect(trimTrailingZeros('')).toBe('')
  })

  it('科学计数法原样返回，不破坏指数里的 0', () => {
    expect(trimTrailingZeros('1.000000000e+10')).toBe('1.000000000e+10')
    expect(trimTrailingZeros('5e-7')).toBe('5e-7')
    expect(trimTrailingZeros('1.5E+20')).toBe('1.5E+20')
  })
})

describe('formatMultiplier', () => {
  it('裁掉倍率末尾无意义的 0', () => {
    expect(formatMultiplier(1)).toBe('1')
    expect(formatMultiplier(2.5)).toBe('2.5')
    expect(formatMultiplier(1.25)).toBe('1.25')
  })

  it('小数倍率保留精度', () => {
    expect(formatMultiplier(0.005)).toBe('0.005')
    expect(formatMultiplier(0.001)).toBe('0.001')
  })
})
