<script setup lang="ts">
import { RouterView, useRouter, useRoute } from 'vue-router'
import { onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Toast from '@/components/common/Toast.vue'
import NavigationProgress from '@/components/common/NavigationProgress.vue'
import { useAppStore, useAuthStore, useSubscriptionStore } from '@/stores'
import { getSetupStatus } from '@/api/setup'
import { applySeoMeta } from '@/utils/seo'

const router = useRouter()
const route = useRoute()
const { t, locale } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()
const subscriptionStore = useSubscriptionStore()

/**
 * Update favicon dynamically
 * @param logoUrl - URL of the logo to use as favicon
 */
function updateFavicon(logoUrl: string) {
  // Find existing favicon link or create new one
  let link = document.querySelector<HTMLLinkElement>('link[rel="icon"]')
  if (!link) {
    link = document.createElement('link')
    link.rel = 'icon'
    document.head.appendChild(link)
  }
  link.type = logoUrl.endsWith('.svg') ? 'image/svg+xml' : 'image/x-icon'
  link.href = logoUrl
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
  const keywords = keywordsKey ? t(keywordsKey) : ''

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

// Watch for authentication state and manage subscription data
watch(
  () => authStore.isAuthenticated,
  (isAuthenticated) => {
    if (isAuthenticated) {
      // User logged in: preload subscriptions and start polling
      subscriptionStore.fetchActiveSubscriptions().catch((error) => {
        console.error('Failed to preload subscriptions:', error)
      })
      subscriptionStore.startPolling()
    } else {
      // User logged out: clear data and stop polling
      subscriptionStore.clear()
    }
  },
  { immediate: true }
)

onMounted(async () => {
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
})
</script>

<template>
  <NavigationProgress />
  <RouterView />
  <Toast />
</template>
