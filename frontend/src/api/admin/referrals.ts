import client from '../client'

export interface AdminReferralRecord {
  id: number
  referrer_id: number
  referrer_email: string
  invitee_id: number
  invitee_email: string
  trigger_order_id: number
  referrer_reward: number
  invitee_reward: number
  skip_referrer_reason?: string
  status: 'pending' | 'rewarded'
  created_at: string
}

export interface AdminReferralListResponse {
  items: AdminReferralRecord[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export interface AdminReferralStats {
  total_records: number
  total_referrers: number
  total_invitees: number
  total_pending: number
  total_rewarded: number
  total_referrer_paid: number
  total_invitee_paid: number
}

export async function getReferrals(
  page: number = 1,
  pageSize: number = 20,
  search?: string
): Promise<AdminReferralListResponse> {
  const params = new URLSearchParams({
    page: String(page),
    page_size: String(pageSize)
  })
  if (search) {
    params.append('search', search)
  }
  const { data } = await client.get<AdminReferralListResponse>(`/admin/referrals?${params.toString()}`)
  return data
}

export async function getReferralStats(): Promise<AdminReferralStats> {
  const { data } = await client.get<AdminReferralStats>('/admin/referrals/stats')
  return data
}
