import { trimTrailingZeros } from './formatters'

/**
 * formatScaled formats a per-token (or per-request) USD price scaled by `scale`.
 *
 *   formatScaled(0.000003, 1_000_000) → "$3"        // per 1M tokens
 *   formatScaled(0.5,        1)        → "$0.5"      // per request
 *   formatScaled(null,       1_000_000) → "-"
 *
 * Uses toPrecision(10) then strips trailing zeros to avoid IEEE 754 display noise.
 */
export function formatScaled(value: number | null, scale: number): string {
  if (value == null) return '-'
  return `$${trimTrailingZeros((value * scale).toPrecision(10))}`
}

/**
 * 默认 CNY/USD 汇率，用于「模型广场」本站价换算在 /pricing/public/fx-rate 拉取
 * 失败时的降级值。本系统采用 1¥=1$ 余额模型，默认为 1（本站价 = 官方价 × 分组倍率，
 * 不额外换汇）；实际值由后端 fx-rate 端点返回。
 */
export const DEFAULT_CNY_PER_USD = 1
