<script setup lang="ts">
import { RouterView, useRouter, useRoute } from 'vue-router'
import { onMounted, onBeforeUnmount, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Toast from '@/components/common/Toast.vue'
import NavigationProgress from '@/components/common/NavigationProgress.vue'
import AdminComplianceDialog from '@/components/admin/AdminComplianceDialog.vue'
import { resolveRouteDocumentTitle } from '@/router/title'
import AnnouncementPopup from '@/components/common/AnnouncementPopup.vue'
import ChatWidget from '@/components/common/ChatWidget.vue'
import { useAppStore, useAuthStore, useSubscriptionStore, useAnnouncementStore, useResellerSettingsStore, useAdminComplianceStore, useAdminSettingsStore } from '@/stores'
import { useAdminChatStore } from '@/stores/adminChat'
import { getSetupStatus } from '@/api/setup'
import { applySeoMeta } from '@/utils/seo'
import { updateFavicon } from '@/utils/branding'

const router = useRouter()
const route = useRoute()
const { t, locale } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const subscriptionStore = useSubscriptionStore()
const announcementStore = useAnnouncementStore()
const resellerSettingsStore = useResellerSettingsStore()
const adminComplianceStore = useAdminComplianceStore()
const adminChatStore = useAdminChatStore()
const adminSettingsStore = useAdminSettingsStore()

function updateDocumentTitle() {
  const customMenuItems = [
    ...(appStore.cachedPublicSettings?.custom_menu_items ?? []),
    ...(authStore.isAdmin ? adminSettingsStore.customMenuItems : []),
  ]
  document.title = resolveRouteDocumentTitle(route, appStore.siteName, customMenuItems)
}

// Watch for site settings changes and update favicon/title
watch(
  () => appStore.siteLogo,
  (newLogo) => {
    if (newLogo) {
      updateFavicon(newLogo)
    }
  },
  { immediate: true }
)

function toAbsoluteUrl(url: string) {
  if (!url) return `${window.location.origin}/logo.png`
  if (url.startsWith('http://') || url.startsWith('https://')) return url
  return new URL(url, window.location.origin).toString()
}

function updateSeo() {
  if (typeof window === 'undefined') return

  const siteName = appStore.cachedPublicSettings?.site_name || appStore.siteName || '码驿站'
  const siteSubtitle =
    appStore.cachedPublicSettings?.site_subtitle || 'AI API Gateway Platform'

  const meta = route.meta as Record<string, unknown>
  const requiresAuth = route.meta.requiresAuth !== false

  const titleKey = typeof meta.titleKey === 'string' ? meta.titleKey : ''
  const rawTitle = titleKey ? t(titleKey, { siteName }) : (meta.title as string | undefined) || siteName
  const title = rawTitle.includes(siteName) ? rawTitle : `${rawTitle} - ${siteName}`

  const descriptionKey = typeof meta.descriptionKey === 'string' ? meta.descriptionKey : ''
  const description = descriptionKey ? t(descriptionKey, { siteName, siteSubtitle }) : siteSubtitle

  const keywordsKey = typeof meta.keywordsKey === 'string' ? meta.keywordsKey : ''
  const keywords = keywordsKey ? t(keywordsKey, { siteName }) : ''

  const canonicalUrl = `${window.location.origin}${route.path}`
  const logoUrl = appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '/logo.png?v=2'
  const imageUrl = toAbsoluteUrl(logoUrl)

  const isHome = route.name === 'Home' || route.path === '/home'
  const structuredData = isHome ? [
    {
      '@context': 'https://schema.org',
      '@type': 'SoftwareApplication',
      name: siteName,
      applicationCategory: 'DeveloperApplication',
      operatingSystem: 'Web',
      url: window.location.origin,
      description,
      image: imageUrl
    },
    {
      '@context': 'https://schema.org',
      '@type': 'FAQPage',
      mainEntity: [1, 2, 3, 4, 5].map((n) => ({
        '@type': 'Question',
        name: t(`home.faq.q${n}`),
        acceptedAnswer: { '@type': 'Answer', text: t(`home.faq.a${n}`) }
      }))
    }
  ] : undefined

  applySeoMeta({
    title,
    description,
    keywords,
    canonicalUrl,
    imageUrl,
    siteName,
    locale: (locale.value === 'zh' ? 'zh' : 'en'),
    noindex: meta.noindex === true || requiresAuth,
    structuredData
  })
}

watch(
  () => [route.fullPath, locale.value, appStore.siteName, appStore.siteLogo, appStore.cachedPublicSettings?.site_subtitle],
  () => updateSeo(),
  { immediate: true }
)

watch(
  [
    () => route.fullPath,
    () => route.meta.title,
    () => route.meta.titleKey,
    () => appStore.siteName,
    () => appStore.cachedPublicSettings?.custom_menu_items,
    () => authStore.isAdmin,
    () => adminSettingsStore.customMenuItems,
  ],
  updateDocumentTitle,
  { deep: true }
)

// Watch for authentication state and manage subscription data + announcements
function onVisibilityChange() {
  if (document.visibilityState === 'visible' && authStore.isAuthenticated) {
    announcementStore.fetchAnnouncements()
  }
}

function onAdminComplianceRequired(event: Event) {
  const detail = (event as CustomEvent<Record<string, string>>).detail || {}
  adminComplianceStore.requireAcknowledgement(detail)
}

watch(
  () => authStore.isAuthenticated,
  (isAuthenticated, oldValue) => {
    if (isAuthenticated) {
      if (authStore.isAdmin) {
        adminComplianceStore.fetchStatus().catch((error) => {
          console.error('Failed to fetch admin compliance status:', error)
        })
        // 轮询客服待处理(未读)会话数,驱动侧边栏「在线客服」徽标(管理员不再有悬浮气泡)
        adminChatStore.startPolling()
      }

      // User logged in: preload subscriptions and start polling
      subscriptionStore.fetchActiveSubscriptions().catch((error) => {
        console.error('Failed to preload subscriptions:', error)
      })
      subscriptionStore.startPolling()

      // Announcements: new login vs page refresh restore
      if (oldValue === false) {
        // New login: delay 3s then force fetch
        setTimeout(() => announcementStore.fetchAnnouncements(true), 3000)
      } else {
        // Page refresh restore (oldValue was undefined)
        announcementStore.fetchAnnouncements()
      }

      // Register visibility change listener
      document.addEventListener('visibilitychange', onVisibilityChange)
    } else {
      // User logged out: clear data and stop polling
      subscriptionStore.clear()
      announcementStore.reset()
      resellerSettingsStore.reset()
      adminComplianceStore.reset()
      adminChatStore.reset()
      document.removeEventListener('visibilitychange', onVisibilityChange)
    }
  },
  { immediate: true }
)

// Route change trigger (throttled by store)
router.afterEach(() => {
  if (authStore.isAuthenticated) {
    announcementStore.fetchAnnouncements()
  }
})

onBeforeUnmount(() => {
  document.removeEventListener('visibilitychange', onVisibilityChange)
  window.removeEventListener('admin-compliance-required', onAdminComplianceRequired)
})

onMounted(async () => {
  window.addEventListener('admin-compliance-required', onAdminComplianceRequired)

  // Check if setup is needed
  try {
    const status = await getSetupStatus()
    if (status.needs_setup && route.path !== '/setup') {
      router.replace('/setup')
      return
    }
  } catch {
    // If setup endpoint fails, assume normal mode and continue
  }

  // Load public settings into appStore (will be cached for other components)
  await appStore.fetchPublicSettings()

  // Re-resolve document title now that site settings are available
  updateDocumentTitle()
})
</script>

<template>
  <NavigationProgress />
  <RouterView />
  <Toast />
  <AnnouncementPopup />
  <!-- 访客客服气泡：仅对非管理员展示；管理员通过侧边栏「在线客服」徽标提醒，无悬浮气泡 -->
  <ChatWidget v-if="!authStore.isAdmin" />
  <AdminComplianceDialog />
</template>
