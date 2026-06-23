/**
 * 统一充值订单（topup）API
 * 把 充值(EPAY) / 支付宝个人免签(AliMPay) / USDT 三套订单按时间混排到一张表。
 * - 用户端：GET /topup/orders        （只看自己的）
 * - 管理端：GET /admin/topup/orders  （看全部，含 user_email、可按 user_id 过滤）
 * 退款仍按通道分发到各自原有的管理端退款接口。
 */
import { apiClient } from './client'
import rechargeOrdersAPI from './admin/rechargeOrders'
import alimpayOrdersAPI from './admin/alimpayOrders'
import usdtOrdersAPI from './admin/usdtOrders'

export type TopupChannel = 'recharge' | 'alimpay' | 'usdt' | 'manual'

export interface MergedTopupOrder {
  channel: TopupChannel
  id: number
  order_no: string
  trade_no: string
  user_id: number
  user_email: string
  amount: string
  credit_amount: string
  status: string
  pay_type: string
  created_at: string
  paid_at: string | null
  expired_at: string | null
  // 通道特有字段（不适用时为 null）
  payment_amount: string | null
  usdt_amount_str: string | null
  usdt_rate: string | null
  usdt_chain: string | null
  source_domain: string | null
  note: string | null // 仅 manual：管理员调整备注/原因
}

export interface MergedTopupOrderListResponse {
  items: MergedTopupOrder[]
  total: number
  page: number
  page_size: number
  pages?: number
}

export interface TopupOrderQuery {
  page?: number
  page_size?: number
  channel?: TopupChannel | ''
  status?: string
  user_id?: number
}

function buildParams(q: TopupOrderQuery): Record<string, any> {
  const params: Record<string, any> = {
    page: q.page ?? 1,
    page_size: q.page_size ?? 20
  }
  if (q.channel) params.channel = q.channel
  if (q.status) params.status = q.status
  if (q.user_id) params.user_id = q.user_id
  return params
}

export async function getUserTopupOrders(q: TopupOrderQuery): Promise<MergedTopupOrderListResponse> {
  const { data } = await apiClient.get<MergedTopupOrderListResponse>('/topup/orders', { params: buildParams(q) })
  return data
}

export async function getAdminTopupOrders(q: TopupOrderQuery): Promise<MergedTopupOrderListResponse> {
  const { data } = await apiClient.get<MergedTopupOrderListResponse>('/admin/topup/orders', { params: buildParams(q) })
  return data
}

/** 退款：按通道分发到各自原有的管理端退款接口 */
export async function refundTopupOrder(channel: TopupChannel, orderNo: string, reason?: string) {
  switch (channel) {
    case 'recharge':
      return rechargeOrdersAPI.refundAdminRechargeOrder(orderNo, reason)
    case 'alimpay':
      return alimpayOrdersAPI.refundAdminAliMPayOrder(orderNo, reason)
    case 'usdt':
      return usdtOrdersAPI.refundAdminUsdtOrder(orderNo, reason)
    default:
      // manual（手工调整）等通道不支持订单退款；UI 不应触达此分支，显式抛错避免静默"成功"。
      throw new Error(`refund not supported for channel: ${channel}`)
  }
}

export const topupOrdersAPI = {
  getUserTopupOrders,
  getAdminTopupOrders,
  refundTopupOrder
}

export default topupOrdersAPI
