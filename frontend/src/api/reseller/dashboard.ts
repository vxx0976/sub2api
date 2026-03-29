/**
 * Reseller Dashboard API endpoints
 */

import { apiClient } from '../client'

export interface ResellerDashboardStats {
  my_balance: number
  domain_count: number
  verified_domains: number
  group_count: number
  key_count: number
  active_key_count: number
  total_quota_used: number
}

/**
 * Get reseller dashboard statistics
 */
export async function getStats(): Promise<ResellerDashboardStats> {
  const { data } = await apiClient.get<ResellerDashboardStats>('/reseller/dashboard')
  return data
}

export const dashboardAPI = { getStats }
export default dashboardAPI
