/**
 * Reseller Commission API endpoints
 */

import { apiClient } from '../client'

export interface CommissionSummary {
  commission_rate: number
  total_cost: string
  total_commission: string
  total_recharge: string
  total_users: number
  today_new_users: number
  today_cost: number
  withdrawn: string
  pending: string
  available: string
}

export interface CommissionDetailItem {
  id: number
  user_id: number
  model: string
  total_cost: string
  merchant_rate_snapshot: string
  platform_cost_snapshot?: string
  commission: string
  created_at: string
}

export interface RechargeDetailItem {
  user_id: number
  order_no: string
  credit_amount: number
  paid_at: string
}

export interface CommissionDetailResponse {
  items: CommissionDetailItem[]
  total: number
  page: number
  page_size: number
  pages: number
}

export interface RechargeDetailResponse {
  items: RechargeDetailItem[]
  total: number
  page: number
  page_size: number
  pages: number
}

export async function getCommissionSummary(): Promise<CommissionSummary> {
  const { data } = await apiClient.get<CommissionSummary>('/reseller/commissions/summary')
  return data
}

export async function getCommissionDetail(params: {
  page?: number
  page_size?: number
  start_date?: string
  end_date?: string
  user_id?: number
}): Promise<CommissionDetailResponse> {
  const { data } = await apiClient.get<CommissionDetailResponse>('/reseller/commissions/detail', { params })
  return data
}

export async function getRechargeDetail(params: {
  page?: number
  page_size?: number
}): Promise<RechargeDetailResponse> {
  const { data } = await apiClient.get<RechargeDetailResponse>('/reseller/commissions/recharges', { params })
  return data
}

export const commissionsAPI = { getCommissionSummary, getCommissionDetail, getRechargeDetail }
export default commissionsAPI
