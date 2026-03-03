import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import i18n, { initI18n } from './i18n'
import { useAppStore } from '@/stores/app'
import './style.css'

function initThemeClass() {
  const savedTheme = localStorage.getItem('theme')
  const shouldUseDark =
    savedTheme === 'dark' ||
    (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  document.documentElement.classList.toggle('dark', shouldUseDark)
}

async function bootstrap() {
  // Apply theme class globally before app mount to keep all routes consistent.
  initThemeClass()

  const app = createApp(App)
  const pinia = createPinia()
  app.use(pinia)

  // Initialize settings from injected config BEFORE mounting (prevents flash)
  // This must happen after pinia is installed but before router and i18n
  const appStore = useAppStore()
  const hasInjectedConfig = appStore.initFromInjectedConfig()

  // Set document title immediately after config is loaded
  function updateTitle() {
    if (appStore.siteName && appStore.siteName !== '码驿站') {
      document.title = `${appStore.siteName} - AI API Gateway`
    }
  }
  updateTitle()

  await initI18n()

  app.use(router)
  app.use(i18n)

  // If no injected config (server HTML injection failed), fetch from API before
  // mounting to prevent flash of main-site branding on reseller domains
  if (!hasInjectedConfig) {
    await Promise.race([
      appStore.fetchPublicSettings(),
      new Promise((resolve) => setTimeout(resolve, 3000)),
    ])
    updateTitle()
  }

  // 等待路由器完成初始导航后再挂载，避免竞态条件导致的空白渲染
  await router.isReady()
  app.mount('#app')
}

bootstrap()
