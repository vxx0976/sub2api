import { trimTrailingZeros } from './formatters'

export const TOKENS_PER_MILLION = 1_000_000

interface TokenPriceFormatOptions {
  fractionDigits?: number
  withCurrencySymbol?: boolean
  emptyValue?: string
  /** 货币符号，默认 '$'。国产人民币计价模型传 '¥'。 */
  currencySymbol?: string
}

/**
 * costSymbol 返回某条用量记录应展示的货币符号。
 * 后端 price_currency==='CNY'（国产官方人民币计价模型）→ '¥'，其余（含缺省/未知）→ '$'。
 * 依赖部署 pricing.cny_to_usd_rate=1.0（1¥=1 余额单位），此时存储金额数值即官方人民币数，
 * 直接换符号即与官网一致；若改了汇率则需另行换算（详见后端 ModelPriceCurrency 注释）。
 */
export function costSymbol(currency?: string | null): string {
  return currency === 'CNY' ? '¥' : '$'
}

function isFiniteNumber(value: unknown): value is number {
  return typeof value === 'number' && Number.isFinite(value)
}

export function calculateTokenUnitPrice(
  cost: number | null | undefined,
  tokens: number | null | undefined
): number | null {
  if (!isFiniteNumber(cost) || !isFiniteNumber(tokens) || tokens <= 0) {
    return null
  }

  return cost / tokens
}

export function calculateTokenPricePerMillion(
  cost: number | null | undefined,
  tokens: number | null | undefined
): number | null {
  const unitPrice = calculateTokenUnitPrice(cost, tokens)
  if (unitPrice == null) {
    return null
  }

  return unitPrice * TOKENS_PER_MILLION
}

export function formatTokenPricePerMillion(
  cost: number | null | undefined,
  tokens: number | null | undefined,
  options: TokenPriceFormatOptions = {}
): string {
  const pricePerMillion = calculateTokenPricePerMillion(cost, tokens)
  if (pricePerMillion == null) {
    return options.emptyValue ?? '-'
  }

  const fractionDigits = options.fractionDigits ?? 4
  const formatted = trimTrailingZeros(pricePerMillion.toFixed(fractionDigits))
  const symbol = options.currencySymbol ?? '$'
  return options.withCurrencySymbol == false ? formatted : `${symbol}${formatted}`
}
