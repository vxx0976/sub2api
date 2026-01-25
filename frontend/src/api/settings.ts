/**
 * Public Settings API endpoints
 * Handles public settings access for all users
 */

import { apiClient } from './client'
import type { PublicSettings } from '@/types'

/**
 * Get public settings (no auth required)
 * @returns Public settings including site config and announcements
 */
export async function getPublicSettings(): Promise<PublicSettings> {
  const { data } = await apiClient.get<PublicSettings>('/settings/public')
  return data
}

export const settingsAPI = {
  getPublicSettings
}

export default settingsAPI
