<template>
  <header
    class="sticky top-0 z-30 border-b border-gray-200/60 bg-white/70 backdrop-blur-xl dark:border-dark-800/60 dark:bg-dark-950/50"
  >
    <nav class="mx-auto flex max-w-7xl items-center justify-between px-6 py-3">
      <!-- Brand -->
      <router-link to="/" class="flex items-center gap-3">
        <div
          class="h-10 w-10 overflow-hidden rounded-xl bg-white shadow-sm ring-1 ring-gray-200/60 dark:bg-dark-900 dark:ring-dark-800"
        >
          <img :src="siteLogo || '/logo.svg?v=2'" alt="Logo" class="h-full w-full object-contain" />
        </div>
        <div class="hidden flex-col sm:flex">
          <span class="text-sm font-semibold text-gray-900 dark:text-white">{{ siteName }}</span>
          <span class="text-xs text-gray-500 dark:text-dark-400">{{ siteSubtitle }}</span>
        </div>
      </router-link>

      <!-- Right Side: Nav + Actions -->
      <div class="flex items-center gap-6">
        <!-- Nav Links -->
        <div class="hidden items-center gap-5 text-sm font-medium lg:flex">
          <router-link
            to="/"
            class="transition-colors"
            :class="isActive('/') ? 'text-gray-900 dark:text-white' : 'text-gray-600 hover:text-gray-900 dark:text-dark-300 dark:hover:text-white'"
          >{{ t('home.nav.home') }}</router-link>
          <router-link
            to="/pricing"
            class="transition-colors"
            :class="isActive('/pricing') ? 'text-gray-900 dark:text-white' : 'text-gray-600 hover:text-gray-900 dark:text-dark-300 dark:hover:text-white'"
          >{{ t('home.nav.pricing') }}</router-link>
          <router-link
            to="/docs"
            class="transition-colors"
            :class="isActive('/docs') ? 'text-gray-900 dark:text-white' : 'text-gray-600 hover:text-gray-900 dark:text-dark-300 dark:hover:text-white'"
          >{{ t('home.nav.docs') }}</router-link>
          <router-link
            to="/status"
            class="flex items-center gap-1.5 transition-colors"
            :class="isActive('/status') ? 'text-gray-900 dark:text-white' : 'text-gray-600 hover:text-gray-900 dark:text-dark-300 dark:hover:text-white'"
          >
            {{ t('home.nav.status') }}
            <span class="relative flex h-2 w-2">
              <span class="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-75"></span>
              <span class="relative inline-flex h-2 w-2 rounded-full bg-emerald-500"></span>
            </span>
          </router-link>
        </div>

        <!-- Actions -->
        <div class="flex items-center gap-2.5">
          <LocaleSwitcher />
          <button
            @click="toggleTheme"
            class="rounded-xl p-2 text-gray-600 transition-colors hover:bg-gray-100 hover:text-gray-900 dark:text-dark-300 dark:hover:bg-dark-900 dark:hover:text-white"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
          >
            <Icon v-if="isDark" name="sun" size="md" />
            <Icon v-else name="moon" size="md" />
          </button>

          <template v-if="isAuthenticated">
            <router-link
              :to="dashboardPath"
              class="inline-flex items-center gap-1.5 rounded-full bg-gray-900 py-1 pl-1 pr-2.5 transition-colors hover:bg-gray-800 dark:bg-gray-800 dark:hover:bg-gray-700"
            >
              <span
                class="flex h-5 w-5 items-center justify-center rounded-full bg-gradient-to-br from-primary-400 to-primary-600 text-[10px] font-semibold text-white"
              >
                {{ userInitial }}
              </span>
              <span class="text-xs font-medium text-white">{{ t('home.dashboard') }}</span>
              <Icon name="arrowRight" size="xs" class="text-gray-300" :stroke-width="2" />
            </router-link>
          </template>
          <template v-else>
            <router-link
              :to="loginPath"
              class="inline-flex items-center rounded-full border border-gray-300 bg-white px-3 py-1.5 text-xs font-medium text-gray-700 transition-colors hover:bg-gray-50 dark:border-dark-600 dark:bg-dark-800 dark:text-dark-200 dark:hover:bg-dark-700"
            >
              {{ t('home.login') }}
            </router-link>
            <router-link
              :to="registerPath"
              class="inline-flex items-center rounded-full bg-gray-900 px-3 py-1.5 text-xs font-medium text-white transition-colors hover:bg-gray-800 dark:bg-gray-800 dark:hover:bg-gray-700"
            >
              {{ t('home.register') }}
            </router-link>
          </template>
        </div>
      </div>
    </nav>
  </header>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAppStore, useAuthStore } from '@/stores'
import Icon from '@/components/icons/Icon.vue'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'

const { t } = useI18n()
const route = useRoute()
const appStore = useAppStore()
const authStore = useAuthStore()

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || '码驿站')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'AI API Gateway')

const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardPath = computed(() => (isAdmin.value ? '/admin/dashboard' : '/console-home'))
const userInitial = computed(() => {
  const user = authStore.user
  if (!user || !user.email) return ''
  return user.email.charAt(0).toUpperCase()
})

// Pass ref parameter from URL to login/register pages
const refCode = computed(() => route.query.ref as string | undefined)
const registerPath = computed(() => refCode.value ? `/register?ref=${refCode.value}` : '/register')
const loginPath = computed(() => refCode.value ? `/login?redirect=${encodeURIComponent(registerPath.value)}` : '/login')

const isDark = ref(document.documentElement.classList.contains('dark'))

function isActive(path: string) {
  if (path === '/') {
    return route.path === '/' || route.path === '/home'
  }
  return route.path === path
}

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}
</script>
