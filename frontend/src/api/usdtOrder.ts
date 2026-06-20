/**
 * USDT (TRC20) 自建收款 Order API
 * 与 /api/alimpay/* 并列：业务同为"CNY 计价 → 余额充值"，差异在支付通道。
 * USDT 走链上轮询（无 webhook），前端同样 5s 轮询订单状态。
 */

import { apiClient } from './client'

export interface UsdtConfig {
  enabled: boolean
  min_amount: number
  max_amount: number
  chains: string[] // 当前可收款的链：trc20 / bep20 / ton
  rate: number // 当前换算汇率：1 USDT = ? CNY
}

export interface UsdtOrderResult {
  order_no: string
  amount: number
  credit_amount: number
  chain: string
  address: string
  usdt_amount: number
  usdt_amount_str: string
  rate: number
  expires_in: number
  expired_at: string
}

export interface UsdtOrderStatus {
  order_no: string
  amount: number
  credit_amount: number
  chain: string
  address: string
  usdt_amount: number
  usdt_amount_str: string
  status: string
  paid_at: string | null
  trade_no: string
  expired_at: string | null
}

export interface UsdtOrderItem {
  id: number
  order_no: string
  trade_no: string
  user_id: number
  amount: number
  credit_amount: number
  chain: string
  usdt_amount: number
  usdt_amount_str: string
  usdt_rate: number
  paid_usdt_amount?: number
  paid_usdt_amount_str?: string
  status: string
  pay_type: string
  paid_at: string | null
  created_at: string
  updated_at: string
  expired_at: string | null
}

export async function getConfig(): Promise<UsdtConfig> {
  const { data } = await apiClient.get<UsdtConfig>('/usdt/config')
  return data
}

export async function createOrder(amount: number, chain: string): Promise<UsdtOrderResult> {
  const { data } = await apiClient.post<UsdtOrderResult>('/usdt/create', { amount, chain })
  return data
}

export async function getOrderStatus(orderNo: string): Promise<UsdtOrderStatus> {
  const { data } = await apiClient.get<UsdtOrderStatus>(`/usdt/status/${orderNo}`)
  return data
}

export async function listOrders(
  page = 1,
  pageSize = 20
): Promise<{
  items: UsdtOrderItem[]
  total: number
  page: number
  page_size: number
}> {
  const { data } = await apiClient.get<{
    items: UsdtOrderItem[]
    total: number
    page: number
    page_size: number
  }>('/usdt/orders', { params: { page, page_size: pageSize } })
  return data
}

export const usdtAPI = {
  getConfig,
  createOrder,
  getOrderStatus,
  listOrders
}

export default usdtAPI
