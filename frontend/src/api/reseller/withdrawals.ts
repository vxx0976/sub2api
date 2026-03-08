/**
 * Reseller Withdrawal API endpoints
 */

import { apiClient } from '../client'

export interface WithdrawalItem {
  id: number
  amount: string
  status: string
  payment_method: string
  payment_account: string
  payment_name: string
  admin_notes: string
  created_at: string
  paid_at: string | null
  rejected_at: string | null
}

export interface WithdrawalListResponse {
  items: WithdrawalItem[]
  total: number
  page: number
  page_size: number
  pages: number
}

export async function getWithdrawals(params: {
  page?: number
  page_size?: number
  status?: string
}): Promise<WithdrawalListResponse> {
  const { data } = await apiClient.get<WithdrawalListResponse>('/reseller/withdrawals', { params })
  return data
}

export async function createWithdrawal(payload: {
  amount: number
  payment_method: string
  payment_account: string
  payment_name: string
}): Promise<WithdrawalItem> {
  const { data } = await apiClient.post<WithdrawalItem>('/reseller/withdrawals', payload)
  return data
}

export async function cancelWithdrawal(id: number): Promise<void> {
  await apiClient.delete(`/reseller/withdrawals/${id}`)
}

export const withdrawalsAPI = { getWithdrawals, createWithdrawal, cancelWithdrawal }
export default withdrawalsAPI
