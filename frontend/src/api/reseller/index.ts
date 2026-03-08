/**
 * Reseller API barrel export
 */

import dashboardAPI from './dashboard'
import domainsAPI from './domains'
import settingsAPI from './settings'
import keysAPI from './keys'
import redeemAPI from './redeem'
import announcementsAPI from './announcements'
import usersAPI from './users'
import commissionsAPI from './commissions'
import withdrawalsAPI from './withdrawals'

export const resellerAPI = {
  dashboard: dashboardAPI,
  domains: domainsAPI,
  settings: settingsAPI,
  keys: keysAPI,
  redeem: redeemAPI,
  announcements: announcementsAPI,
  users: usersAPI,
  commissions: commissionsAPI,
  withdrawals: withdrawalsAPI
}

export {
  dashboardAPI,
  domainsAPI,
  settingsAPI,
  keysAPI,
  redeemAPI,
  announcementsAPI,
  usersAPI,
  commissionsAPI,
  withdrawalsAPI
}

export default resellerAPI

// Re-export types
export type { ResellerDashboardStats } from './dashboard'
export type { ResellerDomain, CreateDomainRequest, UpdateDomainRequest } from './domains'
export type { ResellerSettings } from './settings'
export type { ResellerKey, CreateResellerKeyRequest, UpdateResellerKeyRequest } from './keys'
export type { ResellerRedeemCode, ResellerRedeemCodeListResponse, GenerateRedeemCodesRequest } from './redeem'
export type { ResellerAnnouncement, ResellerAnnouncementListResponse, CreateAnnouncementRequest, UpdateAnnouncementRequest } from './announcements'
export type { ResellerUserListResponse, ResellerBalanceHistoryItem, ResellerBalanceHistoryResponse } from './users'
export type { CommissionSummary, CommissionDetailItem, CommissionDetailResponse } from './commissions'
export type { WithdrawalItem, WithdrawalListResponse } from './withdrawals'
