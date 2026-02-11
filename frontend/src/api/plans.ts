/**
 * Plans and Payment API
 * API for managing purchasable plans and orders
 */

import { apiClient } from './client'

/**
 * Purchasable plan
 */
export interface PurchasablePlan {
  id: number
  name: string
  description: string
  price: number
  validity_days: number
  daily_limit_usd?: number
  weekly_limit_usd?: number
  monthly_limit_usd?: number
  sort_order: number
  is_recommended: boolean
  external_buy_url?: string
}

/**
 * Create order response
 */
export interface CreateOrderResponse {
  order_no: string
  amount: number
  payment_amount: number
  qr_code_url: string
  qr_code: string
  mode: string
}

/**
 * Order status
 */
export interface Order {
  id: number
  order_no: string
  trade_no?: string
  user_id: number
  group_id: number
  amount: number
  status: 'pending' | 'paid' | 'expired' | 'refunded'
  pay_type?: string
  paid_at?: string
  subscription_id?: number
  created_at: string
  updated_at: string
  expired_at?: string
  group?: {
    id: number
    name: string
  }
}

/**
 * Get list of purchasable plans
 */
export async function getPlans(): Promise<PurchasablePlan[]> {
  const response = await apiClient.get<PurchasablePlan[]>('/plans')
  return response.data
}

/**
 * Create a new order
 */
export async function createOrder(groupId: number): Promise<CreateOrderResponse> {
  const response = await apiClient.post<CreateOrderResponse>('/orders', { group_id: groupId })
  return response.data
}

/**
 * Get user's orders
 */
export async function getOrders(page: number = 1, pageSize: number = 20): Promise<{ items: Order[]; total: number }> {
  const response = await apiClient.get<{ items: Order[]; total: number }>('/orders', {
    params: { page, page_size: pageSize }
  })
  return response.data
}

/**
 * Repay an existing pending order
 */
export async function repayOrder(orderNo: string): Promise<CreateOrderResponse> {
  const response = await apiClient.post<CreateOrderResponse>(`/orders/${orderNo}/repay`)
  return response.data
}

/**
 * Payment status response
 */
export interface PaymentStatusResponse {
  status: string
  order_no: string
}

/**
 * Get order payment status (for polling)
 */
export async function getOrderPaymentStatus(orderNo: string): Promise<PaymentStatusResponse> {
  const response = await apiClient.get<PaymentStatusResponse>(`/orders/${orderNo}/payment-status`)
  return response.data
}

export default {
  getPlans,
  createOrder,
  getOrders,
  repayOrder,
  getOrderPaymentStatus
}
