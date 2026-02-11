import { apiClient } from './client'

// ========== 数据类型 ==========

export interface RechargeTier {
  min: number
  max: number | null
  multiplier: number
}

export interface RechargeConfig {
  enabled: boolean
  min_amount: number
  max_amount: number
  tiers: RechargeTier[]
}

export interface RechargeOrder {
  id: number
  order_no: string
  trade_no?: string
  user_id: number
  amount: number
  credit_amount: number
  multiplier: number
  status: 'pending' | 'paid' | 'expired' | 'refunded'
  pay_type?: string
  paid_at?: string
  created_at: string
  updated_at: string
  expired_at?: string
  user?: {
    id: number
    username: string
    email?: string
  }
}

export interface CreateRechargeOrderRequest {
  amount: number
}

export interface CreateRechargeOrderResponse {
  order_no: string
  amount: number
  payment_amount: number
  qr_code_url: string
  qr_code: string
  mode: string
  credit_amount: number
  multiplier: number
}

// ========== API 函数 ==========

/**
 * 获取充值配置
 */
export async function getRechargeConfig(): Promise<RechargeConfig> {
  const response = await apiClient.get<RechargeConfig>('/recharge/config')
  return response.data
}

/**
 * 创建充值订单
 */
export async function createRechargeOrder(
  amount: number
): Promise<CreateRechargeOrderResponse> {
  const response = await apiClient.post<CreateRechargeOrderResponse>(
    '/recharge/orders',
    { amount }
  )
  return response.data
}

/**
 * 获取充值订单列表
 */
export async function getRechargeOrders(
  page: number = 1,
  pageSize: number = 20
): Promise<{ items: RechargeOrder[]; total: number }> {
  const response = await apiClient.get<{ items: RechargeOrder[]; total: number }>(
    '/recharge/orders',
    { params: { page, page_size: pageSize } }
  )
  return response.data
}

/**
 * 继续支付充值订单
 */
export async function repayRechargeOrder(
  orderNo: string
): Promise<CreateRechargeOrderResponse> {
  const response = await apiClient.post<CreateRechargeOrderResponse>(
    `/recharge/orders/${orderNo}/repay`
  )
  return response.data
}

// ========== 管理员 API ==========

/**
 * 获取充值配置（管理员）
 */
export async function getRechargeConfigAdmin(): Promise<RechargeConfig> {
  const response = await apiClient.get<RechargeConfig>('/admin/settings/recharge')
  return response.data
}

/**
 * 更新充值配置（管理员）
 */
export async function updateRechargeConfig(config: RechargeConfig): Promise<void> {
  await apiClient.put('/admin/settings/recharge', config)
}

/**
 * 获取所有充值订单（管理员）
 */
export async function getAllRechargeOrders(
  page: number = 1,
  pageSize: number = 20,
  filters?: {
    user_id?: number
    status?: string
    keyword?: string
  }
): Promise<{ items: RechargeOrder[]; total: number }> {
  const response = await apiClient.get<{ items: RechargeOrder[]; total: number }>(
    '/admin/recharge/orders',
    { params: { page, page_size: pageSize, ...filters } }
  )
  return response.data
}

export const rechargeAPI = {
  getConfig: getRechargeConfig,
  createOrder: createRechargeOrder,
  getOrders: getRechargeOrders,
  repayOrder: repayRechargeOrder,
  getConfigAdmin: getRechargeConfigAdmin,
  updateConfig: updateRechargeConfig,
  getAllOrders: getAllRechargeOrders
}

export default rechargeAPI
