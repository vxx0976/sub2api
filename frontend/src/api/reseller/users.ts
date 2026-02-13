import apiClient from '@/api/client'
import type { User } from '@/types'

export interface ResellerUserListResponse {
  items: User[]
  total: number
  page: number
  page_size: number
  pages: number
}

export interface ResellerBalanceHistoryItem {
  id: number
  code: string
  type: string
  value: number
  status: string
  used_at: string | null
  created_at: string
  notes: string
  validity_days: number
  group?: { id: number; name: string } | null
}

export interface ResellerBalanceHistoryResponse {
  items: ResellerBalanceHistoryItem[]
  total: number
  page: number
  page_size: number
  pages: number
  total_recharged: number
}

export const resellerUsersAPI = {
  async list(params: { page?: number; page_size?: number; search?: string }) {
    const { data } = await apiClient.get<ResellerUserListResponse>('/reseller/users', { params })
    return data
  },

  async updateBalance(userId: number, amount: number, operation: 'add' | 'subtract', notes?: string): Promise<User> {
    const { data } = await apiClient.post<User>(`/reseller/users/${userId}/balance`, {
      amount,
      operation,
      notes: notes || ''
    })
    return data
  },

  async getBalanceHistory(userId: number, page: number = 1, pageSize: number = 15, type?: string): Promise<ResellerBalanceHistoryResponse> {
    const params: Record<string, any> = { page, page_size: pageSize }
    if (type) params.type = type
    const { data } = await apiClient.get<ResellerBalanceHistoryResponse>(
      `/reseller/users/${userId}/balance-history`,
      { params }
    )
    return data
  }
}

export default resellerUsersAPI
