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

/**
 * Generate a temporary Telegram bind code
 */
export async function generateBindCode(): Promise<{ bind_code: string }> {
  const { data } = await apiClient.post<{ bind_code: string }>('/reseller/settings/tg-bind-code')
  return data
}

export const settingsAPI = {
  get,
  update,
  generateBindCode
}

export default settingsAPI
