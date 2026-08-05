import { trimTrailingZeros } from './formatters'

/**
 * formatScaled formats a per-token (or per-request) price scaled by `scale`.
 *
 *   formatScaled(0.000003, 1_000_000)         → "$3"      // per 1M tokens
 *   formatScaled(0.5,        1)               → "$0.5"    // per request
 *   formatScaled(null,       1_000_000)       → "-"
 *   formatScaled(0.000003, 1_000_000, 2)      → "$3.00"   // pad to ≥2 decimals
 *   formatScaled(1.25e-8,  1_000_000, 2)      → "$0.0125" // longer decimals kept as-is
 *   formatScaled(0.000003, 1_000_000, 2, '¥') → "¥3.00"   // 人民币官方计价的国产模型
 *
 * Uses toPrecision(10) then strips trailing zeros to avoid IEEE 754 display noise.
 * `minFractionDigits` pads the result back up to a minimum number of decimals.
 * `symbol` 默认 '$'；国产按人民币官方计价的模型传 '¥'（取自后端 price_currency，
 * 见 utils/usagePricing 的 costSymbol）。1¥=1 余额单位，只换符号不换算数值。
 */
export function formatScaled(
  value: number | null,
  scale: number,
  minFractionDigits = 0,
  symbol = '$'
): string {
  if (value == null) return '-'
  let s = trimTrailingZeros((value * scale).toPrecision(10))
  if (minFractionDigits > 0 && !s.includes('e')) {
    const dot = s.indexOf('.')
    const digits = dot === -1 ? 0 : s.length - dot - 1
    if (digits < minFractionDigits) {
      s = (dot === -1 ? `${s}.` : s) + '0'.repeat(minFractionDigits - digits)
    }
  }
  return `${symbol}${s}`
}

/**
 * 默认 CNY/USD 汇率，用于「模型广场」本站价换算在 /pricing/public/fx-rate 拉取
 * 失败时的降级值。本系统采用 1¥=1$ 余额模型，默认为 1（本站价 = 官方价 × 分组倍率，
 * 不额外换汇）；实际值由后端 fx-rate 端点返回。
 */
export const DEFAULT_CNY_PER_USD = 1
