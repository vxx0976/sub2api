/**
 * Reseller Settings API endpoints
 */

import { apiClient } from '../client'

export interface ResellerSettings {
  [key: string]: string
}

/**
 * Get all reseller settings
 */
export async function get(): Promise<ResellerSettings> {
  const { data } = await apiClient.get<ResellerSettings>('/reseller/settings')
  return data
}

/**
 * Update reseller settings (batch)
 */
export async function update(settings: ResellerSettings): Promise<ResellerSettings> {
  const { data } = await apiClient.put<ResellerSettings>('/reseller/settings', settings)
  return data
}

export const settingsAPI = {
  get,
  update,
}

export default settingsAPI
