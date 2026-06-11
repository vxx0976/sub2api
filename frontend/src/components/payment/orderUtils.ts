/**
 * Shared utility functions for payment order display.
 * Used by AdminOrderDetail, AdminOrderTable, AdminRefundDialog, AdminOrdersView, etc.
 */

const STATUS_BADGE_MAP: Record<string, string> = {
  PENDING: 'badge-warning',
  PAID: 'badge-info',
  RECHARGING: 'badge-info',
  COMPLETED: 'badge-success',
  EXPIRED: 'badge-secondary',
  CANCELLED: 'badge-secondary',
  FAILED: 'badge-danger',
  REFUND_REQUESTED: 'badge-warning',
  REFUNDING: 'badge-warning',
  PARTIALLY_REFUNDED: 'badge-warning',
  REFUNDED: 'badge-info',
  REFUND_FAILED: 'badge-danger',
}

const REFUNDABLE_STATUSES = ['COMPLETED', 'PARTIALLY_REFUNDED', 'REFUND_REQUESTED', 'REFUND_FAILED']

export function statusBadgeClass(status: string): string {
  return STATUS_BADGE_MAP[status] || 'badge-secondary'
}

export function canRefund(status: string): boolean {
  return REFUNDABLE_STATUSES.includes(status)
}

export function formatOrderDateTime(dateStr: string): string {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString()
}

/**
 * 与后端 fee.go 同口径的手续费计算: fee = RoundUp(base * rate / 100, 2)。
 * (baseCents / 100) * rate 直接得到以「分」为单位的手续费(rate 是百分比数值,如 1 表示 1%),
 * 先用 round 抹掉二进制浮点噪声,再 ceil 到整分,避免先乘后 ceil 把误差放大成整 1 分。
 * 后端舍入规则变更时只改这里——PaymentView 与 deriveOrderBaseAmount 共用此实现。
 */
export function calcFeeCents(baseCents: number, feeRatePercent: number): number {
  if (feeRatePercent <= 0 || baseCents <= 0) return 0
  return Math.ceil(Math.round((baseCents / 100) * feeRatePercent * 1e6) / 1e6)
}

/**
 * 从 pay_amount 与 fee_rate(百分比) 精确反推下单基础金额与手续费。
 *
 * 后端口径: pay = base + RoundUp(base * rate / 100, 2)(手续费向上取整到分),
 * 该映射不能按比例除回去(会差 1 分)。但 base ↦ pay 在分位上严格单调递增,
 * 反函数唯一——以比例商为初值,在 ±2 分内搜索使等式精确成立的 base。
 * 注意不能用 order.amount: 它是到账额度(基础额 × 充值倍率),倍率 ≠ 1 时不是实付基础额。
 */
export function deriveOrderBaseAmount(
  payAmount: number,
  feeRatePercent: number
): { base: number; fee: number } {
  const payCents = Math.round(payAmount * 100)
  const rate = feeRatePercent
  if (!Number.isFinite(payCents) || payCents <= 0 || !Number.isFinite(rate) || rate <= 0) {
    return { base: payAmount, fee: 0 }
  }
  const feeCentsFor = (baseCents: number) => calcFeeCents(baseCents, rate)
  const guess = Math.round(payCents / (1 + rate / 100))
  for (const delta of [0, -1, 1, -2, 2]) {
    const baseCents = guess + delta
    if (baseCents <= 0) continue
    if (baseCents + feeCentsFor(baseCents) === payCents) {
      return { base: baseCents / 100, fee: (payCents - baseCents) / 100 }
    }
  }
  // 兜底: 比例估算(异常数据下可能差 1 分)
  const base = Math.round((payAmount / (1 + rate / 100)) * 100) / 100
  return { base, fee: Math.round((payAmount - base) * 100) / 100 }
}
