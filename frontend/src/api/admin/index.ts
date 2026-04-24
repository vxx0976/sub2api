/**
 * Admin API barrel export
 * Centralized exports for all admin API modules
 */

import dashboardAPI from './dashboard'
import usersAPI from './users'
import groupsAPI from './groups'
import accountsAPI from './accounts'
import proxiesAPI from './proxies'
import redeemAPI from './redeem'
import promoAPI from './promo'
import announcementsAPI from './announcements'
import settingsAPI from './settings'
import systemAPI from './system'
import subscriptionsAPI from './subscriptions'
import usageAPI from './usage'
import geminiAPI from './gemini'
import antigravityAPI from './antigravity'
import userAttributesAPI from './userAttributes'
import opsAPI from './ops'
import errorPassthroughAPI from './errorPassthrough'
import dataManagementAPI from './dataManagement'
import apiKeysAPI from './apiKeys'
import scheduledTestsAPI from './scheduledTests'
import merchantsAPI from './merchants'
import merchantWithdrawalsAPI from './merchantWithdrawals'
import rechargeOrdersAPI from './rechargeOrders'
import alimpayOrdersAPI from './alimpayOrders'
import alimpayConfigAPI from './alimpayConfig'
import backupAPI from './backup'
import tlsFingerprintProfileAPI from './tlsFingerprintProfile'
import channelsAPI from './channels'
import adminPaymentAPI from './payment'

/**
 * Unified admin API object for convenient access
 */
export const adminAPI = {
  dashboard: dashboardAPI,
  users: usersAPI,
  groups: groupsAPI,
  accounts: accountsAPI,
  proxies: proxiesAPI,
  redeem: redeemAPI,
  promo: promoAPI,
  announcements: announcementsAPI,
  settings: settingsAPI,
  system: systemAPI,
  subscriptions: subscriptionsAPI,
  usage: usageAPI,
  gemini: geminiAPI,
  antigravity: antigravityAPI,
  userAttributes: userAttributesAPI,
  ops: opsAPI,
  channels: channelsAPI,
  errorPassthrough: errorPassthroughAPI,
  dataManagement: dataManagementAPI,
  apiKeys: apiKeysAPI,
  scheduledTests: scheduledTestsAPI,
  merchants: merchantsAPI,
  merchantWithdrawals: merchantWithdrawalsAPI,
  rechargeOrders: rechargeOrdersAPI,
  alimpayOrders: alimpayOrdersAPI,
  alimpayConfig: alimpayConfigAPI,
  backup: backupAPI,
  tlsFingerprintProfiles: tlsFingerprintProfileAPI,
  payment: adminPaymentAPI
}

export {
  dashboardAPI,
  usersAPI,
  groupsAPI,
  accountsAPI,
  proxiesAPI,
  redeemAPI,
  promoAPI,
  announcementsAPI,
  settingsAPI,
  systemAPI,
  subscriptionsAPI,
  usageAPI,
  geminiAPI,
  antigravityAPI,
  userAttributesAPI,
  opsAPI,
  channelsAPI,
  errorPassthroughAPI,
  dataManagementAPI,
  apiKeysAPI,
  scheduledTestsAPI,
  merchantsAPI,
  merchantWithdrawalsAPI,
  rechargeOrdersAPI,
  alimpayOrdersAPI,
  alimpayConfigAPI,
  backupAPI,
  tlsFingerprintProfileAPI,
  adminPaymentAPI
}

export default adminAPI

// Re-export types used by components
export type { BalanceHistoryItem } from './users'
export type { ErrorPassthroughRule, CreateRuleRequest, UpdateRuleRequest } from './errorPassthrough'
export type { BackupAgentHealth, DataManagementConfig } from './dataManagement'
export type { MerchantUser, MerchantSettings } from './merchants'
export type { AdminWithdrawalItem, AdminWithdrawalListResponse } from './merchantWithdrawals'
export type { AdminRechargeOrderItem, AdminRechargeOrderListResponse } from './rechargeOrders'
export type { AdminAliMPayOrderItem, AdminAliMPayOrderListResponse } from './alimpayOrders'
export type { AdminAliMPayConfig, AdminAliMPayConfigUpdate } from './alimpayConfig'
export type { TLSFingerprintProfile, CreateProfileRequest, UpdateProfileRequest } from './tlsFingerprintProfile'
