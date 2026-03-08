/**
 * Admin Merchant Withdrawals API endpoints
 */

import { apiClient } from '../client'

export interface AdminWithdrawalItem {
  id: number
  reseller_id: number
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

export interface AdminWithdrawalListResponse {
  items: AdminWithdrawalItem[]
  total: number
  page: number
  page_size: number
  pages: number
}

export async function getAdminWithdrawals(params: {
  page?: number
  page_size?: number
  status?: string
  reseller_id?: number
}): Promise<AdminWithdrawalListResponse> {
  const { data } = await apiClient.get<AdminWithdrawalListResponse>('/admin/withdrawals', { params })
  return data
}

export async function payWithdrawal(id: number, payload: { admin_notes?: string }): Promise<AdminWithdrawalItem> {
  const { data } = await apiClient.put<AdminWithdrawalItem>(`/admin/withdrawals/${id}/pay`, payload)
  return data
}

export async function rejectWithdrawal(id: number, payload: { admin_notes?: string }): Promise<AdminWithdrawalItem> {
  const { data } = await apiClient.put<AdminWithdrawalItem>(`/admin/withdrawals/${id}/reject`, payload)
  return data
}

export const merchantWithdrawalsAPI = { getAdminWithdrawals, payWithdrawal, rejectWithdrawal }
export default merchantWithdrawalsAPI
