/**
 * Admin Channels API endpoints
 * Handles channel management for administrators
 */

import { apiClient } from '../client'
import type {
  Channel,
  CreateChannelRequest,
  UpdateChannelRequest,
  PaginatedResponse
} from '@/types'

/**
 * List all channels with pagination
 * @param page - Page number (default: 1)
 * @param pageSize - Items per page (default: 20)
 * @param filters - Optional filters
 * @returns Paginated list of channels
 */
export async function list(
  page: number = 1,
  pageSize: number = 20,
  filters?: {
    platform?: string
    status?: 'active' | 'inactive' | 'error'
    search?: string
  },
  options?: {
    signal?: AbortSignal
  }
): Promise<PaginatedResponse<Channel>> {
  const { data } = await apiClient.get<PaginatedResponse<Channel>>('/admin/channels', {
    params: {
      page,
      page_size: pageSize,
      ...filters
    },
    signal: options?.signal
  })
  return data
}

/**
 * Get channel by ID
 * @param id - Channel ID
 * @returns Channel details
 */
export async function getById(id: number): Promise<Channel> {
  const { data } = await apiClient.get<Channel>(`/admin/channels/${id}`)
  return data
}

/**
 * Create new channel
 * @param channelData - Channel data
 * @returns Created channel
 */
export async function create(channelData: CreateChannelRequest): Promise<Channel> {
  const { data } = await apiClient.post<Channel>('/admin/channels', channelData)
  return data
}

/**
 * Update channel
 * @param id - Channel ID
 * @param updates - Fields to update
 * @returns Updated channel
 */
export async function update(id: number, updates: UpdateChannelRequest): Promise<Channel> {
  const { data } = await apiClient.put<Channel>(`/admin/channels/${id}`, updates)
  return data
}

/**
 * Delete channel
 * @param id - Channel ID
 * @returns Success confirmation
 */
export async function deleteChannel(id: number): Promise<{ message: string }> {
  const { data } = await apiClient.delete<{ message: string }>(`/admin/channels/${id}`)
  return data
}

/**
 * Check channel balance
 * @param id - Channel ID
 * @returns Updated channel with balance info
 */
export async function checkBalance(id: number): Promise<Channel> {
  const { data } = await apiClient.post<Channel>(`/admin/channels/${id}/check-balance`)
  return data
}

/**
 * Toggle channel status
 * @param id - Channel ID
 * @param status - New status
 * @returns Updated channel
 */
export async function toggleStatus(id: number, status: 'active' | 'inactive'): Promise<Channel> {
  return update(id, { status })
}

export const channelsAPI = {
  list,
  getById,
  create,
  update,
  delete: deleteChannel,
  checkBalance,
  toggleStatus
}

export default channelsAPI
