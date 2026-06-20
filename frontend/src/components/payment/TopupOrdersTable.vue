<template>
  <div class="card overflow-hidden">
    <div class="overflow-x-auto">
      <table class="w-full">
        <thead>
          <tr class="border-b border-gray-100 dark:border-dark-700">
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('topupOrders.method') }}</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('topupOrders.orderNo') }}</th>
            <th v-if="admin" class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('topupOrders.user') }}</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('topupOrders.amount') }}</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('topupOrders.creditAmount') }}</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('topupOrders.status') }}</th>
            <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('topupOrders.createdAt') }}</th>
            <th v-if="admin" class="px-4 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400">{{ t('topupOrders.actions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="item in items"
            :key="`${item.channel}-${item.id}`"
            class="border-b border-gray-50 dark:border-dark-800"
          >
            <!-- 通道/方式 -->
            <td class="px-4 py-3">
              <span class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium" :class="methodBadgeClass(item)">
                {{ methodLabel(item) }}
              </span>
            </td>
            <!-- 订单号 -->
            <td class="px-4 py-3">
              <span class="font-mono text-xs text-gray-700 dark:text-gray-300">{{ item.order_no }}</span>
            </td>
            <!-- 用户（管理端） -->
            <td v-if="admin" class="px-4 py-3">
              <div>
                <span class="text-sm text-gray-900 dark:text-white">{{ item.user_email || '—' }}</span>
                <span class="ml-1 text-xs text-gray-400 dark:text-gray-500">#{{ item.user_id }}</span>
              </div>
            </td>
            <!-- 金额（通道相关口径） -->
            <td class="px-4 py-3">
              <span class="text-sm font-medium text-gray-900 dark:text-white">{{ displayAmount(item) }}</span>
            </td>
            <!-- 到账（USD） -->
            <td class="px-4 py-3">
              <span class="text-sm font-medium text-emerald-600 dark:text-emerald-400">${{ item.credit_amount }}</span>
            </td>
            <!-- 状态 -->
            <td class="px-4 py-3">
              <span class="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium" :class="statusClass(item.status)">
                {{ statusLabel(item.status) }}
              </span>
            </td>
            <!-- 时间 -->
            <td class="px-4 py-3">
              <span class="text-sm text-gray-500 dark:text-gray-400">{{ formatDateTime(item.created_at) }}</span>
            </td>
            <!-- 操作（管理端退款） -->
            <td v-if="admin" class="px-4 py-3 text-right">
              <button
                v-if="item.status === 'paid'"
                type="button"
                class="text-sm font-medium text-red-600 hover:text-red-700 dark:text-red-400"
                @click="emit('refund', item)"
              >
                {{ t('topupOrders.refund') }}
              </button>
              <span v-else class="text-sm text-gray-300 dark:text-gray-600">—</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { formatDateTime } from '@/utils/format'
import type { MergedTopupOrder } from '@/api/topupOrders'

withDefaults(defineProps<{ items: MergedTopupOrder[]; admin?: boolean }>(), { admin: false })
const emit = defineEmits<{ (e: 'refund', item: MergedTopupOrder): void }>()

const { t } = useI18n()

const CHAIN_LABELS: Record<string, string> = { trc20: 'TRC20', bep20: 'BEP20', ton: 'TON' }

/** 把 (channel, pay_type) 归一化为面向用户的方式：alipay / wechat / usdt */
function methodKind(item: MergedTopupOrder): 'alipay' | 'wechat' | 'usdt' | 'other' {
  if (item.channel === 'usdt') return 'usdt'
  if (item.channel === 'alimpay') return 'alipay'
  if (item.pay_type === 'wxpay') return 'wechat'
  if (item.pay_type === 'alipay') return 'alipay'
  return 'other'
}

function methodLabel(item: MergedTopupOrder): string {
  const kind = methodKind(item)
  if (kind === 'usdt') {
    const chain = item.usdt_chain ? CHAIN_LABELS[item.usdt_chain] || item.usdt_chain.toUpperCase() : ''
    return chain ? `USDT · ${chain}` : 'USDT'
  }
  if (kind === 'wechat') return t('topupOrders.methodWechat')
  if (kind === 'alipay') return t('topupOrders.methodAlipay')
  return t('topupOrders.methodOther')
}

function methodBadgeClass(item: MergedTopupOrder): string {
  switch (methodKind(item)) {
    case 'alipay':
      return 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
    case 'wechat':
      return 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
    case 'usdt':
      return 'bg-teal-100 text-teal-700 dark:bg-teal-900/30 dark:text-teal-400'
    default:
      return 'bg-gray-100 text-gray-700 dark:bg-dark-700 dark:text-gray-300'
  }
}

/** 金额展示口径：USDT 显示 USDT 数量，其余显示人民币（支付宝免签用实际支付金额） */
function displayAmount(item: MergedTopupOrder): string {
  if (item.channel === 'usdt') {
    return item.usdt_amount_str ? `${item.usdt_amount_str} USDT` : `${item.amount} USDT`
  }
  if (item.channel === 'alimpay' && item.payment_amount) {
    return `¥${item.payment_amount}`
  }
  return `¥${item.amount}`
}

function statusClass(status: string): string {
  switch (status) {
    case 'pending':
      return 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400'
    case 'paid':
      return 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
    case 'expired':
      return 'bg-gray-100 text-gray-700 dark:bg-gray-900/30 dark:text-gray-400'
    case 'refunded':
      return 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
    default:
      return 'bg-gray-100 text-gray-700 dark:bg-gray-900/30 dark:text-gray-400'
  }
}

function statusLabel(status: string): string {
  switch (status) {
    case 'pending':
      return t('topupOrders.statusPending')
    case 'paid':
      return t('topupOrders.statusPaid')
    case 'expired':
      return t('topupOrders.statusExpired')
    case 'refunded':
      return t('topupOrders.statusRefunded')
    default:
      return status
  }
}
</script>
