/**
 * Admin AliMPay Orders API endpoints
 */

import { apiClient } from '../client'

export interface AdminAliMPayOrderItem {
  id: number
  order_no: string
  trade_no: string
  user_id: number
  user_email: string
  amount: string
  payment_amount: string
  credit_amount: string
  multiplier: number
  status: string
  pay_type: string
  paid_at: string | null
  source_domain: string
  created_at: string
  updated_at: string
  expired_at: string
}

export interface AdminAliMPayOrderListResponse {
  items: AdminAliMPayOrderItem[]
  total: number
  page: number
  page_size: number
}

export async function getAdminAliMPayOrders(params: {
  page?: number
  page_size?: number
  status?: string
  user_id?: number
}): Promise<AdminAliMPayOrderListResponse> {
  const { data } = await apiClient.get<AdminAliMPayOrderListResponse>('/admin/alimpay/orders', { params })
  return data
}

export const alimpayOrdersAPI = { getAdminAliMPayOrders }
export default alimpayOrdersAPI
