<template>
  <!-- Reseller domain with doc_url: embed in iframe -->
  <div v-if="isResellerDomain && resellerDocUrl" class="flex min-h-screen flex-col bg-gray-50 dark:bg-dark-950">
    <PublicHeader />
    <main class="flex-1">
      <iframe
        :src="resellerDocUrl"
        class="h-[calc(100vh-57px)] w-full border-0"
        allowfullscreen
        referrerpolicy="no-referrer-when-downgrade"
      ></iframe>
    </main>
  </div>

  <!-- Normal docs page -->
  <PublicLayout v-else>
    <div class="mx-auto max-w-4xl">
      <!-- Header -->
      <div class="mb-8 text-center">
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">{{ t('docs.entry.title') }}</h1>
        <p class="mt-3 text-gray-600 dark:text-gray-400">{{ t('docs.entry.subtitle') }}</p>
      </div>

      <!-- Method Selection Cards -->
      <div class="mb-8 grid gap-4 sm:grid-cols-2">
        <!-- VSCode Card -->
        <router-link
          to="/docs/vscode"
          class="group relative rounded-xl border border-gray-200 bg-white p-6 transition-all hover:border-primary-300 hover:shadow-lg dark:border-dark-700 dark:bg-dark-800 dark:hover:border-primary-600"
        >
          <div class="mb-4 flex items-center gap-3">
            <div class="flex h-12 w-12 items-center justify-center rounded-lg bg-gradient-to-br from-blue-500 to-blue-600 text-white">
              <svg class="h-6 w-6" viewBox="0 0 24 24" fill="currentColor">
                <path d="M17.583 2L6.25 11.042 2.833 8.375.5 9.75l4.917 4.25L.5 18.25l2.333 1.375 3.417-2.667L17.583 26 21.5 24.25V3.75L17.583 2zm-.833 3.958v16.084l-8.917-8.042 8.917-8.042z"/>
              </svg>
            </div>
            <div>
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white">VSCode</h3>
              <span class="inline-block rounded-full bg-green-100 px-2 py-0.5 text-xs font-medium text-green-700 dark:bg-green-900/30 dark:text-green-400">{{ t('docs.entry.recommended') }}</span>
            </div>
          </div>
          <p class="text-sm text-gray-600 dark:text-gray-400">{{ t('docs.entry.vscodeDesc') }}</p>
          <div class="mt-4 flex items-center text-sm font-medium text-primary-600 dark:text-primary-400">
            {{ t('docs.entry.viewGuide') }}
            <svg class="ml-1 h-4 w-4 transition-transform group-hover:translate-x-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
            </svg>
          </div>
        </router-link>

        <!-- CLI Card -->
        <router-link
          to="/docs/cli"
          class="group relative rounded-xl border border-gray-200 bg-white p-6 transition-all hover:border-primary-300 hover:shadow-lg dark:border-dark-700 dark:bg-dark-800 dark:hover:border-primary-600"
        >
          <div class="mb-4 flex items-center gap-3">
            <div class="flex h-12 w-12 items-center justify-center rounded-lg bg-gradient-to-br from-gray-700 to-gray-900 text-white dark:from-gray-600 dark:to-gray-800">
              <svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
            </div>
            <div>
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white">Claude Code CLI</h3>
              <span class="inline-block rounded-full bg-blue-100 px-2 py-0.5 text-xs font-medium text-blue-700 dark:bg-blue-900/30 dark:text-blue-400">{{ t('docs.entry.advanced') }}</span>
            </div>
          </div>
          <p class="text-sm text-gray-600 dark:text-gray-400">{{ t('docs.entry.cliDesc') }}</p>
          <div class="mt-4 flex items-center text-sm font-medium text-primary-600 dark:text-primary-400">
            {{ t('docs.entry.viewGuide') }}
            <svg class="ml-1 h-4 w-4 transition-transform group-hover:translate-x-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
            </svg>
          </div>
        </router-link>

        <!-- Tips Card -->
        <router-link
          to="/docs/tips"
          class="group relative rounded-xl border border-gray-200 bg-white p-6 transition-all hover:border-primary-300 hover:shadow-lg dark:border-dark-700 dark:bg-dark-800 dark:hover:border-primary-600"
        >
          <div class="mb-4 flex items-center gap-3">
            <div class="flex h-12 w-12 items-center justify-center rounded-lg bg-gradient-to-br from-amber-500 to-orange-500 text-white">
              <svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
              </svg>
            </div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white">{{ t('docs.entry.tipsTitle') }}</h3>
          </div>
          <p class="text-sm text-gray-600 dark:text-gray-400">{{ t('docs.entry.tipsDesc') }}</p>
          <div class="mt-4 flex items-center text-sm font-medium text-primary-600 dark:text-primary-400">
            {{ t('docs.entry.viewGuide') }}
            <svg class="ml-1 h-4 w-4 transition-transform group-hover:translate-x-1" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 5l7 7-7 7" />
            </svg>
          </div>
        </router-link>
      </div>

      <!-- Quick Info Box -->
      <div class="mb-8 rounded-xl border border-blue-200 bg-blue-50 p-6 dark:border-blue-800 dark:bg-blue-900/20">
        <h3 class="mb-4 flex items-center gap-2 font-semibold text-blue-800 dark:text-blue-200">
          <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          {{ t('docs.entry.quickInfo.title') }}
        </h3>
        <div class="grid gap-4 sm:grid-cols-3">
          <div class="rounded-lg bg-white p-4 dark:bg-dark-800">
            <div class="mb-2 text-2xl font-bold text-blue-600 dark:text-blue-400">1</div>
            <div class="text-sm font-medium text-gray-900 dark:text-white">{{ t('docs.entry.quickInfo.step1Title') }}</div>
            <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('docs.entry.quickInfo.step1Desc') }}</div>
          </div>
          <div class="rounded-lg bg-white p-4 dark:bg-dark-800">
            <div class="mb-2 text-2xl font-bold text-blue-600 dark:text-blue-400">2</div>
            <div class="text-sm font-medium text-gray-900 dark:text-white">{{ t('docs.entry.quickInfo.step2Title') }}</div>
            <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('docs.entry.quickInfo.step2Desc') }}</div>
          </div>
          <div class="rounded-lg bg-white p-4 dark:bg-dark-800">
            <div class="mb-2 text-2xl font-bold text-blue-600 dark:text-blue-400">3</div>
            <div class="text-sm font-medium text-gray-900 dark:text-white">{{ t('docs.entry.quickInfo.step3Title') }}</div>
            <div class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('docs.entry.quickInfo.step3Desc') }}</div>
          </div>
        </div>
      </div>

      <!-- FAQ Section -->
      <section class="rounded-xl border border-gray-200 bg-white p-6 dark:border-dark-700 dark:bg-dark-800">
        <h3 class="mb-4 text-lg font-semibold text-gray-900 dark:text-white">{{ t('docs.entry.faq.title') }}</h3>
        <div class="space-y-4">
          <div>
            <p class="font-medium text-gray-800 dark:text-gray-200">{{ t('docs.entry.faq.q1') }}</p>
            <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">{{ t('docs.entry.faq.a1') }}</p>
          </div>
          <div>
            <p class="font-medium text-gray-800 dark:text-gray-200">{{ t('docs.entry.faq.q2') }}</p>
            <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">{{ t('docs.entry.faq.a2') }}</p>
          </div>
          <div>
            <p class="font-medium text-gray-800 dark:text-gray-200">{{ t('docs.entry.faq.q3') }}</p>
            <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">{{ t('docs.entry.faq.a3') }}</p>
          </div>
          <div>
            <p class="font-medium text-gray-800 dark:text-gray-200">{{ t('docs.entry.faq.q4') }}</p>
            <p class="mt-1 text-sm text-gray-600 dark:text-gray-400">{{ t('docs.entry.faq.a4') }}</p>
          </div>
        </div>
      </section>
    </div>
  </PublicLayout>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import PublicLayout from '@/components/layout/PublicLayout.vue'
import PublicHeader from '@/components/layout/PublicHeader.vue'

const { t } = useI18n()
const appStore = useAppStore()

// Reseller domain detection
const isResellerDomain = computed(() => !!appStore.cachedPublicSettings?.reseller_id)
const resellerDocUrl = computed(() => appStore.cachedPublicSettings?.doc_url || '')
</script>
