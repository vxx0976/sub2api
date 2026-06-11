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

export interface RefundAliMPayOrderResponse {
  order_no: string
  user_id: number
  credit_amount: number
  status: string
}

// 退款：把已支付订单标记为 refunded 并扣回到账余额（连带回冲商户佣金基数）。
export async function refundAdminAliMPayOrder(
  orderNo: string,
  reason?: string
): Promise<RefundAliMPayOrderResponse> {
  const { data } = await apiClient.post<RefundAliMPayOrderResponse>(
    `/admin/alimpay/orders/${encodeURIComponent(orderNo)}/refund`,
    { reason: reason ?? '' }
  )
  return data
}

export const alimpayOrdersAPI = { getAdminAliMPayOrders, refundAdminAliMPayOrder }
export default alimpayOrdersAPI
