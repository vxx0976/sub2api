/**
 * Admin USDT (TRC20) Orders API endpoints
 */

import { apiClient } from '../client'

export interface AdminUsdtOrderItem {
  id: number
  order_no: string
  trade_no: string
  user_id: number
  user_email: string
  amount: string
  credit_amount: string
  chain: string
  usdt_amount: number
  usdt_amount_str: string
  usdt_rate: number
  paid_usdt_amount?: number
  paid_usdt_amount_str?: string
  receiving_address: string
  from_address: string
  status: string
  pay_type: string
  paid_at: string | null
  source_domain: string
  created_at: string
  updated_at: string
  expired_at: string
}

export interface AdminUsdtOrderListResponse {
  items: AdminUsdtOrderItem[]
  total: number
  page: number
  page_size: number
}

export async function getAdminUsdtOrders(params: {
  page?: number
  page_size?: number
  status?: string
  user_id?: number
}): Promise<AdminUsdtOrderListResponse> {
  const { data } = await apiClient.get<AdminUsdtOrderListResponse>('/admin/usdt/orders', { params })
  return data
}

export interface RefundUsdtOrderResponse {
  order_no: string
  user_id: number
  credit_amount: number
  status: string
}

// 退款：把已支付订单标记为 refunded 并扣回到账余额（链上不可逆，仅冲账户余额）。
export async function refundAdminUsdtOrder(
  orderNo: string,
  reason?: string
): Promise<RefundUsdtOrderResponse> {
  const { data } = await apiClient.post<RefundUsdtOrderResponse>(
    `/admin/usdt/orders/${encodeURIComponent(orderNo)}/refund`,
    { reason: reason ?? '' }
  )
  return data
}

export const usdtOrdersAPI = { getAdminUsdtOrders, refundAdminUsdtOrder }
export default usdtOrdersAPI
