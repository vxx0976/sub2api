<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Announcements -->
      <div v-if="announcementHtml" class="card p-5">
        <div class="mb-3 flex items-center gap-2">
          <Icon name="chatBubble" size="md" class="text-amber-500" />
          <h2 class="text-base font-semibold text-gray-900 dark:text-white">
            {{ t('consoleHome.announcements') }}
          </h2>
        </div>
        <div class="markdown-body prose prose-sm max-w-none dark:prose-invert" v-html="announcementHtml"></div>
      </div>

      <!-- Getting Started + Contact Us (side by side) -->
      <div class="grid gap-4 sm:grid-cols-2">
        <!-- Getting Started Guide -->
        <a
          v-if="docUrl"
          :href="docUrl"
          target="_blank"
          rel="noopener noreferrer"
          class="group card flex items-center gap-4 p-5 transition-all hover:shadow-lg hover:ring-2 hover:ring-primary-500/50"
        >
          <div class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-sky-500 to-blue-600 shadow shadow-blue-500/30 transition-transform group-hover:scale-105">
            <Icon name="book" size="lg" class="text-white" />
          </div>
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2">
              <h3 class="text-base font-bold text-gray-900 dark:text-white">
                {{ t('consoleHome.gettingStarted.title') }}
              </h3>
              <span class="rounded-full bg-amber-100 px-2 py-0.5 text-xs font-semibold text-amber-700 dark:bg-amber-900/30 dark:text-amber-400">
                {{ t('consoleHome.gettingStarted.badge') }}
              </span>
            </div>
            <p class="mt-1 text-sm text-gray-500 dark:text-dark-400 truncate">
              {{ t('consoleHome.gettingStarted.description') }}
            </p>
          </div>
          <Icon name="externalLink" size="md" class="flex-shrink-0 text-gray-400 transition-colors group-hover:text-primary-500 dark:text-dark-500" />
        </a>

        <!-- Contact Us -->
        <div v-if="hasContactInfo" class="card p-5">
          <div class="mb-3 flex items-center gap-2">
            <Icon name="chat" size="md" class="text-sky-500" />
            <h3 class="text-base font-bold text-gray-900 dark:text-white">
              {{ t('consoleHome.contact.title') }}
            </h3>
          </div>
          <div class="flex-1 space-y-2">
              <!-- WeChat -->
              <div v-if="contactWechat" class="flex items-center gap-2">
                <svg class="h-4 w-4 flex-shrink-0 text-green-500" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M8.691 2.188C3.891 2.188 0 5.476 0 9.53c0 2.212 1.17 4.203 3.002 5.55a.59.59 0 01.213.665l-.39 1.48c-.019.07-.048.141-.048.213 0 .163.13.295.29.295a.326.326 0 00.167-.054l1.903-1.114a.864.864 0 01.717-.098 10.16 10.16 0 002.837.403c.276 0 .543-.027.811-.05-.857-2.578.157-4.972 1.932-6.446 1.703-1.415 3.882-1.98 5.853-1.838-.576-3.583-4.196-6.348-8.596-6.348zM5.785 5.991c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 01-1.162 1.178A1.17 1.17 0 014.623 7.17c0-.651.52-1.18 1.162-1.18zm5.813 0c.642 0 1.162.529 1.162 1.18a1.17 1.17 0 01-1.162 1.178 1.17 1.17 0 01-1.162-1.178c0-.651.52-1.18 1.162-1.18zm5.34 2.867c-1.797-.052-3.746.512-5.28 1.786-1.72 1.428-2.687 3.72-1.78 6.22.942 2.453 3.666 4.229 6.884 4.229.826 0 1.622-.12 2.361-.336a.722.722 0 01.598.082l1.584.926a.272.272 0 00.14.045c.134 0 .24-.109.24-.245 0-.06-.024-.12-.04-.177l-.325-1.233a.493.493 0 01.177-.554C23.04 18.423 24 16.837 24 15.069c0-3.07-3.022-5.997-7.062-6.21zM13.544 12.5c.535 0 .969.44.969.982a.976.976 0 01-.969.983.976.976 0 01-.969-.983c0-.542.434-.982.97-.982zm4.844 0c.535 0 .969.44.969.982a.976.976 0 01-.969.983.976.976 0 01-.969-.983c0-.542.434-.982.969-.982z"/>
                </svg>
                <span class="text-sm font-medium text-gray-700 dark:text-dark-300">{{ t('consoleHome.contact.wechat') }}: {{ contactWechat }}</span>
              </div>
              <!-- Telegram -->
              <a
                v-if="contactTelegram"
                :href="`https://t.me/${contactTelegram}`"
                target="_blank"
                rel="noopener noreferrer"
                class="flex items-center gap-2 transition-colors hover:text-blue-700"
              >
                <svg class="h-4 w-4 flex-shrink-0 text-blue-500" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M11.944 0A12 12 0 000 12a12 12 0 0012 12 12 12 0 0012-12A12 12 0 0012 0a12 12 0 00-.056 0zm4.962 7.224c.1-.002.321.023.465.14a.506.506 0 01.171.325c.016.093.036.306.02.472-.18 1.898-.962 6.502-1.36 8.627-.168.9-.499 1.201-.82 1.23-.696.065-1.225-.46-1.9-.902-1.056-.693-1.653-1.124-2.678-1.8-1.185-.78-.417-1.21.258-1.91.177-.184 3.247-2.977 3.307-3.23.007-.032.014-.15-.056-.212s-.174-.041-.249-.024c-.106.024-1.793 1.14-5.061 3.345-.479.33-.913.49-1.302.48-.428-.008-1.252-.241-1.865-.44-.752-.245-1.349-.374-1.297-.789.027-.216.325-.437.893-.663 3.498-1.524 5.83-2.529 6.998-3.014 3.332-1.386 4.025-1.627 4.476-1.635z"/>
                </svg>
                <span class="text-sm font-medium text-gray-700 dark:text-dark-300">{{ t('consoleHome.contact.telegram') }}: @{{ contactTelegram }}</span>
              </a>
              <!-- QQ -->
              <div v-if="contactQQ" class="flex items-center gap-2">
                <svg class="h-4 w-4 flex-shrink-0 text-sky-500" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M12.003 2C6.004 2 3 6.086 3 9.166c0 3.313 1.727 6.286 2.907 7.594-.09.86-.455 2.11-.809 3.063-.18.484.235.696.56.512 1.108-.628 2.613-1.62 3.31-2.12.98.254 1.965.384 3.035.384 6 0 9.003-4.086 9.003-7.166C21.006 6.086 18.003 2 12.003 2z"/>
                </svg>
                <span class="text-sm font-medium text-gray-700 dark:text-dark-300">{{ t('consoleHome.contact.qq') }}: {{ contactQQ }}</span>
              </div>
              <!-- Email -->
              <a
                v-if="contactEmail"
                :href="`mailto:${contactEmail}`"
                class="flex items-center gap-2 transition-colors hover:text-red-700"
              >
                <svg class="h-4 w-4 flex-shrink-0 text-red-500" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M1.5 8.67v8.58a3 3 0 003 3h15a3 3 0 003-3V8.67l-8.928 5.493a3 3 0 01-3.144 0L1.5 8.67z" />
                  <path d="M22.5 6.908V6.75a3 3 0 00-3-3h-15a3 3 0 00-3 3v.158l9.714 5.978a1.5 1.5 0 001.572 0L22.5 6.908z" />
                </svg>
                <span class="text-sm font-medium text-gray-700 dark:text-dark-300">{{ t('consoleHome.contact.email') }}: {{ contactEmail }}</span>
              </a>
            </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { Marked } from 'marked'
