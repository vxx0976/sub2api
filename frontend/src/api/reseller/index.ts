/**
 * Reseller API barrel export
 */

import dashboardAPI from './dashboard'
import domainsAPI from './domains'
import groupsAPI from './groups'
import settingsAPI from './settings'
import keysAPI from './keys'

export const resellerAPI = {
  dashboard: dashboardAPI,
  domains: domainsAPI,
  groups: groupsAPI,
  settings: settingsAPI,
  keys: keysAPI
}

export {
  dashboardAPI,
  domainsAPI,
  groupsAPI,
  settingsAPI,
  keysAPI
}

export default resellerAPI

// Re-export types
export type { ResellerDashboardStats } from './dashboard'
export type { ResellerDomain, CreateDomainRequest, UpdateDomainRequest } from './domains'
export type { ResellerGroup, GroupTemplate, CreateResellerGroupRequest, UpdateResellerGroupRequest } from './groups'
export type { ResellerSettings } from './settings'
export type { ResellerKey, CreateResellerKeyRequest, UpdateResellerKeyRequest } from './keys'
