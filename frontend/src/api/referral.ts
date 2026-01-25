/**
 * Referral API endpoints
 * Handles user referral code and invitee management
 */

import { apiClient } from './client'

export interface ReferralCodeResponse {
  code: string
}

export interface ReferralStats {
  total_invited: number
  total_rewarded: number
  pending_payment: number
  total_earnings: number
}

export interface InviteeInfo {
  id: number
  email: string
  status: string // 'pending' | 'rewarded'
  joined_at: string
  rewarded_at?: string
  reward_amount?: number
  referrer_earning?: number
}

export interface InviteesListResponse {
  items: InviteeInfo[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

/**
 * Get or generate the user's referral code
 */
export async function getReferralCode(): Promise<ReferralCodeResponse> {
  const { data } = await apiClient.get<ReferralCodeResponse>('/referral/code')
  return data
}

/**
 * Get referral statistics for the current user
 */
export async function getReferralStats(): Promise<ReferralStats> {
  const { data } = await apiClient.get<ReferralStats>('/referral/stats')
  return data
}

/**
 * Get the list of invitees for the current user
 */
export async function getInvitees(
  page: number = 1,
  pageSize: number = 10
): Promise<InviteesListResponse> {
  const { data } = await apiClient.get<InviteesListResponse>('/referral/invitees', {
    params: { page, page_size: pageSize }
  })
  return data
}

export default {
  getReferralCode,
  getReferralStats,
  getInvitees
}
