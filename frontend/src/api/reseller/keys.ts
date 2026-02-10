/**
 * Reseller Key API endpoints
 */

import { apiClient } from '../client'
import type { PaginatedResponse } from '@/types'

export interface ResellerKey {
  id: number
  user_id: number
  key: string
  name: string
  notes: string
  group_id: number | null
  status: string
  quota: number
  quota_used: number
  expires_at: string | null
  ip_whitelist: string[]
  ip_blacklist: string[]
  created_at: string
  updated_at: string
  group?: {
    id: number
    name: string
  }
}

export interface CreateResellerKeyRequest {
  name: string
  notes?: string
  group_id?: number
  custom_key?: string
  quota?: number
  expires_in_days?: number
  ip_whitelist?: string[]
  ip_blacklist?: string[]
}

export interface UpdateResellerKeyRequest {
  name?: string
  notes?: string
  group_id?: number
  status?: string
  quota?: number
  expires_at?: string
  ip_whitelist?: string[]
  ip_blacklist?: string[]
}

/**
 * List reseller's API keys
 */
export async function list(
  page: number = 1,
  pageSize: number = 20
): Promise<PaginatedResponse<ResellerKey>> {
  const { data } = await apiClient.get<PaginatedResponse<ResellerKey>>('/reseller/keys', {
    params: { page, page_size: pageSize }
  })
  return data
}

/**
 * Create a new API key
 */
export async function create(keyData: CreateResellerKeyRequest): Promise<ResellerKey> {
  const { data } = await apiClient.post<ResellerKey>('/reseller/keys', keyData)
  return data
}

/**
 * Update an API key
 */
export async function update(id: number, updates: UpdateResellerKeyRequest): Promise<ResellerKey> {
  const { data } = await apiClient.put<ResellerKey>(`/reseller/keys/${id}`, updates)
  return data
}

/**
 * Delete an API key
 */
export async function deleteKey(id: number): Promise<{ message: string }> {
  const { data } = await apiClient.delete<{ message: string }>(`/reseller/keys/${id}`)
  return data
}

/**
 * Reset API key quota
 */
export async function resetQuota(id: number): Promise<ResellerKey> {
  const { data } = await apiClient.post<ResellerKey>(`/reseller/keys/${id}/reset-quota`)
  return data
}

export const keysAPI = {
  list,
  create,
  update,
  delete: deleteKey,
  resetQuota
}

export default keysAPI
