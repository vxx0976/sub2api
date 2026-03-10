/**
 * Admin Merchants API endpoints
 */

import { apiClient } from '../client'
import type { PaginatedResponse } from '@/types'

export interface MerchantUser {
  id: number
  username: string
  email: string
  balance: number
  merchant_mode: string
  platform_cost: string
  price_multiplier: string
  min_withdrawal: string
}

export interface MerchantSettings {
  merchant_mode?: string
  platform_cost?: string
  min_withdrawal?: string
}

export async function getMerchants(params: {
  page?: number
  page_size?: number
  search?: string
}): Promise<PaginatedResponse<MerchantUser>> {
  const { data } = await apiClient.get<PaginatedResponse<MerchantUser>>('/admin/merchants', { params })
  return data
}

export async function getMerchantSettings(id: number): Promise<Record<string, string>> {
  const { data } = await apiClient.get<Record<string, string>>(`/admin/merchants/${id}/settings`)
  return data
}

export async function updateMerchantSettings(id: number, settings: MerchantSettings): Promise<MerchantUser> {
  const { data } = await apiClient.put<MerchantUser>(`/admin/merchants/${id}/settings`, settings)
  return data
}

export async function updateMerchantBalance(
  id: number,
  balance: number,
  operation: 'set' | 'add' | 'subtract',
  notes?: string
): Promise<void> {
  await apiClient.post(`/admin/merchants/${id}/balance`, {
    balance,
    operation,
    notes
  })
}

export const merchantsAPI = { getMerchants, getMerchantSettings, updateMerchantSettings, updateMerchantBalance }
export default merchantsAPI
