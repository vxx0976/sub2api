import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import i18n from './i18n'
import './style.css'

const app = createApp(App)
const pinia = createPinia()
app.use(pinia)

// Initialize settings from injected config BEFORE mounting (prevents flash)
// This must happen after pinia is installed but before router and i18n
import { useAppStore } from '@/stores/app'
const appStore = useAppStore()
const hasInjectedConfig = appStore.initFromInjectedConfig()

// Set document title immediately after config is loaded
function updateTitle() {
  if (appStore.siteName && appStore.siteName !== '码驿站') {
    document.title = `${appStore.siteName} - AI API Gateway`
  }
}
updateTitle()

app.use(router)
app.use(i18n)

async function mount() {
  // If no injected config (server HTML injection failed), fetch from API before
  // mounting to prevent flash of main-site branding on reseller domains
  if (!hasInjectedConfig) {
    await Promise.race([
      appStore.fetchPublicSettings(),
      new Promise((resolve) => setTimeout(resolve, 3000)),
    ])
    updateTitle()
  }

  await router.isReady()
  app.mount('#app')
}

mount()
