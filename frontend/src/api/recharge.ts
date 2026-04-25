/**
 * Recharge API endpoints
 * Handles balance recharge via payment gateway
 */

import { apiClient } from './client'

export interface RechargeConfig {
  enabled: boolean
  min_amount: number
  max_amount: number
  pay_types: string[]
  selling_price: number
}

export interface RechargeOrderResult {
  order_no: string
  amount: number
  credit_amount: number
  pay_url: string
  qrcode: string
  expired_at: string
}

export interface RechargeOrderStatus {
  order_no: string
  amount: number
  credit_amount: number
  status: string
  paid_at: string | null
}

export interface RechargeOrderItem {
  id: number
  order_no: string
  trade_no: string
  user_id: number
  amount: number
  credit_amount: number
  status: string
  pay_type: string
  paid_at: string | null
  created_at: string
  updated_at: string
  expired_at: string | null
}

/**
 * Get recharge page configuration
 */
export async function getConfig(): Promise<RechargeConfig> {
  const { data } = await apiClient.get<RechargeConfig>('/recharge/config')
  return data
}

/**
 * Create a recharge order
 */
export async function createOrder(amount: number, payType: string): Promise<RechargeOrderResult> {
  const { data } = await apiClient.post<RechargeOrderResult>('/recharge/create', {
    amount,
    pay_type: payType
  })
  return data
}

/**
 * Get order status
 */
export async function getOrderStatus(orderNo: string): Promise<RechargeOrderStatus> {
  const { data } = await apiClient.get<RechargeOrderStatus>(`/recharge/status/${orderNo}`)
  return data
}

/**
 * List user's recharge orders
 */
export async function listOrders(page: number = 1, pageSize: number = 20): Promise<{
  items: RechargeOrderItem[]
  total: number
  page: number
  page_size: number
}> {
  const { data } = await apiClient.get<{
    items: RechargeOrderItem[]
    total: number
    page: number
    page_size: number
  }>('/recharge/orders', { params: { page, page_size: pageSize } })
  return data
}

export const rechargeAPI = {
  getConfig,
  createOrder,
  getOrderStatus,
  listOrders
}

export default rechargeAPI
