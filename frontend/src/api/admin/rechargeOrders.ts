/**
 * Admin Recharge Orders API endpoints
 */

import { apiClient } from '../client'

export interface AdminRechargeOrderItem {
  id: number
  order_no: string
  trade_no: string
  user_id: number
  user_email: string
  amount: string
  credit_amount: string
  multiplier: number
  status: string
  pay_type: string
  paid_at: string | null
  created_at: string
  updated_at: string
  expired_at: string
}

export interface AdminRechargeOrderListResponse {
  items: AdminRechargeOrderItem[]
  total: number
  page: number
  page_size: number
}

export async function getAdminRechargeOrders(params: {
  page?: number
  page_size?: number
  status?: string
  user_id?: number
}): Promise<AdminRechargeOrderListResponse> {
  const { data } = await apiClient.get<AdminRechargeOrderListResponse>('/admin/recharge/orders', { params })
  return data
}

export const rechargeOrdersAPI = { getAdminRechargeOrders }
export default rechargeOrdersAPI
