import apiClient from '@/api/client'
import type { User } from '@/types'

export interface ResellerUserListResponse {
  items: User[]
  total: number
  page: number
  page_size: number
  pages: number
}

export const resellerUsersAPI = {
  async list(params: { page?: number; page_size?: number; search?: string }) {
    const { data } = await apiClient.get<ResellerUserListResponse>('/reseller/users', { params })
    return data
  }
}

export default resellerUsersAPI
