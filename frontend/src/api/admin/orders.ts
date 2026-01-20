/**
 * Admin Orders API endpoints
 * Handles order management for administrators
 */

import { apiClient } from '../client'
import type { PaginatedResponse } from '@/types'

/**
 * Admin order with user info
 */
export interface AdminOrder {
  id: number
  order_no: string
  trade_no?: string
  user_id: number
  group_id: number
  amount: number
  status: 'pending' | 'paid' | 'expired' | 'refunded'
  pay_type?: string
  paid_at?: string
  subscription_id?: number
  created_at: string
  updated_at: string
  expired_at?: string
  group?: {
    id: number
    name: string
  }
  user?: {
    id: number
    email: string
    username?: string
  }
}

/**
 * List all orders with pagination and filters
 * @param page - Page number (default: 1)
 * @param pageSize - Items per page (default: 20)
 * @param filters - Optional filters (status, search)
 * @returns Paginated list of orders
 */
export async function list(
  page: number = 1,
  pageSize: number = 20,
  filters?: {
    status?: string
    search?: string
  },
  options?: {
    signal?: AbortSignal
  }
): Promise<PaginatedResponse<AdminOrder>> {
  const { data } = await apiClient.get<PaginatedResponse<AdminOrder>>(
    '/admin/orders',
    {
      params: {
        page,
        page_size: pageSize,
        ...filters
      },
      signal: options?.signal
    }
  )
  return data
}

export const ordersAPI = {
  list
}

export default ordersAPI
