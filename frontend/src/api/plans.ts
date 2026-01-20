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
}

/**
 * Create order response
 */
export interface CreateOrderResponse {
  order_no: string
  pay_url: string
  amount: number
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
 * Payment return parameters
 */
export interface PaymentReturnParams {
  pid: string
  trade_no: string
  out_trade_no: string
  type: string
  name: string
  money: string
  trade_status: string
  sign: string
  sign_type: string
}

/**
 * Verify payment return response
 */
export interface VerifyPaymentResponse {
  status: string
  order_no: string
  paid: boolean
}

/**
 * Verify payment return and process order
 */
export async function verifyPayment(params: PaymentReturnParams): Promise<VerifyPaymentResponse> {
  const response = await apiClient.post<VerifyPaymentResponse>('/payment/verify', params)
  return response.data
}

/**
 * Repay an existing pending order
 */
export async function repayOrder(orderNo: string): Promise<CreateOrderResponse> {
  const response = await apiClient.post<CreateOrderResponse>(`/orders/${orderNo}/repay`)
  return response.data
}

export default {
  getPlans,
  createOrder,
  getOrders,
  verifyPayment,
  repayOrder
}
