/**
 * Status API
 * Group status monitoring endpoints
 */

import apiClient from './client'

// Group Status types
export interface GroupStatusItem {
  id: number
  name: string
  status: 'operational' | 'degraded' | 'down' | 'unknown'
  success_rate: number
  total_requests: number
}

export interface GroupStatusResponse {
  groups: GroupStatusItem[]
  updated_at: string
}

/**
 * Get group status from our backend (public, no auth required)
 * GET /api/v1/group-status
 */
export async function getGroupStatus(): Promise<GroupStatusResponse> {
  const { data } = await apiClient.get<GroupStatusResponse>('/group-status')
  return data
}
