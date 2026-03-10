/**
 * Custom routes added by dev branch
 * These are separated to minimize merge conflicts with upstream
 */
import type { RouteRecordRaw } from 'vue-router'

export const customRoutes: RouteRecordRaw[] = [
  // ==================== Custom User Routes ====================
  {
    path: '/console-home',
    name: 'ConsoleHome',
    component: () => import('@/views/user/ConsoleHomeView.vue'),
    meta: {
      requiresAuth: true,
      requiresAdmin: false,
      title: 'Console Home',
      titleKey: 'consoleHome.title',
      descriptionKey: 'consoleHome.description'
    }
  },
  {
    path: '/docs',
    name: 'DocsPortal',
    component: () => import('@/views/user/DocsPortalView.vue'),
    meta: {
      requiresAuth: false,
      title: 'Documentation',
      titleKey: 'docs.title',
      descriptionKey: 'docs.subtitle'
    }
  },
  {
    path: '/docs/claude',
    name: 'DocsClaude',
    component: () => import('@/views/user/DocsView.vue'),
    meta: {
      requiresAuth: false,
      title: 'Claude Code Guide',
      titleKey: 'docs.guide.title'
    }
  },
  {
    path: '/docs/codex',
    name: 'DocsCodex',
    component: () => import('@/views/user/CodexDocsView.vue'),
    meta: {
      requiresAuth: false,
      title: 'Codex CLI Guide',
      titleKey: 'docs.codex.title'
    }
  },
  {
    path: '/status',
    name: 'Status',
    beforeEnter() {
      window.location.href = 'https://status.claude.com'
      return false
    },
    component: () => import('@/views/HomeView.vue'), // placeholder, never rendered
    meta: { requiresAuth: false }
  },

  // ==================== Merchant Routes ====================
  {
    path: '/merchant',
    redirect: '/dashboard'
  },
  // Legacy reseller paths redirect to merchant
  { path: '/reseller', redirect: '/merchant' },
  { path: '/reseller/domains', redirect: '/merchant/sites' },
  { path: '/reseller/settings', redirect: '/merchant/settings' },
  {
    path: '/merchant/sites',
    name: 'MerchantSites',
    component: () => import('@/views/reseller/DomainsView.vue'),
    meta: {
      requiresAuth: true,
      requiresReseller: true,
      title: 'Site Management',
      titleKey: 'reseller.sites.title'
    }
  },
  {
    path: '/merchant/settings',
    name: 'MerchantSettings',
    component: () => import('@/views/reseller/SettingsView.vue'),
    meta: {
      requiresAuth: true,
      requiresReseller: true,
      title: 'Settings',
      titleKey: 'reseller.settings.title'
    }
  },
  {
    path: '/merchant/redeem',
    name: 'MerchantRedeem',
    component: () => import('@/views/reseller/RedeemView.vue'),
    meta: {
      requiresAuth: true,
      requiresReseller: true,
      title: 'Redeem Codes',
      titleKey: 'reseller.redeem.title'
    }
  },
  {
    path: '/merchant/users',
    name: 'MerchantUsers',
    component: () => import('@/views/reseller/UsersView.vue'),
    meta: {
      requiresAuth: true,
      requiresReseller: true,
      title: 'User Management',
      titleKey: 'reseller.users.title'
    }
  },
  {
    path: '/merchant/announcements',
    name: 'MerchantAnnouncements',
    component: () => import('@/views/reseller/AnnouncementsView.vue'),
    meta: {
      requiresAuth: true,
      requiresReseller: true,
      title: 'Announcements',
      titleKey: 'reseller.announcements.title'
    }
  },
  {
    path: '/merchant/commissions',
    name: 'MerchantCommissions',
    component: () => import('@/views/reseller/CommissionsView.vue'),
    meta: {
      requiresAuth: true,
      requiresReseller: true,
      requiresAgentEnabled: true,
      title: 'Commission Statistics',
      titleKey: 'reseller.commissions.title'
    }
  },
  {
    path: '/merchant/withdrawals',
    name: 'MerchantWithdrawals',
    component: () => import('@/views/reseller/WithdrawalsView.vue'),
    meta: {
      requiresAuth: true,
      requiresReseller: true,
      requiresAgentEnabled: true,
      title: 'Withdrawals',
      titleKey: 'reseller.withdrawals.title'
    }
  },

  // ==================== Custom Admin Routes ====================
  {
    path: '/admin/channels',
    name: 'AdminChannels',
    component: () => import('@/views/admin/ChannelsView.vue'),
    meta: {
      requiresAuth: true,
      requiresAdmin: true,
      title: 'Channel Management',
      titleKey: 'admin.channels.title',
      descriptionKey: 'admin.channels.description'
    }
  },
  {
    path: '/admin/merchants',
    name: 'AdminMerchants',
    component: () => import('@/views/admin/MerchantsView.vue'),
    meta: {
      requiresAuth: true,
      requiresAdmin: true,
      title: 'Merchant Management',
      titleKey: 'admin.merchants.title'
    }
  },
  {
    path: '/admin/merchant-withdrawals',
    name: 'AdminMerchantWithdrawals',
    component: () => import('@/views/admin/MerchantWithdrawalsView.vue'),
    meta: {
      requiresAuth: true,
      requiresAdmin: true,
      title: 'Merchant Withdrawals',
      titleKey: 'admin.merchantWithdrawals.title'
    }
  },

  // ==================== Public Key Query ====================
  {
    path: '/key-query',
    name: 'KeyQuery',
    component: () => import('@/views/KeyQueryView.vue'),
    meta: {
      requiresAuth: false,
      title: 'API Key Query',
      noindex: true
    }
  },
]
