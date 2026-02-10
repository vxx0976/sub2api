/**
 * Reseller Group (Package) API endpoints
 */

import { apiClient } from '../client'
import type { PaginatedResponse } from '@/types'

export interface ResellerGroup {
  id: number
  name: string
  description: string | null
  price: number | null
  daily_limit_usd: number | null
  weekly_limit_usd: number | null
  monthly_limit_usd: number | null
  default_validity_days: number
  is_purchasable: boolean
  sort_order: number
  status: 'active' | 'inactive'
  source_group_id: number | null
  created_at: string
  updated_at: string
}

export interface GroupTemplate {
  id: number
  name: string
  description: string | null
}

export interface CreateResellerGroupRequest {
  name: string
  description?: string | null
  source_group_id: number
  price?: number | null
  daily_limit_usd?: number | null
  weekly_limit_usd?: number | null
  monthly_limit_usd?: number | null
  default_validity_days?: number
  is_purchasable?: boolean
  sort_order?: number
}

export interface UpdateResellerGroupRequest {
  name?: string
  description?: string | null
  price?: number | null
  daily_limit_usd?: number | null
  weekly_limit_usd?: number | null
  monthly_limit_usd?: number | null
  default_validity_days?: number
  is_purchasable?: boolean
  sort_order?: number
  status?: 'active' | 'inactive'
}

/**
 * List admin group templates available for reseller use
 */
export async function listTemplates(): Promise<GroupTemplate[]> {
  const { data } = await apiClient.get<GroupTemplate[]>('/reseller/groups/templates')
  return data
}

/**
 * List reseller's groups (packages)
 */
export async function list(
  page: number = 1,
  pageSize: number = 20
): Promise<PaginatedResponse<ResellerGroup>> {
  const { data } = await apiClient.get<PaginatedResponse<ResellerGroup>>('/reseller/groups', {
    params: { page, page_size: pageSize }
  })
  return data
}

/**
 * Get a group by ID
 */
export async function getById(id: number): Promise<ResellerGroup> {
  const { data } = await apiClient.get<ResellerGroup>(`/reseller/groups/${id}`)
  return data
}

/**
 * Create a new group (package)
 */
export async function create(groupData: CreateResellerGroupRequest): Promise<ResellerGroup> {
  const { data } = await apiClient.post<ResellerGroup>('/reseller/groups', groupData)
  return data
}

/**
 * Update a group
 */
export async function update(id: number, updates: UpdateResellerGroupRequest): Promise<ResellerGroup> {
  const { data } = await apiClient.put<ResellerGroup>(`/reseller/groups/${id}`, updates)
  return data
}

/**
 * Delete a group
 */
export async function deleteGroup(id: number): Promise<{ message: string }> {
  const { data } = await apiClient.delete<{ message: string }>(`/reseller/groups/${id}`)
  return data
}

export const groupsAPI = {
  listTemplates,
  list,
  getById,
  create,
  update,
  delete: deleteGroup
}

export default groupsAPI
