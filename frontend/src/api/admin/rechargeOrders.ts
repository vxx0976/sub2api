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

export interface RefundRechargeOrderResponse {
  order_no: string
  user_id: number
  credit_amount: number
  status: string
}

// 退款：把已支付订单标记为 refunded 并扣回到账余额（连带回冲商户佣金基数）。
export async function refundAdminRechargeOrder(
  orderNo: string,
  reason?: string
): Promise<RefundRechargeOrderResponse> {
  const { data } = await apiClient.post<RefundRechargeOrderResponse>(
    `/admin/recharge/orders/${encodeURIComponent(orderNo)}/refund`,
    { reason: reason ?? '' }
  )
  return data
}

export const rechargeOrdersAPI = { getAdminRechargeOrders, refundAdminRechargeOrder }
export default rechargeOrdersAPI