import DOMPurify from 'dompurify'
import { useAppStore, useAuthStore } from '@/stores'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'

const md = new Marked({ breaks: true, gfm: true })

const { t } = useI18n()
const appStore = useAppStore()
const authStore = useAuthStore()

// Doc URL from settings
const docUrl = computed(() => appStore.cachedPublicSettings?.doc_url || '')

// Dynamic contact info — from public settings (main site or reseller)
const contactWechat = computed(() => appStore.cachedPublicSettings?.contact_wechat || '')
const contactTelegram = computed(() => appStore.cachedPublicSettings?.contact_telegram || '')
const contactQQ = computed(() => appStore.cachedPublicSettings?.contact_qq || '')
const contactEmail = computed(() => (appStore.isResellerDomain || authStore.isResellerUser) ? '' : 'vanxuehan@gmail.com')
const hasContactInfo = computed(() => !!contactWechat.value || !!contactTelegram.value || !!contactQQ.value || !!contactEmail.value)

// Announcements from cachedPublicSettings (injected via __APP_CONFIG__).
// On reseller domains, embed_on.go replaces system announcements with the reseller's own.
const announcementHtml = computed(() => {
  const items = appStore.cachedPublicSettings?.announcements ?? []
  // Merge all content into one markdown document
  const markdown = items
    .map(a => {
      const parts = []
      if (a.title) parts.push(`## ${a.title}`)
      if (a.content) parts.push(a.content)
      return parts.join('\n\n')
    })
    .filter(Boolean)
    .join('\n\n---\n\n')
  if (!markdown) return ''
  return DOMPurify.sanitize(md.parse(markdown) as string)
})
</script>
