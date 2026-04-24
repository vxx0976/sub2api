/**
 * AliMPay (Alipay 个人免签) Order API
 * 与 /api/recharge/* 并列：业务同为"CNY → USD 余额充值"，差异在支付通道
 * AliMPay 走账单轮询（无 webhook），前端同样 5s 轮询订单状态
 */

import { apiClient } from './client'
import type { RechargeTier } from './recharge'

export interface AliMPayConfig {
  enabled: boolean
  min_amount: number
  max_amount: number
  tiers: RechargeTier[]
  selling_price: number
  mode: 'business_qr' | 'transfer' | ''
}

export interface AliMPayOrderResult {
  order_no: string
  amount: number
  payment_amount: number   // 实际支付金额（经过金额偏移的唯一金额）
  credit_amount: number
  multiplier: number
  qrcode_url: string
  mode: 'business_qr' | 'transfer' | ''
  expires_in: number
  expired_at: string
}

export interface AliMPayOrderStatus {
  order_no: string
  amount: number
  payment_amount: number
  credit_amount: number
  status: string
  paid_at: string | null
}

export interface AliMPayOrderItem {
  id: number
  order_no: string
  trade_no: string
  user_id: number
  amount: number
  payment_amount: number
  credit_amount: number
  multiplier: number
  status: string
  pay_type: string
  paid_at: string | null
  created_at: string
  updated_at: string
  expired_at: string | null
}

export async function getConfig(): Promise<AliMPayConfig> {
  const { data } = await apiClient.get<AliMPayConfig>('/alimpay/config')
  return data
}

export async function createOrder(amount: number): Promise<AliMPayOrderResult> {
  const { data } = await apiClient.post<AliMPayOrderResult>('/alimpay/create', { amount })
  return data
}

export async function getOrderStatus(orderNo: string): Promise<AliMPayOrderStatus> {
  const { data } = await apiClient.get<AliMPayOrderStatus>(`/alimpay/status/${orderNo}`)
  return data
}

export async function listOrders(
  page = 1,
  pageSize = 20
): Promise<{
  items: AliMPayOrderItem[]
  total: number
  page: number
  page_size: number
}> {
  const { data } = await apiClient.get<{
    items: AliMPayOrderItem[]
    total: number
    page: number
    page_size: number
  }>('/alimpay/orders', { params: { page, page_size: pageSize } })
  return data
}

export const alimpayAPI = {
  getConfig,
  createOrder,
  getOrderStatus,
  listOrders
}

export default alimpayAPI
