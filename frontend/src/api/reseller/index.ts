/**
 * Reseller API barrel export
 */

import dashboardAPI from './dashboard'
import domainsAPI from './domains'
import groupsAPI from './groups'
import settingsAPI from './settings'
import keysAPI from './keys'
import redeemAPI from './redeem'
import announcementsAPI from './announcements'
import usersAPI from './users'

export const resellerAPI = {
  dashboard: dashboardAPI,
  domains: domainsAPI,
  groups: groupsAPI,
  settings: settingsAPI,
  keys: keysAPI,
  redeem: redeemAPI,
  announcements: announcementsAPI,
  users: usersAPI
}

export {
  dashboardAPI,
  domainsAPI,
  groupsAPI,
  settingsAPI,
  keysAPI,
  redeemAPI,
  announcementsAPI,
  usersAPI
}

export default resellerAPI

// Re-export types
export type { ResellerDashboardStats } from './dashboard'
export type { ResellerDomain, CreateDomainRequest, UpdateDomainRequest } from './domains'
export type { ResellerGroup, GroupTemplate, CreateResellerGroupRequest, UpdateResellerGroupRequest } from './groups'
export type { ResellerSettings } from './settings'
export type { ResellerKey, CreateResellerKeyRequest, UpdateResellerKeyRequest } from './keys'
export type { ResellerRedeemCode, ResellerRedeemCodeListResponse, GenerateRedeemCodesRequest } from './redeem'
export type { ResellerAnnouncement, ResellerAnnouncementListResponse, CreateAnnouncementRequest, UpdateAnnouncementRequest } from './announcements'
export type { ResellerUserListResponse, ResellerBalanceHistoryItem, ResellerBalanceHistoryResponse } from './users'
