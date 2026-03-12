import type { UserSubscription } from '@/types'

// 计算剩余周期数（向上取整，每30天为一个周期）
export function getRemainingCycles(sub: UserSubscription): number {
  if (!sub.expires_at) return 1
  const now = new Date()
  const expires = new Date(sub.expires_at)
  const diff = expires.getTime() - now.getTime()
  const daysRemaining = Math.ceil(diff / (1000 * 60 * 60 * 24))
  if (daysRemaining <= 0) return 0
  return Math.ceil(daysRemaining / 30)
}

// 计算有效月额度（单周期额度 × 剩余周期数）
export function getEffectiveMonthlyLimit(sub: UserSubscription): number {
  const monthlyLimit = sub.group?.monthly_limit_usd || 0
  const cycles = getRemainingCycles(sub)
  return monthlyLimit * cycles
}
